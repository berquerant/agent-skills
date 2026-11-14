package pace_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/pace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaceSessionLifecycle(t *testing.T) {
	tmp := t.TempDir()
	tr, err := pace.NewTracker(tmp)
	require.NoError(t, err)

	t0 := time.Date(2026, 11, 15, 5, 0, 0, 0, time.UTC)

	startOut, err := tr.Start(pace.StartInput{
		PaceID:   "feat-work",
		Profile:  pace.ProfileQuick, // 120s / 10 calls
		Mode:     pace.ExecModeStandard,
		TaskNote: "Implement pace-guard package",
	}, t0)
	require.NoError(t, err)
	assert.Equal(t, 120, startOut.Session.TimeLimitSec)
	assert.Equal(t, 10, startOut.Session.CallLimit)
	assert.Contains(t, startOut.ProposalTemplate, "120 seconds")

	// Parallel read batch counts as 1 non-edit step
	ev1, err := tr.Tick(pace.TickInput{
		PaceID: "feat-work",
		Kind:   pace.TickKindParallelBatch,
		Note:   "Read 5 files in parallel",
	}, t0.Add(30*time.Second))
	require.NoError(t, err)
	assert.Equal(t, pace.PaceStatusOK, ev1.Status)
	assert.Equal(t, 1, ev1.NonEditCalls)
	assert.Equal(t, 9, ev1.RemainingCalls)
	assert.Equal(t, 90, ev1.RemainingSec)

	// Exceeding time threshold triggers EXCEEDED and checkpoint report
	ev2, err := tr.Status(pace.StatusInput{PaceID: "feat-work"}, t0.Add(125*time.Second))
	require.NoError(t, err)
	assert.Equal(t, pace.PaceStatusExceeded, ev2.Status)
	assert.Contains(t, ev2.CheckpointReport, "### ⏸️ Paced Execution Checkpoint")

	// Edit step resets non-edit counter and elapsed timer
	ev3, err := tr.Tick(pace.TickInput{
		PaceID:   "feat-work",
		Kind:     pace.TickKindEdit,
		FilePath: "mcp/internal/pace/tracker.go",
		Note:     "Updated tracker logic",
	}, t0.Add(130*time.Second))
	require.NoError(t, err)
	assert.Equal(t, pace.PaceStatusOK, ev3.Status)
	assert.Equal(t, 0, ev3.NonEditCalls)
	assert.Equal(t, 0, ev3.ElapsedSec)
	assert.Equal(t, "mcp/internal/pace/tracker.go", ev3.LastEditedFile)

	// 3 consecutive cyclic failures trigger EXCEEDED even before time/call limits
	t1 := t0.Add(140 * time.Second)
	for i := range 3 {
		ev, tickErr := tr.Tick(pace.TickInput{
			PaceID:      "feat-work",
			Kind:        pace.TickKindNonEdit,
			CyclicError: true,
			Note:        "Repeated test failure",
		}, t1.Add(time.Duration(i)*time.Second))
		require.NoError(t, tickErr)
		if i < 2 {
			assert.Equal(t, pace.PaceStatusOK, ev.Status)
		} else {
			assert.Equal(t, pace.PaceStatusExceeded, ev.Status)
			assert.Contains(t, ev.Reasons[0], "cyclic failure loop")
		}
	}

	// Reset clears counters and updates profile
	evReset, err := tr.Reset(pace.ResetInput{
		PaceID:  "feat-work",
		Profile: pace.ProfileDeep,
		Mode:    pace.ExecModeHeavy,
		Note:    "User approved deeper investigation",
	}, t0.Add(200*time.Second))
	require.NoError(t, err)
	assert.Equal(t, pace.PaceStatusOK, evReset.Status)
	assert.Equal(t, 300, evReset.TimeLimitSec)
	assert.Equal(t, 30, evReset.CallLimit)
}

