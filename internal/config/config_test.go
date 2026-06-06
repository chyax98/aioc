package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPathsEnvAndYAML(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".aioc.config.env")
	yamlPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(envPath, []byte("AIOC_DEFAULT_PROVIDER=claude\nAIOC_CLAUDE_PATH=/bin/claude\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yamlPath, []byte(`
default_provider: codex
default_model: gpt-test
env:
  AIOC_EXTRA: yes
providers:
  codex:
    path: /bin/codex
    model: gpt-5.5
  openai:
    base_url: http://127.0.0.1:8317/v1
    api_key: test-key
`), 0o644); err != nil {
		t.Fatal(err)
	}
	keys := []string{"AIOC_DEFAULT_PROVIDER", "AIOC_DEFAULT_MODEL", "AIOC_CLAUDE_PATH", "AIOC_CODEX_PATH", "AIOC_CODEX_MODEL", "AIOC_OPENAI_BASE_URL", "AIOC_OPENAI_API_KEY", "AIOC_EXTRA"}
	for _, key := range keys {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
	res, err := LoadPaths([]string{envPath, yamlPath})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Paths) != 2 {
		t.Fatalf("paths = %v", res.Paths)
	}
	want := map[string]string{
		"AIOC_DEFAULT_PROVIDER": "codex",
		"AIOC_DEFAULT_MODEL":    "gpt-test",
		"AIOC_CLAUDE_PATH":      "/bin/claude",
		"AIOC_CODEX_PATH":       "/bin/codex",
		"AIOC_CODEX_MODEL":      "gpt-5.5",
		"AIOC_OPENAI_BASE_URL":  "http://127.0.0.1:8317/v1",
		"AIOC_OPENAI_API_KEY":   "test-key",
		"AIOC_EXTRA":            "yes",
	}
	for key, value := range want {
		if got := os.Getenv(key); got != value {
			t.Fatalf("%s = %q, want %q", key, got, value)
		}
	}
}

func TestLoadPathsDoesNotOverrideOriginalEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("AIOC_DEFAULT_PROVIDER=claude\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AIOC_DEFAULT_PROVIDER", "pi")
	if _, err := LoadPaths([]string{path}); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("AIOC_DEFAULT_PROVIDER"); got != "pi" {
		t.Fatalf("AIOC_DEFAULT_PROVIDER = %q, want original env pi", got)
	}
}
