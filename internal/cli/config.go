package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	aiocconfig "aioc/internal/config"
)

func configCmd(args []string) error {
	jsonOut := hasFlag(args, "--json")
	showSecrets := hasFlag(args, "--show-secrets")
	loaded := aiocconfig.Loaded()
	if jsonOut {
		out := loaded
		if !showSecrets {
			out.Env = redactEnvMap(out.Env)
		}
		enc := json.NewEncoder(stdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
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
			if showSecrets {
				fmt.Printf("  %s=%s\n", key, value)
			} else {
				fmt.Printf("  %s=%s\n", key, redactConfigValue(key, value))
			}
		}
	}
	return nil
}

func redactEnvMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = redactConfigValue(k, v)
	}
	return out
}

func interestingEnvKeys() []string {
	keys := []string{"AIOC_DEFAULT_PROVIDER", "AIOC_DEFAULT_MODEL", "AIOC_AGENT_TIMEOUT", "AIOC_CODEX_SEMANTIC_INACTIVITY_TIMEOUT", "AIOC_AUTO_PRIORITY", "AIOC_THINKING_LEVEL"}
	for _, kv := range os.Environ() {
		key, _, ok := splitEnv(kv)
		if ok && strings.HasPrefix(key, "AIOC_") {
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
	key = strings.ToUpper(key)
	for _, word := range []string{"KEY", "TOKEN", "SECRET", "PASSWORD"} {
		if strings.Contains(key, word) {
			return true
		}
	}
	return false
}
