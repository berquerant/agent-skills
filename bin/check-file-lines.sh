#!/usr/bin/env bash
set -euo pipefail

readonly max_lines="${1:-400}"
readonly target_dir="${2:-mcp}"

failed=0
while IFS= read -r -d '' file; do
  lines="$(wc -l <"${file}" | tr -d ' ')"
  if ((lines > max_lines)); then
    printf 'file line limit exceeded: %s has %d lines (max %d)\n' "${file}" "${lines}" "${max_lines}" >&2
    failed=1
  fi
done < <(find "${target_dir}" -type f -name '*.go' -print0 | sort -z)

exit "${failed}"
