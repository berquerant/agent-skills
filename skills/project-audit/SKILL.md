---
name: project-audit
description: >-
  Use this skill to conduct a comprehensive repository-wide health check, assess
  public release readiness and exposure risks, audit project security, or evaluate
  maintainability and technical debt. Produces a prioritized audit report with
  clear pass/fail ratings and actionable remedies without modifying files.
---

# Project Audit (Health, Security, Public Release Readiness, & Maintainability)

Use this skill to perform an in-depth, repository-wide diagnostic audit. It inspects
project hygiene and build health, identifies security vulnerabilities and public
exposure risks (in both current state and recent commit history), evaluates code
maintainability and technical debt, and provides prioritized, actionable remedies.

This skill is strictly read-only and never modifies files directly. For security and
vulnerability criteria, refer to [`../code-review/references/security-checklist.md`](../code-review/references/security-checklist.md).
For maintainability heuristics, refer to [`references/maintainability-metrics.md`](references/maintainability-metrics.md).

---

## Triggers

Activate this skill when:
- The user requests a general health check, audit, or diagnostic of the repository (e.g. "audit this project", "check repository health").
- The project is being prepared for public release, open-sourcing, or external deployment, and needs an exposure risk assessment.
- Evaluating code quality, technical debt, or refactoring opportunities across the codebase.
- A comprehensive macro-level status audit is needed beyond the lightweight working-tree check provided by `project-status`.

---

## Step 1: Determine Audit Mode & Commit History Window

Clarify the inspection mode and inspect the recent commit window:

### 1.1 Select Mode
- **`full` (Default)**: Executes all checks (Hygiene, Security & Public Exposure, Maintainability).
- **`release`**: Focuses strictly on Security and Public Exposure risks (credentials, private IPs, internal endpoints, licenses).
- **`hygiene`**: Focuses on build, test, lint, configuration hygiene, and repository metadata.
- **`maintainability`**: Focuses on code complexity, technical debt markers, and refactoring opportunities.

### 1.2 Determine Commit History Window
To detect recently committed secrets or sensitive internal details, inspect the larger of:
- The **last 20 commits** (`git log -n 20`), or
- All commits within the **last 1 hour** (`git log --since="1 hour ago"`).

```sh
# Calculate commit count in the last 1 hour vs last 20 commits
COUNT_1H=$(git log --since="1 hour ago" --oneline 2>/dev/null | wc -l | tr -d ' ')
if [ "$COUNT_1H" -gt 20 ]; then
  COMMIT_TARGET_ARG="--since=\"1 hour ago\""
  echo "Using last 1 hour window ($COUNT_1H commits)"
else
  COMMIT_TARGET_ARG="-n 20"
  echo "Using last 20 commits window"
fi
```

Verify: The audit mode and commit history window are explicitly established before running checks.

---

## Step 2: Project Hygiene & Configuration Audit

*(Execute if mode is `full` or `hygiene`)*

Inspect repository configuration and baseline operational health:

1. **Working Tree & Ignored Files**:
   - Check `git status` for unintended untracked files.
   - Inspect `.gitignore` for standard exclusions (build artifacts, temporary files, editor settings, local environments like `.env*`).
2. **Build, Test, and Linter Health**:
   - Run the project's standard build and test suites (e.g. `go test ./...`, `npm test`, `cargo test`, `pytest`).
   - Run linters if configured (e.g. `golangci-lint`, `eslint`, `ruff`, `make lint`).
   - Note failing suites, compilation errors, or excessive warnings.
3. **Repository Metadata & Documentation**:
   - Verify existence of essential project files: `README.md`, `LICENSE`, `CONTRIBUTING.md` (if applicable), and agent operational guidelines (`AGENTS.md`).
4. **Dependency & Lockfile Integrity**:
   - Confirm lockfiles exist and match dependency manifests (e.g. `go.sum`, `package-lock.json`, `Cargo.lock`, `poetry.lock`).

Verify: Hygiene status, build/test execution results, and metadata presence are documented.

---

## Step 3: Security & Public Exposure Risk Assessment

*(Execute if mode is `full` or `release`)*

Inspect current files and the commit window established in Step 1 against
[`../code-review/references/security-checklist.md`](../code-review/references/security-checklist.md):

1. **Credentials & Secrets**:
   - Search working tree and recent commits for API keys, bearer tokens, private keys (`BEGIN PRIVATE KEY`), and passwords.
   - Ensure no `.env`, credentials, or keystore files are committed.
