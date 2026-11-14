---
name: command-snippet
description: >-
  Use this skill when executing complex, chained, or multi-line shell commands,
  normalizing CLI arguments, or constraining operational scope. Converts inline
  commands into documented, parameterized snippet files and audits them via
  snippet-review before execution.
---

# Command Snippet

Use this skill to convert long, chained, or complex shell commands—or commands
requiring approval-friendly argument normalization—into dedicated, documented,
and parameterized snippet files before execution.
Directly running long inline commands reduces readability and makes inspecting,
debugging, and verifying execution difficult for users.

This workflow coordinates [snippet-review](../snippet-review/SKILL.md) (../snippet-review/SKILL.md) to audit
the cognitive intelligibility of every generated snippet and mandate explaining its
intent and side-effects to the user before execution, formatting all script, argument,
and affected file paths per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md).

---

## Step 1: Detect Long, Complex, or Inline Interpreter Commands

Inspect the intended command against the avoidance thresholds:

- **Chained commands**: 4 or more commands chained with operators such as `;`,
  `&&`, `||`, or `|`.
- **Character length**: 120 characters or longer. This count evaluates the entire
  command line, including all flags, options, and quoted payload strings.
- **Line count**: 3 or more lines (multi-line commands).
- **Inline interpreter scripts (`-c`, `-e`, inline code payloads)**:
  Any command invoking a language interpreter or shell (`python`, `python3`,
  `node`, `ruby`, `perl`, `bash -c`, `sh -c`, `awk`, `sed`, etc.) that embeds an
  inline script string matching ANY of the following conditions:
  - **Payload length**: The inline script string itself is **50 characters or longer**.
  - **Multiple statements**: Contains multiple statements (separated by `;` or newlines).
  - **Module imports**: Contains `import ...`, `from ... import`, or `require(...)`.
  - **Control flow**: Uses loops (`for`, `while`), comprehensions, or conditionals (`if`, `elif`, `else`).
  - **Escaped quotes / string nesting**: Requires escaped quotes (`\"`, `\'`) or complex string nesting.
- **Infix variable arguments & approval friction**:
  Commands that embed dynamic context variables in the middle of command lines
  (e.g., `git -C <dir> status`), which disrupt prefix-based approval rules or
  obscure the critical action verb during human review.
- **Trailing target arguments & persistent read approval friction**:
  Commands where the target resource is placed at the end of the command line preceded
  by variable flags and patterns (e.g., `grep [options...] <pattern> <file>`). While
  reading the specific target file is safe once authorized, prefix matching cannot bind
  approval to trailing arguments. Approving `grep` globally grants dangerous blanket
  read access to the entire filesystem, while changing flags (`-i`, `-C 5`) invalidates
  prefix rules and triggers repetitive approval friction. Currying the target file
  into a dedicated search snippet establishes a stable prefix for persistent read
  approval while strictly confining read scope.
- **Overly broad command scope & persistent approval friction**:
  Commands invoking omnipotent CLI tools with wide capability blast radiuses
  (e.g., `gh`, `aws`, `gcloud`, `kubectl`), where approving raw commands imposes
  high cognitive review friction or where continuous/session-long approval requires
  restricting the tool to narrow, benign subcommands or exact targets.

If the command meets any of these criteria, **do not execute it inline**. Halt
inline execution and proceed to create a snippet file with the appropriate extension
(e.g., `.sh`, `.py`, `.js`).

- **Risk-Proportional Lightweight Path for Read-Only Local Inspection**:
  When a command is strictly **read-only, local, and side-effect-free** (e.g., local `find`/`grep`/`awk` inspection pipelines that do not mutate files, access external networks, or require elevated sandbox permissions), prefer splitting it into simple $<120$-character commands first; escalate to full snippet creation and `snippet-review` when the inspection logic must be reused, exceeds simple command splitting, or requires persistent read approval on a specific file without triggering repetitive approval gates.

Verify: You have checked the command against all avoidance thresholds and confirmed
that it requires snippet creation.

---

## Step 2: Determine Scope and Reusability

Evaluate how the snippet will be used:

1. **Reusability Assessment**:
   - Determine whether the snippet will be reused for subsequent tasks in the
     current context or is strictly a one-time execution.
2. **Target File Location**:
   - For one-off snippets: Place them in an appropriate scratch or temporary directory.
   - For reusable snippets: Place them in a designated project script location or
     structured utility path.

