# AGENTS.md

Operational guidelines and conventions for AI agents operating within this repository.

## Repository Purpose

This repository hosts agent skills following the [Agent Skills Specification](https://agentskills.io/specification).
Each skill defines clear triggers, procedures, step-by-step verification methods, and guidelines.

## Development & Maintenance Rules

### 1. Skill Directory Conventions
- Every skill lives under `skills/<skill-name>/`.
- Must contain a valid `SKILL.md` with required frontmatter (`name`, `description`).
- Keep `SKILL.md` concise. Move extensive technical references or specifications to `references/` within the skill directory.
- `name` must be lowercase, alphanumeric with hyphens (e.g. `explore-plan-execute`).
- `description` must be written in the third person and clearly state **when to use** and **what it does**.

### 2. Standard SKILL.md Structure
All skills should adhere to the following section ordering:
1. **Frontmatter** (`name`, `description`)
2. **Title & Summary** (`# <Skill Title>`)
3. **Prerequisites** (if applicable)
4. **Step-by-Step Instructions** (`## Step 1: ...` with explicit `Verify:` checkpoints)
5. **Guidelines** (`## Guidelines` summarizing non-negotiable operational principles)

### 3. Universal & Platform-Agnostic Design Principles
Skills in this repository must be portable across different AI agent platforms (e.g., Antigravity, Claude Code, Cursor, Codex) and project structures.
- **Platform Agnostic**: Avoid hardcoding platform-specific tool names, built-in features, or proprietary commands unless providing them as non-exclusive examples.
- **Project Structure Agnostic**: Do not assume fixed directory layouts, toolchains, package managers, or tech stacks (e.g., provide language-neutral workflows or multi-language examples like Go, Rust, Python, Node).
- **Environment & Tool Adaptability**: Prefer standard shell commands (`git`, `grep`, `find`) and generic terminology (e.g., "run tests with the project's toolchain") over vendor-locked tooling.
- **Relative Path References**: Keep internal links and references relative within each skill directory (e.g., `[references/spec.md](references/spec.md)`).

### 4. File Modification & Git Protocol
- When adding or editing skills, write changes directly to `skills/<name>/`.
- `.agents/skills` is a symlink to `skills/` enabling project-local agent discovery. Do not replace it with a regular directory.
- Verify that markdown files are cleanly formatted and links use relative paths where appropriate.

## Skill Interoperability & Architecture

Skills in this repository follow a **loosely coupled, sequential hand-off architecture**. Skills operate independently and do not hard-depend on or invoke each other as mandatory subroutines. Instead, they produce clear outputs and recommend the next appropriate skill upon user approval.

```mermaid
flowchart TD
    subgraph Meta ["Meta & Skill Creation"]
        SC[skill-creator]
        MC[mcp-creator]
    end

    subgraph Planning ["Planning & Workflow"]
        EPE[explore-plan-execute]
        SPE[scrutinize-plan-execute]
        SG[step-gate]
        GW[git-worktree]
    end

    subgraph Audit ["Audit & Verification - Read-Only"]
        PA["project-audit<br/>(Macro health, security, public release, tech debt)"]
        PS["project-status<br/>(Micro working-tree & drift check)"]
        CR["code-review<br/>(Meso diff/PR review & security audit)"]
    end

    subgraph Remediation ["Remediation & Action"]
        RF[refactor]
        DC[diff-continue]
    end

    %% Workflows & Hand-offs
    EPE -->|sequential hand-off| SG
    SPE -->|sequential hand-off| SG
    GW -.->|worktree isolation| EPE
    GW -.->|worktree isolation| SPE

    PA -->|suggests remediation| RF
    PA -->|suggests phased fixes| SG

    CR -.->|shares security checklist| PA
    RF -.->|consults review criteria| CR

    PS -.->|diff drift verification| DC
```

### Interoperability Principles
- **Read-Only vs. Mutation Separation**: Audit skills (`project-audit`, `project-status`, `code-review`) are strictly read-only and never mutate files. When changes are required, they suggest mutation skills (`refactor`, `step-gate`, `diff-continue`).
- **Sequential Hand-off (User-in-the-Loop)**: Rather than calling other skills directly in a deep nested chain, skills present findings and recommend the next skill for the user to approve.
- **Reference Sharing over Duplication**: When audit standards overlap (such as security and public release checklists), skills share references (e.g. `../code-review/references/security-checklist.md`) via relative links instead of duplicating checklist content.


