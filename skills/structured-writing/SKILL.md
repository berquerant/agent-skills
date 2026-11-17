---
name: structured-writing
description: >-
  Use this skill when drafting new documents, proposals, RFCs, ADRs, technical
  guides, reports, or persuasive messages from scratch. Guides document-type
  classification (Diataxis), framework selection (Pyramid, PREP/BLUF, New
  PASONA, B-C-A/BAB, ADR, STAR), outline construction, and logical self-checks.
license: MIT
---

# Structured Writing

Draft clear, persuasive, and logically sound technical documents, proposals, and communications from scratch by matching the document genre and prose structure to the reader's intent.

For framework definitions, templates, technical examples, and anti-patterns, refer to [frameworks.md](references/frameworks.md) (references/frameworks.md).

---

## Step 1: Identify Purpose, Audience, and Document Type

Clarify what the text must achieve before choosing a structure:

1. **Audience & Desired Outcome**:
   - Identify who will read the text (decision-maker, peer engineer, operator, end-user) and what they should do or understand after reading (approve, align, execute, learn, or look up facts).
2. **Format & Length**:
   - Determine the medium (chat update, issue/PR comment, RFC, ADR, guide, postmortem, release note).
3. **Technical Documentation Genre (When Writing User or Developer Docs)**:
   - Apply **Diataxis** (Daniele Procida) to avoid mixing incompatible documentation modes in a single page:
     - **Tutorial** (learning-oriented): Guided, foolproof steps for a newcomer.
     - **How-to Guide** (task-oriented): Practical steps to solve a specific real-world problem.
     - **Reference** (information-oriented): Concise, objective specifications (APIs, CLI flags, schemas).
     - **Explanation** (understanding-oriented): Context, design rationale, and architecture trade-offs.

Verify: Target audience, desired outcome, medium, and (for technical docs) the Diataxis genre are explicitly defined.

---

## Step 2: Select the Structural Framework

Choose one primary framework (or pair an introduction pattern with a body framework) using the decision table below and the detailed guidance in [frameworks.md](references/frameworks.md) (references/frameworks.md):

| Primary Goal / Situation | Recommended Framework | Typical Engineering Use Cases |
|---|---|---|
| Deliver a decision, request, or brief argument fast | **BLUF** or **PREP** | Status updates, approval requests, concise tech opinions, Q&A |
| Build consensus and prompt action on a shared pain | **New PASONA** (Affinity) | Tooling adoption pitches, refactoring proposals, persuasive RFCs |
| Show transformation, migration steps, and results | **B-C-A** or **BAB** | Migration reports, performance case studies, release highlights |
| Explain complex topics or multi-part logic systematically | **Pyramid Principle** (SCQA + MECE) | Design docs, root-cause analyses, architectural explanations |
| Record an architectural choice and its trade-offs | **ADR** (Nygard) | Architecture Decision Records (`Context / Decision / Consequences`) |
| Report an incident timeline, investigation, or execution | **STAR** | Postmortem timelines, spike/evaluation reports, project summaries |

Verify: The selected framework aligns with Step 1's audience, outcome, and medium, with any genre mixing resolved.

---

## Step 3: Build the Outline and Draft Content

Construct the document top-down rather than writing linearly from background details:

1. **Honor Repository Conventions First**:
   - If the target repository already defines templates (e.g. RFC, ADR, or issue templates), map the chosen framework's logic into those required sections.
2. **Draft the Skeleton**:
   - Populate each slot of the selected framework using concise bullet points before writing full prose.
   - Ensure every claim is backed by verifiable facts, measurements, or explicit placeholders where data is still needed.
3. **Expand with Concrete Evidence**:
   - Replace abstract statements with concrete metrics, reproducible commands, or minimal code examples.
   - Keep sections focused on the single Diataxis mode chosen in Step 1.

Verify: Every slot in the chosen framework is populated with concrete, evidence-backed points, and no unsupported metrics or claims are fabricated.

---

## Step 4: Structural Self-Check and Hand-Off

Audit the draft against the framework-specific checklist in [frameworks.md](references/frameworks.md) (references/frameworks.md):

1. **Logical Integrity Check**:
   - Confirm that the main conclusion appears up front where expected (BLUF, PREP, Pyramid).
   - Verify that horizontal groupings in Pyramid structures are **MECE** (mutually exclusive, collectively exhaustive) and that `Reason` is not confused with `Example` in PREP.
   - Confirm that persuasive frameworks (New PASONA, BAB) rely on technical rationale rather than artificial urgency.
2. **Sequential Hand-Off**:
   - **When integrating into existing documentation with zero code mutation**: Recommend [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md).
   - **When auditing technical accuracy against code reality and exposure safety**: Recommend [doc-review](../doc-review/SKILL.md) (../doc-review/SKILL.md).
   - **When turning the summary into a commit message or PR description**: Recommend [change-message](../change-message/SKILL.md) (../change-message/SKILL.md).

Verify: All framework-specific checks pass, any missing data placeholders are flagged to the user, and appropriate follow-up skills are recommended.

---

## Guidelines

- **Do not fabricate evidence or metrics to complete a template**: Mark missing measurements or unverified claims as explicit placeholders (e.g., `<latency-p99-after>`) instead of inventing plausible numbers, because fabricated data undermines trust.
- **Keep Diataxis genres separated**: Avoid embedding deep architectural explanations inside step-by-step tutorials or mixing reference tables into persuasive proposals; link across documents instead so each piece stays focused.
- **Adapt persuasive tone to engineering norms**: Use New PASONA's Affinity and Narrowing steps to highlight shared operational pain and concrete cost-of-delay rather than marketing hype, keeping technical proposals credible.
- **Defer existing-doc maintenance to `doc-refactor`**: Use this skill for structuring new drafts or major structural outlines; hand off to [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md) when refining existing documentation where 100% context conservation is required.
