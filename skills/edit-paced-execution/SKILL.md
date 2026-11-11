---
name: edit-paced-execution
description: >-
  Use this skill when task execution risks long investigation loops or silent delays
  between file modifications. Sets dynamic time and non-edit tool call thresholds,
  pauses when thresholds are exceeded or obstacles arise, reports progress and blockers,
  and requests user direction before continuing.
---

# Edit Paced Execution

Use this skill to maintain tight user feedback loops and prevent runaway exploration or silent stalls during implementation work.
The agent paces execution by monitoring the interval between file edits, pausing proactively when work exceeds proposed thresholds or hits unexpected delays, and requesting user direction.

For reporting templates and threshold heuristics, refer to [pace-monitoring-guide.md](references/pace-monitoring-guide.md) (references/pace-monitoring-guide.md).

---

## Step 1: Assess Work Complexity and Propose Thresholds

Before beginning task execution, analyze the task requirements and propose pacing thresholds and execution modes to the user.

1. **Establish Baseline Edit-Pacing Thresholds**:
   - Default: **3 minutes** or **20 non-edit tool calls / 6–8 non-edit reasoning turns** (commands, searches, reads, test runs without modifying project files).
   - **Parallel Tool-Call Accounting**: When a model issues multiple read-only tool calls (such as parallel `read_file` or `grep` calls) within a single reasoning turn, count that parallel batch as **1 non-edit turn** (or weight it proportionally) so efficient parallel exploration does not prematurely trigger a stall checkpoint.
2. **Determine Task & Command Execution Mode**:
   - When executing commands or background tasks, determine their execution mode based on duration, retryability, and external volatility:
     - **`standard`**: Routine tasks; easy & fast to retry (e.g. unit tests, linters). Default check interval: **3 minutes**.
     - **`heavy`**: Long-running tasks, but cleanly retryable without special conditions (e.g. integration tests, release builds). Default check interval: **5 minutes**.
     - **`critical`**: Sensitive tasks affected by uncontrollable conditions (network, API limits) or hard to retry (e.g. deployments, remote API batches, migrations). Default check interval: **1 minute**.
     - **`unmonitored`**: Tasks known in advance to take very long where frequent checks are disruptive (e.g. model training, giant clones). Check interval: **60 minutes or none**.
3. **Context-Aware Adjustment**:
   - Tailor thresholds based on work scope (see [pace-monitoring-guide.md](references/pace-monitoring-guide.md) (references/pace-monitoring-guide.md)):
     - Small/localized edits: 1–2 minutes or 5–10 non-edit tool calls.
     - Standard tasks: 3 minutes or 20 non-edit tool calls.
     - Broad cross-cutting exploration: 4–5 minutes or 25–30 non-edit tool calls.
4. **Present Proposal**:
   - Explicitly communicate proposed pacing limits, command execution mode, and rationale before executing actions:
     > *"For this task, I propose an edit-pacing threshold of X minutes or Y non-edit tool calls. Long-running commands will run in `mode` (check interval: Z minutes). If I encounter unexpected delays, stalls, or exceed limits, I will pause and report for guidance."*

Verify: Pacing thresholds and command execution modes are proposed with clear rationale and acknowledged or adjusted by the user.

---

## Step 2: Track Progress and Pace

Execute the planned task while continuously tracking pace against the agreed thresholds.

1. **Define File Modifications**:
   - Creating, editing, replacing, or patching repository code and documentation files counts as an edit event.
   - Reading files, running read-only search commands, checking status, or testing without edits do **not** reset the interval.
2. **Interval Reset**:
   - Each time a target file is modified, reset the non-edit tool call counter and elapsed time baseline.
3. **Pace Monitoring**:
   - Maintain awareness of continuous non-edit operations.
   - If an exploratory command or troubleshooting sequence is anticipated to take significant time, evaluate whether an early checkpoint is warranted before running it.
4. **Command Liveness Tracking**:
   - For long-running or background commands, check execution progress at the mode's interval (`standard`: 3m, `heavy`: 5m, `critical`: 1m, `unmonitored`: 60m).
   - Evaluate liveness signals (CPU/memory consumption, file/artifact changes, and log activity).

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
   - Reset the tool call counter and timer.
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
