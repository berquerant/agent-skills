# Evidentiary Inquiry References

This document provides templates, fact-checking checklists, and formatting guidelines for agents conducting evidence-backed investigations under `evidentiary-inquiry`.

---

## 1. Information Source Scoping Template

When initiating an inquiry in Step 1, propose anticipated sources using the following format:

```markdown
### 🔍 Proposed Investigation Sources

Before beginning the inquiry, I plan to examine the following sources:

- **Priority Sources**:
  - `path/to/primary/file.ext` (Rationale: Core implementation or configuration defining X)
  - `<tool command>` (e.g. `git log -n 5`, `go test ./...`) (Rationale: Direct observation of runtime or history)
- **Secondary / Fallback Sources**:
  - `path/to/secondary/docs/` (Rationale: Consulted only if priority sources lack sufficient context)

**Questions for User**:
1. Are any of these sources excessive or out-of-scope?
2. Are there specific sources missing that should be prioritized?
```

### Heuristics for Source Exclusions
- **Strict Prohibition**: Sources explicitly labeled by the user as excessive, confidential, or irrelevant must not be opened or queried.
- **Priority Bias**: Exhaust priority sources before accessing secondary sources.
- **Dynamic Expansion**: If an unanticipated lead appears in the evidence, you may inspect it provided it does not violate explicit user exclusions; mention the newly discovered lead in the final report.

---

## 2. Pre-Report Fact Verification Checklist

Before including any fact in an investigation report, confirm each item on this checklist:

- [ ] **Exact Output Check**: Did you directly observe the claim in raw tool output or file contents during this session, rather than recalling from general knowledge or assumption?
- [ ] **Exit Code & Error Silence**: Did the tool command exit successfully (`0`) without hidden errors or truncated outputs?
- [ ] **Citation Precision**: Is the file path, function/symbol name, and line number cited exact and currently up-to-date in the workspace?
- [ ] **Non-Contradiction**: Are there conflicting outputs or opposing data points in other parts of the codebase? If yes, is the discrepancy noted?
- [ ] **Reproducibility**: If relying on a command output, can the same output be consistently obtained under the same parameters?

---

## 3. Evidentiary Report Template

Format the final report in Step 4 to ensure clear separation between observation and inference:

```markdown
### 🔎 Evidentiary Investigation Report: <Topic / Problem Statement>

#### 1. Scope & Consulted Sources
- **Agreed Priority Sources**: `<list of files/tools inspected>`
- **Excluded Sources**: `<sources explicitly bypassed per user direction>`
- **Unanticipated Sources Consulted**: `<any additional source consulted with justification, or "None">`

#### 2. Verified Facts (Observations)
> *Note: All items below are verified directly against workspace files and tool outputs.*

1. **<Fact Summary 1>**
   - **Source / Command**: `git log -n 1` or `path/to/file.ext#L20-L35`
   - **Observed Evidence**: `<Snippet or precise output summary showing the fact>`
2. **<Fact Summary 2>**
   - **Source / Command**: `path/to/config.json#L12`
   - **Observed Evidence**: `<Key value is set to false>`

#### 3. Interpretation & Logical Deduction (Reasoning)
- **Bridge to Interpretation**:
  - Fact 1 shows `<X>`, while Fact 2 establishes `<Y>`.
  - Given system invariant `<Z>`, this implies `<Mechanism/Cause>`.
- **Rejected Alternatives**:
  - Alternative hypothesis `<H1>` was ruled out because `<specific verified fact disproving it>`.

#### 4. Conclusions & Recommended Action
- **Conclusion**: `<Direct answer to the user's inquiry or confirmed root cause>`
- **Recommended Next Skill**:
  - Propose [`refactor`](../refactor/SKILL.md) / [`step-gate`](../step-gate/SKILL.md) for code modifications, or [`explore-plan-execute`](../explore-plan-execute/SKILL.md) for full implementation planning.
```
