# Multi-Repository Scheduling & Low-Cost Pipeline

Guidelines for managing dependency PR resolution across single or multiple repositories without idle blocking, ensuring default branch stability, and tracking external API rate limits and costs.

---

## 1. Principles of Cost-Aware Scheduling

When processing dependency PRs across repositories:
1. **External API Rate Limiting & Cost Accounting**:
   - Platform API queries (`gh`, `glab`, etc.) must respect an agreed rate-limit interval (recommended default: **10 seconds per call**; confirm the appropriate interval with the user during planning).
   - Treat each API call as a billable resource/cost metric. For repeatable queries, reuse local caches following [`command-cache-protocol.md`](../../explore-plan-execute/references/command-cache-protocol.md) and dedicated caching skills/MCPs (`cache-memoizer` / `cache-mcp`) per `AGENTS.md` Rule 5. For non-cacheable or state-mutating calls, enforce the rate-limit interval.
   - Pacing cooldowns should be filled with local operations (worktree setups, builds, tests, linting, or log analysis) which incur zero API cost.
2. **Default Branch CI Stability Gate**:
   - Within any single repository, multiple PRs must be processed strictly sequentially.
   - Merging a PR is **not complete** until the default branch CI triggered by the merge commit passes. Do not begin rebasing or verifying the next candidate PR in the same repository until the default branch CI is green.
3. **Utilize Idle Wait Times Across Repositories**:
   - When a repository enters an idle wait state (waiting for PR branch CI or post-merge default branch CI), immediately switch focus to another pending repository to perform triage, checkout, or local verification.
4. **Local-First Verification**:
   - Run local test/lint checks inside the worktree first. Catch broken updates before pushing branches or consuming remote CI / API quota.

---

## 2. Repository Lifecycle State Machine

Track each repository's progress in a state table:

| State | Description | Next Action |
|---|---|---|
| `QUEUED` | Candidate repository identified; PRs not yet triaged. | Fetch PR list in batch and triage candidates locally. |
| `TRIAGING` | Inspecting PR metadata, security, and context. | Filter, audit security, and formulate plan. |
| `PLANNING` | Presenting resolution plan to user (EPE Gate). | Await user approval on execution plan, order, and API rate limit. |
| `CHECKOUT` | Creating isolated worktree and syncing branch. | Rebase and run local checks. |
| `LOCAL_VERIFY` | Running local build/test checks inside worktree. | If pass -> push and transition to `WAITING_PR_CI`. |
| `WAITING_PR_CI` | Remote PR branch CI checks in-flight. **(IDLE WAIT)** | Switch focus to next actionable repository! |
| `REMEDIATING` | Failure occurred; remediation plan awaiting approval. | Await user approval; execute minimal fix. |
| `READY_MERGE` | Remote PR CI passed; ready to merge. | Merge PR and transition to `WAITING_DEFAULT_CI`. |
| `WAITING_DEFAULT_CI` | Default branch CI triggered by merge commit is running. **(IDLE WAIT)** | Switch focus to next repository! (Block next PR in same repo). |
| `HALTED` | Safety halt triggered due to doubts, security, or CI break. | Blocked pending explicit user clearance. |
| `COMPLETED` | All candidate PRs merged and verified on default branch. | Clean up worktrees and report final status. |

---

## 3. Context-Switching Protocol

### 3.1 Detecting Idle Wait
An idle wait state occurs when:
- Remote PR branch CI checks are running (`WAITING_PR_CI`).
- Default branch CI checks are running following a merge (`WAITING_DEFAULT_CI`).
- User input is requested for a remediation plan (`REMEDIATING`) or safety halt (`HALTED`).
- Network-bound long-running operations cannot proceed locally.

### 3.2 Switching Focus
1. **Record State**: Note the current state of repository $R_i$ in the tracking table (including PR number, branch, worktree path, last action, and timestamp).
2. **Select Next Actionable Repository**:
   - Prioritize repositories in `QUEUED`, `TRIAGING`, `CHECKOUT`, or `LOCAL_VERIFY` states (these involve local CPU/disk actions and zero API cost).
   - If no repository has local work ready, check if any repository in `WAITING_DEFAULT_CI` or `WAITING_PR_CI` has elapsed sufficient time and rate limit cooldown to check status.
