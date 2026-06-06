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

  aioc <prompt>
      Run prompt with auto-selected provider.

  aioc <provider> <prompt>
      Provider shorthand, e.g. aioc claude "review current changes".

  aioc run [flags] <prompt>
      Run prompt and stream normalized JSONL events. Provider defaults to auto.

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

  -p <provider>          provider name, or auto; default auto
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

JSONL event types:

  session, status, text, thinking, tool_use, tool_result, error, done

Automation rule:

  Parse stdout as JSONL. Logs are stderr. Wait for final done event.
  Treat done.status != "completed" as failure.

Examples:

  aioc agents --json
  aioc doctor
  aioc models -p claude
  aioc "用默认 agent 回答 hello"
  aioc claude --cwd . "review current changes"
  aioc codex --cwd /repo --timeout 20m "fix tests"
  aioc pi --model openai/gpt-5.5 "用中文总结当前目录"
  aioc run --cwd . --prompt-file task.md

Path overrides:

  AIOC_CLAUDE_PATH=/path/to/claude
  AIOC_CODEX_PATH=/path/to/codex
  AIOC_PI_PATH=/path/to/pi
  AIOC_<PROVIDER>_PATH=/path/to/bin

Usage shadow:

  aioc skills get      Print minimal SKILL.md shadow.
  aioc skills install  Copy minimal shadow to ~/.pi/agent/skills/aioc-cli/SKILL.md.

The shadow is only a discovery hint. This command output is the source of truth.
`

const agentPrompt = `You can use AIOC to call local AI agent CLIs through one stable interface.

Use this workflow:

1. Inspect installed providers:

   aioc agents --json

2. Inspect health when unsure:

   aioc doctor

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

Useful commands:

   aioc usage
   aioc agents --json
   aioc models -p <provider> --json
   aioc "quick question"
   aioc claude --cwd . "review current changes"
   aioc codex --cwd . --timeout 20m "fix failing tests"
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
