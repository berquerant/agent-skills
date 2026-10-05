# Skill Review Report Templates

Report formats for skill review findings. Assign statuses and verdicts using the **Verdict Rules** in [skill-review-checklist.md](skill-review-checklist.md).

---

## Per-Skill Report

```markdown
# Skill Review: `<skill-name>`

**Target Path**: `<path-to-skill>` (git: `<clean / modified / untracked>`)
**Review Basis**: `<checklist version or commit; any interpretations applied>`
**Overall Verdict**: 🟢 **PASS** / 🟡 **NEEDS IMPROVEMENT** / 🔴 **FAIL**

---

## Automated Check Results

| Check | Result |
|---|---|
| SKILL.md lines | `<n>` |
| name (pattern / length / dir match) | `<OK / FAIL ...>` |
| description length | `<n> chars` |
| Steps missing `Verify:` | `<none / list>` |
| Broken / absolute links | `<none / list>` |
| Non-executable scripts | `<none / n/a / list>` |
| Sibling references | `<skill: count, ...>` |
| Repository registration | `<README: n, AGENTS.md: n>` |

---

## Evaluation Summary

| Dimension | Status | Key Observations |
|---|:---:|---|
| **1. Scope & Granularity** | Pass / Warn / Fail | `<Brief observation>` |
| **2. Modularity & File Split** | Pass / Warn / Fail | `<Brief observation>` |
| **3. Maintainability & Conventions** | Pass / Warn / Fail | `<Brief observation>` |
| **4. Loose Coupling** | Pass / Warn / Fail | `<Brief observation>` |
| **5. DRY & Interoperability** | Pass / Warn / Fail | `<Brief observation>` |
| **6. Self-Consistency** | Pass / Warn / Fail | `<Brief observation>` |

---

## Detailed Findings

### 🔴 Critical / Blocking Issues (Fail)
- **`<ID>. <Issue Title>`** `[F]`
  - **Location**: `<file-path>#Lxx-Lyy`
  - **Problem**: `<Which [F] rule is violated and why it matters>`
  - **Suggested Remedy**: `<Concrete recommendation or diff>`

### 🟡 Improvements & Warnings (Needs Improvement)
- **`<ID>. <Issue Title>`** `[W]`
  - **Location**: `<file-path>#Lxx-Lyy`
  - **Problem**: `<Which [W] rule is violated and why it matters>`
  - **Suggested Remedy**: `<Concrete recommendation>`

### 🟢 Good Patterns
- `<Well-designed sections, clean modularity, or effective verification checkpoints>`

---

## Actionable Next Steps

- Priority order: `<IDs, highest first>`
- **Recommended Next Skill**:
  - Phased modifications with approval gates: `step-gate`
  - Behavior-preserving structural changes: `refactor`
```

---

## Cross-Skill Summary (Batch Mode)

```markdown
# Skill Review Summary

**Scope**: `<skills root>` (`<n>` skills)
**Review Basis**: `<checklist version or commit; interpretations applied>`

## Verdict Matrix

| Skill | 1 Scope | 2 Modularity | 3 Conventions | 4 Coupling | 5 DRY | 6 Self-Cons. | Overall |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| `<name>` | P/W/F | P/W/F | P/W/F | P/W/F | P/W/F | P/W/F | 🟢/🟡/🔴 |

## Cross-Skill Issues

- **`<ID>. <Issue>`**: `<affected skills>`, with the problem and a suggested remedy (for example, which skill should own a shared reference)

## Repository-Level Issues

- `<Registration gaps in README / AGENTS.md, diagram mismatches, untracked skills>`

## Prioritized Backlog

| Priority | Item | Skills | Suggested Skill |
|---|---|---|---|
| High | `<item>` | `<skills>` | `step-gate` / `refactor` |
```
