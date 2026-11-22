# Candidate Filter & Eligibility Criteria

Rules for filtering PRs/MRs to target only pure dependency updates or security patches, excluding any PRs that contain manual logic alterations or code edits.

---

## 1. Metadata Pre-filtering

Inspect PR/MR attributes (author, label, title) to quickly identify candidate dependency or security changes:

- **Authors / Bots**:
  - `app/dependabot`, `dependabot[bot]`, `dependabot`
  - `renovate[bot]`, `renovate`
  - `snyk-bot`, `security-bot`, `depfu[bot]`, `pyup-bot`
  - GitLab Dependency Bot / Security Bot, `github-actions[bot]`
- **Labels**:
  - `dependencies`, `security`, `dependencies-update`, `cve`, `deps`, `deps-dev`, `dependency`
- **Title Conventions**:
  - `Bump <pkg> from <v1> to <v2>`
  - `chore(deps): update ...`, `chore(deps-dev): ...`, `build(deps): ...`, `fix(deps): ...`
  - `Security update ...`, `Update dependency ...`, `Lock file maintenance`

---

## 2. File-Level Exclusivity Check (Strict Rule)

A candidate PR/MR **MUST ONLY** touch recognized dependency manifest or lock files.

### Allowed File Patterns

| Ecosystem | Manifest / Lockfile Patterns |
|---|---|
| **JavaScript / TypeScript** | `package.json`, `package-lock.json`, `pnpm-lock.yaml`, `pnpm-workspace.yaml`, `yarn.lock`, `bun.lockb`, `bun.lock`, `.npmrc`, `.yarnrc.yml` |
| **Go** | `go.mod`, `go.sum`, `go.work`, `go.work.sum` |
| **Rust** | `Cargo.toml`, `Cargo.lock` |
| **Python** | `requirements*.txt`, `pyproject.toml`, `poetry.lock`, `Pipfile`, `Pipfile.lock`, `pdm.lock`, `uv.lock`, `setup.cfg`, `setup.py` |
| **Ruby** | `Gemfile`, `Gemfile.lock`, `*.gemspec` |
| **PHP** | `composer.json`, `composer.lock` |
| **Java / Kotlin** | `pom.xml`, `build.gradle`, `build.gradle.kts`, `gradle.properties`, `settings.gradle` |
| **.NET** | `*.csproj`, `*.fsproj`, `*.vbproj`, `packages.lock.json`, `Directory.Packages.props`, `Directory.Build.props`, `nuget.config` |
| **Actions Lock** | `actions.lock`, `.github/workflows/actions.lock` |
| **Swift** | `Package.swift`, `Package.resolved` |
| **Elixir** | `mix.exs`, `mix.lock` |
| **Dart / Flutter** | `pubspec.yaml`, `pubspec.lock` |

### Disqualification Criteria (Exclude Immediately)

If the PR/MR diff touches **any** of the following, disqualify it and leave it for manual / standard code review:
- Application source files across languages (e.g. `*.go`, `*.ts`, `*.tsx`, `*.js`, `*.py`, `*.rs`, `*.java`, `*.kt`, `*.c`, `*.cpp`, `*.cs`, `*.rb`, `*.swift`, `*.dart`, `*.sh`)
- Test suites or test mocks (e.g. `*_test.go`, `*.test.*`, `*.spec.*`, `test_*.py`, `*_test.py`, `tests/`, `test/`, `spec/`)
- CI / workflow files (e.g. `.github/workflows/*`, `.gitlab-ci.yml`, `.circleci/*`) unless the update bot exclusively bumped an action lockfile
- Documentation or general configs (`README.md`, `Dockerfile`, `Makefile`, `compose.yaml`, `mise.toml`)

### Handling Unknown Files and Agent Overrides

When files in a PR diff are not recognized by the default pattern catalog:
1. The tool reports them under `unknown_files` with category `unknown`.
2. The agent inspects the file contents and context directly:
   - If verified as a legitimate custom manifest or lockfile, specify it via `overrides` (`{"<path>": "manifest"}`) or `extra_manifest_patterns`.
   - If verified as harmless metadata, specify `overrides: {"<path>": "ignored"}`.
   - If it contains logic or source modifications, uphold the disqualification (`"source"`).
3. If a non-automated or human PR author made a legitimate dependency bump, the agent can add its number to `allowed_pr_numbers`.

---

## 3. Cost-Aware Batch Filtering

To minimize API consumption and tool latency:
- Fetch candidate metadata and changed file lists in a single batch query and redirect to a local file:
  - **GitHub**:
    ```bash
    gh pr list --state open --json number,title,headRefName,author,labels,files,body > candidate_prs.json
    ```
  - **GitLab**:
    ```bash
    glab api "projects/:id/merge_requests?state=opened" > candidate_mrs.json
    ```
- Filter candidates in memory or against the locally saved snapshot (`candidate_prs.json` / `candidate_mrs.json`) instead of querying file diffs or details one PR at a time via remote CLI commands (e.g. avoid running `gh pr diff <id>` or `gh pr view <id>` in loops).
- For all qualified candidates, proceed to [security-and-reputation-check.md](security-and-reputation-check.md) (security-and-reputation-check.md) for context and legitimacy verification before checkout.

