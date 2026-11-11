---
name: evidentiary-inquiry
description: >-
  Use this skill when an investigation, technical evaluation, or diagnostic task requires
  rigorous evidence, premise clarification, and version-pinned fact verification. Scopes
  trusted sources, separates observed facts from logical deductions, and prevents
  speculative diagnoses.
---

# Evidentiary Inquiry

Use this skill when conducting investigations, root cause analyses, technical evaluations, or fact-finding tasks where precision, traceability, version accuracy, and verifiable evidence are required.
Before exploring, the agent analyzes ambiguous requests to clarify core premises with the user and agrees upon information sources. During investigation, it pins every fact to an exact version, examines feature deltas and unresolved or recently resolved issues across surrounding versions, verifies all findings before presenting them, clearly separates observed facts from inferences, and explains the logical deduction behind every judgment.

For premise clarification checklists, source scoping formats, version delta matrices, and reporting templates, refer to [source-scoping-and-verification.md](references/source-scoping-and-verification.md) (references/source-scoping-and-verification.md).

---

## Step 1: Clarify Premises and Scope Information Sources

Before executing exploratory commands or reading target files, analyze the user's request to identify ambiguities and predict which information sources will be needed.

1. **Analyze Request and Clarify Ambiguous Premises**:
   - **Local Manifest Fast-Path**: Before pausing to ask the user about target versions, architecture, or toolchain constraints, check whether local repository manifests (such as `go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`, or `README.md`) already answer those objective facts unambiguously. Only pause to ask the user when genuine ambiguity, unstated requirements, or external source authorization remains.
   - If the user's request contains ambiguities that could misdirect the investigation, determine which points must be concretized to ensure a reliable inquiry.
   - Evaluate the request across six premise dimensions (proposing workspace-derived default hypotheses where possible to minimize user burden):
     - **Architecture & Environment**: Expected system architecture, runtime topology, and integration boundaries.
     - **Functional Requirements**: Core capabilities, behaviors, and inputs/outputs to achieve or verify.
     - **Non-Functional Requirements**: Performance, latency, throughput, security, availability, or maintainability targets.
     - **Out-of-Scope Boundaries**: Explicit exclusions, non-goals, or scenarios that do not need investigation.
     - **Inviolable Constraints**: Non-negotiable guardrails, backward compatibility requirements, prohibited modifications, or regulatory/licensing limits.
     - **Target Versions**: Expected versions of software, libraries, runtimes, or protocols (and whether version upgrades are permitted).
2. **Forecast and Classify Candidate Sources**:
   - List the information sources anticipated for the inquiry:
     - Specific repository files and directories (e.g., config files, manifests, lockfiles, implementation modules).
     - Local tools, linters, or test runners (e.g., `git`, test suites, static analysis tools).
     - Version history and issue sources (e.g., release notes, changelogs, migration guides, commit logs, issue trackers) when evaluating software or tooling capabilities.
     - Documentation or spec files within local references.
   - Categorize the candidate sources:
     - **Primary / Priority Sources**: First-line sources essential to resolving the inquiry.
     - **Secondary / Fallback Sources**: Used only if primary sources yield incomplete evidence.
   - Present the clarifying questions (with proposed default assumptions) alongside the anticipated sources in a single prompt:
     > *"To ensure an accurate investigation, please confirm or adjust these premises: [List key premise questions/hypotheses]. Based on these, I plan to examine: [List priority and fallback sources]. Are any premises off-target or any sources excessive/missing?"*
3. **Establish Premise and Source Constraints**:
   - Lock in the confirmed requirements, out-of-scope boundaries, inviolable constraints, and target versions as baseline invariants for the investigation.
   - Prioritize sources agreed upon with the user.
   - You are not strictly forbidden from consulting additional unpredicted sources if genuine inquiry demands it; however, **sources designated as excessive or out-of-scope by the user must be strictly avoided**.

