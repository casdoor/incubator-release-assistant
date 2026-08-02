#!/usr/bin/env bash
set -euo pipefail

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
engine_dir="$repo_dir/skills/incubator-release-assistant/scripts/ira"
caller_dir=$(pwd)
args=()

while (($#)); do
  if [[ "$1" == "--config" ]]; then
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

export GOCACHE="$repo_dir/.ira/go-build-cache"
exec go -C "$engine_dir" run ./cmd/ira "${args[@]}"
