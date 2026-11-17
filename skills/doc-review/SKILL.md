---
name: doc-review
description: >-
  Use this skill when reviewing, auditing, or verifying technical documentation,
  specifications, READMEs, or API guides. Analyzes technical accuracy against code reality,
  structural clarity, completeness, hygiene, and public exposure safety without modifying files.
license: MIT
---

# Technical Documentation Review

Use this skill to inspect Markdown documentation, system specifications, API references, user guides, or repository READMEs.
Provides structured, actionable feedback categorized by severity, covering code-truth alignment, structural clarity, asset hygiene, and public exposure risks.

This skill is strictly read-only and directly references three auxiliary guides:
- [doc-review-checklist.md](references/doc-review-checklist.md) (references/doc-review-checklist.md): Core evaluation criteria across five documentation dimensions.
- [code-alignment-guide.md](references/code-alignment-guide.md) (references/code-alignment-guide.md): Procedures for cross-checking documentation against codebase reality.
- [report-template.md](references/report-template.md) (references/report-template.md): Finding format (Where/Rule/Why/How) and structured report templates.

---

## Step 1: Identify Review Scope & Mode

Determine the target documentation files and review mode:

1. **Identify Target Files**:
   - Single document or directory (e.g. `README.md`, `docs/`, `skills/<name>/SKILL.md`).
   - Git topic branch or PR diff (e.g. `git diff origin/main...HEAD -- '*.md'`).
   - Repository-wide documentation audit.
2. **Select Review Mode**:
   - **`quick`**: Fast assessment focusing on structural clarity, broken links, Why-first framing, and public release exposure hazards.
   - **`thorough`**: Exhaustive audit including codebase cross-checking (Code-Truth Alignment), full relative link traversal, Mermaid diagram syntax validation, and information completeness.
3. **Record Workspace Baseline**:
   - Record `git status --short` before starting to verify that no files are modified during review.

Verify: Target files, comparison base (if reviewing a diff), review mode (`quick` vs `thorough`), and baseline git status are identified.

---

## Step 2: Read-Only Inspection & Codebase Cross-Checking

Inspect the target documentation along the five dimensions defined in [doc-review-checklist.md](references/doc-review-checklist.md) (references/doc-review-checklist.md):

1. **Code-Truth Alignment & Drift Detection**:
   - Follow [code-alignment-guide.md](references/code-alignment-guide.md) (references/code-alignment-guide.md) to cross-check commands, code snippets, function signatures, configuration keys, and file paths against actual source files in the repository.
   - Verify that code blocks reflect current reality rather than obsolete implementations.
2. **Structural Clarity & Reader Persona**:
   - Check that explanations lead with the "Why" (rationale, architectural intent) before detailing the "What" or "How".
   - Confirm logical information hierarchy: Overview/Intent → Prerequisites → Procedures/Examples → Verification/Edge Cases.
3. **Technical Hygiene & Assets**:
   - Verify that all relative links, file anchors, and image targets exist and resolve correctly.
   - Validate that Mermaid diagrams render without syntax errors and code blocks specify appropriate language tags.
4. **Public Exposure & Confidentiality Safety**:
   - Cross-check against the confidentiality criteria in [doc-review-checklist.md](references/doc-review-checklist.md) (references/doc-review-checklist.md) to ensure zero leaks of API keys, tokens, internal hostnames, private IPs, or staging environments.
5. **Information Economy & Conservation**:
   - Ensure prerequisites, invariants, and failure modes are explicitly stated without arbitrary omission or unnecessary filler.

Do **not** edit, format, or auto-fix any files during inspection.

Verify: All target documents have been inspected across all five dimensions without modifying any files in the workspace.

---

## Step 3: Categorize Findings by Severity

Organize all review feedback into four severity categories using the schema in [report-template.md](references/report-template.md) (references/report-template.md):

- 🔴 **Critical / Exposure & Correctness Hazards**:
  - Public exposure risks: leaked credentials, intranet domains, staging URLs, private IPs, or confidential architecture data.
  - Harmful errors: misleading commands or code examples that fail, corrupt state, or cause security hazards if executed as written.
  - Broken primary entrypoints or dead links to mandatory resources.
- 🟡 **Warnings / Quality & Drift Debt**:
  - Code-doc drift: stale function signatures, renamed parameters, obsolete flags, or outdated architectural descriptions.
  - Structural defects: missing "Why" rationale, poor hierarchy, omitted prerequisite versions or constraints.
  - Invalid diagram syntax or unescaped characters causing rendering errors.
- 🟢 **Suggestions / Clarity Enhancements**:
  - Phrasing improvements, readability polish, formatting consistency, or missing language tags on code blocks.
  - Recommendations to add illustrative Mermaid diagrams or troubleshooting tips.
- 💡 **Positive Notes**:
  - Commend well-structured explanations, thorough edge-case warnings, or exceptionally clear examples.

Format each finding with:
1. **Where**: Exact file and line reference (e.g. `docs/guide.md#L42-L48`).
2. **Rule**: Applicable standard from [doc-review-checklist.md](references/doc-review-checklist.md) (references/doc-review-checklist.md).
3. **Why**: Rationale detailing the confusion, drift risk, or security hazard caused by the current text.
4. **How**: Concrete replacement text or structural revision demonstrating the recommended improvement.

Verify: Every finding is assigned an explicit severity level with Where, Rule, Why, and How elements.

---

## Step 4: Present Findings & Recommend Next Steps

Present the structured review report using [report-template.md](references/report-template.md) (references/report-template.md):

1. **Highlight Exposure Hazards Upfront**:
   - If any public exposure or critical correctness issue is identified, display an immediate warning banner at the top of the report.
2. **Deliver the Actionable Report**:
   - Present findings grouped by file and severity, followed by positive observations.
3. **Confirm Workspace Baseline**:
   - Run `git status --short` to verify zero file modifications occurred during the review.
4. **Recommend Sequential Hand-Off Options**:
   - **For applying text updates, comment additions, or structural rewrites**: Recommend [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md).
   - **For drafting missing documentation, RFCs, or ADRs from scratch**: Recommend [structured-writing](../structured-writing/SKILL.md) (../structured-writing/SKILL.md).
   - **For posting review comments or conducting interactive PR discussions**: Recommend [review-discussion](../review-discussion/SKILL.md) (../review-discussion/SKILL.md).
   - **For fixing underlying code bugs revealed by documentation drift**: Recommend [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md).
   - **For preparing commit messages and PR descriptions**: Recommend [change-message](../change-message/SKILL.md) (../change-message/SKILL.md).

Verify: The report is presented with all findings, unchanged git status is confirmed, and sequential hand-off options are provided.

---

## Guidelines

- **Maintain a strictly read-only review posture**: Inspect and report findings without editing, formatting, or rewriting files so reviews remain non-destructive and safe.
- **Anchor reviews in codebase reality**: Verify documented code snippets, options, and APIs against actual repository source code rather than evaluating documentation in isolation.
- **Prioritize the "Why" and structural clarity**: Ensure documentation explains the rationale and constraints behind architectural decisions before describing syntax or mechanics.
- **Surface public exposure risks immediately**: Place warnings for leaked secrets, intranet hostnames, and staging credentials prominently at the top of the report.
- **Provide concrete, drop-in replacement remedies**: Include actionable `Where`, `Rule`, `Why`, and `How` replacement text for every finding to minimize ambiguity for the author.
