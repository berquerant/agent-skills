# Maintainability & Technical Debt Metrics

This reference defines criteria and heuristics for evaluating code maintainability,
identifying technical debt, and discovering refactoring opportunities.

---

## 1. Code Complexity & Size Thresholds

While thresholds vary by programming language, the following heuristics provide
general benchmarks for maintainability:

| Dimension | Normal / Good | Caution (Review) | Warning (Action Recommended) |
| :--- | :--- | :--- | :--- |
| **File Length** | < 300 lines | 300 – 600 lines | > 600 lines (consider splitting by concern) |
| **Function / Method Length** | < 40 lines | 40 – 80 lines | > 80 lines (break down into subroutines) |
| **Nesting Depth** | ≤ 3 levels | 4 levels | ≥ 5 levels (extract early returns, guard clauses) |
| **Function Parameters** | ≤ 4 parameters | 5 – 6 parameters | ≥ 7 parameters (use parameter object / config struct) |
| **Cyclomatic Complexity** | < 10 | 10 – 20 | > 20 (simplify conditionals, dispatch table) |

### Detection Tips
```sh
# Find large files (adapt extensions as needed):
find . -type f \( -name "*.go" -o -name "*.py" -o -name "*.ts" -o -name "*.rs" \) \
  -not -path "*/vendor/*" -not -path "*/node_modules/*" -not -path "*/.git/*" \
  -exec wc -l {} + | sort -rn | head -20
```

---

## 2. Technical Debt & Workaround Markers

Track intentional technical debt and unfinished implementations left in comments:

| Marker | Meaning & Impact |
| :--- | :--- |
| `TODO` | Incomplete feature, deferred implementation, or placeholder logic. |
| `FIXME` | Known bug, workaround, or broken edge-case awaiting correction. |
| `HACK` | Brittle solution, workaround for an upstream issue, or temporary fix. |
| `XXX` / `BUG` | Warning of severe hazard, dangerous assumption, or flaw. |
| `DEPRECATED` | Outdated API or function pending removal. |

### Detection Tips
```sh
# Search for tech debt markers excluding vendor/dependencies:
grep -rnE "(TODO|FIXME|HACK|XXX|DEPRECATED)" \
  --exclude-dir={.git,node_modules,vendor,dist,build,target} .
```

Categorize findings into:
1. **Critical Debt**: Unhandled errors, skipped security checks, or race conditions.
2. **Feature Gaps**: Missing optional capabilities or pending optimizations.
3. **Stale Notes**: Outdated markers whose underlying problem is already resolved.

---

## 3. Duplication & Coupling

- **Duplicated Logic (DRY Violations)**: Similar error handling boilerplate, duplicated validation blocks, or copy-pasted data mapping across services.
- **Tight Coupling & Monolith Modules**: A single module or utility file imported by nearly every other component, creating cascading changes.
- **Circular Dependencies**: Packages or modules mutually importing one another, impeding independent testing and isolation.

---

## 4. Testability & Test Coverage

Assess how easily the codebase can be modified without breaking:

- **Missing Test Suites**: Critical business logic or calculation modules lacking corresponding `*_test` or `test_*` files.
- **Untestable Patterns**:
  - Global mutable state or singletons.
  - Direct I/O or network calls embedded deep in business logic without interface/abstraction injection.
  - Non-deterministic behaviors (hardcoded current time, random generators) without mocking hooks.
- **Fragile Tests**: Tests relying on exact sleep timers, unpinned external network access, or rigid execution order.

---

## 5. Dead Code & Unused Artifacts

- **Unreachable Code**: Code paths following unconditional returns or exceptions.
- **Orphaned Files**: Scripts, obsolete configurations, or abandoned feature branches no longer referenced anywhere in the repository.
- **Unused Dependencies**: Packages declared in manifests (`package.json`, `go.mod`, `Cargo.toml`, `requirements.txt`) that are never imported.

---

## 6. Refactoring Prioritization Matrix

When reporting technical debt, prioritize candidates using this matrix:

| Impact on System | Low Effort (Quick Win) | High Effort (Strategic Project) |
| :--- | :--- | :--- |
| **High Impact** (Core logic, high churn) | **P0: Immediate Fix** (Guard clauses, extract helper) | **P1: Planned Refactor** (Modularization, decoupling via `refactor` skill) |
| **Low Impact** (Peripheral, rarely touched) | **P2: Opportunistic** (Clean up during adjacent work) | **P3: Accept Debt** (Document and defer) |
