# Conventional Comments Specification Reference (Condensed)

> [!NOTE] Local Reference Snapshot
> This document is a local, self-contained summary of the Conventional Comments specification.
> Originally derived from: `https://conventionalcomments.org/` (for human citation only).
> **AI agents must strictly rely on this local specification and MUST NOT fetch or access external URLs autonomously.**

---

## 1. Specification Overview

Conventional Comments standardizes the format and labels used in code review comments. It makes expectations clear, avoids miscommunication, and reduces friction between reviewers and authors.

### Format

```text
<label> [decoration]: <subject>

[discussion]
```

- `<label>`: Signifies the kind of comment being left (mandatory).
- `[decoration]`: Extra context or constraints, wrapped in parentheses (optional).
- `<subject>`: Main feedback statement in a concise sentence (mandatory).
- `[discussion]`: Supporting arguments, rationale, or code snippets (optional).

---

## 2. Standard Labels

| Label | Description |
|---|---|
| `praise` | Highlights positive aspects, elegant solutions, or great teamwork. |
| `nitpick` | Trivial preferences or minor polish (e.g. typos, stylistic consistency). Must not block merging. |
| `suggestion` | Proposes a concrete improvement or alternative implementation. |
| `issue` | Highlights a problem, bug, broken contract, or regression that should be resolved. |
| `question` | Inquires about an unclear area or asks for context/intent. |
| `thought` | Shares an idea or observation that is not immediately actionable. |
| `chore` | Meta tasks like linking issues, updating docs, or running verification checks. |

---

## 3. Standard Decorations

Decorations provide explicit signaling regarding blocking status:

- `(blocking)`: Must be resolved before the Pull Request / Merge Request is merged.
- `(non-blocking)`: Should not prevent the PR/MR from being merged. The author can decline or address later.
- `(if-minor)`: The suggestion should only be addressed if it is quick and trivial to do so.

*Default convention if decoration is omitted*:
- `issue` defaults to blocking.
- `suggestion`, `thought`, and `nitpick` default to non-blocking.
