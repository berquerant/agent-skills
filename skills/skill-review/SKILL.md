---
name: skill-review
description: >-
  Use this skill when reviewing, auditing, or evaluating one or more agent skills against the
  Agent Skills Specification, repository conventions, and frontier model alignment. Runs
  reproducible automated checks, scores six dimensions (scope, modularity, conventions, coupling,
  DRY, self-consistency), and outputs a read-only report.
---

# Skill Review

Use this skill to audit existing or newly created agent skills. It combines reproducible automated checks with a structured evaluation across six dimensions—including progressive disclosure, Layer 1 trigger economy, risk-proportional guardrails, and internal version consistency—to produce an actionable per-skill report and a cross-skill summary in batch mode.

This skill evaluates skills as skills: their structure, triggers, context efficiency, and interoperability. For general application code or security review, use [code-review](../code-review/SKILL.md) (../code-review/SKILL.md) instead.

This skill is strictly read-only and directly references three reference documents and one helper script:
- [skill-review-checklist.md](references/skill-review-checklist.md) (references/skill-review-checklist.md): evaluation criteria and verdict rules (single source of truth).
- [automated-checks.md](references/automated-checks.md) (references/automated-checks.md): documentation for checks A through K implemented in [run-checks.sh](scripts/run-checks.sh) (scripts/run-checks.sh).
- [report-template.md](references/report-template.md) (references/report-template.md): per-skill and batch report templates.

---

## Prerequisites

- Read access to the target skill directories (each containing `SKILL.md`) and their sibling skills.
- A POSIX-compatible shell with `grep`, `sed`, `awk`, and `find` to run [run-checks.sh](scripts/run-checks.sh) (scripts/run-checks.sh). If a shell is unavailable, perform the equivalent checks in [automated-checks.md](references/automated-checks.md) (references/automated-checks.md) manually.

---

## Step 1: Identify Review Scope & Baseline

1. Determine the review scope (scaling confirmation gates proportionally to scope):
   - **Single mode**: one skill directory, such as `skills/<name>/` or `.agents/skills/<name>/`. Proceed directly to Step 2 without an extra confirmation gate.
   - **Batch mode**: multiple skills or all skills under a skills root. Group targets by role and confirm the target list with the user before full evaluation.
2. List the files in each target skill directory and record the working tree baseline so you can confirm at completion that no files were modified:
   ```sh
   find skills/<target-skill> -type f | sort
   git status --short
   ```

Verify: Target skill directories (each containing `SKILL.md`) and their file inventories are listed, and the baseline `git status --short` output is recorded.

---

## Step 2: Run Automated Checks

Execute [run-checks.sh](scripts/run-checks.sh) (scripts/run-checks.sh) (as documented in [automated-checks.md](references/automated-checks.md) (references/automated-checks.md)) against each target skill and record the raw output for checks A through K:

```sh
skills/skill-review/scripts/run-checks.sh skills/<target-skill>
```

- **A–D**: File inventory, `SKILL.md` line count (80–250 `[W]`, < 500 `[F]`), direct reference coverage (zero orphan files `[W]`), `name` rules (3–30 chars `[W]`), `description` Layer 1 budget (100–350 chars `[W]`, ≤ 1024 `[F]`) and trigger opener, top-loaded `## Step 1` (≤ line 100 `[W]`), section order ending with `## Guidelines`, and `Verify:` checkpoints.
- **E–K**: Relative/absolute links, script permissions, sibling references, registration across `README.md` / `AGENTS.md` / `intent-matrix.md`, duplicate reference headings, external URLs, and command/prompt hygiene (< 120 chars inline shell commands, unadorned raw link targets, no manual CoT `[W]`).

Verify: Automated check results (both `[F]` errors and `[W]` warnings) are recorded for every target, and any false positives are triaged with evidence.

---

## Step 3: Evaluate Structure & Model Alignment (Dimensions 1–3)

Using the automated results and direct file inspection, evaluate each target against dimensions **1. Scope & Granularity**, **2. Modularity & File Split**, and **3. Maintainability & Conventions** in [skill-review-checklist.md](references/skill-review-checklist.md) (references/skill-review-checklist.md):

- Check that triggers define clear activation boundaries, guardrails scale proportionally with operational risk (low friction for read-only exploration, strict gates for state mutation), deep content is offloaded to one-level-deep `references/` with version/date anchors on external snapshots, and guidelines pair affirmative rules with brief "Why" rationale.
- Record every finding with its severity (`[F]` / `[W]`) and exact location formatted per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md).

Verify: Every checklist item in dimensions 1–3 has an explicit Pass, Warn, or Fail backed by line-level evidence.

---

## Step 4: Evaluate Coupling, Interoperability & Self-Consistency (Dimensions 4–6)

Evaluate dimensions **4. Loose Coupling**, **5. DRY & Interoperability**, and **6. Self-Consistency** in [skill-review-checklist.md](references/skill-review-checklist.md) (references/skill-review-checklist.md):

- Compare overlapping sibling skills and `references/` directly to verify shared references, consistent terminology, and complete registration across repository docs and routing tables.
- Inspect all files inside the target skill (`SKILL.md`, `references/`, `examples/`, `scripts/`) to confirm that SDK imports, CLI flags, version assumptions, and workflow terms are mutually consistent and every auxiliary file is directly referenced from `SKILL.md`.

Verify: Every checklist item in dimensions 4–6 has an explicit Pass, Warn, or Fail, and any cross-skill or internal version inconsistencies are logged.

---

## Step 5: Generate Report & Recommend Hand-Off

1. Compute each dimension status and the overall verdict using the **Verdict Rules** in [skill-review-checklist.md](references/skill-review-checklist.md) (references/skill-review-checklist.md).
2. Format the report using [report-template.md](references/report-template.md) (references/report-template.md) (including the cross-skill summary in batch mode), co-presenting all paths and locations per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md) and providing a concrete remedy for every `[F]` or `[W]`.
3. Re-run `git status --short` to confirm the workspace matches the Step 1 baseline.
4. Recommend the appropriate follow-up mutation skill for user approval: [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md) for phased fixes, [refactor](../refactor/SKILL.md) (../refactor/SKILL.md) for structural changes, or [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md) for documentation refinements.

Verify: The report is complete with concrete remedies, `git status --short` matches the Step 1 baseline, and next-step hand-off options are presented.

---

## Guidelines

- **Preserve workspace integrity through read-only auditing**: Keep the review strictly non-mutating so users can inspect findings safely before approving changes.
- **Anchor every finding in verifiable evidence**: Cite exact check outputs or dual-formatted locations (`[file#Lxx-Lyy](target) (path:xx-yy)`) so every judgment is reproducible.
- **Apply a uniform, self-referential standard**: Evaluate every skill—including `skill-review` itself—against the exact criteria in [skill-review-checklist.md](references/skill-review-checklist.md) (references/skill-review-checklist.md), documenting any context-specific interpretations in **Review Basis**.
- **Pair every warning or failure with an actionable remedy**: Provide concrete replacement text, structural splits, or command fixes so downstream remediation skills can execute immediately.
- **Hand off remediation sequentially**: Recommend [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md), [refactor](../refactor/SKILL.md) (../refactor/SKILL.md), or [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md) for user approval so audit and mutation remain cleanly decoupled.