Verify: The reusability status is determined and an appropriate file path is
selected.

---

## Step 3: Parameterize via CLI Arguments

Design the snippet's interface:

1. **Accept CLI Arguments**:
   - If the snippet is marked for reuse, structure it to accept command-line
     arguments (e.g., `$1`, `$2`, or command-line flags) rather than hardcoding
     context-specific values.
2. **Avoid Environment Variables for Input**:
   - Do **not** use environment variables to pass arguments or configuration to
     the snippet. Environment variables are implicit and difficult for users to
     track in terminal history. Always prefer explicit CLI arguments.
3. **Safe Command Execution over `eval`**:
   - Do **not** use `eval` or raw string concatenation to assemble and execute dynamic commands inside snippets.
   - Use safe parameterization patterns:
     - Array parameterization: `cmd=("$@")` or `opts=("-v" "--filter" "$1"); "${cmd[@]}" "${opts[@]}"`
     - Clean subshell invocation: `bash -c '...' _ "$@"`
4. **Trailing Variable Arguments & Logical Atomic Responsibility**:
   - Apply approvable snippet patterns per [approvable-patterns.md](references/approvable-patterns.md) (references/approvable-patterns.md):
     - **Right-align dynamic variables**: Curry fixed commands and operations inside the snippet so that context-dependent variables (such as directory paths or hostnames) are exposed exclusively as trailing CLI arguments (`$1`, `$2`).
     - **Curry target files for persistent read approval**: When repeatedly searching or inspecting a specific target file (e.g., `grep` over a log file or codebase target), curry the target file path inside the snippet as a constant and expose search patterns and optional flags as arguments. This allows prefix-based approval systems to grant persistent read approval for that file without granting blanket read access to the rest of the filesystem.
     - **Logical atomic task scope over command wrapping**: A snippet should represent a single logical task stage or atomic transaction rather than wrapping a single isolated command. Deterministic sequences that do not require interactive user decisions between steps (e.g. resolve dependencies -> commit -> verify test/lint) should be encapsulated in one atomic snippet, avoiding excessive script proliferation and approval fatigue. However, never pass arbitrary action verbs or subcommands dynamically into a generalized wrapper (e.g. avoid `git -C "$1" "$2"`).
     - **Self-descriptive snippet naming**: Name snippet files so that reviewers can immediately deduce the exact action, target, and scope from the filename alone (e.g. `rebase_and_test_worktree.sh` or `update_deps_and_verify.sh`). Never use generic, ambiguous, or misleading names (e.g. `resolve_and_verify.sh`, `run.sh`, `helper.sh`) that fail to convey concrete functionality or might be mistaken for existing local project utilities.
     - **Universal directory context encapsulation**: Avoid relying on tool-specific directory flags (such as `git -C` vs tools like `gh` that lack `-C`). Standardize on validating and changing into the target directory inside the script (`cd "${target_dir:?}"`) before executing commands.
     - **Constrain operational scope granularity**: Apply the 4-tier scope granularity model. Select parameter-constrained snippets (Level 2) for recurring read-only queries across multiple items, and fully fixed resource snippets (Level 3) for state-mutating operations, workflows strictly focused on a single target, or persistent read access to a specific file, minimizing human review friction without introducing approval fatigue.
     - **Shell hardening**: Include `set -euo pipefail`, explicit argument checks (`"${1:?error message}"`), input existence validation, and clean execution.

Verify: All variable inputs are mapped to explicit positional arguments or
command-line flags, dynamic arguments are right-aligned to trailing positions without generalized multi-action wrappers, no inputs rely on ambient environment variables, and dynamic command invocations avoid `eval`.

---

## Step 4: Add Documentation Header

Place a structured documentation comment header at the very beginning of the
snippet file. The header must detail:

- **Purpose**: What the script does and why it was created.
- **Arguments**: Each positional or optional argument and its expected format.
- **Usage Examples**: Concrete command lines showing how to run the script.
- **Prerequisites**: Any required dependencies, tools, or permissions.

Example header format:

```bash
#!/usr/bin/env bash
#
# Description:
#   Brief description of the script's purpose.
#
# Arguments:
#   $1 - target_directory (path to the target directory)
#   $2 - filter_pattern   (regex pattern to match)
#
# Usage:
#   ./script.sh path/to/dir "pattern"
#
# Prerequisites:
#   ripgrep (rg), jq
```

