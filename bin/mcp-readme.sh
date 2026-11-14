#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
readonly script_dir repo_root
readonly mcp_dir="${repo_root}/mcp"

generate() {
  cd "${mcp_dir}"
  local tmp_dir
  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "${tmp_dir}"' RETURN
  local tmp_bin="${tmp_dir}/agent-skills-mcp"
  go build -o "${tmp_bin}" ./cmd/agent-skills-mcp
  local help_out
  help_out="$("${tmp_bin}" --help 2>&1)"

  cat <<README_EOF
# agent-skills-mcp

${help_out}

## Requirements

- Go \`1.27.4\` or later

## Installation

\`\`\`bash
# Install latest from main
go install github.com/berquerant/agent-skills/mcp/cmd/agent-skills-mcp@latest

# Install a specific release tag (subdirectory module tag format: mcp/vX.Y.Z)
go install github.com/berquerant/agent-skills/mcp/cmd/agent-skills-mcp@mcp/v0.1.0
\`\`\`

## Host Configuration Example

\`\`\`json
{
  "mcpServers": {
    "agent-skills": {
      "command": "agent-skills-mcp",
      "args": ["serve", "--pace-guard=false"]
    }
  }
}
\`\`\`
README_EOF
}

generate
