# Systems Thinking Iceberg Model for Root-Cause & Structural Inquiry

A vertical root-cause and intervention-design framework grounded in Systems Thinking (Peter Senge's *levels of explanation*, formalized as the Iceberg Model by Michael Goodman) and paired with Donella Meadows' *Leverage Points*. Use this framework during exploration to look beneath surface symptoms and identify where changes will produce durable impact rather than recurring regressions.

---

## 1. When to Use

Apply this framework during exploration when surface-level fixes are likely to fail or when a problem keeps recurring across releases, modules, or teams.

### Concrete Scenarios
- **Recurring Bugs, Incidents, or Flaky Tests**: Investigating why the same class of failure keeps reappearing despite previous patches.
- **Chronic Technical Debt & Build/CI Bottlenecks**: Understanding why workarounds accumulate and why velocity degrades over time.
- **Evaluating Architectural or Tooling Changes**: Determining whether a proposed fix merely tweaks a parameter (low leverage) or resolves the structural driver (high leverage).
- **External Dependency Friction**: Analyzing repeated upgrade pain, version conflicts, or ecosystem mismatches across a project.

### When Not to Use
- Isolated, one-off defects with an obvious local cause (e.g., a missing null check or typo).
- Straightforward feature additions with well-defined, non-contentious requirements.

---

## 2. The Four Layers of the Iceberg

Move downward from visible symptoms to hidden drivers:

```text
  [1. Events]            <- Visible tip: What just happened? (Reactive)
──────────────────────────── Waterline ────────────────────────────
  [2. Patterns]          <- Trends over time: What keeps happening? (Anticipatory)
  [3. Structures]        <- Underlying system: What produces these patterns? (Design)
  [4. Mental Models]     <- Deepest layer: What assumptions sustain the structure? (Transformative)
```

---

## 3. Software Engineering Mapping & Investigation Checklist

### 3.1 Events (Above the Waterline) — *"What just happened?"*
Discrete, observable incidents or symptoms.
- [ ] Record the exact failure, error log, regression, or friction point observed right now.
- [ ] Avoid stopping at an Event-level patch (e.g., bumping a timeout or retrying a flaky job) without checking whether it is part of a pattern.

### 3.2 Patterns of Behavior — *"What has been happening over time?"*
Temporal trends and recurring trajectories across commits, releases, or incidents.
- [ ] Inspect `git log`, issue history, CI failure trends, or dependency update history for repeated occurrences.
- [ ] Identify cycles or correlations (e.g., "breaks after every minor release of dependency X", "regressions cluster in module Y").

### 3.3 Systemic Structures — *"What architecture, rules, or feedback loops produce this pattern?"*
The technical, organizational, and tooling relationships that constrain behavior.
- [ ] Map module coupling, shared mutable state, missing boundaries/contracts, or CI/review bottlenecks.
- [ ] Look for feedback loops:
  - **Reinforcing loops** (e.g., lack of tests → fear of refactoring → copy-paste workarounds → even harder to test).
  - **Balancing loops** or **delays** (e.g., slow CI feedback causing developers to batch risky changes).
- [ ] Check information flows: are critical signals (deprecation warnings, test coverage, contract mismatches) hidden from the people making changes?

### 3.4 Mental Models — *"What beliefs, assumptions, or paradigms sustain this structure?"*
The implicit premises held by the codebase authors, team, or upstream library maintainers.
- [ ] Surface unspoken assumptions embedded in the design (e.g., "backward compatibility is never broken", "configuration should live in code", "all traffic is single-region").
- [ ] For external tools/libraries, identify the upstream design philosophy and whether it aligns or conflicts with your project's assumptions.

---

## 4. Choosing Intervention Depth (Meadows' Leverage Points)

Pair the Iceberg analysis with Donella Meadows' *Leverage Points* to decide **where** your plan should intervene:

| Leverage Tier | Iceberg Layer | Software Examples |
|---|---|---|
| **Low Leverage** | Events | Tweaking numeric constants, retry counts, timeouts, or one-off hotfixes. |
| **Medium Leverage** | Patterns / Structures | Adding observability/information flows, strengthening automated test gates, decoupling feedback delays. |
| **High Leverage** | Structures / Rules | Changing system boundaries, enforcing invariants via types/linters, redefining dependency contracts or CI rules. |
| **Highest Leverage** | Mental Models | Challenging and updating core architectural assumptions, goals, or technology paradigms. |

> [!TIP] Avoid Pushing Low-Leverage Points Harder
> If an investigation reveals a **Structure** or **Mental Model** mismatch, patching **Events** (tweaking parameters) usually creates new side effects. Explicitly call this out in the plan's **Risks / Trade-offs**.

---

## 5. Combining with Tinbergen's Four Questions

The two frameworks complement each other cleanly:
- **Iceberg Model** provides **vertical depth** (*Events → Patterns → Structures → Mental Models*) to locate high-leverage root causes.
- **[tinbergen-four-questions.md](tinbergen-four-questions.md) (tinbergen-four-questions.md)** provides **four-axis breadth** (*Mechanism, Ontogeny, Function, Phylogeny*) to ensure the **Structures** and **Mental Models** you uncover are grounded in both local history and upstream ecosystem reality.

---

## 6. Synthesis Template

```markdown
### Iceberg Summary: <Target Symptom / Problem Area>
- **Event (Observed symptom)**: <Specific failure, friction, or trigger>
- **Pattern (Observed trend)**: <Recurring frequency, affected releases/commits>
- **Structure (Systemic drivers)**: <Coupling, feedback loops, missing signals, rules/constraints>
- **Mental Model (Implicit assumptions)**: <Unstated beliefs in local design or upstream tooling>
- **Recommended Leverage Point**: <Where to intervene (parameter vs. information flow vs. structure/rule vs. paradigm) and why>
```

---

## 7. Human-Only Citations

> [!NOTE] Human-Only Citations
> The following external references are provided strictly for human citation and background reading. Agents must rely on this local reference file rather than autonomously fetching external URLs.
> - Peter M. Senge (1990), *The Fifth Discipline: The Art & Practice of The Learning Organization* (conceptual origin of "levels of explanation")
> - Michael Goodman, *"Systems Thinking: What, Why, When, Where, and How?"*, *The Systems Thinker* — `https://thesystemsthinker.com/`
> - Donella H. Meadows (1999), *"Leverage Points: Places to Intervene in a System"*, The Sustainability Institute / The Donella Meadows Project — `https://donellameadows.org/`
> - Change Agent Inc. (Japanese practitioner reference), *氷山モデル (The Iceberg Model)* — `https://www.change-agent.jp/systemsthinking/approach/the_iceberg_model.html`
