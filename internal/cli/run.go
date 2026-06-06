package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	aiocconfig "aioc/internal/config"
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
	providerName := fs.String("p", defaultProvider(), "provider; default from AIOC_DEFAULT_PROVIDER or auto")
	cwd := fs.String("cwd", ".", "working directory")
	model := fs.String("model", defaultModel(), "model; default from AIOC_DEFAULT_MODEL or provider config")
	systemPrompt := fs.String("system", "", "system prompt")
	systemFile := fs.String("system-file", "", "file containing system prompt")
	promptFile := fs.String("prompt-file", "", "file containing prompt")
	resumeSessionID := fs.String("resume", "", "resume session id")
	timeout := fs.Duration("timeout", defaultTimeout(), "timeout; default AIOC_AGENT_TIMEOUT")
	inactivityTimeout := fs.Duration("inactivity-timeout", defaultInactivityTimeout(), "semantic inactivity timeout; default AIOC_CODEX_SEMANTIC_INACTIVITY_TIMEOUT")
	idleWatchdog := fs.Duration("idle-watchdog", durationEnvOrZero("AIOC_AGENT_IDLE_WATCHDOG"), "force-stop when backend emits no messages for duration; default AIOC_AGENT_IDLE_WATCHDOG")
	toolWatchdog := fs.Duration("tool-watchdog", durationEnvOrZero("AIOC_AGENT_TOOL_WATCHDOG"), "force-stop when one tool stays in flight silently for duration; default AIOC_AGENT_TOOL_WATCHDOG")
	mcpConfigPath := fs.String("mcp-config", strings.TrimSpace(os.Getenv("AIOC_MCP_CONFIG")), "MCP config JSON file")
	thinkingLevel := fs.String("thinking", strings.TrimSpace(os.Getenv("AIOC_THINKING_LEVEL")), "reasoning/thinking level")
	maxTurns := fs.Int("max-turns", intEnv("AIOC_MAX_TURNS"), "max turns")
	jsonl := fs.Bool("jsonl", true, "emit JSONL events on stdout")
	var customArgs stringSliceFlag
	var extraArgs stringSliceFlag
	var envVars stringSliceFlag
	fs.Var(&customArgs, "arg", "provider custom argument appended after config args; repeatable")
	fs.Var(&extraArgs, "extra-arg", "provider default argument appended before --arg; repeatable")
	fs.Var(&envVars, "env", "environment variable KEY=VALUE for agent; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*jsonl {
		return errors.New("only JSONL stdout is supported")
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

	d := resolveDetection(*providerName)
	if d.Status != "available" {
		return fmt.Errorf("provider %q unavailable: %s", *providerName, d.Error)
	}
	providerEnvPrefix := providerEnvPrefix(d.Provider)
	if *model == "" {
		*model = providerDefaultModel(d.Provider)
	}
	if *thinkingLevel == "" {
		*thinkingLevel = strings.TrimSpace(os.Getenv(providerEnvPrefix + "THINKING_LEVEL"))
	}
	if *maxTurns == 0 {
		*maxTurns = intEnv(providerEnvPrefix + "MAX_TURNS")
	}
	if !flagPassed(fs, "mcp-config") {
		if v := strings.TrimSpace(os.Getenv(providerEnvPrefix + "MCP_CONFIG")); v != "" {
			*mcpConfigPath = v
		}
	}
	if !flagPassed(fs, "inactivity-timeout") {
		if v, ok := durationEnv(providerEnvPrefix + "SEMANTIC_INACTIVITY_TIMEOUT"); ok {
			*inactivityTimeout = v
		}
	}
	mcpConfig, err := readOptionalJSON(*mcpConfigPath)
	if err != nil {
		return err
	}
	agentEnv, err := parseEnvFlags(envVars)
	if err != nil {
		return err
	}
	agentEnv = mergeEnv(agentEnv, providerChildEnv(d.Provider))
	configExtraArgs, err := aiocconfig.ParseShellArgsEnv(providerEnvPrefix + "ARGS")
	if err != nil {
		return err
	}
	configExtraArgs = append(configExtraArgs, []string(extraArgs)...)
	if *thinkingLevel != "" {
		ok, err := agentpkg.ValidateThinkingLevel(context.Background(), d.Provider, d.Path, *model, *thinkingLevel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: thinking_level catalog lookup failed; passing through: %v\n", err)
		} else if !ok {
			fmt.Fprintf(os.Stderr, "warning: thinking_level %q invalid for provider/model; skipping\n", *thinkingLevel)
			*thinkingLevel = ""
		}
	}

	ctx, agentCancel := context.WithCancel(context.Background())
	defer agentCancel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	backend, err := agentpkg.New(d.Provider, agentpkg.Config{
		ExecutablePath: d.Path,
		Env:            agentEnv,
		Logger:         logger,
	})
	if err != nil {
		return err
	}

	history, err := openRunHistory(d.Provider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: save run history failed: %v\n", err)
	}
	defer history.Close()

	out := io.Writer(os.Stdout)
	meta := map[string]any{"path": d.Path, "protocol": d.Protocol, "cwd": *cwd, "launch": d.Launch, "model": *model}
	if history != nil {
		out = io.MultiWriter(os.Stdout, history.file)
		meta["run_id"] = history.ID
		meta["events_path"] = history.Path
	}
	w := events.NewWriter(out)
	_ = w.Write(events.Event{Type: events.TypeSession, Provider: d.Provider, Status: "starting", Meta: meta})

	session, err := backend.Execute(ctx, prompt, agentpkg.ExecOptions{
		Cwd:                       *cwd,
		Model:                     *model,
		SystemPrompt:              strings.TrimSpace(*systemPrompt),
		MaxTurns:                  *maxTurns,
		Timeout:                   *timeout,
		SemanticInactivityTimeout: *inactivityTimeout,
		ResumeSessionID:           *resumeSessionID,
		ExtraArgs:                 configExtraArgs,
		CustomArgs:                []string(customArgs),
		McpConfig:                 mcpConfig,
		ThinkingLevel:             *thinkingLevel,
	})
	if err != nil {
		_ = w.Write(events.Event{Type: events.TypeError, Provider: d.Provider, Error: err.Error()})
		_ = w.Write(events.Event{Type: events.TypeDone, Provider: d.Provider, Status: "failed", Error: err.Error()})
		return errSilent
	}

	result, ok := drainSession(ctx, agentCancel, session, d.Provider, w, *timeout, *idleWatchdog, *toolWatchdog)
	if !ok {
		return errSilent
	}
	if result.Status != "completed" {
		return errSilent
	}
	return nil
}

