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

## 2. Dynamic Halt Criteria

An agent must pause and report to the user if ANY of the following occur:

1. **Threshold Exceeded**: Elapsed time reaches the agreed time threshold, or the non-edit tool call count reaches the limit without modifying a project file.
2. **Investigation Stall / Deadlock**: The agent hits repetitive errors, cyclic search patterns, or unresolvable test/build failures (> 2 consecutive attempts without new insights).
3. **Unexpected Scope Expansion**: Discovering that the next edit requires touching 3+ unpredicted files, external services, or complex refactoring not in the original plan.
4. **Ambiguous Decision Point**: Reaching a fork where multiple viable architectural or design paths exist, and choosing one requires product/business assumptions.

---

## 3. Intervention Report Template

When halting due to an interval threshold or unexpected delay, format the report as follows:

```markdown
### ⏸️ Paced Execution Checkpoint

**Current Threshold**: <X minutes / Y non-edit tool calls> (Current: <elapsed or tool count>)
**Last File Edited**: `<path/to/last/file>` (or "None yet")

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

---

## 4. Re-engaging After Resumption

Upon user approval to continue:
1. Reset the non-edit tool call counter and timer to zero.
2. Note any refined directions or adjusted thresholds given by the user.
3. Proceed directly toward the target file edit without repeating previously cached investigations.
