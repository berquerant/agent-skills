# Safeguards and Operational Constraints

These safeguards must be adhered to at all stages during multi-repository operations.

---

## 1. Uncommitted Changes Protection

Any repository in a dirty or unsettled state must be **excluded immediately** from the operation to protect ongoing user work.

Check the status of each candidate repository:

```bash
git -C <repo-path> status --porcelain=v1 --untracked-files=all
```

If the output is **not empty**, or if any of the following conditions are present, do **NOT** modify the repository:
- Staged or unstaged changes
- Untracked files
- Merge / rebase / cherry-pick conflicts or in-progress operations

### Forbidden Actions on Dirty Repositories
Under no circumstances should the agent attempt any of the following without explicit, individual user direction:
- `git stash`
- `git reset` (soft, mixed, or hard)
- `git clean`
- Discarding changes via checkout or restore
- Creating branches or worktrees on top of uncommitted changes

Record the skip reason and repository status, and report it clearly in the planning and completion summaries.

---

## 2. Forbidden Operations

Unless the user explicitly grants targeted approval with specific parameters, the agent must **never** execute:

- Direct pushes to the default or protected branch (e.g., `main`, `master`)
- Force pushes (`git push --force`, `git push --force-with-lease`)
- Destructive git operations (`git reset --hard`, `git clean -fdx`)
- Git stashing (`git stash`)
- Unsupervised rebases or amending existing public commits
- Deleting existing branches or tags
- Automatic merging of Pull Requests / Merge Requests
- Disabling git hooks or bypassing checks (`--no-verify`)
- Modifying repository permissions, visibility, or protected branch settings

---

## 3. Secret and Credential Protection

Never inspect, display, modify, or commit:
- Personal Access Tokens (PATs) or API keys
- Private keys (SSH, PGP, SSL/TLS)
- OAuth tokens or bearer credentials
- Credential files or keystores
- Production secrets or `.env` files containing sensitive values

If any sensitive information is detected or suspected during edits or diff reviews, immediately halt work and alert the user.

---

## 4. Definition of Explicit User Approval

Automated environment prompts (e.g., shell sandbox permission dialogues, MCP server execution confirmations) do **not** constitute user approval within this workflow.

Approval gates require:
1. Presenting the exact planned actions, diff stats, or payload in chat.
2. Receiving explicit user confirmation (e.g., "approve plan", "proceed with commit", "push changes", "create PRs").
3. In case of ambiguous responses, asking for clarification before taking mutation actions.

---

## 5. Token & Context Window Management

When dealing with many repositories (typically 4 or more) or large diffs:
- **Summarize Diffs**: Avoid printing full diffs for every file. Provide `git diff --stat`, concise bullet-point descriptions of changes, and targeted diff snippets for safety-critical changes (auth, security, CI, migrations).
- **Compress Validation Logs**: Output pass/fail status and only relevant error excerpts instead of dumping verbose stdout/stderr test logs.
- **Batched Reviews**: If the review payload remains excessively large, present the reports in logical batches (e.g., 3–4 repositories per response) and request approval incrementally.
