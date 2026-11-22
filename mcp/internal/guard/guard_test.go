package guard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSensitivePath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		wantSensitive bool
		wantReasonSub string
	}{
		{name: "exact .env", path: "/app/.env", wantSensitive: true, wantReasonSub: ".env"},
		{name: "uppercase .ENV", path: "/app/.ENV", wantSensitive: true, wantReasonSub: ".env"},
		{name: "flag-prefixed dotenv", path: "--env-file=.env.production", wantSensitive: true, wantReasonSub: ".env.production"},
		{name: "dotenv variant .env.custom", path: "config/.env.custom", wantSensitive: true, wantReasonSub: "dotenv"},
		{name: "private key filename id_ed25519", path: "/home/user/id_ed25519", wantSensitive: true, wantReasonSub: "id_ed25519"},
		{name: "pem extension", path: "/certs/server.PEM", wantSensitive: true, wantReasonSub: ".pem"},
		{name: "aws credentials path", path: "/path/to/home/.aws/credentials", wantSensitive: true, wantReasonSub: "AWS"},
		{name: "ssh config path", path: "/path/to/home/.ssh/config", wantSensitive: true, wantReasonSub: "SSH"},
		{name: "docker config.json", path: "/root/.docker/config.json", wantSensitive: true, wantReasonSub: "Docker"},
		{name: "github cli hosts.yml", path: "/home/user/.config/gh/hosts.yml", wantSensitive: true, wantReasonSub: "GitHub CLI"},
		{name: "etc shadow", path: "/etc/shadow", wantSensitive: true, wantReasonSub: "System auth"},
		{name: "normal markdown file", path: "/tmp/checkpoints/proj/main/ctx/checkpoint-1.md", wantSensitive: false},
		{name: "normal go source file", path: "mcp/internal/guard/paths.go", wantSensitive: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reason := guard.IsSensitivePath(tt.path)
			assert.Equal(t, tt.wantSensitive, got)
			if tt.wantSensitive {
				assert.Contains(t, reason, tt.wantReasonSub)
				require.Error(t, guard.CheckPathAllowed(tt.path))
			} else {
				require.NoError(t, guard.CheckPathAllowed(tt.path))
			}
		})
	}
}

func TestScanTextForSecrets(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantRule string
	}{
		{name: "AWS Access Key", text: "export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE", wantRule: "AWS Access Key"},
		{name: "GitHub Token", text: "token: " + "gh" + "p_1234567890abcdefghijklmnopqrstuvwxyzABCD", wantRule: "GitHub Token"},
		{name: "Slack Bot Token", text: "xox" + "b-123456789012-123456789012-abcdefABCDEF123456789012", wantRule: "Slack Token"},
		{name: "OpenAI Key", text: "sk-" + "proj-1234567890abcdefghij1234567890abcdefghij", wantRule: "OpenAI API Key"},
		{name: "Private Key Header", text: "-----BEGIN OPENSSH PRIVATE KEY-----", wantRule: "Private Key Block"},
		{name: "JWT", text: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abcdef123456", wantRule: "JSON Web Token"},
		{name: "URI with credentials", text: "postgres://admin:s3cr3t@db.example.com:5432/prod", wantRule: "URI with Credentials"},
		{name: "Safe normal text", text: "go test -cover -race ./...", wantRule: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := guard.ScanTextForSecrets(tt.text)
			if tt.wantRule == "" {
				assert.Empty(t, got)
				require.NoError(t, guard.CheckNoSecrets(tt.text))
			} else {
				assert.Contains(t, got, tt.wantRule)
				require.Error(t, guard.CheckNoSecrets(tt.text))
				redacted := guard.RedactSecretsInString(tt.text)
				assert.Contains(t, redacted, guard.RedactedMask)
			}
		})
	}
}

func TestSanitizeForLog(t *testing.T) {
	payload := map[string]any{
		"user":          "alice",
		"client_secret": "top-secret-value",
		"nested": map[string]any{
			"database_url": "postgres://u:p@localhost/db",
			"note":         "leaked AKIAIOSFODNN7EXAMPLE in string",
		},
	}
	sanitized, ok := guard.SanitizeForLog(payload).(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "alice", sanitized["user"])
	assert.Equal(t, guard.RedactedMask, sanitized["client_secret"])

	nested, ok := sanitized["nested"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, guard.RedactedMask, nested["database_url"])
	assert.Contains(t, nested["note"], guard.RedactedMask)
	assert.NotContains(t, nested["note"], "AKIAIOSFODNN7EXAMPLE")
}

func TestCheckPathAllowed_Symlink(t *testing.T) {
	tmp := t.TempDir()
	sshDir := filepath.Join(tmp, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0o700))
	sshCfg := filepath.Join(sshDir, "config")
	require.NoError(t, os.WriteFile(sshCfg, []byte("Host *"), 0o600))

	safeFile := filepath.Join(tmp, "safe.log")
	require.NoError(t, os.WriteFile(safeFile, []byte("ok"), 0o600))

	linkToSSHFile := filepath.Join(tmp, "benign.log")
	require.NoError(t, os.Symlink(sshCfg, linkToSSHFile))

	linkToSSHDir := filepath.Join(tmp, "dir_link")
	require.NoError(t, os.Symlink(sshDir, linkToSSHDir))

	linkToSafe := filepath.Join(tmp, "safe_link.log")
	require.NoError(t, os.Symlink(safeFile, linkToSafe))

	require.Error(t, guard.CheckPathAllowed(linkToSSHFile))
	require.Error(t, guard.CheckPathAllowed(filepath.Join(linkToSSHDir, "nonexistent_file")))
	require.NoError(t, guard.CheckPathAllowed(linkToSafe))
}
