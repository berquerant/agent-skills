package snippetmgr_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/berquerant/agent-skills/mcp/internal/snippetmgr"
	"github.com/berquerant/agent-skills/mcp/internal/testutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleBashV1 = `#!/usr/bin/env bash
#
# Description:
#   Prints repository status for a target directory.
#
# Usage:
#   ./inspect_repo_status.sh /path/to/repo
#
set -euo pipefail

target_dir="${1:?target directory required}"
cd "${target_dir}"
git status -sb
`

const sampleBashV2 = `#!/usr/bin/env bash
#
# Description:
#   Prints repository status and current branch for a target directory.
#
# Usage:
#   ./inspect_repo_status.sh /path/to/repo
#
set -euo pipefail

target_dir="${1:?target directory required}"
cd "${target_dir}"
git status -sb
git branch --show-current
`

const sampleComplexBash = `#!/usr/bin/env bash
#
# Description:
#   Complex multi-function snippet triggering Level 3+.
#
# Usage:
#   ./complex_workflow.sh
#
set -euo pipefail

fn_a() { echo "a"; }
fn_b() { echo "b"; }
fn_c() { echo "c"; }

if true; then
  if true; then
    if true; then
      fn_a
      fn_b
      fn_c
    fi
  fi
fi
`

const samplePythonSnippet = `#!/usr/bin/env python3
"""
Description:
  Prints JSON summary of input arguments.

Usage:
  ./summarize_args.py arg1 arg2
"""
import json
import sys

