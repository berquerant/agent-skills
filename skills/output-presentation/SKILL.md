---
name: output-presentation
description: >-
  Use this skill when communicating file paths, workspace targets, web URLs,
  or commands to the user to format them as plain raw strings without Markdown
  decorations for seamless copy-pasting.
---

# Output Presentation

Use this skill to format user-facing messages, file paths, workspace targets,
and web URLs cleanly without Markdown hyperlinks or inline code backticks.
Unadorned raw string outputs ensure error-free, direct copy-pasting into
terminals, editors, and external tools without capturing unwanted markup or
escape characters.

---

## Step 1: Identify Target Entities

When constructing responses, logs, or reports, identify any entities intended
for user copy-pasting or direct terminal/editor use:

- **File and Directory Paths**: Absolute or relative workspace paths (e.g.,
  `path/to/file.ext`, `/workspace/src/app.ts`).
- **Workspace Targets and Identifiers**: Build targets, package paths, or
  repository references.
- **Web URLs**: Git commit links, PR URLs, documentation links, or external
  API endpoints (e.g., `https://example.com/org/repo/commit/12345`).

Verify: You have pinpointed all path, URL, and workspace targets in the draft
response.

---

## Step 2: Strip Markdown Links and Backtick Decorations

Format the identified entities as plain, unadorned raw strings:

1. **No Markdown Hyperlinks**:
   - Do **not** use Markdown link syntax like `[label](path)` or `[label](url)`.
   - Output the bare string directly.
2. **No Inline Code Backticks for Paths and URLs**:
   - Do **not** enclose paths or URLs in single or triple backticks (e.g.,
     avoid `` `path/to/file.ext` `` or `` `https://...` ``).
   - Backticks introduce unwanted quotation marks or delimiters when copied
     from rich-text or terminal interfaces.

### Examples

| Category | Preferred (Raw String) | Avoid (Decorated) |
|---|---|---|
| File Path | `path/to/file.ext` | `[file.ext](path/to/file.ext)` or `` `path/to/file.ext` `` |
| Directory | `src/components/` | `` `src/components/` `` |
| Web URL | `https://example.com/repo/pull/1` | `[PR #1](https://example.com/repo/pull/1)` or `` `https://...` `` |

Verify: All paths and URLs in the output are unadorned and free of Markdown links
or enclosing backticks.

---

## Step 3: Verify Copy-Paste Reliability

Before outputting to the user, conduct a pre-flight sanity check on the
formatted text:

- Confirm that double-clicking or dragging to select the path or URL captures
  **only** the target string without brackets, parentheses, or trailing punctuation.
- If a sentence ends with a path or URL, ensure trailing punctuation (such as
  periods, colons, or commas) is separated by whitespace or placed outside the
  selectable target string.

Verify: The text can be copied cleanly from the chat interface and pasted
directly into a terminal or editor without requiring manual editing.

---

## Guidelines

- **Raw strings by default.** Always output file paths, workspace targets, and
  web URLs as unadorned plain text.
- **No Markdown hyperlinks.** Never use `[label](url)` or `[label](path)` for
  paths or URLs communicated to the user.
- **No inline backticks on targets.** Avoid enclosing copy-paste targets in backticks.
- **Preserve exact characters.** Do not escape underscores or special characters
  within paths or URLs with backslashes.
