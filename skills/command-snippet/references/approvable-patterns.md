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

### Principle 4: Scope Granularity & Cognitive Review Ease (Principle of Least Privilege)
Review friction is directly tied to the ambiguity and blast radius of the command. When an agent presents a command using an omnipotent CLI tool (such as `gh`, `aws`, `gcloud`, or `kubectl`), human reviewers face high cognitive load: they must verify that no destructive subcommands, hidden options, or privilege escalation paths exist anywhere in the command string.

Constraining broad CLI tools into specialized snippet files minimizes capability scope to the exact operation needed, drastically lowering the cognitive threshold for human approval and enabling safe, continuous approval across sessions.

#### The 4-Tier Scope Granularity Model

| Tier | Granularity Level | Invocation Form | Review Burden & Approval Predictability |
|---|---|---|---|
| **Level 0** | **Raw Broad CLI** | `gh pr view 123` | **High Review Burden**: Approving `gh` grants blanket permissions to destructive capabilities (`gh repo delete`, etc.). Every single invocation must be audited end-to-end. |
| **Level 1** | **Subcommand-Constrained** | `./run_gh_pr_view.sh "$@"` | **Moderate Review Ease**: Locks the action to `pr view`, blocking destructive subcommands, but permits arbitrary flags and targets. Suitable for ad-hoc exploration. |
| **Level 2** | **Parameter-Constrained** | `./view_pr_details.sh "$1"` | **Low Review Friction (Continuous Approval)**: Subcommand and flags (e.g. `--json number,title,body`) are hardcoded. Reviewer only audits the trailing target `$1`. Optimal for repetitive inspection across items. |
| **Level 3** | **Fully Fixed Resource** | `./view_target_pr.sh` | **Zero Cognitive Friction**: Both the action and the target (e.g. PR #123) are curried inside the snippet. Zero blast radius, trivial approval. Ideal for single-resource tasks and mutating operations. |

#### Trade-off Analysis: Per-Approval Anxiety vs. Approval Fatigue

Constraining scope to Level 3 offers maximum safety and easiest per-step approval, but creating dedicated files for every single dynamic resource introduces **Approval Fatigue** (excessive approval gates and scratch file proliferation).

- **Use Level 2 (Parameter-Constrained)** when:
  - Performing **read-only queries across multiple items** (e.g., inspecting several PRs or commit SHAs during a single investigation).
  - A single snippet can be audited once and safely reused with trailing positional arguments.
  - Shell argument validation (`[[ "$1" =~ ^[0-9]+$ ]]`) guarantees the parameter format.
- **Use Level 3 (Fully Fixed Resource)** when:
  - Executing **state-mutating or irreversible operations** (e.g., commenting on or merging a specific PR, deploying a specific tag).
  - The entire workflow is known to target exactly one single item from the start.
  - Absolute confinement is required so the agent cannot inadvertently target an incorrect resource.

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

### Example C: Parameter-Constrained GitHub PR Viewer (Level 2)

```bash
#!/usr/bin/env bash
#
# Description:
#   Fetches metadata for a specific GitHub pull request within the current repository.
#   Restricts gh CLI execution strictly to read-only pr view operations.
#
# Arguments:
#   $1 - pr_number (numeric pull request ID)
#
# Usage:
#   ./view_pr_details.sh 123
#
# Prerequisites:
#   gh (GitHub CLI), authenticated
#
set -euo pipefail

pr_number="${1:?Error: PR number is required}"

if [[ ! "${pr_number}" =~ ^[0-9]+$ ]]; then
  echo "Error: PR number must be positive integer: ${pr_number}" >&2
  exit 1
fi

exec gh pr view "${pr_number}" --json number,title,state,author,url
```

### Example D: Fixed-Resource Targeted GitHub PR Viewer (Level 3)

```bash
#!/usr/bin/env bash
#
# Description:
#   Fetches details for target pull request #123.
#   Completely curries the target resource to eliminate dynamic parameter risks.
#
# Arguments:
#   None
#
# Usage:
#   ./view_target_pr_123.sh
#
# Prerequisites:
#   gh (GitHub CLI), authenticated
#
set -euo pipefail

readonly TARGET_PR="123"

exec gh pr view "${TARGET_PR}" --json number,title,state,body,url
```

---

## 5. Verification Checklist

Before presenting an approvable snippet for user execution:

- [ ] Does the snippet perform exactly one action (no dynamic subcommands)?
- [ ] Is the scope granularity tier (Level 1–3) chosen appropriately to minimize review friction without causing approval fatigue?
- [ ] Are omnipotent tools (e.g., `gh`, `aws`, `kubectl`) constrained to explicit, minimal subcommands?
- [ ] Are dynamic variables exposed exclusively as trailing CLI arguments (`$1`, `$2`), or curried internally (Level 3) when targeting a fixed resource?
- [ ] Are required parameters guarded with `${1:?message}` and format-validated (e.g., regex pattern check)?
- [ ] Is input validation in place for paths or required entities?
- [ ] Is the final command executed via `exec` with fully quoted variables?
- [ ] Has the snippet been audited via `snippet-review` and preceded by the mandatory Pre-Execution Notice?
