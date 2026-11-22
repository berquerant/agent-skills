package snippetmgr

import (
	"fmt"
	"strings"
)

// BuildPreExecutionNotice formats the mandatory Pre-Execution Notice skeleton
// conforming to command-snippet and output-presentation conventions.
func BuildPreExecutionNotice(dualPath, absPath, purpose string, auditLevel int, auditRating string) string {
	purposeText := strings.TrimSpace(purpose)
	if purposeText == "" {
		purposeText = "<Explain what the script does and why a snippet was necessary>"
	}

	var b strings.Builder
	b.WriteString("### 📋 Snippet Pre-Execution Notice\n\n")
	fmt.Fprintf(&b, "- **Script Path**: %s\n", dualPath)
	fmt.Fprintf(&b, "- **Invocation**: `%s [arguments...]`\n", absPath)
	fmt.Fprintf(&b, "- **Intelligibility Audit**: Level %d (%s)\n", auditLevel, auditRating)
	fmt.Fprintf(&b, "- **Purpose**: %s\n", purposeText)
	b.WriteString("- **Parameters & Arguments**:\n")
	b.WriteString("  - <List each positional argument or flag and its value>\n")
	b.WriteString("- **Expected Side-Effects & Safety**:\n")
	b.WriteString("  - **Filesystem**: <Files created, edited, or deleted, or \"Read-only (no filesystem changes)\">\n")
	b.WriteString("  - **Environment / Network**: <Network calls, environment changes, or \"None\">\n")
	b.WriteString("  - **Reversibility**: <Reversible / Non-destructive / Irreversible>\n")
	return b.String()
}
