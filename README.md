# aioc

One command for local AI agents.

AIOC is a standalone Go CLI that runs local agent CLIs through one interface and emits one stable JSONL event stream. Higher-level tools can call `aioc` as a subprocess and stop caring whether the underlying provider is Claude, Codex, Pi, ACP, or another CLI shape.

## Providers

AIOC includes adapters for all providers copied from the Multica worker layer. Provider detection follows Multica daemon semantics: explicit `AIOC_<PROVIDER>_PATH`, `PATH`, login-shell fallback for GUI/non-interactive PATH gaps, and Codex Desktop's bundled CLI path.

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
AIOC_AGENT_TIMEOUT=0
AIOC_CODEX_SEMANTIC_INACTIVITY_TIMEOUT=10m
AIOC_AGENT_IDLE_WATCHDOG=0
AIOC_AGENT_TOOL_WATCHDOG=0
AIOC_AUTO_PRIORITY=claude,codex,pi
AIOC_THINKING_LEVEL=high
AIOC_MCP_CONFIG=.aioc/mcp.json
AIOC_CLAUDE_PATH=/path/to/claude
AIOC_CODEX_PATH=/path/to/codex
AIOC_PI_PATH=/path/to/pi
AIOC_<PROVIDER>_PATH=/path/to/bin
AIOC_<PROVIDER>_MODEL=model
AIOC_<PROVIDER>_THINKING_LEVEL=high
AIOC_<PROVIDER>_MAX_TURNS=20
AIOC_<PROVIDER>_ARGS='--flag value'
AIOC_<PROVIDER>_MCP_CONFIG=.aioc/provider-mcp.json
AIOC_<PROVIDER>_ENV_KEY=value
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
--timeout <duration>   provider run timeout; default AIOC_AGENT_TIMEOUT
--inactivity-timeout <duration>
                       semantic inactivity timeout; default AIOC_CODEX_SEMANTIC_INACTIVITY_TIMEOUT
--idle-watchdog <duration>
                       force-stop when backend emits no messages; default AIOC_AGENT_IDLE_WATCHDOG
--tool-watchdog <duration>
                       force-stop when one tool stays in flight silently; default AIOC_AGENT_TOOL_WATCHDOG
--thinking <level>     reasoning/thinking level; validated against local model catalog when possible
--max-turns <n>        max turns for providers that support it
--mcp-config <path>    MCP config JSON file
--extra-arg <arg>      provider default arg appended before --arg; repeatable
--arg <arg>            provider custom arg appended after config args; repeatable
--env KEY=VALUE        env var for child agent; repeatable
--jsonl                emit JSONL events on stdout; default true
```

## Config discovery

AIOC loads home config first, then project config from git root toward the current working directory. Later files override earlier files; real environment variables override all files. Use `--config <path>`, `AIOC_CONFIG=<path>`, `--no-config`, or `AIOC_NO_CONFIG=1` to control loading.

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
thinking_level: high
agent_timeout: 0
codex_semantic_inactivity_timeout: 10m
agent_idle_watchdog: 0
agent_tool_watchdog: 0
auto_priority: [claude, codex, pi]
mcp_config: .aioc/mcp.json
env:
  AIOC_EXTRA: value
providers:
  claude:
    path: /path/to/claude
    model: claude-sonnet-4-6
    thinking_level: high
    max_turns: 20
    args: ["--max-budget-usd", "1.00"]
    child_env:
      CLAUDE_EXAMPLE: value
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
aioc config --json              # redacted by default
aioc config --json --show-secrets
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
{"type":"done","provider":"claude","status":"completed","output":"...","duration_ms":1234,"usage":{"claude-sonnet-4-6":{"input_tokens":100,"output_tokens":20}}}
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

AIOC pins Go with Proto:

```text
.prototools -> go = "1.26.1"
```

Use system Go:

```bash
make test
make build
make check
```

Use Proto-pinned Go:

```bash
proto install
make proto-check
# or: proto exec go -- make check
```

Smoke real providers:

```bash
make smoke                         # defaults: claude codex pi, skips unavailable
AIOC_SMOKE_PROVIDERS="claude pi" make smoke
```

Project layout:

```text
cmd/aioc/          CLI entry
internal/cli/      command implementation
internal/config/   config/env discovery and YAML/env translation
docs/              config and JSONL event notes
scripts/           real-provider smoke harness
pkg/agent/         copied provider adapters and protocol clients
pkg/events/        normalized event schema + JSONL writer
pkg/provider/      provider registry and binary detection
```
