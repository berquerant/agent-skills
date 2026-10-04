# agent-skills

A curated collection of agent skills designed for AI coding assistants (such as Antigravity, Claude Code, and other skill-compatible agents).

## Overview

This repository provides reusable skills that enforce best practices, structured workflows, safe refactoring, thorough code reviews, and project scaffolding.

All skills are maintained under `skills/` and symlinked via `.agents/skills` for seamless workspace discovery.

## Included Skills

| Skill | Description | Key Features |
|---|---|---|
| [`explore-plan-execute`](skills/explore-plan-execute/SKILL.md) | Enforces structured workflows on non-trivial tasks. | Read before writing, user approval gate, verification loops. |
| [`code-review`](skills/code-review/SKILL.md) | Inspects code and documentation quality, security, and public release safety. | Default branch diff check, public exposure hazard alerts, severity categorization, read-only audit. |
| [`refactor`](skills/refactor/SKILL.md) | Safely transforms code and docs without changing behavior. | Baseline tests, pragmatic DRY, code-review integration, mandatory lint/test passing, doc synchronization. |
| [`diff-continue`](skills/diff-continue/SKILL.md) | Continues work-in-progress git diffs across other files. | Pattern extraction, style preservation, scope detection. |
| [`git-worktree`](skills/git-worktree/SKILL.md) | Manages git worktrees for parallel branch work. | WORKTREE_ROOT organization, subagent parallelism, cleanup on request. |
| [`project-status`](skills/project-status/SKILL.md) | Audits the current state of a project. | Git state inspection, file change analysis, test/lint result reporting, pass/fail judgments. |
| [`project-audit`](skills/project-audit/SKILL.md) | Comprehensive health check, security, public release readiness, and maintainability audit. | Read-only whole-project diagnostics, exposure risk ratings, dynamic commit window, maintainability metrics, hand-off to refactor/step-gate. |
| [`skill-creator`](skills/skill-creator/SKILL.md) | Scaffolds and writes new agent skills. | Frontmatter quality checklist, progressive disclosure, standard directory templates. |
| [`mcp-creator`](skills/mcp-creator/SKILL.md) | Scaffolds and implements Model Context Protocol (MCP) servers. | Python/TS templates, tool/resource/prompt support, host setup guides. |
| [`scrutinize-plan-execute`](skills/scrutinize-plan-execute/SKILL.md) | Enforces critical obstacle review, user consultation, and archiving checks. | Self-critical risk scrutiny, remedy consultation, iterative refinement, no-default file archiving. |
| [`step-gate`](skills/step-gate/SKILL.md) | Executes tasks incrementally with explicit user checkpoints. | Step decomposition, verification criteria, user approval gates between steps. |

## Project Structure

```text
.
├── .agents/
│   └── skills -> ../skills/     # Symlink for agent auto-discovery
├── skills/
│   ├── code-review/
│   │   └── references/          # Security & public exposure checklists
│   ├── diff-continue/
│   ├── explore-plan-execute/
│   ├── git-worktree/
│   │   └── references/          # Git worktree workflow references
│   ├── mcp-creator/
│   │   └── references/          # Detailed MCP protocol specs
│   ├── project-audit/
│   │   └── references/          # Maintainability & debt metrics
│   ├── project-status/
│   ├── refactor/
│   ├── scrutinize-plan-execute/
│   │   └── references/          # Plan scrutiny & obstacle checklists
│   ├── skill-creator/
│   │   └── references/          # Detailed Agent Skill specs
│   └── step-gate/
├── AGENTS.md                    # Operational guidelines & context for AI agents
└── README.md
```

## Setup & Usage

### Project-local usage
Clone or submodule this repository into your workspace, and ensure `.agents/skills` points to `skills/`:

```bash
ln -s skills .agents/skills
```

AI agents that scan `.agents/skills` will automatically discover and activate these skills when triggered.
