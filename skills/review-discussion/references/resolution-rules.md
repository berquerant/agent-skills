# Thread Resolution & Discussion Rules

This reference outlines clear criteria and workflows for resolving review comment threads and concluding discussions on Pull Requests and Merge Requests.

---

## 1. Resolution Ownership (Who Resolves?)

- **Default rule**: The **author of the comment (reviewer)** should resolve the thread once they verify the resolution.
- **Author resolution (when permitted)**:
  - Trivial nitpicks or minor suggestions that the author applied verbatim.
  - Informational questions where the author provided a clear, factual answer and the reviewer acknowledged it.
  - Automated bot comments (e.g. linter / CI notes) that have been satisfied.
- **Never auto-resolve blocking issues unilaterally**: If a reviewer raised an `issue (blocking)` or substantial design objection, the PR author must not resolve the thread without explicit agreement or confirmation from the reviewer.

---

## 2. When to Resolve a Thread

A thread should only be marked resolved when one of the following conditions is met:

1. **Fixed in Code**:
   - The requested change or fix has been committed and pushed.
   - Reply to the thread with the commit reference or a short note:
     > Fixed in `abc1234` by adding the null-check.
2. **Clarified and Agreed**:
   - A question or thought was answered, and the reviewer acknowledges the explanation:
     > Thanks, that makes sense given the backward compatibility constraint.
3. **Deferred to Follow-up Issue**:
   - Both parties agree that the comment is valid but out of scope for the current PR.
   - The author opens a tracking ticket / issue, links it in the thread, and then resolves:
     > Agreed. Logged follow-up issue #456 to address this refactor. Resolving here.
4. **Intentionally Declined with Rationale**:
   - After discussion, the suggestion is declined with mutually understood trade-offs:
     > Decided to keep the current interface to maintain parity with the v1 API.

---

## 3. How to Respond to Comments

When replying as the PR/MR author:

- **State the action clearly**: Don't just reply "done". State *how* it was addressed or refer to the commit.
- **Ask for confirmation when unsure**: If proposing an alternative fix, ask the reviewer before making large changes:
  > Would approach B work better here to avoid changing the public signature?
- **Separate discussions**: If a thread branches into two distinct topics, open a new thread or follow-up issue rather than letting the original discussion sprawl.

---

## 4. Handling Stalemates and Disagreements

When the author and reviewer cannot reach agreement on a comment:

1. **Step back to shared objectives**: Restate user impact, system invariants, or project goals.
2. **Involve a third party**: Tag the technical lead or another domain owner for input.
3. **Document the decision**: Once resolved (or escalated), record the consensus in the PR description or thread summary before merging.
