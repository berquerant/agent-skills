---
name: explore-plan-execute
description: >-
  Use this skill when the user gives a request or prompt requiring non-trivial
  work. Guides the agent to explore the codebase and gather information, build
  a concrete plan with success criteria, present it to the user for approval,
  and only then execute — preventing wasted effort and misaligned output.
---

# Explore → Plan → Execute

This skill enforces a structured workflow that reduces errors and misalignment
on any non-trivial user request: explore first, make a plan with clear success
criteria, get user sign-off, then execute.

It embodies two key best practices:
- **Explore, plan, then execute** — read before writing; think before acting.
- **Establish verification loops** — define what "done" looks like up front so
  correctness can be confirmed at each step.

---

## Step 1: Understand the Request

Read the user's request carefully.

- Identify the **core goal** (what the user ultimately wants to achieve).
- Identify **constraints** (language, framework, style, existing conventions).
- Identify **ambiguities** that would block planning.

If the intent is unclear, ask the user **one focused clarifying question** at a
time. Do not ask multiple questions in one turn.

Verify: You can articulate the goal in a single sentence without hedging.

---

## Step 2: Explore and Gather Information

Before writing any code or making any changes, explore the relevant context.

- Read related source files, configs, tests, and docs.
- Run read-only commands (e.g. `grep`, `find`, `git log`) to understand the
  current state.
- Check for existing patterns, naming conventions, and abstractions to follow.
- Identify potential side-effects or impacted areas.
- **Minimize investigation cost**: When using high-cost exploration methods (e.g. APIs, remote queries, heavy tool executions), minimize total cost across latency, local compute, and server-side resource consumption. Fetch potentially needed content into a local file or temporary cache up front and query the local file for subsequent exploration instead of making repetitive API calls.

Do **not** modify any files during this step.

Verify: You have enough information to write a concrete, step-by-step plan.

---

## Step 3: Build a Plan

Construct a plan that includes:

1. **Objective** — one-sentence restatement of the goal.
2. **Scope** — what will and will not be changed.
3. **Steps** — an ordered list of concrete actions (files to edit, commands to
   run, tests to add, etc.).
