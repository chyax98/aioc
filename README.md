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

Binary path overrides:

```bash
AIOC_CLAUDE_PATH=/path/to/claude
AIOC_CODEX_PATH=/path/to/codex
AIOC_PI_PATH=/path/to/pi
AIOC_GEMINI_PATH=/path/to/gemini
AIOC_CURSOR_PATH=/path/to/cursor-agent
AIOC_KIMI_PATH=/path/to/kimi
AIOC_KIRO_PATH=/path/to/kiro-cli
AIOC_HERMES_PATH=/path/to/hermes
AIOC_OPENCODE_PATH=/path/to/opencode
AIOC_OPENCLAW_PATH=/path/to/openclaw
AIOC_COPILOT_PATH=/path/to/copilot
AIOC_ANTIGRAVITY_PATH=/path/to/agy
```

## Commands

```bash
aioc agents
aioc agents --json
aioc agents --versions
aioc doctor
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
-p <provider>          provider name, or auto
--cwd <dir>            working directory
--model <model>        provider model
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