func drainSession(ctx context.Context, cancel context.CancelFunc, session *agentpkg.Session, providerName string, w *events.Writer, timeout, idleWindow, toolWindow time.Duration) (agentpkg.Result, bool) {
	drainCtx := ctx
	var drainCancel context.CancelFunc
	if timeout > 0 {
		drainCtx, drainCancel = context.WithTimeout(ctx, timeout+30*time.Second)
	} else {
		drainCtx, drainCancel = context.WithCancel(ctx)
	}
	defer drainCancel()

	lastActivity := time.Now()
	inFlightTools := 0
	idleFired := false
	idleReason := ""
	messagesOpen := true
	messages := session.Messages
	var tickerC <-chan time.Time
	var ticker *time.Ticker
	if idleWindow > 0 {
		interval := idleWindow / 2
		if idleWindow >= time.Minute && interval < 30*time.Second {
			interval = 30 * time.Second
		}
		if interval <= 0 {
			interval = idleWindow
		}
		ticker = time.NewTicker(interval)
		defer ticker.Stop()
		tickerC = ticker.C
	}

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				messagesOpen = false
				messages = nil
				continue
			}
			lastActivity = time.Now()
			switch msg.Type {
			case agentpkg.MessageToolUse:
				inFlightTools++
			case agentpkg.MessageToolResult:
				if inFlightTools > 0 {
					inFlightTools--
				}
			}
			_ = w.Write(convertAgentMessage(providerName, msg))
		case result, ok := <-session.Result:
			if !ok {
				_ = w.Write(events.Event{Type: events.TypeDone, Provider: providerName, Status: "failed", Error: "agent result channel closed without result"})
				return agentpkg.Result{}, false
			}
			if idleFired {
				result.Status = "idle_watchdog"
				if result.Error == "" {
					result.Error = idleReason
				}
			}
			_ = w.Write(events.Event{Type: events.TypeDone, Provider: providerName, SessionID: result.SessionID, Status: result.Status, Output: result.Output, Error: result.Error, DurationMS: result.DurationMs, Usage: convertUsage(result.Usage)})
			return result, result.Status == "completed"
		case <-tickerC:
			threshold := idleWindow
			if inFlightTools > 0 {
				if toolWindow <= 0 {
					continue
				}
				threshold = toolWindow
			}
			if threshold > 0 && time.Since(lastActivity) >= threshold && (!messagesOpen || len(session.Messages) == 0) {
				idleFired = true
				idleReason = fmt.Sprintf("agent produced no new messages for %s and message queue was empty; force-stopped by idle watchdog", threshold)
				cancel()
			}
		case <-drainCtx.Done():
			if idleFired {
				result := agentpkg.Result{Status: "idle_watchdog", Error: idleReason}
				_ = w.Write(events.Event{Type: events.TypeDone, Provider: providerName, Status: result.Status, Error: result.Error})
				return result, false
			}
			status := "cancelled"
			err := "agent run cancelled"
			if errors.Is(drainCtx.Err(), context.DeadlineExceeded) {
				status = "timeout"
				err = "agent did not produce result within drain timeout"
			}
			result := agentpkg.Result{Status: status, Error: err}
			_ = w.Write(events.Event{Type: events.TypeDone, Provider: providerName, Status: result.Status, Error: result.Error})
			return result, false
		}
	}
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
	out := map[string]string{}
	for _, v := range values {
		key, val, ok := strings.Cut(v, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid --env %q, want KEY=VALUE", v)
		}
		out[key] = val
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func mergeEnv(base, overlay map[string]string) map[string]string {
	if len(overlay) == 0 {
		return base
	}
	if base == nil {
		base = map[string]string{}
	}
	for k, v := range overlay {
		base[k] = v
	}
	return base
}

func defaultProvider() string {
	if v := strings.TrimSpace(os.Getenv("AIOC_DEFAULT_PROVIDER")); v != "" {
		return v
	}
	return "auto"
}

func defaultModel() string {
	return strings.TrimSpace(os.Getenv("AIOC_DEFAULT_MODEL"))
}

func defaultTimeout() time.Duration {
	if v, ok := durationEnv("AIOC_AGENT_TIMEOUT"); ok {
		return v
	}
	return 0
}

func defaultInactivityTimeout() time.Duration {
	if v, ok := durationEnv("AIOC_CODEX_SEMANTIC_INACTIVITY_TIMEOUT"); ok {
		return v
	}
	return 10 * time.Minute
}

func durationEnvOrZero(name string) time.Duration {
	if v, ok := durationEnv(name); ok {
		return v
	}
	return 0
}

func durationEnv(name string) (time.Duration, bool) {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return 0, false
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, false
	}
	return d, true
}

