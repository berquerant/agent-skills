# Approvable Snippet Patterns: Logical Atomic Scope, Execution Idioms & Approvable Architecture

This reference details design patterns for creating shell and interpreter execution snippets optimized for **approval ease, whitelist predictability, high execution efficiency, and minimal cognitive friction**.

Direct execution of long commands with variable parameters embedded in the middle of command lines (e.g. `git -C <dir> status`), multi-step pipelines split into dozens of micro-scripts, or inefficient execution workflows (such as compound `chmod +x ... && exec ...` lines or generic interpreter prefixes like `bash <script>`) breaks approval rules, multiplies execution latency, and causes severe approval fatigue. Approvable snippet patterns eliminate this friction by currying fixed actions, scoping logical atomic transactions, currying target collections into zero-argument batch monitors, using self-descriptive filenames, and executing snippets directly as standalone command binaries (`./script.sh` or `/path/to/script.sh`).

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

### The Execution Authorization Dilemma: Why Interpreter Prefixes and Compound Chmod Fail
Two prevalent execution anti-patterns disrupt snippet whitelisting and trigger repetitive approval friction:

1. **The Generic Interpreter Prefix Trap (`bash script.sh`)**:
   Invoking snippets through an interpreter prefix (e.g. `bash path/to/script.sh` or `python3 path/to/script.py`) places `bash` or `python3` as the command binary (`argv[0]`). In whitelist and approval engines, this prompts the user to grant execution permission to `bash` or `python3` itself. Granting blanket permission to an interpreter is a dangerous security hole (allowing arbitrary one-liners like `bash -c "rm -rf ..."`), while denying blanket permission forces the user to re-approve every single invocation.
2. **The Compound Chmod Trap (`chmod +x ... && exec ...`)**:
   Chaining `chmod` and execution into a compound line makes `chmod` the leading command verb and creates compound syntax (`&&`) that whitelist engines cannot match to stable script prefixes. Furthermore, because script paths vary, this triggers a fresh approval prompt on every execution.
3. **The Solution (Direct Script-Binary Invocation)**:
   Invoke the snippet directly using its own script path:
   ```sh
   ./script.sh [args...]
   /path/to/script.sh [args...]
   ```
   Ensure execute permissions (`+x`) are set upon saving the file. In the execution turn, run **only** the script path itself without `bash` or `chmod` compound lines. This places the specific, immutable snippet file as `argv[0]`, enabling exact, pinpointed script authorization without blanket interpreter permissions or repetitive approval prompts.

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

### Principle 4: Direct Script-Binary Invocation (Zero Interpreter / Chmod Compound)
To ensure the command line cleanly represents the snippet itself and eliminates repetitive approval prompts:
- Execute snippets directly as standalone command binaries:
  - `./script.sh "$@"` or `/path/to/script.sh "$@"`
  - `./script.py "$@"` or `/path/to/script.py "$@"` (relying on shebang `#!/usr/bin/env python3`)
- **Prohibit Generic Interpreter Prefixes**: Do not prepend `bash` or `python3` to the execution command, which would request permission for the interpreter rather than the script.
- **Prohibit Compound `chmod +x ... && exec ...` Lines**: Set file execute permissions upon saving. Keep the execution turn clean and unchained so whitelist engines can match the script path prefix cleanly.

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
| **Level 3** | **Fully Fixed Resource or Batch** | `./view_target_pr.sh`<br/>`./search_app_log.sh "pattern" [opts]`<br/>`./check_active_prs_status.sh` | **Zero Cognitive Friction & Invariant Approval**: Actions and targets (single resource or multi-target batch) are completely curried inside the snippet. Zero dynamic parameter drift, trivial approval. Ideal for single-resource tasks, mutating operations, persistent read access, or repetitive polling of active batches. |

### Principle 7: Zero-Argument Batch Monitor Pattern (Currying Multi-Target Batches)
When monitoring or polling status across multiple specific targets in parallel workflows (e.g., tracking remote CI progress across 3 concurrently rebased PRs, checking health of 4 local services, or verifying multiple worktree builds):
- **The Problem with Parameterized Polling**:
  Passing target IDs as arguments (e.g., `./check_prs.sh 380 398 395`) exposes dynamic parameters on every call. Reviewers must check for argument drift, and any variation in target list order breaks exact-match approval rules. Furthermore, running separate single-target queries (`gh pr checks 380`, `gh pr checks 398`, `gh pr checks 395`) multiplies command turns and causes severe approval fatigue.
