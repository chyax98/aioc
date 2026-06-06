package provider

import "testing"

func TestBuildLoginShellResolveScript(t *testing.T) {
	script := buildLoginShellResolveScript([]string{"claude", "codex"})
	for _, want := range []string{"for n in claude codex", "unalias", "command -v", "printf '%s\\t%s\\n'"} {
		if !containsString(script, want) {
			t.Fatalf("script missing %q:\n%s", want, script)
		}
	}
}

func TestIsSafeAgentName(t *testing.T) {
	for _, name := range []string{"claude", "cursor-agent", "kiro_cli", "a.b"} {
		if !isSafeAgentName(name) {
			t.Fatalf("%q should be safe", name)
		}
	}
	for _, name := range []string{"", "bad/name", "$(x)", "a b"} {
		if isSafeAgentName(name) {
			t.Fatalf("%q should be unsafe", name)
		}
	}
}

func containsString(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
