# Failure Archetypes & Remediation Playbooks for Automated Dependency PRs

This reference catalogs recurring failure patterns encountered when merging automated dependency update PRs (e.g., Dependabot, Renovate) and provides standard, rapid remediation playbooks without trial-and-error guesswork.

---

## 1. Overview of Automated Update Tool Limitations

Automated dependency tools typically execute narrow, isolated manifest edits (e.g., updating a version string in a package manifest and lockfile). They rarely:
- Account for project-wide CI environment changes or runner version pins merged upstream into the default branch.
- Re-run repository-specific code generators, license notice reporters, or schema compilers that produce derived secondary artifacts.
- Update committed snapshot, golden, or regression test baselines whose rendered output shifts following dependency bumps.
- Possess platform API permissions required to modify protected workflow files.

Recognizing these structural limitations allows AI agents to classify CI failures immediately into one of five standard archetypes.

---

## 2. Failure Archetypes & Remediation Playbooks

### Archetype 1: Outdated Base Branch vs. Evolving CI Environment
- **Symptom**: CI fails on a dependency PR branch with build, toolchain, or runtime errors (e.g., unsupported language versions, broken runner environments), but the dependency bump itself is trivial and clean.
- **Root Cause**: The PR branch was branched from an older base commit before upstream fixes, runner version pins, or environment constraints were merged into the default branch.
- **Playbook**:
  1. Check recent commits on the default branch (`git log -n 5 origin/<default-branch>`) for CI or environment adjustments.
  2. Inside an isolated worktree, rebase the PR branch directly onto the latest default branch:
     ```bash
     git fetch origin <default-branch>
     git rebase origin/<default-branch>
     ```
  3. Run local tests. If passing, force-push with lease (`git push --force-with-lease origin <head-branch>`).
  4. **Do not** apply redundant workaround code to the PR branch.

---

### Archetype 2: Derived & Secondary Artifacts Out of Sync
- **Symptom**: CI checks fail on static analysis, licensing scans, code generation verification, or derived documentation checks despite clean code compilation.
- **Root Cause**: The automated tool updated only the package manifest (and lockfile), but the repository requires regenerating secondary artifacts (e.g., third-party notice/license files, compiled protobufs, generated mock interfaces, or bundled distribution assets) that depend on the manifest state.
- **Playbook**:
  1. Inspect the repository's `README.md`, `AGENTS.md`, or task runners (`Makefile`, `Taskfile`, `package.json` scripts) for artifact generation targets (e.g., license reporting, code generators).
  2. In the worktree, execute the appropriate generator command to synchronize secondary artifacts.
  3. Inspect `git diff` to confirm that the changes in the derived files strictly correspond to the updated dependency.
  4. Commit the derived changes alongside the manifest update and push to remote.

---

### Archetype 3: Snapshot & Regression Test Manifest Drift
- **Symptom**: CI fails specifically on snapshot comparison, golden file diff, or template regression tests after updating a dependency or container image tag.
- **Root Cause**: The updated component legitimately alters rendered configuration, formatted outputs, or template manifests. Automated tools cannot autonomously accept or update committed test snapshot baselines.
- **Playbook**:
  1. Consult repository instructions (`README.md`, `AGENTS.md`) for the standard snapshot/golden regeneration command.
  2. In the worktree, run the test update target.
  3. Scrutinize the resulting diff: verify that output shifts are strictly confined to the expected version bump or configuration changes without unintended side-effects.
  4. Commit the updated snapshots with a clear test-baseline update message.
  5. **Strict Serialization Rule**: If multiple open PRs modify shared snapshot baselines, process them strictly one at a time (rebase -> update snapshot -> verify -> merge -> pull default branch -> rebase next PR). Parallel updates on shared snapshot baselines cause severe merge conflicts.

---

### Archetype 4: Base Branch Protection & Ruleset Policy (Fast-Forward Enforcement)
- **Symptom**: Merging via platform CLI (`gh pr merge` or `glab mr merge`) fails with policy rejection errors (e.g., `the base branch policy prohibits the merge`, `non_fast_forward rule violation`, or `branch is out of date`).
- **Root Cause**: The repository enforces a strict fast-forward or linear history ruleset on the default branch. When a preceding PR merges into the default branch, all remaining open PR branches fall behind and become non-fast-forward candidates.
- **Playbook**:
  1. **Do not** bypass branch policies using elevated flags (such as `--admin`).
  2. In the worktree, fetch the latest default branch and rebase the PR branch:
     ```bash
     git fetch origin <default-branch>
     git rebase origin/<default-branch>
     git push --force-with-lease origin <head-branch>
     ```
  3. Once the PR branch tip is a fast-forward descendant of the default branch and CI re-passes, standard merge policies will accept the merge cleanly.

---

### Archetype 5: Platform API Token Scope Restrictions
- **Symptom**: Platform API commands that attempt to update PR branches (e.g., `gh pr update-branch --rebase`) fail with permission or scope errors (e.g., `GraphQL: refusing to allow an OAuth App to create or update workflow .github/workflows/*.yml without 'workflow' scope`).
- **Root Cause**: The platform REST/GraphQL API enforces explicit, elevated token scopes (such as `workflow`) when altering CI configuration files, which standard user or automation tokens may lack.
- **Playbook**:
  1. Avoid relying on API-level branch mutation commands for workflow-affecting branches.
  2. Perform branch operations entirely through standard Git transport (SSH or HTTPS Git credentials) inside a local worktree:
     ```bash
     git fetch origin <default-branch>
     git rebase origin/<default-branch>
     git push --force-with-lease origin <head-branch>
     ```
  3. Standard Git push operations over authenticated Git transport do not trigger GraphQL API token scope barriers.
