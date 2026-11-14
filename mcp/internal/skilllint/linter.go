package skilllint

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const maxLintFileBytes = 2 * 1024 * 1024

// LintSkill runs automated quality checks A through K on a single skill directory.
func LintSkill(skillDir string, strict bool) (SkillReport, error) {
	skillDirAbs, err := filepath.Abs(skillDir)
	if err != nil {
		return SkillReport{}, fmt.Errorf("resolve skill path: %w", err)
	}
	skillRaw, err := readFileInRoot(skillDirAbs, "SKILL.md")
	if err != nil {
		return SkillReport{}, fmt.Errorf("%s/SKILL.md does not exist: %w", skillDir, err)
	}

	skillName := filepath.Base(skillDirAbs)
	skillsRoot := filepath.Dir(skillDirAbs)
	repoRoot := resolveRepoRoot(skillDirAbs)

	relFiles, allContents, mdContents, err := loadSkillDirectoryFiles(skillDirAbs)
	if err != nil {
		return SkillReport{}, err
	}

	skillContent := string(skillRaw)
	skillLines := countNewlines(skillRaw)
	lines := splitLines(skillContent)

	fm, fmErr := parseFrontmatter(lines)

	var findings []Finding
	if fmErr != nil {
		findings = append(findings, Finding{
			Check:    CheckB,
			Severity: SeverityError,
			File:     "SKILL.md",
			Message:  fmt.Sprintf("frontmatter YAML syntax error: %v", fmErr),
		})
	}
	findings = append(findings, checkAInventoryAndOrphans(skillLines, skillContent, relFiles, mdContents)...)
	findings = append(findings, checkBFrontmatterName(fm.Name, skillName)...)
	findings = append(findings, checkCFrontmatterDesc(fm.Description)...)
	findings = append(findings, checkDStructureAndVerify(lines)...)
	findings = append(findings, checkELinksAndPortability(skillDirAbs, mdContents)...)
	findings = append(findings, checkFScriptPermissions(skillDirAbs, relFiles)...)
	findings = append(findings, checkHRepoRegistration(repoRoot, skillName)...)
	findings = append(findings, checkKCommandAndPromptHygiene(mdContents)...)

	siblingRefs, crossLinks := checkGSiblingReferences(skillsRoot, repoRoot, skillName, allContents)
	dupHeadings := checkIDuplicateHeadings(skillDirAbs, skillsRoot, repoRoot, skillName)
	extURLs := checkJExternalURLs(allContents)

	errCount, warnCount := countSeverities(findings)
	passed := errCount == 0 && (!strict || warnCount == 0)

	if findings == nil {
		findings = []Finding{}
	}
	return SkillReport{
		OK:                true,
		Passed:            passed,
		Strict:            strict,
		SkillName:         skillName,
		TargetDir:         skillDir,
		SkillLines:        skillLines,
		Errors:            errCount,
		Warnings:          warnCount,
		Findings:          findings,
		SiblingRefs:       siblingRefs,
		CrossSkillLinks:   crossLinks,
		DuplicateHeadings: dupHeadings,
		ExternalURLs:      extURLs,
	}, nil
}

// LintAll discovers and lints all skill directories under rootDir.
func LintAll(rootDir string, strict bool) (BatchReport, error) {
	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return BatchReport{}, fmt.Errorf("resolve root dir: %w", err)
	}
	skillDirs, err := discoverSkillDirs(rootAbs)
	if err != nil {
		return BatchReport{}, err
	}

	reports := make([]SkillReport, 0, len(skillDirs))
	passedSkills := 0
	failedSkills := 0
	totalErrors := 0
	totalWarnings := 0

	for _, dir := range skillDirs {
		rep, lintErr := LintSkill(dir, strict)
		if lintErr != nil {
			return BatchReport{}, lintErr
		}
		reports = append(reports, rep)
		totalErrors += rep.Errors
		totalWarnings += rep.Warnings
		if rep.Passed {
			passedSkills++
		} else {
			failedSkills++
		}
	}

	return BatchReport{
		OK:            true,
		Passed:        failedSkills == 0,
		Strict:        strict,
		RootDir:       rootDir,
		TotalSkills:   len(reports),
		PassedSkills:  passedSkills,
		FailedSkills:  failedSkills,
		TotalErrors:   totalErrors,
		TotalWarnings: totalWarnings,
		Reports:       reports,
	}, nil
}

func loadSkillDirectoryFiles(skillDirAbs string) ([]string, map[string]string, map[string]string, error) {
	root, err := os.OpenRoot(skillDirAbs)
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() { _ = root.Close() }()

	var relFiles []string
	allContents := make(map[string]string)
	mdContents := make(map[string]string)

	err = filepath.WalkDir(skillDirAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(skillDirAbs, path)
		if relErr != nil {
			return relErr
		}
		relSlash := filepath.ToSlash(rel)
		relFiles = append(relFiles, relSlash)
		info, infoErr := d.Info()
		if infoErr != nil || !info.Mode().IsRegular() || info.Size() > maxLintFileBytes {
			return nil
		}
		raw, readErr := root.ReadFile(rel)
		if readErr == nil {
			content := string(raw)
			allContents[relSlash] = content
			if strings.HasSuffix(relSlash, ".md") {
				mdContents[relSlash] = content
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}
	slices.Sort(relFiles)
	return relFiles, allContents, mdContents, nil
}

func discoverSkillDirs(rootAbs string) ([]string, error) {
	info, err := os.Stat(rootAbs)
	if err != nil {
		return nil, fmt.Errorf("stat root dir: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", rootAbs)
	}

	if _, err := os.Stat(filepath.Join(rootAbs, "SKILL.md")); err == nil {
		return []string{rootAbs}, nil
	}

	candidates := []string{
		filepath.Join(rootAbs, "skills"),
		filepath.Join(rootAbs, ".agent", "skills"),
		rootAbs,
	}
	seen := make(map[string]bool)
	var dirs []string
	for _, parent := range candidates {
		entries, readErr := os.ReadDir(parent)
		if readErr != nil {
			continue
		}
		for _, d := range entries {
			if !d.IsDir() {
				continue
			}
			cand := filepath.Clean(filepath.Join(parent, d.Name()))
			if _, statErr := os.Stat(filepath.Join(cand, "SKILL.md")); statErr == nil && !seen[cand] {
				seen[cand] = true
				dirs = append(dirs, cand)
			}
		}
	}
	slices.Sort(dirs)
	return dirs, nil
}

func countSeverities(findings []Finding) (int, int) {
	errs := 0
	warns := 0
	for _, f := range findings {
		switch f.Severity {
		case SeverityError:
			errs++
		case SeverityWarning:
			warns++
		}
	}
	return errs, warns
}
