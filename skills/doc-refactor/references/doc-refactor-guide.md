# Documentation & Code Comment Refactoring Guide

This document provides checklists, inventory templates, and zero-code-diff verification guidelines for agents operating under `doc-refactor`.

---

## 1. Information Inventory Template

Before modifying existing comments or documentation in Step 1, catalog all context using this template:

```markdown
### 📝 Target Information Inventory: `<path/to/file.ext#Lxx-Lyy>`

#### 1. Core Purpose & Intent
- [ ] <What is the stated high-level goal of this component/doc?>

#### 2. Non-Obvious Rationale ("Why")
- [ ] <Why was this specific algorithm, structure, or workaround chosen?>
- [ ] <What historical problem, bug, or limitation does this address?>

#### 3. Invariants & Prerequisites
- [ ] <What assumptions must hold true (e.g. locks held, environment variables, ordering)?>

#### 4. Warnings, Edge Cases & Failure Modes
- [ ] <What breaks if misused?>
- [ ] <What performance or security considerations are noted?>
```

---

## 2. Information Conservation Checklist

Before presenting changes in Step 4, audit the rewritten text against this checklist:

- [ ] **No Context Loss**: Does the rewritten explanation contain every item identified in the initial Information Inventory?
- [ ] **No Over-Compression**: Was any subtle nuance or warning discarded purely to make the paragraph shorter?
- [ ] **Added Value**: If new details or examples were added, do they directly clarify ambiguities without inventing non-existent functionality?
- [ ] **Accuracy**: Does the description accurately reflect the actual code reality in the repository?
- [ ] **Logical Hierarchy**: Is the explanation structured logically (Summary/Intent first, followed by Rationale, Mechanics, and Constraints)?

---

## 3. Zero-Code-Diff Verification Checklist

When editing inline comments in source code files, confirm zero mutation to executable instructions:

1. **Syntactic Cleanliness**:
   - Verify changes are confined to comment syntax (e.g., `#`, `//`, `/* ... */`, docstrings `"""`).
   - Confirm no indentation shifts on adjacent executable lines.
2. **Build & Test Parity**:
   - Run the project's build, test, and typecheck commands.
   - All tests must pass with identical status.
3. **AST / Semantic Invariance**:
   - Verify that compiled binaries or interpreted outputs produce zero functional divergence.

---

## 4. Complex Code Commenting Heuristics

When operating in `comprehensive` mode and identifying code blocks that warrant new explanatory comments:

### Where to Add Comments
- **Non-Obvious Branching**: Edge-case guards whose failure condition is counter-intuitive.
- **Complex Expressions**: Regex patterns, bitwise masks, pointer arithmetic, or multi-condition boolean logic.
- **Hidden Invariants**: Ordering dependencies (e.g. "Lock A must precede Lock B", "Cleaned up by defer/finally").
- **Workarounds**: Quirks working around third-party bugs or platform-specific behaviors.

### Comment Construction Template (Why-first)
```
// Rationale: <Why this specific check/algorithm is used instead of the naive alternative>
// Invariant: <What condition must be guaranteed before or after this block>
// Warning/Caveat: <Non-obvious side effect or edge case to be aware of>
```
