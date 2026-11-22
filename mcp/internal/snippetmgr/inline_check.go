package snippetmgr

import (
	"fmt"
	"path/filepath"
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

var (
	broadCLITools   = []string{"gh", "aws", "gcloud", "kubectl"}
	searchFileTools = []string{"grep", "egrep", "fgrep", "rg"}
)

// CheckInline evaluates an inline command against command-snippet Step 1 avoidance thresholds.
// Structural complexity, branching (&&/||/if/for), pipelines, inline interpreters, and hazards
// are delegated to snippet.Analyze so both shell and Python share a single AST analysis engine.
func CheckInline(in CheckInlineInput) CheckInlineOutput {
	lang, _, err := NormalizeLanguage(in.Language)
	if err != nil {
		lang = "bash"
	}

	cmd := strings.TrimSpace(in.Command)
	if cmd == "" {
		return CheckInlineOutput{
			ShouldExtract:     false,
			Reasons:           []string{},
			SuggestedLanguage: lang,
		}
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
	} else if audit.Level >= 3 {
		for _, d := range audit.Drivers {
			reasons = appendUnique(reasons, fmt.Sprintf("complexity Level %d (%s): %s", audit.Level, audit.Rating, d))
		}
	}

	suggestedLang := inferSuggestedLanguage(lang, audit)
	chained := 0
	if lang != "python" {
		var frictionReasons []string
		chained, frictionReasons = inspectShellFriction(cmd)
		if chained >= maxInlineChainedCmds {
			reasons = appendUnique(reasons, fmt.Sprintf("chained commands %d >= %d", chained, maxInlineChainedCmds))
		}
		for _, r := range frictionReasons {
			reasons = appendUnique(reasons, r)
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
		SuggestedLanguage: suggestedLang,
		Audit:             audit,
	}
	if out.ShouldExtract && audit.Level < 3 && !containsBroadOrInterpreter(reasons) {
		out.LightweightHint = "If strictly read-only and local, consider splitting into simple <120-char commands first; otherwise create a snippet."
	}
	return out
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

func inspectShellFriction(cmd string) (int, []string) {
	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))
	file, err := parser.Parse(strings.NewReader(cmd), "inline")
	if err != nil {
		return countHeuristicChains(cmd), []string{"shell syntax parse error (extract to inspectable snippet)"}
	}

	var (
		callCount int
		reasons   []string
	)
	syntax.Walk(file, func(n syntax.Node) bool {
		call, ok := n.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		callCount++
		words := extractCallWords(call)
		bin := filepath.Base(words[0])
		if r := checkApprovalFrictionCall(bin, words[1:]); r != "" {
			reasons = appendUnique(reasons, r)
		}
		return true
	})
	return callCount, reasons
}

func extractCallWords(call *syntax.CallExpr) []string {
	words := make([]string, 0, len(call.Args))
	printer := syntax.NewPrinter()
	for _, arg := range call.Args {
		var b strings.Builder
		_ = printer.Print(&b, arg)
		words = append(words, unquoteSimple(strings.TrimSpace(b.String())))
	}
	return words
}

func unquoteSimple(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}

func checkApprovalFrictionCall(bin string, args []string) string {
	if slices.Contains(broadCLITools, bin) {
		return fmt.Sprintf("broad CLI tool %q (constrain scope in a dedicated snippet)", bin)
	}
	if bin == "git" && slices.Contains(args, "-C") {
		return "infix variable argument (git -C <dir>) disrupts prefix approval"
	}
	if slices.Contains(searchFileTools, bin) && hasTrailingFileArg(args) {
		return fmt.Sprintf("trailing target file argument in %s (curry target file for persistent read approval)", bin)
	}
	return ""
}

func hasTrailingFileArg(args []string) bool {
	nonFlag := 0
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			nonFlag += len(args) - i - 1
			break
		}
		if strings.HasPrefix(a, "-") {
			if a == "-e" || a == "-f" || a == "-C" || a == "-A" || a == "-B" || a == "-m" {
				i++
			}
			continue
		}
		nonFlag++
	}
	return nonFlag >= 2
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

func containsBroadOrInterpreter(reasons []string) bool {
	for _, r := range reasons {
		if strings.Contains(r, "broad CLI tool") || strings.Contains(r, "inline_interpreter") {
			return true
		}
	}
	return false
}
