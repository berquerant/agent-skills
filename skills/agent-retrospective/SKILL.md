---
name: agent-retrospective
description: >-
  Use this skill to inspect and analyze agent interaction logs (transcripts or session histories) across platforms, extracting recurring patterns to propose command snippets, new skills, workflow routine prompts, repository rule updates, and existing skill or reference refinements.
license: MIT
---

# Agent Retrospective

Use this skill to conduct an evidence-backed retrospective analysis of agent execution histories (transcripts, conversation logs, or current session context).
It mines repetitive patterns, tool friction, and user corrections from real agent interactions, turning them into actionable proposals for workflow automation, repository rule hardening, and skill/reference refinements.

This skill is **strictly read-only**. It analyzes logs, diagnoses friction points, drafts concrete templates and snippets, and stops at an approval gate before handing off mutations to dedicated creation and refactoring skills.

For reference details, see:
- [modes.md](references/modes.md) (references/modes.md): The two mode axes (Focus × Scope), defaults, and platform log locations.
- [signals.md](references/signals.md) (references/signals.md): Detection criteria and extraction heuristics for the 4 diagnostic signals.
- [report-template.md](references/report-template.md) (references/report-template.md): The structured retrospective report and proposal template.

---

## Prerequisites

- Read access to the target agent logs, conversation transcripts, or active session context.
- Target scope identified: current session, a specified timeframe (`recent [Nd]`), or custom log directory/files.
- Familiarity with the companion skills for mutation hand-offs:
  - [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md) for parameterizing repeated shell pipelines.
  - [skill-creator](../skill-creator/SKILL.md) (../skill-creator/SKILL.md) for authoring new multi-step workflow skills.
  - [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md) / [refactor](../refactor/SKILL.md) (../refactor/SKILL.md) for modifying documentation, rules, or references.

---

## Step 1: Select Mode & Fix the Log Scope

1. Confirm or select one option for each mode axis, as defined in [modes.md](references/modes.md) (references/modes.md):
   - **Focus Axis (Analysis Purpose)**:
     - `all` (**Default**): Comprehensive retrospective across all four diagnostic signals and improvement categories.
     - `auto`: Workflow automation focus (identifies candidates for `command-snippet`, new skills, or prompt routines).
     - `refine`: Existing asset refinement focus (identifies friction to enrich `troubleshooting.md`, `examples.md`, or modify `SKILL.md` steps).
     - `rules`: Governance & rule focus (identifies user corrections to harden `AGENTS.md`, project rules, or system prompt constraints).
   - **Scope Axis (Timeframe & Source)**:
     - `session` (**Default**): The current interactive session transcript or a single user-specified conversation log.
     - `recent` (Default `3d`, configurable e.g., `recent 7d`): Multi-session logs within the specified rolling window.
     - `custom`: Explicitly provided log files, directories, or glob patterns.
2. If the user did not specify preferences, assume the default combination (`Focus: all` × `Scope: session`) and state it clearly.
3. Locate and list the candidate transcript/log files matching the Scope, verifying read access.

Verify: The Focus mode, Scope timeframe, candidate file list, and target log count are clearly stated.

---

## Step 2: Ingest & Normalize Session Logs

1. Read the scoped logs using platform-agnostic tools (`read_file`, grep, or line-oriented JSON parsers).
2. Normalize turns into an abstract event structure:
   - `Turn { role: "user" | "agent" | "tool", content: string, tool_calls: list, status: "success" | "error", timestamp: string }`
3. **Mandatory Sanitization & Secret Masking**:
   - Redact API keys, authentication tokens, passwords, private domain names, absolute local user paths, and sensitive internal credentials from parsed contents before including them in downstream analysis or reports.
   - Replace project-specific volatile values (e.g., commit hashes, temporary branch names, ephemeral container IDs) with generic placeholders (`<COMMIT_HASH>`, `<BRANCH_NAME>`).

Verify: Log files are ingested, turns are normalized, and sensitive credentials/volatile values are masked.

---

## Step 3: Mine Diagnostic Signals

Examine the normalized conversation turns against the 4 diagnostic signals defined in [signals.md](references/signals.md) (references/signals.md):

