---
name: resolve-dependency-prs
description: >-
  Use this skill to triage, update, verify, and merge dependency update or
  security advisory Pull Requests / Merge Requests on GitHub or GitLab. Filters
  for pure manifest bumps, checks out via git-worktree, rebases onto default,
  verifies via local/remote CI, plans remediation gates upon failure, and merges.
---

# Resolve Dependency PRs

Triage open dependency update and security patch PRs/MRs, verify them in isolated worktrees, ensure they are up to date with the default branch, validate through CI, remediate failures under user approval, and safely merge.

This workflow coordinates existing platform conventions, [`git-worktree`](../git-worktree/SKILL.md) for branch isolation, and [`step-gate`](../step-gate/SKILL.md) for remediation approval gates without duplicating procedures.

---

## Prerequisites

- `git` is installed and the current directory is inside a git repository.
- A supported platform tool is authenticated:
  - GitHub: `gh` CLI (or `github` MCP server) as described in [`multi-repo-change/references/platform-tools.md`](../multi-repo-change/references/platform-tools.md).
  - GitLab: `glab` CLI (or `gitlab` MCP server) as described in [`multi-repo-change/references/platform-tools.md`](../multi-repo-change/references/platform-tools.md).

---

## Step 1: List and Filter Candidate PRs/MRs

1. Detect the remote platform host via `git remote get-url origin`.
2. Retrieve open PRs/MRs:
   - **GitHub**:
     ```bash
     gh pr list --state open --json number,title,headRefName,author,labels
     ```
   - **GitLab**:
     ```bash
     glab mr list --state opened
     ```
3. Apply metadata and strict file-level filtering according to [references/filter-criteria.md](references/filter-criteria.md):
   - Check author, labels, and title for dependency/security patterns.
   - Inspect changed files for each candidate PR (`gh pr diff <num> --name-only` or `glab mr diff <id> --raw`).
   - **Exclude immediately** any PR modifying application code, test logic, or general workflows. Only retain changes restricted to package manager manifest and lockfiles.

Verify: The list of candidate PR/MR numbers and branches is filtered to pure dependency/security updates.

---

## Step 2: Check Out Branch Using git-worktree

For each candidate PR/MR, isolate the worktree following [`git-worktree`](../git-worktree/SKILL.md):

1. Fetch the remote branch and latest default branch:
   ```bash
   git fetch origin <default-branch> <head-branch>
   ```
2. Create an isolated worktree under `WORKTREE_ROOT`:
   - Follow Step 1 and Step 2 of [`git-worktree`](../git-worktree/SKILL.md).
   - Checkout the PR's head branch inside the new worktree.

Verify: The worktree is successfully created, clean, and checked out to the target PR branch.

---

## Step 3: Rebase onto Default Branch and Sync

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
   - Wait for the remote CI checks to trigger.

Verify: The branch contains the latest default branch commits and is pushed cleanly.

---

## Step 4: CI Verification (Remote and Local)

Follow [references/ci-detection.md](references/ci-detection.md) to dynamically inspect project test/lint commands without assuming fixed toolchains:

1. **Local Verification (if feasible)**:
   - Detect project test commands from `.github/workflows/`, `.gitlab-ci.yml`, `Makefile`, `mise.toml`, `package.json`, etc.
   - If the required runtime is available locally, run the fast verification commands inside the worktree.
2. **Remote Verification**:
   - Check or watch remote CI status:
     - **GitHub**: `gh pr checks <number> --watch`
     - **GitLab**: `glab mr view <id>` (inspect pipeline status)

Verify: Verification results from both local execution (if applicable) and remote CI are collected and evaluated.

---

## Step 5: Remediation Plan and Approval Gate

If all CI checks pass, proceed directly to **Step 6**.

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
   - Push to remote and return to **Step 4** to re-verify.
   - If the required refactor is extensive or changes behavioral contracts, propose handing off to [`refactor`](../refactor/SKILL.md) and defer merging.

Verify: The user has approved the remediation plan before any file modifications are applied, and tests pass after changes.

---

## Step 6: Merge and Clean Up

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

- **Never Force Without Lease**: Always use `--force-with-lease` when updating rebased remote branches.
- **Strict Scope Preservation**: If a dependency update requires substantial application logic rewrites beyond minor API signature adjustments, pause and consult the user or hand off to [`refactor`](../refactor/SKILL.md).
- **No Leftover Worktrees**: Always remove worktrees upon completion or cancellation to avoid disk clutter.
