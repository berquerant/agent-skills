---
name: skill-modernize
description: >-
  Use this skill when evolving agent skills and review criteria in this repository
  to align with new frontier model capabilities, harness updates, or SDK versions.
  Analyzes model deltas, updates skill-review standards and checks, self-reviews
  the baseline, and coordinates phased skill remediation.
---

# Skill Modernize

Use this repository-local meta-skill (`.agent/skills/skill-modernize/`) to continuously evolve the `skills/` catalog and its review standards as frontier models, agent harnesses, and external SDKs advance.
Rather than patching individual skills ad hoc, this workflow first updates the evaluation ruler ([skill-review](../../../skills/skill-review/SKILL.md) (../../../skills/skill-review/SKILL.md) and [skill-creator](../../../skills/skill-creator/SKILL.md) (../../../skills/skill-creator/SKILL.md)), verifies the ruler against itself, and then orchestrates a batch audit and phased remediation across the repository.

For the 5-axis model delta matrix, the 3-phase remediation plan template, and the self-modernization checklist, refer to [modernization-playbook.md](references/modernization-playbook.md) (references/modernization-playbook.md).

---

## Prerequisites

- Read and write access to `skills/`, `.agent/skills/`, and repository index files (`README.md`, `AGENTS.md`).
- A POSIX-compatible shell to run `./bin/lint.sh check`, `./bin/check-skills.sh`, and [run-checks.sh](../../../skills/skill-review/scripts/run-checks.sh) (../../../skills/skill-review/scripts/run-checks.sh).

---

## Step 1: Analyze Model, Harness & Ecosystem Deltas

Identify what has changed in model capabilities, harness mechanics, or ecosystem SDKs and map those changes to concrete skill conventions:

1. **Gather Evidence**:
   - Inspect user-provided research reports or conduct an evidence-backed investigation via [evidentiary-inquiry](../../../skills/evidentiary-inquiry/SKILL.md) (../../../skills/evidentiary-inquiry/SKILL.md) to pin exact model/SDK behaviors and version boundaries.
2. **Evaluate Across the Five Modernization Axes**:
   - Map each finding against the **Five-Axis Model & Harness Delta Matrix** in [modernization-playbook.md](references/modernization-playbook.md) (references/modernization-playbook.md):
     1. **Reasoning & Step Granularity**: Replace manual chain-of-thought prompts or over-scripted micro-steps with concise 3–8 step workflows and observable `Verify:` contracts.
     2. **Context & Token Economy**: Enforce Layer 1 `description` budgets (100–350 chars), top-loaded Layer 2 `SKILL.md` files (80–250 lines, `## Step 1` $\le$ line 100), and 1-hop Layer 3 `references/` links.
     3. **Tool Use & Risk-Proportional Guardrails**: Allow low-friction parallel read batches during exploration while enforcing strict approval gates on state-mutating or high-cost operations.
     4. **Instruction Framing & Contracts**: Convert unexplained negative prohibitions into affirmative, **Why-first** principles in `## Guidelines`.
     5. **Ecosystem, SDK & Output Hygiene**: Pin external spec snapshots to versions/dates, unify SDK versions across `SKILL.md` / `references/` / `examples/`, keep inline shell commands $<120$ chars, and enforce [output-presentation](../../../skills/output-presentation/SKILL.md) (../../../skills/output-presentation/SKILL.md) dual links.
3. **Confirm Modernization Scope**:
   - Present the proposed criteria updates and affected skill scope to the user (notifying via [user-notification](../../../skills/user-notification/SKILL.md) (../../../skills/user-notification/SKILL.md)) and obtain approval before editing files.

Verify: Model/harness deltas are mapped across the 5 axes in `modernization-playbook.md`, and the user has approved the modernization scope.

---

## Step 2: Evolve Review Criteria, Automated Checks & Creator Templates

Update the repository's evaluation ruler and scaffolding templates first so all subsequent reviews and newly created skills enforce the new standard:

1. **Update `skill-review` Criteria & Automated Checks**:
   - Update [skill-review-checklist.md](../../../skills/skill-review/references/skill-review-checklist.md) (../../../skills/skill-review/references/skill-review-checklist.md) with explicit `[F]` (hard error) and `[W]` (warning) items for the new model-alignment criteria.
   - Implement reproducible mechanical checks in [run-checks.sh](../../../skills/skill-review/scripts/run-checks.sh) (../../../skills/skill-review/scripts/run-checks.sh) (including `--strict` mode support) and document them in [automated-checks.md](../../../skills/skill-review/references/automated-checks.md) (../../../skills/skill-review/references/automated-checks.md) and [report-template.md](../../../skills/skill-review/references/report-template.md) (../../../skills/skill-review/references/report-template.md).
2. **Synchronize `skill-creator` & Repository Conventions**:
   - Update generation guidelines in [skill-creator](../../../skills/skill-creator/SKILL.md) (../../../skills/skill-creator/SKILL.md) and convention rules in `AGENTS.md` so new skills automatically inherit the updated standards.

Verify: `skill-review` checklists, `run-checks.sh`, `skill-creator`, and `AGENTS.md` are updated and pass `./bin/lint.sh check`.

---

## Step 3: Self-Review & Lock the Baseline Standard

Audit the updated meta-skills against their own criteria before reviewing the rest of the catalog:

1. **Run Strict Automated Checks on Meta-Skills**:
   ```sh
   skills/skill-review/scripts/run-checks.sh --strict skills/skill-review
   skills/skill-review/scripts/run-checks.sh --strict .agent/skills/skill-modernize
   ```
