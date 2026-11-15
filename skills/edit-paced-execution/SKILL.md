---
name: edit-paced-execution
description: >-
  Use this skill when task execution risks long investigation loops or silent delays
  between file modifications. Sets dynamic time and non-edit tool call thresholds,
  pauses when thresholds are exceeded or obstacles arise, reports progress and blockers,
  and requests user direction before continuing.
license: MIT
---

# Edit Paced Execution

Use this skill to maintain tight user feedback loops and prevent runaway exploration or silent stalls during implementation work.
The agent paces execution by monitoring the interval between file edits, pausing proactively when work exceeds proposed thresholds or hits unexpected delays, and requesting user direction.

For reporting templates and threshold heuristics, refer to [pace-monitoring-guide.md](references/pace-monitoring-guide.md) (references/pace-monitoring-guide.md).

---

## Prerequisites

- Access to the companion MCP server (`agent-skills-mcp`) providing the `pace-guard` toolset (`pace_start`, `pace_tick`, `pace_status`, `pace_reset`, `pace_watch`, `pace_liveness`). If the MCP server is unavailable, refer to [pace-monitoring-guide.md](references/pace-monitoring-guide.md) (references/pace-monitoring-guide.md) as the manual pacing and stall reporting fallback.

---

## Step 1: Assess Work Complexity and Propose Thresholds

Before beginning task execution, analyze the task requirements and propose pacing thresholds and execution modes to the user.

1. **Initialize Session**: Call the `pace_start` MCP tool with the desired `preset` (`localized`, `standard`, `cross_cutting`), `time_limit_sec`, `call_limit`, and `command_mode` (`standard`, `heavy`, `critical`, `unmonitored`). The tool returns the structured proposal block.
2. **Present Proposal**: Present the proposed limits and rationale to the user for acknowledgment.

If the MCP server is unavailable, refer to [pace-monitoring-guide.md](references/pace-monitoring-guide.md) (references/pace-monitoring-guide.md) to establish thresholds manually:
- Baseline: **3 minutes** or **20 non-edit tool calls / 6–8 reasoning turns** (batch parallel calls as 1 turn).
- Command modes: `standard` (3m check), `heavy` (5m check), `critical` (1m check), `unmonitored` (60m).

Verify: Pacing thresholds and command execution modes are proposed with clear rationale and acknowledged or adjusted by the user.

---

## Step 2: Track Progress and Pace

Track pace and command liveness using the MCP tools:

- **Interval Tracking**: On each tool call, invoke `pace_tick` with `kind: "edit"`, `"non_edit"`, or `"parallel_batch"`. When a repository file is modified (`kind: "edit"`), the server resets the non-edit counter and elapsed timer automatically.
- **Status Checks**: Call `pace_status` to query elapsed seconds and remaining call allowances without modifying counters.
- **Process and Log Liveness**: Register long-running commands or background tasks via `pace_watch` (passing `pid`, `log_path`, or `artifact_path`) and sample execution health with `pace_liveness`.

If the MCP server is unavailable, follow the manual tracking procedure in [pace-monitoring-guide.md](references/pace-monitoring-guide.md) (references/pace-monitoring-guide.md):
- Reset non-edit counters and time baselines upon file modifications.
- Track non-edit queries and elapsed wall-clock time manually.
- Sample command liveness (CPU, memory, log activity) at the mode's interval.

Verify: Tool execution count, time elapsed since the last file modification, and command execution liveness are tracked accurately.

---

## Step 3: Trigger Checkpoint Halt and Report

Pause immediately and report to the user if ANY of the following occur before the next file edit:

- Elapsed wall-clock time reaches the agreed time threshold.
- Accumulated non-edit tool calls reach the agreed limit.
- Investigation stalls in cyclic failure loops (e.g. repeated test failures or dead-end search queries).
- A monitored command or task appears stalled (0% CPU, no log activity, or frozen artifacts across check intervals).
- A blocking ambiguity or architectural decision point emerges.

### Checkpoint Report Procedure:
1. Stop all further background or exploratory commands.
2. Structure the report using the templates from [pace-monitoring-guide.md](references/pace-monitoring-guide.md) (references/pace-monitoring-guide.md):
   - For edit pacing delays, use the **Paced Execution Checkpoint** template (Current Status, Reason for Delay, Options, Question).
   - For stalled or frozen commands, use the **Command Execution Stall Report** template (Observed Liveness Signals, Diagnostic Assessment, Action Options: Abort vs. Keep Waiting, User Judgment).

Verify: Execution is halted, findings/signals and blocker reasons are clearly articulated, and user judgment (e.g. abort vs. continue waiting) is requested without taking further autonomous actions.

---

## Step 4: Resume Execution on User Direction

Upon receiving user input:

1. **Adopt User Decision**:
   - Adjust strategy according to the selected option or user-supplied instructions.
   - If the user requested adjusting thresholds (e.g., granting more time or tightening limits), apply the new thresholds.
2. **Reset Tracking**:
   - Reset intervals and counters via the `pace_reset` MCP tool (updating thresholds if requested). If the MCP server is unavailable, reset the counter and timer manually.
3. **Proceed to Target Edit**:
   - Execute the agreed remediation or next step directly toward the intended file edit.
   - Continue the monitoring loop from Step 2.

Verify: The agent acts strictly on the approved option, tracking counters are reset, and execution proceeds towards the next milestone.

---

## Guidelines

- **Pause proactively when thresholds are reached**: Halt and consult the user (alerting via [user-notification](../user-notification/SKILL.md) (../user-notification/SKILL.md)) when an investigation does not yield progress toward a file edit within the agreed time or turn threshold, preventing silent runaway loops.
- **Propose task-adapted thresholds**: Suggest time and non-edit turn limits tailored to the task's complexity rather than applying rigid static numbers without rationale.
- **Treat file edits as concrete milestones**: Reset pacing counters only when project code or documentation files are modified so intermediate reads and logs do not mask stalls.
- **Surface blockers transparently**: State what was tried and why progress slowed when halting so the user has full context to choose the next path.
- **Monitor command liveness actively**: Inspect CPU/memory usage, log activity, and artifact changes at the mode's interval rather than waiting indefinitely, and seek user judgment on aborting vs. waiting when a command stalls.
- **Hand off plan restructuring sequentially**: Recommend [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md) or [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md) (and log checkpoints via [checkpoint-log](../checkpoint-log/SKILL.md) (../checkpoint-log/SKILL.md)) when a stall reveals the need for a revised plan.
