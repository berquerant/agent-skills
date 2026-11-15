# agent-skills-mcp

Single-binary Model Context Protocol (MCP) server for `agent-skills`, implemented in Go.
Every feature module can be independently enabled or disabled via CLI flags or environment variables.

## Usage

```shell
# Run MCP server over stdio (default subcommand)
agent-skills-mcp [serve] [flags]

# Run automated skill checks directly without running the MCP server
agent-skills-mcp skill-lint [--strict] [--format=markdown|json] [path...]

# Disable specific features
agent-skills-mcp serve --pace-guard=false --pr-triage=false

# Print build version
agent-skills-mcp version
```

## Configuration & Precedence

Configuration values are resolved in the following order of precedence (highest to lowest):
1. Command-line flags (e.g. `--task-dag=false`)
2. Environment variables prefixed with `AGENT_SKILLS_MCP_` (e.g. `AGENT_SKILLS_MCP_TASK_DAG=false`)
3. Default values defined for each flag

Environment variables are derived from flag names in uppercase with hyphens replaced by underscores.
e.g. `--snippet-analyzer` -> `AGENT_SKILLS_MCP_SNIPPET_ANALYZER`

## Flags

```text
      --cache[=true|false]              enable cache feature (default true)
      --cache-dir string                cache directory (default: $XDG_CACHE_HOME/agent-skills-mcp/cache)
      --cache-max-entry-size int        maximum uncompressed cache entry size in bytes (default 10485760)
      --cache-max-total-size int        maximum total compressed cache size in bytes (default 524288000)
      --cache-max-ttl int               maximum allowed cache TTL in seconds (default 1800)
      --checkpoint[=true|false]         enable checkpoint feature (default true)
      --checkpoint-root string          checkpoint root directory (default: $XDG_CACHE_HOME/agent-skills-mcp/checkpoints)
      --log-file string                 optional log file path (default: stderr)
      --log-level string                log level (debug, info, warn, error) (default "warn")
      --max-output-size int             maximum serialized tool output size in bytes (default 32768)
      --pace-guard[=true|false]         enable pace-guard feature (default true)
      --pr-triage[=true|false]          enable pr-triage feature (default true)
      --skill-lint[=true|false]         enable skill-lint feature (default true)
      --snippet-analyzer[=true|false]   enable snippet-analyzer feature (default true)
      --state-dir string                state directory (default: $XDG_STATE_HOME/agent-skills-mcp)
      --task-dag[=true|false]           enable task-dag feature (default true)
```

## Requirements

- Go `1.27.4` or later

## Installation

```bash
# Install latest from main
go install github.com/berquerant/agent-skills/mcp/cmd/agent-skills-mcp@latest

# Install a specific release tag (subdirectory module tag format: mcp/vX.Y.Z)
go install github.com/berquerant/agent-skills/mcp/cmd/agent-skills-mcp@mcp/v0.1.0
```

## Host Configuration Example

```json
{
  "mcpServers": {
    "agent-skills": {
      "command": "agent-skills-mcp",
      "args": ["serve", "--pace-guard=false"]
    }
  }
}
```

## Architecture & Tools Reference

For internal system architecture, security guardrails, and the complete catalog of tools registered across all 7 features, see [references/architecture.md](references/architecture.md).
