# Doctor Report Template

The finding format extends the one in [`code-review`](../../code-review/SKILL.md) Step 3 (Where, Rule, Why, How). The severity labels (🔴 Critical, 🟡 Warning, 🟢 Suggestion, 💡 Positive) are the same.

---

## Finding Format

Each finding has these fields:

1. **Axis**: A, B, C, or D, as defined in [diagnostic-axes.md](diagnostic-axes.md).
2. **Severity**: 🔴 / 🟡 / 🟢 / 💡.
3. **Where**: relative `path/to/file.ext#Lxx-Lyy`.
4. **Evidence (fact)**: what was directly observed, such as a snippet or tool output. No interpretation.
5. **Reasoning (interpretation)**: why the evidence is a problem, and the failure mode or threat. See the separation rule in [`evidentiary-inquiry`](../../evidentiary-inquiry/references/source-scoping-and-verification.md).
6. **Rule**: the criterion from the sibling reference that applies.
7. **Fix**: a concrete recommendation.
8. **Confidence**: High (directly observed and reproducible), Medium (observed, impact inferred), Low (indirect).

---

## Severity-to-Priority Mapping

Use this single mapping in the improvement plan:

| Severity | Priority | Meaning |
|---|---|---|
| 🔴 Critical | P0 | Fix now. |
| 🟡 Warning | P1 | Fix soon. |
| 🟢 Suggestion | P2 | Improve when convenient. |

For axis B, use the impact-versus-effort matrix in [`project-audit`'s maintainability metrics](../../project-audit/references/maintainability-metrics.md) to decide the severity of each finding. A finding the matrix marks P3 ("accept debt") is recorded in the report but not added to the plan.

---

## Report Template

```markdown
# 🩺 Doctor Report

**Mode**: <project|recent> × <quick|thorough>
**Scope**: <what was examined, and the rule that selected it>
**Limits**: <files not read, tools not run or unavailable>

## Summary
| Axis | Status | Findings (🔴/🟡/🟢) |
|---|---|---|
| A. Documentation & comments | <Good / Needs work / Poor> | <n/n/n> |
| B. Maintainability | <...> | <n/n/n> |
| C. Latent bugs | <...> | <n/n/n> |
| D. Security | <...> | <n/n/n> |

> ⚠️ <Critical security or public exposure risk, if any. Otherwise omit.>

## Findings
### <Axis>: <short title>
- **Severity**: ...
- **Where**: ...
- **Evidence**: ...
- **Reasoning**: ...
- **Rule**: ...
- **Fix**: ...
- **Confidence**: ...

## Improvement Plan
| # | Priority | Items (findings) | Owner skill | Success criterion | Depends on |
|---|---|---|---|---|---|
| 1 | P0 | <action> (<finding refs>) | <skill> | <testable condition> | — |

**Suggested order**: <sequence and reason>
**Risks / assumptions / unchecked areas**: <list>

## Decision
Which items should I proceed with? Nothing will be changed until you approve.
```
