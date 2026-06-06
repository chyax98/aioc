package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	agentpkg "aioc/pkg/agent"
	"aioc/pkg/events"
	"aioc/pkg/provider"
)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string { return strings.Join(*s, " ") }
func (s *stringSliceFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func run(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	providerName := fs.String("p", "", "provider; use auto to select first available provider")
	cwd := fs.String("cwd", ".", "working directory")
	model := fs.String("model", "", "model")
	systemPrompt := fs.String("system", "", "system prompt")
	systemFile := fs.String("system-file", "", "file containing system prompt")
	promptFile := fs.String("prompt-file", "", "file containing prompt")
	resumeSessionID := fs.String("resume", "", "resume session id")
	timeout := fs.Duration("timeout", 0, "timeout")
	mcpConfigPath := fs.String("mcp-config", "", "MCP config JSON file")
	jsonl := fs.Bool("jsonl", true, "emit JSONL events on stdout")
	var customArgs stringSliceFlag
	var envVars stringSliceFlag
	fs.Var(&customArgs, "arg", "extra provider argument; repeatable")
	fs.Var(&envVars, "env", "environment variable KEY=VALUE for agent; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*jsonl {
		return errors.New("only JSONL stdout is supported")
	}
	if *providerName == "" {
		return errors.New("missing provider: -p <provider>")
	}
	promptParts := fs.Args()
	prompt := strings.TrimSpace(strings.Join(promptParts, " "))
	if *promptFile != "" {
		b, err := os.ReadFile(*promptFile)
		if err != nil {
			return fmt.Errorf("read prompt file: %w", err)
		}
		if prompt != "" {
			prompt += "\n"
		}
		prompt += string(b)
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return errors.New("missing prompt")
	}
	if *systemFile != "" {
		b, err := os.ReadFile(*systemFile)
		if err != nil {
			return fmt.Errorf("read system file: %w", err)
		}
		if *systemPrompt != "" {
			*systemPrompt += "\n"
		}
		*systemPrompt += string(b)
	}
	mcpConfig, err := readOptionalJSON(*mcpConfigPath)
	if err != nil {
		return err
	}
	agentEnv, err := parseEnvFlags(envVars)
	if err != nil {
		return err
	}
	d := resolveDetection(*providerName)
	if d.Status != "available" {
		return fmt.Errorf("provider %q unavailable: %s", *providerName, d.Error)
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	backend, err := agentpkg.New(d.Provider, agentpkg.Config{
		ExecutablePath: d.Path,
		Env:            agentEnv,
		Logger:         logger,
	})
	if err != nil {
		return err
	}

	w := events.NewWriter(os.Stdout)
	_ = w.Write(events.Event{Type: events.TypeSession, Provider: d.Provider, Status: "starting", Meta: map[string]any{"path": d.Path, "protocol": d.Protocol, "cwd": *cwd, "launch": d.Launch}})

	session, err := backend.Execute(ctx, prompt, agentpkg.ExecOptions{
		Cwd:             *cwd,
		Model:           *model,
		SystemPrompt:    strings.TrimSpace(*systemPrompt),
		ResumeSessionID: *resumeSessionID,
		CustomArgs:      []string(customArgs),
		McpConfig:       mcpConfig,
		Timeout:         *timeout,
	})
	if err != nil {
		_ = w.Write(events.Event{Type: events.TypeError, Provider: d.Provider, Error: err.Error()})
		_ = w.Write(events.Event{Type: events.TypeDone, Provider: d.Provider, Status: "failed", Error: err.Error()})
		return errSilent
	}

	for msg := range session.Messages {
		_ = w.Write(convertAgentMessage(d.Provider, msg))
	}

	result, ok := <-session.Result
	if !ok {
		_ = w.Write(events.Event{Type: events.TypeDone, Provider: d.Provider, Status: "failed", Error: "agent result channel closed without result"})
		return errSilent
	}
	if result.Status != "completed" {
		_ = w.Write(events.Event{Type: events.TypeDone, Provider: d.Provider, SessionID: result.SessionID, Status: result.Status, Output: result.Output, Error: result.Error, DurationMS: result.DurationMs})
		return errSilent
	}
	_ = w.Write(events.Event{Type: events.TypeDone, Provider: d.Provider, SessionID: result.SessionID, Status: result.Status, Output: result.Output, DurationMS: result.DurationMs})
	return nil
}

func readOptionalJSON(path string) (json.RawMessage, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read mcp config: %w", err)
	}
	if !json.Valid(b) {
		return nil, fmt.Errorf("mcp config is not valid JSON: %s", path)
	}
	return json.RawMessage(b), nil
}

func parseEnvFlags(values []string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(values))
	for _, v := range values {
		key, val, ok := strings.Cut(v, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid --env %q, want KEY=VALUE", v)
		}
		out[key] = val
	}
	return out, nil
}

func resolveDetection(name string) provider.Detection {
	if name != "auto" {
		return findDetection(name)
	}
	priority := []string{"claude", "codex", "pi", "gemini", "cursor", "kimi", "hermes", "kiro", "opencode", "openclaw", "copilot", "antigravity"}
	detections := provider.DetectAll(context.Background())
	byName := map[string]provider.Detection{}
	for _, d := range detections {
		byName[d.Provider] = d
	}
	for _, p := range priority {
		if d, ok := byName[p]; ok && d.Status == "available" {
			return d
		}
	}
	return provider.Detection{Provider: "auto", Status: "missing", Error: "no available provider"}
}

func convertAgentMessage(providerName string, msg agentpkg.Message) events.Event {
	switch msg.Type {
	case agentpkg.MessageText:
		return events.Event{Type: events.TypeText, Provider: providerName, Delta: msg.Content, SessionID: msg.SessionID}
	case agentpkg.MessageThinking:
		return events.Event{Type: events.TypeThinking, Provider: providerName, Delta: msg.Content, SessionID: msg.SessionID}
	case agentpkg.MessageToolUse:
		return events.Event{Type: events.TypeToolUse, Provider: providerName, ToolID: msg.CallID, ToolName: msg.Tool, Input: msg.Input, SessionID: msg.SessionID}
	case agentpkg.MessageToolResult:
		return events.Event{Type: events.TypeToolResult, Provider: providerName, ToolID: msg.CallID, ToolName: msg.Tool, Output: msg.Output, SessionID: msg.SessionID}
	case agentpkg.MessageStatus:
		return events.Event{Type: events.TypeStatus, Provider: providerName, Status: msg.Status, SessionID: msg.SessionID}
	case agentpkg.MessageError:
		return events.Event{Type: events.TypeError, Provider: providerName, Error: msg.Content, SessionID: msg.SessionID}
	case agentpkg.MessageLog:
		return events.Event{Type: events.TypeStatus, Provider: providerName, Status: msg.Content, SessionID: msg.SessionID, Meta: map[string]any{"level": msg.Level}}
	default:
		return events.Event{Type: events.TypeStatus, Provider: providerName, Status: msg.Content, SessionID: msg.SessionID}
	}
}

func findDetection(name string) provider.Detection {
	for _, d := range provider.DetectAll(context.Background()) {
		if d.Provider == name {
			return d
		}
	}
	return provider.Detection{Provider: name, Status: "missing", Error: "unknown provider"}
}
