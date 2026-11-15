---
name: agent-skills-setup
description: >-
  Use this skill when initializing operational protocols, environment awareness,
  and cross-cutting guidelines for AI agents at the start of a session or task.
  Replaces static rule pasting by actively adopting core practices across presentation,
  command execution, inquiry, planning, commits, and user notification.
license: MIT
---

# Agent Skills Setup

Use this skill at the beginning of an agent session or when embarking on a new task to bootstrap operational protocols and align behavior with standard repository practices.

Instead of pasting verbose, static rules into system prompts or session preambles, this skill guides the agent to inspect its host environment, activate baseline cross-cutting operational guidelines, and verify operational readiness.

---

## Prerequisites

- Operating within an interactive AI agent session.
- Access to the target workspace and repository tools.

---

## Step 1: Discover Environment and Available Capabilities

Inspect the current execution environment and available tools:

1. **Workspace & Repository Inspection**:
   - Determine the active repository root, working directory, and git status (`git status`, current branch, dirty worktree files).
2. **Tooling & MCP Capability Mapping**:
   - Identify available tools, shell utilities, and registered MCP servers (e.g., the companion `agent-skills-mcp` server when present).
   - Check available notification channels, formatters, and linters supported in the host environment.
3. **Boundary Identification**:
   - Note sandbox restrictions, network permissions, and external service access constraints.

Verify: Workspace boundaries, active repository state, and available toolsets are clearly identified.

---

## Step 2: Adopt Core Operational Protocols

Adopt the six cross-cutting operational protocols for all subsequent actions in the session:

### 1. Output Presentation (Dual Formatting)
- When communicating file paths, workspace targets, or web URLs to the user, co-present a clickable Markdown link followed by an unadorned raw string in parentheses (e.g., `[label](accessible_target) (raw_target)`).
- Ensure links remain viable regardless of directory shifts and that raw strings can be copied directly without enclosing backticks.
- Follow [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md).

### 2. Command Execution & Scripting
- Before running non-trivial shell commands, write, parameterize, audit, and explain immutable snippet files rather than running ad-hoc inline commands.
- Never yield intermediate turns between a pre-execution explanation and the tool invocation; invoke execution atomically within the same turn.
- Follow [command-snippet](../command-snippet/SKILL.md) (../command-snippet/SKILL.md) and [snippet-review](../snippet-review/SKILL.md) (../snippet-review/SKILL.md).

### 3. Investigation & Fact Verification
- Base diagnostics, evaluations, and root-cause analyses on verifiable evidence rather than speculative assumptions.
- Clarify ambiguous premises (architecture, target versions, out-of-scope boundaries) and pin facts to exact versions.
- Follow [evidentiary-inquiry](../evidentiary-inquiry/SKILL.md) (../evidentiary-inquiry/SKILL.md).

### 4. Task Planning & User Approval Gates
- Before mutating code or configurations on non-trivial tasks, explore first, create a step-by-step plan with measurable verification criteria, and secure explicit user approval.
- Follow [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md) (or [scrutinize-plan-execute](../scrutinize-plan-execute/SKILL.md) (../scrutinize-plan-execute/SKILL.md) for high-risk or architecturally ambiguous changes).

### 5. Commits, Pull Requests, & Discussions
- Adhere to repository conventions and Conventional Commits for all commit messages and PR/MR descriptions.
- Maintain constructive, actionable dialogues and clear ownership in review discussion threads.
- Follow [change-message](../change-message/SKILL.md) (../change-message/SKILL.md) and [review-discussion](../review-discussion/SKILL.md) (../review-discussion/SKILL.md).

### 6. User Notifications
- Dispatch timely notifications upon task completion, when waiting on user approval gates, or immediately prior to running privileged or sandbox-elevated commands.
- For elevated commands, issue the notification in a standalone preceding step before requesting user approval.
- Follow [user-notification](../user-notification/SKILL.md) (../user-notification/SKILL.md).

Verify: The six core protocols and their referenced skills are recognized and adopted as active behavioral constraints.

---

## Step 3: Verify Baseline Readiness

Confirm that the agent is properly configured and ready to execute user requests:

1. **Notification Channel Test**: If a user notification mechanism is available, verify that alerts can be dispatched without blocking.
2. **Session Alignment Confirmation**: Confirm that subsequent actions will follow the planning, snippet, and output presentation conventions established in Step 2.
3. **Readiness Summary**: Briefly inform the user that the session environment and operational protocols are initialized.

Verify: The agent has established its baseline protocols and is ready to process user instructions.

---

## Guidelines

- **Zero Inline Prompt Bloat**: Do not re-paste full rule texts into system prompts; keep startup configuration lightweight by referencing this skill.
- **Platform Agnostic**: Apply these protocols across any agent platform (Antigravity, Claude Code, Cursor, Codex, etc.).
- **Atomic Operations**: Always couple explanations with action in the same turn when tools are about to run.
- **Non-Destructive Exploration**: Never execute irreversible modifications without an approved plan.
- **Prioritize Dedicated Tools**: Prefer dedicated MCP tools over generic shell commands or ad-hoc scripts whenever appropriate.
