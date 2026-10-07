# Commit Message Guide

Default format: [Conventional Commits](https://www.conventionalcommits.org/).
Use the repository's own convention instead if one is detected.

## Structure

```text
<type>(<optional scope>)<!>: <subject>

<body: why the change is needed, then a short summary of what changed>

<footer: BREAKING CHANGE, issue references, co-authors>
```

## Subject Line

- `type`: `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `style`, `revert`.
- `scope`: optional; a module, package, or skill name (e.g. `feat(change-message): ...`).
- Imperative mood, lowercase after the colon, no trailing period.
- Keep within ~72 characters (50 preferred).
- Describe the outcome, not the activity: `fix(auth): reject expired refresh tokens`, not `fix: fix bug`.
- Append `!` after the type/scope for breaking changes.

## Body (omit for trivial changes)

- Wrap at ~72 characters.
- First paragraph: **why** — the problem, motivation, or context.
- Then: **what** — up to ~3 bullets summarizing the change by meaning.
- Mention notable trade-offs or alternatives rejected, only if they help future readers.
- Do not restate details obvious from the diff.

## Footer

- `BREAKING CHANGE: <what breaks and how to migrate>` — mandatory for breaking changes.
- Issue references: `Refs: #123`, `Closes #123` (only real, confirmed references).

## Example

```text
feat(change-message): add skill for commit and PR/MR messages

Commit messages and PR descriptions varied in structure across skills,
making changes hard to review at a glance.

- Add a shared workflow that detects repository conventions first
- Default to English and Conventional Commits
- Reference the skill from skills that create commits or PRs/MRs
```
