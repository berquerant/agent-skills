---
name: scrutinize-plan-execute
description: >-
  Use this skill when the user's task requires rigorous upfront scrutiny, has
  ambiguous conditions or premises, carries high failure risks, or demands
  iterative plan refinement with the user before execution. Scrutinizes draft
  plans for obstacles, consults on remedies, manages file archiving for long
  plans and reports upon user confirmation, and executes safely upon approval.
---

# Scrutinize → Plan → Execute

This skill enforces a rigorous, self-critical planning and execution workflow.
Before executing non-trivial or high-risk tasks, it analyzes the request, drafts
a plan, actively searches for hidden obstacles and ambiguous criteria, consults
with the user on remedies, iterates on the plan through discussion, and executes
only after explicit user approval. It also handles file archiving for extensive
plans and post-execution reports with explicit user path confirmation.

---

## Step 1: Explore and Draft Initial Plan

Investigate the user's request and inspect relevant files, environment context,
and constraints.

1. **Investigate Context**: Read relevant files, configurations, documentation,
   and git history. Do not modify files in this step. When using high-cost
   methods (such as APIs or remote queries), minimize total cost across time,
   local compute, and server load by fetching content once to a local file/cache
   and querying it locally for subsequent analysis.
2. **Draft Plan**: Formulate an initial plan containing:
   - **Objective**: Clear restatement of the goal.
   - **Scope**: What will and will not be changed.
   - **Premises & Dependencies**: Assumptions, external dependencies, tools, or
     prerequisites.
   - **Proposed Steps**: Sequenced actions to accomplish the goal.
   - **Success Criteria**: Observable conditions that define task completion.

Verify: The draft plan outlines the objective, scope, steps, and success
criteria based on contextual exploration.

---

## Step 2: Scrutinize Plan and Identify Obstacles

Perform a self-critical review of the draft plan to uncover potential failure
points and ambiguities before presenting it to the user.

Inspect the plan across the following dimensions (refer to
[references/scrutiny-checklist.md](references/scrutiny-checklist.md) for detailed checklists and remedy patterns):

1. **Ambiguous Conditions & Premises**: Are any assumptions unverified? What
   happens if environment states or inputs differ from expectations?
2. **Vague Verification Criteria**: Is any success criterion subjective or hard
   to measure? How exactly will success or failure be confirmed?
3. **Execution Obstacles & Side Effects**: What could fail during execution?
   Are there potential breaking changes, data loss risks, or roll-back issues?
4. **Missing Steps**: Are pre-checks, migrations, test updates, or cleanup steps
   missing?

For every identified obstacle or ambiguity, draft a concrete **remedy or
mitigation proposal**.

Verify: Every identified obstacle or ambiguous point has a corresponding remedy
or mitigation proposal.

---

## Step 3: Consult and Report Findings to the User

Present the draft plan, the scrutinized findings, and proposed resolutions to
the user in a structured format:

1. **Current Draft Plan**: Objective, scope, proposed steps, and success
   criteria.
2. **Identified Obstacles & Ambiguities**: Specific points that may hinder
   execution or cause uncertainty.
3. **Proposed Remedies**: Actionable options or recommendations to address each
   obstacle.
4. **Questions for Clarification**: Specific decisions or choices requiring user
   input.

Ask the user:
> "Here is the scrutinized plan and the potential obstacles identified. How would
> you like to proceed with the proposed remedies, and are there any adjustments you'd like to make?"

Verify: The plan, identified obstacles, and suggested remedies are clearly
communicated to the user.

---

## Step 4: Iteratively Refine Plan Through Discussion

Engage with the user to resolve ambiguities and refine the plan:

1. Collect the user's feedback, answers, and preferences.
2. Update the plan, premises, and mitigation steps accordingly.
3. Re-evaluate any newly introduced steps or assumptions for fresh obstacles.
4. Continue the discussion loop until the user is satisfied and consensus is
   reached on the final plan.

Verify: All identified ambiguities and obstacles are resolved, and the user has
agreed on the refined plan.

---

## Step 5: Pre-Execution Sign-Off and Plan Archiving

Before starting execution, complete the following gates:

### 1. Plan Archiving Check (if plan is extensive)
Evaluate whether the plan is extensive (e.g. substantial length, detailed
architecture, or multi-phase steps):
- If the plan is long or substantial:
  - Ask the user if they would like to save the plan to a file before execution.
  - **No default path**: Do not assume, suggest, or hardcode a default path.
    Always ask the user directly to specify their desired file path if they wish
    to save it.
  - If the user provides a path, write the plan to that exact file.

