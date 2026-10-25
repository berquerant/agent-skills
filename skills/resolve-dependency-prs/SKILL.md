---
name: resolve-dependency-prs
description: >-
  Use this skill to triage, audit, verify, and merge dependency update or
  security advisory Pull Requests / Merge Requests on GitHub or GitLab. Enforces
  cost-aware local batching, legitimacy & security audits, safety halt on doubts,
  idle-wait switching across repositories, worktree isolation, and CI verification.
---

# Resolve Dependency PRs

Triage open dependency and security PRs/MRs, audit legitimacy and safety, verify in isolated worktrees, validate through CI, remediate failures under user approval, manage multi-repo pipelines without idle waiting, and merge safely.

Coordinates [`explore-plan-execute`](../explore-plan-execute/SKILL.md) for upfront planning, [`git-worktree`](../git-worktree/SKILL.md) for branch isolation, [`step-gate`](../step-gate/SKILL.md) for remediation gates, [`task-dependency-pipeline`](../task-dependency-pipeline/SKILL.md) for cooperative multi-repo scheduling, and [`code-review`](../code-review/SKILL.md) for thorough mode audits.

---

## Execution Modes

This skill supports two execution modes:
- **`standard` (Default)**: Standard automated triage, manifest-level security check, rebase, local test/lint, CI verification, and safe merge.
- **`thorough`**: Deep micro reviews per candidate via [`code-review`](../code-review/SKILL.md), holistic macro conflict detection, co-upgrade bundling, and exhaustive multi-faceted verification prior to merge per [references/thorough-audit-and-verification.md](references/thorough-audit-and-verification.md).

---

## Prerequisites

- `git` is installed and the current directory is inside a git repository (or target repositories are available locally).
- A supported platform tool is authenticated (`gh` CLI for GitHub or `glab` CLI for GitLab; refer to [`multi-repo-change/references/platform-tools.md`](../multi-repo-change/references/platform-tools.md)).
- **Cost & Rate Limits**: Follow cost-aware principles (`AGENTS.md` Rule 5 and [`command-cache-protocol.md`](../explore-plan-execute/references/command-cache-protocol.md)). Confirm API rate limit (default: 10s/call) and execution mode (`standard` or `thorough`).
- **Explore & Plan First**: Follow [`explore-plan-execute`](../explore-plan-execute/SKILL.md) to triage candidates and present a concrete resolution plan before modifying branches.

---

## Step 1: Collect, Filter, and Audit Candidate PRs/MRs

### 1.1 Cost-Aware Batch Collection
Retrieve open PRs/MRs along with files and body in a single query to avoid repetitive platform API calls. Save candidate metadata into a local file (`candidate_prs.json` / `candidate_mrs.json`) first, and inspect candidates locally per [references/filter-criteria.md](references/filter-criteria.md):
- **GitHub**: `gh pr list --state open --json number,title,headRefName,author,labels,files,body > candidate_prs.json`
- **GitLab**: `glab mr list --state opened > candidate_mrs.txt` or `glab api "projects/:id/merge_requests?state=opened" > candidate_mrs.json`

### 1.2 File-Level Exclusivity Check
Filter PRs according to [references/filter-criteria.md](references/filter-criteria.md):
- Ensure changes are strictly confined to package manager manifests and lockfiles.
- **Immediately exclude** any PR modifying application code, test logic, CI configurations, or general project files.

### 1.3 Context, Legitimacy & Security Audit
Audit each remaining candidate against [references/security-and-reputation-check.md](references/security-and-reputation-check.md):
1. **Context & SemVer**: Check version delta and past repository context via git log.
2. **Changelog & Supply Chain**: Confirm release notes, official registry endpoints (guarding against typosquatting/arbitrary git URLs), and CVE/GHSA alignment.

### 1.4 Safety Halt on Doubts
If ANY warning signal is detected (typosquatting risk, unknown registry URL, unannounced major breaking change, sudden dependency tree bloat, missing changelog):
1. **Halt processing immediately** for that PR.
2. Present a **Safety Halt Report** to the user with specific diff/metadata evidence.
3. Do not proceed with that PR until explicit user confirmation is received.

