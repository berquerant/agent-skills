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
strictly preserving external behavior.

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
  - *Pure documentation & comment refinement*: When refining explanations, docstrings, or code comments without modifying any code or losing existing rationale, prefer [`doc-refactor`](../doc-refactor/SKILL.md).
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
2. Coordinate with the `code-review` skill (or apply its review criteria) to evaluate the diff for code quality, design cleanliness, residual duplication, and potential regressions.
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

## Guidelines

- **Behavior preservation first.** Never change external behavior, contracts, or public APIs during a refactoring task.
- **Pragmatic DRY.** Maximize code reuse and eliminate duplicate logic within reasonable bounds; avoid over-complicating abstractions solely for DRY's sake.
- **Mandatory green checks.** The project's configured lint and test tasks must pass with zero failures before completing the refactoring.
- **Continuous diff review.** Leverage the `code-review` skill or its methodology on diffs to drive refactoring quality forward.
- **Keep documentation in sync.** Always verify and update `README.md` and agent-facing docs (`AGENTS.md`, skill descriptions) whenever code changes affect them.
- **Safety baselines & atomic steps.** Never refactor without passing baseline checks, and make one logical transformation at a time.

