#!/usr/bin/env bash
set -euo pipefail

log() {
  echo >&2 "$@"
}

usage() {
  cat <<'EOF' >&2
Usage: bin/install.sh [OPTIONS] [TARGET_DIR]

Installs agent skills from this repository by creating symbolic links
in TARGET_DIR pointing to each skill directory.
Displays the planned actions and requests confirmation before applying.

Arguments:
  TARGET_DIR           Target directory to install skill symlinks.
                       Can also be specified via TARGET_DIR environment variable.
                       If unspecified, prompts interactively.

Options:
  -h, --help           Show this help message and exit.
EOF
}

run_cmd() {
  local dryrun="$1"
  shift
  log "  $*"
  if [ "$dryrun" != "1" ]; then
    "$@"
  fi
}

confirm_proceed() {
  printf "Proceed with creating links? [Y/n]: " >&2
  local answer
  read -r answer
  case "$answer" in
    [yY][eE][sS] | [yY] | "")
      return 0
      ;;
    *)
      log "Aborted."
      exit 0
      ;;
  esac
}

apply_plans() {
  local dryrun="$1"
  local target="$2"
  shift 2
  local skills=("$@")

  run_cmd "$dryrun" mkdir -p "$target"
  for skill_path in "${skills[@]}"; do
    local skill_name
    skill_name="$(basename "${skill_path}")"
    run_cmd "$dryrun" ln -sfn "${skill_path}" "${target}/${skill_name}"
  done
}

main() {
  local target=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h | --help)
        usage
        exit 0
        ;;
      -*)
        log "Error: Unknown option '$1'"
        usage
        exit 1
        ;;
      *)
        if [[ -z "$target" ]]; then
          target="$1"
        else
          log "Error: Unexpected argument '$1'"
          usage
          exit 1
        fi
        shift
        ;;
    esac
  done

  if [[ -z "$target" ]]; then
    target="${TARGET_DIR:-}"
  fi

  if [[ -z "$target" ]]; then
    printf "Enter target directory for skills installation: " >&2
    if read -r input_target; then
      target="$input_target"
    fi
  fi

  if [[ "$target" =~ ^~(/.*)?$ ]]; then
    target="${HOME}${target#"~"}"
  fi

  if [[ -z "$target" ]]; then
    log "Error: Target directory is not specified."
    exit 1
  fi

  local script_dir repo_root skills_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  repo_root="$(cd "${script_dir}/.." && pwd)"
  skills_dir="${repo_root}/skills"

  local skills=()
  for skill_path in "${skills_dir}"/*; do
    if [ -d "${skill_path}" ] && [ -f "${skill_path}/SKILL.md" ]; then
      skills+=("${skill_path}")
    fi
  done

  if [ ${#skills[@]} -eq 0 ]; then
    log "No valid skills found in ${skills_dir}."
    exit 0
  fi

  # 1. Preview planned actions
  log "Planned actions:"
  apply_plans 1 "$target" "${skills[@]}"

  # 2. User confirmation
  confirm_proceed

  # 3. Apply actions
  log "Executing:"
  apply_plans 0 "$target" "${skills[@]}"

  log "Successfully installed ${#skills[@]} skills to ${target}."
}

main "$@"
