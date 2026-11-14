package taskdag

// TaskState represents the formal lifecycle state of a task node in a DAG.
type TaskState string

const (
	StateBlocked   TaskState = "BLOCKED"
	StateReady     TaskState = "READY"
	StateRunning   TaskState = "RUNNING"
	StateWaiting   TaskState = "WAITING"
	StateCompleted TaskState = "COMPLETED"
	StateHalted    TaskState = "HALTED"
)

// WaitProfile indicates whether a task executes locally or triggers an external async wait.
type WaitProfile string

const (
	WaitProfileLocal WaitProfile = "local"
	WaitProfileAsync WaitProfile = "async"
)

// Task defines a single node in the task dependency graph.
type Task struct {
	ID          string      `json:"id" jsonschema:"Unique alphanumeric task identifier"`
	Title       string      `json:"title,omitempty" jsonschema:"Short human-readable task title"`
	DependsOn   []string    `json:"depends_on,omitempty" jsonschema:"Upstream task IDs that must reach COMPLETED first"`
	Resource    string      `json:"resource,omitempty" jsonschema:"Worktree path or resource lock key preventing concurrent execution"`
	WaitProfile WaitProfile `json:"wait_profile,omitempty" jsonschema:"Execution profile ('local' or 'async'; default 'local')"`
	EstimateSec int         `json:"estimate_sec,omitempty" jsonschema:"Estimated duration in seconds (smaller prioritized on tie)"`
	State       TaskState   `json:"state,omitempty" jsonschema:"Current task lifecycle state"`
	NextPollAt  string      `json:"next_poll_at,omitempty" jsonschema:"RFC3339 timestamp before which a WAITING task cannot be polled"`
	Note        string      `json:"note,omitempty" jsonschema:"Latest status or checkpoint note"`
	UpdatedAt   string      `json:"updated_at,omitempty" jsonschema:"RFC3339 timestamp of the last state transition"`
}

// DAG holds the complete persistent state of a task dependency pipeline.
type DAG struct {
	DAGID     string `json:"dag_id"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Tasks     []Task `json:"tasks"`
}

// JournalEntry records an append-only state transition event for a DAG.
type JournalEntry struct {
	Timestamp  string    `json:"timestamp"`
	TaskID     string    `json:"task_id"`
	From       TaskState `json:"from"`
	To         TaskState `json:"to"`
	Note       string    `json:"note,omitempty"`
	NextPollAt string    `json:"next_poll_at,omitempty"`
}

// CreateInput defines parameters for dag_create.
type CreateInput struct {
	DAGID string `json:"dag_id,omitempty" jsonschema:"Optional custom DAG identifier (auto-generated if omitted)"`
	Name  string `json:"name,omitempty" jsonschema:"Optional human-readable pipeline name"`
	Tasks []Task `json:"tasks" jsonschema:"List of task definitions with dependencies"`
}

// CreateOutput holds the result of dag_create.
type CreateOutput struct {
	OK                 bool     `json:"ok"`
	DAGID              string   `json:"dag_id"`
	Name               string   `json:"name,omitempty"`
	TotalTasks         int      `json:"total_tasks"`
	ReadyTasks         []string `json:"ready_tasks"`
	BlockedTasks       []string `json:"blocked_tasks"`
	CriticalPath       []string `json:"critical_path"`
	IndependentStreams int      `json:"independent_streams"`
}

// NextInput defines parameters for dag_next.
type NextInput struct {
	DAGID string `json:"dag_id" jsonschema:"Target DAG identifier"`
}

// NextOutput holds the scheduling recommendation from dag_next.
type NextOutput struct {
	OK               bool   `json:"ok"`
	DAGID            string `json:"dag_id"`
	Status           string `json:"status"`
	Task             *Task  `json:"task,omitempty"`
	Reason           string `json:"reason"`
	UnblocksCount    int    `json:"unblocks_count,omitempty"`
	NextPollAt       string `json:"next_poll_at,omitempty"`
	WaitRemainingSec int    `json:"wait_remaining_sec,omitempty"`
}

// TransitionInput defines parameters for dag_transition.
type TransitionInput struct {
	DAGID        string    `json:"dag_id" jsonschema:"Target DAG identifier"`
	TaskID       string    `json:"task_id" jsonschema:"Target task identifier"`
	To           TaskState `json:"to" jsonschema:"Target state (RUNNING, WAITING, COMPLETED, HALTED, READY)"`
	Note         string    `json:"note,omitempty" jsonschema:"Optional transition note or wait condition"`
	PollAfterSec int       `json:"poll_after_sec,omitempty" jsonschema:"Cooldown seconds before WAITING task can be re-evaluated (default 30)"`
	NextPollAt   string    `json:"next_poll_at,omitempty" jsonschema:"Explicit RFC3339 timestamp for next poll check"`
}

// TransitionOutput holds the result of dag_transition.
type TransitionOutput struct {
	OK                bool     `json:"ok"`
	DAGID             string   `json:"dag_id"`
	Task              Task     `json:"task"`
	PromotedTasks     []string `json:"promoted_tasks"`
	HaltedDescendants []string `json:"halted_descendants"`
	PipelineStatus    string   `json:"pipeline_status"`
}

// StatusInput defines parameters for dag_status.
type StatusInput struct {
	DAGID string `json:"dag_id" jsonschema:"Target DAG identifier"`
}

// StatusOutput holds the full status summary from dag_status.
type StatusOutput struct {
	OK                  bool              `json:"ok"`
	DAG                 DAG               `json:"dag"`
	Counts              map[TaskState]int `json:"counts"`
	CriticalPath        []string          `json:"critical_path"`
	IndependentBranches [][]string        `json:"independent_branches"`
	PipelineStatus      string            `json:"pipeline_status"`
}

// RenderInput defines parameters for dag_render.
type RenderInput struct {
	DAGID string `json:"dag_id" jsonschema:"Target DAG identifier"`
}

// RenderOutput holds the Mermaid flowchart representation of a DAG.
type RenderOutput struct {
	OK      bool   `json:"ok"`
	DAGID   string `json:"dag_id"`
	Mermaid string `json:"mermaid"`
}

// DAGSummary provides metadata for a stored DAG in dag_list.
type DAGSummary struct {
	DAGID          string `json:"dag_id"`
	Name           string `json:"name,omitempty"`
	TotalTasks     int    `json:"total_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
	PipelineStatus string `json:"pipeline_status"`
	UpdatedAt      string `json:"updated_at"`
}

// ListInput defines parameters for dag_list.
type ListInput struct{}

// ListOutput holds the list of stored DAGs.
type ListOutput struct {
	OK   bool         `json:"ok"`
	DAGs []DAGSummary `json:"dags"`
}

// DeleteInput defines parameters for dag_delete.
type DeleteInput struct {
	DAGID string `json:"dag_id" jsonschema:"Target DAG identifier to delete"`
}

// DeleteOutput holds the result of dag_delete.
type DeleteOutput struct {
	OK      bool   `json:"ok"`
	DAGID   string `json:"dag_id"`
	Deleted bool   `json:"deleted"`
}
