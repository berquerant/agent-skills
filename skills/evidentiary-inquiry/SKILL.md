---
name: evidentiary-inquiry
description: >-
  Use this skill when an investigation or diagnostic task requires rigorous evidence,
  transparent source scoping, pre-report fact verification, and clear separation between
  observed facts and logical interpretations. Scopes and agrees on information sources
  with the user before exploring, validates facts prior to reporting, explains reasoning,
  and avoids excessive or unauthorized sources.
---

# Evidentiary Inquiry

Use this skill when conducting investigations, root cause analyses, technical evaluations, or fact-finding tasks where precision, traceability, and verifiable evidence are required.
The agent forecasts and agrees upon information sources before exploring, verifies all findings before presenting them, clearly separates observed facts from inferences, and explains the logical deduction behind every judgment.

For reporting templates, source scoping formats, and fact-checking checklists, refer to [references/source-scoping-and-verification.md](references/source-scoping-and-verification.md).

---

## Step 1: Forecast and Scope Information Sources

Before executing exploratory commands or reading files, predict which information sources will be needed and confirm them with the user.

1. **Forecast Candidate Sources**:
   - List the primary information sources anticipated for the inquiry:
     - Specific repository files and directories (e.g., config files, implementation modules).
     - Local tools, linters, or test runners (e.g., `git`, test suites, static analysis tools).
     - Documentation or spec files within local references.
2. **Classify and Propose Sources**:
   - Categorize the candidate sources:
     - **Primary / Priority Sources**: First-line sources essential to resolving the inquiry.
     - **Secondary / Fallback Sources**: Used only if primary sources yield incomplete evidence.
   - Present the anticipated sources to the user and request review:
     > *"Before investigating, I plan to examine the following sources: [List priority and fallback sources]. Are any of these excessive, insufficient, or in need of adjustment?"*
3. **Establish Source Constraints**:
   - Prioritize sources agreed upon with the user.
   - You are not strictly forbidden from consulting additional unpredicted sources if genuine inquiry demands it; however, **sources designated as excessive or out-of-scope by the user must be strictly avoided**.

Verify: Planned information sources are forecasted, categorized into priority/fallback tiers, presented to the user, and agreed upon with explicit exclusion of excessive sources.

---

## Step 2: Investigate and Self-Verify Evidence

Execute the investigation focusing on the agreed sources, and verify all facts before formulating reports.

1. **Targeted Exploration**:
   - Probe the prioritized sources first.
   - Record exact tool commands, file paths, line ranges, and raw tool outputs.
   - If investigation threatens to become prolonged or stalls, respect interval pacing or consult pacing guidelines.
2. **Pre-Report Fact-Checking (Self-Verification)**:
   - Before presenting any fact to the user, verify its truth and reproducibility:
     - Did the command exit with status 0, or did it fail silently?
     - Does the cited file and line range actually contain the stated code or configuration?
     - Is the observed phenomenon consistently reproducible or a transient side effect?
     - Cross-check against alternative evidence if the finding seems counterintuitive.
   - Discard unverified claims or clearly label them as unconfirmed anomalies.

Verify: All candidate facts have been directly checked against actual tool outputs or file contents, and no unverified claims remain in the findings.

---

## Step 3: Articulate Logical Interpretation & Decisions

Transform verified facts into conclusions by explicitly detailing the deductive or inductive chain of reasoning.

1. **Strictly Separate Facts from Inferences**:
   - **Facts (Observed Reality)**: Direct observations, exact tool outputs, committed file contents, and exit codes.
   - **Inferences / Interpretations (Agent Reasoning)**: Hypotheses, conclusions, causality deductions, or risk assessments derived from those facts.
2. **Explain the Deductive Bridge**:
   - When reaching a judgment or decision based on facts, explain *how* those facts were interpreted to arrive at the conclusion:
     - State the underlying principle, invariant, or assumption applied to the facts.
     - Detail why alternative interpretations were rejected.
   - Continue providing this logical rationale even after decisions or remedies are implemented.

Verify: Every judgment is supported by a documented logical bridge from verified facts to conclusions, maintaining unambiguous separation between observation and speculation.

---

## Step 4: Report with Grounded Evidence and Suggest Next Handoff

Present the structured investigation report to the user with full traceability.

1. **Structure the Evidentiary Report**:
   - Format the findings using the template from [references/source-scoping-and-verification.md](references/source-scoping-and-verification.md):
     - **Agreed Sources & Scope**: Sources consulted and any sources excluded.
     - **Verified Facts**: Concrete findings with exact citations (file paths, line numbers, tool commands, output snippets).
     - **Interpretation & Reasoning**: Step-by-step logic bridging facts to conclusions.
     - **Conclusions / Judgments**: Explicit verdict, root cause, or architectural decision.
2. **Recommend Sequential Handoff**:
   - For remediation or code modification, recommend [`step-gate`](../step-gate/SKILL.md) or [`refactor`](../refactor/SKILL.md).
   - For documentation or comment refinement based on findings, recommend [`doc-refactor`](../doc-refactor/SKILL.md).
   - For planning large feature additions based on findings, recommend [`explore-plan-execute`](../explore-plan-execute/SKILL.md).
   - For further adversarial scrutiny of complex trade-offs, recommend [`scrutinize-plan-execute`](../scrutinize-plan-execute/SKILL.md).

Verify: The report provides complete evidentiary traceability for every finding and proposes appropriate next-step skills for user approval.

---

## Guidelines

- **Pre-report verification is mandatory.** Never communicate an unconfirmed fact to the user as truth; always verify against tool outputs or file contents first.
- **Strictly honor source exclusions.** While exploration may expand to unpredicted sources if necessary, never consult sources that the user flagged as excessive or off-limits.
- **Unambiguous demarcation.** Keep "what was observed" (facts) clearly distinct from "what it means" (interpretations and hypotheses) in both internal reasoning and user-facing reports.
- **Transparent rationale before and after decisions.** Always explain the logical chain of thought leading to a judgment, both when proposing it and after executing it.
- **Read-only investigation.** This skill investigates and diagnoses; it does not modify repository files. Suggest remediation skills (`refactor`, `step-gate`) when changes are required.
