# Snippet Intelligibility & Complexity Rubric

Evaluation criteria and scoring guidelines for measuring how easily a user or reviewer can understand the intent, control flow, and safety of an execution snippet on the fly.

---

## 1. Core Principles

Execution snippets are created dynamically during agent workflows to perform operations that exceed inline command thresholds. Unlike permanent application codebase files, snippets are consumed and verified immediately in the conversation:
- **Zero Hidden Surprises**: A reviewer should be able to look at the snippet and predict exactly what it will touch, create, or modify.
- **Cognitive Economy**: Simple tasks must not be wrapped in convoluted control flow, unnecessary abstractions, or obscure syntax.
- **Self-Documenting Intent**: If non-trivial logic is unavoidable, it must be accompanied by explicit "Why" comments explaining each block.
- **Boundary & Persistence Safety**: Verify the snippet does not write to protected persistence points (`.git/config`, hooks, shell rc, IDE/agent configs) or secret files (`.env*`, `*.pem`, `*.key`), invoke external app launchers/IPC, or transmit raw credentials over plaintext channels per [persistence-and-escape-vectors.md](../../sandbox-review/references/persistence-and-escape-vectors.md) (../../sandbox-review/references/persistence-and-escape-vectors.md). Any such pattern escalates to Level 4 (Opaque / Critical) until justified or removed.

---

## 2. Evaluation Dimensions (5 Metrics)

Evaluate candidate snippets across five primary dimensions:

### Metric 1: Code Scale & Function Granularity
- **Line Count (SLOC)**: Total non-comment, non-empty code lines.
  - ≤ 10 lines: Low complexity (Level 1).
  - 11–20 lines: Moderate complexity (Level 2).
  - 21–25 lines: High complexity (Level 3).
  - ≥ 26 lines: Critical complexity (Level 4); consider whether the snippet should be broken down.
- **Function / Method Count**:
  - 0 functions (linear flow): Lowest cognitive overhead (Level 1).
  - 1 function (≤ 15 lines): Moderate complexity (Level 2); acceptable if modularizing repeated logic.
  - 2 functions (or any function > 15 lines): Complex (Level 3).
  - ≥ 3 functions: Opaque / Critical (Level 4); resembles a full standalone program rather than a task snippet.
- **Function Length**:
  - Individual functions exceeding 15 lines add significant cognitive load.

### Metric 2: Signature & Interface Complexity
- **Input Parameters**:
  - Explicit positional parameters (`$1`, `$2`): Clear and transparent.
  - Excessive parameters (≥ 4 arguments): Hard to track from the invocation command.
  - Implicit inputs: Parsing untyped environment variables, reading from unexpected parent shell state, or ambient files.
- **Return Values / Outputs**:
  - Clear stdout/stderr streams vs. complex serialized string packing or undocumented exit code conventions.

### Metric 3: Control Flow & Branching (Cyclomatic Complexity)
- **Branching Density**:
  - Number of `if`, `elif`, `else`, `case`, loops (`for`, `while`), and boolean operators (`&&`, `||`).
  - Total decision points ≤ 2: Low.
  - Total decision points 3–5: Moderate.
  - Total decision points ≥ 6: High.
- **Nesting Depth**:
  - Nesting depth ≥ 3 levels (e.g., an `if` inside a loop inside another `if`) makes verifying edge cases mentally exhausting on the fly.

### Metric 4: Asynchrony, Concurrency & Background Execution
- **Background Jobs**: Spawning processes with `&`, `coproc`, or subshell forks `(...)`.
- **Parallelism / Async**: Using `xargs -P`, `parallel`, async/await, background wait loops.
- **Signals & Traps**: Custom trap handlers (`trap '...' EXIT INT TERM`).
- *Impact*: Any asynchronous or parallel execution inherently obscures execution order and error propagation, elevating complexity.

### Metric 5: Language-Specific & Shell Idiosyncrasies
- **Inline Interpreter Scripts (`-c`, `-e`, inline code strings)**:
  Executing interpreter code inline (e.g., `python -c "..."`, `node -e "..."`, `bash -c "..."`, `ruby -e "..."`) when the code payload is ≥ 50 characters, contains module imports, multiple statements, or loops/conditionals. These obscure code logic inside string quotes and must be extracted into a dedicated script file.