print(json.dumps(sys.argv[1:]))
`

func TestManagerComplexityGate(t *testing.T) {
	mgr, err := snippetmgr.NewManager(t.TempDir())
	require.NoError(t, err)
	t0 := time.Date(2026, 11, 23, 1, 0, 0, 0, time.UTC)

	blocked, err := mgr.Create(snippetmgr.CreateInput{
		Name:    "complex_workflow",
		Code:    sampleComplexBash,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0)
	require.NoError(t, err)
	assert.False(t, blocked.Saved)
	assert.Equal(t, "simplify_or_consult", blocked.ActionRequired)
	assert.GreaterOrEqual(t, blocked.Audit.Level, 3)

	allowed, err := mgr.Create(snippetmgr.CreateInput{
		Name:         "complex_workflow",
		Code:         sampleComplexBash,
		Project:      "agent-skills",
		Branch:       "feat/snippet-mgr",
		AllowComplex: true,
	}, t0)
	require.NoError(t, err)
	assert.True(t, allowed.Saved)
	assert.Equal(t, 1, allowed.Version)
}

func TestManagerVersioningAndVerify(t *testing.T) {
	mgr, err := snippetmgr.NewManager(t.TempDir())
	require.NoError(t, err)
	t0 := time.Date(2026, 11, 23, 1, 0, 0, 0, time.UTC)

	v1, err := mgr.Create(snippetmgr.CreateInput{
		Name:    "inspect_repo_status",
		Code:    sampleBashV1,
		Scope:   snippetmgr.ScopeReusable,
		Purpose: "Inspect repository git status",
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0)
	require.NoError(t, err)
	assert.True(t, v1.Saved)
	assert.Equal(t, 1, v1.Version)
	assert.Contains(t, v1.DualPath, "](file://")
	assert.Contains(t, v1.PreExecutionNotice, "### 📋 Snippet Pre-Execution Notice")

	infoV1, err := os.Stat(v1.Path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o500), infoV1.Mode().Perm())

	v2, err := mgr.Create(snippetmgr.CreateInput{
		Name:    "inspect_repo_status",
		Code:    sampleBashV2,
		Scope:   snippetmgr.ScopeReusable,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0.Add(time.Minute))
	require.NoError(t, err)
	assert.True(t, v2.Saved)
	assert.Equal(t, 2, v2.Version)
	assert.Equal(t, []int{1}, v2.DeprecatedVersions)

	infoV1After, err := os.Stat(v1.Path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o400), infoV1After.Mode().Perm())

	ver2, err := mgr.Verify(snippetmgr.VerifyInput{
		Name:    "inspect_repo_status",
		Scope:   snippetmgr.ScopeReusable,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	})
	require.NoError(t, err)
	assert.True(t, ver2.Valid)
	assert.True(t, ver2.Executable)
	assert.True(t, ver2.Immutable)

	ver1, err := mgr.Verify(snippetmgr.VerifyInput{
		Name:    "inspect_repo_status",
		Scope:   snippetmgr.ScopeReusable,
		Version: 1,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	})
	require.NoError(t, err)
	assert.False(t, ver1.Valid)
	assert.Equal(t, snippetmgr.StatusDeprecated, ver1.Status)

	// Tamper detection on v2
	require.NoError(t, os.Chmod(v2.Path, 0o600))
	require.NoError(t, os.WriteFile(v2.Path, []byte(sampleBashV2+"\n# tampered\n"), 0o600))
	require.NoError(t, os.Chmod(v2.Path, 0o500))

	tampered, err := mgr.Verify(snippetmgr.VerifyInput{
		Name:    "inspect_repo_status",
		Scope:   snippetmgr.ScopeReusable,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	})
	require.NoError(t, err)
	assert.False(t, tampered.Valid)
	assert.Contains(t, strings.Join(tampered.Issues, " "), "SHA-256 digest mismatch")
}

func TestManagerInvocationAndListing(t *testing.T) {
	mgr, err := snippetmgr.NewManager(t.TempDir())
	require.NoError(t, err)
	t0 := time.Date(2026, 11, 23, 1, 0, 0, 0, time.UTC)

	_, err = mgr.Create(snippetmgr.CreateInput{
		Name:    "inspect_repo_status",
		Code:    sampleBashV1,
		Scope:   snippetmgr.ScopeReusable,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0)
	require.NoError(t, err)

	_, err = mgr.Create(snippetmgr.CreateInput{
		Name:    "inspect_repo_status",
		Code:    sampleBashV2,
		Scope:   snippetmgr.ScopeReusable,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0.Add(time.Minute))
	require.NoError(t, err)

	recOut, err := mgr.RecordInvocation(snippetmgr.RecordInvocationInput{
		Name:       "inspect_repo_status",
		Scope:      snippetmgr.ScopeReusable,
		Args:       []string{"/path/to/repo"},
		ExitCode:   0,
		DurationMs: 42,
		Note:       "clean status",
		Project:    "agent-skills",
		Branch:     "feat/snippet-mgr",
	}, t0.Add(2*time.Minute))
	require.NoError(t, err)
	assert.True(t, recOut.Recorded)
	assert.Equal(t, 2, recOut.Version)
	assert.Equal(t, 1, recOut.TotalRecorded)

	_, err = mgr.RecordInvocation(snippetmgr.RecordInvocationInput{
		Name:    "inspect_repo_status",
		Scope:   snippetmgr.ScopeReusable,
		Args:    []string{"ghp_123456789012345678901234567890123456"},
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0.Add(2*time.Minute))
	require.ErrorContains(t, err, "secret")

	got, err := mgr.Get(snippetmgr.GetInput{
		Name:    "inspect_repo_status",
		Scope:   snippetmgr.ScopeReusable,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, got.Meta.LatestVersion)
	assert.Len(t, got.Meta.Invocations, 1)

	listed, err := mgr.List(snippetmgr.ListInput{
		Project: "agent-skills",
		Scope:   snippetmgr.ScopeReusable,
		Status:  snippetmgr.StatusActive,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, listed.Total)
	assert.Equal(t, "inspect_repo_status", listed.Snippets[0].Name)
}

func TestManagerDeprecateAndCleanup(t *testing.T) {
	mgr, err := snippetmgr.NewManager(t.TempDir())
	require.NoError(t, err)
	t0 := time.Date(2026, 11, 23, 1, 0, 0, 0, time.UTC)

	_, err = mgr.Create(snippetmgr.CreateInput{
		Name:    "one_off_check",
		Code:    sampleBashV1,
		Scope:   snippetmgr.ScopeOneOff,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0)
	require.NoError(t, err)

	v1, err := mgr.Create(snippetmgr.CreateInput{
		Name:    "inspect_repo_status",
		Code:    sampleBashV1,
		Scope:   snippetmgr.ScopeReusable,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0)
	require.NoError(t, err)

	v2, err := mgr.Create(snippetmgr.CreateInput{
		Name:    "inspect_repo_status",
		Code:    sampleBashV2,
		Scope:   snippetmgr.ScopeReusable,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0.Add(time.Minute))
	require.NoError(t, err)

	dry, err := mgr.Cleanup(snippetmgr.CleanupInput{
		Project:             "agent-skills",
		IncludeActiveOneOff: true,
		DryRun:              true,
	}, t0.Add(3*time.Minute))
	require.NoError(t, err)
	assert.True(t, dry.DryRun)
	assert.Equal(t, 2, dry.RemovedCount)

	realClean, err := mgr.Cleanup(snippetmgr.CleanupInput{
		Project:             "agent-skills",
		IncludeActiveOneOff: true,
		DryRun:              false,
	}, t0.Add(4*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 2, realClean.RemovedCount)
	_, statErr := os.Stat(v1.Path)
	assert.True(t, os.IsNotExist(statErr))
	_, statErr2 := os.Stat(v2.Path)
	assert.NoError(t, statErr2, "active reusable snippet must not be removed")

	dep, err := mgr.Deprecate(snippetmgr.DeprecateInput{
		Name:    "inspect_repo_status",
		Scope:   snippetmgr.ScopeReusable,
		Project: "agent-skills",
		Branch:  "feat/snippet-mgr",
	}, t0.Add(5*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, []int{2}, dep.DeprecatedVersions)
}

func TestFeatureMCPIntegration(t *testing.T) {
	ctx := t.Context()
	tmp := t.TempDir()
	now := time.Date(2026, 11, 23, 2, 0, 0, 0, time.UTC)

	srv := mcp.NewServer(&mcp.Implementation{Name: "test-snippetmgr", Version: "1"}, nil)
	f := snippetmgr.NewFeature()
	require.Equal(t, "snippet-manager", f.Name())
	require.Len(t, f.Tools(), 8)
	require.NoError(t, f.Register(srv, feature.Env{
		SnippetDir:    tmp,
		Clock:         testutil.NewManualClock(now),
		MaxOutputSize: feature.DefaultMaxOutputSize,
	}))
	t.Cleanup(func() { _ = f.Close() })

	session := testutil.ConnectInMemory(t, srv)

	// Call snippet_check_inline
	checkRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "snippet_check_inline",
		Arguments: map[string]any{"command": "gh pr list --state open"},
	})
	require.NoError(t, err)
	require.False(t, checkRes.IsError)
	var checkOut snippetmgr.CheckInlineOutput
	require.NoError(t, json.Unmarshal([]byte(checkRes.Content[0].(*mcp.TextContent).Text), &checkOut))
	assert.True(t, checkOut.ShouldExtract)

	// Call snippet_create for Python snippet
	createRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "snippet_create",
		Arguments: map[string]any{
			"name":     "summarize_args",
			"code":     samplePythonSnippet,
			"language": "python",
			"project":  "agent-skills",
			"branch":   "main",
		},
	})
	require.NoError(t, err)
	require.False(t, createRes.IsError)
	var createOut snippetmgr.CreateOutput
	require.NoError(t, json.Unmarshal([]byte(createRes.Content[0].(*mcp.TextContent).Text), &createOut))
	assert.True(t, createOut.Saved)
	assert.Equal(t, "python", createOut.Language)
	assert.True(t, strings.HasSuffix(createOut.Path, "_v1.py"))
}
