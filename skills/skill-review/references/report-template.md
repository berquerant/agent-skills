# Skill Review Report Templates

Report formats for skill review findings. Assign statuses and verdicts using the **Verdict Rules** defined in the skill review checklist.
Per [output-presentation](../../output-presentation/SKILL.md) (../../output-presentation/SKILL.md), output all file paths, line ranges, and URLs by co-presenting accessible Markdown links and unadorned raw strings in parentheses.

---

## Per-Skill Report

```markdown
# Skill Review: `<skill-name>`

**Target Path**: [<skill-name>](<accessible-skill-dir>) (<path-to-skill>) (git: <clean / modified / untracked>)
**Review Basis**: `<checklist version or commit; any interpretations applied>`
**Overall Verdict**: 🟢 **PASS** / 🟡 **NEEDS IMPROVEMENT** / 🔴 **FAIL**

---

## Automated Check Results

| Check | Result |
|---|---|
| SKILL.md lines & Step 1 line | `<n> lines (Step 1 at line <m>)` |
| Orphan / unreferenced files | `<OK / list>` |
| name (pattern / length / dir match) | `<OK / WARN / FAIL ...>` |
| description (length / trigger opener) | `<n> chars (<OK / WARN>)` |
| Section order & `Verify:` checkpoints | `<OK / missing list>` |
| Broken / absolute links | `<none / list>` |
| Non-executable scripts | `<none / n/a / list>` |
| Sibling references | `<skill: count, ...>` |
| Repository registration | `<README: n, AGENTS.md: n, intent-matrix.md: n>` |
| Duplicate reference headings | `<none / list>` |
| External hyperlinks | `<none / annotated human-only citations>` |
| Command & prompt hygiene | `<OK / WARN list>` |

---

## Evaluation Summary

| Dimension | Status | Key Observations |
|---|:---:|---|
| **1. Scope & Granularity** | Pass / Warn / Fail | `<Single responsibility, step count, trigger boundary, risk-proportional gates>` |
| **2. Modularity & File Split** | Pass / Warn / Fail | `<Size, top-loaded Step 1, 3-level progressive disclosure, offline snapshot anchoring>` |
| **3. Maintainability & Conventions** | Pass / Warn / Fail | `<Layer 1 description budget, Verify contracts, positive Why guidance, output/snippet rules>` |
| **4. Loose Coupling** | Pass / Warn / Fail | `<Independent execution, sequential hand-off, read-only vs mutation separation>` |
| **5. DRY & Interoperability** | Pass / Warn / Fail | `<Minimal repetition, shared references, consistent terms, repo & router registration>` |
| **6. Self-Consistency** | Pass / Warn / Fail | `<Description/body match, guidelines honored, API/version coherence, zero orphan files>` |

---

## Detailed Findings

### 🔴 Critical / Blocking Issues (Fail)
- **`<ID>. <Issue Title>`** `[F]`
  - **Location**: [<file>#Lxx-Lyy](<accessible-file-target>) (<file-path>:xx-yy)
  - **Problem**: `<Which [F] rule is violated and why it matters>`
  - **Suggested Remedy**: `<Concrete recommendation or diff>`

### 🟡 Improvements & Warnings (Needs Improvement)
- **`<ID>. <Issue Title>`** `[W]`
  - **Location**: [<file>#Lxx-Lyy](<accessible-file-target>) (<file-path>:xx-yy)
  - **Problem**: `<Which [W] rule is violated and why it matters>`
  - **Suggested Remedy**: `<Concrete recommendation>`

### 🟢 Good Patterns
- `<Well-designed sections, clean progressive disclosure, risk-proportional guardrails, or effective verification contracts>`

---

## Actionable Next Steps

- Priority order: `<IDs, highest first, or "None required">`
- **Recommended Next Skill**:
  - Phased modifications with approval gates: [step-gate](../../step-gate/SKILL.md) (../../step-gate/SKILL.md)
  - Behavior-preserving structural changes: [refactor](../../refactor/SKILL.md) (../../refactor/SKILL.md)
  - Documentation and comment refinements: [doc-refactor](../../doc-refactor/SKILL.md) (../../doc-refactor/SKILL.md)
```

---

## Cross-Skill Summary (Batch Mode)

```markdown
# Skill Review Summary

**Scope**: [<skills-root>](<accessible-skills-root>) (<skills-root>) (<n> skills)
**Review Basis**: `<checklist version or commit; interpretations applied>`

## Verdict Matrix

| Skill | 1 Scope | 2 Modularity | 3 Conventions | 4 Coupling | 5 DRY | 6 Self-Cons. | Overall |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| `<name>` | P/W/F | P/W/F | P/W/F | P/W/F | P/W/F | P/W/F | 🟢/🟡/🔴 |

## Cross-Skill Issues

- **`<ID>. <Issue>`**: `<affected skills>`, with the observed evidence, why it impacts model execution or maintainability, and a concrete remedy.

## Repository-Level Issues

- `<Registration gaps across README.md, AGENTS.md, or intent-matrix.md; Layer 1 token budget totals; diagram mismatches>`

## Prioritized Backlog

| Priority | Item | Skills | Suggested Skill |
|---|---|---|---|
| High | `<item>` | `<skills>` | `step-gate` / `refactor` / `doc-refactor` |
```
