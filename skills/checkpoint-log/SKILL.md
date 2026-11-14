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

For directory hierarchy, generation management, and path conventions, refer to [path-conventions.md](references/path-conventions.md) (references/path-conventions.md). For detailed entry templates and field conventions, refer to [log-format.md](references/log-format.md) (references/log-format.md).

---

## Prerequisites

- Access to the companion MCP server (`agent-skills-mcp`) providing the `checkpoint` toolset (`checkpoint_init`, `checkpoint_append`, `checkpoint_resume`, `checkpoint_set_status`, `checkpoint_list`). If the MCP server is unavailable, refer to [path-conventions.md](references/path-conventions.md) (references/path-conventions.md) and [log-format.md](references/log-format.md) (references/log-format.md) as the manual file management fallback.

---

## Step 1: Resolve the File Path

**Never create or write to a path without explicit user confirmation.**

Resolve the candidate path following [path-conventions.md](references/path-conventions.md) (references/path-conventions.md):
1. **Identify Base Root**: Identify the designated storage root directory configured in the environment or request it from the user.
2. **Derive Project & Branch**:
   - Project: Top-level repository directory name (e.g. `basename $(git rev-parse --show-toplevel)`).
   - Branch: Current active git branch or worktree sanitized for filesystem safety (`/` replaced with `-`).
3. **Determine Context & Document Type**:
   - Context: Brief kebab-case task topic (e.g. `doctor-quick`, `refactor-parser`).
   - Document Type: `checkpoint` (milestone log), `runbook` (procedure guide), or `report` (investigation/audit report).
4. **Propose Candidate Path**:
   - If starting a new run, propose a new generation path with a timestamp: `<root>/<project>/<branch>/<context>/<doc_type>-<YYYYMMDD-HHMMSS>.md`.
   - If resuming or appending to an ongoing run, locate the latest existing generation via `checkpoint_list` (or directory search) and ask whether to append or start a new generation.

> "I propose recording the checkpoint log at:
> `<resolved_file_path>`
> Would you like to proceed with this path or specify an alternative?"

Once the path is confirmed:
1. Ensure the parent directory exists (`mkdir -p "<parent_dir>"`).
2. If resuming, proceed to Step 4; otherwise, proceed to Step 2.

Verify: The exact file path is confirmed by the user, and the parent directory is established.

---

## Step 2: Initialize the Checkpoint Log

Initialize the log file using the `checkpoint_init` MCP tool, specifying `path`, `project`, `branch`, `context`, `title`, and `objective`. The server verifies that the path and contents contain no sensitive keys or paths and atomically writes the header.

If the MCP server is unavailable in the environment, refer to [log-format.md](references/log-format.md) (references/log-format.md) and write the header block manually:

```markdown
# Task Checkpoint Log: <Short Task Title>

- **Created**: <YYYY-MM-DD HH:MM:SS>
- **Log Path**: <resolved_file_path>
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

1. **Append via MCP**: Call the `checkpoint_append` MCP tool with `path`, `step_title`, `status` (`COMPLETED`, `FAILED`, `BLOCKED`, or `ADJUSTED`), `actions`, `findings`, `decisions`, and `next_step`. The server verifies inputs and atomically appends the entry.
2. **Fallback Manual Append**: If the MCP server is unavailable, format the entry per [log-format.md](references/log-format.md) (references/log-format.md) and append to `<log_path>`.
3. **Notify the user**: In the chat update, briefly note that the checkpoint has been recorded to `<log_path>` (formatted per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md)).

Verify: The new milestone entry is cleanly appended to the log file and reflects verified outcomes and rationale.

---

## Step 4: Context Recovery & Review (When Resuming)

When recovering an interrupted task, starting a new session on an ongoing task, or recovering from lost context:

1. **Parse via MCP**: Call `checkpoint_resume` with the log `path` to obtain structured header metadata, current status, latest milestone details, and recorded next steps.
2. **Fallback Manual Inspection**: If the MCP server is unavailable, refer to [path-conventions.md](references/path-conventions.md) (references/path-conventions.md) Section 4 and read the latest generation file to extract state.
3. **Verify ground truth**: Check the local git status or environment to confirm the recorded state matches disk reality.
4. **Report resumption status**: Summarize to the user where the task left off and propose the next action according to the log.

Verify: The agent accurately identifies the latest completed milestone from the validated generation file and resumes without re-executing completed work.

---

## Step 5: Mark Task Completion

When the overall task is finished:

1. Call `checkpoint_set_status` with `path` and `status: "COMPLETED"` (or `"ABORTED"`).
2. Append a final milestone entry via `checkpoint_append` summarizing total achievements and final verification.
3. If the MCP server is unavailable, refer to [log-format.md](references/log-format.md) (references/log-format.md) to append the summary and update the header manually.

Verify: The log reflects overall task completion and final verification results.

---

## Guidelines

- **Confirm context-scoped paths explicitly**: Propose structured paths incorporating project, branch, and task context with generation timestamps per [path-conventions.md](references/path-conventions.md) (references/path-conventions.md), and confirm with the user before writing so logs are stored in predictable, approved locations.
- **Maintain append-only history**: Append new milestone entries without overwriting or deleting prior entries so the audit trail remains intact across sessions.
- **Record both verified facts and rationale**: Document what was executed, what verification showed, and why decisions were made to enable rapid context recovery after compaction.
- **Keep logs sanitized and concise**: Omit API tokens, private keys, and raw log dumps, linking to external output files when needed to prevent secret exposure and file bloat.
- **Use portable GitHub-flavored Markdown**: Format logs in standard Markdown so any editor, script, or agent can read them without proprietary tooling.
