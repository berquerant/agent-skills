# Approvable Example: Git Worktree Cleanup (`git-worktree-cleanup`)

## 1. Overview & Use Cases

Safely removes a completed git worktree directory and optionally deletes its corresponding local tracking branch.

### Triggers & Avoidance Thresholds
- **Multi-Command Chaining**: Combines `git worktree remove` and `git branch -D` into a single atomic teardown stage.
- **Resource Leak Prevention**: Leaving orphan worktree registrations or stale branches bloats the repository state and git metadata.
- **Precondition & Existence Checks**: Validates that the worktree and branch exist before attempting deletion, ensuring idempotent cleanup.

---

## 2. Approvable Script Implementation

Save this script as `cleanup_pr_worktree.sh`:

```bash
#!/usr/bin/env bash
#
# Description:
#   Cleans up a git worktree directory and deletes the corresponding local branch.
#
# Arguments:
#   $1 - worktree_dir (path to the worktree to remove)
#   $2 - branch_name  (optional, local branch name to delete)
#   $3 - repo_dir     (optional, path to main repository root; defaults to current directory)
#
# Usage:
#   ./cleanup_pr_worktree.sh .git-worktree/feature/sample-update
#   ./cleanup_pr_worktree.sh .git-worktree/feature/sample-update feature/sample-update
#   ./cleanup_pr_worktree.sh .git-worktree/feature/sample-update feature/sample-update /path/to/repo
#
# Prerequisites:
#   git
#
set -euo pipefail

worktree_dir="${1:?Error: worktree directory path is required}"
branch_name="${2:-}"
repo_dir="${3:-.}"

if [[ ! -d "${repo_dir}" ]]; then
  echo "Error: Main repository directory not found: ${repo_dir}" >&2
  exit 1
fi

cd "${repo_dir}"

if [[ -d "${worktree_dir}" ]]; then
  echo "Removing worktree at '${worktree_dir}'..."
  git worktree remove --force "${worktree_dir}"
else
  echo "Worktree directory '${worktree_dir}' does not exist. Pruning registrations..."
  git worktree prune
fi

if [[ -n "${branch_name}" ]]; then
  if git show-ref --verify --quiet "refs/heads/${branch_name}"; then
    echo "Deleting local branch '${branch_name}'..."
    git branch -D "${branch_name}"
  else
    echo "Local branch '${branch_name}' does not exist. Skipping."
  fi
fi

echo "Worktree cleanup successfully completed."
```

---

## 3. Parameters & Validation Specifications

| Parameter | Type | Required | Description & Validation |
|---|---|:---:|---|
| `$1` (`worktree_dir`) | Path | Yes | Path to the worktree to remove. Handled gracefully if already deleted. |
| `$2` (`branch_name`) | String | No | Optional local branch name to delete via `git branch -D`. |
| `$3` (`repo_dir`) | Path | No | Optional main repository root. Defaults to `.`. |

- **Exit Codes**:
  - `0`: Success (worktree removed and branch deleted if present).
  - `1`: Invalid main repo directory or git command error.

---

## 4. Safety & Side-Effect Evaluation

- **Mutation Scope**: Deletes the specified worktree directory on the local filesystem and the local branch ref.
- **Isolation Guarantee**: Does not delete remote branches or affect other worktrees or the main checkout.
- **Idempotency**: Safely handles cases where the worktree or branch was already removed, avoiding script crashes.

---

## 5. Pre-Execution Notice Template

Present this notice to the user before running the snippet:

```markdown
### 📋 Snippet Pre-Execution Notice

- **Script Path**: [cleanup_pr_worktree.sh](path/to/scratch/cleanup_pr_worktree.sh) (path/to/scratch/cleanup_pr_worktree.sh)
- **Invocation**: `./cleanup_pr_worktree.sh ".git-worktree/feature/sample-update" "feature/sample-update" "/path/to/repo"`
- **Purpose**: Remove isolated worktree and delete local branch `feature/sample-update`.
- **Parameters & Arguments**:
  - `$1` (`worktree_dir`): `.git-worktree/feature/sample-update`
  - `$2` (`branch_name`): `feature/sample-update`
  - `$3` (`repo_dir`): `/path/to/repo`
- **Expected Side-Effects & Safety**: Removes worktree directory from filesystem and deletes local tracking branch.
```

---

## 6. Variations & Adaptations

### Worktree Prune Only
When cleaning up stale worktree metadata without deleting branches:

```bash
cd "${repo_dir}"
exec git worktree prune
```
