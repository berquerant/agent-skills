package snippet_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/snippet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeRubricLevels(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		lang           string
		wantLevel      int
		wantRating     string
		wantStatus     string
		wantConfidence snippet.Confidence
		wantHazard     snippet.HazardKind
	}{
		{
			name:           "Level 1: simple linear commands with <= 1 pipe and <= 10 SLOC",
			code:           "# list go files\nset -euo pipefail\ngit status --short\nls -la | head -n 5\n",
			wantLevel:      1,
			wantRating:     "Transparent",
			wantStatus:     "approved",
			wantConfidence: snippet.ConfidenceAST,
		},
		{
			name:           "Level 2: simple if branch and 1 helper function",
			code:           "log_info() {\n  echo \"[INFO] $1\"\n}\nif [ -f \"README.md\" ]; then\n  log_info \"found readme\"\nfi\n",
			wantLevel:      2,
			wantRating:     "Moderate",
			wantStatus:     "approved",
			wantConfidence: snippet.ConfidenceAST,
		},
		{
			name:           "Level 3: nested control flow (depth 2) and 2-pipe chain",
			code:           "for f in *.go; do\n  if [ -f \"$f\" ]; then\n    cat \"$f\" | grep package | cut -d' ' -f2\n  fi\ndone\n",
			wantLevel:      3,
			wantRating:     "Complex",
			wantStatus:     "simplify_or_consult",
			wantConfidence: snippet.ConfidenceAST,
		},
		{
			name:           "Level 3: inline python -c script with import and multiple statements",
			code:           `python3 -c "import sys, json; data = json.load(sys.stdin); print(len(data))"`,
			wantLevel:      3,
			wantRating:     "Complex",
			wantStatus:     "simplify_or_consult",
			wantConfidence: snippet.ConfidenceAST,
			wantHazard:     snippet.HazardInlineInterpreter,
		},
		{
			name:           "Level 3: process substitution and high positional parameter $4",
			code:           `diff -u <(sort "$1") <(sort "$4")`,
			wantLevel:      3,
			wantRating:     "Complex",
			wantStatus:     "simplify_or_consult",
			wantConfidence: snippet.ConfidenceAST,
			wantHazard:     snippet.HazardProcessSubstitution,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := snippet.Analyze(tt.code, tt.lang)
			assert.True(t, rep.OK)
			assert.Equal(t, tt.wantLevel, rep.Level)
			assert.Equal(t, tt.wantRating, rep.Rating)
			assert.Equal(t, tt.wantStatus, rep.Status)
			assert.Equal(t, tt.wantConfidence, rep.Confidence)
			if tt.wantLevel > 1 {
				assert.NotEmpty(t, rep.Drivers)
			}
			if tt.wantHazard != "" {
				assertHasHazard(t, rep, tt.wantHazard)
			}
		})
	}
}

func TestAnalyzeOpaqueAndFallback(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		lang           string
		wantConfidence snippet.Confidence
		wantHazard     snippet.HazardKind
	}{
		{
			name:           "Level 4: eval hazard elevates to Opaque",
			code:           `eval "$USER_CMD"`,
			wantConfidence: snippet.ConfidenceAST,
			wantHazard:     snippet.HazardEval,
		},
		{
			name:           "Level 4: deep pipeline (>= 3 pipes) into shell",
			code:           `curl -fsSL https://example.com/install.sh | grep -v debug | sed 's/foo/bar/' | sh`,
			wantConfidence: snippet.ConfidenceAST,
			wantHazard:     snippet.HazardPipeToShell,
		},
		{
			name:           "Level 4: background execution and signal trap",
			code:           "trap 'kill 0' EXIT\n./server &\nwait\n",
			wantConfidence: snippet.ConfidenceAST,
		},
		{
			name:           "zsh syntax error falls back to heuristic confidence",
			code:           "if (( ${+commands[git]} )); then\n  eval \"echo hi\" | sh\nfi\n(( unclosed",
			lang:           "zsh",
			wantConfidence: snippet.ConfidenceHeuristic,
			wantHazard:     snippet.HazardEval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := snippet.Analyze(tt.code, tt.lang)
			assert.True(t, rep.OK)
			assert.Equal(t, 4, rep.Level)
			assert.Equal(t, "Opaque", rep.Rating)
			assert.Equal(t, "halt_and_decompose", rep.Status)
			assert.Equal(t, tt.wantConfidence, rep.Confidence)
			assert.NotEmpty(t, rep.Drivers)
			if tt.wantHazard != "" {
				assertHasHazard(t, rep, tt.wantHazard)
			}
		})
	}
}

