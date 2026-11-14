package prtriage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/prtriage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyFiles(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		wantEligible  bool
		wantEcosystem prtriage.Ecosystem
		wantReasonSub string
	}{
		{
			name:          "pure Go update",
			files:         []string{"go.mod", "go.sum"},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemGo,
		},
		{
			name:          "pure npm update",
			files:         []string{"package.json", "package-lock.json"},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemNode,
		},
		{
			name:          "pure Python update",
			files:         []string{"pyproject.toml", "poetry.lock"},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemPython,
		},
		{
			name:          "actions lock update",
			files:         []string{".github/workflows/actions.lock"},
			wantEligible:  true,
			wantEcosystem: prtriage.EcosystemAction,
		},
		{
			name:          "disqualified by application source code",
			files:         []string{"go.mod", "cmd/app/main.go"},
			wantEligible:  false,
			wantReasonSub: "touches application source code",
		},
		{
			name:          "disqualified by test file",
			files:         []string{"package.json", "src/util.test.ts"},
			wantEligible:  false,
			wantReasonSub: "touches test files",
		},
		{
			name:          "disqualified by documentation",
			files:         []string{"Cargo.toml", "README.md"},
			wantEligible:  false,
			wantReasonSub: "touches repository documentation",
		},
		{
			name:          "empty files list",
			files:         []string{},
			wantEligible:  false,
			wantReasonSub: "no files changed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := prtriage.ClassifyFiles(tt.files)
			require.NoError(t, err)
			assert.Equal(t, tt.wantEligible, out.Eligible)
			if tt.wantEcosystem != "" {
				assert.Equal(t, tt.wantEcosystem, out.Ecosystem)
			}
			if tt.wantReasonSub != "" {
				require.NotEmpty(t, out.Reasons)
				assert.Contains(t, out.Reasons[0], tt.wantReasonSub)
			}
		})
	}
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
	}

	out, err := prtriage.FilterCandidates(prs)
	require.NoError(t, err)
	assert.Equal(t, 3, out.TotalEvaluated)
	assert.Equal(t, 2, out.EligibleCount)
	assert.Equal(t, 1, out.ExcludedCount)
	require.Len(t, out.EligiblePRs, 2)
	assert.Equal(t, 101, out.EligiblePRs[0].Candidate.Number)
	assert.Equal(t, 102, out.EligiblePRs[1].Candidate.Number)
	require.Len(t, out.DisqualifiedPRs, 1)
	assert.Equal(t, 103, out.DisqualifiedPRs[0].Candidate.Number)
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
