---
name: sandbox-review
description: >-
  Use this skill when auditing AI agent sandbox configurations, execution harness
  policies, or sandbox violation logs. Evaluates network, filesystem, secret, and
  process boundaries against platform-agnostic least-privilege requirements and
  outputs a structured read-only report without modifying files.
license: MIT
---

# Sandbox & Execution Harness Review

Use this skill to audit AI agent sandbox configurations, declarative harness policies, or boundary violation logs.
It evaluates execution environments against platform-agnostic security principles without modifying any files, surfacing missing guardrails, escape vectors, and overly permissive rules with concrete, vendor-neutral remediation guidance.

This skill is strictly read-only and directly references three reference documents:
- [harness-security-spec.md](references/harness-security-spec.md) (references/harness-security-spec.md): Core principles, four control-domain requirements, and operational audit criteria.
- [least-privilege-profiles.md](references/least-privilege-profiles.md) (references/least-privilege-profiles.md): Declarative baseline policy profiles starting from an offline local default.
- [persistence-and-escape-vectors.md](references/persistence-and-escape-vectors.md) (references/persistence-and-escape-vectors.md): Catalog of persistence targets, sandbox escape vectors, and violation triage rules.

For general application code or repository public-release audits, use [code-review](../code-review/SKILL.md) (../code-review/SKILL.md) instead.

---

## Step 1: Identify Scope, Mode, and Baseline

Determine what is being reviewed and confirm the working tree starts clean:

1. **Identify Target Artifacts**:
   - Sandbox or harness configuration files (e.g., container/microVM policies, OS sandbox profiles, DevContainer manifests, agent CLI permission configs, CI runner definitions) or declarative policy documents.
   - Boundary denial or violation logs when diagnosing blocked operations.
2. **Select Review Mode**:
   - **Mode 1 — Config & Policy Audit**: Static evaluation of sandbox or harness configurations against [harness-security-spec.md](references/harness-security-spec.md) (references/harness-security-spec.md) and a target baseline in [least-privilege-profiles.md](references/least-privilege-profiles.md) (references/least-privilege-profiles.md).
   - **Mode 2 — Violation Log Triage**: Classification of recorded denial events using [persistence-and-escape-vectors.md](references/persistence-and-escape-vectors.md) (references/persistence-and-escape-vectors.md) to separate harmless runtime noise from legitimate workflow needs and escape indicators.
3. **Record Target Profile & Workspace Baseline**:
   - Identify the intended operational profile (defaulting to Profile 1: Everyday Local Offline Baseline unless the task requires network or external API access).
   - Run `git status --short` when inside a repository so you can verify at completion that zero files were modified.

Verify: Target files or logs, selected review mode, target baseline profile, and initial `git status --short` output are recorded.

---

## Step 2: Audit Control Domains or Triage Violation Logs

Perform read-only inspection according to the selected mode. Do **not** edit, format, or auto-fix any files.

### Mode 1: Config & Policy Audit
Evaluate the target configuration across all four domains in [harness-security-spec.md](references/harness-security-spec.md) (references/harness-security-spec.md) and cross-check every entry in [persistence-and-escape-vectors.md](references/persistence-and-escape-vectors.md) (references/persistence-and-escape-vectors.md):

1. **Network Boundary**: Check default-deny egress, host/port allowlists, strict non-interactive denial, explicit deny-before-allow rules with actionable reasons, post-DNS IP validation (SSRF / DNS rebinding / metadata), local port/IPC/socket blocking, and L7 inspection hygiene.
2. **Filesystem Boundary**: Check read scoping (user home and sibling projects denied; workspace re-allowed), write scoping (workspace + essential temp only; secret files denied over allow), and unconditional protection of persistence points.
3. **Credential & Secret Protection**: Check whether secrets are excluded or injected at the boundary proxy via sentinel masking, whether structured credentials preserve syntax or re-sign HMACs at the boundary, and whether plaintext HTTP injection is forbidden.
4. **Process & OS Execution Boundary**: Check blocking of external app launchers and GUI/scripting IPC, plus disabling of isolation relaxations (nested container relax, new user namespaces, certificate-daemon leak paths, unnecessary PTY allocation).
5. **Operational & Audit Controls**: Check structured violation logging, actionable denial feedback to the agent, and scoped (`command + path`) noise suppression.

### Mode 2: Violation Log Triage
Classify each unique denial event using the rubric in [persistence-and-escape-vectors.md](references/persistence-and-escape-vectors.md) (references/persistence-and-escape-vectors.md):