Verify: The candidate PRs are filtered, audited for security/legitimacy, and classified (eligible, excluded, or halted pending review).

---

## Step 2: Formulate and Present Resolution Plan (Explore-Plan-Execute Gate)

Following [`explore-plan-execute`](../explore-plan-execute/SKILL.md), synthesize the exploration and audit findings into a structured execution plan before checking out branches or altering the repository:

1. **Construct the Resolution Plan**:
   - **Declare Execution Mode**: State whether running in `standard` or `thorough` mode.
   - **Triage Summary**: Total open PRs, eligible candidates, excluded PRs (with reasons), and any safety halts.
   - **Candidate Breakdown**: For each candidate PR:
     - PR number and package name
     - SemVer delta (patch / minor / major) and type (routine bump vs. security advisory)
     - Estimated risk and required work (e.g., clean merge expected vs. breaking changes / local test run needed)
   - **Thorough Mode In-Depth Audit (if mode is `thorough`)**:
     - Follow [references/thorough-audit-and-verification.md](references/thorough-audit-and-verification.md) for micro reviews per candidate via [`code-review`](../code-review/SKILL.md) and macro conflict/topology analysis.
   - **Execution Order**: Planned sequence of processing (e.g. security fixes first, patch updates next, minor/major updates last; or topologically ordered).
   - **Execution Strategy & API Cost Budget**:
     - State whether processing sequentially or concurrently across repositories using [references/multi-repo-scheduling.md](references/multi-repo-scheduling.md).
     - **External API Rate Limit & Cost**: Propose the external API rate-limit interval (recommended default: 10 seconds per call) and estimate total remote API calls (batch collection, CI checks, merges, default branch CI checks). Ask the user to confirm or adjust the rate limit.
     - **Default Branch CI Stability Gate**: Confirm that for single repositories with multiple PRs, each PR merge will be held until the default branch CI passes before rebasing and proceeding to the next candidate PR.

2. **Present Plan to the User**:
   ```markdown
   ### Dependency PR Resolution Plan

   **Mode**: standard / thorough
   **API Rate Limit**: 10s / call (~<X> estimated calls)

   #### Candidate Overview
   | PR # | Package | Version Bump | Type | Risk / Expected Action |
   |---|---|---|---|---|
   | #12 | `foo` | `1.0.1` -> `1.0.2` | Patch (Security) | Low; rebase, run tests, merge |
   | #15 | `bar` | `2.1.0` -> `3.0.0` | Major | High; check breaking changes, run local build |

   #### Holistic Audit & Optimization (Thorough Mode Only)
   - Co-upgrade Bundling: Packages #15 and #16.
   - Excluded / Halted: PR #10 (non-manifest), PR #14 (safety halt).

   #### Proposed Execution Order & Gating
   1. PR #12 (`foo`): Security patch -> Merge -> **Wait Default Branch CI Pass**
   2. PR #15 (`bar`): Major bump (rebase on updated default branch) -> Merge -> **Wait Default Branch CI Pass**

   May I proceed with this resolution plan and the proposed 10-second API rate limit?
   ```

3. **Await Explicit User Confirmation**:
   - Do NOT proceed to Step 3 until the user approves the plan or provides guidance on adjustments.

Verify: The user has approved the resolution plan, execution order, and scope.

---

## Step 3: Check Out Branch Using git-worktree

For each approved candidate PR/MR, isolate the worktree following [`git-worktree`](../git-worktree/SKILL.md):

1. Fetch the remote branch and latest default branch:
   ```bash
   git fetch origin <default-branch> <head-branch>
   ```
2. Create an isolated worktree under `WORKTREE_ROOT`:
   - Follow Step 1 and Step 2 of [`git-worktree`](../git-worktree/SKILL.md).
   - Checkout the PR's head branch inside the new worktree.
3. **Inspect Repository Context**:
   - Immediately check the repository's `README.md` and `AGENTS.md` (or contributing guidelines) inside the worktree to identify project-specific verification commands, derived secondary artifacts, or snapshot update targets.

