# aioc

One command for local AI agents.

AIOC is a standalone Go CLI that runs local agent CLIs through one interface and emits one stable JSONL event stream. Higher-level tools can call `aioc` as a subprocess and stop caring whether the underlying provider is Claude, Codex, Pi, ACP, or another CLI shape.

## Providers

AIOC includes adapters for all providers copied from the Multica worker layer:

| Provider | Command | Transport |
|---|---|---|
| Claude | `claude` | stream JSON |
| Codex | `codex` | native app-server / JSON-RPC |
| Pi | `pi` | JSONL/stream shim |
| Gemini | `gemini` | stream JSON |
| Cursor | `cursor-agent` | stream JSON |
| Kimi | `kimi acp` | ACP |
| Kiro | `kiro-cli acp` | ACP |
| Hermes | `hermes acp` | ACP |
| OpenCode | `opencode` | JSON stream |
| OpenClaw | `openclaw` | JSON stream |
| Copilot | `copilot` | JSON stream |
| Antigravity | `agy` | plain text print mode |

AIOC auto-loads config/env files before every command. Real shell environment variables override config files.

Common env keys:

```bash
AIOC_DEFAULT_PROVIDER=claude
AIOC_DEFAULT_MODEL=claude-sonnet-4-6
AIOC_CLAUDE_PATH=/path/to/claude
AIOC_CODEX_PATH=/path/to/codex
AIOC_PI_PATH=/path/to/pi
AIOC_<PROVIDER>_PATH=/path/to/bin
AIOC_<PROVIDER>_MODEL=model
```

## Commands

```bash
aioc agents
aioc agents --json
aioc agents --versions
aioc doctor
aioc config
aioc config --json
aioc models -p claude
aioc models -p gemini --json
aioc models-all --json
aioc usage
aioc prompt
aioc skills get
aioc skills install
aioc "hello"
aioc claude "hello"
aioc codex --cwd /repo "fix tests"
aioc pi --model openai/gpt-5.1 "review this repo"
aioc run -p claude "explicit provider form"
```

## Run flags

```text
-p <provider>          provider name, or auto; default AIOC_DEFAULT_PROVIDER or auto
--cwd <dir>            working directory
--model <model>        provider model; default AIOC_DEFAULT_MODEL or AIOC_<PROVIDER>_MODEL
--system <text>        system prompt
--system-file <path>   append system prompt from file
--prompt-file <path>   append prompt from file
--resume <id>          resume provider session
--timeout <duration>   provider run timeout
--mcp-config <path>    MCP config JSON file
--arg <arg>            provider-specific extra arg; repeatable
--env KEY=VALUE        env var for child agent; repeatable
--jsonl                emit JSONL events on stdout; default true
```

## Config discovery

AIOC loads home config first, then project config from repository root toward the current working directory. Later files override earlier files; real environment variables override all files.

Home:

```text
~/.config/aioc/config.env
~/.config/aioc/config.yaml
~/.aioc/config.env
~/.aioc/config.yaml
```

Project tree:

```text
.env
.aioc.config.env
.aioc/config.env
.aioc/config.yaml
.config/aioc/config.env
.config/aioc/config.yaml
```

YAML example:

```yaml
default_provider: claude
default_model: claude-sonnet-4-6
env:
  AIOC_EXTRA: value
providers:
  claude:
    path: /path/to/claude
    model: claude-sonnet-4-6
  codex:
    path: /path/to/codex
    model: gpt-5.5
  openai:
    base_url: http://127.0.0.1:8317/v1
    api_key: your-api-key-1
```

Inspect loaded state:

```bash
aioc config
aioc config --json
```

## JSONL event contract

AIOC writes machine-readable JSONL to stdout. Logs go to stderr.

```jsonl
{"type":"session","provider":"claude","status":"starting"}
{"type":"status","provider":"claude","status":"running"}
{"type":"text","provider":"claude","delta":"..."}
{"type":"thinking","provider":"claude","delta":"..."}
{"type":"tool_use","provider":"claude","tool_id":"...","tool_name":"bash","input":{}}
{"type":"tool_result","provider":"claude","tool_id":"...","tool_name":"bash","output":"..."}
{"type":"error","provider":"claude","error":"..."}
{"type":"done","provider":"claude","status":"completed","output":"...","duration_ms":1234}
```

## Self-describing usage

AIOC's CLI is the source of truth for usage.

```bash
aioc usage   # human/agent readable guide
aioc prompt  # prompt for an agent that needs to use AIOC deeply
aioc --help  # compact command list
```

## Usage shadow

AIOC ships a minimal companion skill only as a discovery shadow. It tells agents to run `aioc usage` / `aioc prompt`; it is not the full manual.

Print shadow:

```bash
aioc skills get
```

Copy shadow to Pi global skills:

```bash
aioc skills install
# writes ~/.pi/agent/skills/aioc-cli/SKILL.md
```

Copy shadow elsewhere:

```bash
aioc skills install --dir ~/.agents/skills
```

## Development

```bash
make test
make build
make check
```

Project layout:

```text
cmd/aioc/          CLI entry
internal/cli/      command implementation
pkg/agent/         copied provider adapters and protocol clients
pkg/events/        normalized event schema + JSONL writer
pkg/provider/      provider registry and binary detection
```
