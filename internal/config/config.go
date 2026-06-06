package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type FileConfig struct {
	DefaultProvider string                    `yaml:"default_provider"`
	DefaultModel    string                    `yaml:"default_model"`
	Env             map[string]string         `yaml:"env"`
	Providers       map[string]ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	Path    string `yaml:"path"`
	Command string `yaml:"command"`
	Model   string `yaml:"model"`
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
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
	loadOnce.Do(func() {
		loaded, loadErr = LoadPaths(discoverConfigPaths())
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
	for _, dir := range ancestorDirs(cwd) {
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

func ancestorDirs(cwd string) []string {
	var rev []string
	for {
		rev = append(rev, cwd)
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	out := make([]string, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		out = append(out, rev[i])
	}
	return out
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
	if cfg.DefaultProvider != "" {
		out["AIOC_DEFAULT_PROVIDER"] = cfg.DefaultProvider
	}
	if cfg.DefaultModel != "" {
		out["AIOC_DEFAULT_MODEL"] = cfg.DefaultModel
	}
	for k, v := range cfg.Env {
		if strings.TrimSpace(k) != "" {
			out[strings.TrimSpace(k)] = v
		}
	}
	for name, p := range cfg.Providers {
		prefix := "AIOC_" + providerEnvName(name) + "_"
		if p.Path != "" {
			out[prefix+"PATH"] = p.Path
		}
		if p.Command != "" {
			out[prefix+"PATH"] = p.Command
		}
		if p.Model != "" {
			out[prefix+"MODEL"] = p.Model
		}
		if p.BaseURL != "" {
			out[prefix+"BASE_URL"] = p.BaseURL
		}
		if p.APIKey != "" {
			out[prefix+"API_KEY"] = p.APIKey
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

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
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
