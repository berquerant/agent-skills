# Skill Review Checklist

This file is the single source of truth for skill review criteria. Each item is tagged with the severity it carries when violated:

- **[F]**: violating it makes the dimension **Fail**. These are hard specification rules or rules that affect operational safety.
- **[W]**: violating it makes the dimension at worst **Warn**. These are conventions and best practices.

Authoritative sources:
- Agent Skills Specification: https://agentskills.io/specification
- Authoring conventions: [../../skill-creator/references/skill-spec.md](../../skill-creator/references/skill-spec.md)
- The repository's own agent guidelines, such as `AGENTS.md`, if present.

When a criterion is ambiguous for a particular skill, record the interpretation you applied in the report's **Review Basis** line.

---

## Verdict Rules

**Dimension status**
- **Fail**: at least one [F] item is violated.
- **Warn**: no [F] violations, but at least one [W] item is violated.
- **Pass**: no violations.

**Overall verdict**
- 🔴 **FAIL**: any dimension is Fail.
- 🟡 **NEEDS IMPROVEMENT**: no Fail, but at least one Warn.
- 🟢 **PASS**: every dimension is Pass.

---

## 1. Scope & Granularity

- [ ] **[W] Single responsibility**: The skill covers one focused, coherent operational domain instead of bundling unrelated workflows.
- [ ] **[W] Reasonable step count**: The core workflow has about 3–8 steps. A workflow that needs 10 or more detailed phases should be split, or should delegate to an execution framework through a hand-off.
- [ ] **[W] Distinct trigger**: The trigger in `description` does not substantially overlap with a sibling skill's trigger. Where the two are close, the boundary is stated explicitly.

---

## 2. Modularity & File Split

- [ ] **[F] Size limit**: `SKILL.md` is under 500 lines, as the specification requires.
- [ ] **[W] Concise `SKILL.md`**: `SKILL.md` stays within roughly 80–250 lines and acts as the orchestrator.
- [ ] **[W] Offloaded deep content**:
  - Detailed specifications, checklists, and schemas live in `references/`.
  - Executable helpers live in `scripts/`.
  - Templates and assets live in `resources/` or `assets/`.
  - Reference implementations live in `examples/`.
- [ ] **[W] Shallow references**: Referenced files sit one level deep from `SKILL.md`, with no long chains of reference files pointing to further reference files.
- [ ] **[F] Valid relative links**: Internal links use relative paths that resolve. There are no broken links and no absolute or machine-specific paths.
- [ ] **[W] Executable scripts**: Every file in `scripts/` has execute permission.

---

## 3. Maintainability & Conventions

**Frontmatter, hard rules from the specification**
- [ ] **[F] `name`**: 1–64 characters. Only lowercase letters, digits, and hyphens. Does not start or end with a hyphen and has no consecutive hyphens (`--`). Matches the parent directory name.
- [ ] **[F] `description`**: Non-empty and at most 1024 characters.

**Frontmatter, conventions**
- [ ] **[W] `name` length**: 3–30 characters is recommended, following the skill-creator convention.
- [ ] **[W] `description` format**: The description starts with a trigger clause, `Use this skill when …`. Effects follow as third-person verbs, such as `Analyzes …` or `Produces …`. It states both **when to use** the skill and **what it does**, and avoids vague wording such as "helps with things". The imperative `Use this skill when …` opener is the accepted convention for "third person" in this repository.

**Body structure**
- [ ] **[W] Section order**:
  1. Frontmatter
  2. Title and summary (`# <Skill Title>`)
  3. Prerequisites (only if applicable)
  4. `## Step N: …` sections
  5. `## Guidelines`
- [ ] **[W] Verify checkpoints**: Every `## Step N` section contains an explicit `Verify:` line that a reader can actually check.
- [ ] **[W] Platform and project agnostic**: The skill does not hardcode vendor-specific tool names, proprietary variables, fixed directory layouts, or a single tech stack, unless it labels them as examples and gives a generic fallback.

---

## 4. Loose Coupling

- [ ] **[F] No hard subroutine invocation**: The skill does not require another skill to run as a mandatory nested step.
- [ ] **[W] Sequential hand-off**: The skill communicates through clear outputs or artifacts, and recommends a next skill for the user to approve.
- [ ] **[F] Read-only versus mutation separation**: An audit or review skill never edits, creates, or deletes files in the target. A mutation skill does not mutate files before the user approves, if its own workflow requires approval.
- [ ] **[W] Actionable recommendations**: When the skill identifies improvements or fixes, it names the existing mutation skill that applies, such as `refactor` or `step-gate`.

---

## 5. DRY & Interoperability

- [ ] **[W] Minimal repetition**: The skill does not duplicate checklists, specifications, or command references that already exist in a sibling skill. Within the skill, `SKILL.md` refers to its references rather than restating them.
- [ ] **[W] Shared references**: Where criteria overlap with a sibling skill, the skill links to the shared material with a relative cross-skill path, such as `../code-review/references/security-checklist.md`.
- [ ] **[W] Consistent terminology**: Terms, step names, and status labels match their usage in sibling skills that cover the same concepts.
- [ ] **[W] Repository registration**: The skill is listed in the repository-level docs, such as the README skill table and the architecture diagram in `AGENTS.md`, wherever those docs list skills. Hand-off relationships described there match the skill's actual recommendations.

---

## 6. Self-Consistency

This dimension checks that the target skill is internally coherent.

- [ ] **[W] Description matches body**: Every trigger and effect claimed in `description` and in the summary is implemented by the steps, and no major step goes unmentioned there.
- [ ] **[F] Principles honored**: The steps never contradict the skill's own Guidelines. For example, a skill that declares itself read-only must not include a step that writes files.
- [ ] **[W] Coherent flow**: Inputs, outputs, and terms carry over consistently from one step to the next, with no undefined artifacts and no conflicting instructions.
- [ ] **[W] References match usage**: Every referenced file exists and contains what `SKILL.md` says it contains. No file in the skill directory goes unreferenced.
