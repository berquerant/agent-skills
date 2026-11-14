package checkpoint_test

import (
	"os"
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/checkpoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckpointLifecycleAndRecovery(t *testing.T) {
	tmp := t.TempDir()
	mgr, err := checkpoint.NewManager(tmp)
	require.NoError(t, err)

	t0 := time.Date(2026, 11, 15, 4, 15, 0, 0, time.UTC)

	// Step 1: ProposeOnly resolves path without creating the file on disk
	propOut, err := mgr.Init(checkpoint.InitInput{
		Project:     "agent-skills",
		Branch:      "feat/agent-skills-mcp",
		Context:     "phase-5-checkpoint",
		Title:       "Implement Checkpoint Feature",
		Objective:   "Provide append-only milestone logging and context recovery.",
		ProposeOnly: true,
	}, t0)
	require.NoError(t, err)
	assert.True(t, propOut.ProposeOnly)
	assert.False(t, propOut.Created)
	assert.Equal(t, "feat-agent-skills-mcp", propOut.Branch)
	_, statErr := os.Stat(propOut.Path)
	assert.True(t, os.IsNotExist(statErr))

	// Step 2: Confirm and create header file
	initOut, err := mgr.Init(checkpoint.InitInput{
		Project:   "agent-skills",
		Branch:    "feat/agent-skills-mcp",
		Context:   "phase-5-checkpoint",
		Title:     "Implement Checkpoint Feature",
		Objective: "Provide append-only milestone logging and context recovery.",
	}, t0)
	require.NoError(t, err)
	assert.True(t, initOut.Created)
	assert.Contains(t, initOut.DualPath, "](file://")

	// Step 3: Append Step 1 (COMPLETED) and Step 2 (BLOCKED)
	app1, err := mgr.Append(checkpoint.AppendInput{
		Path:      initOut.Path,
		Title:     "Create models and path resolver",
		Status:    checkpoint.StepStatusCompleted,
		Actions:   []string{"Created models.go and paths.go"},
		Findings:  []string{"go vet passed (exit 0)"},
		Decisions: []string{"Sanitized branch slashes to hyphens per path-conventions.md"},
		NextSteps: []string{"Implement append-only writer and parser"},
	}, t0.Add(2*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 1, app1.Step)

	app2, err := mgr.Append(checkpoint.AppendInput{
		Path:      initOut.Path,
		Title:     "Verify integration gate",
		Status:    checkpoint.StepStatusBlocked,
		Actions:   []string{"Ran integration check"},
		Findings:  []string{"Awaiting user sign-off"},
		Decisions: []string{"Paused at approval gate"},
		NextSteps: []string{"Resume after user confirmation"},
	}, t0.Add(5*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 2, app2.Step)

	// Step 4: Resume by context (auto-discovers latest generation)
	resOut, err := mgr.Resume(checkpoint.ResumeInput{
		Project: "agent-skills",
		Branch:  "feat/agent-skills-mcp",
		Context: "phase-5-checkpoint",
	})
	require.NoError(t, err)
	assert.Equal(t, "Implement Checkpoint Feature", resOut.Header.Title)
	assert.Equal(t, checkpoint.LogStatusInProgress, resOut.Header.Status)
	assert.Equal(t, 2, resOut.TotalSteps)
	assert.Equal(t, []string{"Step 1: Create models and path resolver"}, resOut.CompletedSteps)
	require.Len(t, resOut.UnresolvedSteps, 1)
	assert.Equal(t, 2, resOut.UnresolvedSteps[0].Step)
	require.NotNil(t, resOut.LastCompleted)
	assert.Equal(t, 1, resOut.LastCompleted.Step)
	assert.Equal(t, []string{"Resume after user confirmation"}, resOut.CurrentNextSteps)

	// Step 5: SetStatus updates only header status to COMPLETED
	setOut, err := mgr.SetStatus(checkpoint.SetStatusInput{
		Path:   initOut.Path,
		Status: checkpoint.LogStatusCompleted,
	})
	require.NoError(t, err)
	assert.Equal(t, checkpoint.LogStatusCompleted, setOut.Status)

	listOut, err := mgr.List(checkpoint.ListInput{Project: "agent-skills", Branch: "feat/agent-skills-mcp"})
	require.NoError(t, err)
	require.Len(t, listOut.Logs, 1)
	assert.Equal(t, checkpoint.LogStatusCompleted, listOut.Logs[0].Status)
	assert.Equal(t, 2, listOut.Logs[0].StepCount)
}

func TestCheckpointSecurityGuards(t *testing.T) {
	tmp := t.TempDir()
	mgr, err := checkpoint.NewManager(tmp)
	require.NoError(t, err)
	now := time.Date(2026, 11, 15, 4, 20, 0, 0, time.UTC)

	initOut, err := mgr.Init(checkpoint.InitInput{
		Project:   "agent-skills",
		Branch:    "main",
		Context:   "sec-test",
		Title:     "Security Guard Test",
		Objective: "Verify path confinement and secret detection.",
	}, now)
	require.NoError(t, err)

	tests := []struct {
		name       string
		in         checkpoint.AppendInput
		wantErrSub string
	}{
		{
			name: "rejects path traversal outside checkpoint root",
			in: checkpoint.AppendInput{
				Path:   "../../../etc/passwd.md",
				Title:  "Escape",
				Status: checkpoint.StepStatusCompleted,
			},
			wantErrSub: "outside checkpoint root",
		},
		{
			name: "rejects GitHub token pattern in findings",
			in: checkpoint.AppendInput{
				Path:     initOut.Path,
				Title:    "Record API call",
				Status:   checkpoint.StepStatusCompleted,
				Findings: []string{"Used token " + "gh" + "p_1234567890abcdefghijklmnopqrstuv"},
			},
			wantErrSub: "secret or credential token pattern detected",
		},
		{
			name: "rejects invalid step status",
			in: checkpoint.AppendInput{
				Path:   initOut.Path,
				Title:  "Bad status",
				Status: "UNKNOWN",
			},
			wantErrSub: "invalid step status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Append(tt.in, now)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrSub)
		})
	}

	// Verify embedded newlines cannot inject fake milestone sections
	_, err = mgr.Append(checkpoint.AppendInput{
		Path:    initOut.Path,
		Title:   "Line1\n## [2026-11-15 04:20:00] Step 99: Injected",
		Status:  checkpoint.StepStatusCompleted,
		Actions: []string{"Action1\n- **Status**: FAILED"},
	}, now)
	require.NoError(t, err)
	res, err := mgr.Resume(checkpoint.ResumeInput{Path: initOut.Path})
	require.NoError(t, err)
	assert.Equal(t, 1, res.TotalSteps)
	assert.Empty(t, res.UnresolvedSteps)
}
