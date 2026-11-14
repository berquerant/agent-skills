---
name: resolve-dependency-prs
description: >-
  Use this skill to triage, audit, verify, and merge dependency update or
  security advisory Pull Requests / Merge Requests on GitHub or GitLab. Enforces
  cost-aware local batching, legitimacy & security audits, safety halt on doubts,
  idle-wait switching across repositories, worktree isolation, and CI verification.
---

# Resolve Dependency PRs

Triage open dependency and security PRs/MRs, audit legitimacy and safety, verify in isolated worktrees, validate through CI, remediate failures under user approval, manage multi-repo and monorepo pipelines without idle waiting, and merge safely.

Coordinates [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md) for upfront planning, [git-worktree](../git-worktree/SKILL.md) (../git-worktree/SKILL.md) for branch isolation, [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md) for approvable script encapsulation of merge, cleanup, and verification workflows, [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md) for remediation gates, [task-dependency-pipeline](../task-dependency-pipeline/SKILL.md) (../task-dependency-pipeline/SKILL.md) for cooperative idle-wait scheduling, and [code-review](../code-review/SKILL.md) (../code-review/SKILL.md) for thorough mode audits.

---

## Execution Modes

This skill supports two execution modes:
- **`standard` (Default)**: Standard automated triage, manifest-level security check, rebase, local test/lint, CI verification, and safe merge.
- **`thorough`**: Deep micro reviews per candidate via [code-review](../code-review/SKILL.md) (../code-review/SKILL.md), holistic macro conflict detection, co-upgrade bundling, and exhaustive multi-faceted verification prior to merge per [thorough-audit-and-verification.md](references/thorough-audit-and-verification.md) (references/thorough-audit-and-verification.md).

---

## Prerequisites

- `git` is installed and the current directory is inside a git repository (or target repositories are available locally).
- A supported platform tool is authenticated (`gh` CLI for GitHub or `glab` CLI for GitLab; refer to [platform-tools.md](../multi-repo-change/references/platform-tools.md) (../multi-repo-change/references/platform-tools.md)).
- **Cost & Rate Limits**: Follow cost-aware principles (`AGENTS.md` Rule 5 and [command-cache-protocol.md](../explore-plan-execute/references/command-cache-protocol.md) (../explore-plan-execute/references/command-cache-protocol.md)). Confirm API rate limit (default: 10s/call) and execution mode (`standard` or `thorough`).
- **Explore & Plan First**: Follow [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md) to triage candidates and present a concrete resolution plan before modifying branches.

---

## Step 1: Collect, Filter, and Audit Candidate PRs/MRs

### 1.1 Cost-Aware Batch Collection
Retrieve open PRs/MRs along with files and body in a single query to avoid repetitive platform API calls. Encapsulate batch retrieval and local filtering into a self-contained data pipeline snippet (per [approvable-patterns.md](../command-snippet/references/approvable-patterns.md) (../command-snippet/references/approvable-patterns.md) Example E) or query locally per [filter-criteria.md](references/filter-criteria.md) (references/filter-criteria.md):
- **GitHub**: Fetch open PRs to `candidate_prs.json` via dedicated retrieval snippet.
- **GitLab**: `glab mr list --state opened > candidate_mrs.txt` or `glab api "projects/:id/merge_requests?state=opened" > candidate_mrs.json`

### 1.2 File-Level Exclusivity Check
Filter PRs according to [filter-criteria.md](references/filter-criteria.md) (references/filter-criteria.md):
- Ensure changes are strictly confined to package manager manifests and lockfiles.
- **Immediately exclude** any PR modifying application code, test logic, CI configurations, or general project files.

### 1.3 Context, Legitimacy & Security Audit
Audit each remaining candidate against [security-and-reputation-check.md](references/security-and-reputation-check.md) (references/security-and-reputation-check.md):
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

Following [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md), synthesize the exploration and audit findings into a structured execution plan before checking out branches or altering the repository:

1. **Construct the Resolution Plan**:
   - **Declare Execution Mode**: State whether running in `standard` or `thorough` mode.
   - **Triage Summary**: Total open PRs, eligible candidates, excluded PRs (with reasons), and any safety halts.
   - **Candidate Breakdown**: For each candidate PR:
     - PR number and package name
     - SemVer delta (patch / minor / major) and type (routine bump vs. security advisory)
     - Estimated risk and required work (e.g., clean merge expected vs. breaking changes / local test run needed)
   - **Thorough Mode In-Depth Audit (if mode is `thorough`)**:
     - Follow [thorough-audit-and-verification.md](references/thorough-audit-and-verification.md) (references/thorough-audit-and-verification.md) for micro reviews per candidate via [code-review](../code-review/SKILL.md) (../code-review/SKILL.md) and macro conflict/topology analysis.
   - **Execution Order**: Planned sequence of processing (e.g. security fixes first, patch updates next, minor/major updates last; or topologically ordered).
   - **Execution Strategy & API Cost Budget**:
     - State whether processing sequentially or concurrently across repositories using [multi-repo-scheduling.md](references/multi-repo-scheduling.md) (references/multi-repo-scheduling.md).
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

