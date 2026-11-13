# Tinbergen's Four Questions for Codebase & Dependency Inquiry

A multi-perspective analytical framework adapted from Niko Tinbergen's *Four Questions* (1963) and updated with Bateson & Laland's (2013) dynamic feedback perspective. Use this framework during exploration to avoid single-axis tunnel vision—such as refactoring code based solely on how it looks today without understanding the historical edge cases or external ecosystem constraints that shaped it.

---

## 1. When to Use

Apply this framework during exploration when a task involves understanding **why** a system, module, or external dependency exists in its current form before deciding whether to modify, keep, or replace it.

### Concrete Scenarios
- **Legacy Code Refactoring & Cleanup**: Before deleting or rewriting complex logic, defensive checks, or apparent workarounds (avoiding *Chesterton's Fence* regressions).
- **External Library, Framework, or Tool Evaluation**: Assessing whether to adopt, upgrade, pin, or replace a dependency by examining both local usage history and upstream ecosystem evolution.
- **Unexplained Behavior or Technical Debt Investigation**: Diagnosing code where "how it works" is clear, but "why it was built this way" or "what requirement it protects" is obscured.
- **Codebase & Architecture Onboarding**: Rapidly building a complete mental model of an unfamiliar repository, subsystem, or build/CI pipeline.

### When Not to Use
- Trivial, localized edits (e.g., typo fixes, formatting, isolated test additions).
- Greenfield boilerplate where no prior implementation history or legacy constraints exist.

---

## 2. The Four-Question Matrix

Tinbergen's framework distinguishes **Proximate** (*how* a behavior/implementation operates and developed locally) from **Ultimate** (*why* it is adaptive/valuable and how its lineage evolved), across **Static** (current state) and **Dynamic** (historical/evolutionary) axes:

| Axis | Static (Current State) | Dynamic (History & Evolution) |
|---|---|---|
| **Proximate** (*How* it works & came to be) | **1. Mechanism (Causation)**<br>How does it work right now? | **2. Ontogeny (Development / Local History)**<br>How did it grow into this shape in this codebase? |
| **Ultimate** (*Why* it exists & its lineage) | **3. Function (Adaptive Value)**<br>What problem/constraint does it solve today? | **4. Phylogeny (Evolution / Ecosystem Lineage)**<br>How has this pattern or upstream tool evolved over time? |

---

## 3. Software Engineering Mapping & Investigation Checklist

### 3.1 Mechanism (Proximate, Static) — *"How does it work right now?"*
Focus on observable runtime, structural, and configuration facts in the current workspace.
- [ ] Trace control flow, data flow, and state transitions in the relevant modules.
- [ ] Inspect configuration files, manifests, and lockfiles to pin exact dependency versions.
- [ ] Identify inputs, outputs, side effects, and integration boundaries.

### 3.2 Ontogeny (Proximate, Dynamic) — *"How did it develop in this codebase?"*
Focus on the local repository's developmental history to uncover why past authors made specific choices.
- [ ] Inspect `git log -S`, `git log -p`, and `git blame` on the target files/lines.
- [ ] Review linked PR/MR descriptions, commit messages, ADRs (Architecture Decision Records), and issue references.
- [ ] Distinguish intentional invariants and bug workarounds from accidental accretion.

### 3.3 Function (Ultimate, Static) — *"What purpose or constraint does it serve today?"*
Focus on the current utility, requirements, and trade-offs that justify the component's existence.
- [ ] Check tests, callers, and documentation to identify what functional or non-functional requirements (performance, security, compatibility, compliance) it satisfies.
- [ ] Evaluate what breaks or degrades if this component or dependency is removed or altered.
- [ ] For external tools/libraries, clarify why this option was chosen over alternatives (trade-offs).

### 3.4 Phylogeny (Ultimate, Dynamic) — *"How has its lineage and ecosystem evolved?"*
Focus on the broader technical lineage and upstream trajectory beyond the local repository.
- [ ] Check upstream release notes, changelogs, deprecation notices, and EOL schedules for external dependencies.
- [ ] Compare against modern language/runtime features or newer standard patterns (e.g., whether a workaround was only needed for an older runtime version).
- [ ] Assess future maintenance trajectory: is the upstream active, superseded, or shifting paradigms?

---

## 4. Dynamic Interdependence (Bateson & Laland Update)

Following Bateson & Laland (2013), do **not** treat the four questions as isolated silos. In software systems, **feedback loops** connect them:
- **Niche Construction**: A past tooling or architectural choice (**Ontogeny**) reshapes team workflows, tests, and downstream callers, creating new constraints that reinforce its **Function**.
- **Co-evolution**: Upstream ecosystem shifts (**Phylogeny**) alter local **Mechanism** requirements (e.g., new APIs or deprecations), which in turn trigger new local **Ontogeny** migrations.

When synthesizing findings, explicitly note where changing one axis will ripple into the others.

---

## 5. Synthesis Template

Summarize findings in a compact table before transitioning to planning:

```markdown
### Four-Question Summary: <Target Module / Dependency / Tool>
- **Mechanism (Current behavior)**: <Observed structure, versions, and execution flow>
- **Ontogeny (Local history)**: <Key commits/PRs/ADRs explaining how it became this way>
- **Function (Current value & constraints)**: <Requirements protected; impact if removed/changed>
- **Phylogeny (Ecosystem trajectory)**: <Upstream status, deprecations, modern alternatives>
- **Feedback / Implications for Plan**: <What must be preserved vs. what is safe to refactor or replace>
```

---

## 6. Complementary Frameworks & References

- Combine with [iceberg-model.md](iceberg-model.md) (iceberg-model.md) when you need to trace a recurring symptom down through system structures and implicit mental models.

> [!NOTE] Human-Only Citations
> The following external references are provided strictly for human citation and background reading. Agents must rely on this local reference file rather than autonomously fetching external URLs.
> - Niko Tinbergen (1963), *"On aims and methods of Ethology"*, *Zeitschrift für Tierpsychologie*, 20(4), 410–433 — `https://doi.org/10.1111/j.1439-0310.1963.tb01161.x`
> - Patrick Bateson & Kevin N. Laland (2013), *"Tinbergen's four questions: an appreciation and an update"*, *Trends in Ecology & Evolution*, 28(12), 712–718 — `https://doi.org/10.1016/j.tree.2013.09.013`
