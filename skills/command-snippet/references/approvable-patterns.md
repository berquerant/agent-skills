# Approvable Snippet Patterns: Trailing Variable Arguments & Single Responsibility

This reference details design patterns for creating shell execution snippets optimized for **approval ease, whitelist predictability, and minimal cognitive friction**.

Direct execution of commands with variable parameters embedded in the middle of command lines (e.g. `git -C <dir> status`) breaks whitelist prefix rules and obscures critical actions during human review. Approvable snippet patterns eliminate this friction by currying fixed actions and exposing variable parameters exclusively at the trailing position.

---

## 1. Problem Statement: Infix Arguments & Approval Friction

In agent-human workflows or automated permission harnesses, commands are often evaluated for execution approval using prefix matching or exact rule whitelists.

### The Infix Argument Problem
Commands such as `git -C <dir> <subcommand>` or `ssh <host> <command>` place dynamic context variables in the middle of the command line:

```sh
# Directory varies per invocation, but the action ('status') is fixed:
git -C /path/to/repo-a status
git -C /path/to/repo-b status
```

This structure causes two distinct failures:
1. **Approval Rule Invalidation**: A user or security tool wishing to grant permanent approval for `git status` cannot rely on simple prefix rules because the dynamic directory path splits the command (`git` ... `status`).
2. **Cognitive Audit Fatigue**: When human reviewers inspect commands with long paths or complex options, the critical operation (e.g., whether the command is a read-only `status` or a destructive `reset --hard`) is pushed to the far right or buried inside flags, increasing review error rates.

---

## 2. Core Design Principles

### Principle 1: Trailing Variable Arguments (Right-Aligning Dynamic Inputs)
Encapsulate the fixed command, subcommands, and static flags within an immutable snippet file. Expose variable arguments solely as explicit positional CLI arguments at the end of the invocation:

```sh
# Snippet invocation:
./run_git_status.sh "/path/to/repo-a"
```

- The executable path (`./run_git_status.sh`) becomes a stable, whitelistable prefix.
- The human reviewer immediately sees that the action is `status` and needs only to audit the right-aligned target directory (`$1`).

### Principle 2: Single Responsibility Principle (One Action Per Snippet)
Every approvable snippet must perform **exactly one concrete action**.

Do **NOT** generalize snippets by passing subcommands or actions as secondary arguments:

```bash
# ANTI-PATTERN: Generalized multi-purpose wrapper
#!/usr/bin/env bash
git -C "$1" "$2"  # Dangerous! Privilege escalation hazard.
```

Passing the subcommand dynamically re-introduces the original vulnerability: approving the snippet gives blanket permission to run *any* subcommand (including destructive actions like `clean -fdx` or `push --force`).

Instead, generate separate, dedicated snippets for distinct actions:
- `run_git_status.sh "$1"` — specifically executes `git -C "$1" status`
- `run_git_tag.sh "$1"` — specifically executes `git -C "$1" tag`

### Principle 3: Context-Dependent Variability Analysis
What constitutes a "variable argument" is not fixed by the command's syntax alone; it is determined by the **task context**:

- **Multi-Repository Operations**: When auditing or reporting status across multiple repositories, the repository path (`-C <dir>`) is variable, while the query (`status`, `rev-parse`, `log -1`) is fixed.
- **Multi-Branch Operations**: When inspecting branch heads within a single repository, the directory is fixed, but the branch reference is variable (`git show "$1"`).
- **Remote Host Operations**: When inspecting multiple remote hosts, the host identifier in `ssh "$1" systemctl status service` is variable, while the inspection command is fixed.

Evaluate the execution pattern across the overall workflow before deciding which inputs to parameterize.

---

## 3. Robust Shell Implementation Guidelines

Approvable snippets must not sacrifice robustness for brevity. Always follow standard shell hardening practices:

1. **Safety Flags**: Begin with `set -euo pipefail`.
2. **Strict Parameter Validation**: Guard required arguments using parameter expansion error syntax: `"${1:?Error message}"`.
3. **Precondition Checks**: Verify that target paths exist and match expected types before invoking the core command.
4. **Process Replacement (`exec`)**: Use `exec` on the final command to replace the subshell process, cleanly propagating exit codes and signals without lingering wrapper processes.
5. **Exact Quoting**: Quote all variable references (`"${var}"`) to avoid word splitting and glob expansion.

---

## 4. Standard Templates

### Example A: Directory-Targeted Git Query (`git -C`)

```bash
#!/usr/bin/env bash
#
# Description:
#   Runs read-only git status for a specified target repository.
#
# Arguments:
#   $1 - target_directory (path to repository root)
#
# Usage:
#   ./run_git_status.sh path/to/repo
#
# Prerequisites:
#   git
#
set -euo pipefail

target_dir="${1:?Error: target repository path is required}"

if [[ ! -d "${target_dir}" ]]; then
  echo "Error: Directory not found: ${target_dir}" >&2
  exit 1
fi

exec git -C "${target_dir}" status
```

### Example B: Directory-Targeted Git Tag Listing

```bash
#!/usr/bin/env bash
#
# Description:
#   Lists tags in the specified target repository.
#
# Arguments:
#   $1 - target_directory (path to repository root)
#
# Usage:
#   ./run_git_tag.sh path/to/repo
#
# Prerequisites:
#   git
#
set -euo pipefail

target_dir="${1:?Error: target repository path is required}"

if [[ ! -d "${target_dir}" ]]; then
  echo "Error: Directory not found: ${target_dir}" >&2
  exit 1
fi

exec git -C "${target_dir}" tag
```

---

## 5. Verification Checklist

Before presenting an approvable snippet for user execution:

- [ ] Does the snippet perform exactly one action (no dynamic subcommands)?
- [ ] Are dynamic variables exposed exclusively as trailing CLI arguments (`$1`, `$2`)?
- [ ] Are required parameters guarded with `${1:?message}`?
- [ ] Is input validation in place for paths or required entities?
- [ ] Is the final command executed via `exec` with fully quoted variables?
- [ ] Has the snippet been audited via `snippet-review` and preceded by the mandatory Pre-Execution Notice?
