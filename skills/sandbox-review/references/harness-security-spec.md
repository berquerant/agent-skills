# Generic Agent Execution Harness Security Specification

## Why This Specification Exists

AI agents that execute shell commands, edit files, and invoke tools operate with broad autonomy inside developer machines and CI runners.
Relying solely on prompt instructions or product-specific defaults leaves gaps when prompt injection, buggy commands, or malicious dependencies attempt data exfiltration, host persistence, or sandbox escape.

This specification defines **platform-agnostic security and operational requirements** for any AI agent execution harness—whether implemented via OS sandboxes, containers, microVMs, or proxy-mediated runtimes.

---

## 1. Core Design Principles

1. **Default-Deny & Least Privilege**
   - Block all outbound network traffic, filesystem writes, inter-process communication (IPC), and privileged OS operations by default.
   - Grant only the minimum permissions explicitly required for the active task.
2. **Zero-Trust Secret Handling**
   - Keep real credentials out of the agent execution space whenever possible.
   - When external authentication is unavoidable, inject real secrets at a boundary proxy rather than exposing raw secret values to agent processes.
3. **Actionable Feedback (Explainable Denials)**
   - When the harness blocks an operation, return structured feedback stating **what** was blocked, **why** it is prohibited, and **which safe alternative** is permitted so the agent can self-correct within policy boundaries.

---

## 2. Control Domain Requirements

### 2.1 Network Boundary Control

All agent network traffic should pass through a controlled boundary proxy to prevent data exfiltration and Server-Side Request Forgery (SSRF).

- [ ] **Destination Host & Port Allowlist**: Block all outbound connections except explicitly allowed hostnames and destination ports. Provide a strict non-interactive mode that immediately denies unlisted destinations without prompting for ad-hoc approval.
- [ ] **Deny-Before-Allow with Actionable Reason**: Evaluate explicit deny rules prior to allow rules, returning an agent-readable reason message (for example, pointing to an approved mirror or protocol) on denial.
- [ ] **Post-DNS IP Validation**: Even for allowlisted hostnames, inspect resolved IP addresses immediately before connecting and reject loopback, link-local, host-interface, cloud metadata service, and designated private IP ranges to prevent DNS rebinding and SSRF.
- [ ] **Local Traffic & IPC Blocking**: Deny binding to local host ports, loopback connections, UNIX domain sockets, and OS-specific Remote Procedure Call (RPC) services by default; permit only explicitly vetted sockets when indispensable.
- [ ] **Layer 7 (L7, HTTP/HTTPS) Inspection & Sanitization**: Where appropriate, terminate TLS at the boundary proxy to enforce per-request method, path, and header policies and strip sensitive response headers (such as cookies). Reject requests with Host-header / Server Name Indication (SNI) mismatches or ambiguous path normalization. Treat destinations requiring mutual TLS (mTLS) or certificate pinning as explicit passthrough exceptions.

### 2.2 Filesystem Boundary Control

Prevent destructive changes outside the workspace, unauthorized reading of personal or cross-project secrets, and persistent host compromise.

- [ ] **Scoped Read Access**: Except for read-only OS system binaries and libraries required for execution, deny broad reads across user home directories, credential stores, and unrelated projects; explicitly re-allow only the active workspace directory.
- [ ] **Scoped Write Access with Deny-Over-Allow**: Limit writes strictly to the active workspace directory and essential isolated temporary/stdio paths, evaluating canonical resolved paths so symlinks cannot escape allowed roots. Within allowed write roots, explicitly deny writes to sensitive files (such as environment files and private keys) with deny rules taking precedence over allow rules.
- [ ] **Unconditional Persistence-Point Protection**: Always block writes to files that can trigger arbitrary code execution on the host or in future sessions—such as shell initialization files, Version Control System (VCS) configuration and hooks, IDE configurations, and agent/skill definition files—regardless of workspace allowlists.

### 2.3 Credential & Secret Protection

Prevent API keys, tokens, and signing secrets from leaking through prompt injection, command output, or workspace writes.

- [ ] **Block Mode vs. Boundary Proxy Injection (Masking)**: Support both a **block mode** (completely hiding credential files and environment variables from the execution space) and a **proxy-injection mode** (exposing only session-scoped dummy sentinel values inside the sandbox and substituting real secrets at the boundary proxy immediately upon transmission to an approved TLS endpoint).
- [ ] **Structured Masking & Boundary HMAC Re-Signing**: For structured credential carriers (connection URLs, JSON/YAML configs, JSON Web Tokens (JWTs)), replace only the secret fields with sentinels while preserving valid syntax so client-side validation succeeds without exposing real values. For Hash-based Message Authentication Code (HMAC)-signed requests (such as cloud APIs), recompute signatures at the boundary proxy using the real key.
- [ ] **Strict TLS-Bound Injection**: Restrict sentinel-to-secret substitution to specific, TLS-verified destination hosts. Prohibit secret injection over plaintext HTTP.

### 2.4 Process & OS Execution Boundary

Close sandbox-escape channels that bypass filesystem or network restrictions.

- [ ] **Block External Application Launch & IPC**: Prevent processes inside the execution space from launching host applications or controlling external apps via OS launchers, GUI bridges, or scripting IPC so all child processes remain inside the same isolation boundary.
- [ ] **Disable Isolation Relaxations & Privilege Reconfiguration**: Disable nested-container relaxation modes, creation of new user namespaces, host certificate-verification side channels, and unnecessary pseudo-terminal (PTY) allocation by default.

---

## 3. Operational & Audit Requirements

- [ ] **Structured Violation Logging & Feedback**: Record every blocked network, filesystem, secret, or process event in a structured log format containing timestamp, process/command, target resource, matched rule, and reason so both operators and agents can diagnose issues accurately.
- [ ] **Known-Noise Suppression**: Support narrow suppression rules scoped to `(command, path)` pairs for deterministic, harmless OS or runtime probes so genuine policy deviations are not buried under alert fatigue.

---

## 4. Domain Compliance Matrix Template

Use this summary table when reporting a harness audit:

| Control Domain | Status (`Pass` / `Partial` / `Fail` / `N/A`) | Highest Severity Gap | Notes |
|---|---|---|---|
| Network Boundary | | | |
| Filesystem Boundary | | | |
| Credential & Secret Protection | | | |
| Process & OS Boundary | | | |
| Operational & Audit Controls | | | |
