# Approvable Example: Batch GitHub PR List (`gh-pr-list`)

## 1. Overview & Use Cases

Retrieves open pull requests along with metadata (PR number, title, head branch, author, labels, modified files, body text, and URL) from GitHub CLI in structured JSON format within a single query.

### Triggers & Avoidance Thresholds
- **Character Length**: An inline `gh pr list --state open --json number,title,headRefName,author,labels,files,body,url` exceeds the 120-character threshold.
- **Infix Argument & Directory Portability**: Avoids `gh -C <dir>` inconsistencies across different platforms and CLI versions by encapsulating directory navigation inside the script.
- **Cost-Aware Batch Retrieval**: Replaces multiple fine-grained PR inspection calls with a single batch query, saving API rate limits and execution turns.

---

## 2. Approvable Script Implementation

Save this script as `fetch_repo_prs.sh` (or `fetch_candidate_prs.sh`):

```bash
#!/usr/bin/env bash
#
# Description:
#   Fetches open pull requests with detailed metadata from GitHub CLI
#   within a target repository directory and outputs structured JSON to stdout.
#
# Arguments:
#   $1 - repo_dir (path to local repository root)
#
# Usage:
#   ./fetch_repo_prs.sh path/to/repo
#   ./fetch_repo_prs.sh path/to/repo > candidate_prs.json
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
exec gh pr list --state open --json number,title,headRefName,author,labels,files,body,url
```

---

## 3. Parameters & Validation Specifications

| Parameter | Type | Required | Description & Validation |
|---|---|:---:|---|
| `$1` (`repo_dir`) | Path | Yes | Local filesystem path to repository root. Verified with `[[ -d "${repo_dir}" ]]`. |

- **Exit Codes**:
  - `0`: Success (valid JSON array emitted to stdout).
  - `1`: Missing argument, invalid directory path, or CLI execution failure.

---

## 4. Safety & Side-Effect Evaluation

- **Mutation Risk**: None. Strictly read-only query operation.
- **External Network Access**: Calls GitHub REST/GraphQL API via `gh`. Adheres to API rate limit budgets.
- **Execution Model**: Uses `exec` to replace the subshell with `gh`, ensuring clean signal propagation and zero orphan processes.

---

## 5. Pre-Execution Notice Template

Present this notice to the user before running the snippet:

```markdown
### 📋 Snippet Pre-Execution Notice

- **Script Path**: [fetch_repo_prs.sh](path/to/scratch/fetch_repo_prs.sh) (path/to/scratch/fetch_repo_prs.sh)
- **Invocation**: `./fetch_repo_prs.sh "/path/to/target-repo"`
- **Purpose**: Fetch open pull requests with detailed metadata in structured JSON format.
- **Parameters & Arguments**:
  - `$1` (`repo_dir`): Target local repository directory path.
- **Expected Side-Effects & Safety**: Read-only GitHub API query via `gh`. No local or remote modifications.
```

---

## 6. Variations & Adaptations

### Direct Output to Cache File
If storing results directly to a local JSON cache:

```bash
#!/usr/bin/env bash
set -euo pipefail

repo_dir="${1:?Error: repository directory required}"
output_file="${2:?Error: output file path required}"

if [[ ! -d "${repo_dir}" ]]; then
  echo "Error: Directory not found: ${repo_dir}" >&2
  exit 1
fi

cd "${repo_dir}"
gh pr list --state open --json number,title,headRefName,author,labels,files,body,url > "${output_file}"
```
