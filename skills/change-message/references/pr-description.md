# PR / MR Description Guide

Use the repository's PR/MR template if one exists, filling each section with the
principles below. Otherwise use the default template.

## Title

- Same rules as a commit subject (Conventional Commits by default).
- For a single-commit PR/MR, reuse the commit subject.
- For multiple commits, summarize the overall outcome, not the last commit.

## Default Template

```markdown
## Summary

<1–2 sentences: what this changes and the outcome. A reviewer should be able to stop here.>

## Background

<Why this is needed. Link issues or related PRs/MRs if they exist.>

## Changes

- <Change grouped by meaning, not by file>
- <3–5 bullets maximum>

## Impact & Risks

<Compatibility, performance, security, operational impact, migrations. Write "None" if none.>

## Verification

- <Commands actually run and their result, e.g. `make test` passed>
- <Steps a reviewer can follow to reproduce>

## Review Focus (optional)

<Areas where reviewer attention is most valuable.>
```

## Principles

- **Summary first.** The first section must stand alone.
- **Concise.** Prefer bullets over prose; omit empty optional sections.
- **Accurate.** Every statement must match the diff; mark unrun verification as "not run".
- **Explicit risk.** Never hide breaking changes or migrations in the middle of the text.
- **Cross-links.** For multi-repository changes, add a `## Related PRs/MRs` section listing each counterpart.
