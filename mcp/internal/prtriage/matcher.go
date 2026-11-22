package prtriage

import (
	"regexp"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

type archetypeDef struct {
	archetype FailureArchetype
	title     string
	patterns  []*regexp.Regexp
	rootCause string
	playbook  []string
}

var archetypeDefinitions = []archetypeDef{
	{
		archetype: ArchetypeOutdatedBase,
		title:     "Outdated Base Branch vs. Evolving CI Environment",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(unsupported (go|node|python|rust) version|unsupported runner|base branch out of date|runner image deprecated|runner environment)`),
			regexp.MustCompile(`(?i)(missing toolchain version|environment setup failed|incompatible glibc)`),
		},
		rootCause: "PR branch was created before recent upstream environment fixes, runner version pins, or CI updates were merged into the default branch.",
		playbook: []string{
			"Fetch latest default branch: git fetch origin <default-branch>",
			"Rebase PR branch onto latest default branch: git rebase origin/<default-branch>",
			"Run local tests to confirm clean environment compatibility",
			"Push rebased branch with lease: git push --force-with-lease origin <head-branch>",
		},
	},
	{
		archetype: ArchetypeDerivedArtifacts,
		title:     "Derived & Secondary Artifacts Out of Sync",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(license notice|third_party_notice|third-party licenses|derived artifacts|git diff --exit-code)`),
			regexp.MustCompile(`(?i)(generated code out of (date|sync)|protobuf compiler drift|mockgen out of date|readme-diff)`),
		},
		rootCause: "The update tool bumped package manifests, but secondary artifacts (license notices, generated protobufs/mocks, compiled bundles) were not re-generated.",
		playbook: []string{
			"Inspect task runners (Makefile, Taskfile, mise, package.json) for code generation or license reporting targets",
			"Execute the generator command inside an isolated worktree",
			"Review git diff to ensure changes are strictly confined to the updated dependency",
			"Commit derived changes alongside manifest updates and push",
		},
	},
	{
		archetype: ArchetypeSnapshotDrift,
		title:     "Snapshot & Regression Test Manifest Drift",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(snapshot (comparison )?failed|toMatchSnapshot|golden file mismatch|regression test baseline)`),
			regexp.MustCompile(`(?i)(snapshot manifest drift|test baseline drift)`),
		},
		rootCause: "The updated dependency legitimately shifts rendered configurations, templates, or test snapshots that cannot be updated autonomously by bots.",
		playbook: []string{
			"Execute repository test snapshot update command (e.g. npm test -- -u, go test -update)",
			"Scrutinize resulting diff to ensure shifts strictly correspond to expected version bumps",
			"Commit updated snapshots with a clear test-baseline commit message",
			"Enforce serialization: process shared snapshot baselines strictly one PR at a time",
		},
	},
	{
		archetype: ArchetypeFastForward,
		title:     "Base Branch Protection & Ruleset Policy Violation",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(non_fast_forward|base branch policy prohibits|linear history rule|branch is out of date|must be up to date before merging)`),
			regexp.MustCompile(`(?i)(pull request is not fast[- ]forward)`),
		},
		rootCause: "The default branch enforces strict fast-forward or linear history rules, and earlier merged PRs caused this PR branch to fall behind.",
		playbook: []string{
			"Do not bypass branch protections using elevated or admin flags",
			"In an isolated worktree, fetch latest default branch and rebase: git rebase origin/<default-branch>",
			"Push rebased branch with lease: git push --force-with-lease origin <head-branch>",
			"Wait for CI checks to re-pass and merge cleanly",
		},
	},
	{
		archetype: ArchetypeTokenScope,
		title:     "Platform API Token Scope Restrictions",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(refusing to allow an OAuth App to create or update workflow|without 'workflow' scope|GraphQL: refusing)`),
			regexp.MustCompile(`(?i)(insufficient token scopes? for workflow)`),
		},
		rootCause: "Platform REST/GraphQL APIs reject branch modification commands when changing workflow files without elevated 'workflow' OAuth scopes.",
		playbook: []string{
			"Avoid API-level branch mutation commands (e.g. gh pr update-branch)",
			"Perform rebase and branch operations through standard Git transport (SSH/HTTPS) inside a worktree",
			"Push branch directly via git push --force-with-lease origin <head-branch>",
		},
	},
}

// MatchFailure compares a CI failure log excerpt against the 5 failure archetypes.
func MatchFailure(logExcerpt string) (MatchFailureOutput, error) {
	if err := guard.CheckNoSecrets(logExcerpt); err != nil {
		return MatchFailureOutput{}, err
	}

	trimmed := strings.TrimSpace(logExcerpt)
	if trimmed == "" {
		return MatchFailureOutput{
			OK:      true,
			Matched: false,
			Advice:  "Log excerpt is empty. Provide CI build or test error logs for archetype classification.",
		}, nil
	}
	if len(trimmed) > 64*1024 {
		trimmed = trimmed[:64*1024]
	}

	var matches []ArchetypeMatch
	for _, def := range archetypeDefinitions {
		matchedTerms := findMatchingTerms(trimmed, def.patterns)
		if len(matchedTerms) > 0 {
			conf := "moderate"
			if len(matchedTerms) >= 2 {
				conf = "high"
			}
			matches = append(matches, ArchetypeMatch{
				Archetype:  def.archetype,
				Title:      def.title,
				Confidence: conf,
				MatchedOn:  matchedTerms,
				RootCause:  def.rootCause,
				Playbook:   def.playbook,
			})
		}
	}

	advice := "No standard dependency failure archetype matched. Proceed with standard code review and log inspection."
	if len(matches) > 0 {
		advice = "Matched failure archetype(s). Review proposed root cause and follow remediation playbook."
	}

	return MatchFailureOutput{
		OK:      true,
		Matched: len(matches) > 0,
		Matches: matches,
		Advice:  advice,
	}, nil
}

func findMatchingTerms(log string, patterns []*regexp.Regexp) []string {
	var terms []string
	for _, pat := range patterns {
		if loc := pat.FindString(log); loc != "" {
			terms = append(terms, strings.TrimSpace(loc))
		}
	}
	return terms
}
