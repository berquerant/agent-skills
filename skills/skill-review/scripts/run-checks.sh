#!/bin/sh
# run-checks.sh: Runs automated quality checks for a target agent skill.
# Usage: scripts/run-checks.sh <path-to-skill-dir>

set -eu

if [ $# -lt 1 ]; then
  echo "Usage: $0 <path-to-skill-dir>" >&2
  exit 1
fi

SKILL_DIR="$1"

if [ ! -f "$SKILL_DIR/SKILL.md" ]; then
  echo "Error: $SKILL_DIR/SKILL.md does not exist." >&2
  exit 1
fi

SKILL_DIR_ABS=$(cd "$SKILL_DIR" && pwd)
SKILLS_ROOT=$(dirname "$SKILL_DIR_ABS")
NAME=$(basename "$SKILL_DIR_ABS")

errors=0

echo "=================================================="
echo "Automated Check Report for: $NAME"
echo "Target Directory: $SKILL_DIR"
echo "=================================================="

# Check A: Inventory & Size
echo "--- Check A: Inventory & Line Count ---"
find "$SKILL_DIR" -type f | sort
wc -l "$SKILL_DIR"/SKILL.md "$SKILL_DIR"/*/*.md 2>/dev/null || true
skill_lines=$(wc -l <"$SKILL_DIR/SKILL.md" | tr -d ' ')
if [ "$skill_lines" -gt 500 ]; then
  echo "size: FAIL ($skill_lines lines, exceeds 500)"
  errors=$((errors + 1))
else
  echo "size: OK ($skill_lines lines)"
fi

# Check B: Frontmatter name
echo "--- Check B: Frontmatter 'name' ---"
fm_name=$(awk '
  /^---[[:space:]]*$/ { n++; next }
  n == 1 && /^name:/ {
    sub(/^name:[[:space:]]*/, "");
    sub(/[[:space:]]+#.*$/, "");
    sub(/^["\047]/, "");
    sub(/["\047]$/, "");
    sub(/[[:space:]]+$/, "");
    print;
    exit
  }
' "$SKILL_DIR/SKILL.md")
echo "name=$fm_name len=${#fm_name}"
if [ -z "$fm_name" ]; then
  echo "pattern: FAIL (missing or empty name)"
  errors=$((errors + 1))
elif printf '%s\n' "$fm_name" | grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$'; then
  echo "pattern: OK"
else
  echo "pattern: FAIL"
  errors=$((errors + 1))
fi
if [ "$fm_name" = "$NAME" ]; then
  echo "dir match: OK"
else
  echo "dir match: FAIL (expected $NAME)"
  errors=$((errors + 1))
fi

# Check C: Frontmatter description
echo "--- Check C: Frontmatter 'description' ---"
desc=$(awk '
  /^---[[:space:]]*$/ { n++; next }
  n == 1 && /^description:/ {
    f = 1;
    sub(/^description:[[:space:]]*/, "");
    sub(/[[:space:]]*#.*$/, "");
    sub(/^([>|][-+]?)?[[:space:]]*/, "");
    if ($0 ~ /^["\047]/) {
      q = substr($0, 1, 1);
      sub(/^["\047]/, "");
    }
    desc = $0;
    next
  }
  n == 1 && f && /^[^[:space:]]/ { f = 0 }
  n == 1 && f {
    sub(/^[[:space:]]+/, "");
    sub(/[[:space:]]+$/, "");
    if (desc != "" && $0 != "") desc = desc " " $0;
    else if ($0 != "") desc = $0;
  }
  END {
    sub(/[[:space:]]+$/, "", desc);
    if (q != "" && substr(desc, length(desc), 1) == q) {
      desc = substr(desc, 1, length(desc) - 1);
    }
    print desc;
  }
' "$SKILL_DIR/SKILL.md")
printf 'description: %s\n' "$desc"
desc_len=$(printf '%s' "$desc" | wc -c | tr -d ' ')
echo "char count: $desc_len"
if [ "$desc_len" -eq 0 ]; then
  echo "description: FAIL (empty)"
  errors=$((errors + 1))
elif [ "$desc_len" -gt 1024 ]; then
  echo "description: FAIL (exceeds 1024 chars: $desc_len)"
  errors=$((errors + 1))
else
  echo "description: OK"
fi

# Check D: Section Structure & Verify Checkpoints
echo "--- Check D: Headings & Verify Checkpoints ---"
awk '/^```/ { code = !code; next } !code && /^##? / { print NR ": " $0 }' "$SKILL_DIR/SKILL.md"
missing_verify=$(awk '
  /^```/ { code = !code }
  code { next }
  /^## / { if (step != "" && !v) print "MISSING Verify: " step; step = ($0 ~ /^## Step /) ? $0 : ""; v = 0; next }
  /Verify:/ { v = 1 }
  END { if (step != "" && !v) print "MISSING Verify: " step }
' "$SKILL_DIR/SKILL.md")
if [ -n "$missing_verify" ]; then
  printf '%s\n' "$missing_verify"
  missing_count=$(printf '%s\n' "$missing_verify" | wc -l | tr -d ' ')
  errors=$((errors + missing_count))
fi

# Check E: Links
echo "--- Check E: Relative & Absolute Links ---"
broken_links=$(find "$SKILL_DIR" -name '*.md' | while read -r f; do
  d=$(dirname "$f")
  awk '/^[[:space:]]*```/ { code = !code; next } !code { gsub(/`[^`]*`/, ""); print }' "$f" |
    grep -o '](\([^)]*\))' | sed 's/^](//; s/)$//; s/#.*//; s/[[:space:]].*//' |
    grep -Ev '^(https?:|mailto:|$)' | while read -r p; do
    [ -e "$d/$p" ] || echo "BROKEN link: $f -> $p"
  done
done)
if [ -n "$broken_links" ]; then
  printf '%s\n' "$broken_links"
  broken_count=$(printf '%s\n' "$broken_links" | wc -l | tr -d ' ')
  errors=$((errors + broken_count))
fi
abs_links=$(find "$SKILL_DIR" -name '*.md' ! -name 'automated-checks.md' -exec grep -HnE '\]\((/|file://)|/Users/|/home/[a-z]|[A-Z]:[/\\]' {} + 2>/dev/null || true)
if [ -n "$abs_links" ]; then
  echo "ABSOLUTE/MACHINE paths found:"
  printf '%s\n' "$abs_links"
  abs_count=$(printf '%s\n' "$abs_links" | wc -l | tr -d ' ')
  errors=$((errors + abs_count))
fi

# Check F: Script Permissions
echo "--- Check F: Script Permissions ---"
if [ -d "$SKILL_DIR/scripts" ]; then
  non_exec=$(find "$SKILL_DIR/scripts" -type f ! -perm -u+x)
  if [ -n "$non_exec" ]; then
    printf '%s\n' "$non_exec" | while read -r s; do
      [ -n "$s" ] && echo "NON-EXECUTABLE script: $s"
    done
    non_exec_count=$(printf '%s\n' "$non_exec" | wc -l | tr -d ' ')
    errors=$((errors + non_exec_count))
  fi
fi

# Check G: Sibling References
echo "--- Check G: Sibling Skill References ---"
for s in "$SKILLS_ROOT"/*/; do
  s=$(basename "$s")
  [ "$s" = "$NAME" ] && continue
  case "$s" in
    *[!a-z0-9-]*) continue ;;
  esac
  c=$(grep -rnE "(^|[^a-zA-Z0-9_-])$s([^a-zA-Z0-9_-]|$)" "$SKILL_DIR" 2>/dev/null | wc -l | tr -d ' ')
  [ "$c" -gt 0 ] && echo "$s: $c"
done
grep -rnE '\.\./[a-z0-9-]+/' "$SKILL_DIR" || true

# Check H: Repository Registration
echo "--- Check H: Repository Registration ---"
REPO_ROOT=$(git -C "$SKILL_DIR" rev-parse --show-toplevel 2>/dev/null || (cd "$SKILLS_ROOT/.." 2>/dev/null && pwd) || pwd)
for doc in README.md AGENTS.md; do
  doc_path="$REPO_ROOT/$doc"
  if [ -f "$doc_path" ]; then
    printf '%s: %s\n' "$doc" "$(grep -cF -- "$NAME" "$doc_path" || true)"
  else
    printf '%s: not found in %s\n' "$doc" "$REPO_ROOT"
  fi
done

# Check J: External Hyperlinks
echo "--- Check J: External Hyperlinks ---"
grep -rnE 'https?://' "$SKILL_DIR" || echo "None found (Clean)"

echo "=================================================="
if [ "$errors" -gt 0 ]; then
  echo "Check Result: FAILED ($errors error(s) found)"
  exit 1
else
  echo "Check Result: PASSED"
  exit 0
fi
