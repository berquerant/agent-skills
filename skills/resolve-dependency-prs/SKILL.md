---
name: resolve-dependency-prs
description: >-
  Use this skill to triage, audit, verify, and merge dependency update or
  security advisory Pull Requests / Merge Requests on GitHub or GitLab. Enforces
  cost-aware local batching, legitimacy & security audits, safety halt on doubts,
  idle-wait switching across repositories, worktree isolation, and CI verification.
---

# Resolve Dependency PRs

Triage open dependency update and security patch PRs/MRs, audit their legitimacy and security safety, verify them in isolated worktrees, ensure they are up to date with the default branch, validate through CI, remediate failures under user approval, manage multi-repository pipelines without idle waiting, and safely merge.

This workflow coordinates [`explore-plan-execute`](../explore-plan-execute/SKILL.md) to explore candidates and establish an agreed execution plan before modifying any branches, [`git-worktree`](../git-worktree/SKILL.md) for branch isolation, and [`step-gate`](../step-gate/SKILL.md) for remediation approval gates without duplicating procedures.

---

## Prerequisites

- `git` is installed and the current directory is inside a git repository (or target repositories are available locally).
- A supported platform tool is authenticated:
  - GitHub: `gh` CLI (or `github` MCP server) as described in [`multi-repo-change/references/platform-tools.md`](../multi-repo-change/references/platform-tools.md).
  - GitLab: `glab` CLI (or `gitlab` MCP server) as described in [`multi-repo-change/references/platform-tools.md`](../multi-repo-change/references/platform-tools.md).
- Follow cost-aware principles (`AGENTS.md` Rule 5): batch API queries, cache candidate state locally, avoid excessive CI polling, and rely strictly on local references and PR descriptions.
- **Explore & Plan First**: Do not jump directly into creating worktrees or checking out branches upon invocation. Adhere to [`explore-plan-execute`](../explore-plan-execute/SKILL.md) by exploring open PRs, auditing requirements, and presenting a concrete triage and resolution plan for user approval before modifying any local branches.

---

## Step 1: Collect, Filter, and Audit Candidate PRs/MRs

### 1.1 Cost-Aware Batch Collection
Retrieve open PRs/MRs along with files and body in a single query to avoid repetitive platform API calls. Save candidate metadata into a local JSON file or cache first, and inspect candidate details from that local file rather than calling remote APIs per candidate:
- **GitHub**:
  ```bash
  # Query all candidate PRs in a single API call and save locally
  gh pr list --state open --json number,title,headRefName,author,labels,files,body > candidate_prs.json
  ```
- **GitLab**:
  ```bash
  # Query candidate MRs via glab and save locally (glab api / glab mr list)
  glab mr list --state opened > candidate_mrs.txt
  # Or fetch JSON payload via glab api
  glab api "projects/:id/merge_requests?state=opened" > candidate_mrs.json
  ```

Inspect, filter, and audit candidates against the saved file (`candidate_prs.json` / `candidate_mrs.json`) locally. Avoid issuing separate API queries for each candidate PR.

### 1.2 File-Level Exclusivity Check
Filter PRs according to [references/filter-criteria.md](references/filter-criteria.md):
- Ensure changes are strictly confined to package manager manifests and lockfiles.
- **Immediately exclude** any PR modifying application code, test logic, CI configurations, or general project files.

### 1.3 Context, Legitimacy & Security Audit
Audit each remaining candidate against [references/security-and-reputation-check.md](references/security-and-reputation-check.md):
1. **Context & SemVer Check**: Check version delta (patch, minor, major) and verify past repository context via local git history (e.g. `git log -S "<pkg>"`).
2. **PR Description & Changelog Review**: Confirm release notes or changelogs are referenced, and check for deprecations or license alterations.
3. **Security & Supply Chain Audit**:
   - Check exact package name match (guard against typosquatting).
   - Check lockfile download URLs to confirm official package registry endpoints (guard against arbitrary git URLs or unvetted hosts).
   - Confirm security vulnerability alignment (CVE / GHSA patched version matches declared bump).

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
   - **Triage Summary**: Total open PRs, eligible candidates, excluded PRs (with reasons), and any safety halts.
   - **Candidate Breakdown**: For each candidate PR:
     - PR number and package name
     - SemVer delta (patch / minor / major) and type (routine bump vs. security advisory)
     - Estimated risk and required work (e.g., clean merge expected vs. breaking changes / local test run needed)
   - **Execution Order**: Planned sequence of processing (e.g. security fixes first, patch updates next, minor/major updates last).
   - **Execution Strategy**: Whether processing sequentially or concurrently across repositories using [references/multi-repo-scheduling.md](references/multi-repo-scheduling.md).