Verify: The worktree is successfully created, clean, and checked out to the target PR branch.

---

## Step 4: Rebase onto Default Branch and Sync

Always execute rebase operations via local Git transport inside the worktree (rather than platform API update endpoints) to avoid API token scope limitations on workflow files:

1. Inside the worktree, check whether the branch is behind the default branch:
   ```bash
   git merge-base --is-ancestor origin/<default-branch> origin/<head-branch>
   ```
2. If not an ancestor (the PR branch is lagging behind):
   - Rebase onto the latest remote default branch:
     ```bash
     git rebase origin/<default-branch>
     ```
   - If rebase encounters merge conflicts:
     - Formulate a conflict resolution proposal.
     - Seek user confirmation before forcing resolution.
   - Push the rebased branch via authenticated Git transport:
     ```bash
     git push --force-with-lease origin <head-branch>
     ```

Verify: The branch contains the latest default branch commits and is pushed cleanly.

---

## Step 5: Verification & Multi-Repo Scheduling

### 5.1 Local-First Verification
Follow [references/ci-detection.md](references/ci-detection.md) to detect and execute project test/lint commands locally inside the worktree if runtime dependencies are available. Running local checks first prevents triggering unnecessary remote CI pipelines on obvious syntax or lockfile errors.

In **`thorough` mode**, execute exhaustive multi-faceted checks (static analysis, dependency audits, tree integrity, clean builds) per [references/thorough-audit-and-verification.md](references/thorough-audit-and-verification.md).

### 5.2 Remote CI Verification & Idle Wait Handling
1. Check remote CI status via single-shot queries (never use streaming or loop-polling commands like `--watch`):
   - **GitHub**: `gh pr checks <number>`
   - **GitLab**: `glab ci status` or `glab mr view <id>`
2. **Idle Wait Switching across Repositories**:
   - If processing across multiple repositories (or multiple PRs) and a repository enters an idle wait state (waiting for remote PR CI checks, build completion, or approval):
     - Follow [references/multi-repo-scheduling.md](references/multi-repo-scheduling.md) and [`task-dependency-pipeline`](../task-dependency-pipeline/SKILL.md).
     - Save the repository's state as `WAITING_PR_CI`.
     - **Immediately switch focus to the next queued or pending repository** to begin or resume its triage, checkout, or local verification.
     - Avoid blocking loops (e.g. `while true; do ...; sleep 10; done`); re-check pending CI status only upon transitioning between repository tasks or after the agreed rate-limit cooldown has elapsed.

Verify: Verification results from both local execution and remote CI are evaluated without idle blocking.

---

## Step 6: Remediation Plan and Approval Gate

If all CI checks pass, proceed directly to **Step 7**.

If any local or remote CI checks fail (e.g., breaking API changes, deprecations, type errors):

1. **Analyze Failure**:
   - Inspect build / test failure logs.
   - Cross-reference failure symptoms against [references/failure-archetypes.md](references/failure-archetypes.md) (outdated base branch, unsynced derived artifacts, snapshot drift) to apply standard playbooks before conjecturing code rewrites.
   - Determine the minimal adaptation needed to accommodate the upgraded dependency.
2. **Present Remediation Plan (Human-in-the-Loop)**:
   - Present a concise plan modeled after [`step-gate`](../step-gate/SKILL.md) detailing Failure Summary, Proposed Changes, and Verification commands.
   - Await explicit user approval before modifying files.
3. **Execute & Sync**:
   - Upon user approval, apply the minimal code adjustments.
   - Format commit messages adhering to [`change-message`](../change-message/SKILL.md).
   - Push to remote and return to **Step 5** to re-verify.
   - If the required refactor is extensive or changes behavioral contracts, propose handing off to [`refactor`](../refactor/SKILL.md) and defer merging.

Verify: The user has approved the remediation plan before any file modifications are applied, and tests pass after changes.

---

## Step 7: Merge, Verify Default Branch CI, and Clean Up

