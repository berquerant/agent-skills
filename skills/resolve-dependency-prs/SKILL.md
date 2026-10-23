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

This workflow coordinates [`explore-plan-execute`](../explore-plan-execute/SKILL.md) to explore candidates and establish an agreed execution plan before modifying any branches, [`git-worktree`](../git-worktree/SKILL.md) for branch isolation, and [`step-gate`](../step-gate/SKILL.md) for remediation approval gates without duplicating procedures. In `thorough` mode, it also consults review criteria from [`code-review`](../code-review/SKILL.md) for deep micro/macro audits.

---

## Execution Modes

This skill supports two execution modes:

- **`standard` (Default)**: Standard automated triage, manifest-level security check, rebase, local test/lint, CI verification, and safe merge.
- **`thorough`**: In-depth micro-level code reviews per candidate (via [`code-review`](../code-review/SKILL.md)), holistic macro reviews of the entire candidate set (detecting transitive conflicts, bundling co-upgrades, optimizing execution order, recommending holds/deprecations), and exhaustive multi-faceted local and environmental verification prior to merge.

---

## Prerequisites

- `git` is installed and the current directory is inside a git repository (or target repositories are available locally).
- A supported platform tool is authenticated:
  - GitHub: `gh` CLI (or `github` MCP server) as described in [`multi-repo-change/references/platform-tools.md`](../multi-repo-change/references/platform-tools.md).
  - GitLab: `glab` CLI (or `gitlab` MCP server) as described in [`multi-repo-change/references/platform-tools.md`](../multi-repo-change/references/platform-tools.md).
- Follow cost-aware principles (`AGENTS.md` Rule 5 and [`command-cache-protocol.md`](../explore-plan-execute/references/command-cache-protocol.md)): batch API queries, cache candidate state locally (leveraging `cache-memoizer` if available), avoid excessive CI polling, and rely strictly on local references and PR descriptions.
- **Confirm API Rate Limit & Cost Tracking**:
  External API interactions (`gh`, `glab`, REST APIs) must respect a rate-limit interval to prevent hitting secondary rate limits or 429 errors. The recommended default is **10 seconds per call**. Because this 10-second interval is an example default, explicitly confirm with the user during Step 2 planning what rate limit interval is appropriate for their environment, and treat external API calls as a tracked cost metric.
- **Determine Execution Mode**: Check if the user specified `standard` or `thorough`. If unspecified, default to `standard` (or present `thorough` as an option in Step 2).
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
   - **Declare Execution Mode**: State whether running in `standard` or `thorough` mode.
   - **Triage Summary**: Total open PRs, eligible candidates, excluded PRs (with reasons), and any safety halts.
   - **Candidate Breakdown**: For each candidate PR:
     - PR number and package name
     - SemVer delta (patch / minor / major) and type (routine bump vs. security advisory)
     - Estimated risk and required work (e.g., clean merge expected vs. breaking changes / local test run needed)
   - **Thorough Mode In-Depth Audit (if mode is `thorough`)**:
     - Follow [references/thorough-audit-and-verification.md](references/thorough-audit-and-verification.md).
     - **Micro Code-Review**: Deep review per candidate using [`code-review`](../code-review/SKILL.md) standards (changelogs, breaking changes, deprecations, supply chain security).
     - **Macro Holistic Review**: Review the collective set of candidate PRs:
       - Detect peer dependency conflicts or transitive duplication across candidates.
       - Identify co-upgrade bundling opportunities (e.g. bundling related ecosystem libraries into a single branch/PR).
       - Optimize execution ordering based on dependency topology.
       - Identify updates recommended for holding or rejection (e.g. unstable upstream, massive breaking refactors).
   - **Execution Order**: Planned sequence of processing (e.g. security fixes first, patch updates next, minor/major updates last; or topologically ordered).
   - **Execution Strategy & API Cost Budget**:
     - State whether processing sequentially or concurrently across repositories using [references/multi-repo-scheduling.md](references/multi-repo-scheduling.md).
     - **External API Rate Limit & Cost**: Propose the external API rate-limit interval (recommended default: 10 seconds per call) and estimate total remote API calls (batch collection, CI checks, merges, default branch CI checks). Ask the user to confirm or adjust the rate limit.
     - **Default Branch CI Stability Gate**: Confirm that for single repositories with multiple PRs, each PR merge will be held until the default branch CI passes before rebasing and proceeding to the next candidate PR.

