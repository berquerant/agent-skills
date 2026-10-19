---
name: checkpoint-log
description: >-
  Use this skill when recording step-by-step progress, decisions, and verification
  outcomes to a local file at task milestones, or when resuming interrupted work.
  Enforces explicit user confirmation of the log file path, append-only checkpoint
  entries, and rapid context recovery for both user and agent.
---

# Checkpoint Log

Use this skill to record progressive execution logs to a local file at each task milestone or step, and to review or recover past context quickly.

It benefits both the user and the agent:
- **For the user**: Clear, persistent visibility into the agent's reasoning, executed actions, and ongoing progress without losing track in long chat transcripts.
- **For the agent**: Reliable, low-cost grounding and context recovery across conversation turns or after token context compaction.

For directory hierarchy, generation management, and path conventions, refer to [references/path-conventions.md](references/path-conventions.md). For detailed entry templates and field conventions, refer to [references/log-format.md](references/log-format.md).

---

## Step 1: Resolve the File Path

**Never create or write to a path without explicit user confirmation.**

Resolve the candidate path following [references/path-conventions.md](references/path-conventions.md):
1. **Identify Base Root**: Identify the designated storage root directory configured in the environment or request it from the user.
2. **Derive Project & Branch**:
   - Project: Top-level repository directory name (e.g. `basename $(git rev-parse --show-toplevel)`).
   - Branch: Current active git branch or worktree sanitized for filesystem safety (`/` replaced with `-`).
3. **Determine Context & Document Type**:
   - Context: Brief kebab-case task topic (e.g. `doctor-quick`, `refactor-parser`).
   - Document Type: `checkpoint` (milestone log), `runbook` (procedure guide), or `report` (investigation/audit report).
4. **Propose Candidate Path**:
   - If starting a new run, propose a new generation path with a timestamp: `<root>/<project>/<branch>/<context>/<doc_type>-<YYYYMMDD-HHMMSS>.md`.
   - If resuming or appending to an ongoing run, locate the latest existing generation in `<root>/<project>/<branch>/<context>/` and ask whether to append or start a new generation.

> "I propose recording the checkpoint log at:
> `<resolved_file_path>`
> Would you like to proceed with this path or specify an alternative?"

Once the path is confirmed:
1. Ensure the parent directory exists (`mkdir -p "<parent_dir>"`).
2. If resuming, proceed to Step 4; otherwise, proceed to Step 2.

Verify: The exact file path is confirmed by the user, and the parent directory is established.

---

## Step 2: Initialize the Checkpoint Log

If starting a new task log, write the header block at the top of the file following [references/log-format.md](references/log-format.md):

```markdown
# Task Checkpoint Log: <Short Task Title>

- **Created**: <YYYY-MM-DD HH:MM:SS>
- **Log Path**: `<resolved_file_path>`
- **Project**: `<project_name>`
- **Branch**: `<branch_name>`
- **Context**: `<task_context_slug>`
- **Objective**: <One-sentence summary of the task objective>
- **Status**: IN_PROGRESS

---
```

If resuming an existing log, review existing entries (Step 4) rather than reinitializing.

Verify: The header block exists at the start of the log file with correct task metadata matching project, branch, and context.

---

## Step 3: Record Milestone / Step Checkpoints

At each milestone (e.g. at the completion of each step in `explore-plan-execute`, `step-gate`, or other workflows):

1. **Prepare the entry**: Summarize:
   - **Timestamp**: Current date and time.
   - **Step Title & Status**: e.g. `Step N: <title>` (`COMPLETED`, `FAILED`, `BLOCKED`, or `ADJUSTED`).
   - **Executed Actions**: Specific files edited, tools invoked, or commands run.
   - **Verification & Findings**: Concrete evidence (exit codes, test results, observed facts).
   - **Decisions & Rationale**: Why decisions were made, obstacles handled, or alternatives rejected.
   - **Next Step**: Immediate next action or awaiting user decision.
2. **Append to the file**: Append the entry to the end of the log file without modifying prior entries.
3. **Notify the user**: In the chat update, briefly note that the checkpoint has been appended to `<log_path>`.

Verify: The new milestone entry is cleanly appended to the log file and reflects verified outcomes and rationale.

---

## Step 4: Context Recovery & Review (When Resuming)

When recovering an interrupted task, starting a new session on an ongoing task, or recovering from lost context:

1. **Identify the log file**:
   - Follow [references/path-conventions.md](references/path-conventions.md) Section 4 to locate `<root>/<project>/<branch>/<context>/`.
   - Select the latest generation (`checkpoint-<YYYYMMDD-HHMMSS>.md`) and verify that header metadata (`Project`, `Branch`, `Context`) matches the active task environment.
2. **Read the log file**: Read the entire checkpoint log file (or the latest N entries).
3. **Identify current state**:
   - Locate the last entry marked `COMPLETED` or `BLOCKED`.
   - Read the recorded **Next Step** and **Decisions & Rationale**.
4. **Verify ground truth**: Check the local git status or environment to confirm the recorded state matches disk reality.
5. **Report resumption status**: Summarize to the user where the task left off and propose the next action according to the log.

Verify: The agent accurately identifies the latest completed milestone from the validated generation file and resumes without re-executing completed work.

---

## Step 5: Mark Task Completion

When the overall task is finished:

1. Append a final completion entry summarizing total achievements and final verification.
2. Update the header status from `IN_PROGRESS` to `COMPLETED` (or `ABORTED` if cancelled).

Verify: The log reflects overall task completion and final verification results.

---

## Guidelines

- **Explicit path confirmation with context-scoped conventions.** Propose structured paths incorporating project, branch, and task context with generation timestamps per [references/path-conventions.md](references/path-conventions.md), and always confirm with the user before creating or writing files. Never write to unconfirmed paths.
- **Append-only by default.** Never overwrite, delete, or scramble historical checkpoint entries unless explicitly requested by the user.
- **Record both facts and decisions.** Include not only what was run, but also what the verification showed and why decisions were made.
- **Exclude secrets and massive payloads.** Never record API tokens, private keys, or enormous log dumps. Keep summaries focused and link to raw outputs when necessary.
- **Platform-agnostic & self-contained.** Keep log files standard GitHub-flavored Markdown that can be read by any editor, script, or agent.
