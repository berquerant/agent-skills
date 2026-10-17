# agent-skills

A curated collection of agent skills designed for AI coding assistants (such as Antigravity, Claude Code, and other skill-compatible agents).

## Overview

This repository provides reusable skills that enforce best practices, structured workflows, safe refactoring, thorough code reviews, and project scaffolding.

All skills are maintained under `skills/`.

## Included Skills

| Skill | Description | Key Features |
|---|---|---|
| [`explore-plan-execute`](skills/explore-plan-execute/SKILL.md) | Enforces structured workflows on non-trivial tasks. | Read before writing, user approval gate, verification loops. |
| [`checkpoint-log`](skills/checkpoint-log/SKILL.md) | Records progressive milestone logs and rationale to a user-confirmed local file. | No default paths, append-only step entries, context recovery, decision tracking. |
| [`code-review`](skills/code-review/SKILL.md) | Inspects code and documentation quality, security, and public release safety. | Default branch diff check, public exposure hazard alerts, severity categorization, read-only audit. |
| [`refactor`](skills/refactor/SKILL.md) | Safely transforms code and docs without changing behavior. | Baseline tests, pragmatic DRY, code-review integration, mandatory lint/test passing, doc synchronization. |
| [`doc-refactor`](skills/doc-refactor/SKILL.md) | Refines documentation and code comments with zero code mutation. | Information preservation, clarity & logical depth, zero code diff verification, Why-first explanations. |
| [`diff-continue`](skills/diff-continue/SKILL.md) | Continues work-in-progress git diffs across other files. | Pattern extraction, style preservation, scope detection. |
| [`git-worktree`](skills/git-worktree/SKILL.md) | Manages git worktrees for parallel branch work. | WORKTREE_ROOT organization, subagent parallelism, cleanup on request. |
| [`project-status`](skills/project-status/SKILL.md) | Audits the current state of a project. | Git state inspection, file change analysis, test/lint result reporting, pass/fail judgments. |
| [`project-audit`](skills/project-audit/SKILL.md) | Comprehensive health check, security, public release readiness, and maintainability audit. | Read-only whole-project diagnostics, exposure risk ratings, dynamic commit window, maintainability metrics, hand-off to refactor/step-gate. |
| [`doctor`](skills/doctor/SKILL.md) | One-shot project checkup across docs/comments, maintainability, latent bugs, and security, ending in a prioritized improvement plan. | Mode selection (project/recent × quick/thorough), evidence-backed findings, reuses code-review/project-audit/evidentiary-inquiry criteria, approval gate, hand-off to refactor/doc-refactor/step-gate. |
| [`skill-creator`](skills/skill-creator/SKILL.md) | Scaffolds and writes new agent skills. | Frontmatter quality checklist, progressive disclosure, standard directory templates. |
| [`skill-review`](skills/skill-review/SKILL.md) | Reviews agent skills against the Agent Skills Specification and repository conventions. | Read-only, automated shell checks, six-dimension verdict rules, batch cross-skill summary, hand-off to step-gate/refactor. |
| [`mcp-creator`](skills/mcp-creator/SKILL.md) | Scaffolds and implements Model Context Protocol (MCP) servers. | Python/TS templates, tool/resource/prompt support, host setup guides. |
| [`scrutinize-plan-execute`](skills/scrutinize-plan-execute/SKILL.md) | Enforces critical obstacle review, user consultation, and archiving checks. | Self-critical risk scrutiny, remedy consultation, iterative refinement, no-default file archiving. |
| [`step-gate`](skills/step-gate/SKILL.md) | Executes tasks incrementally with explicit user checkpoints. | Step decomposition, verification criteria, user approval gates between steps. |
| [`multi-repo-change`](skills/multi-repo-change/SKILL.md) | Coordinates multi-repository changes with approval gates and linked PRs/MRs. | Isolated worktrees, pre-flight clean checks, approval gates, draft PR/MR cross-linking. |
| [`change-message`](skills/change-message/SKILL.md) | Writes concise commit messages and PR/MR descriptions. | Repository convention detection, Conventional Commits default, summary-first PR template, self-check list. |
| [`review-discussion`](skills/review-discussion/SKILL.md) | Guides PR/MR review comments, discussions, and thread resolutions. | Conventional Comments, constructive ethos, actionable replies, resolution ownership rules. |
| [`edit-paced-execution`](skills/edit-paced-execution/SKILL.md) | Paces execution intervals between file edits with dynamic pauses. | Adaptive thresholds, non-edit tool call tracking, obstacle reporting, user check-in. |
| [`evidentiary-inquiry`](skills/evidentiary-inquiry/SKILL.md) | Conducts evidence-backed investigations with source scoping and fact verification. | Pre-exploration source scoping, pre-report fact checks, separation of facts and inferences, deductive explanations. |
| [`resolve-dependency-prs`](skills/resolve-dependency-prs/SKILL.md) | Triages, audits, verifies, remediates, and merges dependency and security PRs/MRs. | Manifest filter, cost-aware batching, security/legitimacy audit, safety halt on doubts, idle-wait multi-repo scheduling. |
| [`resolve-recent-dependency-prs`](skills/resolve-recent-dependency-prs/SKILL.md) | Coordinates discovery, mode selection, and resolution of dependency/security PRs/MRs across repositories. | Explicit/recent repository discovery, pre-flight duration confirmation, batch and interactive triage modes with defer/skip, sequential hand-off, non-blocking CI switching. |

