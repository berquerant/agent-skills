#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"

mode="${1:-check}"

list_files() {
  git -C "${repo_root}" ls-files -z '*.sh'
}

if [ -z "$(list_files | head -c 1)" ]; then
  echo "No shell scripts found."
  exit 0
fi

cd "${repo_root}"

case "${mode}" in
  check | lint)
    echo "Running shellcheck..."
    list_files | xargs -0 shellcheck
    echo "Running shfmt -d..."
    list_files | xargs -0 shfmt -i 2 -ci -d
    echo "All lint checks passed."
    ;;
  fmt | format)
    echo "Running shfmt -w..."
    list_files | xargs -0 shfmt -i 2 -ci -w
    echo "Formatting completed."
    ;;
  *)
    echo "Usage: $0 [check|lint|fmt|format]" >&2
    exit 1
    ;;
esac
