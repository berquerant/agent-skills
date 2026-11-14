package taskdag

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var safeIDRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`)

// ValidateAndInitTasks validates task IDs, dependencies, and acyclicity, and initializes task states.
func ValidateAndInitTasks(tasks []Task, nowRFC3339 string) ([]Task, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("at least one task is required")
	}

	byID := make(map[string]int, len(tasks))
	out := make([]Task, len(tasks))
	for i, t := range tasks {
		id := strings.TrimSpace(t.ID)
		if !safeIDRegex.MatchString(id) {
			return nil, fmt.Errorf("invalid task id %q", t.ID)
		}
		if _, exists := byID[id]; exists {
			return nil, fmt.Errorf("duplicate task id %q", id)
		}
		byID[id] = i
		out[i] = normalizeTaskInit(t, id, nowRFC3339)
	}

	if err := validateDependencies(out, byID); err != nil {
		return nil, err
	}
	if cycle := findCycle(out); len(cycle) > 0 {
		return nil, fmt.Errorf("circular dependency detected: %s", strings.Join(cycle, " -> "))
	}
	return out, nil
}

func normalizeTaskInit(t Task, id, nowRFC3339 string) Task {
	t.ID = id
	t.Resource = strings.TrimSpace(t.Resource)
	if t.WaitProfile != WaitProfileAsync {
		t.WaitProfile = WaitProfileLocal
	}
	if t.EstimateSec < 0 {
		t.EstimateSec = 0
	}
	if len(t.DependsOn) == 0 {
		t.State = StateReady
	} else {
		t.State = StateBlocked
	}
	t.UpdatedAt = nowRFC3339
	return t
}

func validateDependencies(tasks []Task, byID map[string]int) error {
	for _, t := range tasks {
		seen := make(map[string]bool, len(t.DependsOn))
		for _, dep := range t.DependsOn {
			if dep == t.ID {
				return fmt.Errorf("task %q cannot depend on itself", t.ID)
			}
			if _, ok := byID[dep]; !ok {
				return fmt.Errorf("task %q depends on unknown task %q", t.ID, dep)
			}
			if seen[dep] {
				return fmt.Errorf("task %q has duplicate dependency %q", t.ID, dep)
			}
			seen[dep] = true
		}
	}
	return nil
}

func findCycle(tasks []Task) []string {
	adj := buildDownstreamMap(tasks)
	state := make(map[string]int, len(tasks)) // 0=unvisited, 1=visiting, 2=visited
	parent := make(map[string]string, len(tasks))

	var dfs func(u string) []string
	dfs = func(u string) []string {
		state[u] = 1
		for _, v := range adj[u] {
			if state[v] == 1 {
				return reconstructCycle(u, v, parent)
			}
			if state[v] == 0 {
				parent[v] = u
				if cyc := dfs(v); len(cyc) > 0 {
					return cyc
				}
			}
		}
		state[u] = 2
		return nil
	}

	for _, t := range tasks {
		if state[t.ID] == 0 {
			if cyc := dfs(t.ID); len(cyc) > 0 {
				return cyc
			}
		}
	}
	return nil
}

func reconstructCycle(curr, target string, parent map[string]string) []string {
	path := []string{target, curr}
	for curr != target {
		curr = parent[curr]
		path = append(path, curr)
	}
	slices.Reverse(path)
	return path
}

func buildDownstreamMap(tasks []Task) map[string][]string {
	adj := make(map[string][]string, len(tasks))
	for _, t := range tasks {
		for _, dep := range t.DependsOn {
			adj[dep] = append(adj[dep], t.ID)
		}
	}
	for k := range adj {
		slices.Sort(adj[k])
	}
	return adj
}

// CountTransitiveDescendants returns the number of incomplete downstream tasks reachable from taskID.
func CountTransitiveDescendants(tasks []Task, taskID string) int {
	return len(CollectTransitiveDescendants(tasks, taskID))
}

// CollectTransitiveDescendants returns sorted IDs of all incomplete downstream tasks reachable from taskID.
func CollectTransitiveDescendants(tasks []Task, taskID string) []string {
	adj := buildDownstreamMap(tasks)
	stateByID := make(map[string]TaskState, len(tasks))
	for _, t := range tasks {
		stateByID[t.ID] = t.State
	}

	visited := make(map[string]bool)
	queue := append([]string(nil), adj[taskID]...)
	var result []string
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if visited[curr] {
			continue
		}
		visited[curr] = true
		if stateByID[curr] != StateCompleted {
			result = append(result, curr)
		}
		queue = append(queue, adj[curr]...)
	}
	slices.Sort(result)
	return result
}

// ComputeCriticalPath returns the longest chain of incomplete tasks weighted by max(1, EstimateSec).
func ComputeCriticalPath(tasks []Task) []string {
	adj := buildDownstreamMap(tasks)
	byID := make(map[string]Task, len(tasks))
	for _, t := range tasks {
		byID[t.ID] = t
	}

	memoWeight := make(map[string]int, len(tasks))
	memoNext := make(map[string]string, len(tasks))

	var longestFrom func(u string) int
	longestFrom = func(u string) int {
		if w, ok := memoWeight[u]; ok {
			return w
		}
		t := byID[u]
		if t.State == StateCompleted {
			memoWeight[u] = 0
			return 0
		}
		selfWeight := max(1, t.EstimateSec)
		bestChildWeight := 0
		bestChild := ""
		for _, v := range adj[u] {
			cw := longestFrom(v)
			if cw > bestChildWeight {
				bestChildWeight = cw
				bestChild = v
			}
		}
		memoWeight[u] = selfWeight + bestChildWeight
		memoNext[u] = bestChild
		return memoWeight[u]
	}

	bestStart := ""
	bestTotal := 0
	for _, t := range tasks {
		if t.State == StateCompleted {
			continue
		}
		w := longestFrom(t.ID)
		if w > bestTotal || (w == bestTotal && (bestStart == "" || t.ID < bestStart)) {
			bestTotal = w
			bestStart = t.ID
		}
	}

	var path []string
	for curr := bestStart; curr != ""; curr = memoNext[curr] {
		if byID[curr].State != StateCompleted {
			path = append(path, curr)
		}
	}
	return path
}

// ComputeIndependentBranches groups incomplete tasks into weakly connected components.
func ComputeIndependentBranches(tasks []Task) [][]string {
	undirected := make(map[string][]string, len(tasks))
	incomplete := make(map[string]bool, len(tasks))
	for _, t := range tasks {
		if t.State != StateCompleted {
			incomplete[t.ID] = true
		}
	}
	for _, t := range tasks {
		if !incomplete[t.ID] {
			continue
		}
		for _, dep := range t.DependsOn {
			if incomplete[dep] {
				undirected[t.ID] = append(undirected[t.ID], dep)
				undirected[dep] = append(undirected[dep], t.ID)
			}
		}
	}

	visited := make(map[string]bool, len(incomplete))
	var branches [][]string
	for _, t := range tasks {
		if !incomplete[t.ID] || visited[t.ID] {
			continue
		}
		comp := collectComponent(t.ID, undirected, visited)
		branches = append(branches, comp)
	}
	return branches
}

func collectComponent(start string, undirected map[string][]string, visited map[string]bool) []string {
	var comp []string
	queue := []string{start}
	visited[start] = true
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		comp = append(comp, curr)
		for _, nb := range undirected[curr] {
			if !visited[nb] {
				visited[nb] = true
				queue = append(queue, nb)
			}
		}
	}
	slices.Sort(comp)
	return comp
}
