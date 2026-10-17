---
name: resolve-recent-dependency-prs
description: >-
  Use this skill when the user wants to audit, verify, and resolve open dependency
  update or security patch PRs/MRs across specified or recently active repositories.
  Discovers candidate repositories, guides multi-repository mode selection (batch
  or interactive), and delegates to resolve-dependency-prs under user approval
  without idle CI waiting.
---

# Resolve Recent Dependency PRs

Coordinate discovery, prioritization, and resolution of open dependency update and security advisory Pull Requests / Merge Requests (PRs/MRs) across specified or recently active repositories.

This skill orchestrates repository discovery (via explicit user input or recent platform activity search), multi-repository triage modes (`batch` vs `interactive`), and delegates repository-level triage, worktree isolation, CI validation, and safe merging to [`resolve-dependency-prs`](../resolve-dependency-prs/SKILL.md) upon user confirmation without idle waiting.

---

## Prerequisites

- `git` is installed and target repositories are accessible locally (or clonable).
- A supported remote platform CLI tool is authenticated:
  - GitHub: `gh` CLI (`gh auth status`).
  - GitLab: `glab` CLI (`glab auth status`).
- Sibling skill [`resolve-dependency-prs`](../resolve-dependency-prs/SKILL.md) is available in the environment to perform per-repository PR/MR audits and merges.

---

## Step 1: Determine Target Discovery Mode (Explicit vs. Recent Search)

Check whether the user provided an explicit list of target repositories (local directories, URLs, or `owner/repo` names):

### 1.1 Path A: Explicit Repository List (User-Specified)
If the user explicitly provided repository paths, URLs, or names:
1. **Bypass time window checks**: Do NOT prompt for or filter by recent update dates.
2. Target all specified repositories directly.
3. Every open PR/MR in each specified repository will be audited in subsequent steps.

### 1.2 Path B: Automated Discovery by Recent Activity (Time Window)
If no explicit repositories were provided:
1. **Pre-flight Confirmation**: Ask the user whether to proceed with the default search duration (past 7 days / 1 week) or adjust the time window (e.g., 3 days, 14 days, or a specific date):
   ```text
   The default scan window is the past 7 days. Would you like to proceed with this duration or specify a different time window?
   ```
2. Calculate the start date (`SINCE_DATE`) based on the confirmed duration:
   ```bash
   # Calculate date in YYYY-MM-DD format (cross-platform fallback)
   SINCE_DATE=$(python3 -c "from datetime import date, timedelta; print(date.today() - timedelta(days=7))" 2>/dev/null || date -d '7 days ago' +%Y-%m-%d 2>/dev/null || date -v-7d +%Y-%m-%d)
   ```
3. Query the platform CLI for non-archived repositories updated since `SINCE_DATE`:
   - **GitHub**:
     ```bash
     gh search repos --owner @me --archived=false --updated ">=${SINCE_DATE}" --json fullName,url > active_repos.json
     ```
   - **GitLab**:
     ```bash
     glab api "/projects?membership=true&archived=false&last_activity_after=${SINCE_DATE}T00:00:00Z" --paginate > active_repos.json
     ```

Verify: The target repository list is established either from user-specified input or from confirmed platform search results.

---

## Step 2: Resolve Local Working Copies and Inspect Open PRs/MRs

For each candidate repository:

1. **Resolve Local Working Directory**:
   - Locate the local checkout directory (e.g. searching existing workspace paths, repository roots, or standard checkouts).
   - If the repository does not exist locally, clone it using standard `git clone` or the platform tool:
     - **GitHub**: `gh repo clone <owner/repo> <local-path>` or `git clone <url> <local-path>`
     - **GitLab**: `glab repo clone <owner/repo> <local-path>` or `git clone <url> <local-path>`
2. **Collect Open Candidate PRs/MRs**:
   Query open Pull Requests / Merge Requests for each repository:
   - **GitHub**:
     ```bash
     gh pr list --repo <owner/repo> --state open --json number,title,headRefName,author,labels > repo_prs.json
     ```
   - **GitLab**:
     ```bash
     glab mr list --repo <owner/repo> --state opened > repo_mrs.txt
     ```
3. Filter out repositories with zero open PRs/MRs. Retain only repositories that have at least one open candidate PR/MR.

Verify: Repositories with active candidate PRs/MRs are identified, and their local working directory paths are verified.

---

## Step 3: Multi-Repository Execution Mode Selection

If only a single repository has candidate PRs/MRs, proceed directly to **Step 4**.

If multiple repositories contain candidate PRs/MRs, present the candidate summary and prompt the user to choose an execution mode:

