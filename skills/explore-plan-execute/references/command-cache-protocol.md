# Command Caching Protocol

Operational protocol for caching high-cost or external communication commands to minimize latency, API limits, and resource consumption across repetitive agent executions.

---

## 1. Idempotence & Cacheability Evaluation

Before executing a high-cost or network-facing command, evaluate its properties to determine whether caching is permissible.

### 1.1 Cacheable Conditions
A command is **cacheable** if and only if it is idempotent, deterministic, and safe under the current execution context:
- **Parameter Completeness**: The result depends solely on the command arguments and flags, without relying on uncaptured runtime or environmental parameters.
- **Finished / Stable State**: The target resource has finished changing (e.g. inspecting completed CI logs via `gh run view <RUN_ID> --log`, inspecting fixed releases/tags, fetching repository metadata at a specific ref).
- **Successful Exit**: Only commands that exit successfully (exit code 0) are saved as cache.

### 1.2 Uncacheable Conditions (Do Not Cache)
A command **must not be cached** if any of the following apply:
- **Time-Varying / Polling Operations**: The agent is monitoring progress, awaiting state transitions, or actively polling (e.g. checking whether an in-progress CI run completed). Caching would cause stale evaluations or infinite loops.
- **Transient / Unintended Failures**: If an execution fails unexpectedly (e.g. rate limit, connection drop, 5xx server error), the result must not be cached because a subsequent retry may produce a different, successful result.
- **State-Mutating Operations**: Commands that create, modify, or delete resources (e.g. `POST`, `PUT`, `DELETE` via `curl`, `gh pr merge`, `gh pr create`, `git push`).

---

## 2. Cache Time-To-Live (TTL)

Cache entries must have an explicit expiration timestamp to prevent stale assumptions:

- **Default TTL**: **5 minutes (300 seconds)** from execution.
- **Dynamic TTL Adjustment**:
  - **Immutable Targets**: Completed CI jobs, tagged releases, or archived run artifacts may extend TTL to **60 minutes or indefinitely for the current session**.
  - **Active Contexts**: In-progress pull request comments, active discussion threads, or mutable repository branches should reduce TTL to **1–2 minutes**, or be bypassed when freshness is critical.

---

## 3. Cache Storage & Naming Specification

### 3.1 Cache Directory
Cache files should be placed in a designated non-tracked cache directory, such as:
- `${XDG_CACHE_HOME:-$HOME/.cache}/agent-skills/cmd-cache/`
- Or a dedicated, git-ignored directory within the workspace (e.g. `.agents/cache/`).

### 3.2 Filename Convention
Cache files must encode scope, identity, command string, and timestamps to ensure isolation and self-describing lifecycles:

```text
<escaped_repo_path>__<escaped_branch>__<escaped_command>__<executed_at>__<expires_at>.cache
```

- **`<escaped_repo_path>`**: Repository relative path or slug where slashes and special characters are escaped (e.g. `github.com_berquerant_agent-skills`).
- **`<escaped_branch>`**: Current git branch name sanitized for filenames (e.g. `main` or `feat_feature-name`).
- **`<escaped_command>`**: The specific command line (or upstream pipeline command) sanitized for filenames (e.g. `gh_run_view_12345`). For excessively long commands, use a readable prefix combined with a deterministic hash (e.g. `gh_run_view_12345_hash_<sha256>`).
- **`<executed_at>`**: Execution timestamp (UNIX epoch seconds or UTC ISO-8601 compact format: `YYYYMMDDTHHMMSSZ`).
- **`<expires_at>`**: Expiration timestamp (UNIX epoch seconds or UTC ISO-8601 compact format).

---

## 4. Execution Workflow

When a candidate command is about to be executed:

```text
                 [Candidate Command Request]
                              │
               1. Is command cacheable & idempotent?
                    ├── [No] ─> Execute directly without cache
                    └── [Yes]
                              │
               2. Does a matching cache file exist?
                    ├── [No] ─> Run command, save output to cache file
                    └── [Yes]
                              │
               3. Is cache valid? (current_time < expires_at)
                    ├── [Expired] ─> Delete stale cache, re-run & create new cache
                    └── [Valid]   ─> Read cached content instead of re-running
```

### Step-by-Step Procedure

1. **Check Existing Cache**:
   - Query the cache directory for files matching `<escaped_repo_path>__<escaped_branch>__<escaped_command>__*`.
2. **Evaluate Expiration**:
   - If a cache exists and `current_time < expires_at`, adopt the cached content.
   - If expired, remove the obsolete cache file and proceed as a cache miss.
3. **Execute & Store (On Cache Miss)**:
   - Run the command.
   - If the command succeeds, write its stdout to `<escaped_repo_path>__<escaped_branch>__<escaped_command>__<now>__<now+ttl>.cache`.
4. **Utilize Output Flexibly**:
   - Do not force redundant piping (e.g. avoid useless `cat` when direct file reading is possible).
   - Use direct file arguments where supported:
     - `grep "PATTERN" "<cache_file>"`
     - `jq '.field' "<cache_file>"`
   - Use stdin / pipes when required by the toolchain:
     - `cat "<cache_file>" | custom_processor`

---

## 5. Cache Lifecycle & Cleanup

- **Branch Deletion / Worktree Teardown**: When a git branch or worktree is deleted or cleaned up, all associated cache files matching `*__<escaped_branch>__*.cache` should be removed.
- **Stale Entry Eviction**: Periodically or upon session conclusion, purge cache files whose `expires_at` has passed.
- **Manual Invalidation**: If the user indicates that remote conditions have changed, immediately clear matching caches and fetch afresh.
