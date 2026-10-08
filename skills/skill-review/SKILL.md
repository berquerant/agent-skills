---
name: skill-review
description: >-
  Use this skill when reviewing, auditing, or evaluating one or more agent skills against the
  Agent Skills Specification and repository conventions. Runs reproducible automated checks,
  rates six dimensions (scope, modularity, conventions, coupling, DRY, self-consistency), and
  produces a structured read-only report without modifying files.
---

# Skill Review

Use this skill to audit existing or newly created agent skills. It combines reproducible automated checks with a judgment-based evaluation and produces an actionable report for each skill, plus a cross-skill summary when several skills are reviewed together.

This skill evaluates skills as skills: their structure, triggers, and how they work together. For general code or prose quality and security review, use `code-review` instead.

This skill is strictly read-only. It uses three references:
- [references/skill-review-checklist.md](references/skill-review-checklist.md): the evaluation criteria and verdict rules. This is the single source of truth for what is checked.
- [references/automated-checks.md](references/automated-checks.md): portable shell commands for the mechanical checks.
- [references/report-template.md](references/report-template.md): the report formats.

---

## Prerequisites

- Read access to the target skill directories (each containing `SKILL.md`) and their sibling skills.
- A POSIX-compatible shell with `grep`, `sed`, `awk`, and `find`, used for the automated checks. If a shell is unavailable, perform the same checks manually.

---

## Step 1: Identify Review Scope & Target Files

1. Decide what to review:
   - **Single mode**: one skill directory, such as `skills/<name>/`, `.agents/skills/<name>/`, or any path the user gives.
   - **Batch mode**: several skills, or every skill under a skills root. List the targets and confirm the order with the user. Grouping skills with similar roles makes cross-skill duplication easier to spot.
2. For each target, list every file and note any pending changes:
   ```bash
   find <skill-dir> -type f | sort
   git status --short <skill-dir>
   ```
3. Record a baseline with `git status --short` so you can show at the end that the review changed nothing.

Verify: Every target has a `SKILL.md`, all of its files are listed, and the baseline git state is recorded.

---

## Step 2: Run Automated Checks

Run the automated checks via `scripts/run-checks.sh <target-skill>` (or follow [references/automated-checks.md](references/automated-checks.md)) for each target and record the raw results:

- File inventory and `SKILL.md` line count
- Frontmatter `name` rules (pattern, length, match with the directory name) and `description` length
- Section headings, plus `## Step` sections that lack a `Verify:` checkpoint
- Broken or absolute links, and scripts that are not executable
- References to sibling skills, and registration in repository-level docs
- External hyperlinks and self-containment status

Verify: Every automated check has a recorded result. False positives, such as links inside code examples, have been triaged and annotated.

---

## Step 3: Evaluate Structure (Dimensions 1–3)

Using the automated results as evidence, evaluate the target against dimensions **1. Scope & Granularity**, **2. Modularity & File Split**, and **3. Maintainability & Conventions** in [references/skill-review-checklist.md](references/skill-review-checklist.md).

Record each finding with a severity (Fail / Warn) and a `file#Lxx-Lyy` location.

Verify: Each item in dimensions 1–3 has a finding or an explicit pass, and every finding has a location.

---

## Step 4: Evaluate Relationships & Coherence (Dimensions 4–6)

Evaluate dimensions **4. Loose Coupling**, **5. DRY & Interoperability**, and **6. Self-Consistency** in [references/skill-review-checklist.md](references/skill-review-checklist.md).

- For dimension 5, read the sibling skills and references that the target overlaps with, and compare their content directly. Do not rely on file names alone.
- In batch mode, keep a running list of cross-skill issues, such as duplicated checklists, inconsistent terms, or mismatched hand-off relationships, for the summary in Step 5.

Verify: Each item in dimensions 4–6 has a finding or an explicit pass, and cross-skill issues are logged.

---

## Step 5: Generate & Present the Review Report

1. Rate each dimension, then determine the overall verdict using the **Verdict Rules** in [references/skill-review-checklist.md](references/skill-review-checklist.md).
2. Fill in the per-skill template in [references/report-template.md](references/report-template.md). Every Warn or Fail must have a concrete remedy.
3. In batch mode, also produce the cross-skill summary (verdict matrix, cross-skill issues, prioritized backlog).
4. Present the reports in the conversation. If the user wants them saved, write them only to a location the user chooses outside the reviewed skills, such as the agent's scratch or artifact area.
5. Recommend a next skill for the user to approve, such as `step-gate` for phased fixes or `refactor` for structural changes.

Verify: Every report is complete, and `git status --short` matches the baseline from Step 1.

---

## Guidelines

- **Strictly read-only**: Never edit, format, or scaffold files in the reviewed skills or the repository. Report findings, then hand remediation off.
- **Evidence over impression**: Back every finding with an automated check result or a specific `file#Lxx-Lyy` location.
- **Uniform standard**: Apply the same checklist and verdict rules to every skill, including this one. Record any interpretation you make in the report.
- **Actionable recommendations**: For every Warn or Fail, give a concrete remedy or example.
- **Sequential hand-off**: Recommend mutation skills (`step-gate`, `refactor`) for the user to approve instead of applying fixes yourself.
