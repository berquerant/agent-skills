# Multi-Repository Scheduling & Low-Cost Pipeline

Guidelines for managing dependency PR resolution across multiple repositories without idle blocking or excessive API resource consumption.

---

## 1. Principles of Cost-Aware Scheduling

When processing dependency PRs across multiple repositories:
1. **Minimize Remote API & CI Queries**: Batch candidate queries up front and store status locally. Do not tight-loop or rapidly poll `gh pr checks` or platform APIs.
2. **Utilize Idle Wait Times**: When a repository enters an idle wait state (e.g., waiting for remote CI workflow completion, artifact building, or user confirmation), immediately switch focus to another pending repository rather than blocking.
3. **Local-First Verification**: Run local test/lint checks inside the worktree first. Catch broken updates before triggering remote CI or spending CI minutes.

---

## 2. Repository Lifecycle State Machine

Track each repository's progress in a state table:

| State | Description | Next Action |
|---|---|---|
| `QUEUED` | Candidate repository identified; PRs not yet triaged. | Fetch PR list in batch and triage candidates. |
| `TRIAGING` | Inspecting PR metadata, security, and context. | Filter and run security audit. |
| `CHECKOUT` | Creating isolated worktree and syncing branch. | Rebase and run local checks. |
| `LOCAL_VERIFY` | Running local build/test checks inside worktree. | If pass -> push and transition to `WAITING_CI`. |
| `WAITING_CI` | Remote CI checks are in-flight. **(IDLE WAIT)** | Switch focus to next runnable repository! |
| `REMEDIATING` | Failure occurred; remediation plan awaiting approval. | Await user approval; execute minimal fix. |
| `READY_MERGE` | Remote CI passed; ready to merge. | Merge PR and clean up worktree. |
| `HALTED` | Safety halt triggered due to doubts or security concern. | Blocked pending explicit user clearance. |
| `COMPLETED` | PR merged and worktree cleaned up. | Done. |

---

## 3. Context-Switching Protocol

### 3.1 Detecting Idle Wait
An idle wait state occurs when:
- Remote CI checks are running (`WAITING_CI`).
- User input is requested for a remediation plan (`REMEDIATING`) or safety halt (`HALTED`).
- Network-bound long-running test execution cannot proceed locally.

### 3.2 Switching Focus
1. **Record State**: Note the current state of repository $R_i$ in the tracking table (including PR number, branch, worktree path, and CI run timestamp).
2. **Select Next Actionable Repository**:
   - Look for any repository in `QUEUED`, `TRIAGING`, `CHECKOUT`, or `LOCAL_VERIFY` states.
   - If none are in active states, inspect repositories in `WAITING_CI` to check if their remote checks have finished since the last switch.
3. **Resume or Start**:
   - Switch active context (directory and worktree) to the selected repository $R_j$.
   - Execute the next logical step.

### 3.3 Batch Status Check (Anti-Polling Rule)
- Do NOT poll remote CI in a short loop or streaming monitor:
  - Anti-pattern: `while true; do gh pr checks; sleep 5; done`
  - Anti-pattern: `gh pr checks <id> --watch`
  - Anti-pattern: `glab ci trace` (streaming continuous output)
- Check remote CI status only when:
  - Transitioning between repository tasks.
  - Or after a pacing interval has elapsed (e.g., at least 2–3 minutes between remote checks).
- Use single-shot check commands:
  - **GitHub**: `gh pr checks <id>` (runs once and exits immediately)
  - **GitLab**: `glab ci status` or `glab mr view <id>` (runs once and exits immediately)

---

## 4. Multi-Repo Progress Summary Template

Keep a lightweight summary table during execution:

```markdown
| Repository | Target PR | Current State | Notes / Blocker |
|---|---|---|---|
| `repo-alpha` | #42 (dep A) | `WAITING_CI` | Rebased & pushed; CI run started 2m ago |
| `repo-beta` | #15 (dep B) | `LOCAL_VERIFY` | Running test suite locally in worktree |
| `repo-gamma` | #88 (dep C) | `HALTED` | Doubts raised on registry URL; awaiting user |
```
