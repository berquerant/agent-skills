# Skill Interoperability & Workflow Architecture

This document details the modular interoperability, sequential hand-off patterns, and domain-specific workflow diagrams for skills across this repository.

Skills adhere to a **loosely coupled, sequential hand-off architecture**. Skills operate independently without rigid subroutine nesting; instead, they output actionable artifacts and suggest the next logical skill for explicit user approval.

---

## 1. High-Level Category Architecture

The agent skills ecosystem is organized into six functional domains:

```mermaid
flowchart TD
    subgraph Dispatch ["1. Dispatch & Routing"]
        SR["skill-router<br/>(Analyze intent & propose pipeline)"]
    end

    subgraph Audit ["2. Audit & Inquiry (Read-Only)"]
        AUD["Audit & Diagnostics<br/>(doctor, code-review, doc-review, project-audit, evidentiary-inquiry)"]
    end

    subgraph Planning ["3. Planning & Scheduling"]
        PLAN["Planning & Isolation<br/>(explore-plan-execute, scrutinize-plan-execute, task-dependency-pipeline)"]
    end

    subgraph Execution ["4. Execution & Pacing"]
        EXEC["Setup, Gating & Logging<br/>(agent-skills-setup, step-gate, edit-paced-execution, checkpoint-log, command-snippet)"]
    end

    subgraph Remediation ["5. Remediation & Action"]
        REM["Code & Doc Authoring<br/>(refactor, doc-refactor, structured-writing, resolve-dependency-prs, multi-repo-change)"]
    end

    subgraph Meta ["6. Meta & Lifecycle"]
        META["Skill Development<br/>(skill-creator, skill-review, skill-modernize, mcp-creator, agent-retrospective)"]
    end

    %% High-level hand-offs
    Dispatch -->|"Routes complex intent"| Planning
    Dispatch -->|"Routes audit/review intent"| Audit
    Dispatch -->|"Routes PR/dep triage"| Remediation
    Audit -->|"Diagnoses suggest planned fix"| Planning
    Audit -->|"Targeted cleanups"| Remediation
    Planning -->|"Approved step-by-step plan"| Execution
    Execution -->|"Executes code mutations"| Remediation
    Remediation -->|"Verifies post-change state"| Audit
    Meta -->|"Evaluates and updates skills"| Execution
```

---

## 2. Planning & Step-Gated Execution Workflow

This workflow governs non-trivial tasks requiring structured investigation, plan scrutiny, incremental execution, and defensive pacing.

```mermaid
flowchart TD
    subgraph PlanningLayer ["Planning & Risk Assessment"]
        EPE["explore-plan-execute<br/>(Standard exploration & plan)"]
        SPE["scrutinize-plan-execute<br/>(High-risk or ambiguous tasks)"]
        GW["git-worktree<br/>(Branch isolation under .git-worktree/)"]
        TDP["task-dependency-pipeline<br/>(DAG modeling & idle-wait scheduling)"]
    end

    subgraph ExecutionLayer ["Incremental Execution & Guards"]
        SG["step-gate<br/>(Step-by-step approval gate)"]
        PACE["edit-paced-execution<br/>(Tool-call & edit pacing monitor)"]
        CPL["checkpoint-log<br/>(Milestone log & rapid recovery)"]
        CS["command-snippet<br/>(Parameterized shell scripts)"]
        SNR["snippet-review<br/>(Intelligibility & complexity audit)"]
    end

    subgraph UserGate ["User Interaction"]
        UN["user-notification<br/>(Alerts & preceding permission notices)"]
        UserApproval{{"User Sign-Off"}}
    end

    %% Connections
    GW -.->|"Provides isolated worktree"| EPE
    GW -.->|"Provides isolated worktree"| SPE
    EPE -->|"Complex or high-risk delegation"| SPE
    EPE -.->|"Parallel tasks or idle wait"| TDP
    EPE -->|"Proposes execution plan"| UserApproval
    SPE -->|"Proposes scrutinized plan"| UserApproval
    TDP -->|"Topological schedule"| UserApproval
    UserApproval -->|"Plan approved"| SG

    SG -->|"Step execution"| PACE
    SG -->|"Records step completion"| CPL
    SG -->|"Complex shell commands"| CS
    CS -->|"Audits script complexity"| SNR
    SG -->|"Gate notification before pause"| UN
```

---

## 3. Audit, Diagnostics & Remediation Workflow

This workflow handles read-only health evaluations, factual inquiries, and code reviews, handing off cleanly to mutation skills only when fixes are needed.

