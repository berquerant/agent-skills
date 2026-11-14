package checkpoint

// LogStatus represents the overall status in the checkpoint log header.
type LogStatus string

const (
	LogStatusInProgress LogStatus = "IN_PROGRESS"
	LogStatusCompleted  LogStatus = "COMPLETED"
	LogStatusAborted    LogStatus = "ABORTED"
)

// StepStatus represents the status of an individual milestone step entry.
type StepStatus string

const (
	StepStatusCompleted StepStatus = "COMPLETED"
	StepStatusFailed    StepStatus = "FAILED"
	StepStatusBlocked   StepStatus = "BLOCKED"
	StepStatusAdjusted  StepStatus = "ADJUSTED"
)

// DocType classifies the artifact document type under the checkpoint hierarchy.
type DocType string

const (
	DocTypeCheckpoint DocType = "checkpoint"
	DocTypeRunbook    DocType = "runbook"
	DocTypeReport     DocType = "report"
)

// Header holds metadata parsed from the top of a checkpoint log file.
type Header struct {
	Title     string    `json:"title"`
	Created   string    `json:"created"`
	LogPath   string    `json:"log_path"`
	Project   string    `json:"project"`
	Branch    string    `json:"branch"`
	Context   string    `json:"context"`
	Objective string    `json:"objective"`
	Status    LogStatus `json:"status"`
}

// StepEntry holds a parsed milestone checkpoint entry.
type StepEntry struct {
	Step      int        `json:"step"`
	Timestamp string     `json:"timestamp"`
	Title     string     `json:"title"`
	Status    StepStatus `json:"status"`
	Actions   []string   `json:"actions"`
	Findings  []string   `json:"findings"`
	Decisions []string   `json:"decisions"`
	NextSteps []string   `json:"next_steps"`
}

// InitInput defines parameters for checkpoint_init.
type InitInput struct {
	Project     string  `json:"project,omitempty" jsonschema:"Project/repository name (auto-derived from git if omitted)"`
	Branch      string  `json:"branch,omitempty" jsonschema:"Active branch name (auto-derived from git if omitted)"`
	Context     string  `json:"context" jsonschema:"Short kebab-case task context slug (e.g. 'refactor-parser')"`
	DocType     DocType `json:"doc_type,omitempty" jsonschema:"Document type ('checkpoint', 'runbook', or 'report'; default 'checkpoint')"`
	Title       string  `json:"title" jsonschema:"Short task title for the log header"`
	Objective   string  `json:"objective" jsonschema:"One-sentence summary of the task objective"`
	WorkDir     string  `json:"work_dir,omitempty" jsonschema:"Working directory used to auto-derive git project and branch"`
	ProposeOnly bool    `json:"propose_only,omitempty" jsonschema:"When true, resolve and return candidate path without writing to disk"`
}

// InitOutput holds the result of checkpoint_init.
type InitOutput struct {
	OK               bool   `json:"ok"`
	Path             string `json:"path"`
	DualPath         string `json:"dual_path"`
	Project          string `json:"project"`
	Branch           string `json:"branch"`
	Context          string `json:"context"`
	DocType          string `json:"doc_type"`
	ProposeOnly      bool   `json:"propose_only"`
	Created          bool   `json:"created"`
	LatestGeneration string `json:"latest_generation,omitempty"`
}

// AppendInput defines parameters for checkpoint_append.
type AppendInput struct {
	Path      string     `json:"path" jsonschema:"Absolute or checkpoint-root-relative path to the checkpoint log file"`
	Step      int        `json:"step,omitempty" jsonschema:"Optional step number (auto-incremented from existing entries if 0)"`
	Title     string     `json:"title" jsonschema:"Milestone or step title"`
	Status    StepStatus `json:"status" jsonschema:"Step outcome ('COMPLETED', 'FAILED', 'BLOCKED', or 'ADJUSTED')"`
	Actions   []string   `json:"actions" jsonschema:"Executed actions, files edited, or commands run"`
	Findings  []string   `json:"findings" jsonschema:"Verified outcomes, exit codes, or observed facts"`
	Decisions []string   `json:"decisions" jsonschema:"Decisions made and rationale"`
	NextSteps []string   `json:"next_steps" jsonschema:"Immediate next step or approval gate"`
}

