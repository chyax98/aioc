package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	aiocconfig "aioc/internal/config"
)

func configCmd(args []string) error {
	jsonOut := hasFlag(args, "--json")
	loaded := aiocconfig.Loaded()
	if jsonOut {
		enc := json.NewEncoder(stdout())
		enc.SetIndent("", "  ")
		return enc.Encode(loaded)
	}
	fmt.Println("config files:")
	if len(loaded.Paths) == 0 {
		fmt.Println("  -")
	} else {
		for _, path := range loaded.Paths {
			fmt.Printf("  %s\n", path)
		}
	}
	fmt.Println("effective:")
	for _, key := range interestingEnvKeys() {
		if value := os.Getenv(key); value != "" {
			fmt.Printf("  %s=%s\n", key, redactConfigValue(key, value))
		}
	}
	return nil
}

func interestingEnvKeys() []string {
	keys := []string{"AIOC_DEFAULT_PROVIDER", "AIOC_DEFAULT_MODEL"}
	for _, kv := range os.Environ() {
		key, _, ok := splitEnv(kv)
		if ok && len(key) > 5 && key[:5] == "AIOC_" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	out := []string{}
	seen := map[string]struct{}{}
	for _, key := range keys {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func splitEnv(kv string) (string, string, bool) {
	for i, c := range kv {
		if c == '=' {
			return kv[:i], kv[i+1:], true
		}
	}
	return "", "", false
}

func redactConfigValue(key, value string) string {
	if len(value) == 0 {
		return value
	}
	if containsSecretWord(key) {
		if len(value) <= 8 {
			return "***"
		}
		return value[:4] + "***" + value[len(value)-4:]
	}
	return value
}

func containsSecretWord(key string) bool {
	for _, word := range []string{"KEY", "TOKEN", "SECRET", "PASSWORD"} {
		if contains(key, word) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
