# AIOC config

AIOC loads config before every command unless disabled.

Priority:

```text
real shell environment > explicit --config/AIOC_CONFIG files > project files > home files > defaults
```

Disable or override:

```bash
aioc --no-config ...
AIOC_NO_CONFIG=1 aioc ...
aioc --config path/to/config.yaml ...
AIOC_CONFIG=path/to/config.yaml aioc ...
```

Discovery:

```text
~/.config/aioc/config.env
~/.config/aioc/config.yaml
~/.config/aioc/config.yml
~/.aioc/config.env
~/.aioc/config.yaml
~/.aioc/config.yml
```

Project discovery starts at git root and walks toward cwd:

```text
.env
.aioc.config.env
.aioc/config.env
.aioc/config.yaml
.aioc/config.yml
.config/aioc/config.env
.config/aioc/config.yaml
.config/aioc/config.yml
```

Execution env keys copied from Multica execution semantics:

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
AIOC_<PROVIDER>_PATH=/path/to/bin
AIOC_<PROVIDER>_MODEL=model
AIOC_<PROVIDER>_THINKING_LEVEL=high
AIOC_<PROVIDER>_MAX_TURNS=20
AIOC_<PROVIDER>_ARGS='--flag value'
AIOC_<PROVIDER>_MCP_CONFIG=.aioc/provider-mcp.json
AIOC_<PROVIDER>_SEMANTIC_INACTIVITY_TIMEOUT=10m
AIOC_<PROVIDER>_ENV_KEY=value
```

YAML:

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
    semantic_inactivity_timeout: 10m
```

Inspect:

```bash
aioc config
aioc config --json              # redacted
aioc config --json --show-secrets
```