func intEnv(name string) int {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func providerDefaultModel(providerName string) string {
	return strings.TrimSpace(os.Getenv(providerEnvPrefix(providerName) + "MODEL"))
}

func providerEnvPrefix(providerName string) string {
	return "AIOC_" + strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(providerName, "-", "_"), ".", "_")) + "_"
}

func providerChildEnv(providerName string) map[string]string {
	prefix := providerEnvPrefix(providerName) + "ENV_"
	out := map[string]string{}
	for _, kv := range os.Environ() {
		key, val, ok := strings.Cut(kv, "=")
		if !ok || !strings.HasPrefix(key, prefix) {
			continue
		}
		childKey := strings.TrimPrefix(key, prefix)
		if childKey != "" {
			out[childKey] = val
		}
	}
	return out
}

func flagPassed(fs *flag.FlagSet, name string) bool {
	passed := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			passed = true
		}
	})
	return passed
}

func autoPriority() []string {
	if raw := strings.TrimSpace(os.Getenv("AIOC_AUTO_PRIORITY")); raw != "" {
		parts := strings.Split(raw, ",")
		out := []string{}
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return []string{"claude", "codex", "pi", "gemini", "cursor", "kimi", "hermes", "kiro", "opencode", "openclaw", "copilot", "antigravity"}
}

func resolveDetection(name string) provider.Detection {
	if name != "auto" {
		return findDetection(name)
	}
	detections := provider.DetectAll(context.Background())
	byName := map[string]provider.Detection{}
	for _, d := range detections {
		byName[d.Provider] = d
	}
	for _, p := range autoPriority() {
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

func convertUsage(usage map[string]agentpkg.TokenUsage) map[string]events.TokenUsage {
	if len(usage) == 0 {
		return nil
	}
	out := map[string]events.TokenUsage{}
	for model, u := range usage {
		out[model] = events.TokenUsage{
			InputTokens:      u.InputTokens,
			OutputTokens:     u.OutputTokens,
			CacheReadTokens:  u.CacheReadTokens,
			CacheWriteTokens: u.CacheWriteTokens,
		}
	}
	return out
}

func findDetection(name string) provider.Detection {
	for _, d := range provider.DetectAll(context.Background()) {
		if d.Provider == name {
			return d
		}
	}
	return provider.Detection{Provider: name, Status: "missing", Error: "unknown provider"}
}
