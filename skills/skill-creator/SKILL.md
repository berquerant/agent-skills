---
name: skill-creator
description: >-
  Use this skill when the user wants to create a new agent skill. Checks for
  overlap with existing skills, guides through requirements collection, generates
  SKILL.md with proper frontmatter and instructions, scaffolds the directory
  structure in the appropriate location, and updates repository skill listings.
---

# Skill Creator

Use this skill to create a new agent skill.
Collect requirements, generate a `SKILL.md` with proper frontmatter and instructions,
and scaffold the directory structure.

For the skill specification, refer to [references/skill-spec.md](references/skill-spec.md).

---

## Step 1: Requirements Collection

If the user's intent is clear, draft a proposal and ask for confirmation.
If it is ambiguous, ask the user for the following information.

**Information to collect:**

1. **Purpose** — What should this skill do?
2. **Trigger conditions** — When should the agent activate this skill? (Provide concrete scenarios.)
3. **Steps** — What specific steps should the agent take?
4. **Placement** — Ask for placement preference (options presented in Step 2). Never assume a default.

5. **Additional files** — Are any of the following needed?
   - `scripts/` : Helper scripts (for automating complex command sequences)
   - `references/` : Detailed documentation (for content that doesn't fit in SKILL.md)
   - `examples/` : Reference implementations (when concrete examples are helpful)
   - `resources/` : Templates or assets (for reusable files)

Verify: Skill purpose, concrete triggers, planned steps, placement preference, and auxiliary file requirements are collected and confirmed with the user.

---

## Step 2: Determine Placement

**Always confirm with the user before placing any files. Do not auto-detect. Do not assume defaults.**

Present the following options and let the user choose:

- **Project-specific**: `skills/<name>/` (when `.agents/skills` is a symlink to `skills/`), or `.agents/skills/<name>/` / `_agents/skills/<name>/` in the workspace
  - If `.agents/`, `_agents/`, `.agent/`, or `_agent/` already exists in the current workspace, inform the user.
  - If `.agents/skills` is a symlink, write to the symlink target and never replace the symlink with a regular directory.
- **Global**: Ask the user for the exact target path directly.
  The path varies by agent tool and environment, so never assume it.

Check for name conflicts: verify that no skill with the same name already exists in the target directory.
If one is found, notify the user and suggest an alternative name.

Check for overlap: skim the `description` of existing skills in the target directory.
If an existing skill already covers part of the purpose, propose either extending that skill
or sharing its `references/` via relative links (e.g. `../<other-skill>/references/<file>.md`) instead of duplicating content.

Verify: Target directory is confirmed with the user, no name conflict exists, and any overlap with existing skills is resolved with the user.

---

## Step 3: Generate SKILL.md

### 3.1 Quality Checklist for the `name` Field

Refer to [references/skill-spec.md](references/skill-spec.md) for character set and length constraints.
Key principles:
- Lowercase, hyphen-separated (e.g., `my-skill`, `deploy-checker`)
- Must match the parent directory name exactly

### 3.2 Quality Checklist for the `description` Field

The `description` is the most critical field — it determines whether the agent activates the skill.
Refer to [references/skill-spec.md](references/skill-spec.md) for length constraints.
All of the following must be satisfied:

- [ ] Written in **third person** (convention: starts with "Use this skill when..." followed by effects like "Guides the agent to...", "Analyzes...")
- [ ] Clearly states **when to use it** (trigger conditions)
- [ ] Clearly states **what it does** (effect or output)
- [ ] Recommended roughly 100–350 characters
- [ ] Avoids vague language ("various tasks", "helps with things")

**Good example:**
```
Use this skill when the user wants to run integration tests for the XYZ service.
Executes the full test suite, captures logs, and reports failures with suggested fixes.
```

**Bad example:**
```
A skill that helps with testing things.
```

### 3.3 SKILL.md Body Structure

Generate the skill content following the template below.
Add or remove sections as appropriate for the skill's nature.

```markdown
---
name: <name>
description: >-
  <description>
---

# <Title>

<One or two sentences describing the purpose of this skill.>

---

## Prerequisites (only if applicable)

- Required tools, environments, or permissions

## Step 1: <First Step Name>

<Specific instructions>

Verify: <How to confirm this step succeeded>

## Step 2: <Next Step Name>

<Specific instructions>

...

## Guidelines

- <Non-negotiable operational principles, one bullet each>
```

**Generation guidelines:**
- Do not include general coding knowledge (the agent already knows this)
- Focus only on procedures, commands, and notes specific to this workflow
- Include a verification method for each step
- Move heavy documentation to `references/` and link to it from SKILL.md
- Keep the skill platform-agnostic: present platform-specific tools or commands only as non-exclusive examples
- If the target repository defines its own skill conventions (e.g. `AGENTS.md`, `CONTRIBUTING.md`), follow them over this template

Verify: Generated `SKILL.md` strictly adheres to the frontmatter specification, includes explicit `Verify:` checkpoints in every step, ends with a `## Guidelines` section, and keeps core instructions concise.

---

## Step 4: Scaffold Directory Structure

1. **Create the main directory**: `<placement>/<name>/`
2. **Write SKILL.md**: with the content generated in Step 3
3. **Create subdirectories** (only if needed):
   - `scripts/` — Helper script stubs (set executable permission with `chmod +x`)
   - `references/` — Reference documentation stubs
   - `examples/` — Reference implementation stubs
   - `resources/` — Directory for templates and assets
4. **Update repository indexes** (only if they exist): add the new skill to skill listings
   such as `README.md` tables, directory trees, or architecture diagrams in `AGENTS.md`.
   If existing skills should hand off to or reference the new skill, propose those edits.
   **Always obtain explicit user confirmation before modifying any files outside the new skill directory.**

Verify: Target directory layout is created with correct permissions, all relative links resolve, and proposed updates to repository indexes are confirmed and applied.

---

## Step 5: Confirm and Follow Up

1. Present the list of generated and modified files to the user.
2. Notify the user of:
   - The exact path where the skill was placed
   - Whether the skill will be auto-discovered or requires manual registration (depends on the agent platform)
   - That the skill can be invoked via a slash command `/<name>` (if supported)
3. Encourage the user to review and edit the content as needed.
4. Recommend reviewing the new skill with `skill-review` (if available).

Verify: File manifest and invocation instructions are presented to the user, and user confirmation is received.

---

## Guidelines

- **Never assume placement locations.** Always ask and confirm the target directory directly with the user before creating files.
- **Third-person descriptions with triggers.** Ensure the `description` frontmatter states both when to use the skill and what it does.
- **Explicit verification for every step.** Provide measurable `Verify:` checkpoints for all procedural steps.
- **Concise core instructions.** Keep `SKILL.md` lean; extract expansive references or specs into `references/`.
