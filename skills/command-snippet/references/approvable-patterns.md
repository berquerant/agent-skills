# Approvable Snippet Patterns: Logical Atomic Scope, Execution Idioms & Approvable Architecture

This reference details design patterns for creating shell and interpreter execution snippets optimized for **approval ease, whitelist predictability, high execution efficiency, and minimal cognitive friction**.

Direct execution of long commands with variable parameters embedded in the middle of command lines (e.g. `git -C <dir> status`), multi-step pipelines split into dozens of micro-scripts, or inefficient execution workflows (such as separate `chmod` turns or indiscriminate interpreter invocations) breaks approval rules, multiplies execution latency, and causes severe approval fatigue. Approvable snippet patterns eliminate this friction by currying fixed actions, scoping logical atomic transactions, right-aligning variable parameters, using self-descriptive filenames, and utilizing pinpointed execution idioms.

---

## 1. Problem Statement: Approval Friction, Execution Overhead & Granularity Dilemmas

In agent-human workflows or automated permission harnesses, commands are evaluated for execution approval using prefix matching, parameter inspection, or exact rule whitelists.

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

### The Trailing Target Argument & Persistent Read Dilemma
A symmetrical but inverted approval friction occurs with inspection commands such as `grep [options...] <pattern> <file>`:

```sh
# Target file is fixed, but options and patterns vary across investigation steps:
grep -i "error" /path/to/app.log
grep -n -C 5 "connection reset" /path/to/app.log
grep -E "(FATAL|CRITICAL)" /path/to/app.log
```

Here, the user or harness may explicitly decide: *"Reading `/path/to/app.log` is safe; once approved, the agent may search it freely throughout this investigation."*

However, standard CLI syntax places the target `<file>` at the very end of the command line, preceded by dynamic options and search patterns:
1. **Blanket Read Approval Hazard**: Granting a prefix-based approval on `grep` or `grep -i` is dangerous because it permits reading **any** arbitrary file on the filesystem (including `.env`, private keys, or `/etc/shadow`), destroying the principle of least privilege.
2. **Persistent Approval Impossibility**: Because the target file is buried at the tail of the command, prefix matching cannot recognize that all these queries target the same pre-approved file. Every variation of flags (`-C 5`, `-E`) and patterns triggers a new approval gate, causing severe **approval fatigue** and interrupting the agent's investigation loop.
3. **The Solution (Currying the Target File)**: By currying the target file path inside a dedicated snippet (e.g. `./search_app_log.sh "pattern" [options...]`), the snippet invocation prefix (`./search_app_log.sh`) becomes stable. A single user approval enables persistent, safe inspection of that specific file without exposing the rest of the filesystem.

### The Execution Authorization Dilemma: Generic Interpreters vs. Standalone `chmod`
Two opposing failure modes frequently disrupt snippet execution:
1. **The Generic Interpreter Approval Trap**:
   Invoking snippets via generic interpreters (e.g., `bash path/to/script.sh` or `python3 path/to/script.py`) places `bash` or `python3` as the command binary (`argv[0]`). In environments evaluating execution authorization by command prefix or binary, granting execution permission to `bash` or `python3` creates a dangerous **blanket authorization hazard**—it permits executing arbitrary shell one-liners (`bash -c "rm -rf ..."`) or arbitrary code scripts across the system.
2. **The Standalone `chmod` Round-Trip Waste**:
   To make scripts executable, agents often execute a standalone `chmod +x <file>` in a separate turn, followed by `<file>` in the next turn. This wastes an entire agent turn, doubles execution latency, and produces redundant approval prompts.
3. **The Solution (Pinpointed Script-Path Prefix & Atomic Chaining)**:
   Keep the script path itself as the command prefix (`/path/to/script.sh [args...]` or `./script.sh [args...]`) so authorization remains strictly confined to that specific, immutable snippet file. If execute permissions must be set, chain them into a single atomic command line:
   ```sh
   chmod +x "/path/to/script.sh" && exec "/path/to/script.sh" [args...]
   ```
   This executes in a single step without extra interaction turns while preserving strict script-level authorization confinement.

