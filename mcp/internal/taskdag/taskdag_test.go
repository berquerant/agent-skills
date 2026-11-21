package taskdag_test

import (
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/taskdag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAndInitTasks(t *testing.T) {
	now := "2026-11-15T04:00:00Z"

	tests := []struct {
		name        string
		tasks       []taskdag.Task
		wantErrSub  string
		wantReady   []string
		wantBlocked []string
	}{
		{
			name: "valid diamond DAG initializes roots as READY and dependents as BLOCKED",
			tasks: []taskdag.Task{
				{ID: "setup-a"},
				{ID: "setup-b"},
				{ID: "merge", DependsOn: []string{"setup-a", "setup-b"}},
			},
			wantReady:   []string{"setup-a", "setup-b"},
			wantBlocked: []string{"merge"},
		},
		{
			name:       "empty task list is rejected",
			tasks:      nil,
			wantErrSub: "at least one task is required",
		},
		{
			name: "duplicate task ID is rejected",
			tasks: []taskdag.Task{
				{ID: "t1"},
				{ID: "t1"},
			},
			wantErrSub: "duplicate task id",
		},
		{
			name: "unknown dependency is rejected",
			tasks: []taskdag.Task{
				{ID: "t1", DependsOn: []string{"missing"}},
			},
			wantErrSub: "depends on unknown task",
		},
		{
			name: "self dependency is rejected",
			tasks: []taskdag.Task{
				{ID: "t1", DependsOn: []string{"t1"}},
			},
			wantErrSub: "cannot depend on itself",
		},
		{
			name: "3-node cycle is detected with cycle path",
			tasks: []taskdag.Task{
				{ID: "a", DependsOn: []string{"c"}},
				{ID: "b", DependsOn: []string{"a"}},
				{ID: "c", DependsOn: []string{"b"}},
			},
			wantErrSub: "circular dependency detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := taskdag.ValidateAndInitTasks(tt.tasks, now)
			if tt.wantErrSub != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrSub)
				return
			}
			require.NoError(t, err)
			var ready, blocked []string
			for _, task := range got {
				switch task.State {
				case taskdag.StateReady:
					ready = append(ready, task.ID)
				case taskdag.StateBlocked:
					blocked = append(blocked, task.ID)
				}
			}
			assert.Equal(t, tt.wantReady, ready)
			assert.Equal(t, tt.wantBlocked, blocked)
		})
	}
}

func TestPrioritySelectionAndCooperativeInterleaving(t *testing.T) {
	baseTime := time.Date(2026, 11, 15, 4, 0, 0, 0, time.UTC)
	nowStr := baseTime.Format(time.RFC3339)

	tasks, err := taskdag.ValidateAndInitTasks([]taskdag.Task{
		{ID: "branch-a-root", WaitProfile: taskdag.WaitProfileLocal, EstimateSec: 20},
		{ID: "branch-a-ci", DependsOn: []string{"branch-a-root"}, WaitProfile: taskdag.WaitProfileAsync, EstimateSec: 60},
		{ID: "branch-a-verify", DependsOn: []string{"branch-a-ci"}, WaitProfile: taskdag.WaitProfileLocal, EstimateSec: 10},
		{ID: "branch-b-solo", WaitProfile: taskdag.WaitProfileLocal, EstimateSec: 5},
	}, nowStr)
	require.NoError(t, err)

	dag := &taskdag.DAG{DAGID: "dag-test", Tasks: tasks}

	// Step 1: branch-a-root unblocks 2 downstream tasks vs branch-b-solo unblocking 0
	next1, _ := taskdag.SelectNext(dag, baseTime)
	require.NotNil(t, next1.Task)
	assert.Equal(t, "branch-a-root", next1.Task.ID)
	assert.Equal(t, 2, next1.UnblocksCount)

	// Step 2: Run and complete branch-a-root -> promotes branch-a-ci to READY
	_, err = taskdag.ApplyTransition(dag, taskdag.TransitionInput{
		TaskID: "branch-a-root",
		To:     taskdag.StateRunning,
	}, baseTime)
	require.NoError(t, err)

	res, err := taskdag.ApplyTransition(dag, taskdag.TransitionInput{
		TaskID: "branch-a-root",
		To:     taskdag.StateCompleted,
	}, baseTime)
	require.NoError(t, err)
	assert.Equal(t, []string{"branch-a-ci"}, res.PromotedTasks)

	// Step 3: Start branch-a-ci and transition to WAITING (poll_after_sec=45)
	_, err = taskdag.ApplyTransition(dag, taskdag.TransitionInput{
		TaskID: "branch-a-ci",
		To:     taskdag.StateRunning,
	}, baseTime)
	require.NoError(t, err)

	_, err = taskdag.ApplyTransition(dag, taskdag.TransitionInput{
		TaskID:       "branch-a-ci",
		To:           taskdag.StateWaiting,
		PollAfterSec: 45,
		Note:         "waiting for remote CI",
	}, baseTime)
	require.NoError(t, err)

	// Step 4: Cooperative switch picks independent READY task branch-b-solo while branch-a-ci waits
	next2, _ := taskdag.SelectNext(dag, baseTime.Add(5*time.Second))
	require.NotNil(t, next2.Task)
	assert.Equal(t, "branch-b-solo", next2.Task.ID)

	// Step 5: Anti-polling guard rejects premature WAITING -> READY before 45s
	_, err = taskdag.ApplyTransition(dag, taskdag.TransitionInput{
		TaskID: "branch-a-ci",
		To:     taskdag.StateReady,
	}, baseTime.Add(10*time.Second))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "anti-polling cooldown active")

	// Step 6: Start branch-b-solo (both branch-b-solo and branch-a-ci have empty Resource "")
	// After 50s (>= 45s), SelectNext auto-awakens branch-a-ci even while branch-b-solo is RUNNING
	_, err = taskdag.ApplyTransition(dag, taskdag.TransitionInput{
		TaskID: "branch-b-solo",
		To:     taskdag.StateRunning,
	}, baseTime.Add(15*time.Second))
	require.NoError(t, err)

	next3, mutated := taskdag.SelectNext(dag, baseTime.Add(50*time.Second))
	assert.True(t, mutated)
	require.NotNil(t, next3.Task)
	assert.Equal(t, "branch-a-ci", next3.Task.ID)
}

