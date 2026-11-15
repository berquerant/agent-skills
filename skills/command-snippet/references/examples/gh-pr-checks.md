# Approvable Example: Batch PR Status Monitor (`gh-pr-checks`)

## 1. Overview & Use Cases

Monitors and displays CI checks and status check rollups across a batch of active pull requests.

### Triggers & Avoidance Thresholds
- **Repetitive Polling & Parameter Drift**: Repeatedly running status checks with varying arguments (`gh pr checks 10`, `gh pr checks 11`) triggers new approval prompts.
- **Zero-Argument Invariant Pattern**: Currying a fixed batch of target PRs into a zero-argument snippet (or right-aligned positional arguments) guarantees invariant command-line invocations for frictionless repetitive polling.
- **Multi-Line Command Threshold**: Replaces multi-line shell loops (`for pr in ...; do gh pr checks ...; done`) with an audited, standalone snippet.

---

## 2. Approvable Script Implementation

Save this script as `check_active_prs_status.sh`:

```bash
#!/usr/bin/env bash
#
# Description:
#   Checks and reports CI checks status for target pull requests.
#   Currying target PRs into an invariant invocation enables repetitive
#   polling without approval friction.
#
# Arguments:
#   $@ - pr_numbers (numeric pull request IDs, space-separated)
#
# Usage:
#   ./check_active_prs_status.sh 42 43 44
#
# Prerequisites:
#   gh (GitHub CLI, authenticated)
#
set -euo pipefail

if [[ "$#" -eq 0 ]]; then
  echo "Error: At least one PR number is required" >&2
  exit 1
fi

echo "=== Active Batch CI Status ($(date '+%Y-%m-%d %H:%M:%S')) ==="
for pr in "$@"; do
  if [[ ! "${pr}" =~ ^[0-9]+$ ]]; then
    echo "Warning: Skipping invalid PR number: ${pr}" >&2
    continue
  fi
  echo "--- PR #${pr} ---"
  gh pr checks "${pr}" || true
done
echo "=== Check complete ==="
```

---

## 3. Parameters & Validation Specifications

| Parameter | Type | Required | Description & Validation |
|---|---|:---:|---|
| `$@` (`pr_numbers`) | Integers | Yes | One or more space-separated PR numbers. Each argument is verified with `[[ "${pr}" =~ ^[0-9]+$ ]]`. |

- **Exit Codes**:
  - `0`: Always exits 0 when valid PR arguments are provided (individual check failures do not abort loop thanks to `|| true`).
  - `1`: Invocation error (no PR numbers provided).

---

## 4. Safety & Side-Effect Evaluation

- **Mutation Risk**: Strictly read-only query operation.
- **External Network Access**: Calls GitHub Status Check API via `gh`.
- **Fault Tolerance**: Contains `|| true` on `gh pr checks` so that pending or failing checks on one PR do not prevent inspecting the remaining candidates.

---

## 5. Pre-Execution Notice Template

Present this notice to the user before running the snippet:

```markdown
### 📋 Snippet Pre-Execution Notice

- **Script Path**: `[check_active_prs_status.sh](path/to/scratch/check_active_prs_status.sh) (path/to/scratch/check_active_prs_status.sh)`
- **Invocation**: `./check_active_prs_status.sh 42 43 44`
- **Purpose**: Poll and report CI checks status for PRs #42, #43, and #44.
- **Parameters & Arguments**:
  - `$@` (`pr_numbers`): List of target pull request numbers (`42 43 44`).
- **Expected Side-Effects & Safety**: Read-only GitHub CLI queries. Safe for repetitive polling.
```

---

## 6. Variations & Adaptations

### Fixed-Batch Zero-Argument Variant (Level 3 Curry)
When polling a known active batch repeatedly over an extended wait loop:

```bash
#!/usr/bin/env bash
set -euo pipefail

readonly TARGET_PRS=(42 43 44)

echo "=== Active Batch CI Status ($(date '+%Y-%m-%d %H:%M:%S')) ==="
for pr in "${TARGET_PRS[@]}"; do
  echo "--- PR #${pr} ---"
  gh pr checks "${pr}" || true
done
echo "=== Check complete ==="
```
Invocation: `./check_active_prs_status.sh` (100% invariant, zero arguments).