3. **Resume or Start**:
   - Switch active context (directory and worktree) to the selected repository $R_j$.
   - Execute the next logical step.

### 3.3 Rate Limiter & Anti-Polling Rules
- **Rate Limit Enforcement**:
  - Keep track of `LAST_API_CALL_TS`.
  - Ensure at least the agreed interval (e.g., 10 seconds) has elapsed since the last external CLI/API execution before issuing another remote query.
  - If the interval has not elapsed, wait or continue local processing.
- **Anti-Polling Rules**:
  - Do NOT poll remote CI in a short loop or streaming monitor:
    - Anti-pattern: `while true; do gh pr checks; sleep 5; done`
    - Anti-pattern: `gh pr checks <id> --watch`
    - Anti-pattern: `gh run watch`
    - Anti-pattern: `glab ci trace`
  - Check remote CI status only when:
    - Transitioning between repository tasks.
    - Or after a meaningful pacing interval (e.g., at least 2–3 minutes between remote status queries, respecting the 10s API rate limit).
- **Single-Shot Check Commands**:
  - **GitHub PR CI**: `gh pr checks <id>`
  - **GitHub Default Branch CI**: `gh run list --branch <default-branch> --commit <merge-sha> --limit 1`
  - **GitLab PR/MR CI**: `glab mr view <id>`
  - **GitLab Default Branch CI**: `glab ci status --branch <default-branch>`

---

## 4. Single-Repository Sequential Pipeline with Default Branch CI Gate

When multiple candidate PRs exist for a single repository:

1. **Strictly Sequential Execution**:
   ```mermaid
   flowchart TD
       P1_LOCAL["PR #1: Rebase & Local Verify"] --> P1_PR_CI["PR #1: Remote PR CI"]
       P1_PR_CI --> P1_MERGE["PR #1: Merge to Default Branch"]
       P1_MERGE --> P1_DEF_CI["PR #1: Wait Default Branch CI Pass"]
       P1_DEF_CI -->|Pass| P2_REBASE["PR #2: Rebase on Updated Default Branch"]
       P1_DEF_CI -->|Fail| HALT["Safety Halt: Default Branch Broken!"]
       P2_REBASE --> P2_LOCAL["PR #2: Local Verify"]
       P2_LOCAL --> P2_PR_CI["PR #2: Remote PR CI"]
       P2_PR_CI --> P2_MERGE["PR #2: Merge to Default Branch"]
       P2_MERGE --> P2_DEF_CI["PR #2: Wait Default Branch CI Pass"]
       P2_DEF_CI -->|Pass| DONE["Repository Completed"]
       P2_DEF_CI -->|Fail| HALT
   ```
2. **Idle Wait Utilization**:
   - During `P1_DEF_CI` (which may take 3–10 minutes), the agent MUST switch to another repository.
   - Never start `P2_REBASE` until `P1_DEF_CI` succeeds.
3. **Incident Response on Default Branch Failure**:
   - If default branch CI fails after merge, immediately freeze processing for this repository.
   - Trigger a Safety Halt report with failure details.
   - Offer the user immediate remediation options (e.g., revert merge commit via `git revert`, or create a fix branch).

---

## 5. Multi-Repo Progress Summary Template

Keep a lightweight summary table during execution:

```markdown
| Repository | Target PR | Current State | API Calls | Notes / Next Action |
|---|---|---|---|---|
| `repo-alpha` | #42 (dep A) | `WAITING_DEFAULT_CI` | 3 | PR merged; waiting on default branch CI (run #102) |
| `repo-beta` | #15 (dep B) | `LOCAL_VERIFY` | 1 | Running tests locally in worktree (0 API calls) |
| `repo-gamma` | #88 (dep C) | `HALTED` | 2 | Doubts raised on registry URL; awaiting user |
```
