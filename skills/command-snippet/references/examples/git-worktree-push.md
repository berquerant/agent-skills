# Approvable Example: Git Worktree Force-Push with Lease (`git-worktree-push`)

## 1. Overview & Use Cases

Pushes a rebased branch from an isolated git worktree back to the remote repository using `--force-with-lease`.

### Triggers & Avoidance Thresholds
- **Critical Safety Guard**: Blindly running `git push -f` or `git push --force` destroys upstream work if remote changes occurred. This snippet mandates `--force-with-lease`.
- **Right-Aligned Parameterization**: Fixes `push --force-with-lease origin` and exposes only the worktree directory and branch name as trailing arguments.
- **Approval Whitelist Stability**: Provides an invariant, auditable script name (`push_worktree.sh`) for user approval review.

---

## 2. Approvable Script Implementation

Save this script as `push_worktree.sh`:

```bash
#!/usr/bin/env bash
#
# Description:
#   Pushes a rebased branch from an isolated git worktree to origin
#   using --force-with-lease to prevent accidental remote overwrite.
#
# Arguments:
#   $1 - worktree_dir (path to the target git worktree)
#   $2 - branch_name  (name of the remote branch to push)
#
# Usage:
#   ./push_worktree.sh /path/to/worktree branch-name
#
# Prerequisites:
#   git, origin remote configured
#
set -euo pipefail

worktree_dir="${1:?Error: worktree directory path is required}"
branch_name="${2:?Error: branch name is required}"

if [[ ! -d "${worktree_dir}" ]]; then
  echo "Error: Directory not found: ${worktree_dir}" >&2
  exit 1
fi

cd "${worktree_dir}"

echo "Pushing '${branch_name}' to origin with --force-with-lease..."
exec git push --force-with-lease origin "${branch_name}"
```

---

## 3. Parameters & Validation Specifications

| Parameter | Type | Required | Description & Validation |
|---|---|:---:|---|
| `$1` (`worktree_dir`) | Path | Yes | Local path to the git worktree. Verified with `[[ -d ... ]]`. |
| `$2` (`branch_name`) | String | Yes | Name of the remote branch. Verified non-empty via `${2:?...}`. |

- **Exit Codes**:
  - `0`: Success (branch pushed to origin).
  - `1`: Directory missing, parameter missing, or push rejected (e.g. lease expired due to upstream commit).

---

## 4. Safety & Side-Effect Evaluation

- **Mutation Scope**: Updates remote tracking branch on origin.
- **Safety Safeguard**: `--force-with-lease` guarantees that the push will fail if someone else (or an automated bot) has updated the remote branch since the last fetch.
- **Execution Model**: Uses `exec` to hand off execution directly to `git push`.

---

## 5. Pre-Execution Notice Template

Present this notice to the user before running the snippet:

```markdown
### 📋 Snippet Pre-Execution Notice

- **Script Path**: `[push_worktree.sh](path/to/scratch/push_worktree.sh) (path/to/scratch/push_worktree.sh)`
- **Invocation**: `./push_worktree.sh "/path/to/worktree" "feature/sample-update"`
- **Purpose**: Push rebased branch `feature/sample-update` to origin with `--force-with-lease`.
- **Parameters & Arguments**:
  - `$1` (`worktree_dir`): `/path/to/worktree`
  - `$2` (`branch_name`): `feature/sample-update`
- **Expected Side-Effects & Safety**: Updates remote branch on origin. Protected by `--force-with-lease` against overwriting unobserved upstream commits.
```

---

## 6. Variations & Adaptations

### Tracking Verification Prior to Push
If verifying that the local HEAD matches the expected branch before pushing:

```bash
cd "${worktree_dir}"
current_branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "${current_branch}" != "${branch_name}" ]]; then
  echo "Error: Expected branch '${branch_name}', but worktree is on '${current_branch}'" >&2
  exit 1
fi
exec git push --force-with-lease origin "${branch_name}"
```
