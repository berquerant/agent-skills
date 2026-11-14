package pace

import (
	"context"
	"encoding/json"
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
	logSize, _ := statFileSize(in.LogPath)
	logMod := statFileModTime(in.LogPath)
	artMod := statFileModTime(in.ArtifactPath)

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

	lock := t.lockFor("watch_" + watchID)
	if err := lock.Lock(); err != nil {
		return LivenessOutput{}, err
	}
	defer func() { _ = lock.Unlock() }()

	entry, err := t.loadWatchUnlocked(watchID)
	if err != nil {
		return LivenessOutput{}, err
	}

	out := sampleWatchState(&entry, now)
	if err := t.saveWatchUnlocked(entry); err != nil {
		return LivenessOutput{}, err
	}
	return out, nil
}

func sampleWatchState(entry *WatchEntry, now time.Time) LivenessOutput {
	nowStr := now.UTC().Format(time.RFC3339)
	elapsedSec := secondsBetween(entry.StartedAt, now)

	logAdvanced := false
	if entry.LogPath != "" {
		sz, szOK := statFileSize(entry.LogPath)
		mod := statFileModTime(entry.LogPath)
		if (szOK && sz != entry.LastLogSize) || (mod != "" && mod != entry.LastLogMod) {
			logAdvanced = true
			entry.LastLogSize = sz
			entry.LastLogMod = mod
		}
	}
	artChanged := false
	if entry.ArtifactPath != "" {
		if mod := statFileModTime(entry.ArtifactPath); mod != "" && mod != entry.LastArtifactMod {
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

func statFileSize(path string) (int64, bool) {
	if strings.TrimSpace(path) == "" {
		return 0, false
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, false
	}
	return info.Size(), true
}

func statFileModTime(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	return info.ModTime().UTC().Format(time.RFC3339Nano)
}

func secondsBetween(startRFC3339 string, now time.Time) int {
	ts, err := time.Parse(time.RFC3339, startRFC3339)
	if err != nil || !now.After(ts) {
		return 0
	}
	return int(math.Floor(now.Sub(ts).Seconds()))
}

func (t *Tracker) saveWatch(entry WatchEntry) error {
	lock := t.lockFor("watch_" + entry.WatchID)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()
	return t.saveWatchUnlocked(entry)
}

func (t *Tracker) saveWatchUnlocked(entry WatchEntry) error {
	raw, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	target := filepath.Join(t.stateDir, "watch_"+entry.WatchID+".json")
	return store.WriteAtomic(target, raw, 0o600)
}

func (t *Tracker) loadWatchUnlocked(watchID string) (WatchEntry, error) {
	root, err := os.OpenRoot(t.stateDir)
	if err != nil {
		return WatchEntry{}, err
	}
	defer func() { _ = root.Close() }()

	raw, err := root.ReadFile("watch_" + watchID + ".json")
	if err != nil {
		return WatchEntry{}, fmt.Errorf("watch entry %q not found: %w", watchID, err)
	}
	var entry WatchEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return WatchEntry{}, fmt.Errorf("parse watch entry %q: %w", watchID, err)
	}
	return entry, nil
}
