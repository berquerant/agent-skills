package skilllint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var crossSkillRelRegex = regexp.MustCompile(`\.\./[a-z0-9-]+/`)

func checkGSiblingReferences(
	skillsRoot, repoRoot, skillName string,
	allContents map[string]string,
) (map[string]int, []string) {
	siblings := listSiblingSkillNames(skillsRoot, repoRoot, skillName)
	siblingCounts := make(map[string]int)

	for _, sib := range siblings {
		re := regexp.MustCompile(`(^|[^a-zA-Z0-9_-])` + regexp.QuoteMeta(sib) + `([^a-zA-Z0-9_-]|$)`)
		count := 0
		for _, content := range allContents {
			for _, line := range splitLines(content) {
				if re.MatchString(line) {
					count++
				}
			}
		}
		if count > 0 {
			siblingCounts[sib] = count
		}
	}

	var crossLinks []string
	for _, rel := range sortedMapKeys(allContents) {
		for idx, line := range splitLines(allContents[rel]) {
			for _, m := range crossSkillRelRegex.FindAllString(line, -1) {
				crossLinks = append(crossLinks, fmt.Sprintf("%s:%d:%s", rel, idx+1, m))
			}
		}
	}
	return siblingCounts, crossLinks
}

func listSiblingSkillNames(skillsRoot, repoRoot, currentSkill string) []string {
	seen := make(map[string]bool)
	for _, parent := range []string{filepath.Join(repoRoot, "skills"), skillsRoot} {
		entries, err := os.ReadDir(parent)
		if err != nil {
			continue
		}
		for _, d := range entries {
			name := d.Name()
			if !d.IsDir() || name == currentSkill || !skillNameRegex.MatchString(name) {
				continue
			}
			seen[name] = true
		}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	slices.Sort(out)
	return out
}

func checkHRepoRegistration(repoRoot, skillName string) []Finding {
	docs := []string{
		"README.md",
		"AGENTS.md",
		filepath.Join("skills", "skill-router", "references", "intent-matrix.md"),
	}
	var findings []Finding
	for _, doc := range docs {
		raw, err := readFileInRoot(repoRoot, doc)
		if err != nil {
			continue
		}
		if strings.Contains(string(raw), skillName) {
			continue
		}
		if doc == filepath.Join("skills", "skill-router", "references", "intent-matrix.md") {
			if skillName == "skill-router" {
				continue
			}
			findings = append(findings, Finding{
				Check:    CheckH,
				Severity: SeverityWarning,
				File:     doc,
				Message:  fmt.Sprintf("WARN: %s is not listed in %s", skillName, filepath.ToSlash(doc)),
			})
			continue
		}
		findings = append(findings, Finding{
			Check:    CheckH,
			Severity: SeverityWarning,
			File:     doc,
			Message:  fmt.Sprintf("WARN: %s is not registered in %s", skillName, filepath.ToSlash(doc)),
		})
	}
	return findings
}

func checkIDuplicateHeadings(skillDirAbs, skillsRoot, repoRoot, skillName string) []string {
	if _, err := os.Stat(filepath.Join(skillDirAbs, "references")); err != nil {
		return nil
	}

	headingSkills := make(map[string][]string)
	inTarget := make(map[string]bool)

	for _, refFile := range collectReferenceMarkdownFiles(skillsRoot, repoRoot) {
		ownerSkill := filepath.Base(filepath.Dir(filepath.Dir(refFile)))
		parentDir := filepath.Dir(refFile)
		raw, err := readFileInRoot(parentDir, filepath.Base(refFile))
		if err != nil {
			continue
		}
		for _, line := range extractMarkdownH2H3Headings(splitLines(string(raw))) {
			if !slices.Contains(headingSkills[line], ownerSkill) {
				headingSkills[line] = append(headingSkills[line], ownerSkill)
			}
			if ownerSkill == skillName {
				inTarget[line] = true
			}
		}
	}

	var duplicates []string
	for heading, owners := range headingSkills {
		if len(owners) > 1 && inTarget[heading] {
			duplicates = append(duplicates, fmt.Sprintf("%s -> %s", heading, strings.Join(owners, " ")))
		}
	}
	slices.Sort(duplicates)
	return duplicates
}

func extractMarkdownH2H3Headings(lines []string) []string {
	inCode := false
	var headings []string
	for _, line := range lines {
		if isFenceDelimiter(line) {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") {
			headings = append(headings, line)
		}
	}
	return headings
}

func collectReferenceMarkdownFiles(skillsRoot, repoRoot string) []string {
	seen := make(map[string]bool)
	var files []string
	for _, parent := range []string{filepath.Join(repoRoot, "skills"), skillsRoot} {
		skillDirs, err := os.ReadDir(parent)
		if err != nil {
			continue
		}
		for _, sd := range skillDirs {
			if !sd.IsDir() {
				continue
			}
			refDir := filepath.Join(parent, sd.Name(), "references")
			refEntries, err := os.ReadDir(refDir)
			if err != nil {
				continue
			}
			for _, re := range refEntries {
				if re.IsDir() || !strings.HasSuffix(re.Name(), ".md") {
					continue
				}
				full := filepath.Clean(filepath.Join(refDir, re.Name()))
				if !seen[full] {
					seen[full] = true
					files = append(files, full)
				}
			}
		}
	}
	slices.Sort(files)
	return files
}
