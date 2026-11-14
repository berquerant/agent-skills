package prtriage

import (
	"path/filepath"
	"slices"
	"strings"
)

type manifestRule struct {
	ecosystem Ecosystem
	match     func(relPath, base string) bool
}

var manifestRules = []manifestRule{
	{
		ecosystem: EcosystemNode,
		match: func(_, base string) bool {
			return base == "package.json" || base == "package-lock.json" ||
				base == "pnpm-lock.yaml" || base == "pnpm-workspace.yaml" ||
				base == "yarn.lock" || base == "bun.lockb" || base == "bun.lock" ||
				base == ".npmrc" || base == ".yarnrc.yml" || base == ".yarnrc"
		},
	},
	{
		ecosystem: EcosystemGo,
		match: func(_, base string) bool {
			return base == "go.mod" || base == "go.sum" || base == "go.work" || base == "go.work.sum"
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
				base == "pdm.lock" || base == "uv.lock" ||
				base == "setup.cfg" || base == "setup.py" ||
				base == "requirements.txt" ||
				(strings.HasPrefix(base, "requirements") && strings.HasSuffix(base, ".txt"))
		},
	},
	{
		ecosystem: EcosystemRuby,
		match: func(_, base string) bool {
			return base == "Gemfile" || base == "Gemfile.lock" || strings.HasSuffix(base, ".gemspec")
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
				base == "build.gradle.kts" || base == "gradle.properties" ||
				base == "settings.gradle" || base == "settings.gradle.kts"
		},
	},
	{
		ecosystem: EcosystemDotNet,
		match: func(_, base string) bool {
			return strings.HasSuffix(base, ".csproj") || strings.HasSuffix(base, ".fsproj") ||
				strings.HasSuffix(base, ".vbproj") || base == "packages.lock.json" ||
				base == "Directory.Packages.props" || base == "Directory.Build.props" ||
				base == "nuget.config"
		},
	},
	{
		ecosystem: EcosystemAction,
		match: func(relPath, base string) bool {
			return base == "actions.lock" || strings.HasSuffix(relPath, ".github/workflows/actions.lock")
		},
	},
	{
		ecosystem: EcosystemElixir,
		match: func(_, base string) bool {
			return base == "mix.exs" || base == "mix.lock"
		},
	},
	{
		ecosystem: EcosystemSwift,
		match: func(_, base string) bool {
			return base == "Package.swift" || base == "Package.resolved"
		},
	},
	{
		ecosystem: EcosystemDart,
		match: func(_, base string) bool {
			return base == "pubspec.yaml" || base == "pubspec.lock"
		},
	},
}

var sourceCodeExtensions = map[string]bool{
	".go": true, ".ts": true, ".tsx": true, ".cts": true, ".mts": true,
	".js": true, ".jsx": true, ".mjs": true, ".cjs": true,
	".py": true, ".pyi": true, ".pyw": true,
	".rs": true, ".java": true, ".kt": true, ".kts": true,
	".scala": true, ".sc": true,
	".c": true, ".cc": true, ".cpp": true, ".cxx": true,
	".h": true, ".hh": true, ".hpp": true, ".hxx": true,
	".cs": true, ".csx": true,
	".rb": true, ".rake": true,
	".php": true, ".phtml": true,
	".swift": true, ".dart": true,
	".sh": true, ".bash": true, ".zsh": true, ".ksh": true,
	".lua": true, ".r": true, ".hs": true, ".lhs": true,
	".ex": true, ".exs": true, ".erl": true, ".hrl": true,
	".zig": true, ".ml": true, ".mli": true,
	".clj": true, ".cljs": true, ".jl": true,
	".pl": true, ".pm": true, ".sol": true,
}

var buildConfigBasenames = map[string]bool{
	"Dockerfile": true, "Containerfile": true,
	"compose.yaml": true, "compose.yml": true,
	"docker-compose.yml": true, "docker-compose.yaml": true,
	".dockerignore": true,
	"Makefile": true, "makefile": true, "GNUmakefile": true,
	"Taskfile.yml": true, "Taskfile.yaml": true,
	"justfile": true, "Justfile": true,
	"mise.toml": true, ".tool-versions": true, "Earthfile": true,
	"CMakeLists.txt": true, "meson.build": true,
	"BUILD.bazel": true, "BUILD": true, "WORKSPACE": true, "WORKSPACE.bazel": true,
	"Procfile": true, "Vagrantfile": true, "Brewfile": true,
	"flake.nix": true, "default.nix": true,
	".editorconfig": true, ".gitattributes": true, ".gitignore": true,
	"README.md": true, "LICENSE": true, "LICENSE.md": true,
	"CHANGELOG.md": true, "CONTRIBUTING.md": true,
	"CODE_OF_CONDUCT.md": true, "SECURITY.md": true, "NOTICE": true,
}

