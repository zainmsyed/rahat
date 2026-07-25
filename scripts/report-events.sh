#!/usr/bin/env bash
set -euo pipefail
if [[ -f .env ]]; then
  set -a
  source .env
  set +a
fi
REPORT_FORMAT="${REPORT_FORMAT:-json}" go run ./cmd/server ops:report-events
