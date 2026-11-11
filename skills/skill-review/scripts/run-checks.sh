#!/bin/sh
# run-checks.sh: Runs automated quality checks for a target agent skill.
# Usage: scripts/run-checks.sh [--strict] <path-to-skill-dir>

set -eu

strict=0
if [ "${1:-}" = "--strict" ]; then
  strict=1
  shift
fi

if [ $# -lt 1 ]; then
  echo "Usage: $0 [--strict] <path-to-skill-dir>" >&2
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
warnings=0

echo "=================================================="
echo "Automated Check Report for: $NAME"
echo "Target Directory: $SKILL_DIR"
echo "=================================================="

# Check A: Inventory, Size & Unreferenced Files
echo "--- Check A: Inventory, Line Count & Orphan Files ---"
find "$SKILL_DIR" -type f | sort
wc -l "$SKILL_DIR"/SKILL.md "$SKILL_DIR"/*/*.md 2>/dev/null || true
skill_lines=$(wc -l <"$SKILL_DIR/SKILL.md" | tr -d ' ')
if [ "$skill_lines" -gt 500 ]; then
  echo "size: FAIL ($skill_lines lines, exceeds 500)"
  errors=$((errors + 1))
elif [ "$skill_lines" -lt 80 ] || [ "$skill_lines" -gt 250 ]; then
  echo "size: WARN ($skill_lines lines, recommended 80-250)"
  warnings=$((warnings + 1))
else
  echo "size: OK ($skill_lines lines)"
fi

orphan_files=$(find "$SKILL_DIR_ABS" -type f ! -name 'SKILL.md' | sort | while read -r f; do
  rel=${f#"$SKILL_DIR_ABS"/}
  base=$(basename "$f")
  if grep -Fq -- "$rel" "$SKILL_DIR_ABS/SKILL.md" 2>/dev/null; then
    continue
  fi
  refs=$(find "$SKILL_DIR_ABS" -name '*.md' ! -path "$f" -exec grep -Fl -- "$base" {} + 2>/dev/null || true)
  if [ -z "$refs" ]; then
    echo "WARN: unreferenced file: $rel"
  else
    echo "WARN: multi-hop reference (not linked directly from SKILL.md): $rel"
  fi
done)
if [ -n "$orphan_files" ]; then
  printf '%s\n' "$orphan_files"
  orphan_count=$(printf '%s\n' "$orphan_files" | wc -l | tr -d ' ')
  warnings=$((warnings + orphan_count))
else
  echo "orphan files: OK (all auxiliary files directly referenced from SKILL.md)"
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
name_len=${#fm_name}
echo "name=$fm_name len=$name_len"
if [ -z "$fm_name" ]; then
  echo "pattern: FAIL (missing or empty name)"
  errors=$((errors + 1))
elif [ "$name_len" -gt 64 ]; then
  echo "pattern: FAIL (exceeds 64 chars: $name_len)"
  errors=$((errors + 1))
elif printf '%s\n' "$fm_name" | grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$'; then
  echo "pattern: OK"
else
  echo "pattern: FAIL"
  errors=$((errors + 1))
fi
if [ "$name_len" -lt 3 ] || [ "$name_len" -gt 30 ]; then
  echo "name length: WARN ($name_len chars, recommended 3-30)"
  warnings=$((warnings + 1))
else
  echo "name length: OK ($name_len chars)"
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
elif [ "$desc_len" -lt 100 ] || [ "$desc_len" -gt 350 ]; then
  echo "description: WARN ($desc_len chars, recommended 100-350 for Layer 1 token economy)"
  warnings=$((warnings + 1))
else
  echo "description: OK ($desc_len chars)"
fi
if printf '%s\n' "$desc" | grep -Eq '^Use this skill (when|to|for) '; then
  echo "trigger opener: OK"
else
  echo "trigger opener: WARN (recommended opener: 'Use this skill when/to/for ...')"
  warnings=$((warnings + 1))
fi

# Check D: Section Structure, First-100-Lines & Verify Checkpoints
echo "--- Check D: Headings, Top-Loaded Workflow & Verify Checkpoints ---"
awk '/^```/ { code = !code; next } !code && /^##? / { print NR ": " $0 }' "$SKILL_DIR/SKILL.md"
step1_line=$(awk '
  /^```/ { code = !code; next }
  !code && /^## Step 1([:[:space:]]|$)/ { print NR; exit }
' "$SKILL_DIR/SKILL.md")
if [ -n "$step1_line" ]; then
  if [ "$step1_line" -le 100 ]; then
    echo "top-loaded workflow: OK (Step 1 at line $step1_line <= 100)"
  else
    echo "top-loaded workflow: WARN (Step 1 at line $step1_line > 100)"
    warnings=$((warnings + 1))
  fi
fi
last_h2=$(awk '
  /^```/ { code = !code; next }
  !code && /^## / { last = $0 }
  END { print last }
' "$SKILL_DIR/SKILL.md")
if [ "$last_h2" = "## Guidelines" ]; then
  echo "final section: OK (## Guidelines)"
else
  echo "final section: WARN (expected '## Guidelines' as final ## section, found '${last_h2:-none}')"
  warnings=$((warnings + 1))
fi
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
else
  echo "verify checkpoints: OK"
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
else
  echo "relative links: OK"
fi
abs_links=$(find "$SKILL_DIR" -name '*.md' ! -name 'automated-checks.md' -exec grep -HnE '\]\((/|file://)|/Users/|/home/[a-z]|[A-Z]:[/\\]' {} + 2>/dev/null || true)
if [ -n "$abs_links" ]; then
  echo "ABSOLUTE/MACHINE paths found:"
  printf '%s\n' "$abs_links"
  abs_count=$(printf '%s\n' "$abs_links" | wc -l | tr -d ' ')
  errors=$((errors + abs_count))
else
  echo "path portability: OK"
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
  else
    echo "script permissions: OK"
  fi
else
  echo "script permissions: n/a (no scripts/ directory)"
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
for doc in README.md AGENTS.md skills/skill-router/references/intent-matrix.md; do
  doc_path="$REPO_ROOT/$doc"
  if [ -f "$doc_path" ]; then
    reg_count=$(grep -cF -- "$NAME" "$doc_path" || true)
    printf '%s: %s\n' "$doc" "$reg_count"
    if [ "$reg_count" -eq 0 ] && [ "$doc" != "skills/skill-router/references/intent-matrix.md" ]; then
      echo "WARN: $NAME is not registered in $doc"
      warnings=$((warnings + 1))
    elif [ "$reg_count" -eq 0 ] && [ "$NAME" != "skill-router" ]; then
      echo "WARN: $NAME is not listed in $doc"
      warnings=$((warnings + 1))
    fi
  fi
done

# Check I: Cross-Skill Duplicate Reference Headings
echo "--- Check I: Cross-Skill Duplicate Reference Headings ---"
if [ -d "$SKILL_DIR/references" ]; then
  dup_headings=$(awk -v target="$NAME" '
    /^###? / {
      n = FILENAME;
      sub(/\/references\/.*/, "", n);
      sub(/.*\//, "", n);
      if (index(s[$0], " " n) == 0) {
        s[$0] = s[$0] " " n;
        c[$0]++;
      }
      if (n == target) {
        in_target[$0] = 1;
      }
    }
    END {
      for (h in c) {
        if (c[h] > 1 && in_target[h]) {
          print h " ->" s[h];
        }
      }
    }
  ' "$SKILLS_ROOT"/*/references/*.md 2>/dev/null | sort || true)
  if [ -n "$dup_headings" ]; then
    printf '%s\n' "$dup_headings"
  else
    echo "None found (Clean)"
  fi
