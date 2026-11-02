---
name: output-presentation
description: >-
  Use this skill when communicating file paths, workspace targets, web URLs,
  or commands to the user to co-present clickable Markdown links with unadorned
  raw strings for both direct access and seamless copy-pasting.
---

# Output Presentation

Use this skill to format user-facing messages, file paths, workspace targets,
and web URLs cleanly by co-presenting a clickable Markdown link alongside an
unadorned raw string in parentheses. This dual format fulfills two complementary
purposes:
1. **Accessible Markdown Link**: Allows the user to click and immediately access or open the file or web resource in an editor or browser. The link target must always resolve reliably to an accessible resource without breaking due to session directory shifts.
2. **Unadorned Raw String**: Provides the clean, copy-pasteable identifier (e.g., repository-relative path, workspace target, or exact string) without markup or backticks, ready for direct terminal or script use.

---

## Step 1: Identify Target Entities

When constructing responses, logs, or reports, identify any entities intended
for navigation, user copy-pasting, or direct terminal/editor use:

- **File and Directory Paths**: Absolute or relative workspace paths (e.g.,
  `path/to/file.ext`, `/workspace/src/app.ts`).
- **Workspace Targets and Identifiers**: Build targets, package paths, or
  repository references.
- **Web URLs**: Git commit links, PR URLs, documentation links, or external
  API endpoints (e.g., `https://example.com/org/repo/commit/12345`).

Verify: You have pinpointed all path, URL, and workspace targets in the draft
response.

---

## Step 2: Co-Present Clickable Links and Raw Strings

Format each identified entity using the co-presentation syntax:
`[label](accessible_target) (raw_target)`

### 1. Guarantee Link Viability
- The Markdown link portion (`[label](accessible_target)`) must point to a target that the user can actually access and open. Broken links are useless.
- For local workspace files, prefer fully-qualified URIs (e.g., `file:///path/to/file.ext`) or paths guaranteed to resolve from the workspace root. Do not emit fragile relative paths that break if the session's current working directory changes.
- For external web resources, use valid, complete HTTPS URLs.

### 2. Format the Unadorned Raw Target in Parentheses
- Place the raw target immediately after the Markdown link enclosed in plain parentheses: `(raw_target)`.
- Inside the parentheses, output the bare, unadorned string without backticks or nested Markdown tags (e.g., `(path/to/file.ext)`, not `(` `path/to/file.ext` `)`).
- The raw target remains meaningful as a copy-pasteable command argument, workspace-relative reference, or terminal input even when the clickable link uses an absolute URI.

### Examples

| Category | Recommended Co-Presentation | Avoid | Reason |
|---|---|---|---|
| File Path | `[file.ext](<file:///workspace/src/file.ext>) (src/file.ext)` | `[file.ext](src/file.ext)` | Relative link breaks if session cwd shifts |
| File Path | `[file.ext](<file:///workspace/src/file.ext>) (src/file.ext)` | `(` `src/file.ext` `)` or `src/file.ext` alone | Lacks clickable link or includes backticks |
| Directory | `[components/](<file:///workspace/src/components/>) (src/components/)` | `` `src/components/` `` | Inline backticks interfere with copy-paste |
| Web URL | `[PR #1](https://example.com/repo/pull/1) (https://example.com/repo/pull/1)` | `[PR #1](https://example.com/repo/pull/1)` alone | Lacks unadorned raw URL for fast terminal pasting |

Verify: Every path or URL features a clickable, working link followed by an unadorned raw target in parentheses without backticks.

---

## Step 3: Verify Copy-Paste Reliability and Link Viability

Before outputting to the user, conduct a pre-flight sanity check on the
formatted text:

- **Link Viability**: Confirm that clicking the Markdown link resolves to the actual file or URL without a 404 or missing-file error.
- **Copy-Paste Isolation**: Confirm that double-clicking or dragging within `(raw_target)` captures only the clean target string without surrounding parentheses, backticks, or trailing punctuation.
- **Punctuation Separation**: If a sentence ends with a co-presented entity, ensure sentence-ending punctuation (such as periods, commas, or colons) is placed outside the closing parenthesis or separated by whitespace.

Verify: The link is clickable and accessible, and the raw target can be copied cleanly without requiring manual editing.

---

## Guidelines

- **Always co-present link and raw string.** Output the clickable Markdown link first, followed by the unadorned raw target in parentheses: `[label](accessible_target) (raw_target)`.
- **Guarantee link viability.** Never provide broken, dead, or cwd-dependent links in the Markdown link portion. The link must be directly accessible to the user.
- **Keep the parenthetical raw target clean.** Do not enclose the raw target in backticks or Markdown formatting within `(...)`.
- **Complementary roles.** Rely on the link for effortless navigation and the raw string for exact, copy-pasteable CLI/editor input.
- **Preserve exact characters.** Do not escape underscores or special characters within raw paths or URLs with backslashes.
