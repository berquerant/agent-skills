# Skill Review Checklist

This file is the single source of truth for skill review criteria. Each item is tagged with the severity it carries when violated:

- **[F]**: violating it makes the dimension **Fail**. These are hard specification rules or rules that affect operational safety.
- **[W]**: violating it makes the dimension at worst **Warn**. These are conventions, model-alignment best practices, and maintainability rules.

Authoritative sources:
- Agent Skills Specification: [agent-skills-spec.md](../../skill-creator/references/agent-skills-spec.md) (../../skill-creator/references/agent-skills-spec.md)
- Authoring conventions: [skill-spec.md](../../skill-creator/references/skill-spec.md) (../../skill-creator/references/skill-spec.md)
- The repository's own agent guidelines, such as `AGENTS.md`, if present.

When a criterion requires interpretation for a specific skill, record that interpretation in the report's **Review Basis** line.

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

- [ ] **[W] Single responsibility**: The skill covers one focused, coherent operational domain instead of bundling unrelated workflows, keeping active context lean.
- [ ] **[W] Reasonable step count**: The core workflow has 3–8 steps. A workflow that needs 10 or more detailed phases splits into focused skills or delegates execution through a sequential hand-off.
- [ ] **[W] Distinct & evaluable trigger**: The trigger in `description` defines clear activation boundaries ("should-trigger" vs. "should-not-trigger" scenarios) without overlapping sibling skills or using vague catch-all phrasing. Where two skills are close, the boundary is stated explicitly.
- [ ] **[W] Risk-proportional guardrails**: Approval gates, snippet extraction rules, and pacing thresholds scale with operational blast radius—allowing low-friction execution and parallel tool calls during read-only or reversible exploration while enforcing strict gates on state-mutating, destructive, or high-cost operations.

---

## 2. Modularity & File Split

- [ ] **[F] Size limit**: `SKILL.md` is under 500 lines, as the specification requires.
- [ ] **[W] Concise & top-loaded `SKILL.md`**: `SKILL.md` stays within roughly 80–250 lines as the orchestrator, and places the start of its core workflow (`## Step 1`) and critical routing boundaries within the **first 100 lines** so models immediately attend to primary execution logic.
- [ ] **[W] Offloaded deep content (3-level progressive disclosure)**:
  - Detailed specifications, checklists, and schemas live in `references/` and are loaded on demand (just-in-time) rather than preloaded unconditionally.
  - Executable helpers live in `scripts/`.
  - Templates and assets live in `resources/` or `assets/`.
  - Reference implementations live in `examples/`.
- [ ] **[W] Shallow references**: Auxiliary files in `references/`, `scripts/`, `examples/`, and `resources/` sit one level deep and are linked directly from `SKILL.md`, avoiding multi-hop reference chains that fragment attention and inflate tool turns.
- [ ] **[F] Valid relative links**: Internal links use relative paths that resolve to existing files, with zero broken links and zero absolute or machine-specific paths.
- [ ] **[F] Offline & self-contained**: Skills rely on local files in `references/` rather than runtime network fetches for instructions, guidelines, or schemas.
- [ ] **[W] Version-anchored snapshots & no autonomous external URL access**: Local snapshots of external specifications, SDKs, or standards record a target version or verification note so staleness is detectable. Any external hyperlinks are strictly human-only citations accompanied by an explicit note forbidding autonomous agent access.
- [ ] **[W] Executable scripts**: Every file in `scripts/` has execute permission (`chmod +x`).

---

## 3. Maintainability & Conventions

**Frontmatter, hard rules from the specification**
- [ ] **[F] `name`**: 1–64 characters. Only lowercase letters, digits, and single hyphens (`^[a-z0-9]+(-[a-z0-9]+)*$`). Matches the parent directory name.
- [ ] **[F] `description`**: Non-empty and at most 1024 characters.

**Frontmatter, conventions (Layer 1 token economy)**
- [ ] **[W] `name` length**: 3–30 characters is recommended, following the skill-creator convention.
- [ ] **[W] `description` format & length**: Starts with a trigger clause (`Use this skill when …` or `Use this skill to …`), followed by third-person effect verbs (`Analyzes …`, `Produces …`). States both **when to use** the skill and **what it does** within **100–350 characters** to keep Layer 1 startup metadata concise across the skill catalog.