func TestResourceCollisionAndSubtreeHalting(t *testing.T) {
	baseTime := time.Date(2026, 11, 15, 4, 0, 0, 0, time.UTC)
	nowStr := baseTime.Format(time.RFC3339)

	tasks, err := taskdag.ValidateAndInitTasks([]taskdag.Task{
		{ID: "t1", Resource: "worktree-main"},
		{ID: "t2", Resource: "worktree-main"},
		{ID: "t1-child", DependsOn: []string{"t1"}},
		{ID: "t3-independent", Resource: "worktree-other"},
	}, nowStr)
	require.NoError(t, err)
	dag := &taskdag.DAG{DAGID: "dag-res", Tasks: tasks}

	// Start t1 on worktree-main
	_, err = taskdag.ApplyTransition(dag, taskdag.TransitionInput{
		TaskID: "t1",
		To:     taskdag.StateRunning,
	}, baseTime)
	require.NoError(t, err)

	// Starting t2 on same resource worktree-main must fail
	_, err = taskdag.ApplyTransition(dag, taskdag.TransitionInput{
		TaskID: "t2",
		To:     taskdag.StateRunning,
	}, baseTime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resource collision")

	// Halting t1 reports t1-child in haltedDescendants, while t2 and t3-independent remain runnable
	haltRes, err := taskdag.ApplyTransition(dag, taskdag.TransitionInput{
		TaskID: "t1",
		To:     taskdag.StateHalted,
		Note:   "compile error",
	}, baseTime)
	require.NoError(t, err)
	assert.Equal(t, []string{"t1-child"}, haltRes.HaltedDescendants)

	// Next task picks t2 (since worktree-main lock was released when t1 halted)
	next, _ := taskdag.SelectNext(dag, baseTime)
	require.NotNil(t, next.Task)
	assert.Equal(t, "t2", next.Task.ID)

	mermaid := taskdag.RenderMermaid(dag)
	assert.Contains(t, mermaid, "flowchart TD")
	assert.Contains(t, mermaid, "[HALTED]")
}

func TestStoragePersistenceAndJournal(t *testing.T) {
	tmp := t.TempDir()
	st, err := taskdag.NewStorage(tmp)
	require.NoError(t, err)

	now := time.Date(2026, 11, 15, 4, 0, 0, 0, time.UTC)
	nowStr := now.Format(time.RFC3339)
	tasks, err := taskdag.ValidateAndInitTasks([]taskdag.Task{
		{ID: "step-1", Title: "Build"},
		{ID: "step-2", Title: "Test", DependsOn: []string{"step-1"}},
	}, nowStr)
	require.NoError(t, err)

	dag := taskdag.DAG{
		DAGID:     "dag-persist",
		Name:      "CI Pipeline",
		CreatedAt: nowStr,
		UpdatedAt: nowStr,
		Tasks:     tasks,
	}
	require.NoError(t, st.Create(dag))
	require.ErrorContains(t, st.Create(dag), "already exists")

	_, err = st.Update("dag-persist", func(d *taskdag.DAG) (*taskdag.JournalEntry, error) {
		res, tErr := taskdag.ApplyTransition(d, taskdag.TransitionInput{
			TaskID: "step-1",
			To:     taskdag.StateRunning,
		}, now)
		return &res.Journal, tErr
	})
	require.NoError(t, err)

	loaded, err := st.Load("dag-persist")
	require.NoError(t, err)
	assert.Equal(t, taskdag.StateRunning, loaded.Tasks[0].State)

	list, err := st.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "dag-persist", list[0].DAGID)

	deleted, err := st.Delete("dag-persist")
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestComputeCriticalPathAndIndependentBranches(t *testing.T) {
	tasks := []taskdag.Task{
		{ID: "t1", EstimateSec: 10},
		{ID: "t2", EstimateSec: 20, DependsOn: []string{"t1"}},
		{ID: "t3", EstimateSec: 5, DependsOn: []string{"t1"}},
		{ID: "t4", EstimateSec: 30, DependsOn: []string{"t2", "t3"}},
		{ID: "t5", EstimateSec: 15}, // independent branch
	}
	crit := taskdag.ComputeCriticalPath(tasks)
	assert.Equal(t, []string{"t1", "t2", "t4"}, crit)

	branches := taskdag.ComputeIndependentBranches(tasks)
	assert.Len(t, branches, 2)
}
