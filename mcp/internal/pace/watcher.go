package pace

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/berquerant/agent-skills/mcp/internal/store"
)

// Watch registers a background process/command and records its initial log/artifact state.
func (t *Tracker) Watch(in WatchInput, now time.Time) (WatchOutput, error) {
	if err := guard.CheckPathAllowed(in.LogPath); err != nil {
		return WatchOutput{}, err
	}
	if err := guard.CheckPathAllowed(in.ArtifactPath); err != nil {
		return WatchOutput{}, err
	}
	if err := guard.CheckNoSecrets(in.WatchID, in.Command, in.LogPath, in.ArtifactPath); err != nil {
		return WatchOutput{}, err
	}
	if in.PID <= 0 && strings.TrimSpace(in.LogPath) == "" && strings.TrimSpace(in.ArtifactPath) == "" {
		return WatchOutput{}, fmt.Errorf("at least one of pid (> 0), log_path, or artifact_path is required")
	}

	watchID := strings.TrimSpace(in.WatchID)
	if watchID == "" {
		id, err := store.GenerateID()
		if err != nil {
			return WatchOutput{}, err
		}
		watchID = "watch-" + id[:12]
	} else if !safePaceIDRe.MatchString(watchID) {
		return WatchOutput{}, fmt.Errorf("invalid watch_id %q", in.WatchID)
	}

	mode, intervalSec, err := ResolveExecMode(in.Mode)
	if err != nil {
		return WatchOutput{}, err
	}

	nowStr := now.UTC().Format(time.RFC3339)
	logSize, logMod, _ := statFileMeta(in.LogPath)
	_, artMod, _ := statFileMeta(in.ArtifactPath)

	entry := WatchEntry{
		WatchID:          watchID,
		PID:              in.PID,
		Command:          strings.TrimSpace(in.Command),
		Mode:             mode,
		CheckIntervalSec: intervalSec,
		LogPath:          strings.TrimSpace(in.LogPath),
		ArtifactPath:     strings.TrimSpace(in.ArtifactPath),
		StartedAt:        nowStr,
		LastActivityAt:   nowStr,
		LastLogSize:      logSize,
		LastLogMod:       logMod,
		LastArtifactMod:  artMod,
	}
	if err := t.saveWatch(entry); err != nil {
		return WatchOutput{}, err
	}
	return WatchOutput{OK: true, Watch: entry}, nil
}

// CheckLiveness samples a registered watch entry's process and log/artifact activity.
func (t *Tracker) CheckLiveness(in LivenessInput, now time.Time) (LivenessOutput, error) {
	watchID := strings.TrimSpace(in.WatchID)
	if !safePaceIDRe.MatchString(watchID) {
		return LivenessOutput{}, fmt.Errorf("invalid watch_id %q", in.WatchID)
	}

	var out LivenessOutput
	err := t.withWatchLock(watchID, func() error {
		entry, err := t.loadWatchUnlocked(watchID)
		if err != nil {
			return err
		}
		out = sampleWatchState(&entry, now)
		return t.saveWatchUnlocked(entry)
	})
	return out, err
}

func sampleWatchState(entry *WatchEntry, now time.Time) LivenessOutput {
	nowStr := now.UTC().Format(time.RFC3339)
	elapsedSec := secondsBetween(entry.StartedAt, now)

	logAdvanced := false
	if entry.LogPath != "" {
		sz, mod, ok := statFileMeta(entry.LogPath)
		if (ok && sz != entry.LastLogSize) || (mod != "" && mod != entry.LastLogMod) {
			logAdvanced = true
			entry.LastLogSize = sz
			entry.LastLogMod = mod
		}
	}
	artChanged := false
	if entry.ArtifactPath != "" {
		if _, mod, _ := statFileMeta(entry.ArtifactPath); mod != "" && mod != entry.LastArtifactMod {
			artChanged = true
			entry.LastArtifactMod = mod
		}
	}
	if logAdvanced || artChanged {
		entry.LastActivityAt = nowStr
		entry.StallChecks = 0
	}

	idleSec := secondsBetween(entry.LastActivityAt, now)
	procAlive, isZombie, cpuPct, rssKB := inspectProcess(entry.PID)
	if cpuPct > 0.1 {
		entry.LastActivityAt = nowStr
		idleSec = 0
		entry.StallChecks = 0
	}

	state, reasons := classifyLiveness(entry, procAlive, isZombie, cpuPct, logAdvanced, artChanged, idleSec)
	report := ""
	if state == LivenessStalled || state == LivenessZombie {
		report = renderStallReport(*entry, elapsedSec, idleSec, cpuPct, rssKB, reasons)
	}

	return LivenessOutput{
		OK:          true,
		WatchID:     entry.WatchID,
		State:       state,
		ElapsedSec:  elapsedSec,
		IdleSec:     idleSec,
		CPU:         cpuPct,
		RSSKB:       rssKB,
		LogAdvanced: logAdvanced,
		Reasons:     reasons,
		StallReport: report,
		Watch:       *entry,
	}
}

