# Agent Skills Specification Reference (Condensed)

> [!NOTE] Local Reference Snapshot
> This document is a local, self-contained summary of the core principles from the Agent Skills Specification.
> Originally derived from: `https://agentskills.io/specification` (for human citation only).
> **AI agents must strictly rely on this local specification and MUST NOT fetch or access external URLs autonomously.**

---

## 1. Directory Structure

Every agent skill resides in its own self-contained directory named after the skill:

```text
skills/<skill-name>/
├── SKILL.md          # Required: Entrypoint and procedural workflow
├── references/       # Optional: In-depth domain manuals, checklists, schemas
├── scripts/          # Optional: Executable automation scripts
├── examples/         # Optional: Concrete code/pattern examples
└── resources/        # Optional: Reusable assets, templates, configurations
```

- **Single source of execution**: `SKILL.md` is the primary orchestrator loaded by the agent.
- **Progressive disclosure**: Detailed specifications and long checklists belong in `references/` and are read on demand.

---

## 2. SKILL.md Frontmatter

`SKILL.md` MUST start with a YAML frontmatter block containing `name` and `description`:

```markdown
---
name: my-skill
description: >-
  Use this skill when ...
---
```

### Constraints

- **`name`**:
  - Length: 1–64 characters (recommended roughly 3–30 characters).
  - Format: Lowercase letters (`a-z`), digits (`0-9`), and single hyphens (`-`).
  - Restrictions: Must NOT begin or end with a hyphen, and consecutive hyphens (`--`) are forbidden.
  - Match: Must match the parent directory name exactly.
- **`description`**:
  - Length: 1–1024 characters (recommended roughly 100–350 characters).
  - Voice: Third person (e.g. starts with `Use this skill when ...`).
  - Content: Must clearly define both the **activation triggers** (when to run) and the **outcomes/effects** (what it does).
  - Avoid vague language ("helps with various things").

---

## 3. Structural Conventions

A standard `SKILL.md` adheres to the following structure:
1. **Frontmatter**: `name` and `description`.
2. **Title & Summary**: `# <Skill Title>` followed by 1–2 sentences summarizing the core responsibility.
3. **Prerequisites** (optional): Required tools, shell commands, or permissions.
4. **Step-by-Step Instructions**: Sequential steps (`## Step 1: ...`, `## Step 2: ...`).
   - Every procedural step MUST conclude with a measurable verification checkpoint (`Verify: ...`).
5. **Guidelines**: `## Guidelines` summarizing non-negotiable operational principles.

---

## 4. Self-Containment and Portability

- **Relative linking**: Internal cross-references must use relative file paths to real files:
  ```text
  [link text](references/spec.md)
  ```