func assertHasHazard(t *testing.T, rep snippet.AuditReport, want snippet.HazardKind) {
	t.Helper()
	var kinds []snippet.HazardKind
	for _, h := range rep.Hazards {
		kinds = append(kinds, h.Kind)
	}
	assert.Contains(t, kinds, want)
}

func TestAnalyzeFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "sample.sh")
	content := strings.Repeat("echo line\n", 12)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	rep, err := snippet.AnalyzeFile(path, "bash")
	require.NoError(t, err)
	assert.Equal(t, 2, rep.Level)
	assert.Equal(t, 12, rep.Metrics.SLOC)
}

func TestAnalyzePythonAST(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		wantLevel  int
		wantRating string
		wantStatus string
		wantHazard snippet.HazardKind
	}{
		{
			name:       "Python Level 1: linear script with docstring and comment",
			code:       "\"\"\"Module docstring.\nspanning two lines.\"\"\"\n# load config\nimport json\nimport sys\nprint(json.dumps({\"ok\": True}))\n",
			wantLevel:  1,
			wantRating: "Transparent",
			wantStatus: "approved",
		},
		{
			name:       "Python Level 2: 1 helper function and 1 if/elif branch",
			code:       "def greet(name):\n    return f\"hello {name}\"\n\nif True:\n    print(greet(\"world\"))\nelif False:\n    print(\"none\")\n",
			wantLevel:  2,
			wantRating: "Moderate",
			wantStatus: "approved",
		},
		{
			name:       "Python Level 3: nested loop + if (depth 2) and sys.argv[4]",
			code:       "import sys\narg = sys.argv[4]\nfor item in [1, 2, 3]:\n    if item > 1:\n        print(item, arg)\n",
			wantLevel:  3,
			wantRating: "Complex",
			wantStatus: "simplify_or_consult",
		},
		{
			name:       "Python Level 4: eval hazard elevates to Opaque",
			code:       "user_expr = \"1 + 2\"\nres = eval(user_expr)\n",
			wantLevel:  4,
			wantRating: "Opaque",
			wantStatus: "halt_and_decompose",
			wantHazard: snippet.HazardEval,
		},
		{
			name:       "Python Level 4: subprocess.run with shell=True",
			code:       "from subprocess import run\nrun(\"ls -la\", shell=True)\n",
			wantLevel:  4,
			wantRating: "Opaque",
			wantStatus: "halt_and_decompose",
			wantHazard: snippet.HazardPipeToShell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := snippet.Analyze(tt.code, "python")
			require.True(t, rep.OK, "expected OK=true, got error: %s", rep.ParseError)
			assert.Equal(t, "python", rep.Language)
			assert.Equal(t, snippet.ConfidenceAST, rep.Confidence)
			assert.Equal(t, tt.wantLevel, rep.Level)
			assert.Equal(t, tt.wantRating, rep.Rating)
			assert.Equal(t, tt.wantStatus, rep.Status)
			if tt.wantHazard != "" {
				assertHasHazard(t, rep, tt.wantHazard)
			}
		})
	}
}

func TestAnalyzePythonSyntaxErrorAndUnsupportedLang(t *testing.T) {
	pyErrRep := snippet.Analyze("def broken(\n    print(1)\n", "python")
	assert.False(t, pyErrRep.OK)
	assert.Equal(t, "analysis_failed", pyErrRep.Status)
	assert.Equal(t, "python", pyErrRep.Language)
	assert.Contains(t, pyErrRep.ParseError, "SyntaxError")

	unsupportedRep := snippet.Analyze("puts 'hello'", "ruby")
	assert.False(t, unsupportedRep.OK)
	assert.Equal(t, "analysis_failed", unsupportedRep.Status)
	assert.Contains(t, unsupportedRep.ParseError, "unsupported language")
}
