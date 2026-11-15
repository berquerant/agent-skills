// Package skilllint implements automated static linting and quality conformance checks (A through K)
// for agent skills against the Agent Skills Specification and repository conventions.
//
// Automated Check Coverage:
//   - Checks A-C: Inventory completeness, frontmatter structure, and third-person trigger descriptions.
//   - Checks D-F: Standard SKILL.md sections (Step/Verify checkpoints), link portability, and script permissions.
//   - Checks G-K: Security patterns, repository skill index synchronization, and reference self-containment.
package skilllint

// Severity represents the severity of a skill lint finding.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// CheckID identifies one of the automated skill checks (A through K).
type CheckID string

const (
	CheckA CheckID = "A"
	CheckB CheckID = "B"
	CheckC CheckID = "C"
	CheckD CheckID = "D"
	CheckE CheckID = "E"
	CheckF CheckID = "F"
	CheckG CheckID = "G"
	CheckH CheckID = "H"
	CheckI CheckID = "I"
	CheckJ CheckID = "J"
	CheckK CheckID = "K"
)

// Finding represents a single lint error or warning in a skill directory.
type Finding struct {
	Check    CheckID  `json:"check"`
	Severity Severity `json:"severity"`
	File     string   `json:"file,omitempty"`
	Line     int      `json:"line,omitempty"`
	Message  string   `json:"message"`
}

// SkillReport holds the full lint results for a single skill directory.
type SkillReport struct {
	OK                bool           `json:"ok"`
	Passed            bool           `json:"passed"`
	Strict            bool           `json:"strict"`
	SkillName         string         `json:"skill_name"`
	TargetDir         string         `json:"target_dir"`
	SkillLines        int            `json:"skill_lines"`
	Errors            int            `json:"errors"`
	Warnings          int            `json:"warnings"`
	Findings          []Finding      `json:"findings"`
	SiblingRefs       map[string]int `json:"sibling_refs,omitempty"`
	CrossSkillLinks   []string       `json:"cross_skill_links,omitempty"`
	DuplicateHeadings []string       `json:"duplicate_headings,omitempty"`
	ExternalURLs      []string       `json:"external_urls,omitempty"`
}

// BatchReport holds the aggregated lint results across multiple skill directories.
type BatchReport struct {
	OK            bool          `json:"ok"`
	Passed        bool          `json:"passed"`
	Strict        bool          `json:"strict"`
	RootDir       string        `json:"root_dir"`
	TotalSkills   int           `json:"total_skills"`
	PassedSkills  int           `json:"passed_skills"`
	FailedSkills  int           `json:"failed_skills"`
	TotalErrors   int           `json:"total_errors"`
	TotalWarnings int           `json:"total_warnings"`
	Reports       []SkillReport `json:"reports"`
}

// BuildBatchReport aggregates individual skill reports into a consolidated BatchReport.
func BuildBatchReport(reports []SkillReport, rootDir string, strict bool) BatchReport {
	passedSkills := 0
	failedSkills := 0
	totalErrors := 0
	totalWarnings := 0

	for _, rep := range reports {
		totalErrors += rep.Errors
		totalWarnings += rep.Warnings
		if rep.Passed {
			passedSkills++
		} else {
			failedSkills++
		}
	}

	return BatchReport{
		OK:            true,
		Passed:        failedSkills == 0,
		Strict:        strict,
		RootDir:       rootDir,
		TotalSkills:   len(reports),
		PassedSkills:  passedSkills,
		FailedSkills:  failedSkills,
		TotalErrors:   totalErrors,
		TotalWarnings: totalWarnings,
		Reports:       reports,
	}
}

// LintSkillInput defines MCP tool parameters for lint_skill.
type LintSkillInput struct {
	Path   string `json:"path" jsonschema:"Path to the target skill directory containing SKILL.md"`
	Strict bool   `json:"strict,omitempty" jsonschema:"When true, warnings also cause passed=false"`
}

// LintAllInput defines MCP tool parameters for lint_all.
type LintAllInput struct {
	Root   string `json:"root" jsonschema:"Path to the repository root or skills parent directory"`
	Strict bool   `json:"strict,omitempty" jsonschema:"When true, warnings also cause passed=false"`
}
