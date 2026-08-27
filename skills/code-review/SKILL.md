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

---

## Step 1: Identify Review Scope & Focus

Determine the target files and focus area:

- **Target scope**: Specific files/directories, a git diff (`git diff HEAD`), or an entire branch/PR.
- **Review mode / focus**:
  - **General Review**: Readability, maintainability, consistency, logic bugs, adherence to conventions.
  - **Security Audit**: Vulnerability assessment (OWASP Top 10, injection, auth/access control, secret leaks, unsafe dependencies, memory safety/concurrency issues).
  - **Documentation Review**: Clarity, completeness, accuracy with code reality, formatting, broken links/examples.
  - **Comprehensive**: All of the above combined.

Verify: Target files and criteria are defined before starting analysis.

---

## Step 2: Static Analysis & Code Exploration

Run relevant read-only checks if tools are available:

- Linters, format checkers, type-checkers (e.g., `golangci-lint`, `ruff`, `eslint`, `cargo check`).
- Security tools or grep searches for secrets / insecure patterns (e.g., hardcoded keys, SQL concatenation).
- Inspect files and diffs thoroughly in context.

Do **not** edit any files during this process.

Verify: All target files/diffs have been inspected.

---

## Step 3: Categorize and Format Findings

Structure the review feedback by severity levels:

- 🔴 **Critical / Security Vulnerabilities**: Direct security risks (secret leaks, injection, auth bypass) or severe bugs causing crashes/data loss.
- 🟡 **Warnings / Improvements**: Suboptimal design, performance bottlenecks, edge-case bugs, maintainability debt.
- 🟢 **Suggestions / Nitpicks**: Style consistency, doc typos, minor refactoring ideas.
- 💡 **Positive Notes**: Well-structured code or good patterns worth acknowledging.

Format each finding with:
1. **File and Line Reference** (relative path, e.g. `[path/to/filename.ext:L10-L15](path/to/filename.ext#L10-L15)`)
2. **Issue Description** (what is wrong and why)
3. **Recommended Fix / Concrete Code Example**


---

## Step 4: Present Findings and Next Steps

Present the structured report to the user.
Provide an executive summary and ask if the user wants to proceed with fixing any of the items (e.g., switching to the `refactor` skill).

Verify: The report is clear, actionable, and non-blocking.

---

## Guidelines

- **Read-only enforcement.** Never edit, rewrite, or auto-fix files during review.
- **Actionable feedback.** Every finding must include concrete rationale and an example fix.
- **Prioritize safety.** Surface critical security risks and crash-inducing bugs prominently at the top.
- **Respect codebase style.** Do not enforce personal styling preferences if the existing project has an established convention.

