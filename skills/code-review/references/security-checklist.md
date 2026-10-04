# Security & Public Release Audit Checklist

This reference provides a structured checklist for auditing code, documentation,
and pull requests for security vulnerabilities and public release safety.

---

## 1. Public Release & Exposure Safety

Before code or documentation is published to a public repository or external
environment, verify zero exposure across the following categories:

### 1.1 Secrets & Credentials
- [ ] **API Keys & Tokens**: No hardcoded API keys, OAuth tokens, personal access
  tokens, webhook URLs with embedded secrets, or service account tokens.
- [ ] **Passwords & Passphrases**: No database passwords, encryption passphrases,
  or administrative credentials.
- [ ] **Private Keys & Certificates**: No SSH private keys, TLS/SSL private keys,
  or signing certificates committed to the repository.
- [ ] **Cloud Provider Credentials**: No AWS access keys, GCP service account
  JSON keys, Azure client secrets, or similar vendor credentials.

### 1.2 Internal Infrastructure & Network Details
- [ ] **Internal Hostnames & Domains**: No intranet domains, internal DNS names,
  or company-internal hostnames (e.g. `*.internal`, `*.corp`, `*.local`).
- [ ] **Internal IP Addresses**: No RFC 1918 private IPv4 addresses (10.0.0.0/8,
  172.16.0.0/12, 192.168.0.0/16) or internal IPv6 addresses, unless clearly
  mocked/RFC-standard documentation ranges (e.g. 192.0.2.0/24, 198.51.100.0/24,
  203.0.113.0/24).
- [ ] **Staging & Non-Public Endpoints**: No staging, dev, or pre-production URLs
  exposed in configurations or tests.

### 1.3 Confidentiality & Compliance
- [ ] **Proprietary Algorithms & Unreleased Features**: No unannounced or
  proprietary business logic committed inadvertently.
- [ ] **PII & Customer Data**: No personally identifiable information, test fixtures
  containing real user data, or production database dumps.
- [ ] **Internal References in Comments**: No developer notes mentioning internal-only
  project codenames, confidential initiatives, or employee names without consent.
- [ ] **Licensing & Copyright**: All third-party libraries and snippets comply with
  applicable open-source licenses, with proper notices preserved.

---

## 2. Application Security Vulnerability Checklist (OWASP-aligned)

### 2.1 Injection & Sanitization
- [ ] **SQL / NoSQL Injection**: All database queries use parameterized queries
  or prepared statements; no string concatenation into query strings.
- [ ] **Command Injection**: Shell executions do not concatenate user input directly;
  proper argument arrays and sanitization are enforced.
- [ ] **Path Traversal**: File path manipulations validate against directory escape
  sequences (e.g., `../`, `..\\`) and resolve to allowed directories.
- [ ] **Cross-Site Scripting (XSS)**: User inputs rendered in HTML/templates are
  context-sensitively encoded; raw HTML injection is prohibited.

### 2.2 Authentication & Authorization
- [ ] **Authentication Robustness**: Passwords hashed with strong algorithms (e.g.
  argon2id, bcrypt, scrypt); no timing attack vectors on token comparisons.
- [ ] **Access Control Enforcement**: Permission checks performed server-side for
  every restricted resource/action; no client-side security assumptions.
- [ ] **Session Management**: Session tokens are cryptographically random, transmitted
  over HTTPS only, and have appropriate expiration and invalidation logic.

### 2.3 Network & Transport
- [ ] **Server-Side Request Forgery (SSRF)**: URLs fetched by backend services are
  validated against allowlists; internal network requests and loopback addresses
  (`localhost`, `127.0.0.1`, cloud metadata `169.254.169.254`) are blocked.
- [ ] **Transport Security**: HTTPS and TLS 1.2+ enforced for all external calls.

### 2.4 Data Integrity & Cryptography
- [ ] **Cryptographic Primitives**: Standard, audited cryptographic libraries used;
  no home-grown cryptography or deprecated algorithms (MD5, SHA1, DES).
- [ ] **Safe Deserialization**: Untrusted inputs are not directly passed to unsafe
  object deserializers (e.g., Python `pickle`, YAML `load` without safe loader).
- [ ] **Concurrency & Memory Safety**: Concurrency primitives (mutexes, channels,
  locks) prevent race conditions, deadlocks, and shared-memory corruption.
