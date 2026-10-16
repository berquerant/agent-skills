# Diagnostic Axes

`doctor` diagnoses four axes. This file does not restate the checks. It points to the sibling skill that defines them, and adds only the depth rules and the owner-skill routing that belong to `doctor`.

---

## Axis A: Documentation & Comments

Covers missing documentation, and comments that are too few, too many, outdated, or low quality.

- **Criteria**:
  - Documentation Review focus in [`code-review`](../../code-review/SKILL.md) Step 1 (clarity, completeness, accuracy against the code, broken links and examples).
  - The "hard-to-understand code" categories and the Why-first principle in [`doc-refactor`](../../doc-refactor/SKILL.md) Step 1 and Step 2, with the checklist in [`doc-refactor-guide.md`](../../doc-refactor/references/doc-refactor-guide.md).
- **Look for**: missing README or usage docs; undocumented public interfaces; complex logic without a Why comment; comments that only restate the code; comments that contradict the code.
- **`quick`**: public interfaces and entry points of the sampled files.
- **`thorough`**: every public symbol and every complex block in scope.

## Axis B: Maintainability

Covers code that is hard to understand, change, or test.

- **Criteria**: [`maintainability-metrics.md`](../../project-audit/references/maintainability-metrics.md) (size and complexity thresholds, debt markers, duplication and coupling, testability, dead code).
- **`quick`**: the detection commands in that file (large files, debt markers).
- **`thorough`**: all sections, including duplication, coupling, test gaps, and dead code.

## Axis C: Latent Bugs

Covers behavior that is likely wrong, even if no test fails today.

- **Criteria**: General Review focus in [`code-review`](../../code-review/SKILL.md) Step 1 and the static analysis in its Step 2.
- **Look for**: unchecked errors and return values; off-by-one and boundary mistakes; nil/null handling; resource leaks; race conditions; time-of-check/time-of-use gaps; unhandled edge cases; logic that contradicts its comments or docs.
- **`quick`**: linter and type-checker output when cheap, plus pattern searches.
- **`thorough`**: trace the logic of each unit in scope, including callers and failure paths, and run the test suite.

## Axis D: Security

Covers vulnerabilities and exposure risks.

- **Criteria**: [`security-checklist.md`](../../code-review/references/security-checklist.md) and [`owasp-secure-coding-guide.md`](../../code-review/references/owasp-secure-coding-guide.md).
- **`quick`**: secret and credential searches, and the injection and unsafe-input patterns from the checklists.
- **`thorough`**: the full checklists, dependency vulnerability scanning when available, and, for `recent`, the commits in scope as well as the working tree.

---

## Owner-Skill Routing

Use this table in Step 4 to assign an owner skill to each plan item.

| Finding type | Owner skill | Notes |
|---|---|---|
| Documentation and comment gaps or quality (A) | [`doc-refactor`](../../doc-refactor/SKILL.md) | Pure documentation and comments, no code change. Use `comprehensive` mode when new comments are needed. |
| Structural problems and debt (B) | [`refactor`](../../refactor/SKILL.md) | Behavior-preserving. |
| Latent bugs (C) | [`explore-plan-execute`](../../explore-plan-execute/SKILL.md) | Fixes change behavior, which `refactor` never does. Plan the fix and its test first. Any behavior-preserving cleanup afterward goes to `refactor`. |
| Security issues (D) | [`explore-plan-execute`](../../explore-plan-execute/SKILL.md) | Treat critical items as P0. Use `step-gate` when the fix is high-risk or spans several steps. Behavior-preserving cleanup afterward goes to `refactor`. |
| Several dependent items, or high-risk changes | [`step-gate`](../../step-gate/SKILL.md) | One approved step at a time. |
| Fixes that need a design | [`explore-plan-execute`](../../explore-plan-execute/SKILL.md) | Full implementation plan first. |

A finding that spans axes (for example a bug caused by a misleading comment) gets one plan item per owner skill, ordered so that code changes come before the comment updates that describe them.
