# Evidentiary Inquiry References

This document provides premise clarification frameworks, source scoping templates, version analysis criteria, fact-checking checklists, and reporting formats for agents conducting evidence-backed investigations under `evidentiary-inquiry`.

---

## 1. Premise Clarification & Information Source Scoping

### Six Premise Analysis Dimensions

When a user request is ambiguous or underspecified, analyze what must be concretized before starting the investigation. Focus only on decision-altering dimensions and propose workspace-derived default hypotheses where possible:

| Dimension | Key Questions to Derive | Example Default Hypothesis |
|---|---|---|
| **1. Architecture & Environment** | What system topology, runtime environment, or component boundary is assumed? | "Assuming single-binary CLI execution in the existing repository layout." |
| **2. Functional Requirements** | What specific behaviors, inputs, outputs, or edge cases must be supported? | "Assuming read-only validation of YAML manifests with exit code reporting." |
| **3. Non-Functional Requirements** | What latency, throughput, memory, security, availability, or maintainability targets apply? | "Assuming sub-second local execution without external network calls." |
| **4. Out-of-Scope Boundaries** | What features, platforms, or migration paths are explicitly excluded from this inquiry? | "Treating automatic schema migration and legacy v1 format support as out of scope." |
| **5. Inviolable Constraints** | What existing APIs, backward compatibility guarantees, security rules, or licenses must never be violated? | "Must not break existing CLI flags or introduce non-MIT/Apache-2.0 dependencies." |
| **6. Target Versions** | Which exact version(s) of the language, library, tool, or runtime are targeted, and are upgrades allowed? | "Targeting version `v2.4.x` pinned in the lockfile (minor patch upgrades allowed, major v3 excluded)." |

### Combined Premise & Source Scoping Template

When initiating an inquiry in Step 1, propose premise clarifications and anticipated sources together:

```markdown
### 🔍 Premise Confirmation & Proposed Investigation Sources

Before beginning the inquiry, please confirm or adjust the following premises and source scope:

#### 1. Proposed Premises & Constraints
- **Architecture & Environment**: `<assumed architecture/runtime, or question>`
- **Functional & Non-Functional Scope**: `<core requirements and performance/security expectations>`
- **Out of Scope**: `<items explicitly excluded from investigation>`
- **Inviolable Constraints**: `<guardrails and compatibility conditions that must not be violated>`
- **Target Version(s)**: `<pinned or assumed version baseline, e.g., v2.4.1>`

#### 2. Anticipated Information Sources
- **Priority Sources**:
  - [file.ext](<accessible-file-target>) (path/to/primary/file.ext) — Rationale: Core implementation or manifest defining X
  - `<tool command>` (e.g., `git log -n 5`, `go test ./...`) — Rationale: Direct observation of runtime behavior or history
- **Secondary / Fallback Sources**:
  - [docs/](<accessible-dir-target>) (path/to/secondary/docs/) — Rationale: Consulted only if priority sources lack sufficient context

**Questions for User**:
1. Do the proposed premises, out-of-scope boundaries, constraints, and target versions match your expectations?
2. Are any of the proposed sources excessive, out-of-scope, or missing?
```

### Heuristics for Premise Clarification & Source Exclusions
- **Hypothesis-First Minimalism**: Do not interrogate the user with an exhaustive questionnaire when workspace manifests or configs already answer the question; state the inferred baseline and ask only where genuine ambiguity remains.
- **Strict Prohibition**: Sources explicitly labeled by the user as excessive, confidential, or irrelevant must not be opened or queried. Inviolable constraints confirmed by the user must never be breached in recommendations.
- **Priority Bias**: Exhaust priority sources before accessing secondary sources.
- **Dynamic Expansion**: If an unanticipated lead appears in the evidence, you may inspect it provided it does not violate explicit user exclusions; mention the newly discovered lead in the final report.

---

## 2. Version Contextualization & Neighborhood Analysis Guide

When investigating software libraries, tools, runtimes, or external systems as implementation means, evaluate facts across a defined version window rather than treating documentation as timeless.

### 1. Version Pinning Rules
- **Explicit Version Attribution**: Every technical claim (e.g., "Feature X is supported", "Flag `--foo` defaults to `true`") must record the exact version tag, commit hash, or release date from which the evidence was drawn.
- **Anchor vs. Neighborhood Window**:
  - **Anchor Version**: The version currently used in the workspace (from lockfiles/manifests) or specified by the user.
  - **Neighborhood Versions**: Adjacent releases around the anchor (e.g., previous minor/patch release, next minor/major release, and latest stable release).

### 2. Feature Delta Classification
Track how relevant capabilities evolve across the anchor and neighborhood versions:
- **Added**: Version in which the capability or API was introduced (confirming whether the anchor version already has it or requires an upgrade).
- **Changed / Deprecated**: Changes in default parameters, return semantics, performance characteristics, or deprecation warnings signaling future removal.
- **Removed / Breaking**: Capabilities or flags deleted or altered incompatibly, along with required migration paths.

### 3. Surrounding Issues Analysis (Unresolved & Recently Resolved)
Inspect changelogs, release notes, commit histories, and issue trackers around the target version for two critical classes of defects:

- **Unresolved Issues (Open Bugs & Limitations)**:
  - Active defects, edge-case failures, or architectural limitations present in the target version.
  - Identify whether a viable workaround exists and whether the issue violates the user's functional/non-functional requirements or inviolable constraints.
