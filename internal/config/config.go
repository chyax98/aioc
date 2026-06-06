package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/mattn/go-shellwords"
	"gopkg.in/yaml.v3"
)

type FileConfig struct {
	DefaultProvider                string                    `yaml:"default_provider"`
	DefaultModel                   string                    `yaml:"default_model"`
	ThinkingLevel                  string                    `yaml:"thinking_level"`
	AgentTimeout                   string                    `yaml:"agent_timeout"`
	CodexSemanticInactivityTimeout string                    `yaml:"codex_semantic_inactivity_timeout"`
	AgentIdleWatchdog              string                    `yaml:"agent_idle_watchdog"`
	AgentToolWatchdog              string                    `yaml:"agent_tool_watchdog"`
	AutoPriority                   []string                  `yaml:"auto_priority"`
	McpConfig                      string                    `yaml:"mcp_config"`
	Env                            map[string]string         `yaml:"env"`
	Providers                      map[string]ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	Path            string            `yaml:"path"`
	Command         string            `yaml:"command"`
	Model           string            `yaml:"model"`
	ThinkingLevel   string            `yaml:"thinking_level"`
	MaxTurns        int               `yaml:"max_turns"`
	Args            []string          `yaml:"args"`
	McpConfig       string            `yaml:"mcp_config"`
	BaseURL         string            `yaml:"base_url"`
	APIKey          string            `yaml:"api_key"`
	Env             map[string]string `yaml:"env"`
	ChildEnv        map[string]string `yaml:"child_env"`
	CustomEnv       map[string]string `yaml:"custom_env"`
	InactivityTimer string            `yaml:"semantic_inactivity_timeout"`
}

type LoadOptions struct {
	NoConfig bool
	Paths    []string
}

type LoadResult struct {
	Paths []string          `json:"paths"`
	Env   map[string]string `json:"env"`
}

var (
	loadOnce sync.Once
	loaded   LoadResult
	loadErr  error
)

func LoadAuto() (LoadResult, error) {
	return LoadAutoWithOptions(LoadOptions{})
}

func LoadAutoWithOptions(opts LoadOptions) (LoadResult, error) {
	loadOnce.Do(func() {
		if opts.NoConfig || truthy(os.Getenv("AIOC_NO_CONFIG")) {
			loaded = LoadResult{Paths: []string{}, Env: map[string]string{}}
			return
		}
		paths := opts.Paths
		if len(paths) == 0 {
			paths = pathsFromEnv("AIOC_CONFIG")
		}
		if len(paths) == 0 {
			paths = discoverConfigPaths()
		}
		loaded, loadErr = LoadPaths(paths)
	})
	return loaded, loadErr
}

func Loaded() LoadResult {
	return loaded
}

func LoadPaths(paths []string) (LoadResult, error) {
	originalEnv := map[string]struct{}{}
	for _, kv := range os.Environ() {
		key, _, ok := strings.Cut(kv, "=")
		if ok {
			originalEnv[key] = struct{}{}
		}
	}
	merged := map[string]string{}
	var loadedPaths []string
	for _, path := range unique(paths) {
		if path == "" || !fileExists(path) {
			continue
		}
		values, err := parseConfigFile(path)
		if err != nil {
			return LoadResult{Paths: loadedPaths, Env: merged}, err
		}
		for k, v := range values {
			merged[k] = v
		}
		loadedPaths = append(loadedPaths, path)
	}
	for k, v := range merged {
		if _, exists := originalEnv[k]; exists {
			continue
		}
		_ = os.Setenv(k, v)
	}
	return LoadResult{Paths: loadedPaths, Env: merged}, nil
}

func discoverConfigPaths() []string {
	paths := []string{}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		paths = append(paths,
			filepath.Join(home, ".config", "aioc", "config.env"),
			filepath.Join(home, ".config", "aioc", "config.yaml"),
			filepath.Join(home, ".config", "aioc", "config.yml"),
			filepath.Join(home, ".aioc", "config.env"),
			filepath.Join(home, ".aioc", "config.yaml"),
			filepath.Join(home, ".aioc", "config.yml"),
		)
	}
	cwd, err := os.Getwd()
	if err != nil || cwd == "" {
		return paths
	}
	for _, dir := range projectDirs(cwd) {
		paths = append(paths,
			filepath.Join(dir, ".env"),
			filepath.Join(dir, ".aioc.config.env"),
			filepath.Join(dir, ".aioc", "config.env"),
			filepath.Join(dir, ".aioc", "config.yaml"),
			filepath.Join(dir, ".aioc", "config.yml"),
			filepath.Join(dir, ".config", "aioc", "config.env"),
			filepath.Join(dir, ".config", "aioc", "config.yaml"),
			filepath.Join(dir, ".config", "aioc", "config.yml"),
		)
	}
	return paths
}

