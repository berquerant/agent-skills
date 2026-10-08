---
name: code-review
description: >-
  Use this skill when the user wants a review or security audit of code or
  documentation. Analyzes quality, readability, maintainability, design, and
  potential vulnerabilities, producing a structured, actionable report without
  modifying files.
---

# Code & Documentation Review (with Security Audit)

Use this skill to inspect code, configurations, or documentation.
Provides clear feedback categorized by severity, covering maintainability,
correctness, and security risks.

For detailed vulnerability and public exposure audit criteria, refer to
[references/security-checklist.md](references/security-checklist.md).

---

## Step 1: Identify Review Scope & Focus

Determine the target files and focus area:

- **Target scope**:
  - Specific files/directories or single commit diffs (`git diff HEAD~1`).
  - **Topic branch review**: When reviewing a topic branch or pull request, compare the branch against the repository's default branch (e.g. `git diff origin/main...HEAD` or `git diff main...HEAD`).
- **Review mode / focus**:
  - **General Review**: Readability, maintainability, consistency, logic bugs, adherence to conventions.
  - **Security & Public Release Audit**:
    - Vulnerability assessment (OWASP-aligned: injection, auth/access control, SSRF, memory safety/concurrency).
    - **Public Exposure Safety**: Ensure the diff is completely safe to publish to the internet. Verify zero leaks of sensitive information (secrets, API keys, tokens, private keys), internal network details (intranet hostnames, internal IPs, staging endpoints), or proprietary data. (See [references/security-checklist.md](references/security-checklist.md)).
  - **Documentation Review**: Clarity, completeness, accuracy with code reality, formatting, broken links/examples.
  - **Comprehensive**: All of the above combined.

Verify: Target files, comparison base (e.g., default branch for topic branches), and criteria are defined before starting analysis.

---

## Step 2: Static Analysis & Code Exploration

Run relevant read-only checks if tools are available:

- Linters, format checkers, type-checkers (e.g., `golangci-lint`, `ruff`, `eslint`, `cargo check`).
- Security tools or grep searches for secrets / insecure patterns (cross-reference against [references/security-checklist.md](references/security-checklist.md)).
- Inspect files and diffs thoroughly in context against the comparison base.
- Specifically verify that no files or changes pose a risk if exposed publicly on the internet.

Do **not** edit any files during this process.

Verify: All target files/diffs have been inspected, and public release safety has been assessed.

---

## Step 3: Categorize and Format Findings

Structure the review feedback by severity levels:

- 🔴 **Critical / Security Vulnerabilities & Public Exposure Risks**:
  - Direct security risks (secret leaks, injection, auth bypass) or severe bugs causing crashes/data loss.
  - **Public Release Hazards**: Any leak of credentials, internal infrastructure identifiers, or unapproved confidential artifacts. **If there is even the slightest risk, prominently alert and warn the user.**
- 🟡 **Warnings / Improvements**: Suboptimal design, performance bottlenecks, edge-case bugs, maintainability debt.
- 🟢 **Suggestions / Nitpicks**: Style consistency, doc typos, minor refactoring ideas.
- 💡 **Positive Notes**: Well-structured code or good patterns worth acknowledging.

Format each finding with:
1. **File and Line Reference** (relative path, e.g. `path/to/filename.ext#L10-L15`)
2. **Issue Description** (what is wrong and why)
3. **Recommended Fix / Concrete Code Example**

Verify: All findings are categorized by severity levels (Critical, Warnings, Suggestions) with file:line references and actionable remedies.

---

## Step 4: Present Findings and Next Steps

Present the structured report to the user.
- If any public exposure or security risk is detected, highlight a clear warning upfront before other findings.
- When posting feedback as inline review comments or notes on a PR/MR, format and discuss them following [`review-discussion`](../review-discussion/SKILL.md).
- Provide an executive summary and ask if the user wants to proceed with fixing any of the items (e.g., switching to the `refactor` skill).

Verify: The report is clear, actionable, alerts on any public release risk, and is non-blocking.

---

## Guidelines

- **Read-only enforcement.** Never edit, rewrite, or auto-fix files during review.
- **Actionable feedback.** Every finding must include concrete rationale and an example fix.
- **Prioritize safety & public release caution.** Surface critical security risks, crash-inducing bugs, and any potential exposure risk when publishing to the internet prominently at the top. If any risk exists, warn the user without hesitation.
- **Respect codebase style.** Do not enforce personal styling preferences if the existing project has an established convention.

