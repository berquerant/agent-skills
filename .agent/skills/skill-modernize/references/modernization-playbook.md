# Skill Modernization Playbook

Reference documentation for `skill-modernize`: evaluating frontier model and harness deltas, evolving review criteria and automated checks, and orchestrating repository-wide phased remediation.

---

## 1. Five-Axis Model & Harness Delta Matrix

When a new model generation, agent harness update, or SDK version is released, evaluate its impact on existing skills across these five axes:

| Axis | Legacy Skill Anti-Pattern | Frontier Model-Aligned Target | Target Review / Check Artifact |
|---|---|---|---|
| **1. Reasoning & Step Granularity** | Manual chain-of-thought prompts (`think step by step`) or rigid 12+ micro-step scripts that constrain native reasoning | Concise 3–8 step workflows governed by observable `Verify:` contracts | `skill-review-checklist.md` (Dims 1 & 3), `run-checks.sh` (Check D & Check K) |
| **2. Context & Token Economy** | Bloated frontmatter `description` (>350 chars), >250-line `SKILL.md` files, or multi-hop reference chains | 100–350 char `description` (Layer 1), 80–250 line top-loaded `SKILL.md` (`## Step 1` $\le$ line 100, Layer 2), and 1-hop `references/` links (Layer 3) | `skill-review-checklist.md` (Dims 2 & 3), `run-checks.sh` (Checks A, C, D) |
| **3. Tool Use & Risk-Proportional Guardrails** | Uniformly high friction on read-only exploration, or per-call counters that penalize parallel read tool batches | Low-friction parallel read batches (counted as 1 turn) paired with strict approval gates on state-mutating, destructive, or external operations | `skill-review-checklist.md` (Dim 1), `edit-paced-execution`, `command-snippet`, `evidentiary-inquiry` |
| **4. Instruction Framing & Contracts** | All-caps negative prohibition lists (`DO NOT`, `NEVER`) without rationale, causing brittle edge-case generalization | Affirmative, **Why-first** principles in `## Guidelines` explaining the architectural invariant before any boundary | `skill-review-checklist.md` (Dim 3), `skill-creator`, `AGENTS.md` |
| **5. Ecosystem, SDK & Output Hygiene** | Version-unanchored specs, internal SDK mismatches (e.g., v1 vs. v2 imports), $\ge 120$-char inline shell commands, or single-format links | Version-dated local snapshots, unified SDK examples, $<120$-char shell commands, and dual-formatted links (`[label](path) (path)`) | `skill-review-checklist.md` (Dims 2, 3, 6), `run-checks.sh` (Checks E, J, K) |

---

## 2. Three-Phase Repository Remediation Template

After updating `skill-review` and locking the baseline via self-review, group batch audit findings into three sequential phases so changes remain reviewable and verifiable at every commit boundary:

```markdown
### Skill Modernization Remediation Plan

- **Baseline Standard Commit**: `<commit-hash>` (`skill-review` & `skill-modernize` self-verified at `0 errors, 0 warnings`)
- **Batch Audit Report**: [<report-file>](<accessible-report-path>) (<raw-report-path>)
- **Checkpoint Log**: [<log-file>](<accessible-log-path>) (<raw-log-path>)

#### Phase 1: Priority 0 — Automated Check & Accuracy Fixes (via `step-gate`)
- **Target Skills**: `<list of skills with [F] or mechanical [W] findings>`
- **Scope**:
  - Resolve internal SDK/CLI version contradictions (`[F]`).
  - Link orphan or multi-hop files directly (1-hop) from `SKILL.md`.
  - Compress frontmatter `description` fields into the 100–350 char Layer 1 budget.
  - Wrap or extract $\ge 120$-char inline shell commands and fix duplicate reference headings.
  - Register missing skills in `README.md`, `AGENTS.md`, and `intent-matrix.md`.
- **Verification Gate**: `run-checks.sh --strict` reports `0 errors, 0 warnings` across all modified skills.

#### Phase 2: Priority 1 — Structural & Risk-Proportional Refactoring (via `refactor`)
- **Target Skills**: `<list of skills needing structural split or guardrail calibration>`
- **Scope**:
  - Extract heavy inline templates or reference tables from `SKILL.md` into `references/` to keep `SKILL.md` $\le 250$ lines.
  - Synchronize `skill-creator` templates and generation checklists with the updated `skill-review` criteria.
  - Calibrate approval gates, snippet thresholds, and pacing counters to be proportional to operational risk and parallel tool-call behavior.
- **Verification Gate**: `./bin/lint.sh check && ./bin/check-skills.sh` passes with zero behavioral regressions.

#### Phase 3: Priority 2 — Convention & Why-First Documentation Polish (via `doc-refactor`)
- **Target Skills**: `<remaining skills + root AGENTS.md>`
- **Scope**:
  - Convert negative prohibition lists in `## Guidelines` into affirmative, Why-first principles.
  - Standardize all file, directory, and cross-skill links to `output-presentation` dual format (`[label](path) (path)`).
  - Align root `AGENTS.md` and `README.md` conventions with the modernized baseline.
- **Verification Gate**: All skills in `skills/*` and `.agent/skills/*` achieve `🟢 PASS (0 [F], 0 [W])` across both `run-checks.sh --strict` and the 6 qualitative review dimensions.
```

---

## 3. Self-Modernization Checklist (Meta-Verification)

When applying `skill-modernize` to itself (`.agent/skills/skill-modernize/`):

1. **Delta Check**: Confirm that `modernization-playbook.md` covers all active axes in [skill-review-checklist.md](../../../../skills/skill-review/references/skill-review-checklist.md) (../../../../skills/skill-review/references/skill-review-checklist.md) and all automated checks A–K in [automated-checks.md](../../../../skills/skill-review/references/automated-checks.md) (../../../../skills/skill-review/references/automated-checks.md).
2. **Mechanical Check**: Run `skills/skill-review/scripts/run-checks.sh --strict .agent/skills/skill-modernize` and verify `PASSED (0 errors, 0 warnings)`.
3. **Six-Dimension Check**: Evaluate `skill-modernize` across all 6 dimensions in `skill-review` (`1. Scope`, `2. Modularity`, `3. Conventions`, `4. Coupling`, `5. DRY`, `6. Self-Consistency`) and confirm `🟢 PASS (0 [F], 0 [W])`.
