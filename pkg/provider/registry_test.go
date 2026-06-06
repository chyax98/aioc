package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"aioc/pkg/agent"
)

func TestSpecsAreBackedByAgentFactory(t *testing.T) {
	for _, spec := range Specs {
		if _, err := agent.New(spec.Name, agent.Config{}); err != nil {
			t.Fatalf("spec %q not supported by agent factory: %v", spec.Name, err)
		}
		if agent.LaunchHeader(spec.Name) == "" {
			t.Fatalf("spec %q missing launch header", spec.Name)
		}
	}
}

func TestDetectUsesEnvPathOverride(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "fake-claude")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AIOC_CLAUDE_PATH", bin)
	d := Detect(context.Background(), Spec{Name: "claude", Command: "claude", EnvPath: "AIOC_CLAUDE_PATH", Protocol: ProtocolJSONL}, false)
	if d.Status != "available" || d.Path != bin || d.Command != bin {
		t.Fatalf("detect = %+v, want available path override", d)
	}
}

func TestDetectMissing(t *testing.T) {
	d := Detect(context.Background(), Spec{Name: "missing", Command: "definitely-not-aioc-test-binary", EnvPath: "AIOC_MISSING_PATH", Protocol: ProtocolJSONL}, false)
	if d.Status != "missing" || d.Error == "" {
		t.Fatalf("detect missing = %+v", d)
	}
}
