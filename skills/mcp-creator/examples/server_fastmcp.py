"""Example MCP server using the Python mcp 2.0+ SDK (MCPServer).

Note: For legacy mcp 1.x environments, replace `from mcp.server import MCPServer`
with `from mcp.server.fastmcp import FastMCP`.
"""

from typing import Annotated

from mcp.server import MCPServer
from pydantic import Field

mcp = MCPServer("my-server")

# Sample in-memory store for demonstration
data_store = {"example": "sample content"}


# Tool
@mcp.tool()
async def my_tool(
    x: Annotated[int, Field(description="First addend", ge=-1_000_000, le=1_000_000)],
    y: Annotated[int, Field(description="Second addend", ge=-1_000_000, le=1_000_000)],
) -> int:
    """Add two bounded numbers."""
    return x + y


# Resource
@mcp.resource("data://{key}")
async def get_data(key: str) -> str:
    """Return stored content by key."""
    if key not in data_store:
        raise ValueError(f"Key not found: {key}")
    return data_store[key]


# Prompt
@mcp.prompt()
def my_prompt(topic: str) -> str:
    """Prompt template for domain expert persona."""
    return f"You are an expert on {topic}."


# Run (stdio)
if __name__ == "__main__":
    mcp.run()
