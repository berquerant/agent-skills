package taskdag

import (
	"cmp"
	"fmt"
	"slices"
	"time"
)

// SelectNext picks the highest-priority READY task or reports waiting/completed/halted pipeline state.
func SelectNext(dag *DAG, now time.Time) (NextOutput, bool) {
	mutated := awakenExpiredWaitersIfIdle(dag, now)
	candidates := collectEligibleReadyTasks(dag.Tasks)
	if len(candidates) > 0 {
		sortCandidatesByPriority(dag.Tasks, candidates)
		chosen := candidates[0]
		unblocks := CountTransitiveDescendants(dag.Tasks, chosen.ID)
		return NextOutput{
			OK:            true,
			DAGID:         dag.DAGID,
			Status:        "ready",
			Task:          &chosen,
			Reason:        fmt.Sprintf("selected %q (unblocks=%d, profile=%s, estimate=%ds)", chosen.ID, unblocks, chosen.WaitProfile, chosen.EstimateSec),
			UnblocksCount: unblocks,
		}, mutated
	}
	return buildNonReadyNextOutput(dag, now), mutated
}

func awakenExpiredWaitersIfIdle(dag *DAG, now time.Time) bool {
	if len(collectEligibleReadyTasks(dag.Tasks)) > 0 {
		return false
	}
	oldestIdx := -1
	var oldestPoll time.Time
	for i, t := range dag.Tasks {
		if t.State != StateWaiting || remainingPollSeconds(t.NextPollAt, now) > 0 {
			continue
		}
		if findRunningResourceHolder(dag.Tasks, t.Resource, t.ID) != "" {
			continue
		}
		ts, _ := time.Parse(time.RFC3339, t.NextPollAt)
		if oldestIdx < 0 || ts.Before(oldestPoll) || (ts.Equal(oldestPoll) && t.ID < dag.Tasks[oldestIdx].ID) {
			oldestIdx = i
			oldestPoll = ts
		}
	}
	if oldestIdx < 0 {
		return false
	}
	nowStr := now.UTC().Format(time.RFC3339)
	dag.Tasks[oldestIdx].State = StateReady
	dag.Tasks[oldestIdx].NextPollAt = ""
	dag.Tasks[oldestIdx].UpdatedAt = nowStr
	dag.UpdatedAt = nowStr
	return true
}

func collectEligibleReadyTasks(tasks []Task) []Task {
	var ready []Task
	for _, t := range tasks {
		if t.State != StateReady {
			continue
		}
		if t.Resource != "" && findRunningResourceHolder(tasks, t.Resource, t.ID) != "" {
			continue
		}
		ready = append(ready, t)
	}
	return ready
}

func sortCandidatesByPriority(allTasks []Task, candidates []Task) {
	unblocksMap := make(map[string]int, len(candidates))
	for _, c := range candidates {
		unblocksMap[c.ID] = CountTransitiveDescendants(allTasks, c.ID)
	}
	slices.SortFunc(candidates, func(a, b Task) int {
		if c := cmp.Compare(unblocksMap[b.ID], unblocksMap[a.ID]); c != 0 {
			return c
		}
		if a.WaitProfile != b.WaitProfile {
			if a.WaitProfile == WaitProfileLocal {
				return -1
			}
			return 1
		}
		if c := cmp.Compare(a.EstimateSec, b.EstimateSec); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})
}

func buildNonReadyNextOutput(dag *DAG, now time.Time) NextOutput {
	status := SummarizePipelineStatus(dag.Tasks)
	if status == "completed" || status == "halted" {
		return NextOutput{
			OK:     true,
			DAGID:  dag.DAGID,
			Status: status,
			Reason: fmt.Sprintf("pipeline is %s; no READY tasks remain", status),
		}
	}
	earliestPoll, minRem := earliestWaitingPoll(dag.Tasks, now)
	if earliestPoll != "" {
		return NextOutput{
			OK:               true,
			DAGID:            dag.DAGID,
			Status:           "waiting",
			Reason:           fmt.Sprintf("all active tasks are WAITING or RUNNING; next poll at %s (%ds remaining)", earliestPoll, minRem),
			NextPollAt:       earliestPoll,
			WaitRemainingSec: minRem,
		}
	}
	return NextOutput{
		OK:     true,
		DAGID:  dag.DAGID,
		Status: status,
		Reason: "no eligible READY task available (tasks are RUNNING or resource-locked)",
	}
}

func earliestWaitingPoll(tasks []Task, now time.Time) (string, int) {
	earliest := ""
	minRem := 0
	for _, t := range tasks {
		if t.State != StateWaiting || t.NextPollAt == "" {
			continue
		}
		rem := remainingPollSeconds(t.NextPollAt, now)
		if earliest == "" || rem < minRem {
			earliest = t.NextPollAt
			minRem = rem
		}
	}
	return earliest, minRem
}

// SummarizePipelineStatus computes the overall pipeline status across all tasks.
func SummarizePipelineStatus(tasks []Task) string {
	counts := CountStates(tasks)
	if counts[StateCompleted] == len(tasks) && len(tasks) > 0 {
		return "completed"
	}
	if counts[StateRunning] > 0 || counts[StateReady] > 0 {
		return "active"
	}
	if counts[StateWaiting] > 0 {
		return "waiting"
	}
	if counts[StateHalted] > 0 {
		return "halted"
	}
	return "blocked"
}

// CountStates tallies tasks by TaskState.
func CountStates(tasks []Task) map[TaskState]int {
	counts := map[TaskState]int{
		StateBlocked:   0,
		StateReady:     0,
		StateRunning:   0,
		StateWaiting:   0,
		StateCompleted: 0,
		StateHalted:    0,
	}
	for _, t := range tasks {
		counts[t.State]++
	}
	return counts
}
