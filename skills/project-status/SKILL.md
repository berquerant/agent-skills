---
name: project-status
description: >-
  Use this skill to reliably audit the current state of a project. Activates
  when the user requests a status check or when applied changes produce a diff
  that diverges from expectations. Systematically inspects git state, file
  changes, recent commits, and test/lint results, then reports findings with
  clear pass/fail judgments.
---

# Project Status

Use this skill to get a reliable, structured picture of the project's current
state — either on user request or whenever a change produces a diff that
differs from what was expected.

---

## Triggers

Activate this skill in either of the following situations:

1. **User request** — the user asks to check the current state, review what
   has changed, or verify the diff (e.g. "what's the status?", "confirm the
   changes", "something looks off").
2. **Drift detected** — after applying changes, the resulting diff contains
   unexpected modifications, missing edits, or unintended side-effects that
   were not part of the plan.

---

## Step 1: Define the Inspection Scope

Before running any commands, state clearly what will be checked and why.

- **User request**: list the areas the user cares about (git state, specific
  files, test results, etc.).
- **Drift detected**: write down the **expected diff** in plain language first
  (e.g. "only `foo.go` should have changed, adding function `Bar`"). This
  anchors the comparison in Step 3.

Verify: The scope is written out and the expected state (if drift-triggered)
is explicitly stated before any commands are run.

---

## Step 2: Check Git State

Run the following commands and record their output:

```sh
# Overall working-tree status
git status

# Unstaged changes (content)
git diff

# Staged changes (content)
git diff --staged

# Recent commits for context
git log --oneline -10
```

Note:
- Which files are modified, staged, or untracked.
- Whether any files appear in an unexpected state (e.g. accidentally staged,
  deleted, or modified).

Verify: You have a complete list of all changed files and their current stage
(unstaged / staged / untracked).

---

## Step 3: Validate Changes Against Expectations

Compare the actual state from Step 2 against the expected state from Step 1.

For each changed file, answer:

| Question | Answer |
|----------|--------|
| Was this file expected to change? | ✅ Yes / ❌ No (unexpected) |
| Is the content of the change correct? | ✅ Yes / ⚠️ Partial / ❌ No |
| Are any expected changes missing? | ✅ None missing / ❌ Missing: `<file>` |

If **user-requested** (not drift-triggered), skip the comparison table and
simply describe what has changed.

Verify: Every changed file is accounted for as either expected or unexpected.

---

## Step 4: Run Tests and Static Analysis (if applicable)

If the project has a test suite or linter, run them to confirm nothing is
broken:

```sh
# Examples — adapt to the project's toolchain:
# go test ./...
# npm test
# pytest
# cargo test
# make lint
```

If no toolchain is detected or tests are not relevant to the check, skip this
step and note the reason.

Verify: Test and lint results are recorded (pass / fail / skipped with reason).

---

## Step 5: Report Findings

Produce a structured status report using the following format:

```
## Project Status Report

### Git State
- Branch: <branch>
- Modified files: <list or "none">
- Staged files: <list or "none">
- Untracked files: <list or "none">
- Recent commits: <last 3 one-liners>

### Change Validation        [✅ / ⚠️ / ❌]
<comparison table from Step 3, or change summary if user-requested>

### Tests / Lint             [✅ Pass / ❌ Fail / — Skipped]
<result summary or skip reason>

### Overall Assessment       [✅ Clean / ⚠️ Needs attention / ❌ Action required]
<one-paragraph summary of findings>

### Recommended Actions (if any)
- <actionable item>
```

Use:
- ✅ when the item matches expectations or passes.
- ⚠️ when something is off but not blocking.
- ❌ when action is required.
- — when the check was intentionally skipped.

Verify: Every scope item from Step 1 has a corresponding entry in the report.

---

## Guidelines

- **State expectations before checking.** For drift-triggered runs, always
  write out the expected diff in Step 1 before running any commands. Comparing
  to an unstated expectation leads to missed discrepancies.
- **Cover all changed files.** Do not focus only on the files you edited;
  check `git status` output completely for surprises.
- **Never assume "good enough".** If something looks unexpected, flag it with
  ⚠️ or ❌ rather than silently accepting it.
- **Skip gracefully.** If tests or lint are not applicable, say so explicitly
  rather than omitting the section.
- **Report before acting.** This skill produces a report only. If fixes are
  needed, propose them separately and get user approval first.
