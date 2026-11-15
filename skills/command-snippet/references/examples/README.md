# Approvable Snippet Catalog & Reference Examples

This directory provides production-ready, approval-friendly reference implementations for frequently executed command sequences. Each example is packaged as a standalone specification document containing a copy-pasteable script, argument validation logic, safety assessment, and a ready-to-use **Pre-Execution Notice template**.

Agents should use these references directly instead of generating ad-hoc or poorly parameterized inline scripts when executing recurring workflows.

---

## Catalog by Operational Category

### 1. GitHub PR & Issue Lifecycle Operations
Constrain omnipotent `gh` CLI commands to explicit, read-only or targeted mutation scopes with stable execution prefixes.

- [gh-pr-list.md](gh-pr-list.md) (gh-pr-list.md): Fetches open pull requests with detailed metadata (title, branch, author, labels, files, body) as structured JSON in a single batch query.
- [gh-pr-merge.md](gh-pr-merge.md) (gh-pr-merge.md): Squash-merges or standard-merges an approved pull request and deletes the remote head branch under strict argument validation.
- [gh-pr-checks.md](gh-pr-checks.md) (gh-pr-checks.md): Monitors and polls status check rollups and CI checks across a batch of candidate pull requests without argument drift.
- [gh-main-ci.md](gh-main-ci.md) (gh-main-ci.md): Monitors GitHub Actions workflow runs on the default branch (`main`) for a specific commit or recent runs.

### 2. Git Worktree Isolation & Lifecycle Operations
Standardize git worktree operations under `.git-worktree/<branch-name>`, right-align variable parameters, and ensure clean cleanup.

- [git-worktree-setup.md](git-worktree-setup.md) (git-worktree-setup.md): Fetches a remote branch from origin and creates an isolated git worktree under `.git-worktree/<branch-name>`.
- [git-worktree-rebase.md](git-worktree-rebase.md) (git-worktree-rebase.md): Safely checks and rebases an isolated worktree branch onto `origin/main` atomically.
- [git-worktree-push.md](git-worktree-push.md) (git-worktree-push.md): Pushes a rebased branch from an isolated worktree to origin using `--force-with-lease`.
- [git-worktree-cleanup.md](git-worktree-cleanup.md) (git-worktree-cleanup.md): Safely removes a completed git worktree directory and deletes its local tracking branch.

---

## Design Principles Followed by These Examples

Every example in this catalog satisfies the core tenets of [approvable-patterns.md](../approvable-patterns.md) (../approvable-patterns.md):

1. **Trailing Variable Arguments**: Fixed action verbs and invariant flags are encapsulated inside the script; only dynamic arguments (repository directory, PR number, branch name) are exposed at the right end (`$1`, `$2`).
2. **Directory Encapsulation**: Navigates to target repositories or worktrees internally (`cd "${target_dir:?}"`), avoiding fragile or unportable flags like `gh -C` or nested subshells.
3. **Strict Parameter Validation**: Validates all arguments immediately (`${1:?Error: ...}`) and verifies types (e.g. numeric regex for PR numbers, directory existence).
4. **Standalone Binary Invocation**: Executable directly (`./script.sh [args]`), eliminating approval-bypassing interpreter prefixes (`bash script.sh`) and compound chaining (`chmod +x ... && ./...`).
5. **Standardized Pre-Execution Notices**: Formulates clear notices to avoid cognitive review fatigue during human approval.
