---
name: command-snippet
description: >-
  Use this skill when executing complex, chained, or multi-line shell commands
  to convert them into structured, documented, and parameterized snippet files,
  review them for intelligibility via snippet-review, and explain their intent to
  the user before execution instead of executing long inline commands.
---

# Command Snippet

Use this skill to convert long, chained, or complex shell commands into
dedicated, documented, and parameterized snippet files before execution.
Directly running long inline commands reduces readability and makes inspecting,
debugging, and verifying execution difficult for users.

This workflow coordinates [`snippet-review`](../snippet-review/SKILL.md) to audit
the cognitive intelligibility of every generated snippet and mandate explaining its
intent and side-effects to the user before execution.

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

If the command meets any of these criteria, **do not execute it inline**. Halt
inline execution and proceed to create a snippet file with the appropriate extension
(e.g., `.sh`, `.py`, `.js`).

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

Verify: All variable inputs are mapped to explicit positional arguments or
command-line flags, and no inputs rely on ambient environment variables.

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

1. **Intelligibility Audit**:
   - Audit the snippet using [`snippet-review`](../snippet-review/SKILL.md) against its four-tier rubric (Transparent, Moderate, Complex, Opaque).
   - If the snippet is rated **Level 3 (Complex)** or **Level 4 (Opaque)**:
     - Attempt to simplify the logic (inline single-use functions, remove dense regex, linearize control flow, add Why comments).
     - If simplification is not feasible, open the **User Consultation Gate** per [`snippet-review`](../snippet-review/SKILL.md) and present the three options:
       1. Split into smaller, sequential steps
       2. Proceed as-is under explicit user acknowledgment
       3. Abandon the snippet and switch to an alternative method
2. **Mandatory Pre-Execution Notice**:
   - Once the snippet is verified as intelligible (Level 1/2 or user-approved Level 3/4), do NOT casually run the command. Output an explicit, prominent notice block in the conversation immediately before invoking the execution tool:

   ```markdown
   ### 📋 Snippet Pre-Execution Notice

   - **Script Path**: <raw unadorned snippet path>
   - **Invocation**: `<interpreter> <script-path> [arguments...]`
   - **Purpose**: <Clear, concrete explanation of what the script does and why a snippet was necessary>
   - **Parameters & Arguments**:
     - `$1` (or `<flag>`): `<value>` — <what this parameter represents>
   - **Expected Side-Effects & Safety**:
     - **Filesystem**: <files created, edited, deleted, or "Read-only (no filesystem changes)">
     - **Environment / Network**: <network calls, environment changes, or "None">
     - **Reversibility**: <Reversible / Non-destructive / Irreversible>
   ```

   - Informal, single-sentence mentions (e.g., "Running script now to check files.") are strictly prohibited. The structured notice must be presented clearly.

Verify: The snippet's intelligibility is audited, any high complexity is simplified or consulted on, and the mandatory Pre-Execution Notice is output before execution.

---

## Step 6: Save, Execute, and Verify

1. **Save Snippet File**:
   - Write the script to the determined path. Ensure file permissions allow
     execution if running directly, or invoke via the interpreter (e.g., `bash path/to/script.sh`).
2. **Execute with Explicit Arguments**:
   - Run the snippet using explicit arguments in the shell invocation.
3. **Verify Outcome**:
   - Check the command exit code and output against expectations.
   - Stop and inspect if the script fails.

Verify: The snippet executed cleanly and satisfied the intended operation.

---

## Step 7: Post-Execution Lifecycle Management

Based on the reusability decision from Step 2:

- **One-off Snippets**: Clean up temporary files if they are no longer needed,
  or leave them in the scratch directory if the user needs them for inspection.
- **Reusable Snippets**: Confirm that the snippet is cleanly formatted, documented,
  and ready for future invocations.

Verify: Unneeded temporary files are disposed of, and reusable scripts are preserved
in their designated locations.

---

## Guidelines

- **Never run complex one-liners or inline scripts.** Adhere strictly to the
  avoidance thresholds (≥4 operators, ≥120 characters, ≥3 lines, or inline
  interpreter scripts `-c`/`-e` with ≥50 chars, imports, multiple statements, or control flow).
- **Explicit arguments over environment variables.** Never use environment
  variables to pass input to snippets; use explicit CLI arguments instead.
- **Document upfront.** Every snippet must include a clear documentation header
  explaining purpose, arguments, and usage.
- **Audit intelligibility before execution.** Every snippet must undergo review
  via [`snippet-review`](../snippet-review/SKILL.md) regardless of line count.
- **Always output the Pre-Execution Notice.** Never execute snippets silently or
  with casual one-line chat messages. Always output the structured Pre-Execution
  Notice block before invoking the execution tool.
- **Keep snippets focused.** A snippet should accomplish a coherent task, making
  it easy for users to read and understand.
