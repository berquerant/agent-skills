# Technical Documentation Review Report Template

Use these templates to structure reports produced during `doc-review`.

---

## 1. Finding Item Schema

Every finding must include four core components:

```markdown
### [Severity Badge] Finding Summary

- **Where**: `path/to/document.md#Lxx-Lyy`
- **Rule**: [Dimension Name & Checklist Item from doc-review-checklist.md]
- **Why**: Explanation of the technical divergence, missing context, readability barrier, or security risk.
- **How**:
  ```markdown
  // Concrete drop-in replacement or structural revision
  ```
```

---

## 2. Single Document Review Report Template

```markdown
# Documentation Review: <Document Name or Path>

> **Scope**: `<path/to/target.md>`  
> **Mode**: `quick` | `thorough`  
> **Baseline State**: Clean (unmodified)

---

<!-- If public exposure or critical execution hazards exist, insert this alert banner: -->
> [!CRITICAL] Public Exposure or Severe Correctness Risk Detected
> <Summary of leaked credentials, internal infrastructure identifiers, or broken critical commands.>

---

## Executive Summary

- **Overall Quality Verdict**: Excellent | Satisfactory | Needs Improvement | Action Required
- **Total Findings**: X (🔴 Critical: A | 🟡 Warnings: B | 🟢 Suggestions: C)

---

## Detailed Findings

### 🔴 Critical Issues
<!-- List Critical findings using the schema above, or note "None identified." -->

### 🟡 Warnings & Drift
<!-- List Warning findings using the schema above, or note "None identified." -->

### 🟢 Suggestions & Clarity
<!-- List Suggestion findings using the schema above, or note "None identified." -->

### 💡 Positive Observations
<!-- Highlight well-written explanations, accurate edge-case caveats, or clear structures. -->

---

## Recommended Next Steps

- [ ] Apply documentation improvements via [`doc-refactor`](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md).
- [ ] Post review comments or engage in discussion via [`review-discussion`](../review-discussion/SKILL.md) (../review-discussion/SKILL.md).
- [ ] Craft commit or PR updates via [`change-message`](../change-message/SKILL.md) (../change-message/SKILL.md).
```

---

## 3. PR Diff & Batch Review Report Template

```markdown
# Documentation Review Summary: <PR / Topic Branch / Multi-Document Scope>

> **Base Comparison**: `<base-branch>...<topic-branch>`  
> **Files Inspected**: N files  
> **Mode**: `quick` | `thorough`

---

## Summary Matrix

| Document Path | Status | 🔴 Critical | 🟡 Warnings | 🟢 Suggestions |
|---|---|---|---|---|
| `docs/architecture.md` | Needs Improvement | 1 | 2 | 0 |
| `README.md` | Pass | 0 | 0 | 1 |

---

## Detailed Findings by File

<!-- Repeat per-file findings using the Finding Item Schema -->

---

## Recommended Next Steps

- [ ] Batch-refactor affected documents via [`doc-refactor`](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md).
- [ ] Formulate thread responses via [`review-discussion`](../review-discussion/SKILL.md) (../review-discussion/SKILL.md).
```
