# Conventional Commits Specification Reference (Condensed)

> [!NOTE] Local Reference Snapshot
> This document is a local, self-contained summary of the Conventional Commits specification.
> Originally derived from: `https://www.conventionalcommits.org/` (for human citation only).
> **AI agents must strictly rely on this local specification and MUST NOT fetch or access external URLs autonomously.**

---

## 1. Specification Overview

The Conventional Commits specification provides a lightweight convention on top of commit messages. It creates an explicit commit history that is easy for humans to review and machines to parse.

### Message Structure

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

---

## 2. Structural Elements

### Type
Communicates the intent of the change:
- `feat`: Introduces a new feature to the codebase (correlates with `MINOR` in SemVer).
- `fix`: Patches a bug in the codebase (correlates with `PATCH` in SemVer).
- `docs`: Documentation only changes.
- `refactor`: A code change that neither fixes a bug nor adds a feature.
- `perf`: A code change that improves performance.
- `test`: Adding missing tests or correcting existing tests.
- `build`: Changes that affect the build system or external dependencies.
- `ci`: Changes to CI configuration files and scripts.
- `chore`: Other changes that do not modify `src` or test files.
- `revert`: Reverts a previous commit.

### Scope (Optional)
A noun describing the section of the codebase enclosed in parentheses:
- Example: `feat(parser): add support for arrays`

### Description
A short, imperative summary of the code changes:
- Use imperative mood: "add", "fix", not "added", "fixes".
- Lowercase first letter after the colon.
- No trailing period.

### Body (Optional)
Detailed motivation and context:
- Explain **why** the change is needed and what was changed.
- Wrap lines at ~72 characters.

### Footer (Optional)
Metadata, issue references, and breaking change declarations:
- One or more footers following the format `token: value` or `token #value`.
- Example: `Refs: #123`, `Closes: #456`.

---

## 3. Breaking Changes

A breaking change indicates a breaking API change (correlates with `MAJOR` in SemVer).

A breaking change MUST be signaled in either of two ways:
1. **Exclamation mark in the header**: Append `!` after the type or scope:
   - `feat!: drop support for Node 14`
   - `feat(api)!: change user endpoint response structure`
2. **`BREAKING CHANGE` footer**: Begin a footer with `BREAKING CHANGE: <explanation and migration path>`:
   ```text
   feat: allow provided config object to extend other configs

   BREAKING CHANGE: `extends` key in config file is now used for extending other config files
   ```