var knownBotAuthors = map[string]struct{}{
	"app/dependabot":        {},
	"dependabot[bot]":       {},
	"dependabot":            {},
	"renovate[bot]":         {},
	"renovate":              {},
	"snyk-bot":              {},
	"security-bot":          {},
	"gitlab-dependency-bot": {},
	"depfu[bot]":            {},
	"depfu":                 {},
	"pyup-bot":              {},
	"greenkeeper[bot]":      {},
	"scala-steward":         {},
	"allstar[bot]":          {},
	"imgbot[bot]":           {},
	"github-actions[bot]":   {},
}

var knownDepLabels = map[string]struct{}{
	"dependencies":        {},
	"dependencies-update": {},
	"dependency":          {},
	"deps":                {},
	"deps-dev":            {},
	"security":            {},
	"cve":                 {},
	"dependency-update":   {},
	"automated pr":        {},
	"bot":                 {},
	"update":              {},
	"renovate":            {},
	"dependabot":          {},
	"snyk":                {},
	"vulnerability":       {},
}

var defaultDependencyTitlePrefixes = []string{
	"bump ",
	"chore(deps):",
	"chore(deps-dev):",
	"build(deps):",
	"build(deps-dev):",
	"fix(deps):",
	"fix(deps-dev):",
	"security update",
	"update dependency",
	"update ",
	"upgrade ",
	"pin dependency",
	"lock file maintenance",
	"[snyk]",
	"[security]",
	"deps:",
	"dep:",
}

var testDirMarkers = []string{
	"/tests/", "tests/", "/test/", "test/",
	"/__tests__/", "__tests__/",
	"/spec/", "spec/", "/specs/", "specs/",
}

func isTestFile(clean, base string) bool {
	for _, marker := range testDirMarkers {
		if strings.Contains(clean, marker) || strings.HasPrefix(clean, marker) {
			return true
		}
	}
	return strings.HasSuffix(base, "_test.go") ||
		strings.Contains(base, ".test.") ||
		strings.Contains(base, ".spec.") ||
		strings.HasPrefix(base, "test_") ||
		strings.HasSuffix(base, "_test.py") ||
		strings.HasSuffix(base, "_spec.rb") ||
		strings.HasSuffix(base, "_test.rb") ||
		strings.HasSuffix(base, "_test.exs") ||
		strings.HasSuffix(base, "_test.rs") ||
		strings.HasPrefix(base, "Test")
}

func isCIWorkflowFile(clean string) bool {
	return strings.HasPrefix(clean, ".github/workflows/") ||
		clean == ".gitlab-ci.yml" ||
		strings.HasPrefix(clean, ".circleci/") ||
		clean == "azure-pipelines.yml" ||
		clean == ".travis.yml"
}

func isRecognizedBot(author string, extraBots []string) bool {
	a := strings.ToLower(strings.TrimSpace(author))
	if _, ok := knownBotAuthors[a]; ok {
		return true
	}
	if strings.HasSuffix(a, "[bot]") || strings.HasSuffix(a, "-bot") {
		return true
	}
	return slices.ContainsFunc(extraBots, func(b string) bool {
		return strings.EqualFold(strings.TrimSpace(b), a)
	})
}

func hasDependencyLabel(labels []string, extraLabels []string) bool {
	return slices.ContainsFunc(labels, func(l string) bool {
		norm := strings.ToLower(strings.TrimSpace(l))
		if _, ok := knownDepLabels[norm]; ok {
			return true
		}
		return slices.ContainsFunc(extraLabels, func(el string) bool {
			return strings.EqualFold(strings.TrimSpace(el), norm)
		})
	})
}

func matchesDependencyTitle(title string, extraPatterns []string) bool {
	t := strings.ToLower(strings.TrimSpace(title))
	for _, prefix := range defaultDependencyTitlePrefixes {
		if strings.HasPrefix(t, prefix) {
			return true
		}
	}
	for _, pat := range extraPatterns {
		p := strings.ToLower(strings.TrimSpace(pat))
		if p != "" && (strings.HasPrefix(t, p) || strings.Contains(t, p)) {
			return true
		}
	}
	return false
}

func matchesCustomManifestPattern(clean, base string, patterns []string) bool {
	for _, pat := range patterns {
		p := strings.TrimSpace(pat)
		if p == "" {
			continue
		}
		if ok, _ := filepath.Match(p, base); ok {
			return true
		}
		if ok, _ := filepath.Match(p, clean); ok {
			return true
		}
	}
	return false
}
