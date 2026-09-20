# MCP Specification Reference

Used by the mcp-creator skill. Covers key protocol concepts needed for
implementation. For the full specification, see:
https://modelcontextprotocol.io/specification/2026-07-28

---

## Architecture Overview

```
MCP Host (AI app, e.g. Claude)
  └── MCP Client  ──(dedicated connection)──  MCP Server
```

- **MCP Host**: AI application that manages one or more MCP clients.
- **MCP Client**: One per server. Maintains the connection, obtains context.
- **MCP Server**: Program that exposes tools, resources, and/or prompts.

---

## Transport Options

| Transport       | Use case               | Auth                          |
|-----------------|------------------------|-------------------------------|
| **stdio**       | Local, single-client   | None (process trust)          |
| **Streamable HTTP** | Remote, multi-client | Bearer token / OAuth (recommended) |

> For STDIO servers: **never write to stdout**. Use stderr for logs.
> Python: `logging.getLogger(__name__)` → writes to stderr.
> TypeScript: use `console.error()`.

---

## Primitives

### Tools — Model-controlled actions

Defined with a `name`, `description`, and `inputSchema` (JSON Schema).
The LLM decides when and how to call tools based on user intent.

```typescript
{
  name: "searchFlights",
  description: "Search for available flights",
  inputSchema: {
    type: "object",
    properties: {
      origin:      { type: "string" },
      destination: { type: "string" },
      date:        { type: "string", format: "date" }
    },
    required: ["origin", "destination", "date"]
  }
}
```

Protocol methods: `tools/list`, `tools/call`

### Resources — Application-controlled context (read-only)

Identified by URIs. Provide passive data for context (files, DB schemas, etc.).

Protocol methods: `resources/list`, `resources/read`

### Prompts — User-controlled templates

Pre-built instruction templates that guide the model. Surfaced in the host UI
for users to select.

Protocol methods: `prompts/list`, `prompts/get`

---

## Protocol Basics

- Based on **JSON-RPC 2.0**.
- **Stateless**: every request carries `_meta` with protocol version and
  client capabilities.
- **Discovery**: clients send `server/discover` to learn what the server
  supports. Response is cacheable (`ttlMs`).
- **Notifications**: opt-in change events via `subscriptions/listen`
  (e.g., `notifications/tools/list_changed`).

---

## Python SDK Quick Reference (mcp 2.0+)

```python
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("my-server")

# Tool
@mcp.tool()
async def my_tool(x: int, y: int) -> int:
    """Add two numbers."""
    return x + y

# Resource
@mcp.resource("data://{key}")
async def get_data(key: str) -> str:
    """Return stored content by key."""
    return data_store[key]

# Prompt
@mcp.prompt()
def my_prompt(topic: str) -> str:
    """Prompt template for domain expert persona."""
    return f"You are an expert on {topic}."

# Run (stdio)
if __name__ == "__main__":
    mcp.run()
```

Dev/inspection:
```bash
uv run mcp dev server.py
```

---

## Host Configuration
 
Most MCP hosts use a standardized JSON format to launch STDIO servers:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "uv",
      "args": ["run", "--directory", "/path/to/project", "python", "server.py"]
    }
  }
}
```

### Typical Host Configuration Paths
- **Claude for Desktop**: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) / `%APPDATA%\Claude\claude_desktop_config.json` (Windows)
- **Antigravity CLI**: `.agents/mcp.json` (workspace) or global configuration
- **VS Code / Cursor / Codex / Other IDEs**: Project workspace `.vscode/mcp.json`, `.cursor/mcp.json`, or relevant client settings


---

## Further Reading

- Getting started: https://modelcontextprotocol.io/docs/2026-07-28/getting-started/intro
- Architecture: https://modelcontextprotocol.io/docs/2026-07-28/learn/architecture
- Server concepts: https://modelcontextprotocol.io/docs/2026-07-28/learn/server-concepts
- Build a server: https://modelcontextprotocol.io/docs/2026-07-28/develop/build-server
- Full specification: https://modelcontextprotocol.io/specification/2026-07-28
- SDKs: https://modelcontextprotocol.io/docs/2026-07-28/sdk