Verify: Ambiguous premises (architecture, functional/non-functional requirements, out-of-scope items, inviolable constraints, target versions) are clarified, candidate information sources are categorized into priority/fallback tiers, and explicit user consensus on premises, scope, and exclusions is obtained.

---

## Step 2: Investigate, Contextualize Versions, and Self-Verify Evidence

Execute the investigation focusing on the agreed sources, contextualize findings by version, and verify all facts before formulating reports.

1. **Targeted Exploration**:
   - Probe the prioritized sources first while respecting confirmed constraints and out-of-scope boundaries.
   - Record exact tool commands, file paths, line ranges, and raw tool outputs.
   - For high-cost or external commands (e.g., CI logs, release metadata, issue queries, platform APIs), evaluate idempotency and prioritize using a dedicated caching or memoization skill, tool, or MCP server if available; otherwise, follow the caching and TTL protocol in [command-cache-protocol.md](../explore-plan-execute/references/command-cache-protocol.md) (../explore-plan-execute/references/command-cache-protocol.md) to preserve evidence locally while minimizing repetitive queries.
   - If investigation threatens to become prolonged or stalls, respect interval pacing or consult pacing guidelines.
2. **Version Contextualization and Neighborhood Analysis**:
   - When investigating software, libraries, frameworks, or tooling as implementation means, rigorously track temporal and version boundaries:
     - **Pin Version Context**: Explicitly distinguish which exact version (version tag, commit hash, or release date) every observed fact, behavior, or specification belongs to. Never treat version-unanchored claims as verified facts.
     - **Compare Feature Deltas Across Surrounding Versions**: Examine differences in the target version and its surrounding versions (e.g., adjacent minor/major releases and latest stable) regarding feature **additions (Added)**, **removals (Removed / Breaking)**, and **behavioral changes or deprecations (Changed / Deprecated)**.
     - **Inspect Surrounding Issues (Unresolved & Recently Resolved)**: Investigate known problems in and around the target version:
       - **Unresolved Issues**: Open bugs, known limitations, edge-case failures, and required workarounds affecting the target version.
       - **Recently Resolved Issues**: Bugs or regressions fixed in the target version or immediately adjacent releases. Pay special attention to patch-level boundaries (whether the target version includes the fix), subtle behavioral side effects introduced by the fix, and incomplete fixes or residual regression risks.
3. **Pre-Report Fact-Checking (Self-Verification)**:
   - Before presenting any fact to the user, verify its truth and reproducibility:
     - Did the command exit with status `0`, or did it fail silently?
     - Does the cited file and line range actually contain the stated code or configuration?
     - Are exact versions, commit hashes, workspace-relative paths, and line numbers rigorously identified for versioned changes and code locations?
     - Is the observed phenomenon verified for the target version rather than conflated with an older or newer release?
     - Is the observed phenomenon consistently reproducible or a transient side effect?
     - Cross-check against alternative evidence if the finding seems counterintuitive.
   - Discard unverified claims or clearly label them as unconfirmed anomalies.

Verify: All candidate facts have been directly checked against actual tool outputs or file contents, exact version contexts and surrounding version deltas/issues (both unresolved and recently resolved) are identified, and no unverified claims remain in the findings.

---

## Step 3: Articulate Logical Interpretation & Decisions

Transform verified facts into conclusions by explicitly detailing the deductive or inductive chain of reasoning against the confirmed premises.

1. **Strictly Separate Facts from Inferences**:
   - **Facts (Observed Reality)**: Direct observations, exact tool outputs, version-pinned specifications, committed file contents, and exit codes.
   - **Inferences / Interpretations (Agent Reasoning)**: Hypotheses, conclusions, causality deductions, compatibility assessments, or risk evaluations derived from those facts.
2. **Explain the Deductive Bridge**:
   - When reaching a judgment or decision based on facts, explain *how* those facts were interpreted to arrive at the conclusion:
     - State the underlying principle, confirmed premise (architecture, functional/non-functional requirement, inviolable constraint), or invariant applied to the facts.
     - Detail how version deltas, unresolved issues, or recently resolved issues impact the viability of the approach under the target version.
     - Detail why alternative interpretations or candidate versions/tools were rejected.
   - Continue providing this logical rationale even after decisions or remedies are implemented.

