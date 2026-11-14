package cache_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T) *cache.Service {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "cache")
	clk := newManualClock(5000)
	svc, err := cache.NewService(cache.Config{
		Dir:                         dir,
		MaxTTL:                      1800,
		MaxCacheSize:                1024 * 1024,
		MaxTotalCompressedCacheSize: 10 * 1024 * 1024,
		Clock:                       clk,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	return svc
}

func TestServiceSet(t *testing.T) {
	svc := newTestService(t)

	tests := []struct {
		name        string
		in          cache.SetInput
		wantSuccess bool
		wantMsgSub  string
	}{
		{
			name:        "rejects non-positive ttl",
			in:          cache.SetInput{Type: "command", Name: "echo", Dir: "/tmp", TTL: 0, Stdout: "hi"},
			wantSuccess: false,
			wantMsgSub:  "ttl must be > 0",
		},
		{
			name:        "stores v1 by computed key",
			in:          cache.SetInput{Type: "command", Name: "echo", Dir: "/tmp", Args: []string{"hello"}, TTL: 300, Stdout: "hello-1\n", Success: true},
			wantSuccess: true,
		},
		{
			name:        "stores v1 by explicit key",
			in:          cache.SetInput{Key: "svc_key", Type: "command", Name: "echo", Dir: "/tmp", TTL: 300, Stdout: "svc-v1", Success: true},
			wantSuccess: true,
		},
		{
			name:        "rejects secret token in stdout",
			in:          cache.SetInput{Type: "command", Name: "echo", Dir: "/tmp", TTL: 300, Stdout: "token=" + "gh" + "p_1234567890abcdefghijklmnopqrstuv", Success: true},
			wantSuccess: false,
			wantMsgSub:  "sensitive secret pattern detected",
		},
		{
			name:        "rejects sensitive file path in args",
			in:          cache.SetInput{Type: "command", Name: "cat", Dir: "/tmp", Args: []string{".env.production"}, TTL: 300, Stdout: "ok", Success: true},
			wantSuccess: false,
			wantMsgSub:  "access to sensitive path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := svc.Set(tt.in)
			assert.Equal(t, tt.wantSuccess, res["success"])
			if tt.wantMsgSub != "" {
				assert.Contains(t, fmt.Sprint(res["message"]), tt.wantMsgSub)
			}
		})
	}
}

func TestServiceQueryVersionAndPurge(t *testing.T) {
	svc := newTestService(t)
	_ = svc.Set(cache.SetInput{Type: "command", Name: "echo", Dir: "/tmp", Args: []string{"hello"}, TTL: 300, Stdout: "h1", Success: true})
	_ = svc.Set(cache.SetInput{Key: "svc_key", Type: "command", Name: "echo", Dir: "/tmp", TTL: 300, Stdout: "v1", Success: true})
	_ = svc.Set(cache.SetInput{Key: "svc_key", Type: "command", Name: "echo", Dir: "/tmp", TTL: 300, Stdout: "v2", Success: true})

	tests := []struct {
		name       string
		op         string
		in         cache.QueryInput
		wantHit    bool
		wantReason string
	}{
		{name: "get missing key and triple", op: "get", in: cache.QueryInput{}, wantHit: false},
		{name: "get invalid explicit version 0", op: "get", in: cache.QueryInput{Key: "svc_key", Version: 0}, wantHit: false, wantReason: "invalid_version"},
		{name: "get hit latest for svc_key", op: "get", in: cache.QueryInput{Key: "svc_key"}, wantHit: true},
		{name: "exist missing key and triple", op: "exist", in: cache.QueryInput{}, wantHit: false},
		{name: "exist hit by triple", op: "exist", in: cache.QueryInput{Type: "command", Name: "echo", Dir: "/tmp", Args: []string{"hello"}}, wantHit: true},
		{name: "get rejects sensitive path in args", op: "get", in: cache.QueryInput{Type: "command", Name: "cat", Dir: "/tmp", Args: []string{"/etc/shadow"}}, wantHit: false, wantReason: "sensitive_path"},
		{name: "delete missing key and triple", op: "delete", in: cache.QueryInput{}, wantHit: false},
		{name: "delete specific version 1 of svc_key", op: "delete", in: cache.QueryInput{Key: "svc_key", Version: 1}, wantHit: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res map[string]any
			switch tt.op {
			case "get":
				res = svc.Get(tt.in)
			case "exist":
				res = svc.Exist(tt.in)
			case "delete":
				res = svc.Delete(tt.in)
			}
			assert.Equal(t, tt.wantHit, res["hit"])
			if tt.wantReason != "" {
				assert.Equal(t, tt.wantReason, res["reason"])
			}
		})
	}

	assert.Equal(t, false, svc.Version(cache.VersionInput{})["hit"])
	assert.Equal(t, true, svc.Version(cache.VersionInput{Key: "svc_key"})["hit"])
	assert.Equal(t, true, svc.Purge(cache.PurgeInput{All: true})["success"])
}
