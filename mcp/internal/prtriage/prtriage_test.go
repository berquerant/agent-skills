package prtriage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/prtriage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyFiles_Ecosystems(t *testing.T) {
	tests := []struct {
		name          string
		input         prtriage.ClassifyFilesInput
		wantEligible  bool
		wantEcosystem prtriage.Ecosystem
	}{
		{
			name: "pure Go update",
			input: prtriage.ClassifyFilesInput{
				Files: []string{"go.mod", "go.sum"},
			},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemGo,
		},
		{
			name: "pure npm update",
			input: prtriage.ClassifyFilesInput{
				Files: []string{"package.json", "package-lock.json"},
			},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemNode,
		},
		{
			name: "pure Python update",
			input: prtriage.ClassifyFilesInput{
				Files: []string{"pyproject.toml", "poetry.lock"},
			},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemPython,
		},
		{
			name: "actions lock update",
			input: prtriage.ClassifyFilesInput{
				Files: []string{".github/workflows/actions.lock"},
			},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemAction,
		},
		{
			name: "Swift package update",
			input: prtriage.ClassifyFilesInput{
				Files: []string{"Package.swift", "Package.resolved"},
			},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemSwift,
		},
		{
			name: "Elixir mix update",
			input: prtriage.ClassifyFilesInput{
				Files: []string{"mix.exs", "mix.lock"},
			},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemElixir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := prtriage.ClassifyFiles(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantEligible, out.Eligible)
			assert.Equal(t, tt.wantEcosystem, out.Ecosystem)
		})
	}
}

func TestClassifyFiles_Disqualifications(t *testing.T) {
	tests := []struct {
		name          string
		input         prtriage.ClassifyFilesInput
		wantReasonSub string
	}{
		{
			name: "disqualified by application source code",
			input: prtriage.ClassifyFilesInput{
				Files: []string{"go.mod", "cmd/app/main.go"},
			},
			wantReasonSub: "touches application source code",
		},
		{
			name: "disqualified by test file",
			input: prtriage.ClassifyFilesInput{
				Files: []string{"package.json", "src/util.test.ts"},
			},
			wantReasonSub: "touches test files",
		},
		{
			name: "disqualified by documentation",
			input: prtriage.ClassifyFilesInput{
				Files: []string{"Cargo.toml", "README.md"},
			},
			wantReasonSub: "touches repository documentation",
		},
		{
			name: "disqualified by build configuration",
			input: prtriage.ClassifyFilesInput{
				Files: []string{"go.mod", "Dockerfile"},
			},
			wantReasonSub: "build config",
		},
		{
			name: "empty files list",
			input: prtriage.ClassifyFilesInput{
				Files: []string{},
			},
			wantReasonSub: "no files changed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := prtriage.ClassifyFiles(tt.input)
			require.NoError(t, err)
			assert.False(t, out.Eligible)
			require.NotEmpty(t, out.Reasons)
			assert.Contains(t, out.Reasons[0], tt.wantReasonSub)
		})
	}
}

func TestClassifyFiles_AgentOverrides(t *testing.T) {
	// Unknown without override
	outUnknown, err := prtriage.ClassifyFiles(prtriage.ClassifyFilesInput{
		Files: []string{"go.mod", "custom_dependency.xyz"},
	})
	require.NoError(t, err)
	assert.False(t, outUnknown.Eligible)
	assert.Len(t, outUnknown.UnknownFiles, 1)
	assert.Contains(t, outUnknown.Reasons[0], "is unclassified (category unknown)")

	// Resolved via agent manifest override
	outManifest, err := prtriage.ClassifyFiles(prtriage.ClassifyFilesInput{
		Files: []string{"go.mod", "custom_dependency.xyz"},
		Overrides: map[string]string{
			"custom_dependency.xyz": "manifest",
		},
	})
	require.NoError(t, err)
	assert.True(t, outManifest.Eligible)
	assert.Equal(t, prtriage.EcosystemMixed, outManifest.Ecosystem)

	// Resolved via agent ignored override
	outIgnored, err := prtriage.ClassifyFiles(prtriage.ClassifyFilesInput{
		Files: []string{"go.mod", "custom_dependency.xyz"},
		Overrides: map[string]string{
			"custom_dependency.xyz": "ignored",
		},
	})
	require.NoError(t, err)
	assert.True(t, outIgnored.Eligible)
	assert.Equal(t, prtriage.EcosystemGo, outIgnored.Ecosystem)

	// Resolved via extra manifest pattern
	outPattern, err := prtriage.ClassifyFiles(prtriage.ClassifyFilesInput{
		Files:                 []string{"go.mod", "dependencies.mycustomlock"},
		ExtraManifestPatterns: []string{"*.mycustomlock"},
	})
	require.NoError(t, err)
	assert.True(t, outPattern.Eligible)
	assert.Equal(t, prtriage.EcosystemMixed, outPattern.Ecosystem)
}

