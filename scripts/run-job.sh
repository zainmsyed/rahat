#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: scripts/run-job.sh <job-name>" >&2
  exit 1
fi

if [[ -f .env ]]; then
  set -a
  source .env
  set +a
fi

GO_BIN="${GO_BIN:-go}"
RAHAT_BIN="${RAHAT_BIN:-/usr/local/bin/rahat-api}"

if command -v "$GO_BIN" >/dev/null 2>&1; then
  "$GO_BIN" run ./cmd/server ops:run-job "$1"
elif [[ -x "$RAHAT_BIN" ]]; then
  "$RAHAT_BIN" ops:run-job "$1"
else
  echo "go not found and no rahat-api binary at $RAHAT_BIN" >&2
  exit 1
fi
