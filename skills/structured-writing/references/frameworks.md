# Structured Writing Frameworks Reference

This reference provides detailed definitions, origins, templates, technical examples, and self-check criteria for the frameworks used in `structured-writing`.

---

## 1. Document Genre Classification: Diataxis

Developed by Daniele Procida, **Diataxis** categorizes technical documentation across two axes—**Practical vs. Theoretical** and **Action vs. Cognition**—to prevent mixing incompatible goals in a single document.

| Quadrant | Orientation | Reader's Goal | Author's Focus | Example |
|---|---|---|---|---|
| **Tutorials** | Learning | Acquire competence by doing | Safe, guided, reproducible lesson | "Build your first plugin in 10 minutes" |
| **How-to Guides** | Task | Solve a specific real-world problem | Direct steps assuming basic competence | "How to rotate signing keys without downtime" |
| **Reference** | Information | Look up exact facts | Dry, structured, exhaustive accuracy | CLI flag table, API schema, config keys |
| **Explanation** | Understanding | Grasp context and design rationale | Discursive "Why", trade-offs, architecture | "Why the worker pool uses bounded queues" |

---

## 2. Framework Selection Matrix

| Framework | Best For | Length / Medium | Core Strength |
|---|---|---|---|
| **BLUF** | Urgent updates, approvals, status messages | Short (1–3 paragraphs) | Immediate actionability for busy readers |
| **PREP** | Concise opinions, Q&A, short proposals | Short to Medium | Clear claim backed by logic and a concrete proof point |
| **Pyramid Principle** | Design docs, RCAs, technical specifications | Medium to Long | Airtight top-down hierarchy with zero logical gaps |
| **New PASONA** | Proposals requiring buy-in and action | Medium | Empathy-first persuasion without aggressive hype |
| **B-C-A / BAB** | Migration reports, case studies, release notes | Short to Medium | Vivid contrast between old pain and new outcome |
| **ADR** | Architectural and tooling decisions | Short to Medium | Preserves context and trade-offs alongside code |
| **STAR** | Postmortem timelines, spikes, project summaries | Medium | Reproducible account of constraints, actions, and results |

---

## 3. Pyramid Principle (SCQA + MECE)

Systematized by Barbara Minto (*The Pyramid Principle*, McKinsey & Company), this framework structures complex arguments top-down so busy readers grasp the governing thought first.

### Components
1. **Introduction (SCQA)**:
   - **S (Situation)**: Undisputed context the reader already knows.
   - **C (Complication)**: The change, bottleneck, or problem that disrupts the situation.
   - **Q (Question)**: The specific question the complication raises.
   - **A (Answer)**: The single governing thought at the top of the pyramid.
2. **Vertical Logic (Q&A Dialogue)**:
   - Every tier below answers a "Why?" or "How?" question triggered by the tier above.
3. **Horizontal Logic (Deductive / Inductive + MECE)**:
   - **Inductive** (most common): Group 2–4 distinct arguments sharing a common property to support the parent claim.
   - **Deductive**: Chain premises (`Premise A -> Premise B -> Conclusion`).
   - **MECE** (*Mutually Exclusive, Collectively Exhaustive*): Sibling points must not overlap and must cover all relevant aspects without gaps.

### Technical Template (Design Doc / RCA)
```markdown
## Context (SCQA)
- **Situation**: <Shared baseline state>
- **Complication**: <Observed failure, bottleneck, or scaling limit>
- **Decision / Answer**: <Single governing recommendation>

## Supporting Pillars (MECE)
1. **<Pillar 1: e.g., Reliability>** — <Evidence / benchmarks>
2. **<Pillar 2: e.g., Latency>** — <Evidence / benchmarks>
3. **<Pillar 3: e.g., Operational Cost>** — <Evidence / benchmarks>
```

---

## 4. PREP & BLUF

Both frameworks reject "bottom-up" storytelling where conclusions are buried at the end.

### BLUF (Bottom Line Up Front)
Originated in U.S. military communications to respect the reader's time. Place the decision, conclusion, or requested action in the very first sentence; add context only if needed.

```markdown
**Action required**: Approve upgrading the CI runner image to `v2.4` by Friday to unblock security patches.

**Context**: The current `v2.1` image reaches end-of-life next week. Dry-run builds on `service-a` passed with zero regressions.
```

### PREP (Point, Reason, Example, Point)
Best when you must both state a conclusion and defend it briefly.

