package snippetmgr

import (
	"fmt"
	"slices"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/snippet"
	"mvdan.cc/sh/v3/syntax"
)

const (
	maxInlineChars       = 120
	maxInlineLines       = 3
	maxInlineChainedCmds = 4
)

// CheckInline evaluates an inline command against objective structural thresholds
// (character length, line count, chained command count) and delegates AST complexity
// and hazard analysis to snippet.Analyze for both shell and Python.
func CheckInline(in CheckInlineInput) (CheckInlineOutput, error) {
	lang, _, err := NormalizeLanguage(in.Language)
	if err != nil {
		return CheckInlineOutput{}, err
	}

	cmd := strings.TrimSpace(in.Command)
	if cmd == "" {
		return CheckInlineOutput{
			ShouldExtract:     false,
			Reasons:           []string{},
			SuggestedLanguage: lang,
		}, nil
	}

	charCount := len(cmd)
	lineCount := countNonEmptyLines(cmd)
	var reasons []string

	if charCount >= maxInlineChars {
		reasons = append(reasons, fmt.Sprintf("character length %d >= %d", charCount, maxInlineChars))
	}
	if lineCount >= maxInlineLines {
		reasons = append(reasons, fmt.Sprintf("line count %d >= %d (multi-line)", lineCount, maxInlineLines))
	}

	audit := snippet.Analyze(cmd, lang)
	if !audit.OK {
		reasons = appendUnique(reasons, fmt.Sprintf("syntax analysis failed: %s", audit.ParseError))
	} else if audit.Level >= snippet.LevelComplex {
		for _, d := range audit.Drivers {
			reasons = appendUnique(reasons, fmt.Sprintf("complexity Level %d (%s): %s", audit.Level, audit.Rating, d))
		}
	}

	chained := 0
	if lang != "python" {
		chained = countShellCommands(cmd)
		if chained >= maxInlineChainedCmds {
			reasons = appendUnique(reasons, fmt.Sprintf("chained commands %d >= %d", chained, maxInlineChainedCmds))
		}
	}

	if reasons == nil {
		reasons = []string{}
	}

	out := CheckInlineOutput{
		ShouldExtract:     len(reasons) > 0,
		Reasons:           reasons,
		CharCount:         charCount,
		LineCount:         lineCount,
		ChainedCommands:   chained,
		SuggestedLanguage: inferSuggestedLanguage(lang, audit),
		Audit:             audit,
	}
	if out.ShouldExtract && audit.Level < snippet.LevelComplex {
		out.LightweightHint = "If strictly read-only and local, consider splitting into simple <120-char commands first; otherwise create a snippet."
	}
	return out, nil
}

func inferSuggestedLanguage(lang string, audit snippet.AuditReport) string {
	if lang == "python" {
		return "python"
	}
	for _, h := range audit.Hazards {
		if h.Kind == snippet.HazardInlineInterpreter &&
			(strings.Contains(h.Detail, "python ") || strings.Contains(h.Detail, "python3 ")) {
			return "python"
		}
	}
	return lang
}

func countNonEmptyLines(s string) int {
	count := 0
	for line := range strings.SplitSeq(s, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func countShellCommands(cmd string) int {
	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))
	file, err := parser.Parse(strings.NewReader(cmd), "inline")
	if err != nil {
		return countHeuristicChains(cmd)
	}
	callCount := 0
	syntax.Walk(file, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.CallExpr); ok && len(call.Args) > 0 {
			callCount++
		}
		return true
	})
	return callCount
}

func countHeuristicChains(cmd string) int {
	count := 1
	for _, op := range []string{"&&", "||", "|", ";"} {
		count += strings.Count(cmd, op)
	}
	return count
}

func appendUnique(dst []string, item string) []string {
	if slices.Contains(dst, item) {
		return dst
	}
	return append(dst, item)
}
