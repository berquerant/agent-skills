---
name: multi-repo-change
description: >-
  Use this skill when applying a coordinated change across multiple local git
  repositories. Inspects repositories, creates isolated worktrees from remote default
  branches, validates changes, commits and pushes through explicit approval gates,
  and opens linked Draft Pull Requests / Merge Requests using platform tools.
---

# Multi-Repo Change

Apply coordinated changes across multiple git repositories, creating matching worktrees and branches, and opening cross-linked Draft Pull Requests / Merge Requests on GitHub or GitLab.

To prevent unintended modifications, destructive git actions, or unauthorized pushes, this skill strictly enforces sequential execution through explicit approval gates.

---

## Prerequisites

- `git` is installed and the target repositories are available locally.
- Platform tools for GitHub or GitLab are configured if opening Change Requests automatically:
  - GitHub: `github` MCP server or `gh` CLI (see [references/platform-tools.md](references/platform-tools.md)).
  - GitLab: `gitlab` MCP server or `glab` CLI (see [references/platform-tools.md](references/platform-tools.md)).
- Operational safeguards and forbidden actions defined in [references/safeguards.md](references/safeguards.md) must be strictly observed.

---

## Step 1: Inspect Candidate Repositories

1. Identify candidate repository paths (from user input, `ghq list -p`, or project discovery).
2. For each repository, check the working tree status:
   ```bash
   git -C <repo-path> status --porcelain=v1 --untracked-files=all
   ```
   **If the output is not empty, exclude the repository immediately.** Do not run `git stash`, `reset`, or `clean`.
3. Fetch remote updates and resolve the remote default branch:
   ```bash
   git -C <repo-path> fetch --prune origin
   git -C <repo-path> symbolic-ref --short refs/remotes/origin/HEAD
   ```
4. Verify that the planned branch name does not already exist locally or on remote:
   ```bash
   git -C <repo-path> show-ref --verify --quiet refs/heads/<branch-name>
   git -C <repo-path> ls-remote --exit-code --heads origin <branch-name>
   ```

Verify: Every candidate repository has a clean working tree and a verified remote default branch without branch name collisions.

---

## Step 2: Formulate the Multi-Repo Plan

Synthesize the inspection findings and build a comprehensive execution plan covering:
- Target repositories and skipped repositories (with exact reasons).
- Proposed branch name and rationale for the changes.
- Per-repository files to modify, repo-specific conventions, and planned verification commands (formatters, linters, tests).
- Representative diff preview.
- Dependencies, execution order, and merge order.

Format the proposal following the **Gate 1: Plan Approval** template in [references/approval-gates.md](references/approval-gates.md).

> [!IMPORTANT]
> **Approval Gate 1**: Stop and present the plan to the user. Do **NOT** create worktrees, branches, or edit files until the user explicitly approves the plan.

Verify: Explicit user approval of the plan is received before proceeding.

---

## Step 3: Set Up Isolated Worktrees and Apply Changes

1. For each target repository, create an isolated worktree under a designated base directory (e.g. `${XDG_CACHE_HOME:-$HOME/.cache}/multi-repo-change/worktrees/<host>/<repo>/<branch-slug>`). For general worktree lifecycle practices and options, consult the sibling [`git-worktree`](../git-worktree/SKILL.md) skill:
   ```bash
   git -C <repo-path> worktree add \
     <worktree-path> \
     -b <branch-name> \
     origin/<default-branch>
   ```
2. Apply the approved changes within each worktree. Follow repository-specific conventions (e.g., code style, formatting, lockfile updates).
3. Execute the planned verification commands (lint, unit tests, type checks) within each worktree.
4. If tests fail or unexpected modifications are needed, resolve within the planned scope or halt and consult the user.

Verify: All planned changes are applied and local verification commands pass in each worktree.

---

## Step 4: Commit Review and Stage

1. Inspect changes and diff stat in each worktree:
   ```bash
   git -C <worktree-path> status --short
   git -C <worktree-path> diff HEAD
   ```
2. Prepare commit messages matching repository conventions.
3. Present the **Gate 2: Commit Review** report as defined in [references/approval-gates.md](references/approval-gates.md).

> [!IMPORTANT]
> **Approval Gate 2**: Obtain explicit user approval before staging or committing files.

4. Upon approval, stage only planned files explicitly (do NOT use `git add .` or `git add -A`):
   ```bash
   git -C <worktree-path> add -- <file-1> <file-2>
   git -C <worktree-path> commit -m "<approved-message>"
   ```

Verify: Clean working tree in each worktree and commit recorded on `<branch-name>`.

---

## Step 5: Push Review and Remote Publishing

1. Check the cumulative diff against the remote default branch:
   ```bash
   git -C <worktree-path> diff --stat origin/<default-branch>...HEAD
   ```
2. Present the **Gate 3: Push Review** report as defined in [references/approval-gates.md](references/approval-gates.md).

> [!IMPORTANT]
> **Approval Gate 3**: Obtain explicit user approval before pushing branches to remote.

3. Upon approval, push each branch safely without force-pushing:
   ```bash
   git -C <worktree-path> push -u origin <branch-name>
   ```

Verify: Remote branches pushed successfully and tracking upstream is established.

---

## Step 6: Create Change Requests (PR / MR) and Cross-Link

1. Determine the hosting platform (GitHub or GitLab) and available integration tools (MCP server or official CLI) as described in [references/platform-tools.md](references/platform-tools.md).
2. Check for existing Change Requests on the branch to avoid duplicates.
3. Prepare titles, descriptions, and labels according to repository templates.
4. Present the **Gate 4: Change Request Approval** report as defined in [references/approval-gates.md](references/approval-gates.md).

> [!IMPORTANT]
> **Approval Gate 4**: Obtain explicit user approval before opening PRs/MRs.

5. Upon approval, open each PR / MR in Draft / WIP mode:
   - GitHub: using `github` MCP tools or `gh pr create --draft ...`
   - GitLab: using `gitlab` MCP tools or `glab mr create ...` (with `Draft:` prefix)
6. Once all PRs/MRs are opened, update their descriptions to cross-link all related PR/MR URLs (see [references/platform-tools.md](references/platform-tools.md)).

Verify: All Draft PRs/MRs are open and bidirectional cross-links are present in each description.

---

## Step 7: Completion Summary and Worktree Lifecycle

1. Present the final report using the **Completion Summary Template** in [references/approval-gates.md](references/approval-gates.md).
2. Retain worktrees and local branches by default. Ask the user if they wish to clean up the worktrees:
   ```bash
   git -C <repo-path> worktree remove <worktree-path>
   ```

Verify: The user receives the full summary table with PR/MR URLs and worktree retention status.

---

## Guidelines

- **Never bypass approval gates**: Human-in-the-loop checkpoints at Plan, Commit, Push, and PR/MR creation are non-negotiable.
- **Protect dirty working trees**: Any uncommitted changes in candidate repositories immediately disqualify them from automated modification.
- **Isolate work in worktrees**: Keep existing user workspaces untouched by creating separate worktrees for cross-repo tasks.
- **Draft mode by default**: Always open Change Requests as Draft / WIP without configuring automerge or direct merge.
- **Cross-link coordination**: Always connect related multi-repo PRs/MRs to give reviewers end-to-end visibility.
