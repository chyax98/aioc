#!/usr/bin/env bash
set -euo pipefail

provider=${1:?usage: smoke-provider.sh <provider> [prompt]}
prompt=${2:-"Reply with exactly: ${provider} ok"}
aioc_bin=${AIOC_BIN:-./bin/aioc}
timeout_arg=${AIOC_SMOKE_TIMEOUT:-60s}

out=$(mktemp)
err=$(mktemp)
trap 'rm -f "$out" "$err"' EXIT

set +e
"$aioc_bin" "$provider" --timeout "$timeout_arg" "$prompt" >"$out" 2>"$err"
code=$?
set -e

if ! python3 - "$out" "$provider" "$code" <<'PY'
import json, sys
path, provider, code = sys.argv[1], sys.argv[2], int(sys.argv[3])
lines = [line for line in open(path, encoding='utf-8') if line.strip()]
if not lines:
    raise SystemExit(f"{provider}: no JSONL output, exit={code}")
try:
    events = [json.loads(line) for line in lines]
except Exception as e:
    raise SystemExit(f"{provider}: invalid JSONL: {e}")
done = [e for e in events if e.get("type") == "done"]
if not done:
    raise SystemExit(f"{provider}: missing done event, exit={code}")
last = done[-1]
if last.get("status") != "completed":
    raise SystemExit(f"{provider}: done.status={last.get('status')!r}, error={last.get('error')!r}, exit={code}")
print(f"{provider}: ok session={last.get('session_id','')} duration_ms={last.get('duration_ms',0)}")
PY
then
  echo "--- ${provider} stdout ---" >&2
  tail -80 "$out" >&2 || true
  echo "--- ${provider} stderr ---" >&2
  tail -120 "$err" >&2 || true
  exit 1
fi

if [[ $code -ne 0 ]]; then
  echo "warning: aioc exited non-zero despite completed done: $code" >&2
fi