func classifyLiveness(
	entry *WatchEntry,
	procAlive, isZombie bool,
	cpuPct float64,
	logAdvanced, artChanged bool,
	idleSec int,
) (LivenessState, []string) {
	if entry.PID > 0 && !procAlive {
		return LivenessExited, []string{fmt.Sprintf("process %d has exited", entry.PID)}
	}
	if isZombie {
		return LivenessZombie, []string{fmt.Sprintf("process %d is in zombie (Z) state", entry.PID)}
	}
	if !logAdvanced && !artChanged && cpuPct <= 0.1 && idleSec >= entry.CheckIntervalSec {
		entry.StallChecks++
		msg := fmt.Sprintf("no log/artifact activity for %ds (>= check interval %ds)", idleSec, entry.CheckIntervalSec)
		if entry.PID > 0 {
			msg = fmt.Sprintf("no log/artifact activity and CPU %.1f%% for %ds (>= check interval %ds)", cpuPct, idleSec, entry.CheckIntervalSec)
		}
		return LivenessStalled, []string{msg}
	}
	return LivenessActive, []string{"process/artifacts showing normal liveness"}
}

func inspectProcess(pid int) (alive bool, zombie bool, cpuPct float64, rssKB int64) {
	if pid <= 0 {
		return true, false, 0, 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "ps", "-o", "state=,pcpu=,rss=", "-p", strconv.Itoa(pid)).Output()
	if err == nil {
		fields := strings.Fields(strings.TrimSpace(string(out)))
		if len(fields) >= 3 {
			st := strings.ToUpper(fields[0])
			cpu, _ := strconv.ParseFloat(fields[1], 64)
			rss, _ := strconv.ParseInt(fields[2], 10, 64)
			return true, strings.HasPrefix(st, "Z"), cpu, rss
		}
	}

	if killErr := syscall.Kill(pid, 0); killErr == nil || errors.Is(killErr, syscall.EPERM) {
		return true, false, 0, 0
	}
	return false, false, 0, 0
}

func statFileMeta(path string) (int64, string, bool) {
	if strings.TrimSpace(path) == "" {
		return 0, "", false
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, "", false
	}
	return info.Size(), info.ModTime().UTC().Format(time.RFC3339Nano), true
}

func secondsBetween(startRFC3339 string, now time.Time) int {
	ts, err := time.Parse(time.RFC3339, startRFC3339)
	if err != nil || !now.After(ts) {
		return 0
	}
	return int(math.Floor(now.Sub(ts).Seconds()))
}

func watchKey(watchID string) string {
	return "watch_" + watchID
}

func (t *Tracker) withWatchLock(watchID string, fn func() error) error {
	return t.locks.WithLock(watchKey(watchID), fn)
}

func (t *Tracker) saveWatch(entry WatchEntry) error {
	return t.withWatchLock(entry.WatchID, func() error {
		return t.saveWatchUnlocked(entry)
	})
}

func (t *Tracker) saveWatchUnlocked(entry WatchEntry) error {
	target := filepath.Join(t.stateDir, watchKey(entry.WatchID)+".json")
	return store.WriteJSONAtomic(target, entry, 0o600)
}

func (t *Tracker) loadWatchUnlocked(watchID string) (WatchEntry, error) {
	entry, err := store.ReadJSONRoot[WatchEntry](t.stateDir, watchKey(watchID)+".json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return WatchEntry{}, fmt.Errorf("watch entry %q not found: %w", watchID, err)
		}
		return WatchEntry{}, fmt.Errorf("load watch entry %q: %w", watchID, err)
	}
	return entry, nil
}