4. **Success criteria** — specific, observable conditions that confirm the
   request is satisfied (e.g. "all existing tests pass", "the new endpoint
   returns 200 for valid input", "the output matches the expected format").
5. **Risks / open questions** — any assumptions made or decisions that may need
   user input.

Keep the plan concise enough to present clearly in one message.

### Git Commits, Remotes, and Change Requests (PR / MR)

If the plan involves creating git commits, interacting with remotes, or opening PRs/MRs, incorporate the following:

- **Commit Author**: Determine the intended commit author `name` and `email` for commits created on the base/target branch. If not specified by the user, include an explicit question asking if referencing the global `$HOME/.gitconfig` (`git config --global user.name` and `user.email`) is permitted.
- **Git `sshCommand`**: Check if remote operations (e.g. push, fetch) will use SSH. Confirm what `sshCommand` (`core.sshCommand` or `GIT_SSH_COMMAND`, specifying SSH key or parameters) should be used.
- **PR / MR Assignee**: Determine who the `assignee` should be for any PR or MR to be created.

### Git Worktree and Branch Isolation

If the task involves modifying files in a git repository:
- Plan to start work in an isolated worktree created from the latest default branch.
- Include steps to fetch the latest remote changes (`git fetch origin`), create a new branch from the latest remote default branch (e.g. `origin/main`), and set up the worktree (following [`git-worktree`](../git-worktree/SKILL.md)).

### Decide Whether to Use `scrutinize-plan-execute`

Analyze whether the task requires deep upfront scrutiny, risk evaluation, or interactive refinement before execution.

**Delegate to [`scrutinize-plan-execute`](../scrutinize-plan-execute/SKILL.md) when any of the following apply:**

| Signal | Example |
|---|---|
| Ambiguous premises or conditions | Unverified environment behavior, incomplete requirements |
| Unclear or subjective success criteria | "Improve performance", ambiguous verification logic |
| High-impact execution obstacles | Potential breaking changes, complex multi-component interactions |
| Need for deep user consultation | Multiple competing solutions with trade-offs requiring user dialogue |
| Need for plan or report archiving | Long architecture plans or extensive reports requiring user-approved file archiving |

If delegating to `scrutinize-plan-execute`, hand off the workflow to [`scrutinize-plan-execute`](../scrutinize-plan-execute/SKILL.md) to complete obstacle scrutiny, user consultation, archiving checks, and execution.

### Decide Whether to Use `step-gate`

Analyze the task's nature and determine whether to apply the
[`step-gate`](../step-gate/SKILL.md) skill during execution.

**Use `step-gate` when any of the following apply:**

| Signal | Example |
|--------|---------|
| Low reversibility | Data deletion, DB migration, production deployment |
| High cost of mid-task failure | Long multi-phase refactor, infrastructure change |
| Many steps (≥ 5) with strong dependencies | Sequential setup with each step gating the next |
| User explicitly asked for caution | "Be careful", "confirm with me", "step by step" |

**Skip `step-gate` when:**

- The task is read-only or exploratory (no file edits or state changes).
- The task completes in 1–2 steps with negligible risk.
- All actions are trivially reversible (e.g. editing a draft file).

State the decision in the plan:

> **Execution mode**: `step-gate` / standard

If using `step-gate`, note that Step 5 will follow the step-gate protocol
(per-step verification + user approval before continuing).

Verify: Each step is actionable. Each success criterion is testable.
Execution mode is explicitly stated in the plan (or delegated to `scrutinize-plan-execute`).

---

## Step 4: Present the Plan to the User

Share the plan with the user before taking any action.

Format the plan clearly (use numbered lists, code spans, or a markdown table as
appropriate). Highlight the success criteria explicitly.

When git commits, remotes, or PRs/MRs are involved, explicitly present the following for confirmation:
- Commit author `name` and `email` to be used for the commits (if unspecified, ask: "May I reference your global `$HOME/.gitconfig`?")
- Git `sshCommand` for remote operations (e.g. SSH key or options)
- MR / PR `assignee`

Ask the user:
> "Does this plan look right? Should I proceed, or would you like to adjust
> anything?"

Do **not** start executing until the user confirms.

Verify: The user has read the plan and given explicit approval (or requested
changes), including author, sshCommand, and assignee confirmations if git operations are planned.

---

## Step 5: Execute

Execute the approved plan according to the execution mode decided in Step 3.

### Mode: `step-gate`

Follow the [`step-gate`](../step-gate/SKILL.md) skill protocol:

- Execute one step at a time.
- Verify the completion criterion after each step.
- Report results to the user and **request explicit approval** before
  proceeding to the next step.
- Stop and diagnose on any failure; do not continue unilaterally.

### Mode: standard

Execute the plan step by step with internal verification loops:

- Run the relevant verification after each meaningful step (tests, linters,
  type-checkers, dry-runs, etc.).
- Confirm the success criterion for that step is met before continuing.
- If a step fails, diagnose the issue and propose a fix — do **not** silently
  skip or work around it.

### Worktree Setup Before Modification

When modifying files in a git repository:
- Fetch the latest changes from the remote repository (`git fetch origin`).
- Create a new branch from the latest remote default branch and set up an isolated worktree (following [`git-worktree`](../git-worktree/SKILL.md)).
- Switch to the newly created worktree directory and begin all file modifications there to keep the primary working tree clean.

### Safeguards for Git Commits, Remotes, and PRs/MRs

When execution involves modifying git history or remote hosting:
- **Before creating commits**: Verify that the commit author `name` and `email` have been explicitly confirmed by the user. If unconfirmed and the user approved referencing `$HOME/.gitconfig`, inspect it (`git config --global user.name` / `user.email`) before committing. Never commit under an arbitrary or unverified author.
- **Before pushing or remote operations**: Verify that the git `sshCommand` (e.g. `core.sshCommand` or `GIT_SSH_COMMAND`) has been confirmed with the user.
- **Before opening PRs/MRs**: Verify that the `assignee` has been explicitly confirmed with the user.

### Completion (both modes)

At the end:
- Verify **all** success criteria from the plan.
- Report the outcome to the user, including any deviations from the original
  plan and why they were made.

Verify: All success criteria are satisfied. No regressions introduced.

---

## Guidelines

- **Read before writing.** Exploration (Step 2) is never optional.
- **One approval gate.** Always pause at Step 4; never skip it.
- **Declare execution mode.** Always state whether `step-gate` or standard
  mode will be used, and why, before the user approves the plan.
- **Small, verifiable steps.** Break large plans into checkpoints so errors are
  caught early.
- **Surface surprises.** If you discover something unexpected during execution,
  stop and tell the user instead of improvising.
- **Prefer reversible actions.** When multiple approaches exist, prefer the one
  that is easiest to undo.
- **Start work in a Git Worktree.** When modifying code in a git repository,
  always fetch the latest remote default branch, create a new branch from it
  using `git worktree`, and begin work inside the worktree before modifying
  files.
- **Pre-confirm commit author, `sshCommand`, and MR/PR assignee.** Never commit,
  push via SSH, or open PRs/MRs with assumed identities or settings. Always
  pre-confirm the commit author `name` and `email` (asking if referencing
  `$HOME/.gitconfig` is allowed if unspecified), git `sshCommand`, and MR/PR
  `assignee`.
- **Cost-aware exploration.** When using APIs or expensive operations to
  gather context, fetch data once into a local file or cache and reference it
  locally to minimize latency, local resources, and server-side consumption.
