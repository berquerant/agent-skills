package skilllint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

var (
	absMachinePathRegex  = regexp.MustCompile(`\]\((/|file://)|/Users/|/home/[a-z]|[A-Z]:[/\\]`)
	externalURLRegex     = regexp.MustCompile(`https?://[^\s)>` + "`" + `"]+`)
	backtickRawTargetRe  = regexp.MustCompile(`\]\([^)]+\)\s+\(` + "`[^`]+`" + `\)`)
	manualCoTPhraseRegex = regexp.MustCompile(`(?i)think step[- ]by[- ]step`)
)

func checkELinksAndPortability(skillDirAbs string, mdContents map[string]string) []Finding {
	var findings []Finding
	relFiles := sortedMapKeys(mdContents)

	for _, rel := range relFiles {
		lines := splitLines(mdContents[rel])
		mdAbsDir := filepath.Dir(filepath.Join(skillDirAbs, rel))
		for _, target := range extractRelativeLinkTargets(lines) {
			resolved := filepath.Join(mdAbsDir, target)
			if _, err := os.Stat(resolved); err != nil {
				findings = append(findings, Finding{
					Check:    CheckE,
					Severity: SeverityError,
					File:     rel,
					Message:  fmt.Sprintf("BROKEN link: %s -> %s", rel, target),
				})
			}
		}
	}

	for _, rel := range relFiles {
		if filepath.Base(rel) == "automated-checks.md" {
			continue
		}
		for idx, line := range splitLines(mdContents[rel]) {
			if absMachinePathRegex.MatchString(line) {
				findings = append(findings, Finding{
					Check:    CheckE,
					Severity: SeverityError,
					File:     rel,
					Line:     idx + 1,
					Message:  fmt.Sprintf("ABSOLUTE/MACHINE path found in %s:%d", rel, idx+1),
				})
			}
			for _, sec := range guard.ScanTextForSecrets(line) {
				findings = append(findings, Finding{
					Check:    CheckE,
					Severity: SeverityError,
					File:     rel,
					Line:     idx + 1,
					Message:  fmt.Sprintf("CREDENTIAL/SECRET pattern detected (%s) in %s:%d", sec, rel, idx+1),
				})
			}
		}
	}
	return findings
}

func checkFScriptPermissions(skillDirAbs string, relFiles []string) []Finding {
	var findings []Finding
	sortedRel := slices.Clone(relFiles)
	slices.Sort(sortedRel)
	for _, rel := range sortedRel {
		if !strings.HasPrefix(rel, "scripts/") && !strings.HasPrefix(rel, "scripts\\") {
			continue
		}
		info, err := os.Stat(filepath.Join(skillDirAbs, rel))
		if err != nil || info.IsDir() {
			continue
		}
		if info.Mode()&0o100 == 0 {
			findings = append(findings, Finding{
				Check:    CheckF,
				Severity: SeverityError,
				File:     rel,
				Message:  fmt.Sprintf("NON-EXECUTABLE script: %s", rel),
			})
		}
	}
	return findings
}

func checkJExternalURLs(allContents map[string]string) []string {
	var urls []string
	for _, rel := range sortedMapKeys(allContents) {
		for idx, line := range splitLines(allContents[rel]) {
			for _, u := range externalURLRegex.FindAllString(line, -1) {
				urls = append(urls, fmt.Sprintf("%s:%d:%s", rel, idx+1, u))
			}
		}
	}
	return urls
}

func checkKCommandAndPromptHygiene(mdContents map[string]string) []Finding {
	var findings []Finding
	for _, rel := range sortedMapKeys(mdContents) {
		base := filepath.Base(rel)
		if base == "automated-checks.md" || base == "skill-review-checklist.md" {
			continue
		}
		findings = append(findings, inspectFileHygiene(rel, splitLines(mdContents[rel]))...)
	}
	return findings
}

func inspectFileHygiene(rel string, lines []string) []Finding {
	var findings []Finding
	inCode := false
	inSh := false

	for idx, line := range lines {
		lineNum := idx + 1
		if isShellFenceStart(line) && !inCode {
			inCode = true
			inSh = true
			continue
		}
		if isFenceDelimiter(line) {
			inCode = !inCode
			inSh = false
			continue
		}
		if inSh {
			trimmed := strings.TrimLeft(line, " \t")
			if !strings.HasPrefix(trimmed, "#") && len(line) >= 120 {
				findings = append(findings, Finding{
					Check:    CheckK,
					Severity: SeverityWarning,
					File:     rel,
					Line:     lineNum,
					Message:  fmt.Sprintf("long inline shell command (%d chars >= 120) in %s:%d", len(line), rel, lineNum),
				})
			}
			continue
		}
		if inCode {
			continue
		}
		findings = append(findings, inspectProseHygieneLine(rel, lineNum, line)...)
	}
	return findings
}

func inspectProseHygieneLine(rel string, lineNum int, line string) []Finding {
	var findings []Finding
	if backtickRawTargetRe.MatchString(line) {
		findings = append(findings, Finding{
			Check:    CheckK,
			Severity: SeverityWarning,
			File:     rel,
			Line:     lineNum,
			Message:  fmt.Sprintf("backtick-wrapped raw target in co-presented link in %s:%d", rel, lineNum),
		})
	}
	cleaned := stripInlineCode(line)
	if manualCoTPhraseRegex.MatchString(cleaned) {
		findings = append(findings, Finding{
			Check:    CheckK,
			Severity: SeverityWarning,
			File:     rel,
			Line:     lineNum,
			Message:  fmt.Sprintf("manual chain-of-thought phrasing in %s:%d", rel, lineNum),
		})
	}
	return findings
}

func sortedMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
