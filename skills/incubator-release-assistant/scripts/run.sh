#!/usr/bin/env bash
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
engine_dir="$script_dir/ira"
caller_dir=$(pwd)
args=()
secret_dir=""

while (($#)); do
  if [[ "$1" == "--secret-dir" ]]; then
    shift
    if (($# == 0)); then
      echo "--secret-dir requires a path" >&2
      exit 2
    fi
    if [[ "$1" = /* ]]; then
      secret_dir="$1"
    else
      secret_dir="$caller_dir/$1"
    fi
  elif [[ "$1" == "--config" ]]; then
    args+=("$1")
    shift
    if (($# == 0)); then
      echo "--config requires a path" >&2
      exit 2
    fi
    if [[ "$1" = /* ]]; then
      args+=("$1")
    else
      args+=("$caller_dir/$1")
    fi
  else
    args+=("$1")
  fi
  shift
done

cache_root=${TMPDIR:-/tmp}
export GOCACHE="$cache_root/incubator-release-assistant-go-build-cache"
export GOTELEMETRY=off

if [[ "${args[0]:-}" == "sign" ]]; then
  if [[ -z "$secret_dir" ]]; then
    secret_dir="$caller_dir/secretkey"
  fi
  repository_root=$(CDPATH= cd -- "$script_dir/../../.." && pwd -P)
  secret_parent=$(dirname -- "$secret_dir")
  secret_name=$(basename -- "$secret_dir")
  if [[ ! -d "$secret_parent" ]]; then
    echo "the parent of the secret directory must already exist: $secret_parent" >&2
    exit 2
  fi
  secret_parent=$(CDPATH= cd -- "$secret_parent" && pwd -P)
  secret_dir="$secret_parent/$secret_name"
  case "$secret_dir/" in
    "$repository_root/"*)
      echo "secret directory must be outside the repository; run from its parent workspace or pass --secret-dir" >&2
      exit 2
      ;;
  esac
  git_probe="$secret_dir"
  while :; do
    if [[ -e "$git_probe/.git" ]]; then
      echo "secret directory must not be inside any Git worktree: $git_probe" >&2
      exit 2
    fi
    git_parent=$(dirname -- "$git_probe")
    if [[ "$git_parent" == "$git_probe" ]]; then
      break
    fi
    git_probe="$git_parent"
  done
  mkdir -p -- "$secret_dir"
  secret_dir=$(CDPATH= cd -- "$secret_dir" && pwd -P)
  case "$secret_dir/" in
    "$repository_root/"*)
      echo "resolved secret directory must be outside the repository" >&2
      exit 2
      ;;
  esac
  git_probe="$secret_dir"
  while :; do
    if [[ -e "$git_probe/.git" ]]; then
      echo "secret directory must not be inside any Git worktree: $git_probe" >&2
      exit 2
    fi
    git_parent=$(dirname -- "$git_probe")
    if [[ "$git_parent" == "$git_probe" ]]; then
      break
    fi
    git_probe="$git_parent"
  done
  chmod 700 "$secret_dir"
  export IRA_SECRET_DIR="$secret_dir"
  export GNUPGHOME="$secret_dir"
fi

exec go -C "$engine_dir" run ./cmd/ira "${args[@]}"
