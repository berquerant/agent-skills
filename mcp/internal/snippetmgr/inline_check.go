package snippetmgr

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

const (
	maxInlineChars          = 120
	maxInlineLines          = 3
	maxInlineChainedCmds    = 4
	maxInlinePayloadChars   = 50
	maxDenseTextEngineChars = 40
)

var (
	inlineInterpreters = []string{"python", "python3", "node", "ruby", "perl", "bash", "sh", "zsh"}
	broadCLITools      = []string{"gh", "aws", "gcloud", "kubectl"}
	textEngineTools    = []string{"awk", "sed", "perl"}
	searchFileTools    = []string{"grep", "egrep", "fgrep", "rg"}
)

// CheckInline evaluates an inline command against command-snippet Step 1 avoidance thresholds.
func CheckInline(in CheckInlineInput) CheckInlineOutput {
	cmd := strings.TrimSpace(in.Command)
	if cmd == "" {
		return CheckInlineOutput{
			ShouldExtract:     false,
			Reasons:           []string{},
			SuggestedLanguage: "bash",
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

	chained, astReasons, suggestedLang := inspectInlineAST(cmd)
	if chained >= maxInlineChainedCmds {
		reasons = append(reasons, fmt.Sprintf("chained commands %d >= %d", chained, maxInlineChainedCmds))
	}
	reasons = append(reasons, astReasons...)

	if reasons == nil {
		reasons = []string{}
	}
	if suggestedLang == "" {
		suggestedLang = "bash"
	}

	out := CheckInlineOutput{
		ShouldExtract:     len(reasons) > 0,
		Reasons:           reasons,
		CharCount:         charCount,
		LineCount:         lineCount,
		ChainedCommands:   chained,
		SuggestedLanguage: suggestedLang,
	}
	if out.ShouldExtract && !containsBroadOrInterpreter(reasons) {
		out.LightweightHint = "If strictly read-only and local, consider splitting into simple <120-char commands first; otherwise create a snippet."
	}
	return out
}

func countNonEmptyLines(s string) int {
	lines := strings.Split(s, "\n")
	count := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			count++
		}
	}
	return count
}

func inspectInlineAST(cmd string) (int, []string, string) {
	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))
	file, err := parser.Parse(strings.NewReader(cmd), "inline")
	if err != nil {
		return countHeuristicChains(cmd), []string{"shell syntax parse error (extract to inspectable snippet)"}, "bash"
	}

	var (
		callCount     int
		reasons       []string
		suggestedLang = "bash"
	)

	syntax.Walk(file, func(n syntax.Node) bool {
		call, ok := n.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		callCount++
		words := extractCallWords(call)
		bin := filepath.Base(words[0])
		args := words[1:]

		if r, lang := checkInlineInterpreterCall(bin, args); r != "" {
			reasons = appendUnique(reasons, r)
			if lang != "" {
				suggestedLang = lang
			}
		}
		if r := checkTextEngineCall(bin, args); r != "" {
			reasons = appendUnique(reasons, r)
		}
		if r := checkApprovalFrictionCall(bin, args); r != "" {
			reasons = appendUnique(reasons, r)
		}
		return true
	})

	return callCount, reasons, suggestedLang
}

func extractCallWords(call *syntax.CallExpr) []string {
	words := make([]string, 0, len(call.Args))
	printer := syntax.NewPrinter()
	for _, arg := range call.Args {
		var b strings.Builder
		_ = printer.Print(&b, arg)
		w := strings.TrimSpace(b.String())
		w = unquoteSimple(w)
		words = append(words, w)
	}
	return words
}

func unquoteSimple(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}

func checkInlineInterpreterCall(bin string, args []string) (string, string) {
	if !slices.Contains(inlineInterpreters, bin) {
		return "", ""
	}
	for i := 0; i < len(args)-1; i++ {
		if args[i] != "-c" && args[i] != "-e" {
			continue
		}
		payload := args[i+1]
		if trigger := inlinePayloadTrigger(payload); trigger != "" {
			lang := "bash"
			if bin == "python" || bin == "python3" {
				lang = "python"
			}
			return fmt.Sprintf("inline interpreter %s %s (%s)", bin, args[i], trigger), lang
		}
	}
	return "", ""
}

func inlinePayloadTrigger(payload string) string {
	switch {
	case len(payload) >= maxInlinePayloadChars:
		return fmt.Sprintf("payload length %d >= %d", len(payload), maxInlinePayloadChars)
	case strings.Contains(payload, ";") || strings.Contains(payload, "\n"):
		return "multiple statements"
	case strings.Contains(payload, "import ") || strings.Contains(payload, "require("):
		return "module import"
	case containsControlFlow(payload):
		return "control flow"
	case strings.Contains(payload, `\"`) || strings.Contains(payload, `\'`):
		return "escaped quotes"
	default:
		return ""
	}
}

func containsControlFlow(s string) bool {
	padded := " " + s + " "
	for _, kw := range []string{" for ", " while ", " if ", " elif ", " else "} {
		if strings.Contains(padded, kw) {
			return true
		}
	}
	return false
}

func checkTextEngineCall(bin string, args []string) string {
	if !slices.Contains(textEngineTools, bin) {
		return ""
	}
	for _, arg := range args {
		if len(arg) >= maxDenseTextEngineChars || strings.Contains(arg, "\n") {
			return fmt.Sprintf("inline %s script (%d chars)", bin, len(arg))
		}
	}
	return ""
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
		if strings.Contains(r, "broad CLI tool") || strings.Contains(r, "inline interpreter") {
			return true
		}
	}
	return false
}
