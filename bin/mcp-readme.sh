#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
readonly script_dir repo_root
readonly mcp_dir="${repo_root}/mcp"
readonly mcp_bin="${mcp_dir}/bin/agent-skills-mcp"

generate() {
  (cd "${mcp_dir}" && go build -o "${mcp_bin}" ./cmd/agent-skills-mcp)
  local help_out
  help_out="$("${mcp_bin}" --help 2>&1)"

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

## Architecture & Tools Reference

For internal system architecture, security guardrails, and the complete catalog of tools registered across all 8 features, see [references/architecture.md](references/architecture.md).
README_EOF
}

generate
