// Package prtriage implements automated Pull Request triage, file classification, local CI task
// detection, and test failure archetype matching for the resolve-dependency-prs skill.
//
// Capabilities:
//   - File Classification: Categorizes touched files into manifests, lockfiles, source, tests, CI, docs, or unknown.
//   - Candidate Filtering: Validates bot authors, dependency labels, title signatures, and disqualification criteria.
//   - Task Detection: Discovers local test, build, lint, and lockfile update commands across 12+ package ecosystems.
//   - Failure Matching: Analyzes CI/build error logs to classify failures into structured archetypes with remediation hints.
//   - Agent Overrides: Allows agents to supply explicit classifications or allowed PR numbers when heuristics return unknown.
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
	EcosystemElixir Ecosystem = "elixir"
	EcosystemSwift  Ecosystem = "swift"
	EcosystemDart   Ecosystem = "dart"
	EcosystemMixed  Ecosystem = "mixed"
	EcosystemOther  Ecosystem = "other"
)

// FileCategory indicates the structural purpose of a file in the repository.
type FileCategory string

const (
	CategoryManifest    FileCategory = "manifest"
	CategorySource      FileCategory = "source"
	CategoryTest        FileCategory = "test"
	CategoryCI          FileCategory = "ci"
	CategoryBuildConfig FileCategory = "build_config"
	CategoryDoc         FileCategory = "doc"
	CategoryIgnored     FileCategory = "ignored"
	CategoryUnknown     FileCategory = "unknown"
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

// ClassifiedFile provides detailed categorization for a single file.
type ClassifiedFile struct {
	Path      string       `json:"path"`
	Category  FileCategory `json:"category"`
	Ecosystem Ecosystem    `json:"ecosystem,omitempty"`
	Reason    string       `json:"reason,omitempty"`
}

// ClassifyFilesInput defines parameters for pr_classify_files.
type ClassifyFilesInput struct {
	Files                 []string          `json:"files" jsonschema:"List of file paths modified in the PR/MR"`
	Overrides             map[string]string `json:"overrides,omitempty" jsonschema:"Optional map of file path to category ('manifest', 'source', 'test', 'ci', 'build_config', 'doc', 'ignored')"`
	ExtraManifestPatterns []string          `json:"extra_manifest_patterns,omitempty" jsonschema:"Optional glob/basename patterns to treat as manifests"`
}

// ClassifyFilesOutput holds the result of pr_classify_files.
type ClassifyFilesOutput struct {
	OK            bool             `json:"ok"`
	Eligible      bool             `json:"eligible"`
	Ecosystem     Ecosystem        `json:"ecosystem"`
	ManifestFiles []string         `json:"manifest_files"`
	ExcludedFiles []string         `json:"excluded_files"`
	UnknownFiles  []string         `json:"unknown_files"`
	Files         []ClassifiedFile `json:"files"`
	Reasons       []string         `json:"reasons"`
}

// FilterCandidatesInput defines parameters for pr_filter_candidates.
type FilterCandidatesInput struct {
	PRs                   []CandidatePR     `json:"prs" jsonschema:"Batch list of open candidate PRs/MRs"`
	FileOverrides         map[string]string `json:"file_overrides,omitempty" jsonschema:"Optional map of file path to category"`
	ExtraManifestPatterns []string          `json:"extra_manifest_patterns,omitempty" jsonschema:"Optional glob/basename patterns to treat as manifests"`
	AllowedPRNumbers      []int             `json:"allowed_pr_numbers,omitempty" jsonschema:"Optional list of PR numbers explicitly approved as dependency updates by agent"`
	ExtraBotAuthors       []string          `json:"extra_bot_authors,omitempty" jsonschema:"Optional additional bot author logins"`
	ExtraDepLabels        []string          `json:"extra_dep_labels,omitempty" jsonschema:"Optional additional dependency labels"`
	ExtraTitlePatterns    []string          `json:"extra_title_patterns,omitempty" jsonschema:"Optional additional title prefixes or substrings"`
}

// EvaluatedPR holds qualification judgment for a single PR.
type EvaluatedPR struct {
	Candidate CandidatePR         `json:"candidate"`
	Eligible  bool                `json:"eligible"`
	FileCheck ClassifyFilesOutput `json:"file_check"`
	Unknown   bool                `json:"unknown"`
	Reasons   []string            `json:"reasons"`
}

// FilterCandidatesOutput holds the batch qualification results.
type FilterCandidatesOutput struct {
	OK              bool          `json:"ok"`
	TotalEvaluated  int           `json:"total_evaluated"`
	EligibleCount   int           `json:"eligible_count"`
	ExcludedCount   int           `json:"excluded_count"`
	EligiblePRs     []EvaluatedPR `json:"eligible_prs"`
	DisqualifiedPRs []EvaluatedPR `json:"disqualified_prs"`
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
	OK         bool            `json:"ok"`
	RepoDir    string          `json:"repo_dir"`
	Runners    []string        `json:"runners"`
	Tasks      []TaskCandidate `json:"tasks"`
	HasTests   bool            `json:"has_tests"`
	HasLinters bool            `json:"has_linters"`
}

// MatchFailureInput defines parameters for pr_match_failure.
type MatchFailureInput struct {
	LogExcerpt string `json:"log_excerpt" jsonschema:"CI build or test failure log excerpt"`
}

// ArchetypeMatch describes a matched failure pattern and its remediation playbook.
type ArchetypeMatch struct {
	Archetype  FailureArchetype `json:"archetype"`
	Title      string           `json:"title"`
	Confidence string           `json:"confidence"`
	MatchedOn  []string         `json:"matched_on"`
	RootCause  string           `json:"root_cause"`
	Playbook   []string         `json:"playbook"`
}

// MatchFailureOutput holds diagnostic failure classification and playbook suggestions.
type MatchFailureOutput struct {
	OK      bool             `json:"ok"`
	Matched bool             `json:"matched"`
	Matches []ArchetypeMatch `json:"matches"`
	Advice  string           `json:"advice"`
}
