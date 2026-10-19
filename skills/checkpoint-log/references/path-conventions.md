# Checkpoint & Artifact Path Conventions

This document specifies the path resolution and generation-managed file naming conventions for checkpoint logs, procedure guides (runbooks), and reports managed by `checkpoint-log` and related workflows.

---

## 1. Directory Structure Overview

All artifact files (checkpoint logs, runbooks, reports) are stored in a hierarchical, context-scoped directory structure under the designated storage root:

```text
<root>/<project>/<branch>/<context>/<doc_type>-<timestamp>.md
```

### Components

| Component | Description | Resolution Method & Examples |
|---|---|---|
| `<root>` | Base storage directory | Base directory configured in the environment or explicitly confirmed by the user. |
| `<project>` | Project / repository identifier | Top-level repository directory name (e.g. `basename $(git rev-parse --show-toplevel)`). Example: `agent-skills`. |
| `<branch>` | Active git branch or worktree | Current branch name sanitized for filesystem safety (`/` replaced with `-`). Example: `main`, `feat-checkpoint-log`. If detached HEAD, use the short commit hash. |
| `<context>` | Specific task or topic slug | Short, kebab-case identifier describing the concrete objective or issue. Example: `doctor-quick`, `auth-token-leak`, `refactor-parser`. |
| `<doc_type>` | Document type classification | Standard document category: `checkpoint` (milestone log), `runbook` (procedure guide), `report` (investigation/audit report). |
| `<timestamp>` | Generation timestamp | ISO-compact local timestamp: `YYYYMMDD-HHMMSS` (e.g. `20261019-130500`). Ensures immutable generation tracking across multiple runs. |

---

## 2. Supported Document Types

| Document Type | Filename Pattern | Typical Use Cases |
|---|---|---|
| **Checkpoint Log** | `checkpoint-<YYYYMMDD-HHMMSS>.md` | Step-by-step milestone execution records, decision tracking, and context recovery anchors. |
| **Runbook / Procedure** | `runbook-<YYYYMMDD-HHMMSS>.md` | Step-by-step operational runbooks, migration steps, reproduction procedures, or deployment guides. |
| **Report** | `report-<YYYYMMDD-HHMMSS>.md` | Investigation findings, code/project audit summaries, doctor diagnostic outputs, or security assessments. |

---

## 3. Path Resolution Protocol

When preparing to write or read a file:

### Step 3.1: Derive Path Components

1. **Resolve Root**: Check the base storage directory configured in the environment or confirm the base path with the user.
2. **Resolve Project**: Extract the repository name from Git root (`git rev-parse --show-toplevel`).
3. **Resolve Branch**: Query current Git branch (`git branch --show-current`). Replace all `/` with `-` (e.g. `feature/login` becomes `feature-login`).
4. **Determine Context**: Identify or confirm the kebab-case task context with the user (e.g. `doctor-quick`, `issue-42-reproduce`).
5. **Format Timestamp**: Generate current local timestamp in `YYYYMMDD-HHMMSS` format.

### Step 3.2: Confirm with User

Before creating a new file, propose the fully-resolved path to the user for explicit confirmation:

> "I propose recording the `<doc_type>` at:
> `<resolved_file_path>`
> Would you like to proceed with this path or specify an alternative?"

### Step 3.3: Directory Initialization

Ensure parent directories are created cleanly before writing:
```bash
mkdir -p "<root>/<project>/<branch>/<context>"
```

---

## 4. Generation Discovery & Context Recovery

When an agent needs to locate or resume an existing checkpoint, runbook, or report:

1. **Locate Target Directory**:
   Target `<root>/<project>/<branch>/<context>/`.
2. **List Existing Generations**:
   List files matching the document type:
   ```bash
   ls "<root>/<project>/<branch>/<context>/<doc_type>-*.md"
   ```
3. **Select Latest Generation**:
   Sort filenames lexicographically (since timestamps use `YYYYMMDD-HHMMSS`, standard sorting yields chronological order). Select the last entry to access the most recent generation.
4. **Header Verification**:
   Inspect the header block of the selected file to verify that `Project`, `Branch`, and `Context` match the active environment before proceeding.

---

## 5. Sanitization & Safety Rules

- **Filesystem Safety**: Characters outside `[a-zA-Z0-9_-]` in `<project>`, `<branch>`, and `<context>` must be replaced with `-` or `_`.
- **No Path Collisions**: Never overwrite past generation files. If a new run starts for the same context, create a fresh file with a new timestamp.
- **Resuming Within a Generation**: If continuing an in-progress task session, append to the current active generation's checkpoint file rather than creating a new timestamped file, unless the user explicitly requests starting a new generation.
