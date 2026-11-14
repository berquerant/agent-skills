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

if [ ! -x "${mcp_bin}" ]; then
  (cd "${repo_root}/mcp" && go build -o "${mcp_bin}" ./cmd/agent-skills-mcp)
fi

exec "${mcp_bin}" skill-lint --strict "$@"
