# Code-Truth Alignment & Cross-Checking Guide

This guide provides concrete inspection procedures for cross-checking technical documentation against actual repository implementation files.

---

## 1. Overview & Verification Philosophy

Documentation frequently suffers from drift: code evolves, but comments, README files, and architecture diagrams remain unchanged.
During `doc-review`, agents must treat the code in the repository as the primary ground truth, actively verifying claims made in documentation.

---

## 2. Cross-Checking Procedures

### A. CLI Commands & Flags
When documentation cites command-line examples (e.g. `tool run --flag value`):
1. **Search Subcommands and Flags**:
   - Locate flag definitions or argument parser schemas in the codebase (e.g. `pflag`, `cobra`, `argparse`, `clap`, `structconfig`).
   - Confirm that the command name, subcommand hierarchy, and flag names match source declarations.
2. **Verify Default Values & Enums**:
   - Check default values mentioned in documentation against actual struct initialization or flag defaults.
   - Confirm that listed acceptable values (enums) are complete and accurate.

### B. Function Signatures & API Interfaces
When documentation presents API references or usage snippets:
1. **Locate Declaration**:
   - Search for the function, method, interface, or struct in the codebase.
2. **Compare Parameters & Returns**:
   - Check parameter names, counts, order, and types.
   - Confirm return values and error handling signatures match.
3. **Check Version Anchors & Deprecations**:
   - Verify whether symbols marked as recommended are active, or if comments indicate deprecation.

### C. Configuration & Schema Declarations
When documentation describes YAML, JSON, TOML, or environment configurations:
1. **Cross-Check Schema Keys**:
   - Match configuration field names against deserialization structs or validation schemas.
   - Verify case-sensitivity conventions (e.g. `camelCase` vs `snake_case` vs `kebab-case`).
2. **Environment Variable Names**:
   - Search for `os.Getenv`, `process.env`, or environment struct tags in the codebase.
   - Confirm documented variable prefixes (e.g. `APP_PORT` vs `SERVICE_PORT`) match implementation.

### D. File Paths & Directory Layouts
When documentation references repository paths or structure:
1. **Path Existence**:
   - Check if the referenced path exists in the repository tree.
2. **Relative Reference Viability**:
   - Ensure relative links from the document's directory lead to an existing file.

---

## 3. Discrepancy Triage Matrix

When a discrepancy is found between documentation and code:

| Observed Discrepancy | Likely Root Cause | Review Finding Severity | Recommended Remediation |
|---|---|---|---|
| Command flag or API parameter does not exist in code | Code evolved; doc is stale | 🔴 Critical / 🟡 Warning | Correct documentation or add deprecation warning via `doc-refactor` |
| Documented behavior contradicts implementation logic | Doc drift or unnoticed code regression | 🟡 Warning / 🔴 Critical | Report finding; hand off to `doc-refactor` (if doc error) or `explore-plan-execute` (if code bug) |
| File path or link in doc points to non-existent target | File moved or deleted | 🟡 Warning | Update path or link via `doc-refactor` |
| Code snippet uses deprecated API without explanation | Incomplete migration | 🟡 Warning | Update snippet to modern idiom or document rationale |
