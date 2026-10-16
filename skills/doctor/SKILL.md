---
name: doctor
description: >-
  Use this skill when the user wants a one-shot checkup of a project (or of its recent
  changes) that covers missing documentation and weak comments, hard-to-maintain code,
  latent bugs, and security risks together. Lets the user pick the target (whole project
  or recent changes) and the depth (quick or thorough), diagnoses with evidence, and
  produces a prioritized improvement plan that hands off to refactor, doc-refactor,
  step-gate, or explore-plan-execute after approval. Prefer project-audit for repository
  hygiene and release readiness, and code-review for reviewing specific files or a
  given diff.
---

# Doctor

Use this skill to run a project checkup that follows the explore → plan → execute pattern.
It diagnoses first, turns the evidence into a prioritized improvement plan, and stops for
user approval before anything is changed.

This skill is **read-only**. It covers the Explore and Plan phases and the approval gate.
The Execute phase belongs to the mutation skills that the user approves in Step 5.

It reuses the criteria of sibling skills instead of restating them:
- [references/modes.md](references/modes.md): the two mode axes, their defaults, and the scope rules.
- [references/diagnostic-axes.md](references/diagnostic-axes.md): the four diagnostic axes, each linked to the sibling criteria that define its checks.
- [references/report-template.md](references/report-template.md): the finding format and the report and plan template.

---

## Prerequisites

- Read access to the target project. A git repository is needed for the `recent` target; for `project` in a non-git directory, the baseline uses a file listing instead.
- The project's own lint, test, and vulnerability tools are optional and used only in `thorough` depth. Without them, perform the same checks by reading the code.

---

## Step 1: Select Mode & Fix the Scope

1. Pick one value per axis, as defined in [references/modes.md](references/modes.md):
   - **Target**: `project` (whole project) or `recent` (recent changes).
   - **Depth**: `quick` (fast and rough) or `thorough` (careful and exhaustive).
2. If the user gave no preference, propose `recent` × `quick` and ask for confirmation.
3. List the files in scope, following the scope rules in [references/modes.md](references/modes.md).
4. Record a baseline: `git status --short` in a git repository, or a file listing with modification times otherwise.

Verify: The target, the depth, the file list in scope, and the baseline state are all stated.

---

## Step 2: Explore (Diagnose)

Examine the scope along the four axes in [references/diagnostic-axes.md](references/diagnostic-axes.md), using the depth chosen in Step 1:

- **A. Documentation & comments**
- **B. Maintainability**
- **C. Latent bugs**
- **D. Security**

Each axis points to the sibling criteria it applies. Open those criteria when you reach the axis; do not recall them from memory. Do not modify any file.

Verify: Every axis has either findings or an explicit "no issue found" with the evidence checked, and the depth's required checks were run or marked skipped with a reason.

---

## Step 3: Verify Evidence

Apply the fact-verification checklist and the fact/interpretation separation from [`evidentiary-inquiry`](../evidentiary-inquiry/references/source-scoping-and-verification.md) to every finding:

- Keep only findings backed by something you directly observed (a `file#Lxx-Lyy` location, a tool output).
- Write the observed evidence and your reasoning separately.
- Drop a finding, or label it low confidence, when the evidence is indirect.

Verify: Every remaining finding has a location, an observed evidence line, a reasoning line, and a confidence level.

---

## Step 4: Plan (Automatic)

Proceed automatically from the verified findings to an improvement plan:

1. Rank the findings as P0 (fix now), P1, or P2 (improve when convenient), using the severity-to-priority mapping in [references/report-template.md](references/report-template.md), which also explains how to use the `project-audit` maintainability matrix for axis B.
2. For each item, name the skill that should carry it out, using the routing table in [references/diagnostic-axes.md](references/diagnostic-axes.md), and give a testable success criterion.
3. Group items that depend on each other, and propose an order.
4. Note risks, assumptions, and anything the diagnosis could not check.

Verify: Every plan item has a priority, an owner skill, a success criterion, and a link to the findings that justify it.

---

## Step 5: Present the Report & Approval Gate

1. Present the report using [references/report-template.md](references/report-template.md). Put any critical security or public exposure risk at the top.
2. Ask which plan items the user approves for execution, or whether they want to adjust the plan:
   > "Which items should I proceed with? I will not change anything until you approve."
3. Confirm that the project state still matches the baseline from Step 1 (`git status --short`, or the file listing with modification times for a non-git project).

Verify: The report is shown, the unchanged git state is confirmed, and the user's decision on each item is recorded.

---

## Step 6: Hand Off for Execution

Do not edit files in this skill. For the approved items, recommend the owner skills from the plan so that the user can start them:

- [`doc-refactor`](../doc-refactor/SKILL.md) for axis A.
- [`refactor`](../refactor/SKILL.md) for axis B, and for behavior-preserving cleanup after a fix. It never changes external behavior.
- [`explore-plan-execute`](../explore-plan-execute/SKILL.md) for axis C and D fixes that need a design or change behavior.
- [`step-gate`](../step-gate/SKILL.md) for multi-step or high-risk fixes, or when several items must run in order.

Verify: Each approved item has a named next skill, and the order is stated when more than one skill is involved.

---

## Guidelines

- **Read-only.** Diagnose and plan only; never modify, format, or create files in the target project.
- **Approval before execution.** Plan automatically, but never move to execution without explicit user approval.
- **Evidence over impression.** Report only what you observed, and keep facts apart from interpretations.
- **Reuse, don't restate.** Apply sibling criteria by reference. Add only the mode rules, the axis routing, and the report format that belong to `doctor`.
- **Sequential hand-off.** Recommend mutation skills for the user to approve; do not run them as nested steps.
- **Honest depth.** State what `quick` did not cover, so a clean quick report is not read as a clean project.
