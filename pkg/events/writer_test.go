package events

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriterWritesJSONLWithTime(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.Write(Event{Type: TypeDone, Provider: "test", Status: "completed"}); err != nil {
		t.Fatal(err)
	}
	var ev Event
	if err := json.Unmarshal(buf.Bytes(), &ev); err != nil {
		t.Fatalf("invalid json: %v; %s", err, buf.String())
	}
	if ev.Type != TypeDone || ev.Provider != "test" || ev.Status != "completed" {
		t.Fatalf("event = %+v", ev)
	}
	if ev.Time.IsZero() {
		t.Fatal("event time not set")
	}
	if got := buf.String(); got == "" || got[len(got)-1] != '\n' {
		t.Fatalf("writer did not emit jsonl newline: %q", got)
	}
}
