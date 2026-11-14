package prtriage

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

var knownBotAuthors = map[string]struct{}{
	"app/dependabot":        {},
	"dependabot[bot]":       {},
	"dependabot":            {},
	"renovate[bot]":         {},
	"renovate":              {},
	"snyk-bot":              {},
	"security-bot":          {},
	"gitlab-dependency-bot": {},
}

var knownDepLabels = map[string]struct{}{
	"dependencies":        {},
	"dependencies-update": {},
	"security":            {},
	"cve":                 {},
	"deps":                {},
	"dependency":          {},
}

type manifestRule struct {
	ecosystem Ecosystem
	match     func(relPath, base string) bool
}

var manifestRules = []manifestRule{
	{
		ecosystem: EcosystemNode,
		match: func(_, base string) bool {
			return base == "package.json" || base == "package-lock.json" ||
				base == "pnpm-lock.yaml" || base == "yarn.lock" || base == "bun.lockb"
		},
	},
	{
		ecosystem: EcosystemGo,
		match: func(_, base string) bool {
			return base == "go.mod" || base == "go.sum"
		},
	},
	{
		ecosystem: EcosystemRust,
		match: func(_, base string) bool {
			return base == "Cargo.toml" || base == "Cargo.lock"
		},
	},
	{
		ecosystem: EcosystemPython,
		match: func(_, base string) bool {
			return base == "pyproject.toml" || base == "poetry.lock" ||
				base == "Pipfile" || base == "Pipfile.lock" ||
				base == "requirements.txt" || (strings.HasPrefix(base, "requirements") && strings.HasSuffix(base, ".txt"))
		},
	},
	{
		ecosystem: EcosystemRuby,
		match: func(_, base string) bool {
			return base == "Gemfile" || base == "Gemfile.lock"
		},
	},
	{
		ecosystem: EcosystemPHP,
		match: func(_, base string) bool {
			return base == "composer.json" || base == "composer.lock"
		},
	},
	{
		ecosystem: EcosystemJava,
		match: func(_, base string) bool {
			return base == "pom.xml" || base == "build.gradle" ||
				base == "build.gradle.kts" || base == "gradle.properties"
		},
	},
	{
		ecosystem: EcosystemDotNet,
		match: func(_, base string) bool {
			return strings.HasSuffix(base, ".csproj") || base == "packages.lock.json" || base == "Directory.Packages.props"
		},
	},
	{
		ecosystem: EcosystemAction,
		match: func(relPath, base string) bool {
			return base == "actions.lock" || strings.HasSuffix(relPath, ".github/workflows/actions.lock")
		},
	},
}

