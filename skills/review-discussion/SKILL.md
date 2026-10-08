---
name: review-discussion
description: >-
  Use this skill when drafting, reviewing, or resolving comments, notes, and
  discussion threads on Pull Requests or Merge Requests. Enforces Conventional
  Comments, constructive tone, explicit action responses, and clear resolution
  ownership without posting without approval.
---

# Review Discussion

Guide the creation, discussion, and resolution of review comments and threads
on PRs and MRs with constructive norms, Conventional Comments conventions,
and clear resolution ownership.

---

## Step 1: Identify Context and Intent

Determine the role and context:
1. **Role**:
   - **Reviewer**: Writing new inline comments or review summary notes.
   - **Author**: Responding to review feedback, explaining decisions, or proposing resolutions.
2. **Review Target / Thread**:
   - Target PR/MR URL, diff file, and line numbers.
   - For existing threads: read existing comments, identifying blocking vs. non-blocking concerns.

Verify: The role (reviewer or author) and target context or thread history are clearly identified.

---

## Step 2: Detect Repository Conventions

Check for existing review guidelines or comment conventions in:
- `CONTRIBUTING.md`, `AGENTS.md`, or `.github/` / `.gitlab/` guidelines
- PR description / review checklist requirements

Decide:
- **Conventions**: Use Conventional Comments (`suggestion:`, `issue:`, `question:`, etc.) by default; align with repo conventions if specified.
- **Language**: Match the repository's primary language; default to English.

Verify: Comment conventions and language preferences are established.

---

## Step 3: Draft Comment or Response

- **When Reviewing (Drafting Comments)**:
  - Follow [references/comment-conventions.md](references/comment-conventions.md).
  - Use appropriate labels (`issue`, `suggestion`, `question`, `nitpick`, `praise`).
  - Mark decorations (`(blocking)` or `(non-blocking)`) when ambiguity exists.
  - State the rationale (why) and provide concrete replacement code or patch where practical.
- **When Responding (Replying to Threads)**:
  - Follow [references/resolution-rules.md](references/resolution-rules.md).
  - Clearly state what action was taken (referencing commit SHA or change details).
  - If deferring to a follow-up issue or declining, provide technical justification and tracking link.

Verify: The drafted comment or reply has a clear label/intent, rationale, and actionable suggestion or response.

---

## Step 4: Evaluate Resolution Readiness

If proposing to resolve a thread:
- Verify resolution criteria against [references/resolution-rules.md](references/resolution-rules.md).
- Check ownership: ensure blocking issues are resolved only with reviewer agreement or verification.
- Ensure external tasks are tracked with issue links if deferred.

Verify: Resolution criteria are verified before marking any thread as resolved.

---

## Step 5: Present for Approval

Present the drafted comments or replies to the user in a formatted block.
- For new comments: show file, line reference, label, and full markdown preview.
- For thread replies: show thread context, proposed response, and resolution status change (e.g. "Mark as Resolved: Yes/No").

Do not post comments or resolve threads via API/CLI tools without explicit user approval.

Verify: The user has reviewed and approved the drafted comment, response, or resolution action.

---

## Guidelines

- **Critique the code, not the person.** Keep discussions objective, constructive, and empathetic.
- **Explain the rationale.** Never request changes without stating the underlying reason or risk.
- **Explicit blocking status.** Prevent stalled PRs by distinguishing blocking defects from optional suggestions.
- **Concrete over abstract.** Provide executable replacement diffs or code snippets when suggesting changes.
- **Respect resolution ownership.** Do not resolve reviewer-blocking issues unilaterally without confirmation or clear code fix verification.