Verify: The snippet file begins with a complete documentation header explaining
purpose, arguments, usage, and prerequisites.

---

## Step 5: Audit Intelligibility & Explain Intent to User

Before executing any snippet, evaluate its cognitive intelligibility and ensure the user understands its operation:

1. **Intelligibility Audit & Immutability**:
   - Audit the snippet using [snippet-review](../snippet-review/SKILL.md) (../snippet-review/SKILL.md) against its four-tier rubric (Transparent, Moderate, Complex, Opaque).
   - If the snippet is rated **Level 3 (Complex)** or **Level 4 (Opaque)**:
     - Attempt to simplify the logic (inline single-use functions, remove dense regex, linearize control flow, add Why comments).
     - **Snippet Immutability Rule**: When simplifying or modifying a snippet, **never edit an existing snippet file in place**. Modifying an existing snippet risks Time-of-Check to Time-of-Use (TOCTOU) discrepancies where code is executed that the user never saw or approved. Always write the simplified code to a **new versioned file** (e.g. `script_v2.sh`).
     - If simplification is not feasible, open the **User Consultation Gate** per [snippet-review](../snippet-review/SKILL.md) (../snippet-review/SKILL.md) and present the three options:
       1. Split into smaller, sequential steps
       2. Proceed as-is under explicit user acknowledgment
       3. Abandon the snippet and switch to an alternative method
2. **Mandatory Pre-Execution Notice**:
   - Once the snippet is verified as intelligible (Level 1/2 or user-approved Level 3/4), do NOT casually run the command. Output an explicit, prominent notice block in the conversation immediately before invoking the execution tool:

   ```markdown
   ### 📋 Snippet Pre-Execution Notice

   - **Script Path**: [<script-name>](<accessible-target>) (<raw-target>)
   - **Invocation**: `<interpreter> <script-path> [arguments...]`
   - **Purpose**: <Clear, concrete explanation of what the script does and why a snippet was necessary>
   - **Parameters & Arguments**:
     - `$1` (or `<flag>`): `<value>` (if a file/directory path, co-present per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md)) — <what this parameter represents>
   - **Expected Side-Effects & Safety**:
     - **Filesystem**: <files created, edited, or deleted, co-presented per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md), or "Read-only (no filesystem changes)">
     - **Environment / Network**: <network calls, environment changes, or "None">
     - **Reversibility**: <Reversible / Non-destructive / Irreversible>
   ```

   - **Output Presentation Compliance**: Format all file and directory paths (including the script path itself, argument target files, and affected filesystem paths) by co-presenting an accessible Markdown link and an unadorned raw string in parentheses per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md) (e.g., `[script.sh](<file:///path/to/script.sh>) (path/to/script.sh)`).
   - Informal, single-sentence mentions (e.g., "Running script now to check files.") are strictly prohibited. The structured notice must be presented clearly.
   - **Approval Invalidation**: If any change or correction is made to the snippet after this notice is presented, the previous approval is automatically void. A new versioned file must be created and a new notice presented.

Verify: The snippet's intelligibility is audited, any high complexity is simplified into a new versioned file or consulted on, and the mandatory Pre-Execution Notice is output before execution with all script and related file paths formatted per `output-presentation`.

---

## Step 6: Save, Execute, and Verify

1. **Save Snippet File as Immutable**:
   - Write the script to the determined path.
   - Once saved, treat the file as **strictly immutable**. Never apply in-place edits. If a fix is needed during verification, create a new file (e.g., `_v2`).
2. **Execute with Pinpointed Script-Path Prefix**:
   - **Pinpointed Script Authorization over Generic Interpreter Permission**:
     Execute the snippet directly as the command binary (e.g., `<script-path> [args...]` or `./<script-name> [args...]`). Invoking generic interpreters directly (`bash <script-path>` or `python3 <script-path>`) risks triggering blanket permission prompts for `bash` or `python3` itself, granting overly broad execution rights across arbitrary shell commands. Executing the script path directly confines approval specifically to that immutable snippet file.
   - **Atomic Single-Step Execution (No Standalone `chmod` Turn)**:
     If execute permissions must be ensured before execution, do **not** run a standalone `chmod +x` tool call in a separate turn. Standalone `chmod` turns waste an entire interaction cycle and create redundant approval friction. Instead, chain permission setup and execution into a single atomic command line:
     ```sh
     chmod +x "<script-path>" && exec "<script-path>" [args...]
     ```
     This confines approval and execution to a single atomic step without separate permission turns.
