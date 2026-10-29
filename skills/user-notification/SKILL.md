---
name: user-notification
description: >-
  Use this skill when notifying the user upon task completion, when user confirmation
  or input is required, or immediately before executing tools or commands that require
  user approval or elevated permissions. Guides the agent to issue timely notifications
  via available environment channels in a standalone preceding step before requesting approval.
---

# User Notification

Use this skill to notify the user whenever:
- A requested task is completed.
- User confirmation, input, or clarification is required to proceed.
- Immediately before invoking any tool or command requiring user approval or elevated permission (such as commands that modify system configurations or bypass sandboxing).

Issuing clear, timely notifications prevents silent idle time, alerts the user to blocking decisions, and ensures intentional review before executing privileged actions.

---

## Prerequisites

Identify the notification mechanisms available in the host execution environment:
- **Dedicated Notification Tool or MCP Server**: Any specialized tool registered to display user-facing dialogs, banners, or system alerts.
- **System Desktop Notification Utility**: Any native command-line utility capable of dispatching non-blocking desktop toast or banner notifications.
- **Audio / Terminal Signal**: An audible control signal (such as an ASCII bell character) or a prominent text banner in the active console.

---

## Step 1: Detect Notification Event and Level

Evaluate the operational context and assign the event to one of three notification levels:

| Level | Trigger Condition | Notification Purpose | Preferred Delivery |
|---|---|---|---|
| **Complete** | A requested task, workflow, or major pipeline has completed. | Alert the user that output is ready for inspection. | Passive banner / toast |
| **Attention** | User input, choice selection, or clarification is required to continue. | Prompt the user to unblock execution. | Banner, toast, or interactive dialog |
| **Elevated** | A tool or command requiring explicit user approval or elevated permission is about to run. | Alert the user immediately before the system approval prompt appears. | Standalone preceding notification / alert |

Verify: The trigger condition and notification level are explicitly determined.

---

## Step 2: Enforce the Preceding Notification Protocol for Elevated Actions

When triggering an **Elevated** level notification or any action that halts execution awaiting user approval:

1. **Strict Step Separation**:
   Do **NOT** batch the notification and the privileged command in the same step or concurrent tool-call execution.
2. **Dispatch Preceding Step**:
   Issue the notification in an independent, preceding step first.
3. **Verify and Proceed**:
   Ensure the notification step successfully finishes before issuing the tool call or command that requests elevated permissions or confirmation.

This separation ensures the user receives an alert before the execution environment enters a blocking state.

Verify: For elevated or approval-required actions, the notification is completed in a standalone preceding step before the privileged action is invoked.

---

## Step 3: Resolve Available Notification Channel and Dispatch

Select the highest-fidelity notification channel accessible in the current environment:

1. **Dedicated Notification Tool Channel**:
   If a specialized notification tool is available in the environment:
   - For **Complete** events: Invoke the tool's non-blocking notification or banner function with a concise title and message.
   - For **Attention** or **Elevated** events: Invoke the tool's dialog, alert, or prompt function when explicit modal attention is appropriate.
2. **System Desktop Utility Channel**:
   If no dedicated notification tool is registered but a system-level desktop notification command is supported by the host, invoke that utility to deliver a desktop alert.
3. **Terminal Signal / Visible Output Fallback**:
   In headless, remote, or restricted environments without graphical display capabilities:
   - Emit an audible terminal signal (such as an ASCII bell character) to catch terminal attention.
   - Output a prominent text banner in the console output.

### Message Formulation Rules
- Keep titles concise (e.g., `Task Completed`, `Approval Required`).
- Focus message copy on the specific user action needed (e.g., `Review test results and approve deployment step`).
- Avoid technical jargon, raw stack traces, or excessive markup within notification body text.

Verify: The notification is dispatched cleanly via the resolved channel with concise, actionable content.

---

## Step 4: Verify Delivery and Resume Primary Workflow

1. Confirm that the dispatch call completed successfully.
2. If the preferred notification channel fails or is unsupported in the current environment, fall back to the next available channel without failing the overall user task.
3. Resume the primary workflow: pause for user confirmation, wait for gate input, or present the completed deliverables.

Verify: Notification delivery or fallback is confirmed, and the subsequent workflow action proceeds smoothly.

---

## Guidelines

- **Preceding step rule for elevated actions.** Always issue the notification in an independent preceding step before invoking tools or commands requiring user approval or elevated permissions. Never combine them into a single step.
- **Prevent notification fatigue.** Do not send notifications for minor intermediate sub-steps, iterative internal command loops, or routine scratch operations. Notify only upon task completion, blocking user requests, or elevated permission gates.
- **Platform and tool agnostic.** Never assume a specific operating system or proprietary tool name. Adapt dynamically to the capabilities provided by the host environment.
- **Concise and actionable.** Keep notification titles and bodies brief, unambiguous, and focused on the immediate user decision.