2. **Internal Network & Infrastructure Exposure**:
   - Grep for internal domain names (`*.internal`, `*.corp`, `*.local`).
   - Grep for private RFC 1918 IPv4 addresses (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`) not used as standard mock fixtures.
   - Identify staging/development endpoint URLs.
3. **Confidentiality & Compliance**:
   - Check for internal employee names, internal codenames, or customer test data.
   - Verify open-source licensing compliance and ensure proprietary code is not inadvertently exposed.
4. **Known Vulnerabilities**:
   - Run dependency vulnerability checks if supported by the project toolchain (e.g. `govulncheck`, `npm audit`, `cargo audit`, `pip-audit`).

```sh
# Sample check across the commit window:
git log $COMMIT_TARGET_ARG -p | grep -inE "(api[_-]?key|secret|token|password|bearer|private[_-]?key)" || true
```

Verify: Both working tree and the selected commit window are scanned for secrets, internal infrastructure identifiers, and vulnerabilities.

---

## Step 4: Maintainability & Technical Debt Analysis

*(Execute if mode is `full` or `maintainability`)*

Analyze the codebase according to [`references/maintainability-metrics.md`](references/maintainability-metrics.md):

1. **Code Size & Complexity**:
   - Identify unusually large files (> 600 lines) and deeply nested functions (> 4 levels).
2. **Debt & Workaround Markers**:
   - Catalog `TODO`, `FIXME`, `HACK`, `XXX`, and `DEPRECATED` comments across non-vendor code.
3. **Duplication & Coupling**:
   - Identify repeated boilerplate, God classes/modules, or circular dependency risks.
4. **Testability & Test Gaps**:
   - Identify core packages or business logic lacking automated unit tests.

Verify: Code metrics, large files, debt markers, and test gap locations are recorded.

---

## Step 5: Synthesize and Present the Audit Report

Produce a structured markdown audit report using the following template:

```markdown
# 🩺 Project Audit Report

**Audit Mode**: <full / release / hygiene / maintainability>
**Commit History Window**: <last 20 commits / last 1 hour (N commits)>

## Overall Verdict: [ ✅ Clean / ⚠️ Caution / ❌ Blocked ]
- **Public Release Readiness**: [ Ready / Blocked / Caution / N/A ]
- **Health & Hygiene**: [ Pass / Needs Attention / Fail / N/A ]
- **Security & Exposure**: [ Pass / High Risk / N/A ]
- **Maintainability & Debt**: [ Good / Moderate Debt / High Debt / N/A ]

---

### 1. Security & Public Exposure Risks
- 🔴 Critical / Blocker: <Findings or "None">
- 🟡 Warnings: <Findings or "None">
- 🛡️ Dependency Vulnerabilities: <Summary or "None reported">

### 2. Project Hygiene & Operational Health
- Build / Test / Lint: [✅ Pass / ❌ Fail / — Skipped]
- Repository Metadata (README, LICENSE, AGENTS.md): <Status>
- Configuration & .gitignore: <Status>

### 3. Maintainability & Technical Debt
- Large Files / High Complexity: <List top items or "None">
- Outstanding Debt Markers (TODO/FIXME): <Total count and critical samples>
- Test Gaps: <Untested core areas or "None">

---

### Prioritized Remediation Plan
1. **P0 (Immediate / Blocking)**: <Actionable item with file:line reference>
2. **P1 (High Priority)**: <Actionable item>
3. **P2 (Improvement / Tech Debt)**: <Actionable item>
```

Verify: The report assigns clear ratings, cites specific files and line numbers for all issues, and states an unambiguous verdict.

---

## Step 6: Recommend Next Actions (Sequential Hand-off)

Do not edit files within this skill. Present the report to the user and recommend the appropriate follow-up:

- **For structural refactoring & debt remediation**: Recommend delegating to the [`refactor`](../refactor/SKILL.md) skill.
- **For multi-step fixes or critical blocker resolution**: Recommend structuring changes with the [`step-gate`](../step-gate/SKILL.md) skill.
- **For post-remediation verification**: Use [`project-status`](../project-status/SKILL.md) to confirm clean working-tree diffs.

Verify: The user is presented with clear options to proceed with remediation without unprompted file mutations.

---

## Guidelines

- **Strict read-only discipline.** Never modify or delete files during an audit. Only inspect, run tests, and report.
- **Zero tolerance on public exposure.** If any credential, private key, or internal infrastructure identifier is detected, mark the verdict as ❌ Blocked and highlight it prominently.
- **Dynamic commit window.** Always calculate whether the last 20 commits or the last 1 hour represents the larger commit span, and audit the larger range.
- **Leverage shared references.** Use [`../code-review/references/security-checklist.md`](../code-review/references/security-checklist.md) as the canonical security standard to prevent rule drift.
- **Actionable reporting.** Avoid vague feedback like "refactor large files." Provide concrete file paths, line ranges, and specific recommendations.
