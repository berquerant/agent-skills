# Comment Conventions Reference

This reference defines conventions for drafting clear, constructive, and actionable review comments and notes on Pull Requests (PRs) and Merge Requests (MRs).

Default standard: [Conventional Comments](conventional-comments-spec.md). Follow repository-specific conventions (e.g. `CONTRIBUTING.md`) if established.

---

## Structure of a Review Comment

A well-formed comment consists of three parts:

```markdown
**<label> [(decorations)]**: <subject>

<discussion / reasoning>

<actionable suggestion or concrete code example>
```

### 1. Label

The label clearly communicates the intent and nature of the comment:

| Label | Meaning | Typical Usage |
|---|---|---|
| `suggestion` | Proposed alternative or improvement | Optimizations, cleaner patterns, better naming |
| `issue` | A defect, bug, or problem that must or should be addressed | Logic bugs, regression risks, broken contracts |
| `question` | Request for clarification or context | Understanding design intent, edge case handling |
| `thought` | Exploratory idea or consideration, not immediately actionable | Future improvements, architecture musings |
| `nitpick` | Minor trivial polish (style, typos) | Docstrings, variable naming, formatting tweaks |
| `praise` | Explicit appreciation of great work | Elegant solutions, thorough test cases, clean docs |
| `chore` | Process or meta task | Ticket linking, dependency bump check |

### 2. Decorations (Optional)

Decorations explicitly convey urgency and blocking status:

- `(blocking)`: Must be addressed or formally resolved before merging.
- `(non-blocking)`: Nice to have; author may choose to address now, defer, or decline.
- `(if-minor)`: Address only if trivial; skip if time-consuming.

*Default behavior if omitted*: `issue` defaults to blocking; `suggestion`, `thought`, and `nitpick` default to non-blocking.

### 3. Subject & Discussion

- **Subject**: A clear, concise statement summarizing the observation in one line.
- **Discussion**:
  - Focus on the **why** (performance implications, maintenance cost, safety risk).
  - Separate the author from the code (use "this approach", "this function" instead of "you").
  - Be specific and constructive.

### 4. Concrete Suggestion

Whenever identifying an issue or suggestion, provide a concrete alternative, patch, or pseudo-code block if practical:

```markdown
```suggestion
def process_items(items: list[str]) -> list[str]:
    return [item.strip() for item in items if item]
```
```

---

## Tone and Etiquette Principles

1. **Critique the code, not the author**: Keep discussions technical, objective, and empathetic.
2. **Explain the rationale**: Never say "change this" without explaining the reasoning or risks.
3. **Acknowledge constraints**: Recognize trade-offs (e.g. deadlines, backward compatibility).
4. **Celebrate good solutions**: Use `praise` to highlight elegant designs or robust test suites.

---

## Examples

### Good Issue Comment (Blocking)
```markdown
**issue (blocking)**: Unhandled potential null pointer when `user.profile` is missing

If an existing user without a profile accesses this route, `user.profile.id` raises an `AttributeError`.

Can we use optional chaining or a safe fallback?
```suggestion
user_id = user.profile.id if user.profile else None
```
```

### Good Suggestion Comment (Non-blocking)
```markdown
**suggestion (non-blocking)**: Consider using `collections.defaultdict` to simplify grouping

Using `defaultdict(list)` would avoid the explicit `if key not in mapping` check.
```

### Good Question Comment
```markdown
**question**: Is the ordering of tasks guaranteed across different workers here?

I noticed workers fetch concurrently from the queue. If task ordering matters for downstream processing, we might need a sequence key.
```
