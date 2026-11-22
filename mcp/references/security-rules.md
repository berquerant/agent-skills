# Security Guard Rules & Continuous Improvement Guide

This document defines the scope, rule inclusion criteria, and maintenance workflow for the declarative security rules in [`mcp/internal/guard/rules.json`](../internal/guard/rules.json).

---

## 1. Threat Model & Scope Boundary

`internal/guard` provides **defense-in-depth** inside `agent-skills-mcp` to prevent sensitive paths or credentials from being read, cached, written to persistent state, or logged.

- **In Scope**:
  - Blocking filesystem reads/writes targeting well-known credential files, private key extensions, and sensitive system/config directories (`CheckPathAllowed` / `IsSensitivePath`).
  - Blocking or redacting high-confidence secret value formats before writing to cache, checkpoints, snippets, or logs (`CheckNoSecrets` / `ScanTextForSecrets` / `RedactSecretsInString`).
  - Scrubbing sensitive JSON keys during diagnostic logging (`SanitizeForLog` / `IsSensitiveKey`).
- **Out of Scope**:
  - Replacing the host OS or agent execution sandbox (network egress control, OS user isolation, command approval engines).
  - Enforcing subjective semantic naming or arbitrary CLI subcommand allowlists (handled by skill guidance and host approval policies).

---

## 2. Single Source of Truth (`rules.json`)

All guard rules live in [`mcp/internal/guard/rules.json`](../internal/guard/rules.json) and are embedded into the binary via `//go:embed` (`rules.go`). Updating rules does **not** require editing Go matching logic.

| JSON Section | Purpose | Example |
|---|---|---|
| `sensitive_filenames` | Exact lowercase basenames denied from reading, caching, or writing | `.env`, `id_ed25519`, `kubeconfig` |
| `sensitive_extensions` | Lowercase file extensions for keys, certs, and keystores | `.pem`, `.key`, `.p12`, `.kdbx` |
| `sensitive_path_patterns` | Named Go RE2 regexes matched against cleaned slash-normalized paths | `(?i)(^\|/)(\.config/gh/hosts\.yml)$` |
| `sensitive_key_substrings` | Case-insensitive substrings matched against JSON field/key names during log scrubbing | `api_key`, `client_secret`, `database_url` |
| `secret_value_patterns` | Named Go RE2 regexes for high-confidence credential/token formats | AWS Access Key, GitHub Token, Private Key Block |

---

## 3. Rule Inclusion & Quality Criteria

Before adding or modifying an entry in `rules.json`, verify all four criteria:

1. **High Confidence, Low False-Positive Rate**:
   - Prefer distinctive prefixes, structural formats, or canonical credential paths (derived from Gitleaks, TruffleHog, git-secrets, or official vendor documentation).
   - Avoid broad generic regexes (e.g. arbitrary 32-char hex strings) that trigger false positives on git commit SHAs, SHA-256 digests, or UUIDs.
2. **Go RE2 Compatibility**:
   - Patterns are compiled with Go's standard `regexp` package (RE2), guaranteeing linear-time matching without catastrophic backtracking (ReDoS).
   - Do not use PCRE-only features (lookahead/lookbehind, backreferences).
3. **Synthetic Test Coverage Required**:
   - Every new rule MUST have both a **positive test** (using synthetic dummy values or RFC `example.com` domains—never real credentials or host paths) and a **negative/safe test** in [`mcp/internal/guard/guard_test.go`](../internal/guard/guard_test.go).
4. **Minimal Overlap & Clear Naming**:
   - Each `namedPatternConfig` must have a concise, descriptive `name` that clearly explains why an access or payload was blocked in error messages.

---

## 4. Maintenance & Continuous Improvement Workflow

Use the following loop whenever new credential formats, tool configs, or false positives are discovered (for example via `agent-retrospective`, `sandbox-review`, or upstream Gitleaks rule updates):

1. **Edit Rules**: Update [`mcp/internal/guard/rules.json`](../internal/guard/rules.json) directly.
2. **Add Tests**: Add synthetic positive and negative cases to [`mcp/internal/guard/guard_test.go`](../internal/guard/guard_test.go).
3. **Verify**:
   ```bash
   mise run mcp:test
   mise run mcp:lint
   ```
   Package initialization validates JSON structure and compiles every regex; any invalid pattern fails tests immediately.
4. **Review & Commit**: Confirm no real secrets or personal paths appear in the diff and commit with Conventional Commits (e.g., `fix(mcp): add <provider> credential rule to guard rules.json`).
