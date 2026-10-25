# Checkpoint Log Format Specification

This document defines the structure, conventions, and entry templates for checkpoint logs managed by `checkpoint-log`. For file storage paths and directory hierarchy, refer to [path-conventions.md](path-conventions.md).
Per `output-presentation`, record all file paths, locations, and targets as unadorned raw strings without backticks or Markdown links.

---

## 1. File Structure Overview

A checkpoint log file consists of two parts:
1. **Header Block**: Task metadata, objective, start timestamp, and file path.
2. **Sequential Entries**: Chronological log entries appended at the end of each milestone or step.

---

## 2. Header Block Template

When initializing a new checkpoint log file, write the header block at the top:

```markdown
# Task Checkpoint Log: <Short Task Title>

- **Created**: <YYYY-MM-DD HH:MM:SS>
- **Log Path**: <exact_file_path>
- **Project**: `<project_name>`
- **Branch**: `<branch_name>`
- **Context**: `<task_context_slug>`
- **Objective**: <One-sentence summary of the task objective>
- **Status**: IN_PROGRESS

---
```

---

## 3. Step Checkpoint Entry Template

Append a new entry at the bottom of the log file whenever a step or milestone is completed:

```markdown
## [<YYYY-MM-DD HH:MM:SS>] Step <N>: <Step Title>

- **Status**: COMPLETED | FAILED | BLOCKED | ADJUSTED
- **Executed Actions**:
  - <Concrete actions taken, files edited, or commands run>
- **Verification & Findings**:
  - <Exact command used, exit code, test results, or observed facts>
- **Decisions & Rationale**:
  - <Why a specific choice was made, trade-offs considered, or obstacles encountered>
- **Next Step**:
  - <Next planned step or awaiting user approval>

---
```

---

## 4. Entry Field Guidelines

| Field | Purpose | Guidelines |
|---|---|---|
| **Timestamp** | Chronological ordering | Use ISO or standard local time format (e.g. `2026-10-13 15:30:00`). |
| **Status** | State of the step | Use standard keywords: `COMPLETED` (success), `FAILED` (verification failed), `BLOCKED` (waiting for user input/external dependency), `ADJUSTED` (plan modified). |
| **Executed Actions** | Action log | List modified files and key commands concisely. |
| **Verification & Findings** | Verifiable evidence | Record actual test/linter outcomes, exit codes, and key findings. Separate raw facts from interpretation. |
| **Decisions & Rationale** | Cognitive context | Explain *why* a particular direction was chosen or how an unexpected problem was handled. This is critical for future context re-anchoring. |
| **Next Step** | Forward momentum | State what should happen next so resumption is instantaneous. |

---

## 5. Information Safety Rules

- **No Secrets**: Never write API keys, tokens, passwords, or personal credentials into the log.
- **No Massive Dumps**: Avoid dumping entire file contents, huge stack traces, or compiler logs. Summarize error messages or link to local log files instead.
- **Preserve Append-Only Integrity**: Always append to the end of the file. Do not rewrite, reorder, or delete prior milestone entries unless explicitly correcting historical inaccuracies upon user request.
