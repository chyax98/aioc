# AIOC JSONL events

AIOC stdout is newline-delimited JSON. stderr is diagnostics/logs. Automation must wait for the final `done` event.

Success:

```text
done.status == "completed"
```

Failure:

```text
done.status != "completed"
process exits non-zero before a completed done event
invalid/missing JSONL
```

Event types:

```text
session     run started; includes provider/path/protocol/cwd/model metadata
status      provider status/log bridge
text        assistant text delta
thinking    reasoning/thinking delta when provider exposes it
tool_use    tool call start
tool_result tool call result
error       provider or AIOC error
done        final event; contains status/output/error/session_id/duration_ms/usage
```

`done.usage` mirrors Multica's `agent.Result.Usage` map keyed by model name:

```json
{
  "type": "done",
  "provider": "claude",
  "status": "completed",
  "output": "...",
  "duration_ms": 1234,
  "usage": {
    "claude-sonnet-4-6": {
      "input_tokens": 100,
      "output_tokens": 20,
      "cache_read_tokens": 0,
      "cache_write_tokens": 0
    }
  }
}
```
