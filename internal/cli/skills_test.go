package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillsInstallWritesUsageShadow(t *testing.T) {
	dir := t.TempDir()
	if err := skills([]string{"install", "--dir", dir, "--print-path"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "aioc-cli", "SKILL.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	for _, want := range []string{"name: aioc-cli", "aioc usage", "aioc prompt", "aioc run -p"} {
		if !strings.Contains(content, want) {
			t.Fatalf("skill shadow missing %q:\n%s", want, content)
		}
	}
}
