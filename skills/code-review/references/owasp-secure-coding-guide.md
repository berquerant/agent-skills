# OWASP Secure Coding Practices Guide

This reference provides core secure coding principles and actionable checklists for application developers and code reviewers.
It condenses key guidance from the OWASP Developer Guide (v4.2.0) and OWASP Top 10 Proactive Controls into an offline, self-contained reference.

> **Human Reference Citation**:
> Origin: [OWASP Developer Guide (DevGuide)](https://github.com/OWASP/DevGuide/)
> *Note for AI agents: Per AGENTS.md Rule 6, external URLs are strictly for human citation. Rely solely on the local checklist below.*

---

## 1. Core Secure Coding Checklist

Review code against these 10 core security dimensions:

### 1.1 Input Validation (C3)
- [ ] **Server-Side Enforcement**: All input validation is enforced on the server/backend, never trusting client-side checks alone.
- [ ] **Allow-list Approach**: Validate data against an allow-list of permitted characters, types, length ranges, and formats (e.g. strict regex, enums) rather than attempting to filter deny-lists.
- [ ] **Structured Validation**: Ensure numeric bounds, date formats, and string lengths are bounded to prevent resource exhaustion or buffer overruns.

### 1.2 Output Encoding & Injection Prevention (C4, C6)
- [ ] **Context-Aware Output Encoding**: Output data rendered into HTML, JavaScript, CSS, XML, or JSON is encoded for that specific context before rendering to prevent Cross-Site Scripting (XSS).
- [ ] **Parameterized Database Queries**: All SQL/NoSQL queries use parameterized queries, prepared statements, or ORM parameter bindings. String concatenation or template interpolation into queries is prohibited.
- [ ] **Command Execution Sanitization**: Avoid passing untrusted input to system shells. When process execution is required, use fixed executable paths and pass arguments as separate array elements, avoiding shell wrappers (`sh -c`).
- [ ] **Path Traversal Protection**: Validate file paths against path traversal sequences (`../`, `..\\`) and resolve absolute canonical paths against an allowed base directory.

### 1.3 Authentication & Password Management (C7)
- [ ] **Strong Password Hashing**: Passwords stored using modern, adaptive, salted hashing functions (Argon2id, bcrypt, PBKDF2 with sufficient iterations).
- [ ] **Constant-Time Comparison**: Secrets, hashes, and authentication tokens are compared using constant-time algorithms to prevent timing attacks.
- [ ] **Rate Limiting & Lockout**: Authentication endpoints protect against brute-force attacks via rate limiting, IP throttling, or CAPTCHA.

### 1.4 Access Control & Authorization (C1)
- [ ] **Fail-Safe Defaults (Deny by Default)**: Access control decisions deny access by default unless explicitly granted.
- [ ] **Object-Level Authorization (Anti-IDOR)**: Verify that the authenticated user possesses permission to access the specific record or resource ID requested, not merely generic endpoint access.
- [ ] **Server-Side Authorization**: Role and permission checks are strictly executed server-side for every single request and state transition.

### 1.5 Cryptographic Practices (C2, C8)
- [ ] **Standard Cryptographic Libraries**: Use established, peer-reviewed cryptographic libraries; custom cryptographic algorithms or proprietary implementations are strictly forbidden.
- [ ] **Algorithm Deprecation**: Avoid broken or deprecated algorithms (MD5, SHA-1, DES, 3DES, RC4, ECB cipher mode). Use AES-GCM, ChaCha20-Poly1305, SHA-256/SHA-3, or modern asymmetric standards (Ed25519, RSA >= 2048-bit).
- [ ] **Cryptographically Secure Randomness**: Use CSPRNG (e.g. `crypto/rand` in Go, `secrets` in Python, `crypto.randomBytes` in Node.js) for tokens, nonces, and salts; never use pseudo-random generators (`math/rand`, `Math.random`).

### 1.6 Error Handling & Exception Management (C3, 4.3.11)
- [ ] **No Sensitive Leakage in Errors**: Error responses returned to clients, APIs, or interfaces must never contain stack traces, database schema details, file paths, or credentials.
- [ ] **Fail Securely**: If an exception or error occurs in security-critical logic (authentication, payment, authorization), the system must default to a secure closed state (e.g. abort transaction, deny access).

### 1.7 Data Protection & Memory Safety (C2)
- [ ] **Sensitive Data in Transit & Rest**: Protect sensitive data in transit using TLS 1.2+ and at rest using strong encryption.
- [ ] **Ephemeral Secrets in Memory**: Sensitive credentials and cryptographic keys held in memory should be overwritten/zeroed out when no longer needed where supported by the language.
- [ ] **No Secrets in Source or Version Control**: Ensure zero API keys, tokens, or private certificates are embedded in source files.

### 1.8 Security Logging & Monitoring (C9)
- [ ] **Audit Trail**: Log significant security events: authentication failures, privilege escalations, access control denials, and administrative changes.
- [ ] **Log Sanitization**: Ensure sensitive data (passwords, payment details, session tokens, PII) is redacted or excluded from application log streams.
- [ ] **Log Injection Prevention**: Sanitize or encode newlines (`\n`, `\r`) and control characters in untrusted input before logging to prevent log forging.

### 1.9 Communication & Network Security (4.3.9, SSRF)
- [ ] **SSRF Defense**: For backend services fetching external URLs, validate the destination against a strict domain/IP allow-list. Disallow internal loopback (`127.0.0.1`, `localhost`), link-local (`169.254.169.254`), and private RFC 1918 networks.
- [ ] **Secure Transport Configuration**: Ensure HTTPS is enforced with valid certificate verification (no disabled certificate checks or `InsecureSkipVerify: true`).

### 1.10 Secure by Default & Configuration (C5)
- [ ] **Disable Unused Features**: Turn off debugging endpoints, default credentials, verbose tracing, and unnecessary administrative ports in production configurations.
- [ ] **Security Headers**: Web interfaces must include defense-in-depth headers: `Content-Security-Policy`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Strict-Transport-Security`.

---

## 2. Secure Coding Finding Report Template

When reporting non-compliance during a secure coding audit, format each finding with the following four mandatory elements:

```markdown
### 🔴 [Critical] / 🟡 [Warning] <Short Vulnerability Title>

1. **Where (Location)**: `<relative/path/to/file.ext#Lxx-Lyy>`
2. **Rule & Principle**: `<OWASP Dimension, e.g. 1.2 Output Encoding (C4) / 1.4 Access Control (C1)>`
3. **Why (Risk & Rationale)**:
   <Explain why the current implementation is vulnerable. Detail the attack vector,
   potential impact (e.g. data exfiltration, privilege escalation, remote code execution),
   and why the current assumption is unsafe.>
4. **How (Recommended Remediation)**:
   <Provide a concrete, language-idiomatic code example illustrating the secure fix.
   Highlight parameterized calls, allow-list validations, or safe library usage.>

// Example:
// Before (Vulnerable):
db.Query("SELECT * FROM users WHERE name = '" + userInput + "'")

// After (Secure):
db.Query("SELECT * FROM users WHERE name = ?", userInput)
```