Once remote and local CI checks succeed:

### 7.1 Merge the PR/MR
1. **Pre-Merge Base Drift Check**: Verify the branch tip is a fast-forward descendant of latest default branch. If behind, rebase in worktree per Step 4 to satisfy fast-forward branch protection rules without resorting to `--admin`.
2. **Merge PR/MR**: Merge adhering to the agreed API rate limit:
   - **GitHub**: `gh pr merge <number> --auto --squash` (or `--merge` / `--rebase` per repository rule)
   - **GitLab**: `glab mr merge <id> --auto-merge`

### 7.2 Post-Merge Default Branch CI Verification Gate
Before proceeding to the next candidate PR in the same repository or concluding the workflow, verify that the default branch CI triggered by the merge commit passes:

1. Identify and inspect the default branch CI run triggered by the merge commit per [references/ci-detection.md](references/ci-detection.md) (e.g., `gh run list --branch <default-branch> --commit "$MERGE_SHA" --limit 1` or `glab ci status --branch <default-branch>`).
2. **Idle Wait Handling**:
   - While the default branch CI is running (`WAITING_DEFAULT_CI`), do NOT block in a loop.
   - If processing across multiple repositories, immediately switch focus to another repository per [references/multi-repo-scheduling.md](references/multi-repo-scheduling.md).
3. **Verify Pass**:
   - Check the run status using single-shot commands after an appropriate interval.
   - Once the default branch CI succeeds:
     - If more candidate PRs remain in this repository: rebase the next candidate PR onto this new default branch commit and proceed to Step 4.
     - If all approved candidate PRs in this repository are resolved: proceed to cleanup.
4. **Safety Halt on Default Branch Failure**:
   - If default branch CI fails, freeze all further PR merges immediately.
   - Present a Safety Halt Incident Report to the user with failure logs, and offer remediation options (e.g., revert the merge commit via `git revert <merge-sha>`, or create an immediate patch).

### 7.3 Clean Up Worktree
Clean up the worktree and head branch following [`git-worktree`](../git-worktree/SKILL.md) (`git worktree remove <worktree-path>`, `git branch -D <head-branch>`).

### 7.4 Follow-Up Skills (Sequential Hand-off)
- **For post-merge working tree verification**: Recommend [`project-status`](../project-status/SKILL.md).
- **For generating release notes or summary messages**: Recommend [`change-message`](../change-message/SKILL.md).

Verify: The PR/MR is merged, default branch CI is verified green, and the temporary worktree is completely removed.

---

## Guidelines

- **Explore & Plan First**: Never mutate branches without an agreed plan. Triage candidates, batch API queries upfront, and confirm execution order with the user.
- **External API Rate Limit & Cost Tracking**: Pre-confirm API rate limits (default: 10s/call), track total API calls, and avoid polling loops.
- **Post-Merge Default Branch CI Stability**: In a single repository, process PRs strictly sequentially. Never advance until the default branch CI for the prior merge commit passes. Halt immediately on failure.
- **Safety First on Doubts**: If there is any suspicion regarding package identity, registry URLs, unexpected dependencies, or security advisories, halt and request user confirmation immediately.
- **No Idle Blocking in Multi-Repo Tasks**: When waiting for CI (either PR CI or post-merge default branch CI) or user input in one repository, switch to another repository immediately to continue work.
- **Never Force Without Lease**: Always use `--force-with-lease` when updating rebased remote branches.
- **Declare Execution Mode Upfront**: Explicitly identify whether running in `standard` or `thorough` mode during plan presentation.
- **Thorough Mode Holistic Scrutiny**: In `thorough` mode, inspect changelogs, audit breaking changes, detect cross-PR peer conflicts, and mandate comprehensive verification per [references/thorough-audit-and-verification.md](references/thorough-audit-and-verification.md).
- **Strict Scope Preservation**: If updates require logic rewrites beyond minor API adjustments, pause and consult or hand off to [`refactor`](../refactor/SKILL.md).
- **No Leftover Worktrees**: Clean up worktrees upon completion to avoid disk clutter.