- **Recently Resolved Issues (Newly Fixed Bugs & Patch Boundaries)**:
  - Defects resolved in the target version or in immediately adjacent patch/minor versions.
  - Evaluate three specific risks:
    1. **Patch Boundary Gap**: Is the fix present in `v1.4.2` while the user's anchor version is pinned to `v1.4.0`?
    2. **Fix Side Effects**: Did the resolution introduce subtle behavioral changes, stricter validation, or performance trade-offs?
    3. **Incomplete Fix / Regression Risk**: Was the fix recently merged with follow-up edge-case reports or regression hazards?

---

## 3. Pre-Report Fact & Version Verification Checklist

Before including any fact in an investigation report, confirm each item on this checklist:

- [ ] **Premise & Constraint Alignment**: Does the investigation respect all confirmed functional/non-functional requirements, out-of-scope boundaries, and inviolable constraints?
- [ ] **Exact Output Check**: Did you directly observe the claim in raw tool output or file contents during this session, rather than recalling from general knowledge or assumption?
- [ ] **Version Pinning & Non-Conflation**: Is every software/tool fact explicitly tied to a specific version or commit hash, ensuring behaviors from older or newer versions are not conflated?
- [ ] **Neighborhood Delta & Issue Coverage**: Have feature changes (added/removed/changed), unresolved issues, and recently resolved issues around the target version been checked?
- [ ] **Exit Code & Error Silence**: Did the tool command exit successfully (`0`) without hidden errors or truncated outputs?
- [ ] **Citation Precision & Dual Presentation**: Are file paths, line ranges, commit hashes, and URLs exact and formatted with both a clickable Markdown link and an unadorned raw string in parentheses per [`output-presentation`](../../output-presentation/SKILL.md)?
- [ ] **Non-Contradiction**: Are there conflicting outputs or opposing data points in other parts of the codebase or release history? If yes, is the discrepancy noted?
- [ ] **Reproducibility**: If relying on a command output, can the same output be consistently obtained under the same parameters?

---

## 4. Evidentiary Report Template

Format the final report in Step 4 to ensure clear separation between confirmed premises, version-pinned observations, and logical inference:

```markdown
### 🔎 Evidentiary Investigation Report: <Topic / Problem Statement>

#### 1. Confirmed Premises & Constraints
- **Architecture & Environment**: `<confirmed architecture and runtime context>`
- **Functional & Non-Functional Requirements**: `<confirmed requirements>`
- **Out-of-Scope Boundaries**: `<confirmed exclusions>`
- **Inviolable Constraints**: `<confirmed guardrails>`
- **Target Version Baseline**: `<confirmed anchor version(s)>`

#### 2. Scope & Consulted Sources
- **Agreed Priority Sources**: `<list of files/tools inspected with dual presentation>`
- **Excluded Sources**: `<sources explicitly bypassed per user direction>`
- **Unanticipated Sources Consulted**: `<any additional source consulted with justification, or "None">`

#### 3. Verified Facts (Version-Pinned Observations)
> *Note: All items below are verified directly against workspace files, release artifacts, and tool outputs.*

1. **<Fact Summary 1>**
   - **Version / Commit**: `<version-tag / commit-hash>` ([Commit Link](<remote-permalink-url>) (<remote-permalink-url>))
   - **Location**: [file.ext#L20-L35](<accessible-file-target>) (path/to/file.ext:20-35)
   - **Observed Evidence**: `<Snippet or precise output summary showing the fact>`
2. **<Fact Summary 2>**
   - **Version / Command**: `<vX.Y.Z>` / [config.json#L12](<accessible-file-target>) (path/to/config.json:12)
   - **Observed Evidence**: `<Key value is set to false>`

#### 4. Version Deltas & Surrounding Issues (When Evaluating Software/Tools)

| Version | Feature / Behavior Delta (Added / Changed / Removed) | Unresolved Issues (Open Bugs / Workarounds) | Recently Resolved Issues (Patch Boundary / Side Effects) | Primary Evidence |
|---|---|---|---|---|
| `<vX.Y-1>` (Prior) | `<Baseline behavior>` | `<Issue #101>` | — | `<Release tag / file>` |
| `<vX.Y.0>` (Target) | `<Added feature Z; Changed default W>` | `<Open Issue #142: workaround required>` | `<Fixed Issue #138 in vX.Y.1 (missing in .0)>` | `<Changelog / Commit>` |
| `<vX.Y+1>` (Newer) | `<Deprecated flag V>` | `<None observed>` | `<Resolved #142; altered error return type>` | `<Release Note>` |

#### 5. Interpretation & Logical Deduction (Reasoning)
- **Bridge to Interpretation**:
  - Fact 1 shows `<X>` at version `<vX.Y.0>`, while Fact 2 and the version delta matrix establish `<Y>`.
  - Given confirmed constraint `<Z>`, this implies `<Mechanism / Cause / Viability>`.
- **Rejected Alternatives**:
  - Alternative hypothesis or version `<H1>` was ruled out because `<specific verified fact or unresolved issue disproving it>`.

#### 6. Conclusions & Recommended Action
- **Conclusion**: `<Direct answer to the user's inquiry, confirmed root cause, or version recommendation>`
- **Recommended Next Skill**:
  - Propose [`refactor`](../../refactor/SKILL.md) / [`step-gate`](../../step-gate/SKILL.md) for code modifications, [`doc-refactor`](../../doc-refactor/SKILL.md) for documentation updates, or [`explore-plan-execute`](../../explore-plan-execute/SKILL.md) / [`scrutinize-plan-execute`](../../scrutinize-plan-execute/SKILL.md) for full implementation planning.
```
