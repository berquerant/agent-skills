# Approvable Example: GitHub PR Merge (`gh-pr-merge`)

## 1. Overview & Use Cases

Merges an approved pull request via GitHub CLI using squash-merge (or standard merge) and deletes the remote head branch upon completion.

### Triggers & Avoidance Thresholds
- **Omnipotent Tool Constraint**: `gh` possesses extensive administrative capabilities. Currying the subcommand to `pr merge --squash --delete-branch` restricts execution to this exact operation.
- **Dynamic Argument Risk**: Passing arbitrary flags or subcommands through wrappers creates security hazards. This snippet fixes the merge verb and validates the PR number strictly as an integer.
- **Approval Friction Reduction**: Establishing a stable binary prefix (`./merge_pr.sh <pr-number>`) enables persistent human review whitelisting without blanket approval for generic `gh` commands.

---

## 2. Approvable Script Implementation

Save this script as `merge_pr.sh`:

```bash
#!/usr/bin/env bash
#
# Description:
#   Squash-merges a specified GitHub pull request and deletes the remote head branch.
#   Restricts gh CLI execution strictly to pr merge --squash --delete-branch.
#
# Arguments:
#   $1 - pr_number (numeric pull request ID)
#   $2 - repo_dir   (optional, path to repository root; defaults to current directory)
#
# Usage:
#   ./merge_pr.sh 123
#   ./merge_pr.sh 123 /path/to/repo
#
# Prerequisites:
#   gh (GitHub CLI, authenticated)
#
set -euo pipefail

pr_number="${1:?Error: PR number is required}"
repo_dir="${2:-.}"

if [[ ! "${pr_number}" =~ ^[0-9]+$ ]]; then
  echo "Error: PR number must be a positive integer: ${pr_number}" >&2
  exit 1
fi

if [[ ! -d "${repo_dir}" ]]; then
  echo "Error: Repository directory not found: ${repo_dir}" >&2
  exit 1
fi

cd "${repo_dir}"
exec gh pr merge "${pr_number}" --squash --delete-branch
```

---

## 3. Parameters & Validation Specifications

| Parameter | Type | Required | Description & Validation |
|---|---|:---:|---|
| `$1` (`pr_number`) | Integer | Yes | Target pull request ID. Validated via regex `^[0-9]+$`. |
| `$2` (`repo_dir`) | Path | No | Optional local repository directory. Defaults to `.`. Verified with `[[ -d ... ]]`. |

- **Exit Codes**:
  - `0`: Success (PR squash-merged and remote branch deleted).
  - `1`: Non-numeric PR number, directory not found, or merge blocked by CI / branch protection.

---

## 4. Safety & Side-Effect Evaluation

- **Mutation Scope**: Mutates remote GitHub repository state (merges pull request into default branch and deletes remote branch).
- **Execution Model**: Process replacement via `exec` ensures that signals (e.g. `SIGINT`) terminate the CLI call immediately.
- **Safety Safeguard**: Regex check `^[0-9]+$` guarantees no malicious command injection can occur through the PR number argument.

---

## 5. Pre-Execution Notice Template

Present this notice to the user before running the snippet:

```markdown
### 📋 Snippet Pre-Execution Notice

- **Script Path**: `[merge_pr.sh](path/to/scratch/merge_pr.sh) (path/to/scratch/merge_pr.sh)`
- **Invocation**: `./merge_pr.sh "123" "/path/to/repo"`
- **Purpose**: Squash-merge pull request #123 and delete the remote branch.
- **Parameters & Arguments**:
  - `$1` (`pr_number`): `123` (target pull request ID)
  - `$2` (`repo_dir`): `/path/to/repo` (repository root directory)
- **Expected Side-Effects & Safety**: Mutates remote GitHub repository. Merges PR #123 into default branch and deletes remote head branch.
```

---

## 6. Variations & Adaptations

### Merge Commit Variant
When the repository convention requires standard merge commits instead of squash merging:

```bash
cd "${repo_dir}"
exec gh pr merge "${pr_number}" --merge --delete-branch
```
