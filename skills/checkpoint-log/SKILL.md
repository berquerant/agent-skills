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

For the detailed entry templates and field conventions, refer to [references/log-format.md](references/log-format.md).

---

## Step 1: Resolve the Log File Path

**Never assume or default to a fixed file path.**

Always ask the user explicitly where they would like to store the checkpoint log file:

> "Which file path should I use to record task progress and checkpoint logs?"

Once the user provides a path:
1. Verify the parent directory is writable.
2. Check whether the file already exists:
   - If it exists and contains an existing log, ask if the agent should **resume/append** to it or start a new file.
   - If it does not exist, prepare to initialize it in Step 2.

Verify: The exact log file path is confirmed by the user, and write access to the target path is established.

---

## Step 2: Initialize the Checkpoint Log

If starting a new task log, write the header block at the top of the file following [references/log-format.md](references/log-format.md):

```markdown
# Task Checkpoint Log: <Short Task Title>

- **Created**: <YYYY-MM-DD HH:MM:SS>
- **Log Path**: `<resolved_file_path>`
- **Objective**: <One-sentence summary of the task objective>
- **Status**: IN_PROGRESS

---
```

If resuming an existing log, review existing entries (Step 4) rather than reinitializing.

Verify: The header block exists at the start of the log file with correct task metadata.

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

1. **Read the log file**: Read the entire checkpoint log file (or the latest N entries).
2. **Identify current state**:
   - Locate the last entry marked `COMPLETED` or `BLOCKED`.
   - Read the recorded **Next Step** and **Decisions & Rationale**.
3. **Verify ground truth**: Check the local git status or environment to confirm the recorded state matches disk reality.
4. **Report resumption status**: Summarize to the user where the task left off and propose the next action according to the log.

Verify: The agent accurately identifies the latest completed milestone and resumes without re-executing completed work.

---

## Step 5: Mark Task Completion

When the overall task is finished:

1. Append a final completion entry summarizing total achievements and final verification.
2. Update the header status from `IN_PROGRESS` to `COMPLETED` (or `ABORTED` if cancelled).

Verify: The log reflects overall task completion and final verification results.

---

## Guidelines

- **Never use default or inferred paths.** Always ask the user for their desired log file path before creating or writing files.
- **Append-only by default.** Never overwrite, delete, or scramble historical checkpoint entries unless explicitly requested by the user.
- **Record both facts and decisions.** Include not only what was run, but also what the verification showed and why decisions were made.
- **Exclude secrets and massive payloads.** Never record API tokens, private keys, or enormous log dumps. Keep summaries focused and link to raw outputs when necessary.
- **Platform-agnostic & self-contained.** Keep log files standard GitHub-flavored Markdown that can be read by any editor, script, or agent.
