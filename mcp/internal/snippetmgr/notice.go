package snippetmgr

import (
	"fmt"
	"strings"
)

// NoticeParams holds inputs for formatting a Snippet Pre-Execution Notice skeleton.
type NoticeParams struct {
	DualPath    string
	AbsPath     string
	Purpose     string
	AuditLevel  int
	AuditRating string
}

// BuildPreExecutionNotice formats the mandatory Pre-Execution Notice skeleton
// conforming to command-snippet and output-presentation conventions.
func BuildPreExecutionNotice(p NoticeParams) string {
	purposeText := strings.TrimSpace(p.Purpose)
	if purposeText == "" {
		purposeText = "<Explain what the script does and why a snippet was necessary>"
	}

	var b strings.Builder
	b.WriteString("### 📋 Snippet Pre-Execution Notice\n\n")
	fmt.Fprintf(&b, "- **Script Path**: %s\n", p.DualPath)
	fmt.Fprintf(&b, "- **Invocation**: `%s [arguments...]`\n", p.AbsPath)
	fmt.Fprintf(&b, "- **Intelligibility Audit**: Level %d (%s)\n", p.AuditLevel, p.AuditRating)
	fmt.Fprintf(&b, "- **Purpose**: %s\n", purposeText)
	b.WriteString("- **Parameters & Arguments**:\n")
	b.WriteString("  - <List each positional argument or flag and its value>\n")
	b.WriteString("- **Expected Side-Effects & Safety**:\n")
	b.WriteString("  - **Filesystem**: <Files created, edited, or deleted, or \"Read-only (no filesystem changes)\">\n")
	b.WriteString("  - **Environment / Network**: <Network calls, environment changes, or \"None\">\n")
	b.WriteString("  - **Reversibility**: <Reversible / Non-destructive / Irreversible>\n")
	return b.String()
}
