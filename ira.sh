#!/usr/bin/env bash
set -euo pipefail

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
exec bash "$repo_dir/skills/incubator-release-assistant/scripts/run.sh" "$@"
