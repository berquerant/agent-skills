# Plan Scrutiny & Obstacle Checklist

This reference provides a systematic checklist to detect hidden obstacles,
ambiguous premises, and failure risks during the plan review phase of the
`scrutinize-plan-execute` workflow.

---

## 1. Ambiguous Conditions & Premises

Unstated or unverified assumptions frequently cause plans to fail mid-execution.

- [ ] **Implicit Environment State**: Does the plan assume specific CLI tools,
  runtime versions, or OS dependencies without verifying their presence first?
- [ ] **Permission & Access Assumptions**: Does execution require elevated rights
  (e.g., `sudo`, root, cloud IAM, sandbox bypass) that might be unavailable?
- [ ] **Network & Resource Availability**: Does the plan rely on external APIs,
  registries, or remote servers being reachable and unauthenticated?
- [ ] **Data Format & Schema Assumptions**: Are input/output schemas confirmed,
  or does the plan assume a structure that hasn't been verified against actual data?
- [ ] **Clean Working Directory**: Does the plan assume the repository is clean,
  or are there unstaged changes that could collide with generated artifacts?

---

## 2. Vague or Untestable Verification Criteria

If you cannot define an unambiguous, automated, or deterministic check, the step
is prone to false positives or unresolved errors.

- [ ] **Subjective Success Criteria**: Are criteria phrased subjectively (e.g.
  "should work well", "looks correct") instead of deterministically?
- [ ] **Missing Concrete Verification Commands**: Does each step specify the exact
  command or observation to confirm success (e.g., exit code 0, specific grep
  output, automated test pass)?
- [ ] **Negative / Failure Case Verification**: Does verification test only the
  happy path, or does it confirm that error handling and validation logic work
  as intended?
- [ ] **Silent Failure Detection**: Could a command fail silently (e.g. return code
  0 despite empty output or warning)?

---

## 3. Execution Hazards & Irreversibility

Identify irreversible changes or high-consequence failure modes early.

- [ ] **Destructive Operations**: Does the plan include deleting files, dropping
  tables, purging caches, or overwriting configuration?
- [ ] **Rollback Feasibility**: If step $N$ fails, can the workspace or system
  be restored to the state before step 1? Is there an explicit rollback plan?
- [ ] **Cascade / Blast Radius**: Could modifying one shared module, schema, or
  interface break unrelated callers or consumers?
- [ ] **Rate Limits & Resource Exhaustion**: Could execution trigger API rate
  limits, disk exhaustion, or excessive memory usage?

---

## 4. Missing Steps & Incomplete Workflows

Check for lifecycle phases that are commonly overlooked.

- [ ] **Pre-Flight Checks**: Is there a step to verify prerequisites and baseline
  tests before modifying state?
- [ ] **Migration / Backward Compatibility**: If schemas or APIs change, is there
  a transition path for existing data or callers?
- [ ] **Cleanup & Teardown**: Are temporary files, test containers, or scratch
  branches cleaned up upon completion or failure?
- [ ] **Documentation & Consumer Sync**: Are documentation, configuration samples,
  and type definitions scheduled to be updated alongside code changes?

---

## 5. Standard Remedy Patterns

When an obstacle is identified, apply one of the following remedy patterns:

| Obstacle Type | Recommended Remedy Pattern |
|---|---|
| Unverified premise | Add an explicit discovery/pre-check step to inspect state before acting. |
| Vague success criterion | Replace with a deterministic assertion (e.g. exit code, grep check, test). |
| Destructive operation | Take a backup or snapshot first; switch execution mode to `step-gate`. |
| High blast radius | Isolate changes in a branch or worktree; stage incremental rollouts. |
| Incomplete workflow | Add dedicated pre-flight and teardown steps to the plan. |
