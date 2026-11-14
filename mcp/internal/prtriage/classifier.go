package prtriage

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

// ClassifyFiles inspects a list of modified file paths and verifies dependency manifest exclusivity.
func ClassifyFiles(in ClassifyFilesInput) (ClassifyFilesOutput, error) {
	for _, f := range in.Files {
		if err := guard.CheckPathAllowed(f); err != nil {
			return ClassifyFilesOutput{}, err
		}
	}

	manifests := []string{}
	excluded := []string{}
	unknowns := []string{}
	reasons := []string{}
	classified := make([]ClassifiedFile, 0, len(in.Files))
	ecosystems := make(map[Ecosystem]bool)

	if len(in.Files) == 0 {
		return ClassifyFilesOutput{
			OK:       true,
			Eligible: false,
			Reasons:  []string{"no files changed in PR diff"},
		}, nil
	}

	for _, path := range in.Files {
		clean := strings.ReplaceAll(filepath.Clean(strings.TrimSpace(path)), `\`, "/")
		base := filepath.Base(clean)

		cf := classifySingleFile(clean, base, in)
		classified = append(classified, cf)

		switch cf.Category {
		case CategoryManifest:
			manifests = append(manifests, clean)
			if cf.Ecosystem != "" {
				ecosystems[cf.Ecosystem] = true
			}
		case CategoryIgnored:
			// Ignored files (e.g. harmless metadata explicitly designated by agent) don't disqualify
		case CategoryUnknown:
			unknowns = append(unknowns, clean)
			excluded = append(excluded, clean)
			reasons = append(reasons, cf.Reason)
		default:
			excluded = append(excluded, clean)
			reasons = append(reasons, cf.Reason)
		}
	}

	resolvedEco := resolveEcosystem(ecosystems)
	eligible := len(excluded) == 0 && len(manifests) > 0

	return ClassifyFilesOutput{
		OK:            true,
		Eligible:      eligible,
		Ecosystem:     resolvedEco,
		ManifestFiles: manifests,
		ExcludedFiles: excluded,
		UnknownFiles:  unknowns,
		Files:         classified,
		Reasons:       reasons,
	}, nil
}

func classifySingleFile(clean, base string, in ClassifyFilesInput) ClassifiedFile {
	if catStr, ok := in.Overrides[clean]; ok {
		cat := normalizeCategory(catStr)
		eco := matchManifestEcosystem(clean, base)
		if cat == CategoryManifest && eco == "" {
			eco = EcosystemOther
		}
		return ClassifiedFile{
			Path:      clean,
			Category:  cat,
			Ecosystem: eco,
			Reason:    clean + " classified as " + string(cat) + " via agent override",
		}
	}

	if matchesCustomManifestPattern(clean, base, in.ExtraManifestPatterns) {
		return ClassifiedFile{
			Path:      clean,
			Category:  CategoryManifest,
			Ecosystem: EcosystemOther,
			Reason:    clean + " matches custom manifest pattern",
		}
	}

	matchedEco := matchManifestEcosystem(clean, base)
	if matchedEco != "" {
		return ClassifiedFile{
			Path:      clean,
			Category:  CategoryManifest,
			Ecosystem: matchedEco,
		}
	}

	return categorizeNonManifestFile(clean, base)
}

func categorizeNonManifestFile(clean, base string) ClassifiedFile {
	ext := strings.ToLower(filepath.Ext(base))

	if isTestFile(clean, base) {
		return ClassifiedFile{
			Path:     clean,
			Category: CategoryTest,
			Reason:   clean + " touches test files or suites (requires manual review)",
		}
	}
	if isCIWorkflowFile(clean) {
		return ClassifiedFile{
			Path:     clean,
			Category: CategoryCI,
			Reason:   clean + " touches CI workflows directly (requires manual review)",
		}
	}
	if sourceCodeExtensions[ext] {
		return ClassifiedFile{
			Path:     clean,
			Category: CategorySource,
			Reason:   clean + " touches application source code (disqualified per strict exclusivity rule)",
		}
	}
	if buildConfigBasenames[base] {
		cat := CategoryBuildConfig
		if ext == ".md" || base == "LICENSE" || base == "NOTICE" {
			cat = CategoryDoc
		}
		return ClassifiedFile{
			Path:     clean,
			Category: cat,
			Reason:   clean + " touches repository documentation or build config",
		}
	}

	return ClassifiedFile{
		Path:     clean,
		Category: CategoryUnknown,
		Reason: clean + " is unclassified (category unknown). " +
			"Agent can inspect file content and specify category via overrides or extra_manifest_patterns.",
	}
}

func normalizeCategory(c string) FileCategory {
	switch strings.ToLower(strings.TrimSpace(c)) {
	case "manifest":
		return CategoryManifest
	case "source":
		return CategorySource
	case "test":
		return CategoryTest
	case "ci":
		return CategoryCI
	case "build_config":
		return CategoryBuildConfig
	case "doc":
		return CategoryDoc
	case "ignored":
		return CategoryIgnored
	default:
		return CategoryUnknown
	}
}

func matchManifestEcosystem(relPath, base string) Ecosystem {
	for _, r := range manifestRules {
		if r.match(relPath, base) {
			return r.ecosystem
		}
	}
	return ""
}

func resolveEcosystem(ecosystems map[Ecosystem]bool) Ecosystem {
	switch len(ecosystems) {
	case 0:
		return EcosystemOther
	case 1:
		for eco := range ecosystems {
			return eco
		}
	}
	return EcosystemMixed
}

// FilterCandidates evaluates a batch list of candidate PRs against bot, label, title, and file rules.
func FilterCandidates(in FilterCandidatesInput) (FilterCandidatesOutput, error) {
	eligible := []EvaluatedPR{}
	disqualified := []EvaluatedPR{}

	for _, pr := range in.PRs {
		if err := guard.CheckNoSecrets(pr.Title, pr.Author, pr.HeadRefName, pr.Body); err != nil {
			return FilterCandidatesOutput{}, err
		}
		for _, f := range pr.Files {
			if err := guard.CheckPathAllowed(f); err != nil {
				return FilterCandidatesOutput{}, err
			}
		}

		eval, err := evaluateSingleCandidate(pr, in)
		if err != nil {
			return FilterCandidatesOutput{}, err
		}
		if eval.Eligible {
			eligible = append(eligible, eval)
		} else {
			disqualified = append(disqualified, eval)
		}
	}

	return FilterCandidatesOutput{
		OK:              true,
		TotalEvaluated:  len(in.PRs),
		EligibleCount:   len(eligible),
		ExcludedCount:   len(disqualified),
		EligiblePRs:     eligible,
		DisqualifiedPRs: disqualified,
	}, nil
}

func evaluateSingleCandidate(pr CandidatePR, in FilterCandidatesInput) (EvaluatedPR, error) {
	fileCheck, err := ClassifyFiles(ClassifyFilesInput{
		Files:                 pr.Files,
		Overrides:             in.FileOverrides,
		ExtraManifestPatterns: in.ExtraManifestPatterns,
	})
	if err != nil {
		return EvaluatedPR{}, err
	}

	reasons := []string{}
	isAllowed := slices.Contains(in.AllowedPRNumbers, pr.Number)
	isBot := isRecognizedBot(pr.Author, in.ExtraBotAuthors)
	hasDepLabel := hasDependencyLabel(pr.Labels, in.ExtraDepLabels)
	hasDepTitle := matchesDependencyTitle(pr.Title, in.ExtraTitlePatterns)

	unknownSignature := false
	if !isAllowed && !isBot && !hasDepLabel && !hasDepTitle {
		unknownSignature = true
		reasons = append(reasons, fmt.Sprintf(
			"author/title/labels do not match automated dependency update signatures "+
				"(agent can review and add PR #%d to allowed_pr_numbers)", pr.Number))
	}
	if !fileCheck.Eligible {
		reasons = append(reasons, fileCheck.Reasons...)
	}

	hasUnknownFiles := len(fileCheck.UnknownFiles) > 0
	isUnknown := unknownSignature || hasUnknownFiles

	return EvaluatedPR{
		Candidate: pr,
		Eligible:  len(reasons) == 0,
		FileCheck: fileCheck,
		Unknown:   isUnknown,
		Reasons:   reasons,
	}, nil
}