// ClassifyFiles inspects a list of modified file paths and verifies dependency manifest exclusivity.
func ClassifyFiles(files []string) (ClassifyFilesOutput, error) {
	for _, f := range files {
		if err := guard.CheckPathAllowed(f); err != nil {
			return ClassifyFilesOutput{}, err
		}
	}

	manifests := []string{}
	excluded := []string{}
	reasons := []string{}
	ecosystems := make(map[Ecosystem]bool)

	if len(files) == 0 {
		return ClassifyFilesOutput{
			OK:       true,
			Eligible: false,
			Reasons:  []string{"no files changed in PR diff"},
		}, nil
	}

	for _, path := range files {
		clean := strings.ReplaceAll(filepath.Clean(strings.TrimSpace(path)), `\`, "/")
		base := filepath.Base(clean)

		matchedEco := matchManifestEcosystem(clean, base)
		if matchedEco != "" {
			manifests = append(manifests, clean)
			ecosystems[matchedEco] = true
			continue
		}

		excluded = append(excluded, clean)
		reasons = append(reasons, formatDisqualificationReason(clean, base))
	}

	resolvedEco := resolveEcosystem(ecosystems)
	eligible := len(excluded) == 0 && len(manifests) > 0

	return ClassifyFilesOutput{
		OK:            true,
		Eligible:      eligible,
		Ecosystem:     resolvedEco,
		ManifestFiles: manifests,
		ExcludedFiles: excluded,
		Reasons:       reasons,
	}, nil
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

var sourceCodeExtensions = map[string]bool{
	".go":   true,
	".ts":   true,
	".js":   true,
	".py":   true,
	".rs":   true,
	".java": true,
	".c":    true,
	".cpp":  true,
}

var buildConfigBasenames = map[string]bool{
	"README.md":    true,
	"Dockerfile":   true,
	"Makefile":     true,
	"Taskfile.yml": true,
	"mise.toml":    true,
}

func isTestFile(clean, base string) bool {
	return strings.Contains(clean, "/tests/") || strings.HasPrefix(clean, "tests/") ||
		strings.Contains(clean, "/test/") || strings.HasPrefix(clean, "test/") ||
		strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, ".test.ts") || strings.HasSuffix(base, ".spec.ts") ||
		strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py")
}

func formatDisqualificationReason(clean, base string) string {
	ext := strings.ToLower(filepath.Ext(base))
	if isTestFile(clean, base) {
		return clean + " touches test files or suites (requires manual review)"
	}
	if sourceCodeExtensions[ext] {
		return clean + " touches application source code (disqualified per strict exclusivity rule)"
	}
	if strings.HasPrefix(clean, ".github/workflows/") || clean == ".gitlab-ci.yml" {
		return clean + " touches CI workflows directly (requires manual review)"
	}
	if buildConfigBasenames[base] {
		return clean + " touches repository documentation or build config"
	}
	return clean + " is not a recognized dependency manifest or lockfile"
}

// FilterCandidates evaluates a batch list of candidate PRs against bot, label, title, and file rules.
func FilterCandidates(prs []CandidatePR) (FilterCandidatesOutput, error) {
	eligible := []EvaluatedPR{}
	disqualified := []EvaluatedPR{}

	for _, pr := range prs {
		if err := guard.CheckNoSecrets(pr.Title, pr.Author, pr.HeadRefName, pr.Body); err != nil {
			return FilterCandidatesOutput{}, err
		}
		for _, f := range pr.Files {
			if err := guard.CheckPathAllowed(f); err != nil {
				return FilterCandidatesOutput{}, err
			}
		}

		eval, err := evaluateSingleCandidate(pr)
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
		TotalEvaluated:  len(prs),
		EligibleCount:   len(eligible),
		ExcludedCount:   len(disqualified),
		EligiblePRs:     eligible,
		DisqualifiedPRs: disqualified,
	}, nil
}

func evaluateSingleCandidate(pr CandidatePR) (EvaluatedPR, error) {
	fileCheck, err := ClassifyFiles(pr.Files)
	if err != nil {
		return EvaluatedPR{}, err
	}

	reasons := []string{}
	isBot := isRecognizedBot(pr.Author)
	hasDepLabel := hasDependencyLabel(pr.Labels)
	hasDepTitle := matchesDependencyTitle(pr.Title)

	if !isBot && !hasDepLabel && !hasDepTitle {
		reasons = append(reasons, "author/title/labels do not match automated dependency update signatures")
	}
	if !fileCheck.Eligible {
		reasons = append(reasons, fileCheck.Reasons...)
	}

	return EvaluatedPR{
		Candidate: pr,
		Eligible:  len(reasons) == 0,
		FileCheck: fileCheck,
		Reasons:   reasons,
	}, nil
}

func isRecognizedBot(author string) bool {
	a := strings.ToLower(strings.TrimSpace(author))
	if _, ok := knownBotAuthors[a]; ok {
		return true
	}
	return strings.HasSuffix(a, "[bot]") || strings.HasSuffix(a, "-bot")
}

func hasDependencyLabel(labels []string) bool {
	return slices.ContainsFunc(labels, func(l string) bool {
		_, ok := knownDepLabels[strings.ToLower(strings.TrimSpace(l))]
		return ok
	})
}

func matchesDependencyTitle(title string) bool {
	t := strings.ToLower(strings.TrimSpace(title))
	return strings.HasPrefix(t, "bump ") ||
		strings.HasPrefix(t, "chore(deps):") ||
		strings.HasPrefix(t, "fix(deps):") ||
		strings.HasPrefix(t, "security update") ||
		strings.HasPrefix(t, "update dependency")
}
