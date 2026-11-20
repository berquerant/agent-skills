# Persistence Points, Sandbox Escape Vectors, and Violation Triage

## Why Workspace Scoping Alone Is Not Enough

Restricting an agent to write only inside the active workspace still leaves critical host-compromise paths open.
Many files inside a repository or user profile execute automatically when the developer next runs `git`, opens an IDE, launches a shell, or starts another agent session.
Similarly, OS IPC mechanisms, DNS rebinding, and isolation relaxations allow code inside a sandbox to reach outside its intended boundary.

Use this catalog during **Mode 1 (Config & Policy Audit)** to verify defensive coverage and during **Mode 2 (Violation Log Triage)** to classify blocked events safely.

---

## 1. Persistence-Point & Sensitive-File Protection Catalog

Writes to these targets must remain **unconditionally denied**, even when they reside inside an otherwise writable workspace directory.

| Category | Representative Paths / Patterns | Attack Vector / Risk | Required Posture |
|---|---|---|---|
| **VCS Config & Hooks** | `.git/config`, `.git/hooks/*`, `.gitconfig`, `.gitmodules` | Setting `core.fsmonitor`, `core.sshCommand`, or hook scripts triggers arbitrary host code execution on the next `git` command. | Always write-denied |
| **Shell Initialization** | `~/.bashrc`, `~/.zshrc`, `~/.profile`, `~/.zshenv`, `/etc/profile*` | Injecting commands achieves persistent code execution in every future interactive or login shell. | Read- and write-denied |
| **IDE & Editor Configs** | `.vscode/settings.json`, `.vscode/tasks.json`, `.idea/**` | Malicious workspace tasks, terminal profiles, or extension paths execute automatically when opened in an editor. | Always write-denied |
| **Agent & Tooling Definitions** | `AGENTS.md`, `CLAUDE.md`, `.cursor/**`, `.gemini/**`, `.mcp.json`, skill/hook definitions | Modifying instructions, hooks, or MCP server commands hijacks subsequent agent sessions or elevates tool permissions. | Always write-denied (unless explicitly editing them in a dedicated reviewed task) |
| **Secrets & Key Material** | `.env`, `.env.*`, `*.pem`, `*.key`, `*.p12`, `~/.ssh/**`, `~/.aws/**`, `~/.config/**` | Overwriting or reading credentials enables exfiltration, key substitution, or unauthorized external access. | Read-denied (outside workspace) / Always write-denied |

---

## 2. Sandbox Escape & Exfiltration Vector Catalog

Verify that the audited harness mitigates every applicable vector below:

| Vector | Control Domain | Mechanism | Required Mitigation |
|---|---|---|---|
| **DNS Rebinding & SSRF** | Network | An allowlisted domain resolves to `127.0.0.1`, `::1`, link-local (`169.254.0.0/16`), cloud metadata (`169.254.169.254`), host interface IPs, or RFC 1918 private networks. | Validate resolved IPs immediately before connecting; reject private, loopback, link-local, host, and metadata ranges. |
| **Host / SNI Mismatch & Path Ambiguity** | Network (L7) | Mismatched TLS SNI and HTTP `Host` headers or unnormalized path traversal (`..`, encoded slashes) bypass proxy allowlists. | Reject Host/SNI mismatches and ambiguous path normalizations at the boundary proxy. |
| **Local Port, Loopback & Socket IPC** | Network / OS | Connecting to host loopback services, binding local ports, or accessing host UNIX domain sockets / OS RPC daemons bypasses egress controls. | Block local port binding, loopback traffic, UNIX sockets, and OS RPC by default; allow only explicitly vetted sockets. |
| **External App Launch & Scripting IPC** | Process / OS | Invoking OS app launchers or desktop scripting bridges spawns processes outside the sandbox boundary. | Block external launcher binaries, GUI IPC, and cross-app scripting mechanisms so child processes stay confined. |
| **Isolation Relaxations** | Process / OS | Enabling nested-container relaxations, creating new user namespaces, exposing host certificate-verification daemons, or allocating unneeded PTYs widens the kernel/host attack surface. | Disable nested relaxations, new user namespaces, cert-daemon side channels, and unnecessary PTY allocation by default. |
| **Symlink & Path Traversal Escape** | Filesystem | Creating or traversing symlinks or `..` segments inside a writable workspace redirects reads/writes to denied paths (e.g., `~/.ssh`, `.git/config`). | Enforce boundary checks on canonical resolved paths at the OS/VFS layer rather than lexical path prefixes. |
| **Plaintext or Unmasked Secret Leakage** | Secrets | Real credentials in env vars/files or proxy injection over plaintext HTTP leak via stdout, error logs, or network capture. | Exclude secrets (block mode) or inject sentinels only over verified TLS to bound hosts; preserve structure and re-sign HMACs at the boundary. |
| **Interactive Egress Bypass** | Network / Ops | Ad-hoc approval prompts for unlisted destinations invite approval fatigue and social-engineering bypasses under prompt injection. | Enforce strict non-interactive denial for unlisted destinations and return structured, actionable denial reasons. |

---

## 3. Violation Log Triage Rubric (Mode 2)

Classify every unique blocked event into one of three categories before recommending any configuration change:

| Category | Signals & Examples | Recommended Action |
|---|---|---|
| **1. Benign Runtime Noise** | Deterministic, non-sensitive probes performed automatically by OS libraries or language runtimes (e.g., checking locale files, timezone data, or standard library caches that gracefully fall back when denied). | **Do not widen permissions.** Add a narrow `(command, path)` suppression entry to reduce log noise while keeping the denial active. |
| **2. Legitimate Blocked Operation** | Operation directly required by the user's stated task (e.g., fetching a pinned dependency or calling an approved API) that fails the workflow when denied. | **Prefer safe alternatives first** (offline cache, read-only mount, boundary proxy injection). Only if no safe alternative exists, recommend a minimal, documented deviation per [least-privilege-profiles.md](least-privilege-profiles.md) (least-privilege-profiles.md). |
| **3. Escape or Injection Indicator** | Access to persistence points (`.git/config`, hooks, shell rc, agent configs), credential stores (`~/.ssh`, `.env`), cloud metadata (`169.254.169.254`), private/loopback networks, external app launchers, or namespace/PTY escalation. | **Keep strictly blocked.** Escalate as a **Critical** security finding and inspect the triggering command or prompt input for injection or malicious dependency behavior. |
