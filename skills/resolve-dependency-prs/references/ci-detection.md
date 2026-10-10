# Dynamic CI & Verification Detection

Guidelines for dynamically inspecting a repository's existing CI workflows and project configurations to determine verification steps without hardcoding or assuming specific tools.

---

## 1. Remote CI Inspection

Do not guess the CI commands. Inspect the actual configuration declared in the repository:

### GitHub Actions
- Look in `.github/workflows/*.yml` or `.github/workflows/*.yaml`.
- Search for steps executing testing, linting, building, or type-checking:
  - `run: npm test`, `run: go test ./...`, `run: cargo test`, `run: pytest`
  - `run: make test`, `run: task test`
- Identify the essential baseline checks that determine PR pass/fail.

### GitLab CI
- Inspect `.gitlab-ci.yml`.
- Search for job `script:` blocks in stages named `test`, `lint`, `check`, or `build`.

---

## 2. Project Task Runner Inspection

If the remote CI triggers standard project task runners, locate the corresponding local entrypoint:

| Configuration File | Entrypoint | Candidate Checks |
|---|---|---|
| `Makefile` | `make` | `make test`, `make lint`, `make check` |
| `Taskfile.yml` / `Taskfile.yaml` | `task` | `task test`, `task lint` |
| `mise.toml` | `mise run` | `mise run test`, `mise run lint`, `mise run check` |
| `justfile` | `just` | `just test`, `just check` |
| `package.json` | `npm` / `pnpm` / `yarn` / `bun` | `npm test`, `npm run lint`, `npm run typecheck` |
| `Cargo.toml` | `cargo` | `cargo test`, `cargo clippy` |
| `go.mod` | `go` | `go test ./...`, `golangci-lint run` |

---

## 3. Local Feasibility Assessment

Before running detected verification commands locally:

1. **Tool Availability**:
   Check if the required compiler/runtime/CLI is available on the machine:
   ```bash
   command -v <tool>
   ```
2. **Decision Matrix**:
   - **Tool Available & Fast**: Execute locally in the worktree. Fast-fails save round-trip time and CI quota.
   - **Tool Missing / Heavy Containerized Pipeline**: Do **NOT** attempt to install dependencies or force local emulation. Push the branch and delegate verification directly to the remote CI pipeline (`gh pr checks --watch` or `glab mr view`).