else
  echo "n/a (no references/ directory)"
fi

# Check J: External Hyperlinks
echo "--- Check J: External Hyperlinks ---"
grep -rnE 'https?://' "$SKILL_DIR" || echo "None found (Clean)"

# Check K: Inline Shell Command Modularity & Prompt Hygiene
echo "--- Check K: Command & Prompt Hygiene ---"
hygiene_warns=$(find "$SKILL_DIR" -name '*.md' ! -name 'automated-checks.md' ! -name 'skill-review-checklist.md' | sort | while read -r f; do
  awk -v file="$f" '
    /^[[:space:]]*```(sh|bash|zsh)[[:space:]]*$/ { in_sh = 1; next }
    /^[[:space:]]*```/ { in_code = !in_code; in_sh = 0; next }
    in_sh && $0 !~ /^[[:space:]]*#/ && length($0) >= 120 {
      printf "WARN: long inline shell command (%d chars >= 120) in %s:%d\n", length($0), file, NR
    }
    !in_code && /\]\([^)]+\)[[:space:]]+\(`[^`]+`\)/ {
      printf "WARN: backtick-wrapped raw target in co-presented link in %s:%d\n", file, NR
    }
    !in_code && tolower($0) ~ /think step[- ]by[- ]step/ {
      printf "WARN: manual chain-of-thought phrasing in %s:%d\n", file, NR
    }
  ' "$f"
done)
if [ -n "$hygiene_warns" ]; then
  printf '%s\n' "$hygiene_warns"
  hygiene_count=$(printf '%s\n' "$hygiene_warns" | wc -l | tr -d ' ')
  warnings=$((warnings + hygiene_count))
else
  echo "command & prompt hygiene: OK (concise shell examples, unadorned raw link targets)"
fi

echo "=================================================="
if [ "$errors" -gt 0 ]; then
  echo "Check Result: FAILED ($errors error(s), $warnings warning(s))"
  exit 1
elif [ "$warnings" -gt 0 ]; then
  echo "Check Result: PASSED ($errors error(s), $warnings warning(s))"
  if [ "$strict" -eq 1 ]; then
    exit 1
  fi
  exit 0
else
  echo "Check Result: PASSED (0 errors, 0 warnings)"
  exit 0
fi