```mermaid
flowchart TD
    subgraph ReadOnlyAudit ["Read-Only Inspection & Inquiry"]
        DOC["doctor<br/>(Mode-selectable health checkup)"]
        EI["evidentiary-inquiry<br/>(Evidence & version-pinned facts)"]
        CR["code-review<br/>(Diff/PR review & security audit)"]
        DCR["doc-review<br/>(Documentation & specification audit)"]
        PA["project-audit<br/>(Repository baseline & tech debt)"]
        PS["project-status<br/>(Working-tree diff & drift check)"]
    end

    subgraph RemediationLayer ["Targeted Mutation Skills"]
        RF["refactor<br/>(Behavior-preserving code cleanups)"]
        DR["doc-refactor<br/>(Doc refinement, zero code change)"]
        SW["structured-writing<br/>(New docs, RFCs & proposals)"]
        DC["diff-continue<br/>(Complete work-in-progress diffs)"]
        RD["review-discussion<br/>(Resolve PR/MR review comments)"]
    end

    subgraph PlanningHandoff ["Planning Gate"]
        EPE["explore-plan-execute"]
        SG["step-gate"]
    end

    %% Handoffs
    DOC -->|"Factual uncertainty"| EI
    DOC -->|"Security / quality standards"| CR
    DOC -->|"Documentation & comment standards"| DCR
    DOC -->|"Maintainability standards"| PA
    DOC -->|"Proposes bug/security fix plan"| EPE
    DOC -->|"Proposes focused refactor"| RF
    DOC -->|"Proposes doc polish"| DR

    EI -->|"Recommends structural remediation"| RF
    EI -->|"Recommends doc remediation"| DR
    EI -->|"Recommends planned fix"| SG

    CR -->|"Dedicated documentation audit"| DCR
    CR -->|"Suggests comment resolutions"| RD
    CR -->|"Suggests documentation clarity"| DR
    CR -->|"Suggests refactoring"| RF

    DCR -->|"Proposes document refinement"| DR
    DCR -->|"Drafts missing doc from scratch"| SW
    DCR -->|"Suggests comment discussions"| RD

    PA -->|"Prioritized debt remediation"| RF
    PA -->|"Phased multi-step fixes"| SG

    PS -->|"Unfinished drift pattern"| DC
    DC -->|"Requires planning"| EPE
```

---

## 4. Multi-Repository & Dependency PR Triage Workflow

This workflow orchestrates dependency update audits, CI log matching, and automated PR merging across single or multiple repositories.

```mermaid
flowchart TD
    subgraph Discovery ["Repository Discovery"]
        RRDP["resolve-recent-dependency-prs<br/>(Discover active repos with open PRs)"]
    end

    subgraph Triage ["PR Triage & Verification"]
        RDP["resolve-dependency-prs<br/>(Audit, build, test & merge PRs)"]
        GW["git-worktree<br/>(Worktree isolation per PR)"]
        TDP["task-dependency-pipeline<br/>(Cooperative idle-wait during CI)"]
    end

    subgraph Actions ["PR Lifecycle & Actions"]
        CM["change-message<br/>(Commit & PR description formatting)"]
        CR["code-review<br/>(Thorough mode security audit)"]
        RF["refactor<br/>(Complex fix hand-off)"]
        PS["project-status<br/>(Post-merge drift check)"]
    end

    %% Connections
    RRDP -->|"Sequentially delegates per repo"| RDP
    GW -.->|"Isolates PR branch"| RDP
    RDP -.->|"Parallel PR waits"| TDP
    RDP -.->|"Formats commit/PR metadata"| CM
    RDP -.->|"Thorough security audit"| CR
    RDP -.->|"Fixes broken updates"| RF
    RDP -.->|"Verifies working-tree post-merge"| PS
```

---

## 5. Skill Meta & Modernization Lifecycle

This workflow governs the creation, review, and evolution of agent skills to align with new model capabilities and harness conventions.

```mermaid
flowchart TD
    subgraph MetaTools ["Skill Authoring & Evaluation"]
        SC["skill-creator<br/>(Scaffold new skills adhering to spec)"]
        SR["skill-review<br/>(Read-only audit: checks A-K)"]
        SM["skill-modernize<br/>(Frontier model alignment & remediation)"]
        MC["mcp-creator<br/>(Scaffold companion Go MCP servers)"]
        AR["agent-retrospective<br/>(Transcript & session log mining)"]
    end

    subgraph FixHandoff ["Remediation Hand-offs"]
        SG["step-gate<br/>(Phased skill fixes)"]
        RF["refactor<br/>(Structural skill refactoring)"]
        DR["doc-refactor<br/>(Skill documentation & rubric polish)"]
        CS["command-snippet<br/>(Parameterize repeated commands)"]
    end

    %% Connections
    SC -->|"New skill validated by"| SR
    SR -.->|"Shares skill spec & checks"| SC
    SR -->|"Suggests phased fixes"| SG
    SR -->|"Suggests structure cleanups"| RF
    SR -->|"Suggests doc refinement"| DR

    AR -->|"Proposes new skills to"| SC
    AR -->|"Proposes command snippets to"| CS
    AR -->|"Proposes doc/reference fixes to"| DR

    SM -->|"Updates review criteria in"| SR
    SM -.->|"Syncs authoring rules to"| SC
    SM -->|"Orchestrates phased modernisation"| SG
    SM -->|"Orchestrates structural fixes"| RF
    SM -->|"Orchestrates doc updates"| DR
```

---

## 6. Interoperability Principles

1. **Intent-Driven Dispatch & Pipeline Routing**: `skill-router` acts as a read-only entrypoint dispatcher to analyze context, map user intent, and present an explicit sequential pipeline for user sign-off before executing any skill.
2. **Read-Only vs. Mutation Separation**: Audit and investigation skills (`project-audit`, `project-status`, `code-review`, `doc-review`, `skill-review`, `evidentiary-inquiry`, `doctor`, `snippet-review`, `skill-router`, `agent-retrospective`) are strictly read-only and never mutate files. When changes are required, they suggest mutation skills (`refactor`, `doc-refactor`, `step-gate`, `diff-continue`).
3. **Sequential Hand-off (User-in-the-Loop)**: Rather than calling other skills directly in a deep nested chain, skills present findings and recommend the next skill for the user to approve.
4. **Reference Sharing over Duplication**: When audit standards overlap (such as security and public release checklists), skills share references (e.g. `../code-review/references/security-checklist.md`) via relative links instead of duplicating checklist content.
5. **Gate Ownership**: When a skill's guidance or artifact is used inside another workflow, the calling workflow owns the approval gates; avoid introducing duplicate or redundant approval checkpoints.