## Project Structure

```text
.
├── .github/
│   ├── dependabot.yml           # Dependabot configuration
│   └── workflows/
│       └── lint.yml             # CI lint workflow
├── bin/
│   ├── install.sh               # Symlinks skills to a target directory
│   └── lint.sh                  # Shellcheck and shfmt runner
├── skills/
│   ├── change-message/
│   │   └── references/          # Commit message & PR/MR description guides
│   ├── checkpoint-log/
│   │   └── references/          # Milestone log format & checkpoint templates
│   ├── code-review/
│   │   └── references/          # Security checklists & OWASP secure coding guide
│   ├── diff-continue/
│   ├── doc-refactor/
│   │   └── references/          # Information inventory & zero-code-diff guide
│   ├── doctor/
│   │   └── references/          # Mode axes, diagnostic axes routing & report template
│   ├── edit-paced-execution/
│   │   └── references/          # Pacing thresholds & checkpoint report guide
│   ├── evidentiary-inquiry/
│   │   └── references/          # Source scoping templates & verification checklist
│   ├── explore-plan-execute/
│   │   └── references/          # Command cache protocol guide
│   ├── git-worktree/
│   │   └── references/          # Git worktree workflow references
│   ├── mcp-creator/
│   │   ├── examples/            # Example server implementations
│   │   └── references/          # Detailed MCP protocol specs
│   ├── multi-repo-change/
│   │   └── references/          # Approval gates, platform tools, safeguards
│   ├── project-audit/
│   │   └── references/          # Maintainability & debt metrics
│   ├── project-status/
│   ├── refactor/
│   ├── resolve-dependency-prs/
│   │   └── references/          # Manifest filtering, security audit, CI & scheduling guides
│   ├── resolve-recent-dependency-prs/
│   ├── review-discussion/
│   │   └── references/          # Comment conventions & resolution rules
│   ├── scrutinize-plan-execute/
│   │   └── references/          # Plan scrutiny & obstacle checklists
│   ├── skill-creator/
│   │   └── references/          # Detailed Agent Skill specs
│   ├── skill-review/
│   │   ├── references/          # Review checklist, automated checks, report templates
│   │   └── scripts/             # Automated validation scripts
│   └── step-gate/
├── AGENTS.md                    # Operational guidelines & context for AI agents
├── .gitignore
├── .mise.toml                   # Tool configuration and mise tasks
└── README.md
```

## Setup & Usage

### Installing Skills into a Target Directory

Use `mise run install` or `bin/install.sh` to create symbolic links pointing to each skill in your target directory:

```bash
# Interactive prompt if target directory is not specified
mise run install

# Or specify TARGET_DIR directly
TARGET_DIR=/path/to/skills mise run install

# Alternatively, invoke the script directly
./bin/install.sh /path/to/skills
```

The installer displays all planned actions and asks for confirmation (`[Y/n]`) before creating symlinks.
Note that `ln -sfn` is used:
- Existing symlinks with the same name in `TARGET_DIR` are overwritten without extra prompts.
- If an actual directory with the same name exists in `TARGET_DIR`, the symlink is created inside that directory.

### Project-local usage
To enable project-local skill auto-discovery in AI agent environments expecting skills under `.agents/skills`, create a symlink:

```bash
mkdir -p .agents
ln -s ../skills .agents/skills
```

AI agents scanning `.agents/skills` will then discover and activate these skills automatically.

## Development

This repository uses [mise](https://mise.jdx.dev/) to manage development tools (`shellcheck`, `shfmt`).

```bash
# Install development tools
mise install

# Run linter checks (shellcheck & shfmt)
mise run lint

# Format shell scripts
mise run fmt
```
