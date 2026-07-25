#!/usr/bin/env bash
set -euo pipefail
if [[ -f .env ]]; then
  set -a
  source .env
  set +a
fi
: "${RAHAT_RESET_CONFIRM:=reset-non-production}"
export RAHAT_RESET_CONFIRM
go run ./cmd/server ops:reset-nonprod
