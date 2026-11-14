// Package pace implements edit-paced execution tracking, non-edit tool call budget enforcement,
// cyclic failure detection, and background process liveness monitoring for the edit-paced-execution skill.
//
// Invariants & Guardrails:
//   - Non-edit thresholds and elapsed time limits prevent unbounded investigative loops without progress.
//   - Cyclic failure detection halts execution if consecutive identical failures occur (cyclic failure limit: 3).
//   - Process watching tracks background command PIDs, heartbeats, and resource consumption.
package pace

// Profile identifies the work complexity preset from pace-monitoring-guide.md §1.
type Profile string

const (
	ProfileQuick    Profile = "quick"
	ProfileStandard Profile = "standard"
	ProfileDeep     Profile = "deep"
)

// ExecMode identifies the task/command execution mode from pace-monitoring-guide.md §2.
type ExecMode string

const (
	ExecModeStandard    ExecMode = "standard"
	ExecModeHeavy       ExecMode = "heavy"
	ExecModeCritical    ExecMode = "critical"
	ExecModeUnmonitored ExecMode = "unmonitored"
)

// TickKind classifies a recorded agent step.
type TickKind string

const (
	TickKindEdit          TickKind = "edit"
	TickKindNonEdit       TickKind = "non_edit"
	TickKindParallelBatch TickKind = "parallel_batch"
)

// PaceStatus indicates whether pacing thresholds are within bounds or exceeded.
type PaceStatus string

const (
	PaceStatusOK       PaceStatus = "OK"
	PaceStatusExceeded PaceStatus = "EXCEEDED"
)

// LivenessState classifies a monitored command/process state.
type LivenessState string

const (
	LivenessActive  LivenessState = "ACTIVE"
	LivenessStalled LivenessState = "STALLED"
	LivenessZombie  LivenessState = "ZOMBIE"
	LivenessExited  LivenessState = "EXITED"
)

// Session stores persistent state for an edit-pacing session.
type Session struct {
	PaceID              string   `json:"pace_id"`
	Profile             Profile  `json:"profile"`
	Mode                ExecMode `json:"mode"`
	TimeLimitSec        int      `json:"time_limit_sec"`
	CallLimit           int      `json:"call_limit"`
	CheckIntervalSec    int      `json:"check_interval_sec"`
	StartedAt           string   `json:"started_at"`
	LastEditAt          string   `json:"last_edit_at"`
	LastEditedFile      string   `json:"last_edited_file,omitempty"`
	NonEditCalls        int      `json:"non_edit_calls"`
	TotalEdits          int      `json:"total_edits"`
	ConsecutiveFailures int      `json:"consecutive_failures"`
	LastNote            string   `json:"last_note,omitempty"`
}

// StartInput defines parameters for pace_start.
type StartInput struct {
	PaceID       string   `json:"pace_id,omitempty" jsonschema:"Session identifier (defaults to 'default')"`
	Profile      Profile  `json:"profile,omitempty" jsonschema:"Complexity preset ('quick', 'standard', or 'deep'; default 'standard')"`
	Mode         ExecMode `json:"mode,omitempty" jsonschema:"Command execution mode ('standard', 'heavy', 'critical', or 'unmonitored'; default 'standard')"`
	TimeLimitSec int      `json:"time_limit_sec,omitempty" jsonschema:"Optional custom time threshold in seconds (> 0 overrides profile)"`
	CallLimit    int      `json:"call_limit,omitempty" jsonschema:"Optional custom non-edit call threshold (> 0 overrides profile)"`
	TaskNote     string   `json:"task_note,omitempty" jsonschema:"Optional summary of task scope or rationale"`
}

// StartOutput holds the initialized pacing session and proposal statement.
type StartOutput struct {
	OK               bool    `json:"ok"`
	Session          Session `json:"session"`
	ProposalTemplate string  `json:"proposal_template"`
}

// TickInput defines parameters for pace_tick.
type TickInput struct {
	PaceID      string   `json:"pace_id,omitempty" jsonschema:"Session identifier (defaults to 'default')"`
	Kind        TickKind `json:"kind" jsonschema:"Step type ('edit', 'non_edit', or 'parallel_batch')"`
	FilePath    string   `json:"file_path,omitempty" jsonschema:"Modified project file path when kind is 'edit'"`
	Note        string   `json:"note,omitempty" jsonschema:"Brief note of what was examined or modified"`
	CyclicError bool     `json:"cyclic_error,omitempty" jsonschema:"True if this step encountered a repeated/cyclic test or command failure"`
}

