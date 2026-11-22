package pathutil_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/pathutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{
			name:  "replaces slashes and spaces with hyphens",
			input: "feat/my branch name",
			want:  "feat-my-branch-name",
		},
		{
			name:  "trims leading and trailing hyphens and underscores",
			input: " _-hello_world-_ ",
			want:  "hello_world",
		},
		{
			name:  "truncates slugs longer than 128 characters",
			input: strings.Repeat("a", 150),
			want:  strings.Repeat("a", 128),
		},
		{
			name:    "rejects empty after sanitization",
			input:   "///---___",
			wantErr: "invalid empty identifier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pathutil.SanitizeSlug(tt.input)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveProjectAndBranch(t *testing.T) {
	t.Run("explicit project and branch are sanitized", func(t *testing.T) {
		proj, br, err := pathutil.ResolveProjectAndBranch("my/repo", "feat/snippet-mgr", "")
		require.NoError(t, err)
		assert.Equal(t, "my-repo", proj)
		assert.Equal(t, "feat-snippet-mgr", br)
	})

	t.Run("rejects sensitive workDir", func(t *testing.T) {
		_, _, err := pathutil.ResolveProjectAndBranch("proj", "main", "/home/user/.ssh/keys")
		require.ErrorContains(t, err, "sensitive path")
	})

	t.Run("auto-derives project and branch from git repository", func(t *testing.T) {
		repoDir := filepath.Join(t.TempDir(), "sample-repo")
		runGitCmd(t, "", "init", "-b", "feat/test-branch", repoDir)
		runGitCmd(t, repoDir, "config", "user.email", "test@example.com")
		runGitCmd(t, repoDir, "config", "user.name", "Test User")
		runGitCmd(t, repoDir, "commit", "--allow-empty", "-m", "initial")

		proj, br, err := pathutil.ResolveProjectAndBranch("", "", repoDir)
		require.NoError(t, err)
		assert.Equal(t, "sample-repo", proj)
		assert.Equal(t, "feat-test-branch", br)
	})

	t.Run("falls back to short HEAD SHA on detached HEAD", func(t *testing.T) {
		repoDir := filepath.Join(t.TempDir(), "detached-repo")
		runGitCmd(t, "", "init", "-b", "main", repoDir)
		runGitCmd(t, repoDir, "config", "user.email", "test@example.com")
		runGitCmd(t, repoDir, "config", "user.name", "Test User")
		runGitCmd(t, repoDir, "commit", "--allow-empty", "-m", "initial")
		runGitCmd(t, repoDir, "checkout", "--detach", "HEAD")

		proj, br, err := pathutil.ResolveProjectAndBranch("", "", repoDir)
		require.NoError(t, err)
		assert.Equal(t, "detached-repo", proj)
		assert.NotEmpty(t, br)
		assert.NotEqual(t, "main", br)
	})
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, string(out))
}
