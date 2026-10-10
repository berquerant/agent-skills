# Security, Context & Reputation Audit Checklist

Guidelines and checklist for auditing dependency update and security patch PRs/MRs. Verifies change legitimacy, repository compatibility, and security safety before checking out or merging code.

---

## 1. Cost-Aware Investigation Protocol

Per repository conventions (`AGENTS.md` Rule 5 & Rule 6):
- **Batch Metadata Collection**: Retrieve PR/MR details, body/description, labels, and file lists in a single platform query (e.g. `gh pr view <id> --json number,title,body,files,labels,commits`) and save or cache locally rather than issuing repetitive API queries.
- **Local-First History Check**: Use local git history (`git log -S "<package>"`, `git log -G "<package>"`) to inspect historical versions and update frequency rather than querying remote repository APIs.
- **Strictly Offline References**: Rely exclusively on PR body notes, local commit history, and bundled manifests. Do not autonomously crawl external package registries or websites.

---

## 2. Context & Legitimacy Audit

Evaluate whether the update is reasonable and aligns with repository history:

### 2.1 Update Scope & Version Jump
- [ ] **SemVer Delta**: Identify whether the version jump is patch, minor, or major.
  - Patch updates (`x.y.1` -> `x.y.2`): Expected to have zero breaking changes; primarily bug or security fixes.
  - Minor updates (`x.1.0` -> `x.2.0`): May introduce new features; should preserve backward compatibility.
  - Major updates (`x.0.0` -> `y.0.0`): High probability of breaking changes; inspect PR body carefully for migration guides.
- [ ] **Historical Context in Repository**:
  - Has this dependency been upgraded frequently or pinned intentionally?
  - Check past commits touching the manifest: `git log -n 5 -p -- "<manifest-file>"`.
  - Check whether a comment exists explaining version pinning (e.g., `# pinned due to compatibility issue`).

### 2.2 PR Description & Changelog Review
- [ ] **Release Notes & Changelog Present**: Does the PR body provide excerpts from the dependency's changelog, release notes, or commit list?
- [ ] **Deprecated / Breaking API Notice**: Are there mentioned deprecations or removed functions that the repository might be using?
- [ ] **License Changes**: Check if the dependency modified its license (e.g., permissive MIT/Apache to copyleft or commercial source-available).

---

## 3. Security & Supply Chain Safety Audit

Audit the diff and metadata against potential supply chain attacks:

### 3.1 Typosquatting & Package Identity
- [ ] **Exact Package Name Match**: Verify the package name in manifest/lockfile matches the previously used dependency character-for-character. Guard against lookalike characters, hyphen vs underscore tricks, or subtle misspellings (e.g. `cross-fetch` vs `crossfetch`).
- [ ] **Scope / Namespace Consistency**: For scoped packages (e.g. `@org/pkg`), verify that the scope remains identical.

### 3.2 Lockfile & Registry URL Integrity
- [ ] **Official Registry Endpoint**: Verify that resolved download URLs in lockfiles point to official, recognized registry endpoints (e.g. `registry.npmjs.org`, `proxy.golang.org`, `crates.io`, `pypi.org`, `repo1.maven.org`).
- [ ] **No Unauthorized Git / URL Overrides**: Verify that dependencies were not quietly redirected to an arbitrary git branch, fork, HTTP URL, or unauthenticated tarball unless explicitly intended by the repository maintainer.
- [ ] **Integrity Hashes Present**: Ensure lockfile entries retain cryptographic hashes (SHA-256, SHA-512, subresource integrity hashes).

### 3.3 Security Advisory Alignment
- [ ] **Vulnerability Justification**: If the PR claims to fix a vulnerability (CVE, GHSA), does the version bump match the patched version declared in the advisory summary?
- [ ] **Disproportionate Dependency Bloat**: Check whether updating one minor package inadvertently pulls in an unusually large number of unexpected transitive dependencies.

---

## 4. Safety Halt Protocol (When Doubts Arise)

If ANY of the following warning signals or anomalies are detected:

| Warning Signal | Description |
|---|---|
| **Package Name Mismatch** | Subtle difference in package name or scope compared to existing configuration. |
| **Unfamiliar Registry URL** | Lockfile points to a non-standard host or unvetted private server. |
| **Abrupt Dependency Tree Bloat** | Massive unexpected addition of transitive dependencies or install scripts. |
| **Unjustified Major Bump** | Unprompted major bump with no release notes or apparent need. |
| **Ambiguous CVE Resolution** | CVE mentioned in PR title/body but patched version does not match upstream notes. |
| **Silent License Shift** | Dependency switched to a restrictive or non-standard license. |

### Immediate Halt Procedure:
1. **Immediately pause all automated execution** (do not rebase, run tests, or merge).
2. **Formulate a Safety Halt Report** stating:
   - Target PR/MR number and package name.
   - Specific anomaly or doubt detected with exact diff/metadata evidence.
   - Potential security or stability risk.
3. **Present the report to the user** and await explicit confirmation:
   > "Safety check raised doubts on PR #<num>: <summary>. Processing halted. Would you like to proceed with this PR, reject/close it, or inspect manually?"
4. **Do not resume** processing for that PR until explicit user clearance is received.
