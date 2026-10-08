#!/bin/sh
# run-checks.sh: Runs automated quality checks for a target agent skill.
# Usage: scripts/run-checks.sh <path-to-skill-dir>

set -eu

if [ $# -lt 1 ]; then
  echo "Usage: $0 <path-to-skill-dir>" >&2
  exit 1
fi

SKILL_DIR="$1"
SKILLS_ROOT=$(dirname "$SKILL_DIR")
NAME=$(basename "$SKILL_DIR")

if [ ! -f "$SKILL_DIR/SKILL.md" ]; then
  echo "Error: $SKILL_DIR/SKILL.md does not exist." >&2
  exit 1
fi

echo "=================================================="
echo "Automated Check Report for: $NAME"
echo "Target Directory: $SKILL_DIR"
echo "=================================================="

# Check A: Inventory & Size
echo "--- Check A: Inventory & Line Count ---"
find "$SKILL_DIR" -type f | sort
wc -l "$SKILL_DIR"/SKILL.md "$SKILL_DIR"/*/*.md 2>/dev/null || true

# Check B: Frontmatter name
echo "--- Check B: Frontmatter 'name' ---"
fm_name=$(awk '/^---[[:space:]]*$/{n++; next} n==1 && /^name:/{sub(/^name:[[:space:]]*/, ""); print; exit}' "$SKILL_DIR/SKILL.md")
echo "name=$fm_name len=${#fm_name}"
if printf '%s\n' "$fm_name" | grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$'; then
  echo "pattern: OK"
else
  echo "pattern: FAIL"
fi
if [ "$fm_name" = "$NAME" ]; then
  echo "dir match: OK"
else
  echo "dir match: FAIL (expected $NAME)"
fi

# Check C: Frontmatter description
echo "--- Check C: Frontmatter 'description' ---"
desc=$(awk '
  /^---[[:space:]]*$/ { n++; next }
  n == 1 && /^description:/ { f = 1; sub(/^description:[[:space:]]*([>|][-+]?)?[[:space:]]*/, ""); if ($0 != "") printf "%s ", $0; next }
  n == 1 && f && /^[^[:space:]]/ { f = 0 }
  n == 1 && f { sub(/^[[:space:]]+/, ""); printf "%s ", $0 }
' "$SKILL_DIR/SKILL.md" | sed 's/[[:space:]]*$//')
printf 'description: %s\n' "$desc"
desc_len=$(printf '%s' "$desc" | wc -c | tr -d ' ')
echo "char count: $desc_len"

# Check D: Section Structure & Verify Checkpoints
echo "--- Check D: Headings & Verify Checkpoints ---"
awk '/^```/ { code = !code; next } !code && /^##? / { print NR ": " $0 }' "$SKILL_DIR/SKILL.md"
awk '
  /^```/ { code = !code }
  code { next }
  /^## / { if (step != "" && !v) print "MISSING Verify: " step; step = ($0 ~ /^## Step /) ? $0 : ""; v = 0; next }
  /Verify:/ { v = 1 }
  END { if (step != "" && !v) print "MISSING Verify: " step }
' "$SKILL_DIR/SKILL.md"

# Check E: Links
echo "--- Check E: Relative & Absolute Links ---"
find "$SKILL_DIR" -name '*.md' | while read -r f; do
  d=$(dirname "$f")
  awk '/^[[:space:]]*```/ { code = !code; next } !code' "$f" |
    grep -o '](\([^)]*\))' | sed 's/^](//; s/)$//; s/#.*//; s/[[:space:]].*//' |
    grep -Ev '^(https?:|mailto:|$)' | while read -r p; do
    [ -e "$d/$p" ] || echo "BROKEN link: $f -> $p"
  done
done
grep -rnE '\]\((/|file://)|/Users/|/home/[a-z]|[A-Z]:[/\\]' "$SKILL_DIR" || true

# Check F: Script Permissions
echo "--- Check F: Script Permissions ---"
if [ -d "$SKILL_DIR/scripts" ]; then
  find "$SKILL_DIR/scripts" -type f ! -perm -u+x | while read -r s; do
    echo "NON-EXECUTABLE script: $s"
  done
fi

# Check G: Sibling References
echo "--- Check G: Sibling Skill References ---"
for s in "$SKILLS_ROOT"/*/; do
  s=$(basename "$s")
  [ "$s" = "$NAME" ] && continue
  c=$(grep -rF -- "$s" "$SKILL_DIR" 2>/dev/null | wc -l | tr -d ' ')
  [ "$c" -gt 0 ] && echo "$s: $c"
done
grep -rnE '\.\./[a-z0-9-]+/' "$SKILL_DIR" || true

# Check H: Repository Registration
echo "--- Check H: Repository Registration ---"
for doc in README.md AGENTS.md; do
  if [ -f "$doc" ]; then
    printf '%s: %s\n' "$doc" "$(grep -cF -- "$NAME" "$doc" || true)"
  fi
done

# Check J: External Hyperlinks
echo "--- Check J: External Hyperlinks ---"
grep -rnE 'https?://' "$SKILL_DIR" || echo "None found (Clean)"
