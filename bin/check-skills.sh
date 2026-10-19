#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
checker="${repo_root}/skills/skill-review/scripts/run-checks.sh"

cd "${repo_root}"

failed=0

for skill_dir in skills/*; do
  [ -d "${skill_dir}" ] || continue
  [ -f "${skill_dir}/SKILL.md" ] || continue

  output="$("${checker}" "${skill_dir}" 2>&1)"
  if echo "${output}" | grep -Eq 'FAIL|BROKEN|MISSING|NON-EXECUTABLE'; then
    echo "========================================" >&2
    echo "Check failed for ${skill_dir}:" >&2
    echo "${output}" | grep -E 'FAIL|BROKEN|MISSING|NON-EXECUTABLE' >&2
    echo "========================================" >&2
    failed=1
  fi
done

if [ "${failed}" -ne 0 ]; then
  echo "Skill consistency checks failed." >&2
  exit 1
fi

echo "All skill consistency checks passed."
