package taskdag

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

const defaultPollCooldownSec = 30

// ApplyTransition validates and applies a state transition to in.TaskID in dag.
func ApplyTransition(dag *DAG, in TransitionInput, now time.Time) (TransitionResult, error) {
	idx := findTaskIndex(dag.Tasks, in.TaskID)
	if idx < 0 {
		return TransitionResult{}, fmt.Errorf("task %q not found in dag %q", in.TaskID, dag.DAGID)
	}

	curr := dag.Tasks[idx]
	if err := validateTransition(dag.Tasks, curr, in, now); err != nil {
		return TransitionResult{}, err
	}

	nowStr := now.UTC().Format(time.RFC3339)
	nextPoll := ""
	if in.To == StateWaiting {
		var err error
		nextPoll, err = resolveNextPollAt(in.PollAfterSec, in.NextPollAt, now)
		if err != nil {
			return TransitionResult{}, err
		}
	}

	fromState := curr.State
	curr.State = in.To
	curr.NextPollAt = nextPoll
	if in.Note != "" {
		curr.Note = in.Note
	}
	curr.UpdatedAt = nowStr
	dag.Tasks[idx] = curr
	dag.UpdatedAt = nowStr

	promoted := []string{}
	if in.To == StateCompleted {
		if p := promoteReadyDependents(dag, nowStr); len(p) > 0 {
			promoted = p
		}
	}
	haltedDesc := []string{}
	if in.To == StateHalted {
		if h := CollectTransitiveDescendants(dag.Tasks, curr.ID); len(h) > 0 {
			haltedDesc = h
		}
	}

	return TransitionResult{
		Task:              curr,
		PromotedTasks:     promoted,
		HaltedDescendants: haltedDesc,
		Journal: JournalEntry{
			Timestamp:  nowStr,
			TaskID:     curr.ID,
			From:       fromState,
			To:         in.To,
			Note:       in.Note,
			NextPollAt: nextPoll,
		},
	}, nil
}

func validateTransition(tasks []Task, curr Task, in TransitionInput, now time.Time) error {
	if err := guard.CheckNoSecrets(in.Note); err != nil {
		return err
	}
	if !isAllowedTransition(curr.State, in.To) {
		return fmt.Errorf("invalid transition for task %q: %s -> %s", curr.ID, curr.State, in.To)
	}
	if in.To == StateRunning && curr.Resource != "" {
		if holder := findRunningResourceHolder(tasks, curr.Resource, curr.ID); holder != "" {
			return fmt.Errorf("resource collision on %q: task %q is already RUNNING", curr.Resource, holder)
		}
	}
	if curr.State == StateWaiting && (in.To == StateReady || in.To == StateRunning) {
		if rem := remainingPollSeconds(curr.NextPollAt, now); rem > 0 {
			return fmt.Errorf("anti-polling cooldown active for task %q: wait %ds until %s", curr.ID, rem, curr.NextPollAt)
		}
	}
	return nil
}

func isAllowedTransition(from, to TaskState) bool {
	switch from {
	case StateReady:
		return to == StateRunning || to == StateHalted
	case StateRunning:
		return to == StateWaiting || to == StateCompleted || to == StateHalted
	case StateWaiting:
		return to == StateReady || to == StateRunning || to == StateCompleted || to == StateHalted
	default:
		return false
	}
}

func findRunningResourceHolder(tasks []Task, resource, excludeTaskID string) string {
	if resource == "" {
		return ""
	}
	for _, t := range tasks {
		if t.ID != excludeTaskID && t.State == StateRunning && t.Resource == resource {
			return t.ID
		}
	}
	return ""
}

func remainingPollSeconds(nextPollAt string, now time.Time) int {
	if nextPollAt == "" {
		return 0
	}
	ts, err := time.Parse(time.RFC3339, nextPollAt)
	if err != nil || !now.Before(ts) {
		return 0
	}
	return int(math.Ceil(ts.Sub(now).Seconds()))
}

func resolveNextPollAt(pollAfterSec int, explicitNextPollAt string, now time.Time) (string, error) {
	if explicitNextPollAt != "" {
		ts, err := time.Parse(time.RFC3339, explicitNextPollAt)
		if err != nil {
			return "", fmt.Errorf("invalid next_poll_at RFC3339 timestamp %q: %w", explicitNextPollAt, err)
		}
		return ts.UTC().Format(time.RFC3339), nil
	}
	sec := pollAfterSec
	if sec <= 0 {
		sec = defaultPollCooldownSec
	}
	return now.UTC().Add(time.Duration(sec) * time.Second).Format(time.RFC3339), nil
}

func promoteReadyDependents(dag *DAG, nowStr string) []string {
	stateByID := make(map[string]TaskState, len(dag.Tasks))
	for _, t := range dag.Tasks {
		stateByID[t.ID] = t.State
	}

	var promoted []string
	for i, t := range dag.Tasks {
		if t.State != StateBlocked {
			continue
		}
		allDone := true
		for _, dep := range t.DependsOn {
			if stateByID[dep] != StateCompleted {
				allDone = false
				break
			}
		}
		if allDone {
			dag.Tasks[i].State = StateReady
			dag.Tasks[i].UpdatedAt = nowStr
			promoted = append(promoted, t.ID)
		}
	}
	slices.Sort(promoted)
	return promoted
}

func findTaskIndex(tasks []Task, id string) int {
	for i, t := range tasks {
		if t.ID == id {
			return i
		}
	}
	return -1
}
