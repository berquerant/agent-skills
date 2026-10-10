---
name: doc-refactor
description: >-
  Use this skill when improving or adding documentation and code comments without
  modifying any project code or functionality. Supports both existing-doc refinement
  and adding new docstrings/comments to complex or undocumented code, ensuring zero loss
  of context, logical depth, and complete code preservation.
---

# Documentation & Code Comment Refactor

Use this skill to refine documentation files, markdown guides, API docstrings, or inline code comments.
Enhances clarity, logical organization, and precision while strictly guaranteeing that executable project code remains completely untouched and no existing context or intent is lost.

For checklists, information extraction templates, and zero-code-diff verification patterns, refer to [references/doc-refactor-guide.md](references/doc-refactor-guide.md).

---

## Step 1: Select Scope Mode, Extract Information & Define Boundaries

Identify the target files, determine the improvement mode with the user, inventory existing context, and freeze code boundaries.

1. **Select Scope Mode**:
   - Confirm or propose one of the following modes:
     - **`existing-only`**: Refine and restructure existing documentation and comments only (no new comment headers on undocumented symbols).
     - **`comprehensive` (or `gap-filling`)**: In addition to refining existing text, add new docstrings/headers to undocumented functions/types, add inline comments to **complex or hard-to-understand code blocks**, and create new documentation guides if needed.
2. **Target & Complexity Identification**:
   - Identify target doc files (`.md`, `.rst`) or source code files.
   - For `comprehensive` mode, identify **hard-to-understand code locations** requiring explanatory comments:
     - Non-obvious branching or implicit business invariants.
     - Intricate regular expressions, bitwise manipulations, or mathematical formulas.
     - Workarounds for external quirks, legacy constraints, or subtle concurrency assumptions.
3. **Information Inventory (For Existing Text)**:
   - Before modifying any existing text, catalog its information units:
     - **Core Purpose & Intent**: Why the code or system exists.
     - **Non-Obvious Rationale ("Why")**: Historical decisions, trade-offs, workarounds.
     - **Prerequisites & Invariants**: Assumptions, expected invariants, dependencies.
     - **Warnings & Edge Cases**: Caveats, performance implications, failure modes.
4. **Establish Zero-Code Boundary**:
   - For comments embedded inside source files, demarcate the comment boundaries. Executable tokens, whitespace surrounding code statements, and program logic must not be modified.

Verify: Scope mode (`existing-only` vs `comprehensive`) is decided, hard-to-understand code targets (if applicable) and existing information inventories are established, and code boundaries are frozen.

---

## Step 2: Restructure and Augment for Clarity, Logic, and Depth

Rewrite existing documentation or add new comments to improve cognitive readability while preserving or expanding necessary detail.

1. **Apply Clarity Principles**:
   - **Lead with the Core Conclusion / Intent**: State what a function, module, or concept achieves first.
   - **Explain the "Why" Over the Obvious "What"**: Do not simply narrate code mechanics that are obvious from syntax; elucidate why specific decisions, structures, or algorithms were chosen.
   - **Logical Flow**: Structure logically (e.g. Overview → Context/Why → Mechanics → Constraints & Edge Cases).
2. **Add Comments to Complex Code (When in `comprehensive` Mode)**:
   - For complex, non-obvious code blocks, add explanatory inline comments focusing on:
     - The rationale behind subtle logic.
     - Invariants that must be maintained across statements.
     - Why simpler or alternative approaches were not viable.
3. **Do Not Measure Quality by Length Alone**:
   - Brevity is valuable, but **never omit nuance or rationale for the sake of shortness**.
   - Adding clarifying explanations, concrete examples, or explicit warnings is a positive improvement whenever it aids understanding.
4. **Incorporate Every Inventoried Item**:
   - Verify that every item recorded in Step 1's inventory is clearly represented in the rewritten draft.

Verify: All existing context is conserved, complex code locations receive clarifying Why-first comments, and the revised documentation is logically structured.

---

## Step 3: Strictly Verify Zero Code Mutation

Ensure with complete certainty that executable project code, syntax, and runtime behavior were not altered.

1. **Codebase Integrity Checks**:
   - Run the project's format checkers, linters, and tests:
     - Formatters / linters (e.g. `npm run lint`, `golangci-lint`, `ruff`, `cargo check`).
     - Test suites (e.g. `npm test`, `go test ./...`, `pytest`, `cargo test`).
2. **Diff Inspection for Pure Documentation Changes**:
   - Inspect the diff carefully:
     - For standalone doc files: verify that only text and doc assets changed.
     - For source code comments: confirm that changes are strictly isolated to comment markers (e.g., `//`, `/* ... */`, `#`, docstrings) and that no executable statements, arguments, or function signatures were shifted or altered.

Verify: Linters and tests pass without regression, and diff inspection confirms zero mutation of executable code.

---

## Step 4: Verify Information Conservation & Present Findings

Perform a final audit against the information conservation checklist before concluding.

1. **Information Conservation Audit**:
   - Compare the final text side-by-side with the Step 1 inventory using [references/doc-refactor-guide.md](references/doc-refactor-guide.md):
     - Are all constraints, prerequisites, and warnings still clearly explained?
     - Are non-obvious design reasons and trade-offs preserved?
     - If detail was added, is it accurate and helpful?
2. **Present Report**:
   - Present the before/after comparison to the user, highlighting:
     - Files updated.
     - Key structural improvements made.
     - Confirmation of zero code mutation and complete information preservation.

Verify: All items in the information conservation checklist pass, and the report is presented to the user.

---

## Guidelines

- **Zero code mutation is absolute.** Never alter code logic, variable names, public signatures, or execution flow under this skill. If code refactoring is needed, recommend [`refactor`](../refactor/SKILL.md).
- **Never lose context or rationale.** Existing comments often document hard-won lessons, edge cases, and non-obvious workarounds. Preserve all such rationale without omission.
- **Quality is not text compression.** A longer explanation with clear context and examples is far superior to a terse summary that loses subtle caveats.
- **Focus on the "Why".** Comments that describe what the code obviously does are low-value; focus improvements on why the code does it that way.
