package skilllint_test

import (
	"bytes"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/skilllint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleBatchReport() skilllint.BatchReport {
	return skilllint.BatchReport{
		OK:            true,
		Passed:        true,
		Strict:        false,
		RootDir:       "skills",
		TotalSkills:   1,
		PassedSkills:  1,
		FailedSkills:  0,
		TotalErrors:   0,
		TotalWarnings: 0,
		Reports: []skilllint.SkillReport{
			{
				OK:         true,
				Passed:     true,
				SkillName:  "test-skill",
				TargetDir:  "skills/test-skill",
				SkillLines: 120,
				Errors:     0,
				Warnings:   0,
				Findings:   nil,
			},
		},
	}
}

func TestRenderOutputFormats(t *testing.T) {
	batch := sampleBatchReport()

	t.Run("render markdown report", func(t *testing.T) {
		var buf bytes.Buffer
		err := skilllint.RenderMarkdown(&buf, batch, "")
		require.NoError(t, err)
		out := buf.String()
		assert.Contains(t, out, "# Skill Lint Report")
		assert.Contains(t, out, "test-skill")
		assert.Contains(t, out, "PASS")
		assert.Contains(t, out, "1 total (1 passed, 0 failed)")
	})

	t.Run("render json report", func(t *testing.T) {
		var buf bytes.Buffer
		err := skilllint.RenderJSON(&buf, batch)
		require.NoError(t, err)
		out := buf.String()
		assert.Contains(t, out, `"ok": true`)
		assert.Contains(t, out, `"passed": true`)
		assert.Contains(t, out, `"skill_name": "test-skill"`)
	})

	t.Run("render output invalid format", func(t *testing.T) {
		var buf bytes.Buffer
		err := skilllint.RenderOutput(&buf, batch, "invalid-format", "")
		require.ErrorContains(t, err, "invalid format")
	})
}

func TestRenderOutputFindingsAndEdgeCases(t *testing.T) {
	t.Run("render markdown with findings", func(t *testing.T) {
		failingBatch := skilllint.BatchReport{
			OK:           true,
			Passed:       false,
			Strict:       true,
			TotalSkills:  1,
			FailedSkills: 1,
			TotalErrors:  1,
			Reports: []skilllint.SkillReport{
				{
					OK:        true,
					Passed:    false,
					SkillName: "bad-skill",
					TargetDir: "/path/to/repo/skills/bad-skill",
					Errors:    1,
					Findings: []skilllint.Finding{
						{
							Check:    skilllint.CheckB,
							Severity: skilllint.SeverityError,
							File:     "SKILL.md",
							Line:     2,
							Message:  "name mismatch",
						},
						{
							Check:    skilllint.CheckA,
							Severity: skilllint.SeverityWarning,
							File:     "references/orphan.md",
							Message:  "orphan file",
						},
					},
				},
			},
		}

		var buf bytes.Buffer
		err := skilllint.RenderOutput(&buf, failingBatch, "markdown", "/path/to/repo")
		require.NoError(t, err)
		out := buf.String()
		assert.Contains(t, out, "**FAIL**")
		assert.Contains(t, out, "skills/bad-skill")
		assert.Contains(t, out, "## Findings")
		assert.Contains(t, out, "[Check B]")
		assert.Contains(t, out, "`SKILL.md:2`")
		assert.Contains(t, out, "`references/orphan.md`")
		assert.Contains(t, out, "name mismatch")
	})

	t.Run("render markdown empty reports", func(t *testing.T) {
		emptyBatch := skilllint.BatchReport{
			OK:     true,
			Passed: true,
		}
		var buf bytes.Buffer
		err := skilllint.RenderMarkdown(&buf, emptyBatch, "")
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "No skills found.")
	})
}