2. **Evaluate All Six Qualitative Dimensions**:
   - Inspect both `skill-review` and `skill-modernize` against all 6 dimensions in [skill-review-checklist.md](../../../skills/skill-review/references/skill-review-checklist.md) (../../../skills/skill-review/references/skill-review-checklist.md) and the **Self-Modernization Checklist** in [modernization-playbook.md](references/modernization-playbook.md) (references/modernization-playbook.md).
   - Remediate any self-review `[F]` or `[W]` findings immediately until both skills achieve `🟢 PASS (0 errors, 0 warnings)`.

Verify: Both `skill-review` and `skill-modernize` pass `run-checks.sh --strict` with `0 errors, 0 warnings` and score `🟢 PASS` across all 6 dimensions.

---

## Step 4: Run Batch Audit & Formulate Phased Remediation Plan

Audit all skills in the repository using the locked `skill-review` baseline and organize fixes into a phased execution plan:

1. **Execute Batch `skill-review`**:
   - Run `run-checks.sh` and the 6-dimension evaluation across all skills in `skills/*` (and `.agent/skills/*`).
   - Save the complete batch review report (with a Verdict Matrix, Cross-Skill Issues, and per-skill findings) to a user-confirmed path following [checkpoint-log](../../../skills/checkpoint-log/SKILL.md) (../../../skills/checkpoint-log/SKILL.md).
2. **Structure the Three-Phase Remediation Plan**:
   - Group all findings using the **Three-Phase Repository Remediation Template** in [modernization-playbook.md](references/modernization-playbook.md) (references/modernization-playbook.md):
     - **Phase 1 (Priority 0 — [step-gate](../../../skills/step-gate/SKILL.md) (../../../skills/step-gate/SKILL.md))**: Resolve all `[F]` errors, SDK contradictions, orphan/multi-hop links, $>350$-char descriptions, $\ge 120$-char commands, duplicate headings, and registration gaps.
     - **Phase 2 (Priority 1 — [refactor](../../../skills/refactor/SKILL.md) (../../../skills/refactor/SKILL.md))**: Extract heavy inline templates to `references/` and calibrate risk-proportional guardrails.
     - **Phase 3 (Priority 2 — [doc-refactor](../../../skills/doc-refactor/SKILL.md) (../../../skills/doc-refactor/SKILL.md))**: Standardize `output-presentation` dual links and affirmative Why-first `## Guidelines` across all skills.
3. **Request User Approval**:
   - Present the batch report and 3-phase plan to the user and confirm whether to execute step-by-step or phase-by-phase.

Verify: The batch audit report is saved, findings are categorized into Phases 1–3, and explicit user approval is obtained before modifying target skills.

---

## Step 5: Coordinate Phased Execution & Verify Repository-Wide Compliance

Execute the approved phases sequentially, logging milestones to [checkpoint-log](../../../skills/checkpoint-log/SKILL.md) (../../../skills/checkpoint-log/SKILL.md) and verifying repository-wide compliance at each gate:

1. **Execute & Verify Each Phase**:
   - Hand off or execute Phase 1 ([step-gate](../../../skills/step-gate/SKILL.md) (../../../skills/step-gate/SKILL.md)), Phase 2 ([refactor](../../../skills/refactor/SKILL.md) (../../../skills/refactor/SKILL.md)), and Phase 3 ([doc-refactor](../../../skills/doc-refactor/SKILL.md) (../../../skills/doc-refactor/SKILL.md)) with user confirmation and clean commits via [change-message](../../../skills/change-message/SKILL.md) (../../../skills/change-message/SKILL.md) at each phase boundary.
2. **Run Full Repository Strict Verification**:
   ```sh
   ./bin/lint.sh check && ./bin/check-skills.sh
   ```
3. **Record Post-Remediation Matrix**:
   - Append the before/after verdict matrix to the batch report and final milestone entry to the checkpoint log, then recommend [code-review](../../../skills/code-review/SKILL.md) (../../../skills/code-review/SKILL.md) for final branch audit before merging.

Verify: `./bin/lint.sh check && ./bin/check-skills.sh` exits with `0` (`0 errors, 0 warnings` across every skill in `skills/*` and `.agent/skills/*`), and the final verification matrix is recorded.

---

## Guidelines

- **Update the evaluation ruler before patching individual skills**: Evolve `skill-review` criteria and `run-checks.sh` first so every subsequent skill edit is measured against a single, reproducible standard.
- **Enforce self-referential baseline verification**: Require `skill-review` and `skill-modernize` to pass `run-checks.sh --strict` with zero warnings before auditing the rest of the catalog.
- **Scale guardrails proportionally to operational risk**: Calibrate skill instructions so read-only local exploration remains low-friction and parallelizable while state-mutating or external actions retain explicit approval gates.
- **Execute repository-wide changes in reviewable phases**: Separate mechanical/accuracy fixes (Phase 1), structural/guardrail refactoring (Phase 2), and convention/documentation polish (Phase 3) with explicit checkpoints and commits.
- **Preserve full traceability via dual links and checkpoint logs**: Co-present all file paths and line ranges per [output-presentation](../../../skills/output-presentation/SKILL.md) (../../../skills/output-presentation/SKILL.md) and record every phase outcome in a user-confirmed [checkpoint-log](../../../skills/checkpoint-log/SKILL.md) (../../../skills/checkpoint-log/SKILL.md).
