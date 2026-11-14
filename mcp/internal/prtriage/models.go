package prtriage

// Ecosystem identifies package management ecosystems from filter-criteria.md §2.
type Ecosystem string

const (
	EcosystemNode   Ecosystem = "npm/yarn/pnpm"
	EcosystemGo     Ecosystem = "go"
	EcosystemRust   Ecosystem = "cargo"
	EcosystemPython Ecosystem = "python"
	EcosystemRuby   Ecosystem = "bundler"
	EcosystemPHP    Ecosystem = "composer"
	EcosystemJava   Ecosystem = "maven/gradle"
	EcosystemDotNet Ecosystem = "dotnet"
	EcosystemAction Ecosystem = "actions-lock"
	EcosystemMixed  Ecosystem = "mixed"
	EcosystemOther  Ecosystem = "other"
)

// FailureArchetype identifies the 5 recurring failure patterns from failure-archetypes.md.
type FailureArchetype string

const (
	ArchetypeOutdatedBase     FailureArchetype = "outdated_base_branch"
	ArchetypeDerivedArtifacts FailureArchetype = "derived_artifacts_drift"
	ArchetypeSnapshotDrift    FailureArchetype = "snapshot_manifest_drift"
	ArchetypeFastForward      FailureArchetype = "fast_forward_policy_violation"
	ArchetypeTokenScope       FailureArchetype = "token_scope_restriction"
)

// CandidatePR represents input PR metadata from gh pr list or glab api.
type CandidatePR struct {
	Number      int      `json:"number" jsonschema:"PR/MR number"`
	Title       string   `json:"title" jsonschema:"PR/MR title"`
	Author      string   `json:"author" jsonschema:"Author username or bot login"`
	HeadRefName string   `json:"head_ref_name,omitempty" jsonschema:"Branch name"`
	Labels      []string `json:"labels,omitempty" jsonschema:"PR labels"`
	Files       []string `json:"files,omitempty" jsonschema:"List of changed file paths"`
	Body        string   `json:"body,omitempty" jsonschema:"Optional PR body description"`
}

// ClassifyFilesInput defines parameters for pr_classify_files.
type ClassifyFilesInput struct {
	Files []string `json:"files" jsonschema:"List of file paths modified in the PR/MR"`
}

// ClassifyFilesOutput holds the result of pr_classify_files.
type ClassifyFilesOutput struct {
	OK            bool      `json:"ok"`
	Eligible      bool      `json:"eligible"`
	Ecosystem     Ecosystem `json:"ecosystem"`
	ManifestFiles []string  `json:"manifest_files"`
	ExcludedFiles []string  `json:"excluded_files"`
	Reasons       []string  `json:"reasons"`
}

// FilterCandidatesInput defines parameters for pr_filter_candidates.
type FilterCandidatesInput struct {
	PRs []CandidatePR `json:"prs" jsonschema:"Batch list of open candidate PRs/MRs"`
}

// EvaluatedPR holds qualification judgment for a single PR.
type EvaluatedPR struct {
	Candidate CandidatePR         `json:"candidate"`
	Eligible  bool                `json:"eligible"`
	FileCheck ClassifyFilesOutput `json:"file_check"`
	Reasons   []string            `json:"reasons"`
}

// FilterCandidatesOutput holds the batch qualification results.
type FilterCandidatesOutput struct {
	OK               bool          `json:"ok"`
	TotalEvaluated   int           `json:"total_evaluated"`
	EligibleCount    int           `json:"eligible_count"`
	ExcludedCount    int           `json:"excluded_count"`
	EligiblePRs      []EvaluatedPR `json:"eligible_prs"`
	DisqualifiedPRs  []EvaluatedPR `json:"disqualified_prs"`
}

// DetectTasksInput defines parameters for pr_detect_tasks.
type DetectTasksInput struct {
	RepoDir string `json:"repo_dir" jsonschema:"Local repository root directory"`
}

// TaskCandidate represents an inspected test, lint, or check command.
type TaskCandidate struct {
	Category string `json:"category"`
	Source   string `json:"source"`
	Command  string `json:"command"`
}

// DetectTasksOutput holds the discovered task runners and verification commands.
type DetectTasksOutput struct {
	OK          bool            `json:"ok"`
	RepoDir     string          `json:"repo_dir"`
	Runners     []string        `json:"runners"`
	Tasks       []TaskCandidate `json:"tasks"`
	HasTests    bool            `json:"has_tests"`
	HasLinters  bool            `json:"has_linters"`
}

// MatchFailureInput defines parameters for pr_match_failure.
type MatchFailureInput struct {
	LogExcerpt string `json:"log_excerpt" jsonschema:"CI build or test failure log excerpt"`
}

// ArchetypeMatch describes a matched failure pattern and its remediation playbook.
type ArchetypeMatch struct {
	Archetype   FailureArchetype `json:"archetype"`
	Title       string           `json:"title"`
	Confidence  string           `json:"confidence"`
	MatchedOn   []string         `json:"matched_on"`
	RootCause   string           `json:"root_cause"`
	Playbook    []string         `json:"playbook"`
}

// MatchFailureOutput holds diagnostic failure classification and playbook suggestions.
type MatchFailureOutput struct {
	OK      bool             `json:"ok"`
	Matched bool             `json:"matched"`
	Matches []ArchetypeMatch `json:"matches"`
	Advice  string           `json:"advice"`
}