**Body structure & instruction design**
- [ ] **[W] Section order**:
  1. Frontmatter
  2. Title and summary (`# <Skill Title>`)
  3. Prerequisites (only if applicable)
  4. `## Step N: …` sections
  5. `## Guidelines` (final top-level section)
- [ ] **[W] Boundary & contract over manual CoT (`Verify:` checkpoints)**: Every `## Step N` section ends with an observable, testable `Verify:` contract rather than scripting manual chain-of-thought prompts ("think step by step") that interfere with native reasoning models.
- [ ] **[W] Positive guidance with rationale ("Why")**: Instructions and `## Guidelines` pair directives with brief rationale ("why" the rule matters) and prefer affirmative guidance so models generalize accurately across edge cases, rather than listing unexplained negative prohibitions.
- [ ] **[W] Output presentation compliance**: Instructions, report templates, and procedural steps communicate file paths, line ranges, and web URLs by co-presenting accessible Markdown links and unadorned raw strings in parentheses without inline backticks on raw targets, following [output-presentation](../../output-presentation/SKILL.md) (../../output-presentation/SKILL.md).
- [ ] **[W] Command snippet modularity**: Inline shell examples stay concise (under 4 chained operators, under 120 characters, and under 3 lines) and delegate multi-line or complex logic to `scripts/`, following [command-snippet](../../command-snippet/SKILL.md) (../../command-snippet/SKILL.md).
- [ ] **[W] Platform and project agnostic**: The skill uses portable terminology and standard tools, presenting vendor-specific CLIs, model parameters, or tech stacks only as non-exclusive examples with generic fallbacks.

---

## 4. Loose Coupling

- [ ] **[F] No hard subroutine invocation**: The skill operates independently without requiring another skill to run as a mandatory nested step.
- [ ] **[W] Sequential hand-off & context isolation**: The skill communicates through clear outputs or externalized artifacts (supporting context compaction) and recommends a next skill for the user to approve, avoiding cascading multi-skill preloads in a single turn.
- [ ] **[F] Read-only versus mutation separation**: An audit or review skill remains strictly read-only and never edits, creates, or deletes files in the target workspace. A mutation skill obtains explicit user approval before modifying files whenever its workflow defines an approval gate.
- [ ] **[W] Actionable recommendations**: When the skill identifies improvements or fixes, it names the existing mutation skill that applies, such as `refactor`, `step-gate`, or `doc-refactor` (for documentation and comment refinements).

---

## 5. DRY & Interoperability

- [ ] **[W] Minimal repetition**: The skill links to its `references/` and sibling references rather than restating checklists, specifications, or generic coding knowledge the model already possesses.
- [ ] **[W] Shared references**: Where criteria overlap with a sibling skill, the skill links to the shared material with a relative cross-skill path, such as [security-checklist.md](../../code-review/references/security-checklist.md) (../../code-review/references/security-checklist.md).
- [ ] **[W] Consistent terminology**: Terms, step names, and status labels match their usage in sibling skills that cover the same concepts.
- [ ] **[W] Repository registration**: The skill is listed across repository-level docs—such as the `README.md` skill tables, the architecture diagram in `AGENTS.md`, and routing matrices like [intent-matrix.md](../../skill-router/references/intent-matrix.md) (../../skill-router/references/intent-matrix.md) where present—and hand-off relationships match across all listings.

---

## 6. Self-Consistency

This dimension checks that the target skill is internally coherent across all of its files.

- [ ] **[W] Description matches body**: Every trigger and effect claimed in `description` and in the summary is implemented by the steps, and no major step goes unmentioned there.
- [ ] **[F] Principles honored**: The steps honor the skill's own `## Guidelines` and declared constraints (for example, a read-only skill contains zero file-mutating steps).
- [ ] **[W] Coherent flow & API/version consistency**: Inputs, outputs, terms, CLI flags, SDK imports, and version assumptions carry over consistently across `SKILL.md`, `references/`, `examples/`, and `scripts/` without internal contradictions or version mismatches.
- [ ] **[W] References match usage & zero orphan files**: Every referenced file exists and contains what `SKILL.md` says it contains, and every file in the skill directory (`references/`, `scripts/`, `examples/`, `resources/`) is directly referenced from `SKILL.md`.
