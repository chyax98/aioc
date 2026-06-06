---
name: aioc-cli
description: Minimal discovery shadow for AIOC, the local AI all-in-one CLI. Use when user asks to call local agents, route work to Claude/Codex/Pi/etc, inspect provider availability, or parse AIOC JSONL. Triggers: aioc, AI all-in-one CLI, unified agent CLI, run local agent, 调本机 agent, 统一调度 Agent.
---

# AIOC CLI Shadow

This skill is only a small hint so agents discover AIOC.

Source of truth lives in the CLI itself:

```bash
aioc usage
aioc prompt
aioc --help
```

## What AIOC is

AIOC runs local AI agent CLIs through one command and emits normalized JSONL events on stdout.

## First moves

```bash
aioc agents --json
aioc doctor
aioc usage
```

## Run an agent

```bash
aioc run -p claude --cwd . "review current changes"
aioc run -p codex --cwd . "fix failing tests"
aioc run -p pi --cwd . "用中文总结当前目录"
aioc run -p auto --cwd . "pick first available provider"
```

## Parse contract

- stdout: JSONL events
- stderr: provider logs / diagnostics
- wait for final `done` event
- success only when `done.status == "completed"`

Event types:

```text
session status text thinking tool_use tool_result error done
```

## Providers

```text
claude codex pi gemini cursor kimi kiro hermes opencode openclaw copilot antigravity auto
```

## Deeper instructions

When unsure, ask AIOC for its own agent-facing prompt:

```bash
aioc prompt
```

Then follow that output rather than this shadow.
