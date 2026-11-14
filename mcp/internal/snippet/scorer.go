package snippet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

const maxSnippetFileBytes = 1024 * 1024

type scoreRule struct {
	level  int
	match  func(m Metrics, hazards []Hazard) bool
	driver func(m Metrics, hazards []Hazard) string
}

var levelRules = []scoreRule{
	{
		level:  4,
		match:  func(m Metrics, _ []Hazard) bool { return m.SLOC >= 26 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("SLOC=%d (>= 26)", m.SLOC) },
	},
	{
		level:  4,
		match:  func(m Metrics, _ []Hazard) bool { return m.FunctionCount >= 3 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("functions=%d (>= 3)", m.FunctionCount) },
	},
	{
		level:  4,
		match:  func(m Metrics, _ []Hazard) bool { return m.MaxNestingDepth >= 3 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("max_nesting_depth=%d (>= 3)", m.MaxNestingDepth) },
	},
	{
		level:  4,
		match:  func(m Metrics, _ []Hazard) bool { return m.BranchCount >= 6 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("branch_count=%d (>= 6)", m.BranchCount) },
	},
	{
		level:  4,
		match:  func(m Metrics, _ []Hazard) bool { return m.MaxPipelineLength >= 3 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("max_pipeline_length=%d (>= 3 pipes)", m.MaxPipelineLength) },
	},
	{
		level:  4,
		match:  func(m Metrics, _ []Hazard) bool { return m.BackgroundCount > 0 || m.TrapCount > 0 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("async/concurrency constructs (background=%d, traps=%d)", m.BackgroundCount, m.TrapCount) },
	},
	{
		level: 4,
		match: func(_ Metrics, h []Hazard) bool { return hasHighSevHazard(h) },
		driver: func(_ Metrics, h []Hazard) string {
			return fmt.Sprintf("high-severity hazard (%s)", formatHazardKinds(h, true))
		},
	},
	{
		level:  3,
		match:  func(m Metrics, _ []Hazard) bool { return m.SLOC >= 21 && m.SLOC <= 25 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("SLOC=%d (21-25)", m.SLOC) },
	},
	{
		level:  3,
		match:  func(m Metrics, _ []Hazard) bool { return m.FunctionCount == 2 || m.MaxFunctionLines > 15 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("functions=%d (max_lines=%d)", m.FunctionCount, m.MaxFunctionLines) },
	},
	{
		level:  3,
		match:  func(m Metrics, _ []Hazard) bool { return m.MaxNestingDepth == 2 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("max_nesting_depth=%d", m.MaxNestingDepth) },
	},
	{
		level:  3,
		match:  func(m Metrics, _ []Hazard) bool { return m.BranchCount >= 3 && m.BranchCount <= 5 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("branch_count=%d (3-5)", m.BranchCount) },
	},
	{
		level:  3,
		match:  func(m Metrics, _ []Hazard) bool { return m.MaxPipelineLength == 2 },
		driver: func(m Metrics, _ []Hazard) string { return "max_pipeline_length=2 (3 chained commands)" },
	},
	{
		level:  3,
		match:  func(m Metrics, _ []Hazard) bool { return m.SubshellCount > 0 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("subshell_count=%d", m.SubshellCount) },
	},
	{
		level:  3,
		match:  func(m Metrics, _ []Hazard) bool { return m.MaxPositionalArg >= 4 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("max_positional_arg=$%d (>= 4)", m.MaxPositionalArg) },
	},
	{
		level: 3,
		match: func(_ Metrics, h []Hazard) bool { return len(h) > 0 && !hasHighSevHazard(h) },
		driver: func(_ Metrics, h []Hazard) string {
			return fmt.Sprintf("shell idiosyncrasies (%s)", formatHazardKinds(h, false))
		},
	},
	{
		level:  2,
		match:  func(m Metrics, _ []Hazard) bool { return m.SLOC >= 11 && m.SLOC <= 20 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("SLOC=%d (11-20)", m.SLOC) },
	},
	{
		level:  2,
		match:  func(m Metrics, _ []Hazard) bool { return m.FunctionCount == 1 && m.MaxFunctionLines <= 15 },
		driver: func(_ Metrics, _ []Hazard) string { return "1 helper function" },
	},
	{
		level:  2,
		match:  func(m Metrics, _ []Hazard) bool { return m.MaxNestingDepth == 1 },
		driver: func(_ Metrics, _ []Hazard) string { return "max_nesting_depth=1" },
	},
	{
		level:  2,
		match:  func(m Metrics, _ []Hazard) bool { return m.BranchCount >= 1 && m.BranchCount <= 2 },
		driver: func(m Metrics, _ []Hazard) string { return fmt.Sprintf("branch_count=%d (1-2)", m.BranchCount) },
	},
}

