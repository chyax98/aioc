#!/usr/bin/env bash
set -euo pipefail

aioc_bin=${AIOC_BIN:-./bin/aioc}
providers=${AIOC_SMOKE_PROVIDERS:-claude codex pi}

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT

"$aioc_bin" agents --versions
"$aioc_bin" agents --json >"$tmp"
for p in $providers; do
  if python3 - "$tmp" "$p" <<'PY'
import json, sys
path, provider = sys.argv[1], sys.argv[2]
data = json.load(open(path, encoding='utf-8'))
for row in data:
    if row.get("provider") == provider and row.get("status") == "available":
        raise SystemExit(0)
raise SystemExit(1)
PY
  then
    scripts/smoke-provider.sh "$p"
  else
    echo "$p: skip unavailable" >&2
  fi
done
