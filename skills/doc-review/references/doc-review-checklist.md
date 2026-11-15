# Technical Documentation Review Checklist

This document defines the evaluation criteria across five core dimensions for agents operating under `doc-review`.

---

## Dimension 1: Code-Truth Alignment & Drift Detection

Documentation that diverges from actual code behavior is actively harmful. Verify that text reflects current codebase reality:

- [ ] **Executable Command Veracity**:
  - Do CLI commands, script names, subcommands, and flags exist in the current codebase?
  - Are default parameter values, environment variable names, and configuration paths up to date?
- [ ] **API & Interface Synchronization**:
  - Do documented function signatures, types, parameter lists, and return values match source code declarations?
  - Are deprecated interfaces clearly marked with upgrade paths, or have obsolete functions been purged from active docs?
- [ ] **Working Example Integrity**:
  - Do code snippets compile or run against the current project dependencies and SDK versions?
  - Are imports, namespaces, and module names valid?
- [ ] **Structural & File Path Accuracy**:
  - Do referenced file paths, directory layouts, and configuration file names match the actual repository tree?

---

## Dimension 2: Structural Clarity & Reader Persona

Documentation must provide immediate orientation and scale appropriately to the intended reader's needs:

- [ ] **Why-first Framing**:
  - Does the document explain *why* the component, architecture, or convention exists before describing mechanics?
  - Are non-obvious design choices, trade-offs, and historical context preserved?
- [ ] **Logical Information Hierarchy**:
  - Does the document follow a progressive information structure (Overview/Intent → Prerequisites → Step-by-Step Instructions → Verification & Troubleshooting)?
  - Are headings logical, unique within the file, and properly nested (avoid skipping heading levels, e.g. `##` to `####`)?
- [ ] **Target Audience Alignment**:
  - Is the level of technical abstraction appropriate for the reader (end-user, library consumer, internal contributor)?
  - Are specialized terms, domain acronyms, or project-specific idioms defined on first use?
- [ ] **Scannability & Visual Navigation**:
  - Are lists, tables, callouts, and code blocks used effectively to break up walls of text?
  - Are key takeaways, warnings, and prerequisites easily scannable?

---

## Dimension 3: Technical Hygiene & Asset Integrity

Documentation assets must be syntactically valid, accessible, and self-contained:

- [ ] **Link & Anchor Viability**:
  - Do all relative markdown links (`[text](../path/file.md)`) resolve to real files in the repository?
  - Do line anchors (`#L10-L20`) and section anchor fragments (`#section-title`) match actual targets?
- [ ] **Diagram & Visualization Syntax**:
  - Do Mermaid code blocks render correctly without syntax errors?
  - Are node labels containing special characters (brackets, parentheses) properly quoted?
  - Are diagram directions and relationships logically structured?
- [ ] **Code Block Hygiene**:
  - Does every fenced code block specify an appropriate syntax highlighting language tag (e.g. ```` ```yaml ````, ```` ```go ````, ```` ```sh ````)?
  - Are shell commands concise, formatted for readability, and free from dangerous or unquoted parameters?
- [ ] **Standard Markdown Compliance**:
  - Are lists consistently formatted, tables well-aligned, and blockquotes properly styled?

---

## Dimension 4: Public Exposure & Confidentiality Safety

Documentation intended for publication or external distribution must never leak sensitive details:

- [ ] **Zero Credential Leaks**:
  - Are there any real API keys, bearer tokens, private keys, SSH keys, passwords, or secret hashes?
  - Do documentation examples strictly use standard dummy placeholders (e.g. `sk_test_...`, `EXAMPLE_TOKEN`)?
- [ ] **Zero Infrastructure & Network Leaks**:
  - Are internal domain names (e.g. `.corp`, `.internal`, `.local`), intranet hostnames, or corporate VPN endpoints exposed?
  - Are private IPv4/IPv6 addresses (e.g. `10.x.x.x`, `192.168.x.x`, `172.16-31.x.x`) or proprietary staging URLs leaked?
- [ ] **Zero Proprietary & Customer Data Leaks**:
  - Does the text expose confidential business metrics, proprietary algorithms, unreleased feature codenames, or customer PII?

---

## Dimension 5: Information Economy & Conservation

Documentation must balance completeness with cognitive conciseness:

- [ ] **Prerequisites & Invariants Coverage**:
  - Are all environmental assumptions, required tool versions, and necessary permissions explicitly documented?
  - Are invariant rules, constraints, and side effects clearly articulated?
- [ ] **Edge Cases & Failure Modes**:
  - Are non-obvious failure modes, common mistakes, and troubleshooting steps documented?
- [ ] **No Code Paraphrasing Filler**:
  - Does the documentation avoid redundant line-by-line paraphrasing of obvious code syntax?
- [ ] **No Context Loss During Updates**:
  - When reviewing document revisions or PR diffs, confirm that hard-won institutional context, subtle caveats, or historical rationale were not accidentally discarded.