// PaceEvaluation holds the result of pace_tick, pace_status, or pace_reset.
type PaceEvaluation struct {
	OK               bool       `json:"ok"`
	PaceID           string     `json:"pace_id"`
	Status           PaceStatus `json:"status"`
	ElapsedSec       int        `json:"elapsed_sec"`
	TimeLimitSec     int        `json:"time_limit_sec"`
	RemainingSec     int        `json:"remaining_sec"`
	NonEditCalls     int        `json:"non_edit_calls"`
	CallLimit        int        `json:"call_limit"`
	RemainingCalls   int        `json:"remaining_calls"`
	LastEditedFile   string     `json:"last_edited_file,omitempty"`
	Reasons          []string   `json:"reasons"`
	CheckpointReport string     `json:"checkpoint_report,omitempty"`
	Session          Session    `json:"session"`
}

// StatusInput defines parameters for pace_status.
type StatusInput struct {
	PaceID string `json:"pace_id,omitempty" jsonschema:"Session identifier (defaults to 'default')"`
}

// ResetInput defines parameters for pace_reset.
type ResetInput struct {
	PaceID       string   `json:"pace_id,omitempty" jsonschema:"Session identifier (defaults to 'default')"`
	Profile      Profile  `json:"profile,omitempty" jsonschema:"Optional updated complexity profile"`
	Mode         ExecMode `json:"mode,omitempty" jsonschema:"Optional updated command execution mode"`
	TimeLimitSec int      `json:"time_limit_sec,omitempty" jsonschema:"Optional updated time threshold in seconds"`
	CallLimit    int      `json:"call_limit,omitempty" jsonschema:"Optional updated non-edit call threshold"`
	Note         string   `json:"note,omitempty" jsonschema:"User direction or rationale recorded on reset"`
}

// WatchEntry stores the baseline state for a monitored process/command.
type WatchEntry struct {
	WatchID          string   `json:"watch_id"`
	PID              int      `json:"pid,omitempty"`
	Command          string   `json:"command,omitempty"`
	Mode             ExecMode `json:"mode"`
	CheckIntervalSec int      `json:"check_interval_sec"`
	LogPath          string   `json:"log_path,omitempty"`
	ArtifactPath     string   `json:"artifact_path,omitempty"`
	StartedAt        string   `json:"started_at"`
	LastActivityAt   string   `json:"last_activity_at"`
	LastLogSize      int64    `json:"last_log_size"`
	LastLogMod       string   `json:"last_log_mod,omitempty"`
	LastArtifactMod  string   `json:"last_artifact_mod,omitempty"`
	StallChecks      int      `json:"stall_checks"`
}

// WatchInput defines parameters for pace_watch.
type WatchInput struct {
	WatchID      string   `json:"watch_id,omitempty" jsonschema:"Watch identifier (auto-generated if omitted)"`
	PID          int      `json:"pid,omitempty" jsonschema:"Optional OS process ID to monitor"`
	Command      string   `json:"command,omitempty" jsonschema:"Command line or task description"`
	Mode         ExecMode `json:"mode,omitempty" jsonschema:"Execution mode ('standard', 'heavy', 'critical', 'unmonitored'; default 'standard')"`
	LogPath      string   `json:"log_path,omitempty" jsonschema:"Optional log file path to check for new output"`
	ArtifactPath string   `json:"artifact_path,omitempty" jsonschema:"Optional output artifact path to check for modifications"`
}

// WatchOutput holds the registered WatchEntry.
type WatchOutput struct {
	OK    bool       `json:"ok"`
	Watch WatchEntry `json:"watch"`
}

// LivenessInput defines parameters for pace_liveness.
type LivenessInput struct {
	WatchID string `json:"watch_id" jsonschema:"Watch identifier returned by pace_watch"`
}

// LivenessOutput holds the sampled liveness state and diagnostic report.
type LivenessOutput struct {
	OK          bool          `json:"ok"`
	WatchID     string        `json:"watch_id"`
	State       LivenessState `json:"state"`
	ElapsedSec  int           `json:"elapsed_sec"`
	IdleSec     int           `json:"idle_sec"`
	CPU         float64       `json:"cpu_percent"`
	RSSKB       int64         `json:"rss_kb"`
	LogAdvanced bool          `json:"log_advanced"`
	Reasons     []string      `json:"reasons"`
	StallReport string        `json:"stall_report,omitempty"`
	Watch       WatchEntry    `json:"watch"`
}
