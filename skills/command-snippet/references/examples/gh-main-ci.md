# Approvable Example: Default Branch CI Monitor (`gh-main-ci`)

## 1. Overview & Use Cases

Monitors the latest GitHub Actions workflow runs on the default branch (`main` or `master`) or for a specific commit SHA following a merge.

### Triggers & Avoidance Thresholds
- **Character Length Threshold**: Running `gh run list --branch main --limit 3 --json status,conclusion,name,databaseId,headSha,event,url` inline exceeds 120 characters.
- **Sequential Stability Gate**: In automated PR triage, confirming that the default branch CI passes before rebasing the next candidate PR is mandatory. This snippet provides a stable read-only probe.
- **Directory Portability**: Encapsulates target directory navigation so external repositories can be inspected without global path flags.

---

## 2. Approvable Script Implementation

Save this script as `check_main_ci.sh`:

```bash
#!/usr/bin/env bash
#
# Description:
#   Checks the latest GitHub Actions workflow runs on the default branch.
#   Optionally filters by commit SHA.
#
# Arguments:
#   $1 - repo_dir   (path to local repository root)
#   $2 - commit_sha (optional, git commit SHA to filter runs)
#
# Usage:
#   ./check_main_ci.sh path/to/repo
#   ./check_main_ci.sh path/to/repo a1b2c3d
#
# Prerequisites:
#   gh (GitHub CLI, authenticated)
#
set -euo pipefail

repo_dir="${1:?Error: repository directory is required}"
commit_sha="${2:-}"

if [[ ! -d "${repo_dir}" ]]; then
  echo "Error: Directory not found: ${repo_dir}" >&2
  exit 1
fi

cd "${repo_dir}"

if [[ -n "${commit_sha}" ]]; then
  exec gh run list --branch main --commit "${commit_sha}" --json databaseId,name,status,conclusion,url
else
  exec gh run list --branch main --limit 3 --json databaseId,name,status,conclusion,event,url
fi
```

---

## 3. Parameters & Validation Specifications

| Parameter | Type | Required | Description & Validation |
|---|---|:---:|---|
| `$1` (`repo_dir`) | Path | Yes | Local repository directory path. Verified with `[[ -d "${repo_dir}" ]]`. |
| `$2` (`commit_sha`) | String | No | Optional commit hash. When present, filters runs strictly to that commit. |

- **Exit Codes**:
  - `0`: Success (workflow runs listed in JSON format).
  - `1`: Invalid directory or GitHub API failure.

---

## 4. Safety & Side-Effect Evaluation

- **Mutation Risk**: None. Strictly read-only query operation.
- **External Network Access**: Calls GitHub Actions API via `gh`.
- **Clean Execution**: Invokes `exec` to hand off subshell execution directly to `gh`.

---

## 5. Pre-Execution Notice Template

Present this notice to the user before running the snippet:

```markdown
### 📋 Snippet Pre-Execution Notice

- **Script Path**: `[check_main_ci.sh](path/to/scratch/check_main_ci.sh) (path/to/scratch/check_main_ci.sh)`
- **Invocation**: `./check_main_ci.sh "/path/to/repo" "a1b2c3d"`
- **Purpose**: Verify GitHub Actions workflow status on `main` branch for commit `a1b2c3d`.
- **Parameters & Arguments**:
  - `$1` (`repo_dir`): `/path/to/repo` (repository root directory)
  - `$2` (`commit_sha`): `a1b2c3d` (target commit SHA)
- **Expected Side-Effects & Safety**: Read-only GitHub Actions API query. No repository changes.
```

---

## 6. Variations & Adaptations

### Plain Text Summary for Quick Terminal Audit
When structured JSON output is not required and human-readable terminal table output is preferred:

```bash
cd "${repo_dir}"
exec gh run list --branch main --limit 5
```