### The Granularity Dilemma: Single Command Wrapping vs. Atomic Task Workflows
When applying single responsibility, a frequent mistake is extreme micro-fragmentation: creating a separate snippet for every single command (e.g., separate snippets for `fetch_prs.sh`, `setup_worktree.sh`, `rebase.sh`, `test.sh`, `lint.sh`, `commit.sh`, `push.sh`, `merge.sh`, `cleanup.sh`).
- **Approval Fatigue & Fragility**: Generating 10+ snippets for a single task forces the reviewer through 10+ separate Pre-Execution Notices and approval gates. Furthermore, if intermediate checks are not atomically chained, the environment is left in a dirty, inconsistent intermediate state upon failure.
- **The Solution (Logical Atomic Task Stage)**: Group deterministic, unbranching sequences of commands that form a single logical milestone (e.g. `rebase_and_test_worktree.sh`) into one atomic snippet protected by `set -euo pipefail`. If any step fails, the entire transaction stops immediately without side-effect escalation, while requiring only a single human review gate.

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

### Principle 2: Logical Atomic Responsibility & Self-Descriptive Naming
A snippet must perform **one logical atomic task stage**.

- **Anti-Pattern 1: Generalized Multi-Action Wrapper**:
  ```bash
  # DANGEROUS: Privilege escalation hazard
  #!/usr/bin/env bash
  git -C "$1" "$2"
  ```
  Passing arbitrary subcommands or actions dynamically re-introduces blanket execution hazards. Never pass action verbs as parameters.

- **Anti-Pattern 2: Micro-Script Fragmentation (Approval Exhaustion)**:
  Creating 10 separate 1-line scripts for a deterministic 10-step build/verify pipeline.

- **Best Practice: Atomic Task Stage Pipeline**:
  Encapsulate a complete, deterministic workflow stage (e.g., run code generators -> update licenses -> run test suite -> run linter) in one script. If any command fails, `set -e` halts execution cleanly.

- **Naming Rule: Self-Descriptive Snippet Filenames**:
  Every snippet must be named so that human reviewers can instantly deduce its exact scope, action, and target from the filename alone (e.g. `rebase_and_test_worktree.sh` or `search_app_log.sh`). Never use vague or ambiguous names (e.g. `run.sh`, `helper.sh`, `resolve_and_verify.sh`) that fail to convey functionality or could be confused with existing local project utilities.

### Principle 3: Universal Directory Context Encapsulation
Do not rely on tool-specific directory flags (e.g. `git -C <dir>` works, but `gh -C <dir>` or `make -C <dir>` may be unsupported or behave inconsistently across platforms).
Standardize directory scoping inside snippets using explicit, safe directory navigation:

```bash
target_dir="${1:?Error: target directory required}"
if [[ ! -d "${target_dir}" ]]; then
  echo "Error: Directory not found: ${target_dir}" >&2
  exit 1
fi
cd "${target_dir}"
```

This guarantees that subsequent commands (`gh`, `go`, `make`, `npm`, `cargo`) execute within the intended workspace context cleanly.

### Principle 4: Pinpointed Script Authorization & Atomic Execution
To prevent granting blanket execution rights to generic shells or interpreters (`bash` or `python3`), invoke the snippet directly via its script path:
- Direct script invocation: `./script.sh "$@"` or `/path/to/script.sh "$@"`
- Atomic single-step invocation when setting execute bits:
  ```sh
  chmod +x "/path/to/script.sh" && exec "/path/to/script.sh" "$@"
  ```

This locks the authorization boundary to the exact script file path without leaking blanket execution privileges to `bash` or wasting turns on standalone `chmod` steps.

### Principle 5: Context-Dependent Variability Analysis
What constitutes a "variable argument" is determined by the **task context**:

