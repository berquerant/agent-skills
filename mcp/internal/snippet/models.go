// Package snippet performs static AST analysis, complexity scoring, and hazard detection
// on shell snippets and inline command scripts for the snippet-review skill.
//
// Analysis Capabilities:
//   - AST Extraction: Uses mvdan.cc/sh/v3 to parse POSIX/Bash syntax into concrete syntax trees
//     to compute exact SLOC, nesting depth, branch count, pipeline length, and variable counts.
//   - Hazard Detection: Detects high-risk shell idioms (e.g. `eval`, unquoted variable expansion,
//     pipe-to-shell, arbitrary interpreter execution) and flags their line numbers.
//   - Tier Classification: Maps metrics against the 4-level Intelligibility Rubric (Level 1 to 4).
package snippet

// Confidence indicates whether metrics were extracted via full AST parsing or heuristic fallback.
type Confidence string

const (
	ConfidenceAST       Confidence = "ast"
	ConfidenceHeuristic Confidence = "heuristic"
)

// HazardKind categorizes a shell or language construct that degrades snippet intelligibility or safety.
type HazardKind string

const (
	HazardEval                  HazardKind = "eval"
	HazardSource                HazardKind = "source"
	HazardProcessSubstitution   HazardKind = "process_substitution"
	HazardInlineInterpreter     HazardKind = "inline_interpreter"
	HazardPipeToShell           HazardKind = "pipe_to_shell"
	HazardDenseTextEngine       HazardKind = "dense_text_engine"
	HazardObscureParamExpansion HazardKind = "obscure_param_expansion"
)

// Hazard represents a specific construct detected in the snippet with its 1-based line number.
type Hazard struct {
	Kind    HazardKind `json:"kind"`
	Line    int        `json:"line"`
	Detail  string     `json:"detail"`
	HighSev bool       `json:"high_severity"`
}

// Metrics holds the 5 evaluation dimensions defined in the snippet intelligibility rubric.
type Metrics struct {
	TotalLines        int `json:"total_lines"`
	SLOC              int `json:"sloc"`
	CommentLines      int `json:"comment_lines"`
	FunctionCount     int `json:"function_count"`
	MaxFunctionLines  int `json:"max_function_lines"`
	MaxPositionalArg  int `json:"max_positional_arg"`
	BranchCount       int `json:"branch_count"`
	MaxNestingDepth   int `json:"max_nesting_depth"`
	BackgroundCount   int `json:"background_count"`
	SubshellCount     int `json:"subshell_count"`
	TrapCount         int `json:"trap_count"`
	MaxPipelineLength int `json:"max_pipeline_length"`
}

// AuditReport holds the complete snippet intelligibility evaluation result.
type AuditReport struct {
	OK         bool       `json:"ok"`
	Level      int        `json:"level"`
	Rating     string     `json:"rating"`
	Status     string     `json:"status"`
	Language   string     `json:"language"`
	Confidence Confidence `json:"confidence"`
	ParseError string     `json:"parse_error,omitempty"`
	Metrics    Metrics    `json:"metrics"`
	Hazards    []Hazard   `json:"hazards"`
	Drivers    []string   `json:"drivers"`
}

// AuditInput defines MCP tool parameters for snippet_audit.
type AuditInput struct {
	Code     string `json:"code,omitempty" jsonschema:"Inline snippet source code to audit (provide either code or path)"`
	Path     string `json:"path,omitempty" jsonschema:"Path to a saved snippet file to audit (provide either code or path)"`
	Language string `json:"language,omitempty" jsonschema:"Snippet language ('bash', 'sh', 'zsh', 'mksh', 'python'; default 'bash')"`
}
