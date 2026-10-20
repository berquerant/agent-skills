---
name: skill-router
description: >-
  Use this skill to analyze user requests, workspace context, and project state, then select and route to the optimal agent skill or pipeline. Formulates execution sequences with approval gates before handing off.
---

# Skill Router

Use this skill to triage user requests, inspect the current workspace context, and select the most appropriate agent skill or multi-skill pipeline.
It acts as an intelligent entry dispatcher that aligns user intent with available skills, formulates a sequential execution plan, and pauses for user approval before handing off.

This skill is **strictly read-only**. It never edits code, creates commits, or modifies files.

---

## Prerequisites

- Read access to the target workspace or repository.
- Standard shell inspection tools (`git`, `grep`, `find`) to examine workspace status.

---

## Step 1: Analyze Request & Workspace Context

Examine the user prompt and the current state of the workspace to capture relevant context:

1. **User Request Parsing**:
   - Identify the user's primary goal (e.g. feature addition, bug fix, refactoring, audit, documentation, dependency update, branch isolation).
   - Note any stated constraints, preferences (e.g. "step-by-step", "read-only", "fast"), or target files/modules.
2. **Workspace & Git State Inspection**:
   - Check working tree state: `git status --short`.
   - Check if uncommitted changes or WIP diffs exist: `git diff --stat`.
   - Check current branch name and default branch tracking.
   - Check if recent commits or PRs were referenced.

Verify: The user request intent, workspace git state, and target scope are recorded.

---

## Step 2: Match Intent & Select Candidate Skills

Match the analyzed request and context against [references/intent-matrix.md](references/intent-matrix.md):

1. **Category Mapping**:
   - Classify the request into one of the primary intent categories (e.g. Feature Work, Holistic Health Check, Code Review, Refactoring, Dependency PRs).
2. **Skill Candidate Evaluation**:
   - If the task is simple and matches a single skill directly (e.g. pure commit message generation), select that skill directly.
   - If the task is non-trivial or carries risk, evaluate whether a composite pipeline is needed (e.g. diagnosis before refactoring, or planning before step-gate execution).
   - Check sibling skill descriptions and criteria to confirm fit.

Verify: The mapped intent category, selected candidate skill(s), and selection rationale are clearly stated.

---

## Step 3: Compose Skill Sequence & Execution Plan

Formulate a clear execution sequence based on the selected candidates and pipeline patterns in [references/intent-matrix.md](references/intent-matrix.md):

1. **Define Pipeline Sequence**:
   - **Entry Skill**: The first skill to activate (e.g. [`doctor`](../doctor/SKILL.md), [`explore-plan-execute`](../explore-plan-execute/SKILL.md), or [`git-worktree`](../git-worktree/SKILL.md)).
   - **Follow-up Skills (Sequential Hand-offs)**: The subsequent skills that will handle execution, review, or finalization upon user approval.
2. **Draft Proposed Plan**:
   - State the target scope and baseline.
   - Summarize the stages and what each skill will produce.
   - Specify the hand-off checkpoints between skills.

Verify: The execution plan defines the entry skill, downstream hand-off sequence, and success criteria.

---

## Step 4: Present Plan & User Approval Gate

Present the routing recommendation and execution plan to the user:

```markdown
### 🧭 Recommended Skill Execution Plan

**Task Intent**: <Identified category and goal>
**Target Scope**: <Files, branches, or repositories involved>

#### Proposed Skill Sequence
1. **Entry**: `<skill-name>` — <Purpose and expected output>
2. **Next**: `<skill-name>` — <Purpose upon approval>
3. **Finalization**: `<skill-name>` — <Commit, PR, or verification>

#### Why this sequence?
<1-2 sentences explaining why this pipeline best matches the request>
```

Ask the user:
> "Does this execution sequence look right? Shall I proceed by delegating to `<entry-skill>`?"

Do **not** modify files or invoke mutation skills before receiving user confirmation.

Verify: The proposed skill pipeline is presented to the user, and explicit approval is obtained.

---

## Step 5: Hand Off to Initial Skill

Upon user approval, hand off execution to the approved entry skill:

- Provide the entry skill with the analyzed context, scope, and objectives.
- Do not perform the downstream steps here; allow each skill in the sequence to manage its own lifecycle and approval checkpoints.

Verify: The approved entry skill is initiated with the established context.

---

## Guidelines

- **Strictly read-only.** Never modify code, configuration, or documentation within this skill.
- **Always seek approval before dispatching.** Never unilaterally start a multi-skill pipeline without the user confirming the proposed flow.
- **Loosely coupled hand-offs.** Skills do not execute each other in hidden chains; each hand-off must produce visible outputs and respect user checkpoints.
- **Prefer minimal pipelines.** Do not chain unnecessary skills for straightforward tasks. If a task is a pure commit message request, route directly to [`change-message`](../change-message/SKILL.md).
