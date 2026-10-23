---
name: snippet-review
description: >-
  Use this skill when auditing ad-hoc or on-the-fly execution snippets for
  intelligibility, cognitive complexity, and safety. Evaluates snippet length,
  functions, interface complexity, branching, async constructs, and shell
  idiosyncrasies across a four-tier scale, driving simplification or guided
  user consultation before execution.
---

# Snippet Review

Audit ad-hoc, temporary, or task-generated execution snippets (such as shell scripts or interpreter one-offs) for on-the-fly cognitive intelligibility and safety before execution.

This skill is a specialized branch of code review, specifically tailored for code generated on the fly during agent workflows. Unlike standard codebase reviews that focus on long-term maintainability, this skill ensures that human users and reviewers can immediately understand the snippet's intent, control flow, and side-effects without cognitive friction.

---

## Prerequisites

- An execution snippet has been drafted or saved to a file (e.g. via [`command-snippet`](../command-snippet/SKILL.md)).
- The snippet code is accessible for read-only inspection.
- The evaluation criteria in [references/intelligibility-rubric.md](references/intelligibility-rubric.md) are available.

---

## Step 1: Inventory Snippet Structure & Metrics

Inspect the snippet against the five evaluation dimensions in [references/intelligibility-rubric.md](references/intelligibility-rubric.md):

1. **Code Scale & Function Count**:
   - Count the non-empty, non-comment lines of code (SLOC).
   - Count the number of functions/subroutines defined and check if any function exceeds 15 lines.
2. **Interface & Arguments**:
   - Inspect positional arguments (`$1`, `$2`), flags, or environment variable reads.
   - Verify all inputs are explicitly parameterized.
3. **Control Flow & Branching**:
   - Count decision points (`if`, `elif`, `case`, loops, `&&`, `||`).
   - Measure maximum nesting depth (flag if ≥ 3 levels).
4. **Asynchrony & Concurrency**:
   - Detect background processes (`&`), subshell parallelism, async constructs, or custom signal traps.
5. **Language & Shell Idiosyncrasies**:
   - Check for deep pipelines (≥3 pipes), `eval`, process substitution (`<()`), or dense, uncommented regex patterns (`sed`, `awk`, `perl`).

Verify: All five structural metrics are measured and recorded for the candidate snippet.

---

## Step 2: Rate Intelligibility & Complexity Tier

Score the snippet using the Four-Tier Rubric in [references/intelligibility-rubric.md](references/intelligibility-rubric.md):

- **Level 1 (Transparent)**: Simple, linear command flow (≤ 10 SLOC, no custom functions, ≤ 1 pipe, trivial branching).
- **Level 2 (Moderate)**: Clear logic with minor branching or a simple loop (11–20 SLOC, explicit comments, cleanly documented inputs).
- **Level 3 (Complex)**: Multiple functions, nested branches, dense regex, subshells, or ≥ 21 SLOC. Hard to verify at a glance.
- **Level 4 (Opaque / Critical)**: Deep pipelines, `eval`, concurrency/async background tasks, nesting depth ≥ 3, or high-risk destructive commands.

Verify: The snippet is assigned an explicit rating (Level 1, Level 2, Level 3, or Level 4) with documented evidence.

---

## Step 3: Handle Outcome & Drive Remediation

Determine the workflow path based on the tier assigned in Step 2:

### Path A: Low Complexity (Level 1 or Level 2)
The snippet is sufficiently intelligible for execution.
1. Ensure all non-obvious blocks contain concise "Why" comments.
2. Formulate a **Pre-Execution Explanation** containing:
   - **Purpose**: What the script accomplishes.
   - **Arguments**: What values are passed and their roles.
   - **Side-Effects**: What files, directories, or processes are affected.
3. Proceed directly to presenting the explanation to the user before running.

### Path B: High Complexity (Level 3 or Level 4)
The snippet requires intervention to prevent black-box execution:
1. **Attempt Simplification**:
   - Apply simplification strategies from [references/intelligibility-rubric.md](references/intelligibility-rubric.md) (inlining single-use helpers, breaking dense pipes into temporary variables, adding explanatory comments).
   - Re-evaluate the modified snippet. If the score drops to Level 1 or Level 2, transition to Path A.
2. **Open User Consultation Gate (If Simplification Fails)**:
   - If the task's inherent nature prevents reducing complexity to Level 1 or 2, do NOT execute silently.
   - Present the **Snippet Intelligibility Consultation** using the template in [references/intelligibility-rubric.md](references/intelligibility-rubric.md):
     - Explain why the snippet is complex.
     - Detail the intended logic, arguments, and potential side-effects.
     - Present the three user choices:
       1. **Split into Smaller Steps (Recommended)**: Divide the snippet into smaller, step-by-step executions.
       2. **Proceed as-is**: Execute after explicit user acknowledgment.
       3. **Alternative Approach**: Discard the snippet in favor of standard commands or manual action.
   - Await user instruction before proceeding.

Verify: The snippet is either confirmed intelligible (Level 1/2) with pre-execution explanations prepared, or successfully simplified, or gated under explicit user consultation.

---

## Step 4: Present Audit Findings & Hand Off

1. Present the audit summary in the conversation:
   ```markdown
   ### Snippet Intelligibility Audit: `<snippet-path-or-name>`
   - **Complexity Tier**: Level 1 (Transparent) / Level 2 (Moderate) / Level 3 (Complex) / Level 4 (Opaque)
   - **Key Metrics**: <SLOC>, <functions count>, <max nesting depth>, <branching points>
   - **Intelligibility Status**: Approved for execution / Simplified / Awaiting consultation
   ```
2. If user consultation was required, present the consultation options and wait for direction.
3. If approved for execution, proceed back to [`command-snippet`](../command-snippet/SKILL.md) to explain the intent to the user and execute.

Verify: The intelligibility evaluation report is communicated clearly to the user with no unverified assumptions.

---

## Guidelines

- **Zero Black-Box Execution**: Never execute ad-hoc snippets without auditing how understandable they are to human reviewers.
- **Universal Application**: Apply this review to all generated execution snippets, regardless of line count.
- **Prioritize Simplification over Consultation**: Always attempt to refactor and simplify complex snippets first before escalating to the user consultation gate.
- **Explain Intent Before Execution**: Every snippet execution must be preceded by a clear summary of purpose, inputs, and side-effects.
- **Three-Choice Consultation**: When high complexity is unavoidable, always offer the user the choice to split, proceed as-is, or adopt an alternative approach.
