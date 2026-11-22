package snippet

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

type astCollector struct {
	metrics Metrics
	hazards []Hazard
}

func analyzeAST(in shellAnalyzeInput) (analysisResult, error) {
	variant := resolveLangVariant(in.lang)
	parser := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(variant))
	file, err := parser.Parse(strings.NewReader(in.code), "snippet")
	if err != nil {
		return analysisResult{}, err
	}

	c := &astCollector{
		metrics: Metrics{
			TotalLines:   in.counts.total,
			SLOC:         in.counts.sloc,
			CommentLines: in.counts.comments,
		},
	}
	c.walkFile(file, 0)
	return analysisResult{
		metrics: c.metrics,
		hazards: c.hazards,
	}, nil
}

func resolveLangVariant(lang string) syntax.LangVariant {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "sh", "posix":
		return syntax.LangPOSIX
	case "mksh":
		return syntax.LangMirBSDKorn
	default:
		return syntax.LangBash
	}
}

func (c *astCollector) walkFile(node syntax.Node, depth int) {
	if node == nil {
		return
	}
	if depth > c.metrics.MaxNestingDepth {
		c.metrics.MaxNestingDepth = depth
	}

	syntax.Walk(node, func(n syntax.Node) bool {
		if n == nil || n == node {
			return true
		}
		return c.inspectNode(n, depth)
	})
}

func (c *astCollector) inspectNode(n syntax.Node, depth int) bool {
	switch x := n.(type) {
	case *syntax.Stmt:
		if x.Background {
			c.metrics.BackgroundCount++
		}
	case *syntax.FuncDecl:
		c.recordFunction(x, depth)
		return false
	case *syntax.IfClause:
		c.recordIfClause(x, depth)
		return false
	case *syntax.ForClause, *syntax.WhileClause, *syntax.CaseClause:
		c.metrics.BranchCount++
		c.walkChildrenAtDepth(x, depth+1)
		return false
	case *syntax.Subshell:
		c.metrics.SubshellCount++
		c.walkChildrenAtDepth(x, depth+1)
		return false
	case *syntax.CoprocClause:
		c.metrics.BackgroundCount++
		c.walkChildrenAtDepth(x, depth+1)
		return false
	case *syntax.BinaryCmd:
		return c.inspectBinaryCmd(x, depth)
	case *syntax.CallExpr:
		c.inspectCallExpr(x)
	case *syntax.ProcSubst:
		c.addHazard(Hazard{
			Kind:   HazardProcessSubstitution,
			Line:   int(x.Pos().Line()),
			Detail: "process substitution <(...) or >(...)",
		})
	case *syntax.ParamExp:
		c.inspectParamExp(x)
	}
	return true
}

func (c *astCollector) walkChildrenAtDepth(compound syntax.Node, nextDepth int) {
	if nextDepth > c.metrics.MaxNestingDepth {
		c.metrics.MaxNestingDepth = nextDepth
	}
	switch x := compound.(type) {
	case *syntax.ForClause:
		for _, s := range x.Do {
			c.walkFile(s, nextDepth)
		}
	case *syntax.WhileClause:
		for _, s := range x.Do {
			c.walkFile(s, nextDepth)
		}
	case *syntax.CaseClause:
		for _, item := range x.Items {
			c.metrics.BranchCount++
			for _, s := range item.Stmts {
				c.walkFile(s, nextDepth)
			}
		}
	case *syntax.Subshell:
		for _, s := range x.Stmts {
			c.walkFile(s, nextDepth)
		}
	case *syntax.CoprocClause:
		c.walkFile(x.Stmt, nextDepth)
	}
}

func (c *astCollector) recordFunction(fn *syntax.FuncDecl, depth int) {
	c.metrics.FunctionCount++
	span := int(fn.End().Line()) - int(fn.Pos().Line()) + 1
	if span > c.metrics.MaxFunctionLines {
		c.metrics.MaxFunctionLines = span
	}
	c.walkFile(fn.Body, depth)
}

