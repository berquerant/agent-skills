# Workflow Coordination Matrix

Guidelines, decision signals, and plan integration templates for coordinating specialized execution, pacing, milestone logging, and task scheduling skills during the Plan phase of `explore-plan-execute`.

---

## 1. Skill Selection Matrix

When formulating the execution plan in Step 3, evaluate whether the task requires any of the following specialized workflow skills:

| Target Skill | Primary Purpose | Trigger Signals | Key Prerequisites / Setup |
|---|---|---|---|
| [`scrutinize-plan-execute`](../../scrutinize-plan-execute/SKILL.md) | Upfront obstacle review, risk scrutiny, and interactive refinement | Ambiguous premises, unclear verification criteria, high architectural risk, multi-approach trade-offs | Delegates complete planning, archiving, and execution gates |
| [`task-dependency-pipeline`](../../task-dependency-pipeline/SKILL.md) | DAG modeling with cooperative idle-wait context switching | Multiple independent branches, long-running CI/build wait states | Separate git worktrees or resource isolation |
| [`step-gate`](../../step-gate/SKILL.md) | Reliable, gated step-by-step execution with explicit checkpoints | Low reversibility (DB, prod deploy), ≥5 sequential steps, high cost of mid-task failure | Independent verification & completion criterion per step |
| [`checkpoint-log`](../../checkpoint-log/SKILL.md) | Persistent local milestone progress and context tracking | Multi-session work, complex refactors, user requested progress log | Explicit user confirmation of log file path |
| [`edit-paced-execution`](../../edit-paced-execution/SKILL.md) | Liveness pacing and threshold monitoring between file edits | High debugging uncertainty, long external commands, risk of long silent loops | Configured timeout / non-edit tool call thresholds |

---

## 2. Decision Criteria & Plan Integration

### 2.1 Delegate to `scrutinize-plan-execute`
Delegate the entire planning and execution phase to [`scrutinize-plan-execute`](../../scrutinize-plan-execute/SKILL.md) when:
- **Ambiguous premises or conditions**: Unverified environment behavior, missing API contracts, or incomplete requirements.
- **Unclear success criteria**: Subjective goals ("make code cleaner", "improve performance") without observable pass/fail checks.
- **High-impact obstacles**: Breaking API changes, database schema alterations, or multi-component race conditions.
- **Deep user consultation needed**: Multiple viable architectures where trade-offs must be evaluated collaboratively.
- **Plan Archiving**: Extensive plans (> 300 lines) requiring iterative user sign-off and permanent disk archiving.

**Plan Declaration**:
> **Hand-off**: Delegate to [`scrutinize-plan-execute`](../../scrutinize-plan-execute/SKILL.md) for critical obstacle analysis and interactive refinement.

---

### 2.2 Adopt `task-dependency-pipeline`
Adopt [`task-dependency-pipeline`](../../task-dependency-pipeline/SKILL.md) when the task decomposes into a Directed Acyclic Graph (DAG) with concurrency or wait-state interleaving opportunities:
- **Multiple independent branches**: Work spanning decoupled modules, independent PRs, or multiple repositories.
- **Asynchronous idle wait states**: Operations triggering remote CI pipelines, large builds, or rate-limit cooldowns.
- **Cooperative interleaving**: Progressing local verification or setup on Task B while Task A waits for remote completion.

**Plan Declaration**:
> **Scheduling mode**: `task-dependency-pipeline` (cooperative idle-wait DAG execution)

---

### 2.3 Adopt `step-gate`
Adopt [`step-gate`](../../step-gate/SKILL.md) when the execution requires fine-grained user checkpoint gates between individual steps:
- **Low reversibility**: Destructive operations, data deletion, remote migrations, production pushes.
- **High failure blast-radius**: Long multi-phase refactoring where downstream steps depend strictly on upstream invariants.
- **Step count**: 5 or more sequential steps with tight dependencies.
- **User instruction**: User explicitly requested "step by step", "ask before every step", or "proceed with caution".

**Plan Declaration**:
> **Execution mode**: `step-gate` (per-step verification + user approval before continuing)

*Note: Skip `step-gate` for read-only tasks, trivial 1–2 step edits, or easily reversible changes.*

---

### 2.4 Propose `checkpoint-log`
Propose [`checkpoint-log`](../../checkpoint-log/SKILL.md) when a persistent, disk-backed milestone record provides recovery or audit benefits:
- The task spans across multiple turns or extended debugging sessions.
- Detailed architectural decisions and verification commands should be preserved locally for context recovery.

**Plan Declaration**:
> **Milestone logging**: Recommend [`checkpoint-log`](../../checkpoint-log/SKILL.md) to record progress. Please confirm your preferred log file path (e.g., `docs/checkpoints/task-progress.md`).

---

### 2.5 Propose `edit-paced-execution`
Propose [`edit-paced-execution`](../../edit-paced-execution/SKILL.md) when execution risks long periods of non-editing activity or silent loops:
- Trial-and-error exploration or troubleshooting unfamiliar toolchains before editing files.
- Executing long-running external processes or remote builds.

**Plan Declaration**:
> **Pacing guard**: Recommend [`edit-paced-execution`](../../edit-paced-execution/SKILL.md) (threshold: <X minutes / Y non-edit calls>, command mode: `standard` / `heavy` / `critical` / `unmonitored`) to monitor liveness and halt if progress stalls.
