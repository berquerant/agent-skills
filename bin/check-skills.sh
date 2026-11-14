#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
readonly script_dir repo_root
readonly mcp_bin="${repo_root}/mcp/bin/agent-skills-mcp"

cd "${repo_root}"

if command -v agent-skills-mcp >/dev/null 2>&1; then
  exec agent-skills-mcp skill-lint --strict "$@"
fi

if [ -x "${mcp_bin}" ]; then
  exec "${mcp_bin}" skill-lint --strict "$@"
fi

exec go run ./mcp/cmd/agent-skills-mcp skill-lint --strict "$@"