- **P (Point)**: Core claim or recommendation.
- **R (Reason)**: Logical principle or architectural cause behind the claim.
- **E (Example)**: Concrete data, benchmark, or observed incident proving the reason.
- **P (Point)**: Restated conclusion to close cleanly.

```markdown
We should adopt structured JSON logging for `worker-service` (Point).
Plain-text logs break automated parsing when stack traces span multiple lines, delaying incident triage (Reason).
In last week's outage, engineers spent 25 minutes grepping multiline logs compared to 2 minutes in services already emitting JSON (Example).
Therefore, migrating `worker-service` to structured JSON logging will cut mean-time-to-diagnose (Point).
```

---

## 5. New PASONA (Affinity Model)

Updated by Masanori Kanda in 2016 (*Kasegu Kotoba no Hosoku*), **New PASONA** replaces the 1999 model's `Agitation` (fear-mongering) with `Affinity` (empathy) and separates `Offer` from `Solution`.

| Step | Marketing Meaning | Engineering & Proposal Adaptation |
|---|---|---|
| **P**roblem | Identify the reader's pain | State the concrete operational or developer pain |
| **A**ffinity | Empathize with the struggle | Acknowledge team constraints and why the pain persists |
| **S**olution | Present the general approach | Describe the technical direction or pattern |
| **O**ffer | Detail the concrete package | Specify migration scope, effort estimate, and rollout plan |
| **N**arrowing Down | Reason to act now | Note compounding tech debt, upcoming freezes, or SLA risk |
| **A**ction | Clear next step | Request a specific decision (e.g., approve RFC, schedule spike) |

---

## 6. B-C-A & BAB (Before - After - Bridge)

Both contrast a problematic current state with an improved outcome and explain how to cross the gap.

- **BAB (Before -> After -> Bridge)**: Popularized in digital copywriting to inspire action. Show the current pain (**Before**), paint the desired target state (**After**), and present the solution as the **Bridge**.
- **B-C-A (Before -> Change -> After)**: Chronological variant common in technical case studies and post-migration reports. Describe the legacy state (**Before**), the engineering intervention (**Change**), and the measured outcome (**After**).

```markdown
- **Before**: Deployments to staging required 45 minutes of manual script execution and frequently failed due to host drift.
- **Change**: We containerized the pipeline and moved environment configuration into declarative manifests.
- **After**: Deployments now finish in 6 minutes with zero drift-related rollbacks over the past 30 runs.
```

---

## 7. ADR (Architecture Decision Record)

Coined by Michael Nygard (2011) to capture significant technical decisions alongside the codebase.

```markdown
# ADR-001: <Short Noun Phrase Decision>

- **Status**: Proposed | Accepted | Deprecated | Superseded by ADR-XXX

## Context
<Technical, organizational, and operational forces at play.>

## Decision
<Active-voice statement: "We will use ...">

## Consequences
- **Positive**: <Benefits gained>
- **Negative / Trade-offs**: <Accepted complexity, constraints, or migration debt>
```

---

## 8. STAR (Situation, Task, Action, Result)

Structured narrative for incident reviews, technical spikes, and engineering retrospectives.

- **S (Situation)**: Baseline system state and triggering event.
- **T (Task)**: Specific goal, SLA, or constraint that had to be met.
- **A (Action)**: Concrete steps, experiments, or fixes executed.
- **R (Result)**: Measurable outcome and follow-up learnings.

---

## 9. Self-Check & Anti-Patterns

Run these checks before finalizing a draft:

- **BLUF / PREP**:
  - [ ] Is the conclusion in the first sentence/paragraph rather than buried at the end?
  - [ ] In PREP, is `Reason` a general logical cause rather than a single anecdote (`Example`)?
- **Pyramid Principle**:
  - [ ] Does SCQA establish shared context before presenting the governing thought?
  - [ ] Are sibling pillars MECE (no overlapping categories, no obvious missing dimension)?
- **New PASONA**:
  - [ ] Does `Affinity` acknowledge real constraints without patronizing language?
  - [ ] Does `Narrowing Down` cite objective technical urgency rather than artificial scarcity?
- **B-C-A / BAB**:
  - [ ] Are `Before` and `After` measured on comparable dimensions (e.g., both in minutes or error rate)?
- **ADR / STAR**:
  - [ ] Does the ADR explicitly record negative consequences and trade-offs?
  - [ ] Does STAR clearly separate constraints (`Task`) from executed steps (`Action`)?