Verify: Every judgment is supported by a documented logical bridge from version-pinned verified facts to conclusions, evaluated against confirmed premises and constraints, maintaining unambiguous separation between observation and speculation.

---

## Step 4: Report with Grounded Evidence and Suggest Next Handoff

Present the structured investigation report to the user with full traceability.

1. **Structure the Evidentiary Report**:
   - Format the findings using the template from [source-scoping-and-verification.md](references/source-scoping-and-verification.md) (references/source-scoping-and-verification.md), co-presenting all file paths, line ranges, and URLs with clickable Markdown links and unadorned raw strings per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md) (e.g., `[file.ext#L10-L20](<accessible-target>) (path/to/file.ext:10-20)`):
     - **Confirmed Premises & Constraints**: Architecture, functional/non-functional requirements, out-of-scope items, inviolable constraints, and target versions.
     - **Agreed Sources & Scope**: Sources consulted and any sources excluded.
     - **Verified Facts (Version-Pinned)**: Concrete findings with exact citations:
       - **Version / Commit Hash**: Explicit version number and commit hash (abbreviated or full) at which the fact holds true.
       - **Path & Line Numbers**: Clickable link and workspace-relative file path with precise line numbers or ranges.
       - **Target URL**: Clickable link and raw URL for the corresponding remote repository permalink if configured; state local relative path if unpushed or operating in a local-only environment.
       - **Outputs**: Exact tool commands and raw output snippets.
     - **Version Deltas & Surrounding Issues**: Feature additions/removals/changes across adjacent versions, unresolved issues (with workarounds), and recently resolved issues (with patch boundary and regression notes).
     - **Interpretation & Reasoning**: Step-by-step logic bridging facts and version behavior to conclusions under the confirmed constraints.
     - **Conclusions / Judgments**: Explicit verdict, root cause, or architectural/tooling recommendation.
2. **Recommend Sequential Handoff**:
   - For remediation or code modification, recommend [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md) or [refactor](../refactor/SKILL.md) (../refactor/SKILL.md).
   - For documentation or comment refinement based on findings, recommend [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md).
   - For planning large feature additions based on findings, recommend [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md).
   - For further adversarial scrutiny of complex trade-offs, recommend [scrutinize-plan-execute](../scrutinize-plan-execute/SKILL.md) (../scrutinize-plan-execute/SKILL.md).

Verify: The report provides complete evidentiary and version traceability for every finding, formats all paths and URLs with dual presentation, and proposes appropriate next-step skills for user approval.

---

## Guidelines

- **Resolve objective premises from local manifests first, then clarify ambiguities**: Inspect local manifests (`go.mod`, `package.json`, `Cargo.toml`, etc.) for objective version and toolchain facts, and prompt the user with focused questions and default hypotheses when genuine ambiguity remains.
- **Pin facts to exact versions and inspect surrounding deltas**: Record the exact version or commit hash for every observed fact and compare feature additions, removals, and known issues across adjacent releases.
- **Verify all claims against primary outputs before reporting**: Check exit codes, file contents, and line numbers directly so only confirmed facts are presented as truth.
- **Co-present evidence targets per `output-presentation`**: Cite exact versions, commit hashes, workspace-relative paths, line numbers, and permalinks using dual link and raw-string presentation per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md).
- **Honor source exclusions and inviolable constraints**: Restrict exploration and recommendations to user-approved sources and guardrails.
- **Separate observed facts from logical deductions**: Keep direct observations (`Facts`) distinct from hypotheses and conclusions (`Interpretations`) so reasoning remains transparent and auditable.
- **Maintain a strictly read-only scope**: Investigate and report without modifying repository files, handing off to [refactor](../refactor/SKILL.md) (../refactor/SKILL.md), [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md), or [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md) when changes are needed.
