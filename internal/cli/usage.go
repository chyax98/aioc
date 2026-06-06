package cli

import "fmt"

const usageGuide = `# AIOC Usage

AIOC is a local AI-agent command router. It runs installed provider CLIs and emits one JSONL event stream on stdout.

Core commands:

  aioc agents [--json] [--versions]
      Detect installed providers.

  aioc doctor
      Print environment/provider health.

  aioc models -p <provider> [--json]
      List one provider's model catalog.

  aioc models-all [--json]
      List every provider's model catalog.

  aioc config [--json] [--show-secrets]
      Show discovered config files and effective AIOC_* settings. JSON is redacted unless --show-secrets.

  aioc runs [list]
      List saved run ids. Runs are saved as JSONL under AIOC_RUNS_DIR or ~/.aioc/runs.

  aioc runs output [latest|id]
      Print the final done.output from a saved run.

  aioc runs events [latest|id]
      Print the raw saved JSONL event stream.

  aioc <prompt>
      Run prompt with auto-selected provider.

  aioc <provider> <prompt>
      Provider shorthand, e.g. aioc claude "review current changes".

  aioc run [flags] <prompt>
      Run prompt and stream normalized JSONL events. Provider defaults to config or auto.

  aioc run -p <provider> [flags] <prompt>
      Explicit provider form.

  aioc usage
      Print this usage guide.

  aioc prompt
      Print agent-facing prompt for using AIOC deeply.

Providers:

  claude, codex, pi, gemini, cursor, kimi, kiro, hermes,
  opencode, openclaw, copilot, antigravity, auto

Run flags:

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

Run history:

  Each run tees stdout JSONL into ~/.aioc/runs/<run_id>.jsonl by default.
  The session event meta includes run_id and events_path when saving succeeds.
  Disable with AIOC_SAVE_RUNS=0. Override directory with AIOC_RUNS_DIR=/path.

JSONL event types:

  session, status, text, thinking, tool_use, tool_result, error, done

Automation rule:

  Parse stdout as JSONL. Logs are stderr. Wait for final done event.
  Treat done.status != "completed" as failure.

Examples:

  aioc agents --json
  aioc doctor
  aioc config
  aioc runs output latest
  aioc models -p claude
  aioc "用默认 agent 回答 hello"
  aioc claude --cwd . "review current changes"
  aioc codex --cwd /repo --timeout 20m "fix tests"
  aioc pi --model openai/gpt-5.5 "用中文总结当前目录"
  aioc run --cwd . --prompt-file task.md

Config auto-discovery:

  AIOC loads config before every command. Real environment variables win over files.
  Later local files override earlier/home files.

  Home:
    ~/.config/aioc/config.env
    ~/.config/aioc/config.yaml
    ~/.aioc/config.env
    ~/.aioc/config.yaml

  Project tree, from repo root toward cwd:
    .env
    .aioc.config.env
    .aioc/config.env
    .aioc/config.yaml
    .config/aioc/config.env
    .config/aioc/config.yaml

Env keys:

  AIOC_DEFAULT_PROVIDER=claude
  AIOC_DEFAULT_MODEL=...
  AIOC_CLAUDE_PATH=/path/to/claude
  AIOC_CODEX_PATH=/path/to/codex
  AIOC_PI_PATH=/path/to/pi
  AIOC_AGENT_TIMEOUT=0
  AIOC_CODEX_SEMANTIC_INACTIVITY_TIMEOUT=10m
  AIOC_AGENT_IDLE_WATCHDOG=0
  AIOC_AGENT_TOOL_WATCHDOG=0
  AIOC_AUTO_PRIORITY=claude,codex,pi
  AIOC_THINKING_LEVEL=high
  AIOC_MCP_CONFIG=.aioc/mcp.json
  AIOC_RUNS_DIR=~/.aioc/runs
  AIOC_SAVE_RUNS=1
  AIOC_<PROVIDER>_PATH=/path/to/bin
  AIOC_<PROVIDER>_MODEL=model
  AIOC_<PROVIDER>_THINKING_LEVEL=high
  AIOC_<PROVIDER>_MAX_TURNS=20
  AIOC_<PROVIDER>_ARGS='--flag value'
  AIOC_<PROVIDER>_MCP_CONFIG=.aioc/provider-mcp.json
  AIOC_<PROVIDER>_ENV_KEY=value

YAML example:

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

Usage shadow:

  aioc skills get      Print minimal SKILL.md shadow.
  aioc skills install  Copy minimal shadow to ~/.pi/agent/skills/aioc-cli/SKILL.md.

The shadow is only a discovery hint. This command output is the source of truth.
`

const agentPrompt = `You can use AIOC to call local AI agent CLIs through one stable interface.

Use this workflow:

1. Inspect installed providers:

   aioc agents --json

2. Inspect health/config when unsure:

   aioc doctor
   aioc config

3. Pick provider:

   - claude: strong code reasoning/review
   - codex: coding agent / repo work
   - pi: local Pi agent
   - gemini/cursor/kimi/hermes/opencode/openclaw/copilot/antigravity when available
   - auto: first available provider

4. Run task:

   aioc --cwd <repo> "<task>"
   aioc <provider> --cwd <repo> "<task>"
   aioc run -p <provider> --cwd <repo> "<task>"

5. Parse stdout JSONL. Ignore stderr except diagnostics.
   Wait for {"type":"done",...}. Success only when done.status == "completed".
   done.usage contains per-model token counts when provider reports them.
   Each run is also saved to ~/.aioc/runs by default; use aioc runs output latest for final text.

Useful commands:

   aioc usage
   aioc agents --json
   aioc models -p <provider> --json
   aioc runs output latest
   aioc "quick question"
   aioc claude --cwd . "review current changes"
   aioc codex --cwd . --timeout 20m --inactivity-timeout 10m "fix failing tests"
   aioc claude --thinking high --max-turns 20 "investigate"
   aioc pi --cwd . "summarize repo in Chinese"

Event schema:

   session/status/text/thinking/tool_use/tool_result/error/done

Never guess provider-native flags first. Prefer AIOC. Fall back to native CLI only when AIOC cannot express required behavior.
`

func usage(args []string) error {
	_ = args
	fmt.Print(usageGuide)
	return nil
}

func prompt(args []string) error {
	_ = args
	fmt.Print(agentPrompt)
	return nil
}
