---
name: step-gate
description: >-
  Use this skill when the user wants reliable, incremental task execution with
  explicit checkpoints. Decomposes a task into appropriately-sized steps,
  generates an execution plan with verification and completion criteria for each
  step, then executes one step at a time — reporting results and requesting user
  approval before proceeding to the next.
---

# Step Gate

Use this skill to execute tasks with maximum reliability by breaking work into
small, independently verifiable steps and gating each one on explicit user
approval before continuing.

---

## Step 1: Understand and Decompose the Task

Read the task carefully and decompose it into subtasks.

**Decomposition guidelines:**

- Each subtask must be **independently verifiable** — success or failure can be
  confirmed without completing later steps.
- Each subtask should be **easy to roll back** if it fails.
- Prefer smaller steps over larger ones when in doubt.
- Order steps by dependency (prerequisite steps come first).

**Avoid these pitfalls:**

- Do not bundle multiple concerns into one step.
- Do not create steps so small they add no value (e.g., "open a file").

Verify: Every subtask has a single, clear responsibility.

---

## Step 2: Build the Execution Plan

For **each** subtask, define the following three elements:

1. **How to execute** — concrete commands, file edits, or actions to perform.
2. **How to verify** — specific commands or observations that confirm the step
   worked (e.g., `go test ./...`, `curl -s http://...`, `grep ... file`).
3. **Completion criteria** — a precise, observable condition that declares this
   step done (e.g., "all tests pass", "file contains the expected line",
   "command exits with code 0").

Present the plan in this format:

```
## Execution Plan

### Step 1: <title>
- **Execute**: <what to do>
- **Verify**: <how to check>
- **Done when**: <completion criterion>

### Step 2: <title>
...
```

Verify: Every step has all three elements defined and each completion criterion
is measurable.

---

## Step 3: Get Plan Approval

Present the full execution plan to the user and ask:

> "Does this plan look right? Should I proceed with Step 1, or would you like
> to adjust anything?"

Do **not** start executing until the user approves the plan.

Verify: The user has explicitly approved the plan (or requested revisions).

---

## Step 4: Execute Steps One at a Time

For each step in the approved plan, repeat this loop:

1. **Execute** the step as defined in the plan.
2. **Verify** using the defined verification method.
3. **Evaluate** against the completion criterion:
   - ✅ **Pass**: report success to the user.
   - ❌ **Fail**: stop, diagnose the issue, propose a fix, and ask the user how
     to proceed. Do **not** continue to the next step.
4. **Report** to the user:
   - What was done
   - What the verification showed
   - Whether the completion criterion was met
5. **Ask for approval** to continue:
   > "Step N complete. Shall I proceed to Step N+1: `<title>`?"

Wait for explicit approval before moving on.

Verify: Each step's completion criterion is satisfied before the next step begins.

---

## Step 5: Final Summary

After all steps are complete, report:

- A summary of all steps completed.
- Any deviations from the original plan and the reasons.
- The overall outcome.

Verify: All completion criteria from the plan are satisfied.

---

## Guidelines

- **One step at a time.** Never execute the next step without explicit user approval.
- **Stop on failure.** Do not silently skip or work around a failed step.
- **Measurable criteria only.** Completion criteria must be observable, not vague ("looks good").
- **Prefer reversible actions.** When multiple approaches exist, prefer the one easiest to undo.
- **No implicit continuation.** Even if the user seems in a hurry, always ask before proceeding.
