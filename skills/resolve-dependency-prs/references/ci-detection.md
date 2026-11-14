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

## 2. Project Task Runner Inspection & Task Discovery Protocol

Do NOT guess task or target names (e.g. guessing `ui-lint` when the defined task is `ui:lint`). If the remote CI or repository utilizes a standard task runner, inspect and verify available task definitions locally before execution:

### 2.1 Task Discovery Checklist
- **`Taskfile.yml` / `Taskfile.yaml`**: Run `./task --list` or `task --list` (or inspect `Taskfile.yml` directly) to confirm the exact task name syntax (e.g. colon-separated vs hyphen-separated namespaces).
- **`package.json`**: Inspect `scripts` object via `npm run` (shows available scripts) or `jq '.scripts' package.json` before running `npm run <script>` or `pnpm run <script>`.
- **`Makefile`**: Inspect target rules via `make -qp | awk -F':' '/^[a-zA-Z0-9][^$#\/\t=]*:([^=]|$)/ {split($1,A,/ /);print A[1]}' | sort -u` or search targets directly in `Makefile`.
- **`mise.toml`**: Run `mise tasks` to list configured task names.
- **`justfile`**: Run `just --list` to inspect defined recipes.

### 2.2 Standard Task Runners Reference

| Configuration File | Entrypoint | Discovery Command | Common Targets |
|---|---|---|---|
| `Makefile` | `make` | Search `Makefile` targets | `make test`, `make lint`, `make check` |
| `Taskfile.yml` / `Taskfile.yaml` | `task` | `./task --list` or `task --list` | `task test`, `task lint`, `task <ns>:<task>` |
| `mise.toml` | `mise run` | `mise tasks` | `mise run test`, `mise run lint` |
| `justfile` | `just` | `just --list` | `just test`, `just check` |
| `package.json` | `npm` / `pnpm` / `yarn` / `bun` | `npm run` or read `scripts` | `npm test`, `npm run lint`, `npm run typecheck` |
| `Cargo.toml` | `cargo` | Built-in subcommands | `cargo test`, `cargo clippy` |
| `go.mod` | `go` | Built-in subcommands | `go test ./...`, `golangci-lint run` |

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
   - **Tool Missing / Heavy Containerized Pipeline**: Do **NOT** attempt to install dependencies or force local emulation. Push the branch and delegate verification directly to the remote CI pipeline via single-shot queries (`gh pr checks <id>` or `glab mr view <id>`).

---

## 4. Post-Merge Default Branch CI Inspection

After merging a candidate PR into the default branch, the default branch CI must be identified and verified before proceeding to the next candidate PR or concluding:

### GitHub Actions
1. Confirm the PR has finalized merging (especially if `--auto` was queued), then retrieve the merge commit SHA:
   ```bash
   # Ensure PR is merged, then retrieve the merge commit SHA
   JQ_MERGE='select(.state=="MERGED") | .mergeCommit.oid'
   MERGE_SHA=$(gh pr view <number> --json state,mergeCommit --jq "$JQ_MERGE" 2>/dev/null)
   MERGE_SHA=${MERGE_SHA:-$(git rev-parse origin/<default-branch>)}
   ```
2. Query the workflow run triggered on the default branch for that merge commit:
   ```bash
   gh run list --branch <default-branch> --commit "$MERGE_SHA" --limit 1 --json databaseId,status,conclusion,name
   ```
3. For single-shot status checks:
   ```bash
   gh run view <run-id>
   ```

### GitLab CI
1. Retrieve the latest pipeline on the default branch following the merge commit:
   ```bash
   glab ci status --branch <default-branch>
   ```
2. Check pipeline details via single-shot view:
   ```bash
   glab ci view --branch <default-branch>
   ```
