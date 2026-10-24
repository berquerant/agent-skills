# Global AGENTS.md

Operational guidelines and core protocols for AI agents operating across projects.
Unless project-specific instructions dictate otherwise, agents must adhere to the following principles and refer to the specialized skills by name for detailed procedures:

---

## 1. Output Presentation

- **Raw String Presentation for Paths and URLs**:
  When communicating file paths, workspace targets, or web URLs to the user, output them as plain, unadorned raw strings without Markdown hyperlinks (e.g., `[label](url)`) or inline code backticks (e.g., `` `path` ``). Follow the **`output-presentation`** skill for formatting and verification details.

---

## 2. Command Execution & Scripting

- **Read Snippet Skills Before Command Execution**:
  Before running non-trivial shell commands, AI agents must read and follow the **`command-snippet`** and **`snippet-review`** skills to save, parameterize, audit, and explain immutable snippet files instead of executing inline scripts.

---

## 3. Investigation & Fact Verification

- **Evidence-Backed Inquiry**:
  Do not act on speculation, unverified assumptions, or unchecked premises during root-cause analysis, bug diagnosis, or codebase exploration. Follow the **`evidentiary-inquiry`** skill to scope sources, distinguish verifiable facts from deductions, and avoid hallucinated diagnoses.

---

## 4. Task Planning & User Approval Gates

- **Explore, Plan, Then Execute**:
  Before modifying code or configuration on non-trivial tasks, explore first, construct a step-by-step plan with measurable verification criteria, and secure explicit user approval before execution. Refer to the **`explore-plan-execute`** skill (or **`scrutinize-plan-execute`** for high-risk or ambiguous tasks).

---

## 5. Commits, Pull Requests, & Discussions

- **Structured Commit & PR Messages**:
  When crafting git commit messages and PR/MR titles or summaries, adhere to repository conventions and Conventional Commits by following the **`change-message`** skill.
- **Constructive Review Discussions**:
  When initiating or replying to PR/MR review comments and resolution threads, follow the **`review-discussion`** skill to maintain actionable, constructive dialogue and explicit resolution ownership.
