# Candidate Filter & Eligibility Criteria

Rules for filtering PRs/MRs to target only pure dependency updates or security patches, excluding any PRs that contain manual logic alterations or code edits.

---

## 1. Metadata Pre-filtering

Inspect PR/MR attributes (author, label, title) to quickly identify candidate dependency or security changes:

- **Authors / Bots**:
  - `app/dependabot`, `dependabot[bot]`
  - `renovate[bot]`, `renovate`
  - `snyk-bot`, `security-bot`
  - GitLab Dependency Bot / Security Bot
- **Labels**:
  - `dependencies`, `security`, `dependencies-update`, `cve`
- **Title Conventions**:
  - `Bump <pkg> from <v1> to <v2>`
  - `chore(deps): update ...`, `fix(deps): ...`
  - `Security update ...`, `Update dependency ...`

---

## 2. File-Level Exclusivity Check (Strict Rule)

A candidate PR/MR **MUST ONLY** touch recognized dependency manifest or lock files.

### Allowed File Patterns

| Ecosystem | Manifest / Lockfile Patterns |
|---|---|
| **JavaScript / TypeScript** | `package.json`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `bun.lockb` |
| **Go** | `go.mod`, `go.sum` |
| **Rust** | `Cargo.toml`, `Cargo.lock` |
| **Python** | `requirements*.txt`, `pyproject.toml`, `poetry.lock`, `Pipfile`, `Pipfile.lock` |
| **Ruby** | `Gemfile`, `Gemfile.lock` |
| **PHP** | `composer.json`, `composer.lock` |
| **Java / Kotlin** | `pom.xml`, `build.gradle`, `build.gradle.kts`, `gradle.properties` |
| **.NET** | `*.csproj`, `packages.lock.json`, `Directory.Packages.props` |

### Disqualification Criteria (Exclude Immediately)

If the PR/MR diff touches **any** of the following, disqualify it and leave it for manual / standard code review:
- Application source files (e.g. `*.go`, `*.ts`, `*.js`, `*.py`, `*.rs`, `*.java`, `*.c`, `*.cpp`)
- Test suites or test mocks (e.g. `*_test.go`, `*.test.ts`, `tests/`)
- CI / workflow files (e.g. `.github/workflows/*`, `.gitlab-ci.yml`) unless the update bot exclusively bumped an action version (still recommend manual review)
- Documentation or general configs (`README.md`, `Dockerfile`, Makefile)

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
- For all qualified candidates, proceed to [security-and-reputation-check.md](security-and-reputation-check.md) for context and legitimacy verification before checkout.

