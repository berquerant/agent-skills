// Package snippetmgr manages the full lifecycle of command-snippet scripts:
// inline-command extraction checks, self-descriptive naming and header validation,
// write-once immutable versioning with executable permissions (0500), integrated
// AST intelligibility auditing, SHA-256 integrity verification, invocation logging,
// and deprecation/cleanup under an XDG-compliant root directory.
package snippetmgr

import (
	"github.com/berquerant/agent-skills/mcp/internal/snippet"
)

// Scope categorizes snippet longevity.
type Scope string

const (
	ScopeOneOff   Scope = "one-off"
	ScopeReusable Scope = "reusable"
)

// VersionStatus tracks the lifecycle state of a snippet version.
type VersionStatus string

const (
	StatusActive     VersionStatus = "active"
	StatusDeprecated VersionStatus = "deprecated"
	StatusCleaned    VersionStatus = "cleaned"
)

// CheckInlineInput defines parameters for snippet_check_inline.
type CheckInlineInput struct {
	Command  string `json:"command" jsonschema:"Inline shell or interpreter command string to evaluate against command-snippet extraction thresholds"`
	Language string `json:"language,omitempty" jsonschema:"Language for AST complexity analysis ('bash', 'sh', 'zsh', 'mksh', 'python'; default 'bash')"`
}

// CheckInlineOutput reports whether an inline command should be extracted into a snippet file.
type CheckInlineOutput struct {
	ShouldExtract     bool                `json:"should_extract"`
	Reasons           []string            `json:"reasons"`
	CharCount         int                 `json:"char_count"`
	LineCount         int                 `json:"line_count"`
	ChainedCommands   int                 `json:"chained_commands"`
	SuggestedLanguage string              `json:"suggested_language"`
	Audit             snippet.AuditReport `json:"audit"`
	LightweightHint   string              `json:"lightweight_hint,omitempty"`
}

// VersionRecord captures immutable metadata for a single saved snippet version.
type VersionRecord struct {
	Version      int                 `json:"version"`
	Filename     string              `json:"filename"`
	Path         string              `json:"path"`
	DualPath     string              `json:"dual_path"`
	SHA256       string              `json:"sha256"`
	Status       VersionStatus       `json:"status"`
	AuditLevel   int                 `json:"audit_level"`
	AuditRating  string              `json:"audit_rating"`
	AuditDrivers []string            `json:"audit_drivers"`
	AllowComplex bool                `json:"allow_complex,omitempty"`
	CreatedAt    string              `json:"created_at"`
	DeprecatedAt string              `json:"deprecated_at,omitempty"`
	CleanedAt    string              `json:"cleaned_at,omitempty"`
	Audit        snippet.AuditReport `json:"audit"`
}

// InvocationRecord logs a host-side execution of a managed snippet version.
type InvocationRecord struct {
	Version    int      `json:"version"`
	ExecutedAt string   `json:"executed_at"`
	Args       []string `json:"args,omitempty"`
	ExitCode   int      `json:"exit_code"`
	DurationMs int64    `json:"duration_ms,omitempty"`
	Note       string   `json:"note,omitempty"`
}

