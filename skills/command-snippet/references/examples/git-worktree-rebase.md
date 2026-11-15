# Approvable Example: Git Worktree Rebase (`git-worktree-rebase`)

## 1. Overview & Use Cases

Inspects an isolated git worktree branch and safely rebases it onto `origin/main` atomically.

### Triggers & Avoidance Thresholds
- **Unnecessary Rebase Prevention**: Checks `git merge-base --is-ancestor origin/main HEAD` first; if the branch is already up to date, skips rebasing without mutating ref logs.
- **Directory Encapsulation**: Navigates inside the worktree directory safely, right-aligning the worktree path as `$1`.
- **Atomic Single Stage**: Encapsulates ancestor verification and rebase execution into one reviewable transaction.

---

## 2. Approvable Script Implementation

Save this script as `rebase_worktree.sh`:

```bash
#!/usr/bin/env bash
#
# Description:
#   Rebases a target git worktree branch onto origin/main.
#   Skips cleanly if the branch is already an ancestor of origin/main.
#
# Arguments:
#   $1 - worktree_dir (path to the target git worktree)
#
# Usage:
#   ./rebase_worktree.sh /path/to/worktree
#
# Prerequisites:
#   git, origin/main fetched
#
set -euo pipefail

worktree_dir="${1:?Error: worktree directory path is required}"

if [[ ! -d "${worktree_dir}" ]]; then
  echo "Error: Directory not found: ${worktree_dir}" >&2
  exit 1
fi

cd "${worktree_dir}"

echo "Fetching latest origin/main..."
git fetch origin main

echo "Checking if branch is already up to date with origin/main..."
if git merge-base --is-ancestor origin/main HEAD; then
  echo "Branch is already up to date with origin/main. No rebase needed."
  exit 0
fi

echo "Rebasing branch onto origin/main..."
git rebase origin/main
echo "Rebase successfully completed."
```

---

## 3. Parameters & Validation Specifications

| Parameter | Type | Required | Description & Validation |
|---|---|:---:|---|
| `$1` (`worktree_dir`) | Path | Yes | Filesystem path to the git worktree. Verified with `[[ -d ... ]]`. |

- **Exit Codes**:
  - `0`: Success (rebase succeeded or already up to date).
  - `1`: Directory not found, git error, or rebase conflict.

---

## 4. Safety & Side-Effect Evaluation

- **Mutation Scope**: Rewrites git commit history of the local worktree branch. Does not push to remote.
- **Conflict Handling**: If a rebase conflict occurs, `git rebase` halts and reports conflict files. `set -e` ensures no follow-up steps run in a dirty state.
- **Safety Safeguard**: Does not run `--autostash` or dangerous reset flags unless explicitly configured.

---

## 5. Pre-Execution Notice Template

Present this notice to the user before running the snippet:

```markdown
### 📋 Snippet Pre-Execution Notice

- **Script Path**: `[rebase_worktree.sh](path/to/scratch/rebase_worktree.sh) (path/to/scratch/rebase_worktree.sh)`
- **Invocation**: `./rebase_worktree.sh "/path/to/worktree"`
- **Purpose**: Check and rebase local worktree branch onto `origin/main`.
- **Parameters & Arguments**:
  - `$1` (`worktree_dir`): `/path/to/worktree` (target git worktree path)
- **Expected Side-Effects & Safety**: Rewrites local commit history on the worktree branch. Skips safely if already up to date. Halts cleanly on conflicts.
```

---

## 6. Variations & Adaptations

### Explicit Remote Base Branch Parameter
When the default branch is named `master` or a custom branch (e.g. `develop`):

```bash
worktree_dir="${1:?Error: worktree directory required}"
base_branch="${2:-origin/main}"

cd "${worktree_dir}"
if git merge-base --is-ancestor "${base_branch}" HEAD; then
  echo "Already up to date with ${base_branch}."
  exit 0
fi
git rebase "${base_branch}"
```