func (c *astCollector) recordIfClause(ic *syntax.IfClause, depth int) {
	c.metrics.BranchCount++
	nextDepth := depth + 1
	if nextDepth > c.metrics.MaxNestingDepth {
		c.metrics.MaxNestingDepth = nextDepth
	}
	for _, s := range ic.Cond {
		c.walkFile(s, depth)
	}
	for _, s := range ic.Then {
		c.walkFile(s, nextDepth)
	}
	if ic.Else != nil {
		if len(ic.Else.Cond) > 0 {
			c.recordIfClause(ic.Else, depth)
		} else {
			for _, s := range ic.Else.Then {
				c.walkFile(s, nextDepth)
			}
		}
	}
}

func (c *astCollector) inspectBinaryCmd(b *syntax.BinaryCmd, depth int) bool {
	switch b.Op {
	case syntax.AndStmt, syntax.OrStmt:
		c.metrics.BranchCount++
	case syntax.Pipe, syntax.PipeAll:
		pipeLen, rhsCmd := measurePipeline(b)
		if pipeLen > c.metrics.MaxPipelineLength {
			c.metrics.MaxPipelineLength = pipeLen
		}
		if isShellName(rhsCmd) {
			c.addHazard(Hazard{
				Kind:    HazardPipeToShell,
				Line:    int(b.Pos().Line()),
				Detail:  fmt.Sprintf("pipeline into shell (%s)", rhsCmd),
				HighSev: true,
			})
		}
	}
	c.walkFile(b.X, depth)
	c.walkFile(b.Y, depth)
	return false
}

func measurePipeline(b *syntax.BinaryCmd) (int, string) {
	pipes := 0
	var curr syntax.Node = b
	shellTarget := ""
	for {
		bin, ok := curr.(*syntax.BinaryCmd)
		if !ok || (bin.Op != syntax.Pipe && bin.Op != syntax.PipeAll) {
			break
		}
		pipes++
		if shellTarget == "" {
			if target := extractShellTarget(bin.Y); target != "" {
				shellTarget = target
			}
		}
		curr = bin.X.Cmd
	}
	return pipes, shellTarget
}

func extractShellTarget(stmt *syntax.Stmt) string {
	if stmt == nil {
		return ""
	}
	call, ok := stmt.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) == 0 {
		return ""
	}
	for i := 0; i < len(call.Args); i++ {
		w := filepath.Base(wordToString(call.Args[i]))
		if w == "sudo" || w == "env" {
			continue
		}
		if strings.HasPrefix(w, "-") || strings.Contains(w, "=") {
			continue
		}
		if isShellName(w) {
			return w
		}
		return ""
	}
	return ""
}

type callContext struct {
	cmd  string
	args []string
	line int
}

func (c *astCollector) inspectCallExpr(call *syntax.CallExpr) {
	if len(call.Args) == 0 {
		return
	}
	words := make([]string, 0, len(call.Args))
	for _, w := range call.Args {
		words = append(words, wordToString(w))
	}
	ctx := callContext{
		cmd:  filepath.Base(words[0]),
		args: words[1:],
		line: int(call.Pos().Line()),
	}

	switch ctx.cmd {
	case "eval":
		c.addHazard(Hazard{Kind: HazardEval, Line: ctx.line, Detail: "dynamic evaluation via eval", HighSev: true})
	case "source", ".":
		c.addHazard(Hazard{Kind: HazardSource, Line: ctx.line, Detail: fmt.Sprintf("external script sourcing via %s", ctx.cmd)})
	case "trap":
		c.metrics.TrapCount++
	case "wait", "parallel":
		c.metrics.BackgroundCount++
	case "xargs":
		if hasXargsParallelFlag(ctx.args) {
			c.metrics.BackgroundCount++
		}
	case "awk", "sed", "perl", "jq":
		c.checkDenseTextEngine(ctx)
	default:
		c.checkInlineInterpreter(ctx)
	}
}

