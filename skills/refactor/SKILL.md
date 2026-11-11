---
name: refactor
description: >-
  Use this skill when the user wants to refactor code or improve documentation
  structure without changing external behavior. Guides through planning,
  incremental transformations, and verification loops to ensure safety.
---

# Code & Documentation Refactor

Use this skill to refactor existing code, architecture, or documentation.
Improves readability, maintainability, modularity, and performance while
strictly preserving external behavior. See [refactoring-patterns.md](references/refactoring-patterns.md) (references/refactoring-patterns.md) for the atomic refactoring catalog and safety invariants.

---

## Step 1: Define Refactoring Goals & Scope

Clarify the specific objective:

- **Target**: Specific modules, classes, functions, or documentation files.
- **Goals**:
  - **Pragmatic DRY (Don't Repeat Yourself)**: Eliminate duplicate logic, repeated blocks, and redundant declarations to the extent reasonable, avoiding premature abstraction or over-engineering.
  - Code smell elimination (long functions, deep nesting, dead code, excessive coupling).
  - Modernization / idiom adoption (leveraging newer language features and conventions).
  - Architectural / structural cleanup (separation of concerns, dependency injection, modularity).
  - Documentation reorganization (splitting large docs, standardizing structure).
  - *Pure documentation & comment refinement*: When refining explanations, docstrings, or code comments without modifying any code or losing existing rationale, prefer [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md).
- **Invariants**: Explicitly list behaviors, public APIs, contracts, and interfaces that must NOT change.

Verify: Invariants, DRY targets, and goals are clearly defined.

---

## Step 2: Establish Safety Baselines & Verification

Before modifying anything, detect and run the project's quality checks to ensure a working baseline:

- Identify configured lint and test tasks (e.g. `go test ./...`, `npm run lint && npm test`, `pytest`, `cargo test`, `make test`, `pre-commit`).
- Run the baseline test and lint suites.
- If test coverage is insufficient for the target code, write characterization / baseline tests first.
- For documentation, verify build/linting tools (e.g. markdown linter, doc generator).

Verify: Baseline tests and linters pass before any refactoring begins.

---

## Step 3: Incremental Refactoring

Refactor in small, atomic steps:

1. Apply **one** transformation at a time (e.g. Extract Function/Class, Parameterize Method, De-duplicate logic, Rename Variable).
2. Maintain existing coding style, naming conventions, and documentation comments.
3. Run verification / tests after each individual step.
4. If a step breaks tests or behavior, revert immediately and re-evaluate.

Verify: Tests pass after each atomic change.

---

## Step 4: Diff Review & Iterative Refactoring

Engage in iterative review to guide and advance the refactoring:

1. Inspect the incremental or intermediate diff (`git diff HEAD`).
2. Coordinate with the [code-review](../code-review/SKILL.md) (../code-review/SKILL.md) skill (or apply its review criteria) to evaluate the diff for code quality, design cleanliness, residual duplication, and potential regressions.
3. Incorporate review findings into subsequent incremental refactoring passes until the code is clean, cohesive, and sufficiently DRY.

Verify: The diff has been reviewed against code-review standards and further refined.

---

## Step 5: Final Verification, Task Execution & Documentation Sync

1. **Mandatory Lint & Test Success**: Run all lint, formatting, type-check, and test tasks configured in the project. **All tasks must pass cleanly (100% success)**; fix any lint warnings, formatting errors, or test failures before concluding.
2. **Documentation & Spec Synchronization**:
   - Inspect the codebase diff against project documentation.
   - Check if `README.md`, agent guidance documents (`AGENTS.md`, prompt templates, skills), or API docs are affected by the changes or need clarification/updating.
   - Update and correct any outdated explanations, configuration examples, or references to accurately reflect the current project state.
3. **Final Presentation**: Present the refactoring summary, diff highlights, and documentation updates to the user.

Verify: All configured lint and test tasks pass completely. Documentation (README and agent-facing docs) accurately reflects the current state of the project.

---

## Step 6: Hand Off for Verification & Commit

Do not modify files in this step. Recommend follow-up skills to complete the workflow:

- **For post-refactoring status & diff inspection**: Recommend [project-status](../project-status/SKILL.md) (../project-status/SKILL.md).
- **For drafting commit messages and PR/MR descriptions**: Recommend [change-message](../change-message/SKILL.md) (../change-message/SKILL.md).
- **For thorough code review against security and design standards**: Recommend [code-review](../code-review/SKILL.md) (../code-review/SKILL.md).

Verify: Clear next steps and recommended follow-up skills are presented to the user without unprompted changes.

---

## Guidelines

- **Preserve external behavior strictly**: Keep public APIs, contracts, and observable outputs unchanged throughout the refactoring so callers experience zero regressions.
- **Apply pragmatic DRY**: Consolidate duplicated domain logic where it improves clarity while avoiding premature abstractions that obscure control flow.
- **Verify green checks at every step**: Establish a passing test/lint baseline in Step 2 and re-run checks after each atomic transformation in Step 3 so any breakage is caught and reverted immediately.
- **Review intermediate diffs**: Apply [code-review](../code-review/SKILL.md) (../code-review/SKILL.md) criteria to `git diff HEAD` to catch residual duplication or style drift.
- **Synchronize affected documentation**: Update `README.md` and agent-facing docs (`AGENTS.md`, skill references) whenever structural changes affect file paths or module boundaries.

