# Retrospective Report Template

Use this template to present the findings and actionable proposals at the conclusion of an `agent-retrospective` analysis.

---

## Retrospective Report Template

```markdown
# Agent Retrospective Report

**Execution Mode**: Focus: `<all | auto | refine | rules>` × Scope: `<session | recent [Nd] | custom>`  
**Analyzed Scope**: `<N>` sessions / `<N>` turns (`<Date / Session IDs>`)  
**Status**: Ready for User Review (Read-Only)

---

## 1. Executive Summary

- **Total Signals Detected**: `<N>` (Corrections: `<N>`, Friction: `<N>`, Deviations: `<N>`, Repetitions: `<N>`)
- **Key Bottlenecks**: <Brief 1-2 sentence description of the primary sources of friction or inefficiency observed>
- **High-Impact Opportunities**: <Brief summary of the most valuable snippet, rule, or skill updates>

---

## 2. Findings Matrix

| ID | Signal Type | Source / Turn | Summary of Observation | Proposed Action | Impact |
|---|---|---|---|---|---|
| F-01 | Correction (S1) | Turn `<idx>` / `<session>` | User corrected tool usage: "<quote>" | Add rule to `AGENTS.md` | High |
| F-02 | Friction (S2) | Turn `<idx>` / `<session>` | Command `<cmd>` failed with `<error>` | Add entry to `troubleshooting.md` | Medium |
| F-03 | Deviation (S3) | Turn `<idx>` / `<session>` | Skipped verification step in `<skill>` | Clarify Step `<N>` in `SKILL.md` | Medium |
| F-04 | Repetition (S4) | `<N>` occurrences | Repeated pipeline: `<command chain>` | Parameterize with `command-snippet` | High |

---

## 3. Prioritized Improvement Proposals

### Category A: Workflow Automation Proposals (Snippets & New Skills)

#### [A-01] Command Snippet: `<proposed-snippet-name>`
- **Observed Frequency**: Executed `<N>` times across `<N>` turns.
- **Draft Implementation**:
  \`\`\`bash
  #!/usr/bin/env bash
  # Description: <Brief purpose>
  # Usage: <snippet-name> <arg1> [arg2]
  set -euo pipefail
  <parameterized_command_chain>
  \`\`\`
- **Recommended Hand-off**: [command-snippet](skills/command-snippet/SKILL.md) (skills/command-snippet/SKILL.md)

---

### Category B: Existing Asset Refinements (Skills & References)

#### [B-01] Skill Improvement: `skills/<name>/SKILL.md`
- **Observed Friction**: <Description of confusion, failure, or skipped step>
- **Proposed Diff / Addition**:
  \`\`\`markdown
  ## Step <N>: <Step Title>
  + - Verify: <concrete verification command to prevent failure>
  \`\`\`

#### [B-02] Troubleshooting Entry: `skills/<name>/references/troubleshooting.md`
- **Error Pattern**: `<error output or symptom>`
- **Root Cause & Fix**:
  \`\`\`markdown
  ### Issue: <Title>
  **Symptom**: `<error text>`
  **Resolution**: Run `<working command>` instead of `<broken command>`.
  \`\`\`

---

### Category C: Governance & Rule Additions (AGENTS.md & Rules)

#### [C-01] Repository Rule: `<Rule Title>`
- **Observed User Correction**: "<Quote user instruction>" (Turn `<idx>`)
- **Draft Rule Text for `AGENTS.md`**:
  \`\`\`markdown
  - **<Rule Title>**:
    <Clear, imperative rule instructing the agent on the required behavior or forbidden action.>
  \`\`\`

---

## 4. User Action Gate

Please select which recommendations to adopt:
- [ ] Adopt **A-01** (Generate command snippet via `command-snippet`)
- [ ] Adopt **B-01** (Patch skill step instructions via `doc-refactor`)
- [ ] Adopt **B-02** (Add troubleshooting entry via `doc-refactor`)
- [ ] Adopt **C-01** (Append rule to `AGENTS.md`)
- [ ] Reject all / Adjust proposals
```