3. **Verify Outcome**:
   - Check the command exit code and output against expectations.
   - Stop and inspect if the script fails. If failure remediation requires code changes, create a new versioned snippet file (`_v2`) and restart Step 5.

Verify: The snippet executed cleanly and satisfied the intended operation without in-place file modifications or separate standalone `chmod` turns.

---

## Step 7: Post-Execution Lifecycle Management

Based on the reusability decision from Step 2:

- **One-off Snippets**: Clean up temporary files if they are no longer needed,
  or leave them in the scratch directory if the user needs them for inspection.
  When multiple versions were created (`_v1`, `_v2`), remove or deprecate superseded versions (`.deprecated`) to prevent accidental execution.
- **Reusable Snippets**: Confirm that the final snippet version is cleanly formatted, documented,
  and ready for future invocations.

Verify: Unneeded temporary files and superseded snippet versions are disposed of, and reusable scripts are preserved in their designated locations.

---

## Guidelines

- **Extract complex commands into documented snippet files**: Convert commands meeting the avoidance thresholds (≥4 operators, ≥120 characters, ≥3 lines, or non-trivial `-c`/`-e` payloads) into snippet files so users can inspect and approve them cleanly.
- **Execute via script path to confine approval scope**: Execute the script directly as the command binary (or atomic `chmod +x <path> && exec <path> [args...]`) so that approval systems bind permissions to the specific snippet file rather than granting dangerous blanket approval to generic interpreters (`bash` or `python3`). Never run standalone `chmod` commands in a separate turn.
- **Require self-descriptive snippet names**: Give snippets explicit, self-evident names reflecting exact scope and action (e.g. `rebase_and_test_worktree.sh`), avoiding ambiguous or generic names that fail to convey functionality or could be mistaken for pre-existing local repository tools.
- **Pass inputs via explicit CLI arguments**: Expose inputs as positional arguments or flags rather than ambient environment variables so every invocation is self-contained in terminal history.
- **Use array parameterization instead of `eval`**: Construct dynamic command lists with positional arrays (`"${cmd[@]}"`) rather than `eval` to prevent quoting bugs and injection hazards.
- **Include a complete documentation header**: Document purpose, arguments, usage, and prerequisites at the top of every snippet file.
- **Treat saved snippets as write-once (immutable)**: Write any revision, bug fix, or simplification to a new versioned file (e.g., `_v2`) rather than editing in place to prevent TOCTOU approval discrepancies.
- **Audit intelligibility before execution**: Evaluate every snippet with [snippet-review](../snippet-review/SKILL.md) (../snippet-review/SKILL.md) prior to execution.
- **Output the structured Pre-Execution Notice**: Present the `### 📋 Snippet Pre-Execution Notice` block before invoking the execution tool, co-presenting all script and target paths per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md).
- **Encapsulate atomic task stages rather than single command wrapping**: Group deterministic sequences of commands forming a single logical milestone (e.g., resolve -> commit -> verify) into an atomic snippet per [approvable-patterns.md](references/approvable-patterns.md) (references/approvable-patterns.md) to prevent script proliferation and approval fatigue, while keeping dynamic subcommands strictly forbidden.
- **Universal directory context encapsulation**: Validate target paths and enter the target directory via `cd "${target_dir:?}"` inside the script per [approvable-patterns.md](references/approvable-patterns.md) (references/approvable-patterns.md) instead of relying on tool-specific options (`git -C` vs `gh`).
- **Curry target files for persistent read approval**: When repeatedly searching or inspecting a specific target file (e.g., `grep [options] <pattern> <file>`), curry the target file path inside a dedicated snippet per [approvable-patterns.md](references/approvable-patterns.md) (references/approvable-patterns.md) so users can grant persistent read approval for that file via prefix matching without granting blanket read access to the rest of the filesystem.
- **Constrain operational scope**: Restrict broad CLI tools (`gh`, `aws`, `kubectl`) to specific subcommands or exact targets per [approvable-patterns.md](references/approvable-patterns.md) (references/approvable-patterns.md) to reduce review friction.