func TestFilterCandidates(t *testing.T) {
	prs := []prtriage.CandidatePR{
		{
			Number:      101,
			Title:       "Bump github.com/stretchr/testify from 1.9.0 to 1.10.0",
			Author:      "app/dependabot",
			HeadRefName: "dependabot/go_modules/testify-1.10.0",
			Labels:      []string{"dependencies"},
			Files:       []string{"go.mod", "go.sum"},
		},
		{
			Number:      102,
			Title:       "Update dependency typescript to v5.8.0",
			Author:      "renovate[bot]",
			HeadRefName: "renovate/typescript-5.x",
			Labels:      []string{"dependencies"},
			Files:       []string{"package.json", "yarn.lock"},
		},
		{
			Number:      103,
			Title:       "feat: implement new auth flow",
			Author:      "alice",
			HeadRefName: "feature/auth",
			Labels:      []string{"enhancement"},
			Files:       []string{"go.mod", "auth.go"},
		},
		{
			Number:      104,
			Title:       "manual bump for internal dependency",
			Author:      "security-officer",
			HeadRefName: "patch/security",
			Labels:      []string{"internal"},
			Files:       []string{"go.mod", "go.sum"},
		},
	}

	// Default evaluation: PR 104 is disqualified by non-automated author/title signature
	out, err := prtriage.FilterCandidates(prtriage.FilterCandidatesInput{
		PRs: prs,
	})
	require.NoError(t, err)
	assert.Equal(t, 4, out.TotalEvaluated)
	assert.Equal(t, 2, out.EligibleCount)
	assert.Equal(t, 2, out.ExcludedCount)
	require.Len(t, out.EligiblePRs, 2)
	assert.Equal(t, 101, out.EligiblePRs[0].Candidate.Number)
	assert.Equal(t, 102, out.EligiblePRs[1].Candidate.Number)

	// Agent-specified AllowedPRNumbers: approves PR 104
	outWithAllowed, err := prtriage.FilterCandidates(prtriage.FilterCandidatesInput{
		PRs:              prs,
		AllowedPRNumbers: []int{104},
	})
	require.NoError(t, err)
	assert.Equal(t, 3, outWithAllowed.EligibleCount)
	assert.Equal(t, 1, outWithAllowed.ExcludedCount)
}

func TestDetectTasks(t *testing.T) {
	tmp := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module testpkg\n\ngo 1.27\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "Makefile"), []byte("test:\n\tgo test ./...\nlint:\n\tgolangci-lint run\n"), 0o600))
	pkgJSON := `{"scripts": {"test": "jest", "lint": "eslint .", "build": "tsc"}}`
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "package.json"), []byte(pkgJSON), 0o600))

	out, err := prtriage.DetectTasks(tmp)
	require.NoError(t, err)
	assert.True(t, out.HasTests)
	assert.True(t, out.HasLinters)
	assert.Contains(t, out.Runners, "go")
	assert.Contains(t, out.Runners, "make")
	assert.Contains(t, out.Runners, "npm/package.json")

	commands := []string{}
	for _, tsk := range out.Tasks {
		commands = append(commands, tsk.Command)
	}
	assert.Contains(t, commands, "make test")
	assert.Contains(t, commands, "make lint")
	assert.Contains(t, commands, "npm run test")
}

func TestMatchFailure(t *testing.T) {
	tests := []struct {
		name          string
		log           string
		wantMatched   bool
		wantArchetype prtriage.FailureArchetype
	}{
		{
			name:          "outdated runner and base branch",
			log:           "Error: unsupported runner version encountered. base branch out of date.",
			wantMatched:   true,
			wantArchetype: prtriage.ArchetypeOutdatedBase,
		},
		{
			name:          "derived artifacts drift",
			log:           "FAILURE: git diff --exit-code failed! license notice or generated code out of sync.",
			wantMatched:   true,
			wantArchetype: prtriage.ArchetypeDerivedArtifacts,
		},
		{
			name:          "snapshot comparison failed",
			log:           "FAIL: toMatchSnapshot failed. regression test baseline mismatch.",
			wantMatched:   true,
			wantArchetype: prtriage.ArchetypeSnapshotDrift,
		},
		{
			name:          "fast-forward policy violation",
			log:           "Error: non_fast_forward rule violation: base branch policy prohibits merge.",
			wantMatched:   true,
			wantArchetype: prtriage.ArchetypeFastForward,
		},
		{
			name:          "token scope restriction",
			log:           "GraphQL: refusing to allow an OAuth App to create or update workflow without 'workflow' scope",
			wantMatched:   true,
			wantArchetype: prtriage.ArchetypeTokenScope,
		},
		{
			name:        "empty log",
			log:         "",
			wantMatched: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := prtriage.MatchFailure(tt.log)
			require.NoError(t, err)
			assert.Equal(t, tt.wantMatched, out.Matched)
			if tt.wantArchetype != "" {
				require.NotEmpty(t, out.Matches)
				assert.Equal(t, tt.wantArchetype, out.Matches[0].Archetype)
				assert.NotEmpty(t, out.Matches[0].Playbook)
			}
		})
	}
}
