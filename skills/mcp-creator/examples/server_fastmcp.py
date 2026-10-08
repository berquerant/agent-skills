"""Example FastMCP server (Python mcp 2.0+ SDK)."""

from mcp.server.fastmcp import FastMCP

mcp = FastMCP("my-server")

# Sample in-memory store for demonstration
data_store = {"example": "sample content"}


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
