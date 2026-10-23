# Intent Matrix & Dispatch Routing Reference

This reference provides the intent classification criteria, recommended entry skills, and sequential hand-off pipelines for agent tasks.

---

## 1. Primary Intent Classification Matrix

| Category | Typical User Prompts / Signals | Recommended Entry Skill | Typical Follow-up Pipeline |
|---|---|---|---|
| **Feature Work & Non-Trivial Implementation** | "Add a new feature...", "Implement XYZ", "Create new component" | `explore-plan-execute` | `step-gate` (if complex) → `change-message` |
| **High-Risk / Ambiguous Tasks** | "Redesign architecture", "Major migration", unclear trade-offs | `scrutinize-plan-execute` | `step-gate` → `checkpoint-log` → `change-message` |
| **Stepwise Critical Execution** | "Execute this plan step by step", destructive operations, DB changes | `step-gate` | `checkpoint-log` → `project-status` → `change-message` |
| **Holistic Health Check & Triage** | "Check the project", "Diagnose repository", "Find issues" | `doctor` | `refactor` / `doc-refactor` / `step-gate` |
| **Code Review & Security Audit** | "Review this PR", "Review recent commits", "Security check" | `code-review` | `review-discussion` / `refactor` / `doc-refactor` |
| **Deep Project & Debt Audit** | "Analyze maintainability", "Audit tech debt & exposure" | `project-audit` | `refactor` / `step-gate` |
| **Working Tree & Diff Drift** | "Check current git status", "Did tests pass?", "Any uncommitted changes?" | `project-status` | `diff-continue` / `refactor` |
| **Behavior-Preserving Refactoring** | "Clean up this module without changing behavior", "DRY this code" | `refactor` | `code-review` → `change-message` |
| **Documentation & Comment Clarification** | "Improve docstrings", "Clarify comments without touching code" | `doc-refactor` | `project-status` → `change-message` |
| **Continuing Partial Diffs** | "Apply this WIP pattern to the rest of the codebase" | `diff-continue` | `refactor` → `change-message` |
| **Parallel Branch / Isolated Work** | "Work in a separate branch/worktree", "Isolate changes" | `git-worktree` | `explore-plan-execute` / `multi-repo-change` |
| **Multi-Repository Coordination** | "Update across multiple repos", "Coordinate linked PRs" | `multi-repo-change` | `git-worktree` → `change-message` |
| **Dependency & Security PR Resolution** | "Merge dependabot PRs", "Update vulnerable dependencies" | `resolve-dependency-prs` | `project-status` → `change-message` |
| **Cross-Repo Dependency Discovery** | "Check open dependency PRs across all repositories" | `resolve-recent-dependency-prs` | `resolve-dependency-prs` |
| **Commit & PR Message Crafting** | "Write a commit message", "Describe this PR" | `change-message` | — |
| **PR Comment Discussion & Resolution** | "Reply to review comments", "Resolve PR discussion" | `review-discussion` | `refactor` / `change-message` |
| **Evidence-Backed Investigation** | "Investigate root cause with evidence", "Verify facts" | `evidentiary-inquiry` | `explore-plan-execute` / `refactor` |
| **Creating New Skills** | "Create a new agent skill", "Add a skill for XYZ" | `skill-creator` | `skill-review` |
| **Auditing Skills** | "Review this skill against specs", "Check skills consistency" | `skill-review` | `step-gate` / `refactor` |
| **Creating MCP Servers** | "Build an MCP server", "Scaffold MCP tools" | `mcp-creator` | `explore-plan-execute` |
| **Command Scripting & Snippet Parameterization** | "Save command as script", "Format complex command into snippet", "Parameterize command" | `command-snippet` | `step-gate` / `explore-plan-execute` |
| **Output String & Copy-Paste Presentation** | "Format output as raw string", "Make paths/URLs copy-pasteable without markdown" | `output-presentation` | — |

---

## 2. Dispatch Decision Flow

```text
User Request / Query
       │
       ▼
Is the request clear and targeted at a single skill?
       ├─ YES ──> Hand off directly to target skill
       └─ NO  ──> Analyze Context:
                   ├─ Workspace state: git status, diffs, open PRs
                   ├─ Task complexity: multi-step, high-risk, multi-repo
                   └─ Match against Intent Matrix above
                           │
                           ▼
                  Formulate Pipeline Proposal
                  (Entry Skill -> Hand-off Skills)
                           │
                           ▼
                  Present to User & Await Approval
                           │
                           ▼
                  Hand Off to Initial Skill
```

---

## 3. Pipeline Synthesis Patterns

### Pattern A: Diagnostic & Remediation
1. Entry: `doctor` (or `project-audit` / `code-review`)
2. Remediation Gate:
   - For code cleanup: `refactor`
   - For doc/comment polish: `doc-refactor`
   - For architecture/bug fixes: `explore-plan-execute` + `step-gate`
3. Finalization: `project-status` → `change-message`

### Pattern B: High-Risk Feature / Bug Fix
1. Entry: `scrutinize-plan-execute` (or `explore-plan-execute`)
2. Isolation: `git-worktree`
3. Incremental Execution: `step-gate` (with optional `checkpoint-log` and `edit-paced-execution`)
4. Finalization: `code-review` → `change-message`

### Pattern C: Dependency Upgrades
1. Discovery (multi-repo): `resolve-recent-dependency-prs`
2. Execution (single-repo): `resolve-dependency-prs`
3. Complex fix hand-off: `refactor` (if API breaking changes occur)
4. Finalization: `project-status` → `change-message`
