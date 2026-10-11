# Thorough Audit, Holistic Review, and Verification Protocol

Operational guide for the `thorough` execution mode in [`resolve-dependency-prs`](../SKILL.md). Extends routine triage with in-depth micro/macro code reviews and exhaustive verification.

---

## 1. Micro Code-Review (Per-Candidate Inspection)

In `thorough` mode, each eligible candidate PR/MR undergoes a detailed code review adhering to criteria shared with [`code-review`](../../code-review/SKILL.md):

### 1.1 Changelog & Breaking Change Audit
- **Breaking API Changes**: Inspect upstream changelogs, release notes, and migration guides to identify removed, renamed, or modified functions/interfaces used by the repository.
- **Deprecations**: Identify whether the new version deprecates methods currently used in the codebase.
- **Behavioral Shifts**: Look for default behavior changes, altered error types, or environment requirement bumps (e.g. required Node.js, Go, or Python runtime versions).

### 1.2 Supply Chain & Script Safety
- **Lifecycle Scripts**: Verify if new packages introduce `preinstall`, `install`, or `postinstall` hook scripts.
- **Maintainer / Registry Anomalies**: Confirm the publisher identity and download registry against known official registries.
- **Lockfile Hygiene**: Ensure no extraneous dependencies or unexpected package replacements were injected into the lockfile.

---

## 2. Macro Holistic Review (All Candidates Combined)

Rather than treating candidates as isolated updates, evaluate the candidates collectively to optimize the overall resolution path.

### 2.1 Dependency Graph & Transitive Conflicts
- **Peer Dependency Alignment**: Check if updating one candidate causes peer dependency warnings or mismatches with other open PRs.
- **Transitive Duplication**: Identify whether updating PR A and PR B simultaneously causes duplicated transitive packages in lockfiles.

### 2.2 Co-Upgrade Bundling (Consolidation)
Propose bundling updates into a single coordinated PR/branch when candidates are tightly coupled:
- **Ecosystem Packages**: e.g., TypeScript ESLint plugins + parser, React + React-DOM, Vitest + Vite.
- **Runtime & Type Definitions**: e.g., `@types/node` alongside Node engine version bumps.
- **Monorepo / Multi-Package Updates**: Packages released from the same upstream repository with matching version tags.

### 2.3 Optimal Execution Ordering
Order updates to minimize rebase cycles and cascading conflicts:
1. **Critical Security Patches**: Patches resolving known high-severity vulnerabilities.
2. **Foundational Tooling / Formatters**: Low-risk build tools that don't alter application runtime.
3. **Core Application Dependencies**: Layered from foundational/shared libraries up to leaf packages.
4. **Major / High-Risk Upgrades**: Evaluated last after low-risk changes are settled.

### 2.4 Hold / Deprecation Recommendations
Explicitly recommend holding or closing PRs when:
- Upstream has known regressions or unresolved issues reported for that release.
- Upstream introduces breaking changes that require architectural refactoring better suited for a dedicated feature branch.
- The dependency is deprecated or superseded by an alternative.

---

## 3. Exhaustive Pre-Merge Verification

In `thorough` mode, passing the default CI test suite is considered the baseline requirement. Execute all applicable local and environmental checks before approving merge.

### 3.1 Type & Static Analysis Checks
- **TypeScript**: `npx tsc --noEmit`
- **Go**: `go vet ./...`
- **Rust**: `cargo check --all-targets`
- **Python**: `mypy .` or `pyright`

### 3.2 Vulnerability & Audit Scans
Run ecosystem-specific security scanning tools:
- **Node.js**: `npm audit` / `pnpm audit` / `yarn audit`
- **Go**: `govulncheck ./...`
- **Rust**: `cargo audit`
- **Python**: `pip-audit` / `safety check`
- **Ruby**: `bundle exec bundler-audit check --update`

### 3.3 Lockfile & Dependency Tree Integrity
Confirm clean dependency tree resolution without orphaned or missing links:
- **Node.js**: `npm ls` / `pnpm doctor`
- **Go**: `go mod verify`
- **Rust**: `cargo verify-project`

### 3.4 Clean Build Verification
Verify that the project builds cleanly from scratch without relying on stale build caches:
- Purge local build caches if feasible (e.g. `rm -rf dist build target .next`).
- Run the full build command (`npm run build`, `cargo build --release`, `go build ./...`).
