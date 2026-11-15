# Automated Checks

Reference documentation for the reproducible, mechanical checks used during skill review.

All checks are implemented in Go within the `agent-skills-mcp` companion module (`skill-lint` subcommand and `lint_skill`/`lint_all` tools).
The standalone script [run-checks.sh](../scripts/run-checks.sh) (../scripts/run-checks.sh) is **DEPRECATED**, maintained only as a frozen legacy shell fallback, and will not receive further updates.

---

## Running the Automated Checks

Prefer using `agent-skills-mcp skill-lint` (or the MCP tools `lint_skill` / `lint_all`):

```sh
# Run checks on a skill directory
agent-skills-mcp skill-lint skills/<target-skill>

# Run checks in strict mode
agent-skills-mcp skill-lint --strict skills/<target-skill>
```

Legacy fallback (deprecated):

```sh
skills/skill-review/scripts/run-checks.sh skills/<target-skill>
```

The script runs read-only inspection commands and prints a structured report covering checks A through K:
- **Hard errors (`[F]`)** increment the `errors` counter and cause the script to exit with status `1`.
- **Convention & model-alignment warnings (`[W]`)** increment the `warnings` counter and are surfaced in the output while exiting with status `0` when `errors` is `0` (or status `1` when invoked with `--strict`, as in `bin/check-skills.sh`).

---

## Checks Summary and Criteria

### A. Inventory, Size & Unreferenced Files (Dimensions 2 & 6)
- **Check**: Lists all files in the skill directory, counts lines in `SKILL.md` and `references/*.md`, and verifies that every auxiliary file (`references/`, `scripts/`, `examples/`, `resources/`) is referenced directly from `SKILL.md`.
- **Criteria**: `SKILL.md` line count under 500 lines `[F]`; 80–250 lines recommended `[W]`. Zero unreferenced (orphan) auxiliary files `[W]`.

### B. Frontmatter `name` (Dimension 3)
- **Check**: Validates regex `^[a-z0-9]+(-[a-z0-9]+)*$`, length bounds, and exact match with the parent directory name.
- **Criteria**: 1–64 characters, valid pattern, and exact directory match `[F]`. Recommended 3–30 characters `[W]`.

### C. Frontmatter `description` & Layer 1 Budget (Dimensions 1 & 3)
- **Check**: Extracts `description`, computes character length, and checks the trigger opener clause.
- **Criteria**: Non-empty, 1–1024 characters `[F]`. Recommended 100–350 characters to preserve Layer 1 startup context budget `[W]`. Starts with `Use this skill when/to/for ...` `[W]`.

### D. Section Structure, Top-Loaded Workflow & Verify Checkpoints (Dimensions 2 & 3)
- **Check**: Prints section headings with line numbers, checks that `## Step 1` starts at or before line 100, verifies that `## Guidelines` is the final top-level section, and confirms that every `## Step N` section contains an explicit `Verify:` checkpoint.
- **Criteria**: Every step ends with an observable `Verify:` checkpoint `[F]`/`[W]`. `## Step 1` begins within the first 100 lines `[W]`. Standard section ordering ending with `## Guidelines` followed `[W]`.

### E. Links, Path Portability & Credential Sanitization (Dimension 2)
- **Check**: Validates internal relative markdown links outside code blocks, scans for hardcoded absolute or machine-specific paths, and scans for high-entropy secret/credential patterns.
- **Criteria**: Zero broken relative links, zero machine-specific paths, and zero committed secret/credential patterns `[F]`. Never commit real names, host-specific paths, internal hostnames, or actual environment configs/credentials even as examples; always use synthetic placeholders or relative paths.

### F. Script Permissions (Dimension 2)
- **Check**: Verifies executable permission (`u+x`) on all files under `scripts/`.
- **Criteria**: Every script in `scripts/` is executable `[F]`/`[W]`.

### G. Sibling References & Coupling (Dimensions 4 & 5)
- **Check**: Counts mentions of sibling skills and lists relative cross-skill links (`../<sibling>/`).
- **Criteria**: Uses loosely coupled sequential hand-offs and shared references `[W]`; zero mandatory nested subroutine calls `[F]`.

### H. Repository Registration (Dimension 5)
- **Check**: Locates the repository root and checks registration counts across `README.md`, `AGENTS.md`, and [intent-matrix.md](../../skill-router/references/intent-matrix.md) (../../skill-router/references/intent-matrix.md) when present.
- **Criteria**: Registered across repository skill tables, architecture diagrams, and routing matrices `[W]`.

### I. Cross-Skill Duplicate Reference Headings (Dimension 5)
- **Check**: Scans `references/*.md` across sibling skills in the same skills root to identify duplicate markdown headings that appear in the target skill's `references/` and at least one sibling skill.
- **Criteria**: Shared checklists and specifications are linked via relative cross-skill paths rather than duplicated `[W]`.

### J. External Hyperlinks & Self-Containment (Dimensions 2 & 5)
- **Check**: Scans markdown files in the target skill for `https?://` URLs.
- **Criteria**: Core procedures and specifications are self-contained locally `[F]`. Any external URLs are strictly human-only citations accompanied by an explicit notice forbidding autonomous agent access and a version/verification note `[W]`.

### K. Inline Shell Command Modularity & Prompt Hygiene (Dimension 3)
- **Check**: Scans fenced shell code blocks (`sh`, `bash`, `zsh`) for non-comment single-line commands reaching or exceeding 120 characters, checks for backtick-wrapped raw targets in co-presented links, and flags manual chain-of-thought phrases (`think step by step`).
- **Criteria**: Inline shell examples remain concise (< 120 characters, < 4 chained operators, < 3 lines) and delegate complex logic to `scripts/`, raw parenthetical targets remain unadorned, and verification relies on `Verify:` contracts rather than manual CoT `[W]`.
