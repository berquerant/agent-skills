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
	}
}

func TestCheckInline(t *testing.T) {
	for _, tt := range checkInlineTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			out, err := snippetmgr.CheckInline(snippetmgr.CheckInlineInput{Command: tt.cmd, Language: tt.lang})
			require.NoError(t, err)
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

	t.Run("rejects unsupported language", func(t *testing.T) {
		_, err := snippetmgr.CheckInline(snippetmgr.CheckInlineInput{Command: "puts 'hi'", Language: "ruby"})
		require.ErrorContains(t, err, "unsupported language")
	})
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "valid snake_case name", input: "rebase_and_test_worktree", want: "rebase_and_test_worktree"},
		{name: "valid kebab-case name", input: "check-active-prs", want: "check-active-prs"},
		{name: "empty name", input: "", wantErr: "required"},
		{name: "manual version suffix rejected", input: "my_script_v2", wantErr: "_vN"},
		{name: "uppercase rejected", input: "BadCase", wantErr: "lowercase"},
		{name: "path separator rejected", input: "a/b", wantErr: "path separators"},
		{name: "file extension rejected", input: "foo.sh", wantErr: "file extensions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := snippetmgr.ValidateName(tt.input)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateContent(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		lang    string
		wantErr string
	}{
		{name: "valid bash snippet", code: sampleBashV1, lang: "bash"},
		{name: "valid python snippet", code: samplePythonSnippet, lang: "python"},
		{name: "missing shebang", code: "echo hi", lang: "bash", wantErr: "shebang"},
		{name: "CRLF line endings rejected", code: "#!/usr/bin/env bash\r\n# Description: test\r\n# Usage: ./x.sh\r\nset -euo pipefail\r\n", lang: "bash", wantErr: "LF"},
		{name: "missing documentation header", code: "#!/usr/bin/env bash\nset -euo pipefail\necho hi", lang: "bash", wantErr: "documentation header"},
		{name: "missing set -euo pipefail", code: "#!/usr/bin/env bash\n# Description: test\n# Usage: ./x.sh\necho hi", lang: "bash", wantErr: "set -euo pipefail"},
		{name: "commented set -euo pipefail does not count", code: "#!/usr/bin/env bash\n# Description: test\n# Usage: ./x.sh\n# set -euo pipefail\necho hi", lang: "bash", wantErr: "set -euo pipefail"},
		{name: "embedded secret rejected", code: sampleBashV1 + "\nTOKEN=ghp_123456789012345678901234567890123456", lang: "bash", wantErr: "secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := snippetmgr.ValidateContent(tt.code, tt.lang)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