1. **Benign Runtime Noise**: Deterministic, harmless OS or runtime probes — candidate for narrow `command + path` log suppression without widening permissions.
2. **Legitimate Blocked Operation**: Required by the intended task — evaluate safe alternatives first (offline cache, read-only mount, boundary proxy injection) before recommending a minimal scoped exception.
3. **Escape or Injection Indicator**: Attempts targeting persistence points, credential stores, metadata endpoints, local IPC, or privilege relaxation — keep blocked and escalate as a security finding.

Verify: Every relevant control domain (Mode 1) or unique violation event (Mode 2) has been evaluated with file/line or log-entry evidence and zero files have been modified.

---

## Step 3: Categorize Findings and Draft Vendor-Neutral Remedies

Assign each finding a severity level and concrete remediation guidance:

- **Critical**: Missing default-deny boundary, unprotected persistence points (`.git/config`, hooks, shell rc, agent configs), raw secret exposure or plaintext secret injection, SSRF/DNS-rebinding gap to loopback/metadata, or active escape/injection indicator.
- **High**: Overbroad home-directory read access, local IPC/socket or external app-launcher exposure, enabled isolation relaxations (nested relax, new userns, unnecessary PTY), or interactive bypass of denied egress.
- **Medium**: Missing L7 Host/SNI or path-normalization checks, unstructured denial errors lacking actionable feedback, or missing `command + path` noise suppression causing alert fatigue.
- **Low**: Minor policy documentation gaps, overly broad noise suppression rules, or unclear profile deviation rationale.

Format every finding with:
1. **Where**: Exact file path and line range (or log identifier) formatted per [output-presentation](../output-presentation/SKILL.md) (../output-presentation/SKILL.md).
2. **Requirement**: Specific control or vector from [harness-security-spec.md](references/harness-security-spec.md) (references/harness-security-spec.md) or [persistence-and-escape-vectors.md](references/persistence-and-escape-vectors.md) (references/persistence-and-escape-vectors.md).
3. **Why**: Concrete threat scenario, escape path, or operational failure mode.
4. **How**: Platform-agnostic policy fix (declarative rule or minimal profile alignment from [least-privilege-profiles.md](references/least-privilege-profiles.md) (references/least-privilege-profiles.md)), plus an example mapping when a specific runtime is under review.

Verify: All findings carry explicit severity, Where/Requirement/Why/How fields, and baseline profile gap summary.

---

## Step 4: Present Report and Recommend Next Steps

1. **Highlight Critical Hazards Upfront**: Surface any Critical boundary gaps, unprotected persistence points, or active escape indicators at the top of the report.
2. **Deliver Structured Summary**: Present an executive summary (target, mode, baseline profile, and the Domain Compliance Matrix from [harness-security-spec.md](references/harness-security-spec.md) (references/harness-security-spec.md)), prioritized findings, and positive controls worth preserving.
3. **Confirm Read-Only Integrity**: Re-run `git status --short` (when in a repository) to confirm the workspace matches the Step 1 baseline.
4. **Recommend Sequential Hand-Off**:
   - **To plan and apply policy or harness configuration fixes**: Recommend [explore-plan-execute](../explore-plan-execute/SKILL.md) (../explore-plan-execute/SKILL.md) or [step-gate](../step-gate/SKILL.md) (../step-gate/SKILL.md).
   - **For high-risk architecture or boundary redesign**: Recommend [scrutinize-plan-execute](../scrutinize-plan-execute/SKILL.md) (../scrutinize-plan-execute/SKILL.md).
   - **To refine declarative policy documents without changing runtime configs**: Recommend [doc-refactor](../doc-refactor/SKILL.md) (../doc-refactor/SKILL.md).
   - **To audit application code or repository public-release safety**: Recommend [code-review](../code-review/SKILL.md) (../code-review/SKILL.md).

Verify: Report is delivered with dual-formatted paths, unchanged `git status --short` is confirmed, and follow-up hand-off options are presented.

---

## Guidelines

- **Maintain a strictly read-only audit posture**: Inspect configurations and logs without editing files so security reviews remain safe, reproducible, and side-effect-free.
- **Default to least-privilege baselines**: Measure every configuration against Profile 1 (offline local baseline) first and require explicit task justification for any wider network, filesystem, secret, or process permission.
- **Prefer safe alternatives over blanket policy relaxation**: During violation triage, recommend offline substitutes, read-only mounts, or boundary proxy injection before suggesting broader allowlist rules so security boundaries do not erode over time.
- **Keep recommendations platform-agnostic**: State requirements and fixes in declarative harness terms first, adding product-specific syntax only as non-exclusive examples so policies remain portable across runtimes.
- **Sanitize all reported excerpts**: Mask or omit real credentials, private hostnames, internal IP addresses, and host-specific user paths in audit output so reports do not leak sensitive environment details.
