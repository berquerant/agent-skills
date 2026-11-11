# User Notification Channel Patterns

This reference provides environment-adaptive delivery patterns for the three notification tiers in `user-notification`.

---

## 1. Notification Tier Selection Matrix

| Tier | Trigger Event | Recommended Delivery Mode | Standalone Preceding Step Required? |
|---|---|---|---|
| **Complete** | Task, milestone, or pipeline finished | Passive desktop banner / toast or terminal summary | No |
| **Attention** | User input, choice selection, or gate approval needed | Banner, toast, or dialog prompt | No (may accompany the question turn) |
| **Elevated** | Privileged command, sandbox bypass, or system-level approval about to be requested | Prominent alert / notification | **Yes — must complete in a standalone step before the privileged tool call** |

---

## 2. Environment Channel Fallbacks

Evaluate channels in order of fidelity without assuming a fixed OS or proprietary tool:

1. **Dedicated Notification Tool or MCP Server** (Preferred when registered):
   - Invoke the environment's registered notification or alert tool with a concise title (`<project>: <status>`) and a 1-sentence actionable message.
2. **Native Desktop Notification Utility** (When running in a local desktop session without a dedicated tool):
   - Dispatch a non-blocking system toast or banner using the host OS's standard notification utility.
3. **Terminal Bell / Prominent Banner Fallback** (Headless, SSH, or container environments):
   - Emit an audible terminal bell (`\a`) and render a clearly delineated Markdown alert block so the user notices the gate or completion state.