```markdown
### Discovered Repositories with Candidate PRs/MRs
| Repository | Local Path | Open PR/MR Count |
|---|---|---|
| `owner/repo-a` | `/path/to/repo-a` | 2 |
| `owner/repo-b` | `/path/to/repo-b` | 3 |

Please select an execution mode:
1. `batch` (Batch Planning Mode): Triage all repositories together, formulate a consolidated resolution plan, and execute after single user approval.
2. `interactive` (Interactive Triage Mode): Process one repository at a time, prompting for action (Start, Hold, or Skip) per repository.
```

### Execution Mode Characteristics:
- **`batch` (Batch Planning Mode)**:
  - Explores all candidate repositories upfront.
  - Constructs a consolidated cross-repository plan.
  - Requests a single user approval, then proceeds sequentially (interleaving idle CI waits).
- **`interactive` (Interactive Triage Mode)**:
  - Takes one repository at a time and formulates an individual triage proposal.
  - Asks the user for one of three decisions:
    - **`start` (Start)**: Execute the resolution plan for this repository immediately. Upon completion, advance to the next repository.
    - **`defer` (Hold)**: Defer this repository for later. Advance to the next repository immediately; remind the user of all held repositories after the remaining repositories are processed.
    - **`skip` (Skip)**: Skip this repository completely, discard it from memory, and advance to the next repository.

Verify: The user has selected either `batch` or `interactive` mode (or single-repository execution is established).

---

## Step 4: Delegate to resolve-dependency-prs with Idle-Wait Switching

For each repository to be processed:

1. **Obtain Explicit User Approval**:
   Following the Sequential Hand-off principle, present the repository's triage plan and request explicit user confirmation before initiating file modifications or branch checkouts.
2. **Execute via `resolve-dependency-prs`**:
   Upon user approval, switch working directory to the repository and carry out the [`resolve-dependency-prs`](../resolve-dependency-prs/SKILL.md) workflow:
   - Filter changes to package manifests and lockfiles.
   - Run context, legitimacy, and security audits.
   - Create isolated worktrees via `git-worktree`.
   - Rebase onto default branch and run local tests/linters.
   - Validate remote CI checks and merge safely.
3. **Idle-Wait Switching (Non-Blocking Progress)**:
   - If remote CI pipelines are running, do NOT block in active sleep or polling loops.
   - Record the current repository's status as `WAITING_CI`.
   - **Immediately switch focus** to the next pending repository (or a previously held repository) to perform discovery, local audit, or testing.
   - Re-check pending CI status only when transitioning between repository checkpoints.

Verify: Each repository is processed through [`resolve-dependency-prs`](../resolve-dependency-prs/SKILL.md) under user approval, and idle CI waits are interleaved productively.

---

## Step 5: Remind Held Repositories and Report Summary

1. **Remind Held Repositories (Interactive Mode)**:
   If any repositories were marked as `defer` (Hold) in `interactive` mode, present the held repository list to the user:
   ```markdown
   ### Deferred Repositories Reminder
   The following repositories were held during the initial pass:
   - `owner/repo-b` (3 open PRs)

   Would you like to process these deferred repositories now?
   ```
   If approved, process them following **Step 4**.
2. **Consolidated Summary Report**:
   Present an end-to-end outcome report covering all target repositories:
   - **Merged PRs/MRs**: Repository, PR/MR number, package name, version delta.
   - **Skipped / Halted PRs/MRs**: Any PRs excluded due to non-manifest changes, breaking changes, or user request.
   - **Pending Status**: Any items still awaiting external reviews or lengthy CI pipelines.
   - **Cleanup**: Confirm all temporary worktrees were cleanly removed.

Verify: All held repositories have been addressed or explicitly acknowledged, and a comprehensive summary report is delivered.

---

## Guidelines

- **Sequential Hand-Off over Hard Coupling**: Do not invoke sibling skills as non-interactive automated subroutines. Always present findings, formulate a concrete plan, and obtain explicit user approval before delegating execution to [`resolve-dependency-prs`](../resolve-dependency-prs/SKILL.md).
- **Respect User-Specified Scope**: When explicit repository paths or URLs are provided, strictly honor the input and bypass time-window restrictions. Audit all open PRs/MRs for those repositories.
- **Pre-Flight Duration Confirmation**: When discovering repositories automatically via platform activity, always confirm or adjust the scan window with the user before querying.
- **Platform-Agnostic Workflows**: Keep core instructions portable across platforms (GitHub, GitLab) and avoid tying workflows to proprietary tools. Provide platform commands (`gh`, `glab`) as non-exclusive examples.
- **Never Idle-Wait on Remote CI**: When a repository is waiting for remote CI completion or approvals, switch to another candidate repository immediately to ensure continuous forward progress.
- **Clear Decision Semantics in Interactive Mode**: Adhere strictly to the three triage states (`start` executes and advances; `defer` postpones and reminds at the end; `skip` drops the repository completely).
