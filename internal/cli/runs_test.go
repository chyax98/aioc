package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractRunOutputReturnsLastDoneOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.jsonl")
	content := strings.Join([]string{
		`{"type":"session","provider":"fake"}`,
		`{"type":"text","delta":"hello"}`,
		`{"type":"done","status":"completed","output":"first"}`,
		`{"type":"done","status":"completed","output":"final"}`,
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	got, ok, err := extractRunOutput(path)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || got != "final" {
		t.Fatalf("output = %q, ok=%v; want final true", got, ok)
	}
}

func TestResolveRunPathLatestUsesNewestLexicographicRunID(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIOC_RUNS_DIR", dir)
	oldPath := filepath.Join(dir, "20260101T000000Z-claude-old.jsonl")
	newPath := filepath.Join(dir, "20260102T000000Z-claude-new.jsonl")
	if err := os.WriteFile(oldPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := resolveRunPath("latest")
	if err != nil {
		t.Fatal(err)
	}
	if got != newPath {
		t.Fatalf("latest = %q, want %q", got, newPath)
	}
}

func TestOpenRunHistoryCanBeDisabled(t *testing.T) {
	t.Setenv("AIOC_SAVE_RUNS", "0")
	h, err := openRunHistory("claude")
	if err != nil {
		t.Fatal(err)
	}
	if h != nil {
		t.Fatalf("history = %#v, want nil", h)
	}
}

func TestNewRunIDIncludesSanitizedProvider(t *testing.T) {
	id := newRunID("bad/provider name")
	if !strings.Contains(id, "bad-provider-name") {
		t.Fatalf("run id %q does not contain sanitized provider", id)
	}
	if strings.Contains(id, "/") || strings.Contains(id, " ") {
		t.Fatalf("run id %q contains unsafe chars", id)
	}
}
