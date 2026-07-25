#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: scripts/issue-beta-access.sh <user-id-or-email>" >&2
  exit 1
fi

if [[ -f .env ]]; then
  set -a
  source .env
  set +a
fi

GO_BIN="${GO_BIN:-go}"
if ! command -v "$GO_BIN" >/dev/null 2>&1; then
  if [[ -x /home/zain/.local/go/bin/go ]]; then
    GO_BIN=/home/zain/.local/go/bin/go
  else
    echo "go command not found; set GO_BIN=/path/to/go" >&2
    exit 1
  fi
fi

"$GO_BIN" run ./cmd/server ops:issue-access-link "$1"
