package snippetmgr_test

import (
	"strings"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/snippetmgr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type checkInlineTestCase struct {
	name           string
	cmd            string
	lang           string
	wantExtract    bool
	wantReasonSub  string
	wantLang       string
	wantHintSubstr string
}

func checkInlineTestCases() []checkInlineTestCase {
	return []checkInlineTestCase{
		{
			name:        "short simple command does not require extraction",
			cmd:         "git status -sb",
			wantExtract: false,
			wantLang:    "bash",
		},
		{
			name:           "command >= 120 chars triggers extraction",
			cmd:            "echo " + strings.Repeat("a", 120),
			wantExtract:    true,
			wantReasonSub:  "character length",
			wantLang:       "bash",
			wantHintSubstr: "read-only and local",
		},
		{
			name:          "multi-line command (>= 3 lines) triggers extraction",
			cmd:           "echo one\necho two\necho three",
			wantExtract:   true,
			wantReasonSub: "line count",
			wantLang:      "bash",
		},
		{
			name:          "4 chained commands triggers extraction",
			cmd:           "ls | grep foo | sort | uniq",
			wantExtract:   true,
			wantReasonSub: "chained commands",
			wantLang:      "bash",
		},
		{
			name:          "inline python3 -c with import triggers extraction via snippet.Analyze and suggests python",
			cmd:           `python3 -c "import json; print(json.dumps({'ok': True}))"`,
			wantExtract:   true,
			wantReasonSub: "inline_interpreter",
			wantLang:      "python",
		},
		{
			name:          "short complex &&/|| branching triggers extraction via snippet.Analyze",
			cmd:           "a && b && c || d",
			wantExtract:   true,
			wantReasonSub: "branch_count=3",
			wantLang:      "bash",
		},
		{
			name:          "short eval command triggers extraction via snippet.Analyze",
			cmd:           `eval "$CMD"`,
			wantExtract:   true,
			wantReasonSub: "complexity Level 4",
			wantLang:      "bash",
		},
		{
			name:          "python code delegates complexity check to Python AST analyzer",
			cmd:           `res = eval("1+2")`,
			lang:          "python",
			wantExtract:   true,
			wantReasonSub: "complexity Level 4",
			wantLang:      "python",
		},
		{
			name:          "broad CLI tool gh triggers extraction",
			cmd:           "gh pr list --state open",
			wantExtract:   true,
			wantReasonSub: `broad CLI tool "gh"`,
			wantLang:      "bash",
		},
		{
			name:          "git -C infix argument triggers extraction",
			cmd:           "git -C /path/to/repo status",
			wantExtract:   true,
			wantReasonSub: "infix variable argument",
			wantLang:      "bash",
		},
		{
			name:          "grep with trailing target file triggers extraction",
			cmd:           `grep -i "ERROR" /path/to/app.log`,
			wantExtract:   true,
			wantReasonSub: "trailing target file argument",
			wantLang:      "bash",
		},
	}
}

func TestCheckInline(t *testing.T) {
	for _, tt := range checkInlineTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			out := snippetmgr.CheckInline(snippetmgr.CheckInlineInput{Command: tt.cmd, Language: tt.lang})
			assert.Equal(t, tt.wantExtract, out.ShouldExtract)
			assert.Equal(t, tt.wantLang, out.SuggestedLanguage)
			assert.True(t, out.Audit.OK)
			if tt.wantReasonSub != "" {
				assert.Contains(t, strings.Join(out.Reasons, " | "), tt.wantReasonSub)
			}
			if tt.wantHintSubstr != "" {
				assert.Contains(t, out.LightweightHint, tt.wantHintSubstr)
			}
		})
	}
}

func TestNamingAndContentValidation(t *testing.T) {
	t.Run("ValidateName rules", func(t *testing.T) {
		valid, err := snippetmgr.ValidateName("rebase_and_test_worktree")
		require.NoError(t, err)
		assert.Equal(t, "rebase_and_test_worktree", valid)

		for _, bad := range []string{"", "run", "helper", "script", "my_script_v2", "BadCase", "a/b", "foo.sh"} {
			_, err := snippetmgr.ValidateName(bad)
			assert.Error(t, err, "expected error for %q", bad)
		}
	})

	t.Run("ValidateContent rules", func(t *testing.T) {
		require.NoError(t, snippetmgr.ValidateContent(sampleBashV1, "bash"))
		require.NoError(t, snippetmgr.ValidateContent(samplePythonSnippet, "python"))

		assert.ErrorContains(t, snippetmgr.ValidateContent("echo hi", "bash"), "shebang")
		assert.ErrorContains(t, snippetmgr.ValidateContent("#!/usr/bin/env bash\nset -euo pipefail\necho hi", "bash"), "documentation header")
		assert.ErrorContains(t, snippetmgr.ValidateContent("#!/usr/bin/env bash\n# Description: test\n# Usage: ./x.sh\necho hi", "bash"), "set -euo pipefail")
		assert.ErrorContains(t, snippetmgr.ValidateContent(sampleBashV1+"\nTOKEN=ghp_123456789012345678901234567890123456", "bash"), "secret")
	})
}
