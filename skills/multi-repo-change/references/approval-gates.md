# Approval Gates and Reporting Templates

The multi-repo change workflow relies on four explicit user approval checkpoints to prevent unintended mutations. Use the standardized templates below.

---

## Gate 1: Plan Approval

Present this report before creating any worktrees, branches, or modifying files.

```markdown
## Multi-Repository Change Plan

### Candidate Repositories

| Repository | Local Path | Remote Origin | Target Branch | Status |
|---|---|---|---|---|
| `<repo-1>` | `<path>` | `<remote-url>` | `<branch>` | Target |
| `<repo-2>` | `<path>` | `<remote-url>` | `<branch>` | Skipped: `<reason (e.g. uncommitted changes)>` |

### Proposed Branch Name
`<branch-name>`

### Change Overview
<High-level summary of the cross-cutting modifications and rationale>

### Repository Plans

#### `<repo-1>`
- **Target Files**: `<file-1>`, `<file-2>`
- **Changes**: `<bullet points of modifications>`
- **Repository Notes**: `<repo-specific conventions, template paths, or dependencies>`
- **Validation**: `<lint command>`, `<test command>`

#### `<repo-2>`
...

### Representative Diff Preview (`<representative-repo>`)
```diff
<diff excerpt or simulated diff>
```

### Inter-Repository Dependencies
- Can be performed in parallel: Yes / No
- Required implementation order: `<order or None>`
- Required merge order: `<order or None>`

### Verification Commands Matrix
| Repository | Lint / Format | Unit / Integration Test | Build / Codegen |
|---|---|---|---|
| `<repo>` | `<command>` | `<command>` | `<command or N/A>` |

### Potential Risks & Confirmations
- `<risk or open question>`

### Actions Upon Approval
1. Create isolated worktree & branch from `origin/<default-branch>` for each target repo.
2. Apply changes and execute verification commands.
3. Present diff stats and commit review for Gate 2.
```

---

## Gate 2: Commit Review & Approval

Present this review after applying changes and executing local verification in each isolated worktree.

```markdown
## Pre-Commit Review

### `<repo-1>`
- **Changed Files**:
  - `<file-1>`
  - `<file-2>`
- **Diff Stat**:
  ```text
  <git diff --stat HEAD>
  ```
- **Validation Results**:
  - Lint: Passed / Failed / Skipped
  - Test: Passed / Failed / Skipped
- **Proposed Commit Message**:
  ```text
  <commit message (subject and body)>
  ```

### `<repo-2>`
...

### Request
Approve committing changes with the messages above across target repositories?
```

---

## Gate 3: Push Review & Approval

Present this review after commits are created locally.

```markdown
## Pre-Push Review

### `<repo-1>`
- **Branch**: `<branch-name>`
- **Target Remote**: `origin/<default-branch>`
- **Commit**: `<sha> <subject>`
- **Cumulative Diff Stat**:
  ```text
  <git diff --stat origin/<default-branch>...HEAD>
  ```

### `<repo-2>`
...

### Guarantees
- Pushes to new branch `<branch-name>` only.
- No direct push to `<default-branch>`.
- Force push (`--force`) is NOT used.
- Creation of Pull / Merge Requests will require separate Gate 4 approval.

### Request
Approve pushing the branches to remotes?
```

---

## Gate 4: Change Request (PR / MR) Approval

Present this review after remote branches have been pushed successfully.

```markdown
## Change Request (PR / MR) Creation Plan

### `<repo-1>`
- **Platform**: `<GitHub | GitLab | Bitbucket | Gitea>`
- **Source Branch**: `<branch-name>`
- **Target Branch**: `<default-branch>`
- **Draft / WIP**: Yes
- **Title**: `<title>`
- **Reviewers**: `<reviewers or none>`
- **Labels**: `<labels or none>`
- **Description Preview**:
  ```markdown
  <description content including checklist and related PR placeholder>
  ```

### `<repo-2>`
...

### Sequence Upon Approval
1. Open Draft PR / MR on each hosting platform using available platform tools.
2. Update descriptions across all created PRs/MRs with cross-links.
3. No automatic merging will be configured.

### Request
Approve opening these Change Requests?
```

---

## Final Completion Summary Template

```markdown
## Multi-Repository Execution Summary

### Created Change Requests
| Repository | Branch | Commit | Change Request (PR/MR) |
|---|---|---|---|
| `<repo-1>` | `<branch>` | `<sha>` | `<PR/MR URL>` |
| `<repo-2>` | `<branch>` | `<sha>` | `<PR/MR URL>` |

### Verification Status
| Repository | Lint / Format | Test Suite | Additional Checks |
|---|---|---|---|
| `<repo-1>` | Passed | Passed | `<result>` |

### Skipped Repositories
| Repository | Reason |
|---|---|
| `<repo-skipped>` | `<uncommitted changes / not applicable>` |

### Active Worktrees (Retained)
| Repository | Worktree Path |
|---|---|
| `<repo-1>` | `<worktree-path>` |
| `<repo-2>` | `<worktree-path>` |

> [!NOTE]
> All Change Requests are opened as Draft / WIP.
> Worktrees and local branches have been retained for review. Would you like to clean up the worktrees now?
```