func TestPaceWatchAndGuards(t *testing.T) {
	tmp := t.TempDir()
	tr, err := pace.NewTracker(tmp)
	require.NoError(t, err)

	t0 := time.Date(2026, 11, 15, 5, 30, 0, 0, time.UTC)
	logPath := filepath.Join(tmp, "build.log")
	require.NoError(t, os.WriteFile(logPath, []byte("line 1\n"), 0o600))

	wOut, err := tr.Watch(pace.WatchInput{
		WatchID: "build-watch",
		Command: "make build",
		Mode:    pace.ExecModeCritical, // 60s check interval
		LogPath: logPath,
	}, t0)
	require.NoError(t, err)
	assert.Equal(t, 60, wOut.Watch.CheckIntervalSec)

	// Before check interval: ACTIVE even with no new log lines
	live1, err := tr.CheckLiveness(pace.LivenessInput{WatchID: "build-watch"}, t0.Add(20*time.Second))
	require.NoError(t, err)
	assert.Equal(t, pace.LivenessActive, live1.State)

	// After check interval (65s) with no log change: STALLED
	live2, err := tr.CheckLiveness(pace.LivenessInput{WatchID: "build-watch"}, t0.Add(65*time.Second))
	require.NoError(t, err)
	assert.Equal(t, pace.LivenessStalled, live2.State)
	assert.Contains(t, live2.StallReport, "### ⚠️ Command Execution Stall Report")

	// Appending to log restores ACTIVE state
	require.NoError(t, os.WriteFile(logPath, []byte("line 1\nline 2\n"), 0o600))
	live3, err := tr.CheckLiveness(pace.LivenessInput{WatchID: "build-watch"}, t0.Add(70*time.Second))
	require.NoError(t, err)
	assert.Equal(t, pace.LivenessActive, live3.State)
	assert.True(t, live3.LogAdvanced)

	// Verify guard rejects sensitive file path and embedded secrets
	_, err = tr.Tick(pace.TickInput{
		Kind:     pace.TickKindEdit,
		FilePath: ".env.production",
	}, t0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "access to sensitive path")

	_, err = tr.Start(pace.StartInput{
		TaskNote: "token=" + "gh" + "p_1234567890abcdefghijklmnopqrstuv",
	}, t0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sensitive secret pattern detected")
}

func TestPaceReportSanitizationAndLogMod(t *testing.T) {
	tmp := t.TempDir()
	tr, err := pace.NewTracker(tmp)
	require.NoError(t, err)

	t0 := time.Date(2026, 11, 15, 6, 0, 0, 0, time.UTC)

	// Note with embedded newlines should be sanitized to single line in report
	_, err = tr.Start(pace.StartInput{
		PaceID:       "sanitize-test",
		TimeLimitSec: 10,
		TaskNote:     "Heading injection\n## Injected Section\n- **Status**: BAD",
	}, t0)
	require.NoError(t, err)

	ev, err := tr.Status(pace.StatusInput{PaceID: "sanitize-test"}, t0.Add(15*time.Second))
	require.NoError(t, err)
	assert.Equal(t, pace.PaceStatusExceeded, ev.Status)
	assert.NotContains(t, ev.CheckpointReport, "\n## Injected Section")
	assert.Contains(t, ev.CheckpointReport, "Heading injection ## Injected Section - **Status**: BAD")

	// Empty log file modification timestamp change is detected even if size is 0
	emptyLog := filepath.Join(tmp, "empty.log")
	require.NoError(t, os.WriteFile(emptyLog, []byte(""), 0o600))

	wOut, err := tr.Watch(pace.WatchInput{
		WatchID: "empty-log-watch",
		LogPath: emptyLog,
		Mode:    pace.ExecModeCritical,
	}, t0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), wOut.Watch.LastLogSize)

	// Touch empty file without changing size
	require.NoError(t, os.Chtimes(emptyLog, t0.Add(10*time.Second), t0.Add(10*time.Second)))
	live, err := tr.CheckLiveness(pace.LivenessInput{WatchID: "empty-log-watch"}, t0.Add(15*time.Second))
	require.NoError(t, err)
	assert.True(t, live.LogAdvanced)
	assert.Equal(t, pace.LivenessActive, live.State)
}