### 2. Execution Mode Selection (`standard` vs. `step-gate`)
Determine whether to apply the [`step-gate`](../step-gate/SKILL.md) skill
protocol during execution:

- **Use `step-gate` mode** if the plan involves:
  - Destructive or low-reversibility operations (DB changes, file deletions).
  - High blast-radius refactors or high failure costs.
  - Multi-step dependencies where each step must be verified before proceeding.
  - Explicit user preference for checkpointed, step-by-step approval.
- **Use `standard` mode** if actions are easily reversible and safe to execute
  as a single continuous batch with internal verification checks.

State the chosen execution mode to the user:
> **Execution mode**: `step-gate` / `standard`

### 3. Pacing Guard Proposal (`edit-paced-execution`)
Assess if unexpected investigation delays or trial-and-error loops risk stalling execution between file edits.
If appropriate, recommend [`edit-paced-execution`](../edit-paced-execution/SKILL.md) alongside the plan:
> **Pacing guard**: Recommend [`edit-paced-execution`](../edit-paced-execution/SKILL.md) (threshold: <X minutes / Y non-edit calls>) to halt and consult if file edits are delayed.

### 4. Explicit Execution Permission
Confirm the user gives explicit approval to proceed with execution.

Ask the user:
> "Are you ready to proceed with executing this plan in <standard/step-gate> mode?"

Do **not** execute any modifications until the user explicitly approves.

Verify: The execution mode is decided, pacing guard proposals are stated if applicable, the user has given explicit permission to
execute, and if requested, the plan has been saved to the user-specified file path.

---

## Step 6: Execute the Plan

Execute the approved plan according to the mode selected in Step 5:

### Mode: `step-gate`
Follow the [`step-gate`](../step-gate/SKILL.md) skill protocol:
1. Execute **one step at a time**.
2. Run the defined verification check for that step.
3. Report the result to the user and **request explicit approval** before
   advancing to the next step.
4. If a step fails, stop immediately, diagnose the issue, and consult the user.

### Mode: `standard`
Execute the plan step by step with internal verification checks:
1. Follow the sequenced steps agreed upon in Step 4.
2. Apply verification checks after each step to ensure expected behavior.
3. If an unforeseen error or roadblock arises, stop execution, diagnose the
   cause, and consult the user rather than making unilateral assumptions.

Verify: All steps of the plan are executed and each step's verification check
passes under the selected execution mode.

---

## Step 7: Post-Execution Report and Archiving

Present the results to the user:

1. **Execution Summary**: What was accomplished versus the agreed success
   criteria.
2. **Changes Made**: Files created, modified, or deleted, and commands executed.
3. **Verification Results**: Test outputs, linter checks, or manual inspection
   results confirming correctness.
4. **Deviations (if any)**: Any adjustments made during execution and why.

### Report Archiving Check (if report is extensive)
Evaluate whether the post-execution report is extensive:
- If the report is long or detailed:
  - Ask the user if they would like to save the report to a file.
  - **No default path**: Do not assume or suggest a default path. Ask the user
    directly to specify their desired file path if they want it saved.
  - If the user provides a path, write the report to that exact file.

Verify: The user receives the full completion report, and if requested, the
report is saved to the user-specified file path.

---

## Guidelines

- **Active Obstacle Hunting**: Never assume a plan is complete without
  critically checking for ambiguous assumptions, untestable conditions, and
  hidden edge cases in Step 2.
- **No Unilateral Execution**: Never begin modifying code or state without
  explicit user approval in Step 5.
- **Prefer `step-gate` for High Risk**: If the scrutinized plan carries low
  reversibility, critical dependencies, or destructive operations, actively
  recommend and use `step-gate` execution mode.
- **Never Assume File Paths for Archiving**: When asking the user whether to
  save an extensive plan or report, never recommend or pick a default path.
  Always ask the user for their desired file path.
- **Iterative Alignment**: Do not skip the dialogue in Step 4; align thoroughly
  before committing to execution.
- **Transparent Reporting**: Report all deviations or unexpected findings
  encountered during execution.
- **Cost-Aware Investigation**: When using APIs or expensive operations to
  explore context, minimize total cost across time and resources by fetching
  content once into a local file or cache and referencing it locally.
