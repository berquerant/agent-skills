# Skill Specification Reference

This reference is used by the skill-creator skill when generating new skills.

---

## Directory Structure

```text
skills/<skill_name>/
├── SKILL.md          # Required: Main instruction file with frontmatter
├── scripts/          # Optional: Helper scripts and utilities
├── examples/         # Optional: Reference implementations and usage patterns
├── resources/        # Optional: Additional files, templates, or assets
└── references/       # Optional: Detailed documentation or manuals
```

---

## SKILL.md Frontmatter

```markdown
---
name: my-specialized-skill
description: >-
  Describe when the agent should use this skill. Use third-person.
  Example: "Use this skill when the user asks to run integration tests for the XYZ service."
---
```

### Field Definitions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ | 1–64 characters. Lowercase letters, numbers, and single hyphens only. Must not start or end with a hyphen, and consecutive hyphens (`--`) are forbidden. **Must match the parent directory name exactly.** Recommended 3–30 characters. |
| `description` | string | ✅ | 1–1024 characters. Non-empty. Explains when the agent should activate the skill and what it does. Uses third-person conventions (recommended roughly 100–350 characters). |

---

## Discovery and Placement Locations

Always confirm the placement location with the user. Never assume a default.

### Workspace (Project-Specific)

Standard discovery paths vary by agent platform. Common conventions include:

```
<workspace_root>/.agents/skills/<name>/
<workspace_root>/_agents/skills/<name>/
<workspace_root>/.agent/skills/<name>/
<workspace_root>/_agent/skills/<name>/
```

> [!TIP]
> Alternatively, maintain skills in `<workspace_root>/skills/<name>/` and symlink
> `<workspace_root>/.agents/skills` -> `skills` for direct workspace visibility.

Committing to VCS (Git) allows sharing with the team.

### Global (Machine-Wide)

Place the skill inside the agent's global configuration directory.
The exact path varies by agent tool and environment — always confirm with the user or check the agent platform's documentation.

Applies across all projects.

---

## Progressive Disclosure

- Only the skill's **name and description** are injected into context by default
- Full content is loaded only when explicitly activated by the agent or user
- Keep `SKILL.md` concise; move heavy documentation to `references/`

---

## Best Practices

1. **Progressive Disclosure**
   Keep `SKILL.md` brief. Move heavy documentation to `references/` and link to it from `SKILL.md`.
   The agent reads those files only when needed.

2. **Executable Helpers**
   Encapsulate complex command sequences in scripts under `scripts/`.
   Use relative links so the agent can easily find and run them.

3. **Validation Steps**
   Include a way to verify success for each step
   (e.g., checking a log file, running a dry-run command).

4. **No Duplication**
   Do not include general coding practices the agent already knows.
   Focus strictly on the unique procedures of this workflow.

---

## Reference Documentation

Refer to the official documentation for the latest specification:

- Home / Overview: https://agentskills.io/home
- Specification: https://agentskills.io/specification
- Quickstart: https://agentskills.io/skill-creation/quickstart
- Best practices: https://agentskills.io/skill-creation/best-practices
- Optimizing descriptions: https://agentskills.io/skill-creation/optimizing-descriptions
- Evaluating skills: https://agentskills.io/skill-creation/evaluating-skills
- Using scripts: https://agentskills.io/skill-creation/using-scripts