For each approved candidate PR/MR, isolate the worktree following [git-worktree](../git-worktree/SKILL.md) (../git-worktree/SKILL.md):

1. Fetch the remote branch and latest default branch:
   ```bash
   git fetch origin <default-branch> <head-branch>
   ```
2. Create an isolated worktree under `WORKTREE_ROOT`:
   - Follow Step 1 and Step 2 of [git-worktree](../git-worktree/SKILL.md) (../git-worktree/SKILL.md).
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
Follow [ci-detection.md](references/ci-detection.md) (references/ci-detection.md) to detect and execute project test/lint commands locally inside the worktree if runtime dependencies are available.
- **Task Discovery Protocol**: Before running local task runner commands, always verify defined task names (e.g. `./task --list`, `npm run`, or `make -qp`) per [ci-detection.md](references/ci-detection.md) (references/ci-detection.md) rather than guessing task names.
- **Composite Execution**: Encapsulate multi-step verification into an atomic workflow snippet (e.g. `rebase_and_test_worktree.sh`) per [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md).

In **`thorough` mode**, execute exhaustive multi-faceted checks (static analysis, dependency audits, tree integrity, clean builds) per [thorough-audit-and-verification.md](references/thorough-audit-and-verification.md) (references/thorough-audit-and-verification.md).

### 5.2 Remote CI Verification & Idle Wait Handling
1. Check remote CI status via single-shot queries (never use streaming or loop-polling commands like `--watch`):
   - **GitHub**: `gh pr checks <number>`
   - **GitLab**: `glab ci status` or `glab mr view <id>`
2. **Idle Wait Switching across Repositories & Independent Sub-Components**:
   - If processing across multiple repositories OR across independent, disjoint sub-components within the same repository (e.g., UI vs Go backend) and an item enters an idle wait state (`WAITING_PR_CI` or `WAITING_DEFAULT_CI`):
     - Follow [multi-repo-scheduling.md](references/multi-repo-scheduling.md) (references/multi-repo-scheduling.md) and [task-dependency-pipeline](../task-dependency-pipeline/SKILL.md) (../task-dependency-pipeline/SKILL.md).
     - Save the item's state as `WAITING_PR_CI`.
     - **Immediately switch focus to the next queued or pending repository or independent sub-component** to begin or resume its triage, checkout, or local verification.
     - Avoid blocking loops or repetitive sleep timers; check pending CI status upon transitioning between tasks or after the agreed rate-limit cooldown has elapsed.

Verify: Verification results from both local execution and remote CI are evaluated without idle blocking.

---

## Step 6: Remediation Plan and Approval Gate

If all CI checks pass, proceed directly to **Step 7**.

If any local or remote CI checks fail (e.g., breaking API changes, deprecations, type errors):

1. **Analyze Failure**:
   - Inspect build / test failure logs.
   - Cross-reference failure symptoms against [failure-archetypes.md](references/failure-archetypes.md) (references/failure-archetypes.md) (outdated base branch, unsynced derived artifacts, snapshot drift) to apply standard playbooks before conjecturing code rewrites.
   - Determine the minimal adaptation needed to accommodate the upgraded dependency.
2. **Present Remediation Plan (Human-in-the-Loop)**:
   - Present a concise plan modeled after [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md) detailing Failure Summary, Proposed Changes, and Verification commands.
   - Await explicit user approval before modifying files.
3. **Execute & Sync**:
   - Upon user approval, apply the minimal code adjustments.
   - Format commit messages adhering to [change-message](../change-message/SKILL.md) (../change-message/SKILL.md).
   - Push to remote and return to **Step 5** to re-verify.
   - If the required refactor is extensive or changes behavioral contracts, propose handing off to [refactor](../refactor/SKILL.md) (../refactor/SKILL.md) and defer merging.

Verify: The user has approved the remediation plan before any file modifications are applied, and tests pass after changes.

---

## Step 7: Merge, Verify Default Branch CI, and Clean Up

Once remote and local CI checks succeed:

