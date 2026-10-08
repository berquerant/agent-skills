# Automated Checks

Portable POSIX shell commands for the mechanical parts of a skill review. Every command is read-only.

Set these variables first:

```sh
SKILL_DIR=skills/my-skill        # target skill directory
SKILLS_ROOT=$(dirname "$SKILL_DIR")
NAME=$(basename "$SKILL_DIR")
```

Each result feeds the checklist item named in its heading. Triage false positives manually and annotate them in the report. Links inside fenced code examples are a common source.

---

## A. Inventory & Size (Dimension 2)

```sh
find "$SKILL_DIR" -type f | sort
wc -l "$SKILL_DIR"/SKILL.md "$SKILL_DIR"/*/*.md 2>/dev/null
```

Line count of `SKILL.md`: under 500 is required [F]; about 80–250 is recommended [W].

## B. Frontmatter `name` (Dimension 3)

```sh
fm_name=$(awk '/^---[[:space:]]*$/{n++; next} n==1 && /^name:/{sub(/^name:[[:space:]]*/, ""); print; exit}' "$SKILL_DIR/SKILL.md")
echo "name=$fm_name len=${#fm_name}"
printf '%s\n' "$fm_name" | grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$' && echo "pattern: OK" || echo "pattern: FAIL"
[ "$fm_name" = "$NAME" ] && echo "dir match: OK" || echo "dir match: FAIL ($NAME)"
```

## C. Frontmatter `description` (Dimension 3)

This prints the description joined into a single line, followed by its character count. It handles plain, `>-`, and `|` block styles.

```sh
desc=$(awk '
  /^---[[:space:]]*$/ { n++; next }
  n == 1 && /^description:/ { f = 1; sub(/^description:[[:space:]]*([>|][-+]?)?[[:space:]]*/, ""); if ($0 != "") printf "%s ", $0; next }
  n == 1 && f && /^[^[:space:]]/ { f = 0 }
  n == 1 && f { sub(/^[[:space:]]+/, ""); printf "%s ", $0 }
' "$SKILL_DIR/SKILL.md" | sed 's/[[:space:]]*$//')
printf '%s\n' "$desc"
printf '%s' "$desc" | wc -c
```

Check that the count is between 1 and 1024 [F]. Then read the text against the format convention [W].

## D. Section Structure & Verify Checkpoints (Dimension 3)

```sh
awk '/^```/ { code = !code; next } !code && /^##? / { print NR ": " $0 }' "$SKILL_DIR/SKILL.md"
awk '
  /^```/ { code = !code }
  code { next }
  /^## / { if (step != "" && !v) print "MISSING Verify: " step; step = ($0 ~ /^## Step /) ? $0 : ""; v = 0; next }
  /Verify:/ { v = 1 }
  END { if (step != "" && !v) print "MISSING Verify: " step }
' "$SKILL_DIR/SKILL.md"
```

The first command lists the headings so you can check their order. The second prints nothing when every Step section has a `Verify:` line.

## E. Links (Dimension 2)

Relative links that do not resolve. Fenced code blocks are skipped.

```sh
find "$SKILL_DIR" -name '*.md' | while read -r f; do
  d=$(dirname "$f")
  awk '/^[[:space:]]*```/ { code = !code; next } !code' "$f" |
    grep -o '](\([^)]*\))' | sed 's/^](//; s/)$//; s/#.*//; s/[[:space:]].*//' |
    grep -Ev '^(https?:|mailto:|$)' | while read -r p; do
      [ -e "$d/$p" ] || echo "BROKEN: $f -> $p"
    done
done
```

Absolute or machine-specific paths:

```sh
grep -rnE '\]\((/|file://)|/Users/|/home/[a-z]|[A-Z]:\\' "$SKILL_DIR"
```

## F. Script Permissions (Dimension 2)

```sh
[ -d "$SKILL_DIR/scripts" ] && find "$SKILL_DIR/scripts" -type f ! -perm -u+x
```

This prints nothing when every script is executable.

## G. Sibling References (Dimensions 4 & 5)

Count how often each sibling skill is mentioned:

```sh
for s in "$SKILLS_ROOT"/*/; do
  s=$(basename "$s"); [ "$s" = "$NAME" ] && continue
  c=$(grep -rF -- "$s" "$SKILL_DIR" | wc -l)
  [ "$c" -gt 0 ] && echo "$s: $c"
done
grep -rnE '\.\./[a-z0-9-]+/' "$SKILL_DIR"     # cross-skill file links
```

For every mention, read the surrounding text and classify it as one of:
- a recommended hand-off [OK]
- a shared reference [OK]
- a mandatory subroutine [F]

## H. Repository Registration (Dimension 5)

Run this from the repository root and adjust the file names to the repository:

```sh
for doc in README.md AGENTS.md; do
  [ -f "$doc" ] && printf '%s: %s\n' "$doc" "$(grep -cF -- "$NAME" "$doc")"
done
```

A count of `0` in a doc that lists skills means the skill is not registered [W].

## I. Cross-Skill Duplication Hints (Dimension 5, batch mode)

Headings that appear in more than one skill's references suggest duplicated content:

```sh
awk '
  /^###? / { n = FILENAME; sub(/\/references\/.*/, "", n); sub(/.*\//, "", n)
             if (index(s[$0], " " n) == 0) { s[$0] = s[$0] " " n; c[$0]++ } }
  END { for (h in c) if (c[h] > 1) print h " ->" s[h] }
' "$SKILLS_ROOT"/*/references/*.md
```

This only flags candidates. Confirm real duplication by comparing the content side by side.

## J. External Hyperlinks & Self-Containment (Dimension 2 & 5)

Scan for external HTTP/HTTPS hyperlinks in markdown files. Fenced code examples and git remote regexes may produce false positives and should be triaged:

```sh
grep -rnE 'https?://' "$SKILL_DIR"
```

Verify that:
- Core procedures, checklists, and specifications do not delegate to external URLs [F].
- Any remaining external URLs are strictly non-essential human citations accompanied by an explicit note forbidding autonomous agent access [W].
