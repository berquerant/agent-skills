package pace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/berquerant/agent-skills/mcp/internal/store"
)

const (
	defaultPaceID       = "default"
	cyclicFailureLimit  = 3
	paceLockTimeout     = 10 * time.Second
)

var safePaceIDRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`)

// Tracker manages edit-pacing sessions and process watch entries under stateDir.
type Tracker struct {
	stateDir string
}

// NewTracker creates a Tracker rooted at stateDir.
func NewTracker(stateDir string) (*Tracker, error) {
	if err := guard.CheckPathAllowed(stateDir); err != nil {
		return nil, err
	}
	absDir, err := filepath.Abs(stateDir)
	if err != nil {
		return nil, fmt.Errorf("resolve pace state dir: %w", err)
	}
	if err := os.MkdirAll(absDir, 0o700); err != nil {
		return nil, fmt.Errorf("create pace state dir: %w", err)
	}
	return &Tracker{stateDir: absDir}, nil
}

func (t *Tracker) lockFor(name string) *store.FileLock {
	return store.NewFileLock(filepath.Join(t.stateDir, name+".lock"), paceLockTimeout)
}

func normalizePaceID(id string) (string, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return defaultPaceID, nil
	}
	if !safePaceIDRe.MatchString(trimmed) {
		return "", fmt.Errorf("invalid identifier %q", id)
	}
	return trimmed, nil
}

// ResolvePreset returns (Profile, timeLimitSec, callLimit, error).
func ResolvePreset(p Profile, customTimeSec, customCalls int) (Profile, int, int, error) {
	prof := Profile(strings.ToLower(strings.TrimSpace(string(p))))
	if prof == "" {
		prof = ProfileStandard
	}
	var timeSec, calls int
	switch prof {
	case ProfileQuick:
		timeSec, calls = 120, 10
	case ProfileStandard:
		timeSec, calls = 180, 20
	case ProfileDeep:
		timeSec, calls = 300, 30
	default:
		return "", 0, 0, fmt.Errorf("unsupported profile %q (expected quick, standard, or deep)", p)
	}
	if customTimeSec > 0 {
		timeSec = customTimeSec
	}
	if customCalls > 0 {
		calls = customCalls
	}
	return prof, timeSec, calls, nil
}

// ResolveExecMode returns (ExecMode, checkIntervalSec, error).
func ResolveExecMode(m ExecMode) (ExecMode, int, error) {
	mode := ExecMode(strings.ToLower(strings.TrimSpace(string(m))))
	if mode == "" {
		mode = ExecModeStandard
	}
	switch mode {
	case ExecModeStandard:
		return mode, 180, nil
	case ExecModeHeavy:
		return mode, 300, nil
	case ExecModeCritical:
		return mode, 60, nil
	case ExecModeUnmonitored:
		return mode, 3600, nil
	default:
		return "", 0, fmt.Errorf("unsupported mode %q (expected standard, heavy, critical, or unmonitored)", m)
	}
}

// Start initializes and persists a pacing session.
func (t *Tracker) Start(in StartInput, now time.Time) (StartOutput, error) {
	if err := guard.CheckNoSecrets(in.PaceID, in.TaskNote); err != nil {
		return StartOutput{}, err
	}
	paceID, err := normalizePaceID(in.PaceID)
	if err != nil {
		return StartOutput{}, err
	}
	prof, timeSec, calls, err := ResolvePreset(in.Profile, in.TimeLimitSec, in.CallLimit)
	if err != nil {
		return StartOutput{}, err
	}
	mode, intervalSec, err := ResolveExecMode(in.Mode)
	if err != nil {
		return StartOutput{}, err
	}

	nowStr := now.UTC().Format(time.RFC3339)
	sess := Session{
		PaceID:           paceID,
		Profile:          prof,
		Mode:             mode,
		TimeLimitSec:     timeSec,
		CallLimit:        calls,
		CheckIntervalSec: intervalSec,
		StartedAt:        nowStr,
		LastEditAt:       nowStr,
		LastNote:         strings.TrimSpace(in.TaskNote),
	}
	if err := t.saveSession(sess); err != nil {
		return StartOutput{}, err
	}

	proposal := fmt.Sprintf(
		"For this task, I propose an edit-pacing threshold of %d seconds (%.1f minutes) or %d non-edit tool calls (profile: %s). Long-running commands will run in `%s` mode (check interval: %d seconds).",
		timeSec, float64(timeSec)/60.0, calls, prof, mode, intervalSec,
	)
	return StartOutput{
		OK:               true,
		Session:          sess,
		ProposalTemplate: proposal,
	}, nil
}

// Tick records an edit or non-edit step and evaluates pacing thresholds.
func (t *Tracker) Tick(in TickInput, now time.Time) (PaceEvaluation, error) {
	if err := guard.CheckPathAllowed(in.FilePath); err != nil {
		return PaceEvaluation{}, err
	}
	if err := guard.CheckNoSecrets(in.PaceID, in.FilePath, in.Note); err != nil {
		return PaceEvaluation{}, err
	}
	paceID, err := normalizePaceID(in.PaceID)
	if err != nil {
		return PaceEvaluation{}, err
	}

	lock := t.lockFor("session_" + paceID)
	if err := lock.Lock(); err != nil {
		return PaceEvaluation{}, err
	}
	defer func() { _ = lock.Unlock() }()

	sess, err := t.loadSessionUnlocked(paceID)
	if err != nil {
		return PaceEvaluation{}, err
	}
	if err := applyStepToSession(&sess, in, now); err != nil {
		return PaceEvaluation{}, err
	}
	if err := t.saveSessionUnlocked(sess); err != nil {
		return PaceEvaluation{}, err
	}
	return evaluateSession(sess, now), nil
}

func applyStepToSession(sess *Session, in TickInput, now time.Time) error {
	nowStr := now.UTC().Format(time.RFC3339)
	switch in.Kind {
	case TickKindEdit:
		sess.NonEditCalls = 0
		sess.ConsecutiveFailures = 0
		sess.TotalEdits++
		sess.LastEditAt = nowStr
		if strings.TrimSpace(in.FilePath) != "" {
			sess.LastEditedFile = strings.TrimSpace(in.FilePath)
		}
	case TickKindNonEdit, TickKindParallelBatch:
		sess.NonEditCalls++
		if in.CyclicError {
			sess.ConsecutiveFailures++
		} else {
			sess.ConsecutiveFailures = 0
		}
	default:
		return fmt.Errorf("invalid tick kind %q (expected edit, non_edit, or parallel_batch)", in.Kind)
	}
	if strings.TrimSpace(in.Note) != "" {
		sess.LastNote = strings.TrimSpace(in.Note)
	}
	return nil
}

// Status evaluates the current session against real elapsed time without modifying counters.
func (t *Tracker) Status(in StatusInput, now time.Time) (PaceEvaluation, error) {
	paceID, err := normalizePaceID(in.PaceID)
	if err != nil {
		return PaceEvaluation{}, err
	}
	lock := t.lockFor("session_" + paceID)
	if err := lock.RLock(); err != nil {
		return PaceEvaluation{}, err
	}
	defer func() { _ = lock.Unlock() }()

	sess, err := t.loadSessionUnlocked(paceID)
	if err != nil {
		return PaceEvaluation{}, err
	}
	return evaluateSession(sess, now), nil
}

// Reset resets the session timer and counters after user direction, optionally updating thresholds.
func (t *Tracker) Reset(in ResetInput, now time.Time) (PaceEvaluation, error) {
	if err := guard.CheckNoSecrets(in.PaceID, in.Note); err != nil {
		return PaceEvaluation{}, err
	}
	paceID, err := normalizePaceID(in.PaceID)
	if err != nil {
		return PaceEvaluation{}, err
	}

	lock := t.lockFor("session_" + paceID)
	if err := lock.Lock(); err != nil {
		return PaceEvaluation{}, err
	}
	defer func() { _ = lock.Unlock() }()

	sess, err := t.loadSessionUnlocked(paceID)
	if err != nil {
		return PaceEvaluation{}, err
	}
	if err := updateSessionThresholds(&sess, in); err != nil {
		return PaceEvaluation{}, err
	}
	sess.NonEditCalls = 0
	sess.ConsecutiveFailures = 0
	sess.LastEditAt = now.UTC().Format(time.RFC3339)
	if strings.TrimSpace(in.Note) != "" {
		sess.LastNote = strings.TrimSpace(in.Note)
	}
	if err := t.saveSessionUnlocked(sess); err != nil {
		return PaceEvaluation{}, err
	}
	return evaluateSession(sess, now), nil
}

func updateSessionThresholds(sess *Session, in ResetInput) error {
	if in.Profile != "" || in.TimeLimitSec > 0 || in.CallLimit > 0 {
		p := in.Profile
		if p == "" {
			p = sess.Profile
		}
		prof, timeSec, calls, err := ResolvePreset(p, in.TimeLimitSec, in.CallLimit)
		if err != nil {
			return err
		}
		sess.Profile = prof
		if in.Profile != "" || in.TimeLimitSec > 0 {
			sess.TimeLimitSec = timeSec
		}
		if in.Profile != "" || in.CallLimit > 0 {
			sess.CallLimit = calls
		}
	}
	if in.Mode != "" {
		mode, intervalSec, err := ResolveExecMode(in.Mode)
		if err != nil {
			return err
		}
		sess.Mode = mode
		sess.CheckIntervalSec = intervalSec
	}
	return nil
}

func (t *Tracker) saveSession(sess Session) error {
	lock := t.lockFor("session_" + sess.PaceID)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()
	return t.saveSessionUnlocked(sess)
}

func (t *Tracker) saveSessionUnlocked(sess Session) error {
	raw, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	target := filepath.Join(t.stateDir, "session_"+sess.PaceID+".json")
	return store.WriteAtomic(target, raw, 0o600)
}

func (t *Tracker) loadSessionUnlocked(paceID string) (Session, error) {
	root, err := os.OpenRoot(t.stateDir)
	if err != nil {
		return Session{}, err
	}
	defer func() { _ = root.Close() }()

	raw, err := root.ReadFile("session_" + paceID + ".json")
	if err != nil {
		return Session{}, fmt.Errorf("pacing session %q not found (call pace_start first): %w", paceID, err)
	}
	var sess Session
	if err := json.Unmarshal(raw, &sess); err != nil {
		return Session{}, fmt.Errorf("parse pacing session %q: %w", paceID, err)
	}
	return sess, nil
}