### 7.1 Merge the PR/MR via Approvable Snippet
1. **Pre-Merge Base Drift Check**: Verify the branch tip is a fast-forward descendant of latest default branch. If behind, rebase in worktree per Step 4 to satisfy fast-forward branch protection rules without resorting to `--admin`.
2. **Execute Merge via Approvable Snippet**:
   - Do **not** execute raw inline merge commands (e.g., inline `gh pr merge ...`). Encapsulate the operation into a self-descriptive snippet (e.g., `merge_pr_squash.sh`) per [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md).
   - Invoke directly as a standalone binary (`./merge_pr_squash.sh <pr-num>` or `<path> <pr-num>`) preceded by the mandatory Pre-Execution Notice. Do **not** use generic interpreter prefixes (`bash`) or compound `chmod +x ... && exec ...` patterns.

### 7.2 Post-Merge Default Branch CI Verification Gate
Before proceeding to the next dependent candidate PR in the same repository or concluding the workflow, verify that the default branch CI triggered by the merge commit passes:

1. Identify and inspect the default branch CI run triggered by the merge commit per [ci-detection.md](references/ci-detection.md) (references/ci-detection.md) (e.g., `gh run list --branch <default-branch> --commit "$MERGE_SHA" --limit 1` or `glab ci status --branch <default-branch>`).
2. **Idle Wait Interleaving**: While default branch CI is running (`WAITING_DEFAULT_CI`), do NOT block in an idle loop or timer sleep; if other repositories or independent sub-components have work pending, switch focus immediately per [multi-repo-scheduling.md](references/multi-repo-scheduling.md) (references/multi-repo-scheduling.md).
3. **Verify Pass**: Check the run status using single-shot commands after an appropriate interval. Once default branch CI succeeds, rebase the next candidate PR in this component onto the new default branch commit (Step 4) or proceed to cleanup if all approved PRs are resolved.
4. **Safety Halt on Default Branch Failure**: If default branch CI fails, freeze all further PR merges immediately, present a Safety Halt Incident Report with failure logs, and offer remediation options (`git revert <merge-sha>` or immediate patch).

### 7.3 Clean Up Worktree via Approvable Snippet
Clean up the worktree and head branch via a dedicated cleanup snippet (e.g., `cleanup_pr_worktree.sh`) invoked directly per [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md) rather than running inline chained deletion commands.

### 7.4 Follow-Up Skills (Sequential Hand-off)
- **For post-merge working tree verification**: Recommend [project-status](../project-status/SKILL.md) (../project-status/SKILL.md).
- **For generating release notes or summary messages**: Recommend [change-message](../change-message/SKILL.md) (../change-message/SKILL.md).

Verify: The PR/MR is merged, default branch CI is verified green, and the temporary worktree is completely removed using approvable snippets.

---

## Guidelines

- **Approve the triage and resolution plan first**: Batch-collect candidates, audit manifests, and secure explicit user sign-off in Step 2 before checking out branches or rebasing.
- **Respect API rate limits and local caching**: Pre-confirm API rate limits (default: 10s/call), save query responses locally, and use single-shot CI checks to avoid rate-limit exhaustion.
- **Execute state changes via approvable snippets directly**: Encapsulate state-mutating actions (merge, worktree cleanup, rebase) into self-descriptive snippet files per [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md). Invoke directly as standalone binaries (`./<name>.sh [args...]`); strictly avoid raw inline merges, generic interpreter prefixes (`bash`), and compound `chmod +x ... && exec ...` patterns.
- **Discover task runner targets before execution**: Confirm exact task runner target syntax via `./task --list`, `npm run`, or `make -qp` per [ci-detection.md](references/ci-detection.md) (references/ci-detection.md) to prevent guessing errors.
- **Parallelize independent sub-components in monorepos**: In repositories with disjoint components (e.g. UI vs backend), parallelize worktree setup and verification, and interleave idle CI wait states across components per [multi-repo-scheduling.md](references/multi-repo-scheduling.md) (references/multi-repo-scheduling.md).
- **Verify default-branch CI after every merge**: Process dependent PRs sequentially within a component and wait for the prior merge commit's default-branch CI to pass before rebasing the next PR.
- **Halt immediately on supply-chain or security doubts**: Pause and present a Safety Halt Report whenever package names, registry URLs, or changelogs look suspicious.
- **Switch context during CI waits**: Interleave work across pending repositories or independent sub-components per [multi-repo-scheduling.md](references/multi-repo-scheduling.md) (references/multi-repo-scheduling.md) rather than idling on remote CI.
- **Push rebased branches with `--force-with-lease`**: Use `--force-with-lease` so concurrent remote updates are never overwritten.
- **Declare `standard` vs. `thorough` mode upfront**: State the selected depth in Step 2 and follow [thorough-audit-and-verification.md](references/thorough-audit-and-verification.md) (references/thorough-audit-and-verification.md) when `thorough` is chosen.
- **Hand off non-trivial code adaptations**: Delegate broader code rewrites to [refactor](../refactor/SKILL.md) (../refactor/SKILL.md) and clean up temporary worktrees upon completion.
