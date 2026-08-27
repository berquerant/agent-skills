---
name: diff-continue
description: >-
  Use this skill when the user has a work-in-progress git diff and wants the
  agent to continue the same changes — inferring the pattern from what is
  already done and applying it to the remaining files or locations.
---

# Diff Continue

Use this skill to extend a work-in-progress change set.
Read the existing `git diff`, extract the intent and patterns, then
apply those patterns to the files or sections that haven't been touched yet.

---

## Step 1: Capture the Current Diff

Run the following to get the current uncommitted changes:

```bash
git diff          # unstaged changes
git diff --cached # staged changes
git diff HEAD     # all uncommitted changes (staged + unstaged)
```

If the user specifies a particular base (e.g., a branch or commit), use that:

```bash
git diff <base>
```

Read the full diff output carefully before proceeding.

Verify: You have the complete diff in context.

---

## Step 2: Understand the Pattern

Analyse the diff to extract:

1. **What is being changed** — Which constructs are being added, removed, or
   modified? (e.g., adding a new field to every struct, wrapping every function
   with error handling, inserting a log statement after each return)
2. **The transformation rule** — Can you describe the change as a rule?
   (e.g., "every handler gains a `ctx context.Context` first argument and
   passes it to downstream calls")
3. **Conventions** — Naming patterns, indentation, comment style, import
   order, etc. observed in the already-changed hunks.
4. **Scope** — Which files / packages / modules have been changed so far,
   and which have not?

Write a one-paragraph summary of the pattern before touching any files.

Verify: The pattern is specific enough to apply mechanically. If not, ask
the user to clarify.

---

## Step 3: Identify What Remains

Find the files and locations that still need the same change:

```bash
# List files containing a pattern that the diff modifies
git grep -l '<search-term>'

# Compare against already-changed files
git diff --name-only HEAD
```

Build a list of remaining targets: files, functions, structs, lines — wherever
the pattern should be applied but hasn't been yet.

Verify: The remaining-targets list is complete and correct.
Confirm with the user if the scope is large or unclear.

---

## Step 4: Apply the Pattern

Work through the remaining targets one file at a time (or one logical group
at a time for large changes).

For each target:
1. Read the existing code in context.
2. Apply the transformation following the **exact same conventions** observed
   in Step 2 — same naming, same indentation, same comment style.
3. Do not refactor or improve anything outside the scope of the pattern.
4. Do not change files that are already correctly modified.

After each file (or small batch), run a quick sanity check:

```bash
git diff          # review your own additions
```

Verify: Each change looks identical in style to the ones in the original diff.

---

## Step 5: Verify the Result

After all targets are done:

1. Run the project's tests or build:
   ```bash
   # Adapt to the project's toolchain
   go test ./...   # Go
   cargo test      # Rust
   pytest          # Python
   npm test        # Node
   ```
2. Review the full diff one more time:
   ```bash
   git diff HEAD
   ```
3. Confirm that every intended target has been covered and no unintended
   files were modified.

Verify: Build/tests pass. Diff is consistent end-to-end. No regressions.

---

## Guidelines

- **Pattern first, code second.** Always write down the rule (Step 2) before
  editing files. This prevents drift as you go through many files.
- **Preserve style exactly.** Micro-differences in indentation, spacing, or
  naming create noisy diffs and reviewer confusion.
- **Do not over-reach.** Only apply the change the diff demonstrates. If you
  spot other improvements, note them for the user but do not apply them now.
- **Ask before expanding scope.** If the remaining-targets list is ambiguous
  (e.g., "should this apply to test files too?"), ask the user first.
- **One file at a time for large changes.** Commit-sized batches are easier
  to review and easier to roll back.
