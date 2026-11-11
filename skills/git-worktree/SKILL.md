---
name: git-worktree
description: >-
  Use this skill when the user wants to start work on a new branch using git
  worktree. Creates a worktree directory under WORKTREE_ROOT (organized by
  origin URL and branch name), optionally manages parallel worktrees via
  subagents, and cleans up completed worktrees on request.
---

# Git Worktree Workflow

Use this skill to create a git worktree for a new branch, work inside it,
and remove it when the work is complete. See [git-worktree-commands.md](references/git-worktree-commands.md) (references/git-worktree-commands.md)
for the complete command and URL parsing reference.

---

## Prerequisites

- `git` is installed and the current directory is inside a git repository.
- The user has provided (or can confirm) the **WORKTREE_ROOT** — the base
  directory where all worktrees are created.

---

## Step 1: Resolve WORKTREE_ROOT

If the user has not specified WORKTREE_ROOT, ask:

> "Which directory should be used as the root for git worktrees (WORKTREE_ROOT)?"

Once you have a value, verify read/write access:

```sh
ls <WORKTREE_ROOT>
```

If access is denied, inform the user and request permission before continuing (following [user-notification](../user-notification/SKILL.md) (../user-notification/SKILL.md) for elevated permission prompts).
Access to paths outside the sandbox or restricted directories may require elevated user permissions or running outside sandbox isolation.

Verify: WORKTREE_ROOT exists and is writable.

---

## Step 2: Determine the Worktree Path

1. Get the origin remote URL:

   ```sh
   git remote get-url origin
   ```

2. Extract `<owner>/<repo>` from the URL:
   - SSH format: `git@github.com:owner/repo.git` → `owner/repo`
   - HTTPS format: `https://github.com/owner/repo.git` → `owner/repo`

3. Get the branch name from the user or the current context.

4. Compose the worktree path:

   ```
   <WORKTREE_ROOT>/<owner>/<repo>/<branch-name>
   ```

   Example: `~/worktrees/berquerant/agent-skills/feat-my-feature`

Verify: The path components are correct and the branch name is valid (no spaces, no special characters).

---

## Step 3: Create the Worktree

1. Create the parent directory if it does not exist:

   ```sh
   mkdir -p <WORKTREE_ROOT>/<owner>/<repo>
   ```

2. Add the worktree with a new branch:

   ```sh
   git worktree add <path> -b <branch-name>
   ```

   If the branch already exists, use `-B` to reset it, or omit `-b` to check
   out the existing branch. Confirm with the user before overwriting an
   existing branch.

3. Verify the worktree was created:

   ```sh
   git worktree list
   ```

Verify: The new worktree appears in `git worktree list` with the correct path and branch.

---

## Step 4: Work Inside the Worktree

Perform all file edits and commands within the worktree directory:

```
<WORKTREE_ROOT>/<owner>/<repo>/<branch-name>/
```

Treat this directory as the project root for the duration of the task, and proceed with the primary task workflow (such as [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md), [scrutinize-plan-execute](../scrutinize-plan-execute/SKILL.md) (../scrutinize-plan-execute/SKILL.md), or [multi-repo-change](../multi-repo-change/SKILL.md) (../multi-repo-change/SKILL.md)).

### Parallel Work with Subagents (optional)

If subagent functionality is available **and** the user requests work on
multiple branches simultaneously:

1. Repeat Steps 2–3 for each branch to create separate worktree directories.
2. Spawn one subagent per worktree, passing each subagent its dedicated
   worktree path and the specific task to perform.
3. Coordinate results and merge or review as directed by the user.

Verify: Each subagent is working in its own worktree directory without conflicts.

---

## Step 5: Remove the Worktree When Done

Once the user confirms that the work in a worktree is complete:

1. Remove the worktree:

   ```sh
   git worktree remove <path>
   ```

   If there are uncommitted changes, `--force` is required. Confirm with the
   user before using `--force`.

2. Optionally delete the branch if it has been merged:

   ```sh
   git branch -d <branch-name>
   ```

3. Verify no stale worktree entries remain:

   ```sh
   git worktree list
   git worktree prune
   ```

Verify: The removed worktree no longer appears in `git worktree list`.

---

## Guidelines

- **Confirm `WORKTREE_ROOT` explicitly with the user**: Ask for the base directory whenever it is not specified so worktrees are created in a predictable, user-approved location.
- **Verify before resetting or force-removing**: Request explicit user confirmation before passing `-B` or `--force` so existing branches and uncommitted changes are protected.
- **Retain worktrees until user-confirmed completion**: Remove a worktree only after the user confirms that branch work is complete.
- **Follow the preceding notification protocol for outside-sandbox paths**: When `WORKTREE_ROOT` lies outside the sandbox, notify the user per [user-notification](../user-notification/SKILL.md) (../user-notification/SKILL.md) and request permission before accessing it.
- **Use parallel worktrees only when subagents are supported**: Propose multi-worktree parallel execution only in environments with confirmed subagent capabilities.
