# Git Worktree Command Reference

Quick reference for `git worktree` commands used in the `git-worktree` skill.

---

## Create a Worktree

```sh
# New branch
git worktree add <path> -b <new-branch>

# Existing branch (check out without creating)
git worktree add <path> <existing-branch>

# Reset an existing branch to current HEAD
git worktree add <path> -B <branch>
```

## List Worktrees

```sh
git worktree list
# Porcelain output (machine-readable):
git worktree list --porcelain
```

## Remove a Worktree

```sh
# Safe remove (fails if dirty)
git worktree remove <path>

# Force remove (discards uncommitted changes)
git worktree remove --force <path>
```

## Prune Stale Entries

```sh
# Removes administrative files for worktrees whose directories no longer exist
git worktree prune
```

## Move a Worktree

```sh
git worktree move <current-path> <new-path>
```

---

## URL Parsing Examples

| Origin URL | Extracted `<owner>/<repo>` |
|------------|---------------------------|
| `git@github.com:owner/repo.git` | `owner/repo` |
| `https://github.com/owner/repo.git` | `owner/repo` |
| `git@gitlab.com:group/project.git` | `group/project` |
| `https://gitlab.com/group/project` | `group/project` |

### Shell snippet to extract owner/repo

```sh
git remote get-url origin | sed -E 's#^(git@[^:]+:|https?://[^/]+/)##; s#\.git$##'
```

---

## Worktree Path Convention

```text
<WORKTREE_ROOT>/<owner>/<repo>/<branch-name>   # Shared root
.git-worktree/<branch-name>                   # Project-local alternative
```

Examples:

```text
~/worktrees/owner/repo/feat-my-feature
~/worktrees/myorg/backend/fix-login-bug
.git-worktree/chore-update-deps
```
