# Retrospective Mode Specification & Scope Rules

This reference defines the two orthogonal mode axes for `agent-retrospective`, the scope discovery rules, supported platform log locations, and the normalized turn schema.

---

## 1. The Two Mode Axes

Every retrospective execution evaluates along two axes: **Focus** (what to look for and propose) and **Scope** (what time window or files to inspect).

```
                      ┌─────────────── Scope Axis ────────────────┐
                      │ session (single) │ recent (window)│ custom (path) │
┌───────────┬─────────┼──────────────────┼────────────────┼───────────────┤
│           │ all     │ [Default]        │ Multi-day audit│ Target audit  │
│           ├─────────┼──────────────────┼────────────────┼───────────────┤
│ Focus     │ auto    │ Quick snippets   │ Batch routines │ -             │
│ Axis      ├─────────┼──────────────────┼────────────────┼───────────────┤
│           │ refine  │ Error retro      │ Skill hardening│ -             │
│           ├─────────┼──────────────────┼────────────────┼───────────────┤
│           │ rules   │ Feedback retro   │ Policy mining  │ -             │
└───────────┴─────────┴──────────────────┴────────────────┴───────────────┘
```

### 1.1 Focus Axis (Analysis Target)

| Focus Value | Name | Description | Output Categories |
|---|---|---|---|
| **`all`**<br>*(Default)* | **Comprehensive Retrospective** | Evaluates all four signals (Corrections, Friction, Deviations, Repetitions). Provides a balanced, prioritized action plan. | Automation + Refinements + Rules |
| **`auto`** | **Workflow Automation** | Filters specifically for repetitive command sequences and multi-step manual workflows. | Command snippets, new skill proposals, user prompt templates |
| **`refine`** | **Asset Refinement** | Focuses on tool errors, recovery loops, and deviations to harden existing skills and documentation. | SKILL.md step diffs, `references/troubleshooting.md`, `references/examples.md` |
| **`rules`** | **Governance & Rules** | Extracts user correction patterns to establish proactive behavioral constraints. | `AGENTS.md`, project instructions, system prompt rules |

### 1.2 Scope Axis (Target Window & Source)

| Scope Value | Scope Definition | Resolution Mechanism |
|---|---|---|
| **`session`**<br>*(Default)* | **Single Session Context** | 1. If running interactively, the current conversation history.<br>2. If given a session ID, `<appDataDir>/brain/<id>/.system_generated/logs/transcript.jsonl` (or platform equivalent).<br>3. If given a file path, that exact file. |
| **`recent`** | **Rolling Time Window** | Scans logs modified within the window (Default: `3d`).<br>Syntax: `recent [Nd|Nw]` (e.g., `recent 3d`, `recent 7d`, `recent 2w`). |
| **`custom`** | **Explicit Path or Pattern** | User-specified path, directory, or glob pattern (e.g., `/path/to/logs/**/*.jsonl`). |

---

## 2. Common Platform Log Locations

To remain platform-agnostic, the agent checks the following locations when discovering log files in `session`, `recent`, or `custom` scopes:

| Platform / Agent | Common Log / Transcript Location | Format |
|---|---|---|
| **Antigravity** | `<appDataDir>/brain/<conversation-id>/.system_generated/logs/transcript.jsonl` | JSON Lines (step_index, source, type, content, tool_calls) |
| **Claude Code** | `~/.claude/projects/<project>/history.jsonl`, `~/.claude/sessions/` | JSON Lines / JSON array |
| **Cursor** | `~/Library/Application Support/Cursor/User/workspaceStorage/<hash>/state.vscdb` or exported chat transcripts | SQLite / Exported Markdown |
| **Codex / OpenCode / Aider** | `.aider.chat.history.md`, `.aider.input.history`, CLI session stdout logs | Markdown / Text / JSONL |
| **Generic CLI** | Specified output redirection or captured transcript files | JSON Lines, JSON, or Plaintext |

If the agent environment does not provide direct filesystem access to other session logs, `Scope: session` operates directly on the active conversation context.

---

## 3. Normalized Turn Schema

Regardless of input log format, logs are normalized into this conceptual model before signal mining:

```json
{
  "session_id": "string",
  "turn_index": 42,
  "timestamp": "2026-11-15T18:00:00Z",
  "role": "user | agent | tool",
  "content": "Normalized text content (secrets masked)",
  "tool_calls": [
    {
      "tool_name": "run_command",
      "arguments": { "CommandLine": "git status" },
      "exit_code": 0,
      "output_snippet": "On branch main...",
      "is_error": false
    }
  ],
  "user_correction": false
}
```
