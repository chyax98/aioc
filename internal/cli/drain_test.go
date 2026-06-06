package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"aioc/pkg/agent"
	"aioc/pkg/events"
)

func TestDrainSessionIdleWatchdog(t *testing.T) {
	msgs := make(chan agent.Message)
	res := make(chan agent.Result, 1)
	defer close(msgs)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf bytes.Buffer
	result, ok := drainSession(ctx, cancel, &agent.Session{Messages: msgs, Result: res}, "fake", events.NewWriter(&buf), 0, 20*time.Millisecond, 0)
	if ok {
		t.Fatal("idle watchdog should not report success")
	}
	if result.Status != "idle_watchdog" {
		t.Fatalf("status = %q", result.Status)
	}
	var last map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		if err := json.Unmarshal(line, &last); err != nil {
			t.Fatal(err)
		}
	}
	if last["type"] != "done" || last["status"] != "idle_watchdog" {
		t.Fatalf("last event = %#v", last)
	}
}
