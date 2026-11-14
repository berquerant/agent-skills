package feature

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"
	"text/template"
)

//go:embed help.tmpl
var helpTemplateSource string

var (
	helpTmplOnce sync.Once
	helpTmpl     *template.Template
)

func getHelpTemplate() *template.Template {
	helpTmplOnce.Do(func() {
		helpTmpl = template.Must(template.New("help").Parse(helpTemplateSource))
	})
	return helpTmpl
}

// HelpMarkdown returns the markdown help body rendered from help.tmpl.
func HelpMarkdown() string {
	var sb strings.Builder
	_ = getHelpTemplate().Execute(&sb, nil)
	return sb.String()
}

// Usage returns the complete CLI usage and flags markdown block.
func Usage() string {
	return fmt.Sprintf("%s\n```\n%s\n```", HelpMarkdown(), FlagUsages())
}
