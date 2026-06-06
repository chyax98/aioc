package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var defaultCommandNames = []string{
	"claude", "codex", "opencode", "openclaw", "hermes",
	"gemini", "pi", "cursor-agent", "copilot", "kimi", "kiro-cli", "agy",
}

var shellResolveOnce sync.Once
var shellResolved map[string]string

var codexDesktopAppBundlePaths = func() []string {
	paths := []string{"/Applications/Codex.app/Contents/Resources/codex"}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, "Applications", "Codex.app", "Contents", "Resources", "codex"))
	}
	return paths
}

const loginShellResolveTimeout = 3 * time.Second
const loginShellResolveWaitDelay = 2 * time.Second

var supportedLoginShells = map[string]struct{}{
	"bash": {},
	"zsh":  {},
	"sh":   {},
	"dash": {},
	"ksh":  {},
}

func lookPathWithShellFallback(cmd, defaultCmd string) (string, error) {
	path, err := exec.LookPath(cmd)
	if err == nil {
		return path, nil
	}
	if strings.ContainsAny(cmd, "/\\") {
		return "", err
	}
	shellResolveOnce.Do(func() {
		shellResolved = resolveAgentsViaLoginShell(defaultCommandNames)
	})
	if path, ok := shellResolved[cmd]; ok {
		return path, nil
	}
	if defaultCmd == "codex" && cmd == defaultCmd {
		for _, path := range codexDesktopAppBundlePaths() {
			if _, statErr := os.Stat(path); statErr == nil {
				return path, nil
			}
		}
	}
	return "", err
}

func resolveAgentsViaLoginShell(names []string) map[string]string {
	out := map[string]string{}
	if len(names) == 0 {
		return out
	}
	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" {
		return out
	}
	if _, ok := supportedLoginShells[filepath.Base(shell)]; !ok {
		return out
	}
	safe := make([]string, 0, len(names))
	for _, n := range names {
		if isSafeAgentName(n) {
			safe = append(safe, n)
		}
	}
	if len(safe) == 0 {
		return out
	}
	ctx, cancel := context.WithTimeout(context.Background(), loginShellResolveTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, shell, "-ilc", buildLoginShellResolveScript(safe))
	cmd.WaitDelay = loginShellResolveWaitDelay
	raw, err := cmd.Output()
	if err != nil {
		return out
	}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		name, path := parts[0], strings.TrimSpace(parts[1])
		if !filepath.IsAbs(path) {
			continue
		}
		if _, err := exec.LookPath(path); err != nil {
			continue
		}
		out[name] = path
	}
	return out
}

func buildLoginShellResolveScript(names []string) string {
	var b strings.Builder
	b.WriteString("for n in")
	for _, n := range names {
		b.WriteByte(' ')
		b.WriteString(n)
	}
	b.WriteString("; do\n")
	b.WriteString("  unalias \"$n\" 2>/dev/null\n")
	b.WriteString("  unset -f \"$n\" 2>/dev/null\n")
	b.WriteString("  p=$(command -v \"$n\" 2>/dev/null) || continue\n")
	b.WriteString("  [ -n \"$p\" ] || continue\n")
	b.WriteString("  case \"$p\" in /*) ;; *) continue ;; esac\n")
	b.WriteString("  d=$(dirname \"$p\") && f=$(basename \"$p\") && c=$(cd \"$d\" 2>/dev/null && pwd -P) || continue\n")
	b.WriteString("  printf '%s\\t%s\\n' \"$n\" \"$c/$f\"\n")
	b.WriteString("done\n")
	return b.String()
}

func isSafeAgentName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '.':
		default:
			return false
		}
	}
	return true
}