// AppendOutput holds the result of checkpoint_append.
type AppendOutput struct {
	OK        bool       `json:"ok"`
	Path      string     `json:"path"`
	DualPath  string     `json:"dual_path"`
	Step      int        `json:"step"`
	Timestamp string     `json:"timestamp"`
	Status    StepStatus `json:"status"`
}

// ResumeInput defines parameters for checkpoint_resume.
type ResumeInput struct {
	Path    string  `json:"path,omitempty" jsonschema:"Explicit path to a checkpoint log file (if omitted, resolves latest generation by context)"`
	Project string  `json:"project,omitempty" jsonschema:"Project name (auto-derived from git if omitted)"`
	Branch  string  `json:"branch,omitempty" jsonschema:"Branch name (auto-derived from git if omitted)"`
	Context string  `json:"context,omitempty" jsonschema:"Task context slug used when path is omitted"`
	DocType DocType `json:"doc_type,omitempty" jsonschema:"Document type ('checkpoint', 'runbook', 'report'; default 'checkpoint')"`
	WorkDir string  `json:"work_dir,omitempty" jsonschema:"Working directory used to auto-derive git project and branch"`
}

// ResumeOutput holds the structured context recovery summary from checkpoint_resume.
type ResumeOutput struct {
	OK               bool        `json:"ok"`
	Path             string      `json:"path"`
	DualPath         string      `json:"dual_path"`
	Header           Header      `json:"header"`
	TotalSteps       int         `json:"total_steps"`
	CompletedSteps   []string    `json:"completed_steps"`
	UnresolvedSteps  []StepEntry `json:"unresolved_steps"`
	LastCompleted    *StepEntry  `json:"last_completed,omitempty"`
	LatestEntry      *StepEntry  `json:"latest_entry,omitempty"`
	CurrentNextSteps []string    `json:"current_next_steps"`
	ParseWarnings    []string    `json:"parse_warnings"`
}

// SetStatusInput defines parameters for checkpoint_set_status.
type SetStatusInput struct {
	Path   string    `json:"path" jsonschema:"Path to the checkpoint log file"`
	Status LogStatus `json:"status" jsonschema:"New overall task status ('IN_PROGRESS', 'COMPLETED', or 'ABORTED')"`
}

// SetStatusOutput holds the result of checkpoint_set_status.
type SetStatusOutput struct {
	OK       bool      `json:"ok"`
	Path     string    `json:"path"`
	DualPath string    `json:"dual_path"`
	Status   LogStatus `json:"status"`
}

// ListInput defines optional filters for checkpoint_list.
type ListInput struct {
	Project string  `json:"project,omitempty" jsonschema:"Filter by project name"`
	Branch  string  `json:"branch,omitempty" jsonschema:"Filter by branch name"`
	Context string  `json:"context,omitempty" jsonschema:"Filter by context slug"`
	DocType DocType `json:"doc_type,omitempty" jsonschema:"Filter by document type"`
}

// LogSummary describes a single checkpoint log file in checkpoint_list.
type LogSummary struct {
	Path      string    `json:"path"`
	DualPath  string    `json:"dual_path"`
	Project   string    `json:"project"`
	Branch    string    `json:"branch"`
	Context   string    `json:"context"`
	Title     string    `json:"title"`
	Status    LogStatus `json:"status"`
	StepCount int       `json:"step_count"`
	UpdatedAt string    `json:"updated_at"`
}

// ListOutput holds the list of discovered checkpoint logs.
type ListOutput struct {
	OK   bool         `json:"ok"`
	Logs []LogSummary `json:"logs"`
}
