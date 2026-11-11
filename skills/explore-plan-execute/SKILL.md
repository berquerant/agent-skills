---
name: explore-plan-execute
description: >-
  Use this skill when a non-trivial task requires structured exploration, concrete
  planning with verification criteria, and user approval before execution. For
  high-risk or architecturally ambiguous tasks, prefer scrutinize-plan-execute.
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
- **Minimize investigation cost**: When using high-cost exploration methods (e.g. APIs, remote queries, heavy tool executions), minimize total cost across latency, local compute, and server-side resource consumption. Fetch potentially needed content into a local file or temporary cache up front and query the local file for subsequent exploration instead of making repetitive API calls. When caching command or tool (including MCP) execution results, prioritize using a dedicated caching or memoization skill, tool, or MCP server if available in the environment; otherwise, evaluate idempotency and follow the file-based caching and TTL protocol in [command-cache-protocol.md](references/command-cache-protocol.md) (references/command-cache-protocol.md).
- **Evidence-backed inquiry**: If exploration requires diagnosing ambiguous bugs, investigating root causes, or establishing verifiable evidence before planning, adopt the source scoping and fact verification guidelines from [evidentiary-inquiry](../evidentiary-inquiry/SKILL.md) (../evidentiary-inquiry/SKILL.md).

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

### Execution Safeguards & Pre-Confirmations

Incorporate mandatory environment and safety protocols per [execution-safeguards.md](references/execution-safeguards.md) (references/execution-safeguards.md):
- **Git Commits & PRs/MRs**: Pre-confirm commit author identity, SSH configuration, assignee, and draft messages following [change-message](../change-message/SKILL.md) (../change-message/SKILL.md).
- **Git Worktree Isolation**: Plan work inside an isolated worktree created from the remote default branch per [git-worktree](../git-worktree/SKILL.md) (../git-worktree/SKILL.md).
- **File Archiving**: Propose and confirm file paths for extensive documents (≥300 lines) instead of outputting directly to chat.
- **Command Snippets**: Plan complex shell commands as parameterized snippet files per [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md).

### Workflow Coordination and Secondary Skills

Evaluate whether the task requires specialized workflow, scheduling, gating, or logging skills during execution. Refer to [workflow-coordination-matrix.md](references/workflow-coordination-matrix.md) (references/workflow-coordination-matrix.md) for full decision signals, prerequisites, and plan declaration templates.

| Need / Signal | Candidate Skill | Action in Plan |
|---|---|---|
| Ambiguous premises, high risk, multi-approach trade-offs | [scrutinize-plan-execute](../scrutinize-plan-execute/SKILL.md) (../scrutinize-plan-execute/SKILL.md) | Hand off planning, scrutiny, and execution |
| Independent task branches, asynchronous idle waits (CI/builds) | [task-dependency-pipeline](../task-dependency-pipeline/SKILL.md) (../task-dependency-pipeline/SKILL.md) | Declare `task-dependency-pipeline` scheduling mode |
| Low reversibility, ≥5 steps, user asked for step-by-step caution | [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md) | Declare `step-gate` execution mode |
| Extended multi-turn work, complex refactors needing audit trail | [checkpoint-log](../checkpoint-log/SKILL.md) (../checkpoint-log/SKILL.md) | Propose milestone log and confirm path with user |
| High uncertainty, slow commands, risk of silent stalls | [edit-paced-execution](../edit-paced-execution/SKILL.md) (../edit-paced-execution/SKILL.md) | Propose pacing guard thresholds and command mode |

Declare the selected execution mode, scheduling mode, milestone log proposal, and pacing guard in the plan presented to the user.

Verify: Each step is actionable. Each success criterion is testable.
Execution mode, any milestone logging proposal, and any recommended pacing guard are explicitly stated in the plan (or delegated to `scrutinize-plan-execute`).


---

## Step 4: Present the Plan to the User

Share the plan with the user before taking any action.

Format the plan clearly (use numbered lists, code spans, or a markdown table as
appropriate). Highlight the success criteria explicitly.

If the plan, report, or procedure exceeds 300 lines or is intended for repeated reference in subsequent tasks, propose saving it to a local file and confirm the target path with the user before proceeding.

