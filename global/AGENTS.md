# Global AGENTS.md

Operational guidelines and core protocols for AI agents operating across projects.
Unless project-specific instructions dictate otherwise, agents must adhere to the following principles and refer to the specialized skills by name for detailed procedures:

---

## 1. Output Presentation

- **Raw String Presentation for Paths and URLs**:
  When communicating file paths, workspace targets, or web URLs to the user, output them as plain, unadorned raw strings without Markdown hyperlinks (e.g., `[label](url)`) or inline code backticks (e.g., `` `path` ``). Follow the **`output-presentation`** skill for formatting and verification details.

---

## 2. Command Execution & Scripting

- **Avoid Long Inline Commands & Audit Snippets**:
  AI agents must not execute long, overly complex, or multi-line commands directly inline in the shell. When commands reach avoidance thresholds (chained operators, high character count, or multiple lines), follow the **`command-snippet`** skill to save, parameterize (via explicit CLI arguments instead of environment variables), and document them in snippet files before execution. Every generated snippet must be audited for intelligibility via **`snippet-review`**, and its purpose, arguments, and side-effects must be explained to the user before running.

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
