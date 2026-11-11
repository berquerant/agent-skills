---
name: change-message
description: >-
  Use this skill when the user or another workflow needs to write a git commit
  message or a pull request / merge request title and description. Analyzes the
  diff and intent, follows repository conventions (defaulting to English and
  Conventional Commits), and produces concise, review-friendly messages.
---

# Change Message

Write commit messages and PR/MR descriptions that let a reviewer grasp the
what, why, and risk of a change at a glance.

---

## Step 1: Collect the Change and Its Intent

1. Identify the change scope:
   - Commit: staged changes (`git diff --staged`, `git diff --staged --stat`).
   - PR/MR: the branch against its target (`git diff <target>...HEAD`, `git log --oneline <target>..HEAD`).
2. Identify the intent: the user's request, linked issues, or the plan being executed.
   If the motivation cannot be determined from context, ask the user instead of guessing.
3. For commits, check the staged scope:
   - If nothing is staged, inspect `git diff HEAD` and ask the user what to stage. Do not stage files on your own.
   - If the staged diff mixes unrelated changes, propose splitting it into separate commits before drafting.

Verify: The exact diff range and the motivation for the change are known, and each commit to be drafted covers a single logical change.

## Step 2: Detect Repository Conventions

Repository conventions take precedence over the defaults in this skill. Check, if present:

- `CONTRIBUTING.md`, `AGENTS.md`, or similar contributor docs
- Commit linters (e.g. `commitlint.config.*`, `.commitlintrc*`, `.gitmessage`)
- PR/MR templates (`.github/pull_request_template.md`, `.github/PULL_REQUEST_TEMPLATE/`, `.gitlab/merge_request_templates/`)
- Recent history style: `git log --format='%s' -n 30`

Decide:
- **Language**: match the repository's existing messages; default to English.
- **Format**: Conventional Commits by default; follow the repository's style if it clearly differs.
- **Description layout**: use the repository template if one exists; otherwise use the default template.

Verify: Language, commit format, and description template are decided, with the source of each decision noted.

## Step 3: Draft the Message

- Commit message: follow [commit-message.md](references/commit-message.md) (references/commit-message.md) and [conventional-commits-spec.md](references/conventional-commits-spec.md) (references/conventional-commits-spec.md).
- PR/MR title and description: follow [pr-description.md](references/pr-description.md) (references/pr-description.md).

Verify: A draft exists for every requested artifact (commit message, PR/MR title, description).

## Step 4: Self-Check

Check the draft against every item below and fix any failures:

- [ ] The subject/title alone tells what changed (no "fix bug", "update", "WIP").
- [ ] The *why* is stated; the *what* is summarized, not narrated file by file.
- [ ] Every claim matches the actual diff; nothing is speculative or invented.
- [ ] Breaking changes, migrations, and risks are explicitly called out.
- [ ] Verification steps reflect commands actually run (or are marked as not run).
- [ ] No secrets, credentials, internal hostnames, or personal data are included (consult [security-checklist.md](../code-review/references/security-checklist.md) (../code-review/references/security-checklist.md)).
- [ ] Language and format match the decisions in Step 2.

Verify: All checklist items pass.

## Step 5: Present for Approval

Present the final message(s) in a code block, together with the conventions
applied (language, format, template source).

- **Standalone use**: ask the user to approve or edit the message. After approval,
  recommend a workflow skill for committing, pushing, or opening the PR/MR
  (e.g. [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md), or [multi-repo-change](../multi-repo-change/SKILL.md) (../multi-repo-change/SKILL.md) for multi-repository work).
- **Used within another workflow** (e.g. `multi-repo-change`, `explore-plan-execute`):
  return the draft to that workflow and present it at its own approval gate.
  Do not request a separate approval here.

Do not commit, push, or open a PR/MR as part of this skill unless the user explicitly asks.

Verify: The user has approved or edited the message, either directly (standalone) or at the calling workflow's approval gate.

## Guidelines

- **Follow repository conventions first**: Inspect recent git history and templates first, falling back to English and Conventional Commits when the repository has no established pattern.
- **Lead with a self-contained subject line**: Summarize the core outcome in the first line so reviewers immediately grasp the intent.
- **Focus the body on the "Why"**: Explain the motivation, trade-offs, and context behind the change rather than narrating the diff file by file.
- **Group changes by logical meaning**: Summarize related edits into concise bullets to keep PR descriptions scannable.
- **Ground every claim in verified diff evidence**: Cite only actual changes, issue numbers, and commands that were genuinely run.
- **Keep commits atomic**: Recommend splitting mixed, unrelated changes into separate commits before finalizing the message.