When git commits, remotes, or PRs/MRs are involved, explicitly present the following for confirmation:
- Commit author `name` and `email` to be used for the commits (if unspecified, ask: "May I reference your global `$HOME/.gitconfig`?")
- Git `sshCommand` for remote operations (e.g. SSH key or options)
- MR / PR `assignee`

Ask the user (notifying via [user-notification](../user-notification/SKILL.md) (../user-notification/SKILL.md) when appropriate):
> "Does this plan look right? Should I proceed, or would you like to adjust
> anything?"

Do **not** start executing until the user confirms.

Verify: The user has read the plan and given explicit approval (or requested
changes), including author, sshCommand, and assignee confirmations if git operations are planned.

---

## Step 5: Execute

Execute the approved plan according to the execution mode decided in Step 3.

### Mode: `step-gate`

Follow the [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md) skill protocol:

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
- If [checkpoint-log](../checkpoint-log/SKILL.md) (../checkpoint-log/SKILL.md) is enabled, append a checkpoint entry recording actions, verification outcomes, decisions, and next steps to the user-confirmed log file.
- If a step fails, diagnose the issue and propose a fix — do **not** silently
  skip or work around it.


### Execution Safeguards Enforcement

Strictly apply the protocols detailed in [execution-safeguards.md](references/execution-safeguards.md) (references/execution-safeguards.md) during execution:
- **Worktree Isolation**: Begin work in the designated git worktree directory to keep the primary working tree clean per [git-worktree](../git-worktree/SKILL.md) (../git-worktree/SKILL.md).
- **Command Snippets**: Route complex shell commands through versioned snippet files following [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md) and [snippet-review](../snippet-review/SKILL.md) (../snippet-review/SKILL.md).
- **Git Pre-Confirmations**: Verify that commit author identity, SSH configuration, and PR/MR assignee have been explicitly confirmed before committing, pushing, or opening PRs/MRs.

### Completion (both modes)

At the end:
- Verify **all** success criteria from the plan.
- Report the outcome to the user, including any deviations from the original
  plan and why they were made.
- Propose appropriate follow-up skills (Sequential Hand-off):
  - **For drafting commit messages and PR/MR descriptions**: Recommend [change-message](../change-message/SKILL.md) (../change-message/SKILL.md).
  - **For inspecting post-execution working tree and git drift**: Recommend [project-status](../project-status/SKILL.md) (../project-status/SKILL.md).
  - **For conducting a comprehensive diff review**: Recommend [code-review](../code-review/SKILL.md) (../code-review/SKILL.md).

Verify: All success criteria are satisfied. No regressions introduced.

---

## Guidelines

- **Explore before editing**: Complete read-only exploration (Step 2) before modifying files so plans are grounded in actual repository conventions and architecture.
- **Secure explicit approval at Step 4**: Present the plan and wait for user sign-off before executing changes so scope and verification criteria are aligned up front.
- **Declare execution mode with rationale**: State whether `step-gate` or `standard` mode will be used, and why, before the user approves the plan.
- **Decompose into verifiable checkpoints**: Structure steps with observable completion criteria so regressions are caught immediately.
- **Surface unexpected findings immediately**: Pause and inform the user when surprises arise during execution rather than improvising unapproved changes.
- **Prefer reversible actions**: Select the approach easiest to roll back when multiple valid implementation paths exist.
- **Isolate branch work in a Git Worktree**: Fetch the latest remote default branch and create an isolated worktree per [git-worktree](../git-worktree/SKILL.md) (../git-worktree/SKILL.md) when modifying repository code.
- **Pre-confirm git identity, `sshCommand`, and assignee**: Verify commit author `name` and `email`, git `sshCommand`, and MR/PR `assignee` with the user before committing or pushing.
- **Archive extensive documents**: Save reports ≥ 300 lines or reusable procedure manuals to a user-confirmed file path rather than outputting massive blocks in chat.
- **Keep shell commands concise**: Follow [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md) to convert long (≥120 chars), chained (≥4 operators), or multi-line (≥3 lines) commands into documented snippet files with explicit CLI arguments.
- **Cache high-cost exploration results**: Fetch expensive API or query outputs once into a dedicated cache tool or local file following [command-cache-protocol.md](references/command-cache-protocol.md) (references/command-cache-protocol.md) to minimize latency and resource consumption.
