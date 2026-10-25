# Execution Safeguards & Environment Protocols

Mandatory safeguards and environment protocols to incorporate during the planning and execution phases of task workflows.
Per `output-presentation`, communicate all file paths, locations, and URLs as unadorned plain raw strings without backticks or Markdown links.

---

## 1. Git Commits, Remotes, and Change Requests (PR / MR)

When planning or executing modifications involving git history, remote repositories, or PRs/MRs:

- **Commit Author**: Determine the intended commit author name and email for commits created on the base/target branch. If not specified by the user, explicitly ask if referencing the global user git configuration is permitted. Never commit under an arbitrary, assumed, or unverified author.
- **Git SSH Configuration**: Check if remote operations require SSH authentication. Pre-confirm any specific SSH command, key, or parameter settings with the user.
- **PR / MR Assignee**: Pre-confirm who the assignee should be for any PR or MR to be opened.
- **Message Formatting**: Draft commit messages and PR/MR titles and descriptions following [`change-message`](../../change-message/SKILL.md) (Conventional Commits and English by default), presenting drafts to the user for approval.

---

## 2. Git Worktree and Branch Isolation

When modifying files in a git repository:

- **Isolated Worktree**: Always fetch the latest remote changes and create a new branch from the remote default branch in an isolated worktree before editing files.
- **Consult `git-worktree`**: Follow the procedures in [`git-worktree`](../../git-worktree/SKILL.md) for path resolution and lifecycle management.
- **Keep Primary Tree Clean**: Perform all modifications, builds, and test runs within the designated worktree directory.

---

## 3. File Archiving for Long Reports or Reusable Procedures

To avoid chat context pollution and token bloat:

- **Archiving Threshold**:
  - Investigation reports, technical findings, or architecture evaluations exceeding 300 lines.
  - Detailed procedure runbooks, execution checklists, or multi-step recovery guides intended for repeated reference.
- **Storage Protocol**:
  - Do not dump extensive documents directly into the chat buffer. Save them to a local file.
  - **Confirm File Path with User**: Propose a recommended path (e.g., in a designated logging location, project documentation directory, or task scratch area) and obtain explicit user confirmation before writing. Never write to arbitrary unconfirmed paths.

---

## 4. Command Execution & Snippet Protocol

When running shell commands during exploration or execution:

- **Threshold Guards**: Avoid executing long, chained, or complex inline commands that are difficult for users to inspect (≥4 chained operators, ≥120 characters, ≥3 lines, or complex inline interpreter payloads).
- **Convert to Snippets**: Follow [`command-snippet`](../../command-snippet/SKILL.md) to save command logic into dedicated, documented snippet files with explicit CLI arguments (avoiding environment variables for input).
- **Audit Intelligibility**: Ensure snippets are audited via [`snippet-review`](../../snippet-review/SKILL.md) and present the mandatory Pre-Execution Notice before execution.
