package cli

import "testing"

func TestRedactEnvMap(t *testing.T) {
	got := redactEnvMap(map[string]string{"AIOC_OPENAI_API_KEY": "your-api-key-1", "AIOC_DEFAULT_PROVIDER": "claude"})
	if got["AIOC_OPENAI_API_KEY"] == "your-api-key-1" {
		t.Fatal("secret was not redacted")
	}
	if got["AIOC_DEFAULT_PROVIDER"] != "claude" {
		t.Fatalf("non-secret redacted unexpectedly: %q", got["AIOC_DEFAULT_PROVIDER"])
	}
}
