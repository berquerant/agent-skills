---
name: command-snippet
description: >-
  Use this skill when executing complex, chained, or multi-line shell commands
  to convert them into structured, documented, and parameterized snippet files
  instead of executing long inline commands.
---

# Command Snippet

Use this skill to convert long, chained, or complex shell commands into
dedicated, documented, and parameterized snippet files before execution.
Directly running long inline commands reduces readability and makes inspecting,
debugging, and verifying execution difficult for users.

---

## Step 1: Detect Long or Complex Commands

Inspect the intended command against the avoidance thresholds:

- **Chained commands**: 4 or more commands chained with operators such as `;`,
  `&&`, `||`, or `|`.
- **Character length**: 120 characters or longer.
- **Line count**: 3 or more lines (multi-line commands).

If the command meets any of these criteria, **do not execute it inline**. Halt
inline execution and proceed to create a snippet file.

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

## Step 5: Save, Execute, and Verify

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

## Step 6: Post-Execution Lifecycle Management

Based on the reusability decision from Step 2:

- **One-off Snippets**: Clean up temporary files if they are no longer needed,
  or leave them in the scratch directory if the user needs them for inspection.
- **Reusable Snippets**: Confirm that the snippet is cleanly formatted, documented,
  and ready for future invocations.

Verify: Unneeded temporary files are disposed of, and reusable scripts are preserved
in their designated locations.

---

## Guidelines

- **Never run complex one-liners.** Always adhere to the avoidance thresholds (≥4
  operators, ≥120 characters, ≥3 lines).
- **Explicit arguments over environment variables.** Never use environment
  variables to pass input to snippets; use explicit CLI arguments instead.
- **Document upfront.** Every snippet must include a clear documentation header
  explaining purpose, arguments, and usage.
- **Keep snippets focused.** A snippet should accomplish a coherent task, making
  it easy for users to read and understand.