2. **Present Plan to the User**:
   ```markdown
   ### Dependency PR Resolution Plan

   #### Candidate Overview
   | PR # | Package | Version Bump | Type | Risk / Expected Action |
   |---|---|---|---|---|
   | #12 | `foo` | `1.0.1` -> `1.0.2` | Patch (Security) | Low; rebase, run tests, merge |
   | #15 | `bar` | `2.1.0` -> `2.2.0` | Minor | Low; rebase, verify CI |
   | #18 | `baz` | `3.0.0` -> `4.0.0` | Major | High; check breaking changes, run local build |

   #### Excluded / Halted PRs
   - PR #10: Excluded (modifies non-manifest files `src/index.ts`)
   - PR #14: Halted (typosquatting concern or unverified registry URL)

   #### Proposed Execution Order
   1. PR #12 (`foo`): Security patch
   2. PR #15 (`bar`): Minor bump
   3. PR #18 (`baz`): Major bump

   May I proceed with this resolution plan?
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

Verify: The worktree is successfully created, clean, and checked out to the target PR branch.

---

## Step 4: Rebase onto Default Branch and Sync

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
   - Push the rebased branch:
     ```bash
     git push --force-with-lease origin <head-branch>
     ```

Verify: The branch contains the latest default branch commits and is pushed cleanly.

---

## Step 5: Verification & Multi-Repo Scheduling

### 5.1 Local-First Verification
Follow [references/ci-detection.md](references/ci-detection.md) to detect and execute project test/lint commands locally inside the worktree if runtime dependencies are available. Running local checks first prevents triggering unnecessary remote CI pipelines on obvious syntax or lockfile errors.

### 5.2 Remote CI Verification & Idle Wait Handling
1. Check remote CI status via single-shot queries (never use streaming or loop-polling commands like `--watch`):
   - **GitHub**:
     ```bash
     # Single-shot check; do NOT use `gh pr checks <number> --watch`
     gh pr checks <number>
     ```
   - **GitLab**:
     ```bash
     # Single-shot check; query pipeline status once
     glab ci status
     # Or view MR pipeline details once
     glab mr view <id>
     ```
2. **Idle Wait Switching across Repositories**:
   - If processing across multiple repositories (or multiple PRs) and a repository enters an idle wait state (waiting for remote CI checks, build completion, or approval):
     - Follow [references/multi-repo-scheduling.md](references/multi-repo-scheduling.md).
     - Save the repository's state as `WAITING_CI`.
     - **Immediately switch focus to the next queued or pending repository** to begin or resume its triage, checkout, or local verification.
     - Avoid blocking loops (e.g. `while true; do ...; sleep 10; done`); re-check pending CI status only upon transitioning between repository tasks or after an appropriate interval.

Verify: Verification results from both local execution and remote CI are evaluated without idle blocking.

---

## Step 6: Remediation Plan and Approval Gate

If all CI checks pass, proceed directly to **Step 7**.

If any local or remote CI checks fail (e.g., breaking API changes, deprecations, type errors):

1. **Analyze Failure**:
   - Inspect build / test failure logs.
   - Determine the minimal adaptation needed to accommodate the upgraded dependency.
2. **Present Remediation Plan (Human-in-the-Loop)**:
   - Follow the gate structure modeled after [`step-gate`](../step-gate/SKILL.md):
   ```markdown
   ### Remediation Plan for PR #<number> (<package-name>)
   - **Failure Summary**: <Reason for CI failure, e.g. compiler error, deprecated method>
   - **Proposed Changes**: <Exact file modifications and rationale>
   - **Verification**: <Commands to run locally and in CI to confirm resolution>
   ```
   - Await explicit user approval before modifying files.
3. **Execute & Sync**:
   - Upon user approval, apply the minimal code adjustments.
   - Format commit messages adhering to [`change-message`](../change-message/SKILL.md).
   - Push to remote and return to **Step 5** to re-verify.
   - If the required refactor is extensive or changes behavioral contracts, propose handing off to [`refactor`](../refactor/SKILL.md) and defer merging.

Verify: The user has approved the remediation plan before any file modifications are applied, and tests pass after changes.

---

## Step 7: Merge and Clean Up

Once remote and local CI checks succeed:

1. Merge the PR/MR:
   - **GitHub**:
     ```bash
     gh pr merge <number> --auto --squash # or --merge / --rebase per repository rule
     ```
   - **GitLab**:
     ```bash
     glab mr merge <id> --auto-merge
     ```
2. Clean up the worktree following [`git-worktree`](../git-worktree/SKILL.md):
   ```bash
   git worktree remove <worktree-path>
   git branch -D <head-branch>
   ```

Verify: The PR/MR is merged (or queued for auto-merge), and the temporary worktree is completely removed.

---

## Guidelines

- **Explore & Plan First**: Never jump directly into branch checkouts or modifications. Always triage open PRs, formulate an explicit resolution plan with candidate breakdown and execution order, and obtain user confirmation first.

- **Minimize Total Cost**: Batch API requests upfront (`gh pr list --json ... > prs.json`), cache candidate metadata locally, inspect details locally rather than querying remote APIs per PR, prioritize fast local checks over remote CI runs, and strictly avoid tight polling loops (`--watch` or sleep loops).
- **Safety First on Doubts**: If there is any suspicion regarding package identity, registry URLs, unexpected dependencies, or security advisories, halt and request user confirmation immediately.
- **No Idle Blocking in Multi-Repo Tasks**: When waiting for CI or user input in one repository, switch to another repository to continue work.
- **Never Force Without Lease**: Always use `--force-with-lease` when updating rebased remote branches.
- **Strict Scope Preservation**: If a dependency update requires substantial application logic rewrites beyond minor API signature adjustments, pause and consult the user or hand off to [`refactor`](../refactor/SKILL.md).
- **No Leftover Worktrees**: Always remove worktrees upon completion or cancellation to avoid disk clutter.
