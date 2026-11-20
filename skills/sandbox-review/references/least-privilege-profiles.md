# Least-Privilege Sandbox Policy Profiles

## Why Declarative Profiles Matter

Writing sandbox policies directly in runtime-specific syntax makes security intent hard to review and hard to port across tools or operating systems.
These declarative profiles separate **policy intent ("What")** from **runtime implementation ("How")**, giving reviewers a clear baseline to compare any concrete configuration against.

Always start from **Profile 1 (Everyday Local Offline Baseline)** and escalate to Profile 2 or Profile 3 only when a specific task requires it.

---

## Profile 1: Everyday Local Offline Baseline (Default Strict)

Intended for routine local coding, refactoring, testing, and documentation tasks that do not require network access or external credentials.

### 1.1 Network Boundary Policy
1. **Complete Outbound Block**: Deny all outbound traffic across all domains and ports; keep the destination allowlist empty.
2. **Local Traffic & IPC Block**: Deny listening (`bind`) on host local ports, deny loopback connections, and deny all connections to host UNIX domain sockets and OS-specific RPC services.

### 1.2 Filesystem Boundary Policy
1. **Workspace-Scoped Read Access**: Except for essential read-only OS system paths, deny reads across the entire user home directory (including sibling projects, credential directories, and personal files), then re-allow reads only under the active workspace directory. This also neutralizes implicit write access to default log paths under the user home directory.
2. **Workspace-Scoped Write Access & Secret File Protection**: Restrict writes to the active workspace directory plus essential standard I/O and isolated temporary storage. Within the workspace, explicitly deny writes to environment files (`.env`, `.env.*`) and private key or certificate files (`*.pem`, `*.key`).
3. **Persistence-Point & Config Protection**: Unconditionally deny writes to VCS configuration and hook files (e.g., `.git/config`, `.git/hooks/*`), shell initialization files, IDE settings, and agent configuration files.

### 1.3 Credential & Secret Policy
1. **Block Mode Only**: Do not mount personal credential directories or pass real secret environment variables into the execution space.

### 1.4 Process & OS Execution Boundary Policy
1. **No External Application Launch**: Prohibit launching or controlling applications outside the isolation boundary through OS application launchers or cross-app scripting mechanisms.
2. **No Isolation Relaxations**: Disable nested-execution relaxation modes, host certificate-verification leak paths, creation of new user namespaces inside the sandbox, and pseudo-terminal (PTY) allocation.

---

## Profile 2: Controlled Dependency & VCS Access (Scoped Egress)

Use when the task requires fetching dependencies from package registries or interacting with a remote Git host. Inherits all Profile 1 restrictions with the following narrow additions:

- **Network**: Allowlist only the exact required registry or Git hostnames and HTTPS/SSH ports in strict non-interactive mode. Enforce post-DNS IP validation (blocking loopback, link-local, private, and metadata ranges) and deny-before-allow rules with actionable feedback.
- **Filesystem**: Keep home-directory reads denied; mount only an isolated package cache directory inside the sandbox if needed, while preserving all secret-file and persistence-point write denials.
- **Secrets**: Prefer unauthenticated read-only access; if authentication is required, escalate secret handling to Profile 3's proxy-injection controls.
- **Process & OS**: Retain all Profile 1 process, IPC, and isolation-relaxation blocks.

---

## Profile 3: Authenticated API / Cloud Integration (Zero-Trust Proxy Injection)

Use when the task must call authenticated external APIs or cloud services. Inherits Profile 2 restrictions with the following controls:

- **Secrets**: Enable boundary proxy injection with session-scoped sentinel values inside the execution space. Use structured masking for URLs, JSON/YAML, and JWTs, and recompute HMAC signatures at the boundary proxy.
- **Network**: Bind each injected credential strictly to its TLS-verified destination host. Forbid plaintext HTTP injection. Enable L7 HTTP/HTTPS method/path/header inspection and Host/SNI mismatch rejection; document explicit passthrough exceptions only for mTLS or certificate-pinned endpoints.
- **Audit**: Require structured violation logs and verify that sentinel values—never real secrets—appear in agent stdout, stderr, and logs.

---

## Profile Comparison Matrix

| Control Area | Profile 1 (Offline Default) | Profile 2 (Scoped Egress) | Profile 3 (Proxy-Injected API) |
|---|---|---|---|
| **Outbound Network** | Denied (empty allowlist) | Host/port allowlist + post-DNS IP check | Allowlist + L7 inspection + TLS-bound hosts |
| **Local Ports / IPC** | Denied | Denied | Denied |
| **Filesystem Read** | OS essentials + workspace | OS essentials + workspace (+ isolated cache) | OS essentials + workspace (+ isolated cache) |
| **Filesystem Write** | Workspace + isolated temp | Workspace + isolated temp | Workspace + isolated temp |
| **Persistence & Secret Files** | Always write-denied | Always write-denied | Always write-denied |
| **Secrets** | Block mode (excluded) | Block mode (or Profile 3 proxy) | Sentinel masking + boundary proxy injection |
| **Process / OS Relaxations** | All disabled | All disabled | All disabled |

---

## Deviation & Escalation Rules

When an audited configuration relaxes any rule relative to its target profile, verify all three criteria:
1. **Explicit Justification**: A concrete task requirement explains why the stricter profile cannot satisfy the workflow.
2. **Minimum Scope**: The relaxation is narrowed to the exact host, port, path, or HTTP method/path needed—never a wildcard catch-all.
3. **Compensating Control**: High-risk relaxations (such as authenticated egress) are paired with boundary protections (post-DNS IP check, TLS-only sentinel injection, structured violation logging).
