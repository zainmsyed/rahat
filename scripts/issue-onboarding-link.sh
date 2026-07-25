#!/usr/bin/env bash
set -euo pipefail

if [[ -f .env ]]; then
  set -a
  source .env
  set +a
fi

WEB_ORIGIN="${WEB_ORIGIN:-http://localhost:5200}"
INVITE_CODE="${ONBOARDING_INVITE_CODE:-rahat-beta}"

printf '%s/onboarding?invite=%s\n' "${WEB_ORIGIN%/}" "$INVITE_CODE"