func hasXargsParallelFlag(args []string) bool {
	for _, a := range args {
		if a == "-P" || (strings.HasPrefix(a, "-P") && len(a) > 2) || strings.HasPrefix(a, "--max-procs") {
			return true
		}
	}
	return false
}

const (
	denseTextEngineMinChars = 40
	inlinePayloadMinChars   = 50
)

func (c *astCollector) checkDenseTextEngine(ctx callContext) {
	for _, arg := range ctx.args {
		if len(arg) >= denseTextEngineMinChars {
			c.addHazard(Hazard{
				Kind:   HazardDenseTextEngine,
				Line:   ctx.line,
				Detail: fmt.Sprintf("dense %s expression (%d chars)", ctx.cmd, len(arg)),
			})
			return
		}
	}
}

func (c *astCollector) checkInlineInterpreter(ctx callContext) {
	interpreters := []string{"python", "python3", "node", "ruby", "bash", "sh", "zsh"}
	if !slices.Contains(interpreters, ctx.cmd) {
		return
	}
	for i := 0; i < len(ctx.args)-1; i++ {
		if ctx.args[i] == "-c" || ctx.args[i] == "-e" {
			payload := ctx.args[i+1]
			if isComplexInlinePayload(payload) {
				c.addHazard(Hazard{
					Kind:   HazardInlineInterpreter,
					Line:   ctx.line,
					Detail: fmt.Sprintf("inline %s %s script (%d chars)", ctx.cmd, ctx.args[i], len(payload)),
				})
				return
			}
		}
	}
}

func isComplexInlinePayload(payload string) bool {
	if len(payload) >= inlinePayloadMinChars ||
		strings.Contains(payload, ";") ||
		strings.Contains(payload, "\n") ||
		strings.Contains(payload, "import ") ||
		strings.Contains(payload, "require(") ||
		strings.Contains(payload, `\"`) ||
		strings.Contains(payload, `\'`) {
		return true
	}
	padded := " " + payload + " "
	for _, kw := range []string{" for ", " while ", " if ", " elif ", " else "} {
		if strings.Contains(padded, kw) {
			return true
		}
	}
	return false
}

func (c *astCollector) inspectParamExp(pe *syntax.ParamExp) {
	if pe.Param == nil {
		return
	}
	if n, err := strconv.Atoi(pe.Param.Value); err == nil && n > c.metrics.MaxPositionalArg {
		c.metrics.MaxPositionalArg = n
	}
	if pe.Repl != nil || pe.Excl || isObscureExpansion(pe.Exp) {
		c.addHazard(Hazard{
			Kind:   HazardObscureParamExpansion,
			Line:   int(pe.Pos().Line()),
			Detail: fmt.Sprintf("complex parameter expansion ${%s...}", pe.Param.Value),
		})
	}
}

func isObscureExpansion(exp *syntax.Expansion) bool {
	if exp == nil {
		return false
	}
	// Standard required-arg (${1:?msg}, ${1?msg}) and default-value (${var:-def}, ${var-def})
	// guards are recommended by command-snippet and are transparent to human reviewers.
	switch exp.Op.String() {
	case ":?", "?", ":-", "-":
		return false
	default:
		return true
	}
}

func (c *astCollector) addHazard(h Hazard) {
	c.hazards = append(c.hazards, h)
}

func wordToString(w *syntax.Word) string {
	var b strings.Builder
	printer := syntax.NewPrinter()
	_ = printer.Print(&b, w)
	s := strings.TrimSpace(b.String())
	s = strings.TrimPrefix(strings.TrimPrefix(s, `"`), `'`)
	s = strings.TrimSuffix(strings.TrimSuffix(s, `"`), `'`)
	return s
}

func isShellName(cmd string) bool {
	switch cmd {
	case "sh", "bash", "zsh", "mksh", "dash", "ksh":
		return true
	default:
		return false
	}
}
