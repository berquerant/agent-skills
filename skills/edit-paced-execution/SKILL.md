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

For reporting templates and threshold heuristics, refer to [references/pace-monitoring-guide.md](references/pace-monitoring-guide.md).

---

## Step 1: Assess Work Complexity and Propose Thresholds

Before beginning task execution, analyze the task requirements and propose pacing thresholds to the user.

1. **Establish Baseline Thresholds**:
   - Default: **3 minutes** or **20 non-edit tool calls** (commands, searches, reads, test runs without modifying project files).
2. **Context-Aware Adjustment**:
   - Tailor the threshold based on the work scope (see [references/pace-monitoring-guide.md](references/pace-monitoring-guide.md)):
     - Small/localized edits: 1–2 minutes or 5–10 non-edit tool calls.
     - Standard tasks: 3 minutes or 20 non-edit tool calls.
     - Broad cross-cutting exploration: 4–5 minutes or 25–30 non-edit tool calls.
3. **Present Proposal**:
   - Explicitly communicate the proposed limits and rationale before executing actions:
     > *"For this task, I propose an edit-pacing threshold of X minutes or Y non-edit tool calls between file modifications. If I encounter unexpected delays or exceed this limit, I will pause and report for guidance."*

Verify: Pacing thresholds are proposed with clear rationale and acknowledged or adjusted by the user.

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

Verify: Tool execution count and time elapsed since the last file modification are tracked accurately.

---

## Step 3: Trigger Checkpoint Halt and Report

Pause immediately and report to the user if ANY of the following occur before the next file edit:

- Elapsed wall-clock time reaches the agreed time threshold.
- Accumulated non-edit tool calls reach the agreed limit.
- Investigation stalls in cyclic failure loops (e.g. repeated test failures or dead-end search queries).
- A blocking ambiguity or architectural decision point emerges.

### Checkpoint Report Procedure:
1. Stop all further background or exploratory commands.
2. Structure the report using the template from [references/pace-monitoring-guide.md](references/pace-monitoring-guide.md):
   - **Current Status**: Actions taken, files examined, findings discovered.
   - **Reason for Delay**: Specific obstacle, ambiguity, or why the file edit has not yet occurred.
   - **Options**: 2–3 actionable choices for how to proceed (including recommended path).
   - **Question**: Direct request for user guidance.

Verify: Execution is halted, findings and blocker reasons are clearly articulated, and user judgment is requested without taking further autonomous actions.

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

- **Never spin silently.** If an investigation does not yield concrete progress towards an edit within the agreed threshold, pause and ask.
- **Propose, don't mandate.** Always suggest thresholds adapted to the task at hand rather than rigidly applying static numbers without justification.
- **Respect file edits as genuine milestones.** Intermediate reading and logging are not completions; progress is marked by concrete, verifiable changes.
- **Transparent blockers.** When halting, clearly state what was tried and why it failed or slowed down—never hide investigation friction.
- **Sequential hand-off.** When large plan restructuring is required upon halting, recommend planning skills (`step-gate` or `explore-plan-execute`) for user approval.