- **Obscure Parameter Expansion**: e.g., `${var//pat/repl}`, `${var#*:}`, `${!ref*}` without explanation.
- **Deep Pipelines**: Chaining 3 or more pipes (`cmd1 | cmd2 | cmd3 | cmd4`).
- **Dynamic Evaluation**: `eval`, `source` / `.`, dynamically constructed code strings (inherently elevates snippet to Level 3 / Complex due to quoting bugs and injection hazards; always replace with array parameterization `"${cmd[@]}"` or clean subshell execution).
- **Process Substitution & Heredocs**: `<(...)`, `>(...)`, multi-line nested heredocs.
- **Uncommented Regex / Script Engines**: Dense, unannotated `sed`, `awk`, or `perl` one-liners embedded inside shell scripts.

---

## 3. Four-Tier Intelligibility Rubric

Assign the snippet to one of four levels based on the dimensions above:

| Level | Rating | Definition | Typical Characteristics | Required Action |
|---|---|---|---|---|
| **Level 1** | **Transparent** | Immediately obvious to any reviewer at a glance. | Linear command flow, ≤ 10 lines, no custom functions, ≤ 1 simple pipe, clear inputs/outputs. | Explain intent, inputs, and side-effects concisely to the user, then proceed to execute. |
| **Level 2** | **Moderate** | Clear and intelligible with minimal explanation. | 11–20 lines, 1 helper function (≤ 15 lines) or simple loop, 1–2 straightforward branches, explicit comments. | Ensure all blocks have Why-first comments; explain intent and arguments, then proceed. |
| **Level 3** | **Complex** | Hard to parse quickly; carries cognitive debt or subtle failure modes. | 21–25 lines, 2 functions (or function > 15 lines), 2 pipes (3 chained commands), nested branches (depth 2), 3–5 branches, subshells, or ≥ 4 positional parameters. | **Mandatory Simplification Loop**. Attempt to refactor. If impossible, open User Consultation Gate. |
| **Level 4** | **Opaque / Critical** | Extremely difficult to verify mentally on the fly; high risk of unintended side-effects. | ≥ 26 lines, ≥ 3 functions, deep pipelines (≥ 3 pipes), ≥ 6 branches, `eval`, background concurrency, nesting depth ≥ 3, multi-responsibility logic, or writes to protected persistence/secret paths or plaintext credential leaks. | **Halt Execution**. Attempt aggressive decomposition. If task inherently requires this, consult user on alternatives. |

---

## 4. Simplification Strategies (Remediation)

When a snippet is evaluated as Level 3 (Complex) or Level 4 (Opaque), apply the following simplification patterns before asking the user:

- **Rule 0: Snippet Immutability**: Never edit an existing snippet file in place. Any simplification or remediation must be written to a **new versioned file** (e.g. `script_v2.sh`) to prevent TOCTOU approval bypass. Any prior rating or user notice for the previous file is void.
1. **Linearize & Inline**:
   - Replace single-use helper functions with sequential in-line steps.
2. **Replace Dense One-Liners with Explicit Commands**:
   - Replace cryptic `awk`/`sed` scripts with explicit tool invocations (e.g. `grep`, `cut`, `tr`, `jq`) or add clear comment lines above each pattern explaining the regex.
3. **Add Why-First Explanations**:
   - Add a comment above each conditional check or data transformation explaining *why* the condition exists and what case it guards against.
4. **Remove Ambient State**:
   - Replace implicit environment variable dependencies with explicit script arguments.
5. **Re-evaluate**:
   - Score the simplified snippet again. If reduced to Level 1 or Level 2, proceed to pre-execution user explanation.

---

## 5. User Consultation Gate Template (When Simplification Fails)

If the snippet cannot be simplified to Level 1 or 2 because the task inherently requires complex multi-step orchestration, present the following structured consultation to the user before executing:

```markdown
### Snippet Intelligibility Consultation

**Snippet Path / Command**: [<script-name>](<accessible-target>) (<raw-target>)
**Intelligibility Rating**: Level 3 (Complex) / Level 4 (Opaque)

#### Why is this snippet complex?
- **Factors**: <List primary complexity drivers: e.g., nested loops, multi-tool pipeline, parallel background jobs>
- **Purpose**: <What the snippet achieves and why a simpler single command is insufficient>
- **Potential Risks / Side-effects**: <Files modified, network endpoints accessed, potential failure modes>

#### How would you like to proceed?
1. **Split into Smaller Steps (Recommended)**: Decompose this snippet into 2–3 smaller, independently verifiable steps/snippets executed sequentially.
2. **Proceed as-is**: Execute this snippet as written, acknowledging the logic and risks described above.
3. **Alternative Approach**: Abandon this snippet and switch to another method (e.g. standard CLI commands, interactive tool invocations, or manual user action).
```