- **The Solution (Currying the Target Collection)**:
  Hardcode the active target collection (e.g., `TARGET_PRS=(380 398 395)`) inside the snippet as a constant array and expose **zero arguments**:
  ```sh
  ./check_active_prs_status.sh
  ```
- **Key Advantages**:
  1. **Invariant Command Line**: Because there are zero dynamic arguments, the command line is completely static. Whitelist engines and human reviewers can approve it once, allowing frictionless re-polling throughout the entire wait phase.
  2. **Zero Target-Drift / Typo Risk**: Confinement guarantees the script cannot inadvertently query or mutate unintended targets.
  3. **Batch Output Aggregation**: One single command turn collects and summarizes status for all targets simultaneously, drastically reducing latency and token churn.

---

## 3. Robust Shell Implementation Guidelines

1. **Safety Flags**: Begin with `set -euo pipefail`.
2. **Strict Parameter Validation**: Guard required arguments using parameter expansion error syntax: `"${1:?Error message}"` (when parameters are accepted).
3. **Precondition Checks**: Verify target paths or dependencies exist before invoking core commands.
4. **Clean Execution**: Use `exec` for single terminal commands to replace the subshell process, or sequential execution under `set -e` for atomic pipelines and batch loops.
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

### Example H: Zero-Argument Batch PR Status Monitor (Level 3 Multi-Target Batch)
Curries a specific batch of active pull requests into a zero-argument monitoring snippet. Enables invariant, frictionless repetitive status polling across multiple concurrent PRs under approval rules without parameter drift.

```bash
#!/usr/bin/env bash
#
# Description:
#   Monitors and summarizes CI status for active concurrent pull requests (#380, #398, #395).
#   Currying target PRs into a zero-argument snippet guarantees invariant command-line
#   invocations for frictionless polling under approval whitelists.
#
# Arguments:
#   None
#
# Usage:
#   ./check_active_prs_status.sh
#
# Prerequisites:
#   gh (GitHub CLI, authenticated)
#
set -euo pipefail

# Fixed active batch under verification
readonly TARGET_PRS=(380 398 395)

echo "=== Active Batch CI Status ($(date '+%Y-%m-%d %H:%M:%S')) ==="
for pr in "${TARGET_PRS[@]}"; do
  echo "--- PR #${pr} ---"
  gh pr checks "${pr}" || true
done
echo "=== Check complete ==="
```

---

## 5. Verification Checklist

Before presenting an approvable snippet for user execution:

- [ ] Does the snippet represent a **single logical atomic task stage** (no interactive branching or dynamic subcommand wrappers)?
- [ ] Is the snippet named **self-descriptively** (e.g., `rebase_and_test_worktree.sh` or `check_active_prs_status.sh`), avoiding vague or generic names?
- [ ] For repetitive polling over multi-target batches, are targets curried internally into a **zero-argument snippet** (Principle 7)?
- [ ] Is the invocation executed **directly as a standalone binary** (`./script.sh` or `/path/to/script.sh`) without generic interpreter prefixes (`bash`, `python3`) or compound `chmod` lines?
- [ ] Has any `chmod +x ... && exec ...` or interpreter prefix been eliminated from the command line?
- [ ] Is the scope granularity tier (Level 1–3) chosen appropriately to minimize review friction without causing approval fatigue?
- [ ] Are omnipotent tools (e.g., `gh`, `aws`, `kubectl`) constrained to explicit, minimal subcommands?
- [ ] For directory-dependent tools, is directory navigation safely encapsulated (`cd "${target_dir:?}"`) inside the script?
- [ ] For repetitive file inspections (e.g., `grep`), is the target file curried internally (Level 3) to enable persistent read approval without broad filesystem read permissions?
- [ ] Are dynamic variables exposed exclusively as trailing CLI arguments (`$1`, `$2`), or curried internally (Level 3) when targeting a fixed resource/batch?
- [ ] Are required parameters guarded with `${1:?message}` and format-validated (when parameters are used)?
- [ ] Is input validation in place for paths or required entities?
- [ ] Is the final command executed cleanly via `exec` (for single commands) or protected by `set -euo pipefail` (for multi-step pipelines and batch loops)?
- [ ] Has the snippet been audited via `snippet-review` and preceded by the mandatory Pre-Execution Notice?
