# Approvable Example: Git Worktree Setup (`git-worktree-setup`)

## 1. Overview & Use Cases

Fetches a remote branch from origin and adds an isolated git worktree under `.git-worktree/<branch-name>` relative to the target repository root.

### Triggers & Avoidance Thresholds
- **Multi-Command Chaining**: Combines `mkdir -p`, `git fetch`, and `git worktree add` into a single atomic stage, preventing multi-command line proliferation.
- **Convention Enforcement**: Enforces the project convention of placing worktrees strictly under `.git-worktree/<branch-name>`.
- **Precondition Check**: Prevents duplicate worktree creation errors by validating path existence upfront.

---

## 2. Approvable Script Implementation

Save this script as `setup_pr_worktree.sh`:

```bash
#!/usr/bin/env bash
#
# Description:
#   Fetches a remote PR branch and adds an isolated git worktree under
#   .git-worktree/<branch-name> relative to the target repository root.
#
# Arguments:
#   $1 - branch_name (remote branch name to checkout)
#   $2 - repo_dir    (optional, path to main repository root; defaults to current directory)
#
# Usage:
#   ./setup_pr_worktree.sh feature/sample-update
#   ./setup_pr_worktree.sh feature/sample-update /path/to/repo
#
# Prerequisites:
#   git, origin remote configured
#
set -euo pipefail

branch_name="${1:?Error: branch name is required}"
repo_dir="${2:-.}"

if [[ ! -d "${repo_dir}" ]]; then
  echo "Error: Repository directory not found: ${repo_dir}" >&2
  exit 1
fi

cd "${repo_dir}"

worktree_path=".git-worktree/${branch_name}"

if [[ -d "${worktree_path}" ]]; then
  echo "Worktree already exists at ${worktree_path}"
  exit 0
fi

if git show-ref --verify --quiet "refs/heads/${branch_name}"; then
  echo "Local branch '${branch_name}' already exists. Fetching remote updates..."
  git fetch origin "${branch_name}"
else
  echo "Fetching remote branch '${branch_name}' from origin..."
  git fetch origin "${branch_name}:${branch_name}"
fi

echo "Creating parent directory..."
mkdir -p "$(dirname "${worktree_path}")"

echo "Adding worktree at ${worktree_path}..."
git worktree add "${worktree_path}" "${branch_name}"

echo "Worktree successfully created at ${worktree_path}"
```

---

## 3. Parameters & Validation Specifications

| Parameter | Type | Required | Description & Validation |
|---|---|:---:|---|
| `$1` (`branch_name`) | String | Yes | Name of the remote branch. Guarded by `${1:?Error: ...}`. |
| `$2` (`repo_dir`) | Path | No | Path to repository root. Defaults to `.`. Verified with `[[ -d ... ]]`. |

- **Exit Codes**:
  - `0`: Success (worktree created or already exists).
  - `1`: Invalid arguments, missing repository, or git fetch/worktree failure.

---

## 4. Safety & Side-Effect Evaluation

- **Mutation Scope**: Creates local worktree directory under `.git-worktree/` and local git ref. Does not mutate `main` or active workspace files.
- **Idempotency**: Safely detects existing worktree paths and exits cleanly with message.
- **Fail-Fast**: `set -euo pipefail` halts execution immediately if `git fetch` fails (e.g. branch deleted remotely).

---

## 5. Pre-Execution Notice Template

Present this notice to the user before running the snippet:

```markdown
### 📋 Snippet Pre-Execution Notice

- **Script Path**: `[setup_pr_worktree.sh](path/to/scratch/setup_pr_worktree.sh) (path/to/scratch/setup_pr_worktree.sh)`
- **Invocation**: `./setup_pr_worktree.sh "feature/sample-update" "/path/to/repo"`
- **Purpose**: Fetch branch `feature/sample-update` and create isolated worktree under `.git-worktree/feature/sample-update`.
- **Parameters & Arguments**:
  - `$1` (`branch_name`): `feature/sample-update`
  - `$2` (`repo_dir`): `/path/to/repo`
- **Expected Side-Effects & Safety**: Fetches remote ref from origin and adds local worktree directory. No impact on main repository branch.
```

---

## 6. Variations & Adaptations

### Tracking Ref Setup (Alternative for Non-Local Tracking)
When creating a worktree directly tracking `origin/<branch>`:

```bash
git fetch origin "${branch_name}"
mkdir -p "$(dirname "${worktree_path}")"
git worktree add -b "${branch_name}" "${worktree_path}" "origin/${branch_name}"
```