- **Multi-Repository Operations**: When auditing multiple repositories, repository root path is variable (`$1`), while the query (`status`, `log -1`) is fixed.
- **Multi-Branch Operations**: When inspecting branch heads, directory is fixed, but branch ref is variable (`$1`).
- **Target-Focused File/Log Inspection**: When iteratively searching a specific log file (`app.log`), the target file path is fixed inside the snippet, while search patterns and filter flags (`-i`, `-C 3`) are variable (`$1`, `$@`).
- **Resource Actions**: When operating on a specific resource (e.g., PR #123), curry the resource ID inside a Level 3 snippet to eliminate parameter injection risks.

### Principle 6: Scope Granularity & Cognitive Review Ease (4-Tier Model)

| Tier | Granularity Level | Invocation Form | Review Burden & Approval Predictability |
|---|---|---|---|
| **Level 0** | **Raw Broad CLI** | `gh pr view 123`<br/>`grep -i "err" /path/to/app.log` | **High Review Burden**: Approving raw commands grants blanket permissions (destructive actions or arbitrary filesystem read). Every single invocation must be audited end-to-end. |
| **Level 1** | **Subcommand-Constrained** | `./run_gh_pr_view.sh "$@"`<br/>`./run_grep.sh "$@"` | **Moderate Review Ease**: Locks the action/binary, blocking destructive subcommands, but permits arbitrary flags and target files. Suitable for ad-hoc exploration. |
| **Level 2** | **Parameter-Constrained** | `./view_pr_details.sh "$1"`<br/>`./search_logs_dir.sh "$1"` | **Low Review Friction (Continuous Approval)**: Subcommand and search scope are hardcoded. Reviewer only audits the trailing target `$1`. Optimal for repetitive inspection across items. |
| **Level 3** | **Fully Fixed Resource** | `./view_target_pr.sh`<br/>`./search_app_log.sh "pattern" [opts]` | **Zero Cognitive Friction & Persistent Approval**: Both action and target resource (PR #123 or `app.log`) are curried inside the snippet. Zero blast radius, trivial approval. Ideal for single-resource tasks, mutating operations, or persistent read access to a specific file. |

---

## 3. Robust Shell Implementation Guidelines

1. **Safety Flags**: Begin with `set -euo pipefail`.
2. **Strict Parameter Validation**: Guard required arguments using parameter expansion error syntax: `"${1:?Error message}"`.
3. **Precondition Checks**: Verify target paths or dependencies exist before invoking core commands.
4. **Clean Execution**: Use `exec` for single terminal commands to replace the subshell process, or sequential execution under `set -e` for atomic pipelines.
5. **Exact Quoting**: Quote all variable references (`"${var}"`) to avoid word splitting and glob expansion.

---

## 4. Standard Approvable Templates

### Example A: Directory-Targeted Git Query
Encapsulates directory validation and execution with right-aligned repository path.

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

---

### Example B: Universal Directory Context Encapsulation (`cd` idiom)
Executes CLI tools (such as `gh`, `make`, `npm`) that do not uniformly support `-C <dir>` flags by safely changing directory internally.

```bash
#!/usr/bin/env bash
#
# Description:
#   Fetch open pull requests metadata from GitHub CLI within a target repository directory.
#
# Arguments:
#   $1 - repo_dir (path to local repository root)
#
# Usage:
#   ./fetch_repo_prs.sh path/to/repo
#
# Prerequisites:
#   gh (GitHub CLI, authenticated)
#
set -euo pipefail

repo_dir="${1:?Error: repository directory is required}"

if [[ ! -d "${repo_dir}" ]]; then
  echo "Error: Repository directory not found: ${repo_dir}" >&2
  exit 1
fi

cd "${repo_dir}"
exec gh pr list --state open --json number,title,headRefName,author,labels,files,body
```

---

### Example C: Atomic Task Workflow Pipeline (Composite Transaction)
Encapsulates a multi-step, deterministic workspace task (rebase -> generate artifacts -> verify -> commit) into a single atomic stage with a self-descriptive filename, eliminating micro-script proliferation.

```bash
#!/usr/bin/env bash
#
# Description:
#   Resolves dependencies, updates third-party notices, commits changes if dirty,
#   and runs test and lint suites in an isolated worktree.
#
# Arguments:
#   $1 - worktree_dir (path to target git worktree)
#
# Usage:
#   ./rebase_and_test_worktree.sh path/to/worktree
#
# Prerequisites:
#   go, make, git
#
set -euo pipefail

worktree_dir="${1:?Error: worktree directory is required}"

if [[ ! -d "${worktree_dir}" ]]; then
  echo "Error: Worktree directory not found: ${worktree_dir}" >&2
  exit 1
fi

cd "${worktree_dir}"

echo "=== Step 1: Tidying Go dependencies ==="
go mod tidy
git add go.mod go.sum

echo "=== Step 2: Updating third-party notice ==="
./hack/license.sh report > NOTICE
if ! git diff --quiet NOTICE; then
  echo "NOTICE modified; staging and committing..."
  git add NOTICE
  git commit -m "build(deps): update NOTICE for dependency update"
else
  echo "NOTICE is up to date."
fi

echo "=== Step 3: Running test suite ==="
make test

echo "=== Step 4: Running linter ==="
make lint

echo "=== Worktree resolution and verification successful ==="
```

---

### Example D: Fixed-File Search Snippet (Persistent Read Approval for `grep`)
Curries the target file path internally so that users or automated harnesses can grant persistent read approval via prefix matching (`./search_app_log.sh`) without granting blanket filesystem read access.

```bash
#!/usr/bin/env bash
#
# Description:
#   Searches specifically within the target application log file.
#   Curries the target file path inside the snippet so users can grant
#   persistent read approval via prefix matching (./search_app_log.sh)
#   without exposing arbitrary filesystem read access via raw grep commands.
#
# Arguments:
#   $1 - pattern    (regular expression or string to search)
#   $@ - extra_opts (optional grep flags, e.g., -i, -n, -C 3, -m 50)
#
# Usage:
#   ./search_app_log.sh "ERROR"
#   ./search_app_log.sh "Exception" -i -C 3
#
# Prerequisites:
#   grep
#
set -euo pipefail

readonly TARGET_FILE="/var/log/app.log"
pattern="${1:?Error: search pattern is required}"
shift

if [[ ! -f "${TARGET_FILE}" ]]; then
  echo "Error: Target file not found: ${TARGET_FILE}" >&2
  exit 1
fi

exec grep "$@" -e "${pattern}" "${TARGET_FILE}"
```

---

### Example E: Self-Contained Structured Data Pipeline (Python / JSON)
Fetches and processes structured data (e.g. GitHub PRs or container statuses) in a single self-contained script with a self-descriptive filename, avoiding multi-file pipeline fragmentation (`list.sh` -> `temp.json` -> `parse.py`).

```python
#!/usr/bin/env python3
"""
Description:
  Fetches open pull requests matching dependency labels and filters candidate PRs
  ready for automated verification and merge. Outputs structured summary to stdout.

Arguments:
  sys.argv[1] - repo_dir (path to local repository root)

Usage:
  ./filter_candidate_prs.py /path/to/repo

Prerequisites:
  gh CLI (authenticated), python3
"""

import json
import subprocess
import sys

def main():
    if len(sys.argv) < 2:
        print("Error: Target repository directory required", file=sys.stderr)
        sys.exit(1)

    repo_dir = sys.argv[1]
    cmd = [
        "gh", "pr", "list",
        "--state", "open",
        "--json", "number,title,headRefName,labels,statusCheckRollup"
    ]

    try:
        res = subprocess.run(cmd, cwd=repo_dir, capture_output=True, text=True, check=True)
        prs = json.loads(res.stdout)
    except subprocess.CalledProcessError as e:
        print(f"Error executing gh: {e.stderr}", file=sys.stderr)
        sys.exit(e.returncode)

    candidates = []
    for pr in prs:
        labels = [l.get("name", "") for l in pr.get("labels", [])]
        if any("dependencies" in l.lower() or "renovate" in l.lower() for l in labels):
            candidates.append({
                "number": pr["number"],
                "title": pr["title"],
                "branch": pr["headRefName"]
            })

    print(json.dumps(candidates, indent=2))

if __name__ == "__main__":
    main()
```

---

### Example F: Parameter-Constrained GitHub PR Viewer (Level 2)
Restricts `gh` CLI execution strictly to read-only `pr view` operations with parameter format validation.

```bash
#!/usr/bin/env bash
#
# Description:
#   Fetches metadata for a specific GitHub pull request within the target repository.
#   Restricts gh CLI execution strictly to read-only pr view operations.
#
# Arguments:
#   $1 - target_repo_dir (path to repository)
#   $2 - pr_number       (numeric pull request ID)
#
# Usage:
#   ./view_pr_details.sh path/to/repo 123
#
# Prerequisites:
#   gh (GitHub CLI, authenticated)
#
set -euo pipefail

repo_dir="${1:?Error: repository path is required}"
pr_number="${2:?Error: PR number is required}"

if [[ ! -d "${repo_dir}" ]]; then
  echo "Error: Directory not found: ${repo_dir}" >&2
  exit 1
fi

if [[ ! "${pr_number}" =~ ^[0-9]+$ ]]; then
  echo "Error: PR number must be positive integer: ${pr_number}" >&2
  exit 1
fi

cd "${repo_dir}"
exec gh pr view "${pr_number}" --json number,title,state,author,url
```

---

### Example G: Fixed-Resource Targeted GitHub PR Viewer (Level 3)
Completely curries the target resource to eliminate dynamic parameter risks.

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
#   gh (GitHub CLI, authenticated)
#
set -euo pipefail

readonly TARGET_PR="123"

exec gh pr view "${TARGET_PR}" --json number,title,state,body,url
```

---

## 5. Verification Checklist

Before presenting an approvable snippet for user execution:

- [ ] Does the snippet represent a **single logical atomic task stage** (no interactive branching or dynamic subcommand wrappers)?
- [ ] Is the snippet named **self-descriptively** (e.g., `rebase_and_test_worktree.sh`), avoiding vague or generic names?
- [ ] Is the invocation executed with the **script path as the command binary** (or atomic `chmod +x <path> && exec <path> [args...]`) to bind authorization specifically to the snippet rather than granting blanket permission to generic interpreters?
- [ ] Has any standalone `chmod +x` tool call in a separate turn been eliminated?
- [ ] Is the scope granularity tier (Level 1–3) chosen appropriately to minimize review friction without causing approval fatigue?
- [ ] Are omnipotent tools (e.g., `gh`, `aws`, `kubectl`) constrained to explicit, minimal subcommands?
- [ ] For directory-dependent tools, is directory navigation safely encapsulated (`cd "${target_dir:?}"`) inside the script?
- [ ] For repetitive file inspections (e.g., `grep`), is the target file curried internally (Level 3) to enable persistent read approval without broad filesystem read permissions?
- [ ] Are dynamic variables exposed exclusively as trailing CLI arguments (`$1`, `$2`), or curried internally (Level 3) when targeting a fixed resource?
- [ ] Are required parameters guarded with `${1:?message}` and format-validated (e.g., regex pattern check)?
- [ ] Is input validation in place for paths or required entities?
- [ ] Is the final command executed cleanly via `exec` (for single commands) or protected by `set -euo pipefail` (for multi-step pipelines)?
- [ ] Has the snippet been audited via `snippet-review` and preceded by the mandatory Pre-Execution Notice?
