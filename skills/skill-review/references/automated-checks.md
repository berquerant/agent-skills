# Automated Checks

Reference documentation for the automated, mechanical checks used during skill review.

All checks are implemented in the self-contained script [`../scripts/run-checks.sh`](../scripts/run-checks.sh).

---

## Running the Automated Checks

Run the automated check script against the target skill directory:

```sh
skills/skill-review/scripts/run-checks.sh skills/<target-skill>
```

The script runs read-only commands and prints a report covering checks A through J.

---

## Checks Summary and Criteria

### A. Inventory & Size (Dimension 2)
- **Check**: Lists all files and line counts.
- **Criteria**: `SKILL.md` line count under 500 lines [F]; roughly 80–250 lines recommended [W].

### B. Frontmatter `name` (Dimension 3)
- **Check**: Validates regex `^[a-z0-9]+(-[a-z0-9]+)*$` and ensures name matches parent directory.
- **Criteria**: Must match directory name and character constraints [F]. Recommended 3–30 chars [W].

### C. Frontmatter `description` (Dimension 3)
- **Check**: Extracts description and computes character count.
- **Criteria**: Non-empty, 1–1024 characters [F]. Third-person phrasing, 100–350 chars recommended [W].

### D. Section Structure & Verify Checkpoints (Dimension 3)
- **Check**: Inspects headings order and confirms every `## Step N` has an explicit `Verify:` line.
- **Criteria**: Every step has an observable `Verify:` checkpoint [W]. Standard section order followed [W].

### E. Links & Path Portability (Dimension 2)
- **Check**: Validates internal relative markdown links; scans for absolute machine paths (`/Users/`, `/home/`, `C:\`).
- **Criteria**: No broken relative links; no hardcoded absolute machine paths [F].

### F. Script Permissions (Dimension 2)
- **Check**: Checks executable bit (`+x`) on all files under `scripts/`.
- **Criteria**: All scripts executable [W].

### G. Sibling References & Coupling (Dimensions 4 & 5)
- **Check**: Counts mentions of sibling skills and cross-skill relative links (`../<sibling>/`).
- **Criteria**: Recommended hand-off or shared references [OK]. No mandatory nested subroutines [F].

### H. Repository Registration (Dimension 5)
- **Check**: Checks presence in `README.md` and `AGENTS.md`.
- **Criteria**: Registered in repo listings [W].

### I. Cross-Skill Duplication Hints (Dimension 5, Batch Mode)
Headings that appear in more than one skill's references suggest duplicated content:
```sh
awk '
  /^###? / { n = FILENAME; sub(/\/references\/.*/, "", n); sub(/.*\//, "", n)
             if (index(s[$0], " " n) == 0) { s[$0] = s[$0] " " n; c[$0]++ } }
  END { for (h in c) if (c[h] > 1) print h " ->" s[h] }
' skills/*/references/*.md
```

### J. External Hyperlinks & Self-Containment (Dimensions 2 & 5)
- **Check**: Scans for `https?://` in markdown files.
- **Criteria**: Core procedures and specs must not delegate to external URLs [F]. Any human citation URLs must have an explicit notice forbidding autonomous agent access [W].
