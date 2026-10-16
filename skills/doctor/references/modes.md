# Doctor Modes

`doctor` has two independent mode axes. Choose one value on each, giving four combinations.

---

## 1. Target Axis

| Mode | Scope |
|---|---|
| `project` | Every tracked source, configuration, and documentation file, excluding vendored and generated directories (for example `vendor/`, `node_modules/`, `dist/`, `build/`, `target/`). |
| `recent` | Files touched by recent changes, plus the uncommitted working tree. |

### Scope rules for `recent`

Combine these, in order, and state which one applied:

1. Uncommitted changes: `git status --short`, `git diff HEAD`.
2. On a topic branch: the diff against the default branch, as described in Step 1 of [`code-review`](../../code-review/SKILL.md) (for example `git diff origin/main...HEAD`).
3. On the default branch with a clean tree: the commit window used by [`project-audit`](../../project-audit/SKILL.md) Step 1.2, which is the larger of the last 20 commits or the last 1 hour.

For `recent`, read the touched files in full where needed to judge context, but report findings only on lines in or near the change unless a related problem is clearly caused by the change.

---

## 2. Depth Axis

| Mode | Coverage | Checks |
|---|---|---|
| `quick` | Sample the most important files: entry points, files with the largest or most frequent changes, and public interfaces. | Grep-based pattern searches and a read of the sampled files. No build or test runs. Report only the top findings per axis. |
| `thorough` | Read every file in scope. | Everything in `quick`, plus the project's lint, type-check, test, and vulnerability tools when available. Examine each finding's surrounding code and callers. |

Each depth must declare its limits in the report:
- `quick`: list the files that were not read and the tools that were not run.
- `thorough`: list any tools that were unavailable or failed, and the checks done by hand instead.

---

## 3. Defaults and Combinations

- If the user gives no preference, propose `recent` × `quick` and ask for confirmation.
- `project` × `quick`: a rough overview of a large project. Expect sampling gaps.
- `recent` × `thorough`: a careful review before merging or releasing a change.
- `project` × `thorough`: the most expensive combination. Before starting, tell the user about the size of the scope (file count) and ask for confirmation if it is large.
