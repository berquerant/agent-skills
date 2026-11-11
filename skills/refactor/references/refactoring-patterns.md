# Behavior-Preserving Refactoring Patterns

This reference summarizes common structural refactoring patterns and safety invariants used by `refactor`.

---

## 1. Core Refactoring Catalog

| Pattern | When to Apply | Safety Invariant |
|---|---|---|
| **Extract Function / Method** | A code block exceeds ~25 lines or requires a comment to explain what it computes | All local variables passed/returned explicitly; unit tests unchanged |
| **Inline Trivial Indirection** | A helper wrapper adds no abstraction value and is called from a single site | Caller behavior and error propagation remain identical |
| **Consolidate Duplicate Logic (Pragmatic DRY)** | Two or more locations implement the exact same domain rule | Shared helper is parameterized only on genuine variance; avoid premature abstraction |
| **Replace Nested Conditionals with Guard Clauses** | Deep `if/else` nesting ($\ge 3$ levels) obscures the normal execution path | Early-return conditions preserve exact side-effect ordering |
| **Split Overloaded Module / File** | A single file mixes unrelated responsibilities (e.g., transport, business logic, persistence) | Public exports or re-exports preserve existing caller imports where required |

---

## 2. Atomic Verification Loop

1. Run baseline test and lint suites (`0` failures required before starting).
2. Apply **one** structural pattern from the catalog above.
3. Re-run test and lint suites immediately.
4. Inspect `git diff` to confirm zero unintended behavioral or API changes before proceeding to the next transformation.