func projectDirs(cwd string) []string {
	cwd = filepath.Clean(cwd)
	root := findGitRoot(cwd)
	if root == "" {
		return []string{cwd}
	}
	var rev []string
	for dir := cwd; ; dir = filepath.Dir(dir) {
		rev = append(rev, dir)
		if dir == root {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	out := make([]string, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		out = append(out, rev[i])
	}
	return out
}

func findGitRoot(cwd string) string {
	for dir := filepath.Clean(cwd); ; dir = filepath.Dir(dir) {
		if fileOrDirExists(filepath.Join(dir, ".git")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
	}
}

func parseConfigFile(path string) (map[string]string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return parseYAML(path)
	default:
		return parseEnvFile(path)
	}
}

func parseYAML(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg FileConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := map[string]string{}
	set(out, "AIOC_DEFAULT_PROVIDER", cfg.DefaultProvider)
	set(out, "AIOC_DEFAULT_MODEL", cfg.DefaultModel)
	set(out, "AIOC_THINKING_LEVEL", cfg.ThinkingLevel)
	set(out, "AIOC_AGENT_TIMEOUT", cfg.AgentTimeout)
	set(out, "AIOC_CODEX_SEMANTIC_INACTIVITY_TIMEOUT", cfg.CodexSemanticInactivityTimeout)
	set(out, "AIOC_AGENT_IDLE_WATCHDOG", cfg.AgentIdleWatchdog)
	set(out, "AIOC_AGENT_TOOL_WATCHDOG", cfg.AgentToolWatchdog)
	set(out, "AIOC_AUTO_PRIORITY", strings.Join(cfg.AutoPriority, ","))
	set(out, "AIOC_MCP_CONFIG", cfg.McpConfig)
	for k, v := range cfg.Env {
		set(out, k, v)
	}
	for name, p := range cfg.Providers {
		prefix := "AIOC_" + providerEnvName(name) + "_"
		set(out, prefix+"PATH", firstNonEmpty(p.Command, p.Path))
		set(out, prefix+"MODEL", p.Model)
		set(out, prefix+"THINKING_LEVEL", p.ThinkingLevel)
		if p.MaxTurns > 0 {
			set(out, prefix+"MAX_TURNS", strconv.Itoa(p.MaxTurns))
		}
		if len(p.Args) > 0 {
			set(out, prefix+"ARGS", shellQuoteJoin(p.Args))
		}
		set(out, prefix+"MCP_CONFIG", p.McpConfig)
		set(out, prefix+"BASE_URL", p.BaseURL)
		set(out, prefix+"API_KEY", p.APIKey)
		set(out, prefix+"SEMANTIC_INACTIVITY_TIMEOUT", p.InactivityTimer)
		for k, v := range p.Env {
			set(out, k, v)
		}
		for k, v := range p.ChildEnv {
			set(out, prefix+"ENV_"+providerEnvName(k), v)
		}
		for k, v := range p.CustomEnv {
			set(out, prefix+"ENV_"+providerEnvName(k), v)
		}
	}
	return out, nil
}

func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	s := bufio.NewScanner(f)
	lineNo := 0
	for s.Scan() {
		lineNo++
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("parse %s:%d: want KEY=VALUE", path, lineNo)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("parse %s:%d: empty key", path, lineNo)
		}
		out[key] = unquoteEnvValue(strings.TrimSpace(value))
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func ParseShellArgsEnv(name string) ([]string, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil, nil
	}
	args, err := shellwords.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", name, err)
	}
	return args, nil
}

func ParseDurationEnv(name string) (string, bool) {
	v := strings.TrimSpace(os.Getenv(name))
	return v, v != ""
}

func pathsFromEnv(name string) []string {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil
	}
	return splitList(raw)
}

func splitList(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == filepath.ListSeparator })
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func unquoteEnvValue(value string) string {
	if len(value) >= 2 {
		quote := value[0]
		if (quote == '\'' || quote == '"') && value[len(value)-1] == quote {
			value = value[1 : len(value)-1]
		}
	}
	if i := strings.Index(value, " #"); i >= 0 {
		value = strings.TrimSpace(value[:i])
	}
	return value
}

func providerEnvName(name string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(name, "-", "_"), ".", "_"))
}

func shellQuoteJoin(args []string) string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, strconv.Quote(arg))
	}
	return strings.Join(out, " ")
}

func set(out map[string]string, key, value string) {
	key = strings.TrimSpace(key)
	if key != "" && value != "" {
		out[key] = value
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func fileOrDirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func unique(paths []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, path := range paths {
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}
