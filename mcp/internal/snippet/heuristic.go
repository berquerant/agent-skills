package snippet

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	heurFuncRegex       = regexp.MustCompile(`^\s*(function\s+[a-zA-Z_][a-zA-Z0-9_]*|[a-zA-Z_][a-zA-Z0-9_]*\s*\(\s*\))`)
	heurBranchRegex     = regexp.MustCompile(`\b(if|elif|case|for|while|until)\b|&&|\|\|`)
	heurEvalRegex       = regexp.MustCompile(`(^|[;&|()\s])eval\s+`)
	heurProcSubstRegex  = regexp.MustCompile(`[<>]\([^)]+\)`)
	heurPositionalRegex = regexp.MustCompile(`\$([1-9])`)
	heurPipeToShellRe   = regexp.MustCompile(`\|\s*(sh|bash|zsh)\b`)
	heurBgRegex         = regexp.MustCompile(`(^|[^&])&\s*$|\b(coproc|wait)\b`)
	heurTrapRegex       = regexp.MustCompile(`(^|[;&|()\s])trap\s+`)
)

func countSLOCAndComments(code string) (int, int, int) {
	normalized := strings.ReplaceAll(code, "\r\n", "\n")
	trimmed := strings.TrimRight(normalized, "\n")
	if strings.TrimSpace(trimmed) == "" {
		return 0, 0, 0
	}
	lines := strings.Split(trimmed, "\n")
	sloc := 0
	comments := 0
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			comments++
			continue
		}
		sloc++
	}
	return sloc, comments, len(lines)
}

func analyzeHeuristic(code string, sloc, commentLines, totalLines int) (Metrics, []Hazard) {
	m := Metrics{
		TotalLines:   totalLines,
		SLOC:         sloc,
		CommentLines: commentLines,
	}
	var hazards []Hazard
	normalized := strings.ReplaceAll(code, "\r\n", "\n")
	lines := strings.Split(strings.TrimRight(normalized, "\n"), "\n")
	depth := 0

	for idx, raw := range lines {
		lineNum := idx + 1
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		depth = updateHeuristicDepth(&m, line, depth)
		scanHeuristicMetrics(&m, line)
		hazards = append(hazards, scanHeuristicHazards(line, lineNum)...)
	}
	return m, hazards
}

func updateHeuristicDepth(m *Metrics, line string, depth int) int {
	if strings.HasPrefix(line, "if ") || strings.HasPrefix(line, "for ") || strings.HasPrefix(line, "while ") || strings.HasPrefix(line, "case ") {
		depth++
		if depth > m.MaxNestingDepth {
			m.MaxNestingDepth = depth
		}
	}
	if line == "fi" || line == "done" || line == "esac" {
		depth = max(0, depth-1)
	}
	return depth
}

func scanHeuristicMetrics(m *Metrics, line string) {
	if heurFuncRegex.MatchString(line) {
		m.FunctionCount++
	}
	m.BranchCount += len(heurBranchRegex.FindAllString(line, -1))
	if heurBgRegex.MatchString(line) {
		m.BackgroundCount++
	}
	if heurTrapRegex.MatchString(line) {
		m.TrapCount++
	}
	clean := strings.ReplaceAll(line, "||", "")
	pipes := strings.Count(clean, "|")
	if pipes > m.MaxPipelineLength {
		m.MaxPipelineLength = pipes
	}
	for _, match := range heurPositionalRegex.FindAllStringSubmatch(line, -1) {
		if n, err := strconv.Atoi(match[1]); err == nil && n > m.MaxPositionalArg {
			m.MaxPositionalArg = n
		}
	}
}

func scanHeuristicHazards(line string, lineNum int) []Hazard {
	var hazards []Hazard
	if heurEvalRegex.MatchString(line) {
		hazards = append(hazards, Hazard{Kind: HazardEval, Line: lineNum, Detail: "dynamic evaluation via eval", HighSev: true})
	}
	if heurProcSubstRegex.MatchString(line) {
		hazards = append(hazards, Hazard{Kind: HazardProcessSubstitution, Line: lineNum, Detail: "process substitution", HighSev: false})
	}
	if heurPipeToShellRe.MatchString(line) {
		hazards = append(hazards, Hazard{Kind: HazardPipeToShell, Line: lineNum, Detail: "pipeline into shell", HighSev: true})
	}
	return hazards
}
