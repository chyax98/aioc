package provider

import (
	"context"
	"os"
	"strings"
	"time"

	"aioc/pkg/agent"
)

type Spec struct {
	Name     string
	Command  string
	EnvPath  string
	Protocol Protocol
}

var Specs = []Spec{
	{Name: "claude", Command: "claude", EnvPath: "AIOC_CLAUDE_PATH", Protocol: ProtocolJSONL},
	{Name: "codex", Command: "codex", EnvPath: "AIOC_CODEX_PATH", Protocol: ProtocolNative},
	{Name: "pi", Command: "pi", EnvPath: "AIOC_PI_PATH", Protocol: ProtocolJSONL},
	{Name: "gemini", Command: "gemini", EnvPath: "AIOC_GEMINI_PATH", Protocol: ProtocolJSONL},
	{Name: "cursor", Command: "cursor-agent", EnvPath: "AIOC_CURSOR_PATH", Protocol: ProtocolJSONL},
	{Name: "kimi", Command: "kimi", EnvPath: "AIOC_KIMI_PATH", Protocol: ProtocolACP},
	{Name: "kiro", Command: "kiro-cli", EnvPath: "AIOC_KIRO_PATH", Protocol: ProtocolACP},
	{Name: "hermes", Command: "hermes", EnvPath: "AIOC_HERMES_PATH", Protocol: ProtocolACP},
	{Name: "opencode", Command: "opencode", EnvPath: "AIOC_OPENCODE_PATH", Protocol: ProtocolJSONL},
	{Name: "openclaw", Command: "openclaw", EnvPath: "AIOC_OPENCLAW_PATH", Protocol: ProtocolJSONL},
	{Name: "copilot", Command: "copilot", EnvPath: "AIOC_COPILOT_PATH", Protocol: ProtocolJSONL},
	{Name: "antigravity", Command: "agy", EnvPath: "AIOC_ANTIGRAVITY_PATH", Protocol: ProtocolPlain},
}

func DetectAll(ctx context.Context) []Detection {
	out := make([]Detection, 0, len(Specs))
	for _, spec := range Specs {
		out = append(out, Detect(ctx, spec, false))
	}
	return out
}

func DetectAllWithVersions(ctx context.Context) []Detection {
	out := make([]Detection, 0, len(Specs))
	for _, spec := range Specs {
		out = append(out, Detect(ctx, spec, true))
	}
	return out
}

func Detect(ctx context.Context, spec Spec, withVersion bool) Detection {
	cmd := strings.TrimSpace(os.Getenv(spec.EnvPath))
	if cmd == "" {
		cmd = spec.Command
	}
	d := Detection{
		Provider:                spec.Name,
		Command:                 cmd,
		Protocol:                spec.Protocol,
		Status:                  "missing",
		Launch:                  agent.LaunchHeader(spec.Name),
		ModelSelectionSupported: agent.ModelSelectionSupported(spec.Name),
		EnvPath:                 spec.EnvPath,
	}
	path, err := lookPathWithShellFallback(cmd, spec.Command)
	if err != nil {
		d.Error = err.Error()
		return d
	}
	d.Path = path
	d.Status = "available"
	if withVersion {
		versionCtx := ctx
		if versionCtx == nil {
			versionCtx = context.Background()
		}
		var cancel context.CancelFunc
		versionCtx, cancel = context.WithTimeout(versionCtx, 5*time.Second)
		defer cancel()
		version, err := agent.DetectVersion(versionCtx, path)
		if err != nil {
			d.Error = err.Error()
		} else {
			d.Version = version
			if err := agent.CheckMinVersion(spec.Name, version); err != nil {
				d.Status = "unsupported"
				d.Error = err.Error()
			}
		}
	}
	return d
}

func SpecByName(name string) (Spec, bool) {
	for _, spec := range Specs {
		if spec.Name == name {
			return spec, true
		}
	}
	return Spec{}, false
}
