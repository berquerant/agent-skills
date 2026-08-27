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
  - Code smell elimination (long functions, duplicated code, deep nesting).
  - Modernization / idiom adoption (leveraging newer language features).
  - Architectural / structural cleanup (separation of concerns, dependency injection).
  - Documentation reorganization (splitting large docs, standardizing structure).
- **Invariants**: Explicitly list behaviors and interfaces that must NOT change.

Verify: Invariants and goals are clearly defined.

---

## Step 2: Establish Safety Baselines & Verification

Before modifying anything, ensure you can verify behavior:

- Run existing test suites (e.g. `go test ./...`, `npm test`, `pytest`, `cargo test`).
- If test coverage is insufficient for the target code, write characterization / baseline tests first.
- For documentation, verify build/linting tools (e.g. markdown linter, doc generator).

Verify: All baseline tests pass before any refactoring begins.

---

## Step 3: Incremental Refactoring

Refactor in small, atomic steps:

1. Apply **one** transformation at a time (e.g. Extract Function, Rename Variable, Replace Conditional with Polymorphism).
2. Maintain existing coding style, naming conventions, and documentation comments.
3. Run verification / tests after each individual step.
4. If a step breaks tests or behavior, revert immediately and re-evaluate.

Verify: Tests pass after each atomic change.

---

## Step 4: Final Verification & Review

1. Run the entire test suite and linters across the project.
2. Inspect the overall diff (`git diff HEAD`) to ensure no unintended modifications or behavior changes occurred.
3. Present the refactoring summary and diff highlights to the user.

Verify: Zero functional regressions. Codebase is cleaner and tests are passing.

---

## Guidelines

- **Behavior preservation first.** Never change external behavior or public APIs during a refactoring task.
- **Safety baselines.** Never refactor without passing baseline tests or a reliable verification loop.
- **Atomic steps.** Make one logical transformation at a time, testing immediately after each step.
- **Stop and surface unexpected issues.** If tests fail or assumptions break, revert to the last working state before retrying.