1. **Signal 1: Correction Signals (User Interventions & Explicit Constraints)**
   - Identify user prompts that countermanded or redirected agent actions (e.g., "no, use X instead", "don't touch Y", "prefer Z").
   - Extract the implicit preference or invariant rule behind the correction.
2. **Signal 2: Friction Signals (Failures, Retries & Tool Confusion)**
   - Scan for non-zero tool exit codes, tool invocation syntax errors, missing command prerequisites, or repeated back-and-forth attempts.
   - Note the root cause: missing documentation, platform-specific CLI quirks (e.g., BSD vs GNU utilities), or ambiguous skill instructions.
3. **Signal 3: Deviation Signals (Workflow Step Skips & Instruction Drift)**
   - For turns where an existing skill was active, compare actual tool actions against the skill's defined steps.
   - Detect whether the agent skipped verification checkpoints or had to invent ad-hoc steps not documented in `SKILL.md`.
4. **Signal 4: Repetition Signals (Recurring Commands & Routine Prompts)**
   - Identify identical or near-identical shell command pipelines across turns or sessions.
   - Identify repetitive manual prompt sequences that the user repeatedly provides to start tasks.

Verify: Every detected signal is recorded with concrete evidence (turn index, log file, error snippet, or occurrence frequency).

---

## Step 4: Formulate Prioritized Improvement Proposals

Organize the findings into actionable improvement proposals according to the active Focus mode, using the structure in [report-template.md](references/report-template.md) (references/report-template.md):

1. **Category A: Automation Candidates (from S4 & S2)**
   - **Command Snippets**: Propose parameterized snippet definitions (with required flags, avoid inline raw execution) for [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md).
   - **New Skills**: For recurring multi-step procedures (exploration + execution + verification), draft the `SKILL.md` outline for [skill-creator](../skill-creator/SKILL.md) (../skill-creator/SKILL.md).
   - **Routine Prompts**: Draft standardized prompt templates for repetitive user requests.
2. **Category B: Existing Asset Refinements (from S2 & S3)**
   - **Skill Step Clarifications**: Propose exact diffs or bullet additions to existing `skills/<name>/SKILL.md` to prevent step skips or clarify preconditions.
   - **Reference Additions**: Draft concrete Q&A entries for `references/troubleshooting.md` or platform-specific examples for `references/examples.md`.
3. **Category C: Rules & Governance Hardening (from S1)**
   - Propose candidate rules for `AGENTS.md`, project instruction files, or `.gemini/rules/` to proactively eliminate recurring user corrections.

Verify: Each proposal is categorized, provides a concrete draft (code, rule snippet, or text diff), and cites supporting evidence.

---

## Step 5: User Approval Gate & Hand-off

1. Present the retrospective report clearly to the user with a summary table prioritizing recommendations by impact (High / Medium / Low).
2. Explicitly ask the user which recommendations they wish to adopt:
   - "Which of these proposed improvements would you like to implement?"
3. Upon receiving explicit user approval, recommend or hand off to the appropriate mutation skill:
   - For command snippets → Hand off to [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md).
   - For new skills → Hand off to [skill-creator](../skill-creator/SKILL.md) (../skill-creator/SKILL.md).
   - For documentation / rule / reference updates → Hand off to [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md).

Verify: No files or configurations are mutated without explicit user approval of specific proposals.

---

## Guidelines

- **Strictly Read-Only Analysis**: Do not create or edit snippet files, rule files, or skill definitions during this retrospective. Diagnose and propose first; let the user select what to execute.
- **Evidence-Backed Recommendations**: Never speculate. Every recommendation must cite at least one turn index, command error output, or user correction event from the logs.
- **Privacy & Sanitization by Default**: Treat all transcripts as potentially containing sensitive information. Redact secrets, tokens, internal paths, and project-specific identifiers before presenting proposals.
- **Universal & Platform-Agnostic**: Support log analysis regardless of whether the agent runs in Antigravity, Claude Code, Cursor, Codex, or generic CLI harnesses.
- **Cost & Token Awareness**: When running in `recent` mode over large log sets, use targeted search (grep, pattern scanning) rather than reading thousands of raw log lines entirely into the primary conversation context.
