# Edit Paced Execution References

This document provides reporting templates, pace monitoring heuristics, and troubleshooting patterns for agents operating under `edit-paced-execution`.

---

## 1. Threshold Evaluation Heuristics

Agents must evaluate work characteristics before execution and dynamically adapt monitoring thresholds.

| Work Type | Characteristic | Recommended Time Threshold | Recommended Tool Call Threshold |
|---|---|---|---|
| **Quick Refactor / Small Edits** | Direct code tweaks, localized edits, fast feedback loops | 1–2 minutes | 5–10 non-edit calls |
| **Standard Implementation** | Typical feature work, light investigation + code change (Default) | 3 minutes | 20 non-edit calls |
| **Deep Investigation / Multi-Repo** | Heavy log analysis, unfamiliar codebase exploration, distributed tracing | 5 minutes | 30 non-edit calls |

### Adjusting Thresholds
- **Pre-execution Proposal**: When starting a task, state the suggested threshold along with the rationale:
  > *"Because this task involves exploratory debugging across 4 packages, I propose a threshold of 4 minutes or 25 non-edit tool calls before reporting progress."*
- **User Override**: If the user specifies explicit limits (e.g., "1 minute", "every 5 tool calls"), honor them strictly.

---

## 2. Task & Command Execution Modes

Before executing non-trivial tasks or commands (especially background, long-running, or external commands), determine the execution mode and check interval:

| Mode | Target Task Characteristic | Retry Feasibility | Default Interval | Typical Examples |
|---|---|---|---|---|
| **`standard`** | Routine tasks. Failure does not take long to retry. | High (fast & easy retry) | **3 minutes** | Unit test runs, standard builds, lint / static analysis |
| **`heavy`** | Heavy tasks expected to take substantial time, but retryable without special conditions. | High (clean retryable) | **5 minutes** | Full integration test suites, release builds, large data processing |
| **`critical`** | Critical tasks sensitive to uncontrollable external conditions (network, API limits) or hard to retry. | Low (hard to retry / fragile) | **1 minute** | Deployments, remote API batches, external DB migrations, live services |
| **`unmonitored`** | Exceptional tasks known in advance to require very long durations, where frequent polling is disruptive. | Special | **60 minutes or none** | Deep learning training, extensive benchmark runs, giant repo full clones |

---

## 3. Liveness and Progress Monitoring

When periodically checking on running tasks or commands, evaluate whether the process is making healthy progress versus stalled, deadlocked, or orphaned:

### 3.1 Observation Signals
- **Resource Consumption**: CPU and memory utilization trends. Is CPU activity sustained, or has it dropped to 0% continuously? Is memory growing steadily toward an OOM limit?
- **Side Effects & Artifacts**: Changes to output files, build artifacts, temporary directories, or database records (e.g. file size increases, updated timestamps).
- **Log Stream Activity**: New log lines produced. Are logs advancing, completely silent for multiple minutes, or repeating identical error loops?

### 3.2 Stall & Abnormality Criteria
Classify a command or task as **potentially stalled** if:
- No log output or artifact change has occurred across multiple check intervals while CPU utilization remains at 0%.
- Process is caught in an infinite error cycle (e.g. repeated socket reconnection failures).
- Memory consumption has frozen or spiked uncontrollably.

---

## 4. Dynamic Halt Criteria

An agent must pause and report to the user if ANY of the following occur:

1. **Threshold Exceeded**: Elapsed time reaches the agreed time threshold, or the non-edit tool call count reaches the limit without modifying a project file.
2. **Investigation Stall / Deadlock**: The agent hits repetitive errors, cyclic search patterns, or unresolvable test/build failures (> 2 consecutive attempts without new insights).
3. **Execution Stall or Abnormal Resource Utilization**: A monitored command or task exhibits stall criteria (no logs, 0% CPU, missing file artifacts, or cyclic error loops).
4. **Unexpected Scope Expansion**: Discovering that the next edit requires touching 3+ unpredicted files, external services, or complex refactoring not in the original plan.
5. **Ambiguous Decision Point**: Reaching a fork where multiple viable architectural or design paths exist, and choosing one requires product/business assumptions.

---

## 5. Intervention & Stall Report Templates

### 5.1 General Paced Execution Checkpoint
When halting due to an edit-pacing interval threshold or unexpected investigation delay, format the report as follows:

```markdown
### ⏸️ Paced Execution Checkpoint

**Current Threshold**: <X minutes / Y non-edit tool calls> (Current: <elapsed or tool count>)
**Last File Edited**: <path/to/last/file> (or "None yet")

#### 1. Current Status & Findings
- <What has been examined, executed, or verified so far>
- <Key discovery or diagnostic result>

#### 2. Reason for Delay / Obstacle
- <Why the next file edit has not yet occurred>
- <Specific blocker: e.g. unexpected dependency, contradictory test failure, API documentation discrepancy>

#### 3. Proposed Next Actions
- **Option A (Recommended)**: <Primary recommended path forward and next planned edit>
- **Option B**: <Alternative workaround or pragmatic shortcut>
- **Option C**: <Abort / Rollback / Deeper architectural discussion>

#### 4. User Judgment Needed
<Direct question asking how the user wishes to proceed.>
```

### 5.2 Command / Task Stall Report
When a command or background task appears stalled or abnormal during periodic status checks:

```markdown
### ⚠️ Command Execution Stall Report

**Monitored Command / Task**: `<command line or task ID>`
**Execution Mode**: `standard` / `heavy` / `critical` / `unmonitored` (Check Interval: <X minutes>)
**Elapsed Time**: <X minutes>

#### 1. Observed Liveness Signals
- **Resource Usage**: CPU <e.g. 0% for past Y minutes>, Memory <e.g. constant Z MB>
- **Log Activity**: <e.g. No new lines for Y minutes / Repeating loop: "...">
- **File & Artifact Side Effects**: <e.g. Target artifact <path> unmodified since HH:MM>

#### 2. Diagnostic Assessment
- <Evaluation of whether the process is deadlocked, hung on network I/O, or merely computationally intensive>

#### 3. Action Options for User
- **Option A (Abort & Diagnose)**: Terminate the command/task, inspect lingering traces, and formulate an alternative approach.
- **Option B (Extend Wait)**: Continue waiting for another <X minutes> interval.
- **Option C (Manual Inspection)**: Pause for user manual intervention.

#### 4. User Judgment Needed
Do you want to terminate this command/task, or continue waiting?
```

---

## 6. Re-engaging After Resumption

Upon user approval to continue:
1. Reset the non-edit tool call counter, timer, or command monitoring interval.
2. Note any refined directions or adjusted thresholds given by the user.
3. Proceed directly toward the target file edit or command completion without repeating previously cached investigations.