2. **Present Plan to the User**:
   ```markdown
   ### Dependency PR Resolution Plan

   **Mode**: standard / thorough
   **API Rate Limit**: 10 seconds / call (default; please adjust if needed)
   **Estimated API Calls**: ~<X> calls (~<X*10>s API time)

   #### Candidate Overview
   | PR # | Package | Version Bump | Type | Risk / Expected Action |
   |---|---|---|---|---|
   | #12 | `foo` | `1.0.1` -> `1.0.2` | Patch (Security) | Low; rebase, run tests, merge |
   | #15 | `bar` | `2.1.0` -> `2.2.0` | Minor | Low; rebase, verify CI |
   | #18 | `baz` | `3.0.0` -> `4.0.0` | Major | High; check breaking changes, run local build |

   #### Holistic Audit & Optimization (Thorough Mode Only)
   - Co-upgrade Bundling: Recommend bundling PR #15 and PR #16 together.
   - Hold Recommendations: Hold PR #18 due to breaking API shifts; recommend separate refactor branch.

   #### Excluded / Halted PRs
   - PR #10: Excluded (modifies non-manifest files `src/index.ts`)
   - PR #14: Halted (typosquatting concern or unverified registry URL)

   #### Proposed Execution Order & Gating
   1. PR #12 (`foo`): Security patch -> Merge -> **Wait Default Branch CI Pass**
   2. PR #15 (`bar`): Minor bump (rebase onto updated default branch) -> Merge -> **Wait Default Branch CI Pass**
   3. PR #18 (`baz`): Major bump (rebase onto updated default branch) -> Merge -> **Wait Default Branch CI Pass**

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

In **`thorough` mode**, local tests passing is considered a baseline requirement. Execute exhaustive multi-faceted checks per [references/thorough-audit-and-verification.md](references/thorough-audit-and-verification.md):
- **Type Checking / Static Analysis**: e.g., `npx tsc --noEmit`, `go vet ./...`, `cargo check`, `mypy .`
- **Security & Dependency Audit Scans**: e.g., `npm audit`, `cargo audit`, `govulncheck ./...`, `pip-audit`
- **Lockfile & Tree Integrity**: e.g., `npm ls`, `pnpm doctor`, `go mod verify`
- **Clean Build Verification**: Purge build caches and verify a complete scratch build (`npm run build`, `cargo build --release`, `go build ./...`)

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
   - If processing across multiple repositories (or multiple PRs) and a repository enters an idle wait state (waiting for remote PR CI checks, build completion, or approval):
     - Follow [references/multi-repo-scheduling.md](references/multi-repo-scheduling.md).
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

## Step 7: Merge, Verify Default Branch CI, and Clean Up

Once remote and local CI checks succeed:

### 7.1 Merge the PR/MR
Merge the PR/MR adhering to the agreed API rate limit (cooldown check):
- **GitHub**:
  ```bash
  gh pr merge <number> --auto --squash # or --merge / --rebase per repository rule
  ```
- **GitLab**:
  ```bash
  glab mr merge <id> --auto-merge
  ```

### 7.2 Post-Merge Default Branch CI Verification Gate
Before proceeding to the next candidate PR in the same repository or concluding the workflow, verify that the default branch CI triggered by the merge commit passes:

1. Identify the default branch CI run triggered by the merge commit per [references/ci-detection.md](references/ci-detection.md):
   - **GitHub**:
     ```bash
     # Retrieve merge commit SHA and inspect the triggered workflow run
     MERGE_SHA=$(gh pr view <number> --json mergeCommit --jq .mergeCommit.oid 2>/dev/null || git rev-parse origin/<default-branch>)
     gh run list --branch <default-branch> --commit "$MERGE_SHA" --limit 1
     ```
   - **GitLab**:
     ```bash
     glab ci status --branch <default-branch>
     ```
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
Clean up the worktree following [`git-worktree`](../git-worktree/SKILL.md):
```bash
git worktree remove <worktree-path>
git branch -D <head-branch>
```

### 7.4 Follow-Up Skills (Sequential Hand-off)
- **For post-merge working tree verification**: Recommend [`project-status`](../project-status/SKILL.md).
- **For generating release notes or summary messages**: Recommend [`change-message`](../change-message/SKILL.md).

Verify: The PR/MR is merged, default branch CI is verified green, and the temporary worktree is completely removed.

---

## Guidelines

- **Explore & Plan First**: Never jump directly into branch checkouts or modifications. Always triage open PRs, formulate an explicit resolution plan with candidate breakdown and execution order, and obtain user confirmation first.
- **External API Rate Limit & Cost Tracking**: Always confirm the API rate limit with the user during planning (default recommended: 10s per call). Treat API requests as billable costs, track total calls, and use cooldowns to execute zero-API local operations.
- **Post-Merge Default Branch CI Stability**: In any single repository, multiple PRs must be processed strictly sequentially. Never advance to the next candidate PR until the default branch CI triggered by the prior merge commit passes. Halt immediately if default branch CI breaks.
- **Minimize Total Cost**: Batch API requests upfront (`gh pr list --json ... > prs.json`), cache candidate metadata locally, inspect details locally rather than querying remote APIs per PR, prioritize fast local checks over remote CI runs, and strictly avoid tight polling loops (`--watch` or sleep loops).
- **Safety First on Doubts**: If there is any suspicion regarding package identity, registry URLs, unexpected dependencies, or security advisories, halt and request user confirmation immediately.
- **No Idle Blocking in Multi-Repo Tasks**: When waiting for CI (either PR CI or post-merge default branch CI) or user input in one repository, switch to another repository immediately to continue work.
- **Never Force Without Lease**: Always use `--force-with-lease` when updating rebased remote branches.
- **Declare Execution Mode Upfront**: Explicitly identify whether running in `standard` or `thorough` mode during plan presentation.
- **Thorough Mode Holistic Scrutiny**: In `thorough` mode, treat single-PR merges as part of the broader dependency graph: inspect changelogs, audit breaking changes, detect cross-PR peer conflicts, consider bundling co-upgrades, and mandate comprehensive local verification (type checks, audits, clean builds) beyond basic CI.
- **Strict Scope Preservation**: If a dependency update requires substantial application logic rewrites beyond minor API signature adjustments, pause and consult the user or hand off to [`refactor`](../refactor/SKILL.md).
- **No Leftover Worktrees**: Always remove worktrees upon completion or cancellation to avoid disk clutter.
