# Retrospective Diagnostic Signals

This reference details the four diagnostic signals mined during an agent retrospective, their detection heuristics, and extraction procedures.

---

## 1. The Four Diagnostic Signals

```
                    ┌──────────────────────────────────────────────┐
                    │      Session Transcript / Normalized Turns   │
                    └──────────────────────┬───────────────────────┘
                                           │
         ┌──────────────────┬──────────────┴─────┬──────────────────┐
         ▼                  ▼                    ▼                  ▼
┌──────────────────┐┌──────────────────┐┌──────────────────┐┌──────────────────┐
│ S1: Correction   ││ S2: Friction     ││ S3: Deviation    ││ S4: Repetition   │
│ Human redirection││ Errors & retries ││ Skipped steps    ││ Repeated command │
│ & implicit rules ││ Tool confusion   ││ Ad-hoc steps     ││ & prompt patterns│
└────────┬─────────┘└────────┬─────────┘└────────┬─────────┘└────────┬─────────┘
         ▼                   ▼                   ▼                   ▼
┌──────────────────┐┌──────────────────┐┌──────────────────┐┌──────────────────┐
│ Governance Rules ││ Troubleshooting  ││ Skill Step Diffs ││ Snippets & New   │
│ AGENTS.md / rules││ & Examples       ││ & Preconditions  ││ Skills (Creator) │
└──────────────────┘└──────────────────┘└──────────────────┘└──────────────────┘
```

---

## 2. Signal 1: Correction Signals (User Interventions)

### Definition
Events where the user manually redirected, corrected, or constrained the agent's behavior after the agent took an unwanted action or proposed an inappropriate approach.

### Detection Heuristics
- **Semantic Cues in User Turns**:
  - Negations and redirections: `"no"`, `"don't"`, `"do not"`, `"stop"`, `"instead"`, `"never"`, `"prefer"`, `"ではなく"`, `"しないで"`, `"使わないで"`, `"方針変更"`.
  - Preference clarification: `"always use <X>"`, `"ensure you run <Y>"`, `"remember to"`.
- **Structural Cues**:
  - User turn immediately following an agent turn with tool execution where the user requests rolling back or doing it differently.

### Extraction & Rule Formulation
1. Identify the *agent action* that triggered the correction.
2. Formulate the *invariant rule* that would have prevented the mistake.
3. Recommend placement:
   - Repository-specific: `AGENTS.md` or `.gemini/rules/`.
   - Global: Agent system prompt / global instructions.

---

## 3. Signal 2: Friction Signals (Failures & Retries)

### Definition
Moments of inefficiency, tool execution errors, syntax mismatches, or multi-turn trial-and-error where the agent struggled before succeeding (or gave up).

### Detection Heuristics
- **Tool Exit Failures**:
  - Shell command exit code $\neq 0$.
  - MCP tool returns `isError: true` or error strings (`"command not found"`, `"permission denied"`, `"syntax error"`).
- **Repetitive Retry Loops**:
  - 3 or more consecutive tool calls attempting to modify the same file or re-running slightly altered commands to fix an immediate failure.
  - Repeated `grep` or file search calls due to broad or unhelpful queries.
- **Platform Incompatibilities**:
  - macOS vs Linux CLI flag differences (e.g., `sed -i`, `xargs`, `find`, `stat`).
- **Recurring Sandbox / Harness Denials**:
  - Repeated boundary blocks on network, filesystem, or IPC operations across turns or sessions.

### Extraction & Remediation Formulation
1. Group retries into single incident clusters: `[Root Cause] -> [Trial Actions] -> [Working Solution]`.
2. Propose documentation additions:
   - For recurring command gotchas: Add to `skills/<name>/references/troubleshooting.md`.
   - For platform differences: Add working examples to `skills/<name>/references/examples.md`.
   - For recurring sandbox denials: Classify via the violation triage rubric in [persistence-and-escape-vectors.md](../../sandbox-review/references/persistence-and-escape-vectors.md) (../../sandbox-review/references/persistence-and-escape-vectors.md) to propose narrow `(command, path)` noise suppression, safe alternative snippets, or a [sandbox-review](../../sandbox-review/SKILL.md) (../../sandbox-review/SKILL.md) follow-up.

---

## 4. Signal 3: Deviation Signals (Workflow Step Skips)

### Definition
Instances where an active skill was intended to guide the workflow, but the agent skipped mandatory checkpoints, bypassed verification steps, or had to invent undocumented steps.

### Detection Heuristics
- **Skipped Verification**:
  - Agent performed a file modification or executed changes without running the skill's explicit `Verify:` command.
- **Undocumented Ad-Hoc Interventions**:
  - The agent ran 3+ consecutive tools that are not mentioned or anticipated in the active skill's step instructions.
- **Gate Disregard**:
  - The agent executed actions that required an explicit user approval gate without asking.

### Extraction & Skill Patching
1. Identify the discrepancy between the SKILL.md definition and the agent's actual trajectory.
2. Formulate concrete edits for `skills/<name>/SKILL.md`:
   - Add missing prerequisite checks.
   - Adjust unrealistic verification steps.
   - Clarify ambiguous step instructions.

---

## 5. Signal 4: Repetition Signals (Recurring Patterns)

### Definition
Identical or structurally similar actions executed repeatedly across turns or sessions that consume manual effort, token bandwidth, or round-trip turns.

### Detection Heuristics
- **Command Sequence N-Grams**:
  - Exact or parameterized shell pipelines executed 2+ times across sessions (e.g., complex `git`, `docker`, `jq`, or build invocations).
- **Repetitive User Prompts**:
  - Similar prompt structures initiating tasks (e.g., "build, test, lint, and report diff").
- **Multi-Step Workflows**:
  - Repeating sequences involving 3+ distinct tools (e.g., explore -> diff -> test -> commit) that lack a dedicated skill.

### Extraction & Automation Formulation
1. **Command Snippets**:
   - Extract raw command string, replace volatile values with CLI flags or arguments, and draft a snippet file following `command-snippet`.
2. **New Skills**:
   - If the sequence involves verification, multi-step branching, or user approval gates, outline a candidate skill structure for `skill-creator`.
3. **Routine Prompts**:
   - Draft a reusable user prompt or slash command specification.
