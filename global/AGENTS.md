# Global AGENTS.md

Operational guidelines and conventions for AI agents operating across projects.
Unless project-specific instructions dictate otherwise, agents must adhere to the following principles.

---

## 1. Raw String Presentation for Paths and URLs

When communicating file paths, workspace targets, or web URLs to the user, output them as plain, unadorned raw strings without Markdown hyperlinks (e.g., `[label](url)`) or inline code backticks (e.g., `` `path` ``).

- **Rationale**: Ensures seamless, reliable, and error-free copy-pasting directly into terminals, editors, and external tools without capturing unwanted syntax, escape characters, or formatting artifacts.
- **Format Convention**:
  - Preferred: Output `path/to/file.ext` or `https://github.com/example/repo/commit/12345` directly.
  - Avoid: `[file.ext](path/to/file.ext)` or `` `path/to/file.ext` ``.
