package skilllint

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var (
	skillNameRegex     = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	triggerOpenerRegex = regexp.MustCompile(`^Use this skill (when|to|for) `)
	step1HeadingRegex  = regexp.MustCompile(`^## Step 1([:[:space:]]|$)`)
)

func checkAInventoryAndOrphans(
	skillLines int,
	skillContent string,
	relFiles []string,
	mdContents map[string]string,
) []Finding {
	var findings []Finding
	if skillLines > 500 {
		findings = append(findings, Finding{
			Check:    CheckA,
			Severity: SeverityError,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("size: FAIL (%d lines, exceeds 500)", skillLines),
		})
	} else if skillLines < 80 || skillLines > 250 {
		findings = append(findings, Finding{
			Check:    CheckA,
			Severity: SeverityWarning,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("size: WARN (%d lines, recommended 80-250)", skillLines),
		})
	}

	sortedRel := slices.Clone(relFiles)
	slices.Sort(sortedRel)
	for _, rel := range sortedRel {
		if rel == "SKILL.md" {
			continue
		}
		if strings.Contains(skillContent, rel) {
			continue
		}
		base := filepath.Base(rel)
		if isReferencedInOtherMarkdown(rel, base, mdContents) {
			findings = append(findings, Finding{
				Check:    CheckA,
				Severity: SeverityWarning,
				File:     rel,
				Message:  fmt.Sprintf("multi-hop reference (not linked directly from SKILL.md): %s", rel),
			})
		} else {
			findings = append(findings, Finding{
				Check:    CheckA,
				Severity: SeverityWarning,
				File:     rel,
				Message:  fmt.Sprintf("unreferenced file: %s", rel),
			})
		}
	}
	return findings
}

func isReferencedInOtherMarkdown(targetRel, base string, mdContents map[string]string) bool {
	for rel, content := range mdContents {
		if rel == targetRel {
			continue
		}
		if strings.Contains(content, base) {
			return true
		}
	}
	return false
}

func checkBFrontmatterName(fmName, dirName string) []Finding {
	var findings []Finding
	nameLen := len(fmName)
	if fmName == "" {
		findings = append(findings, Finding{
			Check:    CheckB,
			Severity: SeverityError,
			File:     "SKILL.md",
			Message:  "pattern: FAIL (missing or empty name)",
		})
	} else if nameLen > 64 {
		findings = append(findings, Finding{
			Check:    CheckB,
			Severity: SeverityError,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("pattern: FAIL (exceeds 64 chars: %d)", nameLen),
		})
	} else if !skillNameRegex.MatchString(fmName) {
		findings = append(findings, Finding{
			Check:    CheckB,
			Severity: SeverityError,
			File:     "SKILL.md",
			Message:  "pattern: FAIL",
		})
	}

	if nameLen < 3 || nameLen > 30 {
		findings = append(findings, Finding{
			Check:    CheckB,
			Severity: SeverityWarning,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("name length: WARN (%d chars, recommended 3-30)", nameLen),
		})
	}

	if fmName != dirName {
		findings = append(findings, Finding{
			Check:    CheckB,
			Severity: SeverityError,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("dir match: FAIL (expected %s)", dirName),
		})
	}
	return findings
}

func checkCFrontmatterDesc(desc string) []Finding {
	var findings []Finding
	descLen := len(desc)
	if descLen == 0 {
		findings = append(findings, Finding{
			Check:    CheckC,
			Severity: SeverityError,
			File:     "SKILL.md",
			Message:  "description: FAIL (empty)",
		})
	} else if descLen > 1024 {
		findings = append(findings, Finding{
			Check:    CheckC,
			Severity: SeverityError,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("description: FAIL (exceeds 1024 chars: %d)", descLen),
		})
	} else if descLen < 100 || descLen > 350 {
		findings = append(findings, Finding{
			Check:    CheckC,
			Severity: SeverityWarning,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("description: WARN (%d chars, recommended 100-350 for Layer 1 token economy)", descLen),
		})
	}

	if !triggerOpenerRegex.MatchString(desc) {
		findings = append(findings, Finding{
			Check:    CheckC,
			Severity: SeverityWarning,
			File:     "SKILL.md",
			Message:  "trigger opener: WARN (recommended opener: 'Use this skill when/to/for ...')",
		})
	}
	return findings
}

func checkDStructureAndVerify(lines []string) []Finding {
	step1Line, lastH2, missingVerifySteps := scanMarkdownHeadingsAndSteps(lines)
	var findings []Finding
	for _, step := range missingVerifySteps {
		findings = append(findings, Finding{
			Check:    CheckD,
			Severity: SeverityError,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("MISSING Verify: %s", step),
		})
	}
	if step1Line > 100 {
		findings = append(findings, Finding{
			Check:    CheckD,
			Severity: SeverityWarning,
			File:     "SKILL.md",
			Line:     step1Line,
			Message:  fmt.Sprintf("top-loaded workflow: WARN (Step 1 at line %d > 100)", step1Line),
		})
	}
	if lastH2 != "## Guidelines" {
		shown := lastH2
		if shown == "" {
			shown = "none"
		}
		findings = append(findings, Finding{
			Check:    CheckD,
			Severity: SeverityWarning,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("final section: WARN (expected '## Guidelines' as final ## section, found '%s')", shown),
		})
	}
	return findings
}

func scanMarkdownHeadingsAndSteps(lines []string) (int, string, []string) {
	inCode := false
	step1Line := 0
	lastH2 := ""
	currentStep := ""
	hasVerify := false
	var missing []string

	for idx, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		if step1Line == 0 && step1HeadingRegex.MatchString(line) {
			step1Line = idx + 1
		}
		if strings.HasPrefix(line, "## ") {
			if currentStep != "" && !hasVerify {
				missing = append(missing, currentStep)
			}
			lastH2 = line
			currentStep = ""
			if strings.HasPrefix(line, "## Step ") {
				currentStep = line
			}
			hasVerify = false
			continue
		}
		if strings.Contains(line, "Verify:") {
			hasVerify = true
		}
	}
	if currentStep != "" && !hasVerify {
		missing = append(missing, currentStep)
	}
	return step1Line, lastH2, missing
}
