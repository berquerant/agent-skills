---
name: mcp-creator
description: >-
  Use this skill when the user wants to create a new MCP (Model Context
  Protocol) server. Guides through requirements collection, generates
  implementation code with tools/resources/prompts, and scaffolds the
  project directory structure.
---

# MCP Creator

Use this skill to create a new MCP server from scratch.
Collect requirements, scaffold the project, and generate working
server code that exposes tools, resources, and/or prompts.

For MCP concepts and protocol details, refer to
[references/mcp-spec.md](references/mcp-spec.md).

---

## Step 1: Requirements Collection

If the user's intent is clear, draft a proposal and ask for confirmation.
If it is ambiguous, ask for the following information.

**Information to collect:**

1. **Purpose** — What does this MCP server expose? (e.g., "query a PostgreSQL
   database", "wrap a REST API", "read local files")
2. **Primitives to implement** — Which of the following?
   - **Tools** — Functions the LLM can call (write, compute, call APIs)
   - **Resources** — Read-only data the AI app can read for context
   - **Prompts** — Reusable instruction templates
3. **Transport** — `stdio` (local, single-client) or `streamable-http` (remote,
   multi-client)?
4. **Language / runtime** — Python (recommended: `mcp[cli]` SDK), TypeScript,
   or other?
5. **Placement** — Where should the project be created?
6. **Authentication** — Does the server need auth? (relevant for HTTP
   transport; OAuth recommended)

Verify: Server requirements (purpose, primitives, transport, runtime, and placement) are explicitly confirmed with the user.

---

## Step 2: Design the Interface

Before writing code, define the server's surface area.

For each **Tool**:
- `name` (snake_case, unique within server)
- `description` (what it does and when the LLM should call it)
- `inputSchema` (JSON Schema — required and optional parameters with types)
- Return value shape

For each **Resource**:
- URI pattern (e.g., `file:///path/{filename}`, `db://table/{table}`)
- MIME type
- Whether it is static or dynamic

For each **Prompt**:
- `name`
- `description`
- Arguments (if any)
- Template text

Verify: The interface is specific enough to implement without ambiguity.

---

## Step 3: Scaffold the Project

### Python (recommended)

```bash
uv init <project-name>
cd <project-name>
uv venv && source .venv/bin/activate
uv add "mcp[cli]"
touch server.py
```

### TypeScript

```bash
npm init -y
npm install @modelcontextprotocol/sdk
```

> [!NOTE]
> For STDIO servers: **never write to stdout** (it corrupts JSON-RPC messages).
> Use `logging` to stderr in Python, or `console.error` in TypeScript.

Verify: `uv run mcp dev server.py` (Python) or equivalent starts without errors.

---

## Step 4: Implement the Server

Follow the patterns in [references/mcp-spec.md](references/mcp-spec.md) for
detailed protocol rules. Key implementation points:

### Tools (Python example)

```python
from mcp.server import MCPServer

mcp = MCPServer("<server-name>")

@mcp.tool()
async def my_tool(param: str) -> str:
    """Description used as the tool's description for the LLM.

    Args:
        param: Description of this parameter.
    """
    # implementation
    return result
```

### Resources (Python example)

```python
@mcp.resource("myscheme://{name}")
async def my_resource(name: str) -> str:
    """Return content for the resource URI."""
    return content
```

### Prompts (Python example)

```python
@mcp.prompt()
def my_prompt(topic: str) -> str:
    """Prompt description."""
    return f"You are an expert on {topic}. ..."
```

### Running the server

```bash
# Development / inspection
uv run mcp dev server.py

# Production (stdio)
uv run python server.py
```

Verify: Use `mcp dev` to inspect registered tools/resources/prompts interactively.

---

## Step 5: Connect to a Host

Register the server in your MCP client/host configuration. Most hosts use a JSON
configuration specifying the executable command and arguments:

```json
{
  "mcpServers": {
    "<server-name>": {
      "command": "uv",
      "args": ["run", "--directory", "/path/to/project", "python", "server.py"]
    }
  }
}
```

### Common Host Locations:
- **Claude for Desktop**: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows)
- **Antigravity CLI**: `.agents/mcp.json` (workspace) or global config
- **VS Code / Cursor / Other MCP-compatible Clients**: Project-local or user-level MCP configuration file

Verify: The host discovers the server, lists its capabilities (tools/resources/prompts), and can execute them.


---

## Step 6: Confirm and Follow Up

1. Show the user the generated files and directory structure.
2. Confirm that the server registered correctly with the target host.
3. Suggest next steps:
   - Add error handling and logging
   - Write tests using `mcp.test_client()`
   - Publish to the [MCP Registry](https://registry.modelcontextprotocol.io)
     if intended for public use

> [!TIP]
> Run `uv run mcp dev server.py` during development to interactively inspect
> and test tools, resources, and prompts without a full host client.

Verify: Generated files and host registration are confirmed, and next steps are presented to the user.

---

## Guidelines

- **Never pollute stdout in STDIO servers.** Writing non-JSON-RPC output to standard output breaks protocol framing; always send logs to stderr.
- **Strict schema validation.** Always declare clear types and descriptions for tool inputs so host LLMs can invoke them accurately.
- **Verify host compatibility.** Ensure registered tool and resource names are unique and match the transport specifications before deployment.
- **Secure by default.** Validate and sanitize resource URIs and tool parameters to prevent path traversal or unauthorized access.