// SnippetMeta is the persisted JSON ledger for a named snippet across versions.
type SnippetMeta struct {
	Name          string             `json:"name"`
	Project       string             `json:"project"`
	Branch        string             `json:"branch"`
	Scope         Scope              `json:"scope"`
	Language      string             `json:"language"`
	Extension     string             `json:"extension"`
	Purpose       string             `json:"purpose,omitempty"`
	LatestVersion int                `json:"latest_version"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
	Versions      []VersionRecord    `json:"versions"`
	Invocations   []InvocationRecord `json:"invocations,omitempty"`
}

// CreateInput defines parameters for snippet_create.
type CreateInput struct {
	Name         string `json:"name" jsonschema:"Self-descriptive snippet name (lowercase alphanumeric with _ or -, no extension or _vN suffix)"`
	Code         string `json:"code" jsonschema:"Complete snippet source code including shebang, documentation header, and shell safety flags"`
	Language     string `json:"language,omitempty" jsonschema:"Snippet language ('bash', 'sh', 'zsh', 'mksh', 'python'; default 'bash')"`
	Scope        Scope  `json:"scope,omitempty" jsonschema:"Snippet scope ('one-off' or 'reusable'; default 'one-off')"`
	Purpose      string `json:"purpose,omitempty" jsonschema:"Concise human-readable purpose used in metadata and the Pre-Execution Notice"`
	Project      string `json:"project,omitempty" jsonschema:"Project slug (auto-derived from git top-level directory if omitted)"`
	Branch       string `json:"branch,omitempty" jsonschema:"Branch slug (auto-derived from git branch or HEAD SHA if omitted)"`
	WorkDir      string `json:"work_dir,omitempty" jsonschema:"Working directory used for git project/branch auto-derivation"`
	AllowComplex bool   `json:"allow_complex,omitempty" jsonschema:"Set true only after explicit user consultation to allow saving Level 3/4 snippets"`
}

// CreateOutput reports the result of snippet_create.
type CreateOutput struct {
	Saved                bool                `json:"saved"`
	ActionRequired       string              `json:"action_required,omitempty"`
	Name                 string              `json:"name"`
	Project              string              `json:"project"`
	Branch               string              `json:"branch"`
	Scope                Scope               `json:"scope"`
	Language             string              `json:"language"`
	Version              int                 `json:"version,omitempty"`
	Path                 string              `json:"path,omitempty"`
	DualPath             string              `json:"dual_path,omitempty"`
	SHA256               string              `json:"sha256,omitempty"`
	DeprecatedVersions   []int               `json:"deprecated_versions,omitempty"`
	Audit                snippet.AuditReport `json:"audit"`
	PreExecutionNotice   string              `json:"pre_execution_notice,omitempty"`
	DirectInvocationHint string              `json:"direct_invocation_hint,omitempty"`
}

// GetInput defines parameters for snippet_get.
type GetInput struct {
	Name    string `json:"name" jsonschema:"Snippet name"`
	Scope   Scope  `json:"scope,omitempty" jsonschema:"Snippet scope ('one-off' or 'reusable'; default 'one-off')"`
	Version int    `json:"version,omitempty" jsonschema:"Optional version number to inspect (defaults to latest version)"`
	Project string `json:"project,omitempty" jsonschema:"Project slug (auto-derived from git if omitted)"`
	Branch  string `json:"branch,omitempty" jsonschema:"Branch slug (auto-derived from git if omitted)"`
	WorkDir string `json:"work_dir,omitempty" jsonschema:"Working directory for git auto-derivation"`
}

// GetOutput returns metadata and the selected version record for a snippet.
type GetOutput struct {
	Meta            SnippetMeta    `json:"meta"`
	SelectedVersion *VersionRecord `json:"selected_version,omitempty"`
}

// ListInput defines parameters for snippet_list.
type ListInput struct {
	Project string        `json:"project,omitempty" jsonschema:"Filter by project slug (empty lists all projects)"`
	Branch  string        `json:"branch,omitempty" jsonschema:"Filter by branch slug (empty lists all branches)"`
	Scope   Scope         `json:"scope,omitempty" jsonschema:"Filter by scope ('one-off' or 'reusable')"`
	Status  VersionStatus `json:"status,omitempty" jsonschema:"Filter by latest version status ('active', 'deprecated', 'cleaned')"`
}

// SnippetSummary is a compact entry returned by snippet_list.
type SnippetSummary struct {
	Name            string        `json:"name"`
	Project         string        `json:"project"`
	Branch          string        `json:"branch"`
	Scope           Scope         `json:"scope"`
	Language        string        `json:"language"`
	LatestVersion   int           `json:"latest_version"`
	LatestStatus    VersionStatus `json:"latest_status"`
	LatestPath      string        `json:"latest_path"`
	LatestDualPath  string        `json:"latest_dual_path"`
	LatestAuditLvl  int           `json:"latest_audit_level"`
	InvocationCount int           `json:"invocation_count"`
	UpdatedAt       string        `json:"updated_at"`
}

// ListOutput returns matching managed snippets.
type ListOutput struct {
	Snippets []SnippetSummary `json:"snippets"`
	Total    int              `json:"total"`
}

// VerifyInput defines parameters for snippet_verify.
type VerifyInput struct {
	Name    string `json:"name" jsonschema:"Snippet name"`
	Scope   Scope  `json:"scope,omitempty" jsonschema:"Snippet scope ('one-off' or 'reusable'; default 'one-off')"`
	Version int    `json:"version,omitempty" jsonschema:"Version number to verify (defaults to latest version)"`
	Project string `json:"project,omitempty" jsonschema:"Project slug (auto-derived from git if omitted)"`
	Branch  string `json:"branch,omitempty" jsonschema:"Branch slug (auto-derived from git if omitted)"`
	WorkDir string `json:"work_dir,omitempty" jsonschema:"Working directory for git auto-derivation"`
}

// VerifyOutput reports integrity and executability status prior to host execution.
type VerifyOutput struct {
	Valid          bool          `json:"valid"`
	Name           string        `json:"name"`
	Version        int           `json:"version"`
	Status         VersionStatus `json:"status"`
	Path           string        `json:"path"`
	DualPath       string        `json:"dual_path"`
	ExpectedSHA256 string        `json:"expected_sha256"`
	ActualSHA256   string        `json:"actual_sha256,omitempty"`
	Executable     bool          `json:"executable"`
	Immutable      bool          `json:"immutable"`
	Issues         []string      `json:"issues"`
}

// RecordInvocationInput defines parameters for snippet_record_invocation.
type RecordInvocationInput struct {
	Name       string   `json:"name" jsonschema:"Snippet name"`
	Scope      Scope    `json:"scope,omitempty" jsonschema:"Snippet scope ('one-off' or 'reusable'; default 'one-off')"`
	Version    int      `json:"version,omitempty" jsonschema:"Executed version number (defaults to latest version)"`
	Args       []string `json:"args,omitempty" jsonschema:"CLI arguments passed to the snippet"`
	ExitCode   int      `json:"exit_code" jsonschema:"Process exit code returned by host execution"`
	DurationMs int64    `json:"duration_ms,omitempty" jsonschema:"Execution duration in milliseconds"`
	Note       string   `json:"note,omitempty" jsonschema:"Optional concise outcome note"`
	Project    string   `json:"project,omitempty" jsonschema:"Project slug (auto-derived from git if omitted)"`
	Branch     string   `json:"branch,omitempty" jsonschema:"Branch slug (auto-derived from git if omitted)"`
	WorkDir    string   `json:"work_dir,omitempty" jsonschema:"Working directory for git auto-derivation"`
}

// RecordInvocationOutput confirms an invocation entry was recorded.
type RecordInvocationOutput struct {
	Recorded        bool             `json:"recorded"`
	Name            string           `json:"name"`
	Version         int              `json:"version"`
	TotalRecorded   int              `json:"total_recorded"`
	LastInvocation  InvocationRecord `json:"last_invocation"`
	MetaStoragePath string           `json:"meta_storage_path"`
}

// DeprecateInput defines parameters for snippet_deprecate.
type DeprecateInput struct {
	Name    string `json:"name" jsonschema:"Snippet name"`
	Scope   Scope  `json:"scope,omitempty" jsonschema:"Snippet scope ('one-off' or 'reusable'; default 'one-off')"`
	Version int    `json:"version,omitempty" jsonschema:"Version number to deprecate (0 deprecates all active versions)"`
	Project string `json:"project,omitempty" jsonschema:"Project slug (auto-derived from git if omitted)"`
	Branch  string `json:"branch,omitempty" jsonschema:"Branch slug (auto-derived from git if omitted)"`
	WorkDir string `json:"work_dir,omitempty" jsonschema:"Working directory for git auto-derivation"`
}

// DeprecateOutput reports which versions were marked deprecated and stripped of execute permission.
type DeprecateOutput struct {
	Name               string `json:"name"`
	DeprecatedVersions []int  `json:"deprecated_versions"`
}

// CleanupInput defines parameters for snippet_cleanup.
type CleanupInput struct {
	Project             string `json:"project,omitempty" jsonschema:"Filter by project slug (auto-derived if empty and all_projects is false)"`
	Branch              string `json:"branch,omitempty" jsonschema:"Filter by branch slug (empty cleans all branches in project)"`
	Scope               Scope  `json:"scope,omitempty" jsonschema:"Filter by scope ('one-off' or 'reusable')"`
	Name                string `json:"name,omitempty" jsonschema:"Optional specific snippet name to clean"`
	IncludeActiveOneOff bool   `json:"include_active_one_off,omitempty" jsonschema:"When true, also remove active one-off snippet files (reusable active files are never removed)"`
	DryRun              bool   `json:"dry_run,omitempty" jsonschema:"When true, report files that would be removed without deleting them"`
	AllProjects         bool   `json:"all_projects,omitempty" jsonschema:"When true, clean across all projects under snippet root"`
	WorkDir             string `json:"work_dir,omitempty" jsonschema:"Working directory for git auto-derivation"`
}

// CleanupOutput reports files removed (or slated for removal in dry_run mode).
type CleanupOutput struct {
	DryRun       bool     `json:"dry_run"`
	RemovedFiles []string `json:"removed_files"`
	RemovedCount int      `json:"removed_count"`
}