// Analyze evaluates snippet source code and returns a deterministic 4-tier AuditReport.
// For Python snippets, AST parsing uses Python's built-in ast module and fails immediately on syntax errors without heuristic fallback.
func Analyze(code, language string) AuditReport {
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "" {
		lang = "bash"
	}
	if !isSupportedLanguage(lang) {
		return failedReport(lang, fmt.Sprintf("unsupported language %q (supported: bash, sh, zsh, mksh, python)", lang))
	}

	if isPythonLanguage(lang) {
		return analyzePythonSnippet(code)
	}

	sloc, commentLines, totalLines := countSLOCAndComments(code)
	metrics, hazards, parseErr := analyzeAST(code, lang, sloc, commentLines, totalLines)
	conf := ConfidenceAST
	parseErrStr := ""
	if parseErr != nil {
		conf = ConfidenceHeuristic
		parseErrStr = parseErr.Error()
		metrics, hazards = analyzeHeuristic(code, sloc, commentLines, totalLines)
	}
	return buildReport(lang, conf, parseErrStr, metrics, hazards)
}

func analyzePythonSnippet(code string) AuditReport {
	metrics, hazards, err := analyzePythonAST(code)
	if err != nil {
		return failedReport("python", err.Error())
	}
	return buildReport("python", ConfidenceAST, "", metrics, hazards)
}

func failedReport(lang, parseErr string) AuditReport {
	return AuditReport{
		OK:         false,
		Status:     "analysis_failed",
		Language:   lang,
		ParseError: parseErr,
		Hazards:    []Hazard{},
		Drivers:    []string{},
	}
}

func buildReport(lang string, conf Confidence, parseErrStr string, metrics Metrics, hazards []Hazard) AuditReport {
	level, drivers := scoreMetricsAndHazards(metrics, hazards)
	rating, status := ratingAndStatusForLevel(level)
	if hazards == nil {
		hazards = []Hazard{}
	}
	if drivers == nil {
		drivers = []string{}
	}

	return AuditReport{
		OK:         true,
		Level:      level,
		Rating:     rating,
		Status:     status,
		Language:   lang,
		Confidence: conf,
		ParseError: parseErrStr,
		Metrics:    metrics,
		Hazards:    hazards,
		Drivers:    drivers,
	}
}

// AnalyzeFile reads a snippet file from disk safely and evaluates its intelligibility.
func AnalyzeFile(path, language string) (AuditReport, error) {
	if err := guard.CheckPathAllowed(path); err != nil {
		return AuditReport{}, err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return AuditReport{}, fmt.Errorf("resolve snippet path: %w", err)
	}
	dir := filepath.Dir(absPath)
	base := filepath.Base(absPath)

	root, err := os.OpenRoot(dir)
	if err != nil {
		return AuditReport{}, fmt.Errorf("open snippet parent dir: %w", err)
	}
	defer func() { _ = root.Close() }()

	info, err := root.Stat(base)
	if err != nil {
		return AuditReport{}, fmt.Errorf("stat snippet file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return AuditReport{}, fmt.Errorf("snippet path is not a regular file: %s", path)
	}
	if info.Size() > maxSnippetFileBytes {
		return AuditReport{}, fmt.Errorf("snippet file exceeds %d bytes: %d", maxSnippetFileBytes, info.Size())
	}

	raw, err := root.ReadFile(base)
	if err != nil {
		return AuditReport{}, fmt.Errorf("read snippet file: %w", err)
	}
	rep := Analyze(string(raw), language)
	if !rep.OK {
		return rep, fmt.Errorf("snippet analysis failed: %s", rep.ParseError)
	}
	return rep, nil
}

func scoreMetricsAndHazards(m Metrics, hazards []Hazard) (int, []string) {
	level := 1
	var drivers []string
	for _, r := range levelRules {
		if r.match(m, hazards) {
			if r.level > level {
				level = r.level
			}
			drivers = append(drivers, r.driver(m, hazards))
		}
	}
	return level, drivers
}

func ratingAndStatusForLevel(level int) (string, string) {
	switch level {
	case 1:
		return "Transparent", "approved"
	case 2:
		return "Moderate", "approved"
	case 3:
		return "Complex", "simplify_or_consult"
	default:
		return "Opaque", "halt_and_decompose"
	}
}

func hasHighSevHazard(hazards []Hazard) bool {
	for _, h := range hazards {
		if h.HighSev {
			return true
		}
	}
	return false
}

func formatHazardKinds(hazards []Hazard, onlyHigh bool) string {
	var parts []string
	for _, h := range hazards {
		if onlyHigh && !h.HighSev {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s@L%d", h.Kind, h.Line))
	}
	return strings.Join(parts, ", ")
}
