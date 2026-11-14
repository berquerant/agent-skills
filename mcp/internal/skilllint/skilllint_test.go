package skilllint_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/skilllint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validSkillMarkdown(name string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + name + "\n")
	b.WriteString("description: >-\n")
	b.WriteString("  Use this skill when testing the automated skill linter across all checks A through K\n")
	b.WriteString("  to ensure repository conventions, links, and frontmatter rules are enforced.\n")
	b.WriteString("---\n\n")
	b.WriteString("# Test Skill\n\n")
	b.WriteString("See [ref.md](references/ref.md) and [run.sh](scripts/run.sh).\n\n")
	b.WriteString("## Step 1: Execute Check\n\n")
	for range 75 {
		b.WriteString("- Perform deterministic validation step.\n")
	}
	b.WriteString("\nVerify: Output matches expectations.\n\n")
	b.WriteString("## Guidelines\n\n")
	b.WriteString("- Keep rules deterministic.\n")
	return b.String()
}

func setupTestRepo(t *testing.T, skillName, skillMD string) (string, string) {
	t.Helper()
	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "README.md"), []byte(skillName+"\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "AGENTS.md"), []byte(skillName+"\n"), 0o600))

	routerRefDir := filepath.Join(repoRoot, "skills", "skill-router", "references")
	require.NoError(t, os.MkdirAll(routerRefDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(routerRefDir, "intent-matrix.md"), []byte(skillName+"\n"), 0o600))

	skillDir := filepath.Join(repoRoot, "skills", skillName)
	require.NoError(t, os.MkdirAll(filepath.Join(skillDir, "references"), 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(skillDir, "scripts"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("## Shared Heading\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "scripts", "run.sh"), []byte("#!/bin/sh\n"), 0o700))
	return repoRoot, skillDir
}

func TestLintSkillChecks(t *testing.T) {
	tests := []struct {
		name         string
		skillName    string
		mutate       func(t *testing.T, repoRoot, skillDir string)
		wantPassed   bool
		wantCheckIDs []skilllint.CheckID
	}{
		{
			name:       "valid skill passes strict mode with zero findings",
			skillName:  "valid-skill",
			wantPassed: true,
		},
		{
			name:      "check A and B: short SKILL.md, orphan file, and mismatched frontmatter name",
			skillName: "bad-meta",
			mutate: func(t *testing.T, _, skillDir string) {
				md := "---\nname: wrong_NAME\ndescription: >-\n  Use this skill when testing short skill files with invalid frontmatter name.\n---\n# Title\n## Step 1: Go\nVerify: ok\n## Guidelines\n"
				require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(md), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(skillDir, "references", "orphan.md"), []byte("hi\n"), 0o600))
			},
			wantPassed:   false,
			wantCheckIDs: []skilllint.CheckID{skilllint.CheckA, skilllint.CheckB, skilllint.CheckC},
		},
		{
			name:      "check D and E: missing Verify, missing Guidelines, broken link, machine path",
			skillName: "bad-structure",
			mutate: func(t *testing.T, _, skillDir string) {
				md := validSkillMarkdown("bad-structure")
				md = strings.ReplaceAll(md, "Verify: Output matches expectations.", "No verify here.")
				md = strings.ReplaceAll(md, "## Guidelines", "## Other Ending")
				md += "\nBroken [missing](references/missing.md) and machine path /Users/someone/secret.\n"
				require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(md), 0o600))
			},
			wantPassed:   false,
			wantCheckIDs: []skilllint.CheckID{skilllint.CheckD, skilllint.CheckE},
		},
		{
			name:      "check F, H, K: non-executable script, unregistered repo doc, long shell command and CoT",
			skillName: "bad-hygiene",
			mutate: func(t *testing.T, repoRoot, skillDir string) {
				require.NoError(t, os.Chmod(filepath.Join(skillDir, "scripts", "run.sh"), 0o600))
				require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "README.md"), []byte("other\n"), 0o600))
				ref := "# Ref\n[f](ref.md) (`ref.md`)\nPlease think step by step.\n```sh\n" + strings.Repeat("echo a && ", 15) + "true\n```\n"
				require.NoError(t, os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte(ref), 0o600))
			},
			wantPassed:   false,
			wantCheckIDs: []skilllint.CheckID{skilllint.CheckF, skilllint.CheckH, skilllint.CheckK},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot, skillDir := setupTestRepo(t, tt.skillName, validSkillMarkdown(tt.skillName))
			if tt.mutate != nil {
				tt.mutate(t, repoRoot, skillDir)
			}
			rep, err := skilllint.LintSkill(skillDir, true)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPassed, rep.Passed)

			var gotChecks []skilllint.CheckID
			for _, f := range rep.Findings {
				gotChecks = append(gotChecks, f.Check)
			}
			for _, wantID := range tt.wantCheckIDs {
				assert.Contains(t, gotChecks, wantID)
			}
		})
	}
}

func TestRepositorySkillsParity(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err)

	batch, err := skilllint.LintAll(repoRoot, true)
	require.NoError(t, err)
	assert.True(t, batch.OK)
	assert.True(t, batch.Passed, "all repository skills should pass strict lint")
	assert.GreaterOrEqual(t, batch.TotalSkills, 25)
	assert.Equal(t, 0, batch.TotalErrors)
	assert.Equal(t, 0, batch.TotalWarnings)
}
