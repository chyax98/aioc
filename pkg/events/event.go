package events

import "time"

type Type string

const (
	TypeSession    Type = "session"
	TypeStatus     Type = "status"
	TypeText       Type = "text"
	TypeThinking   Type = "thinking"
	TypeToolUse    Type = "tool_use"
	TypeToolResult Type = "tool_result"
	TypeError      Type = "error"
	TypeDone       Type = "done"
)

type Event struct {
	Type       Type           `json:"type"`
	Provider   string         `json:"provider,omitempty"`
	SessionID  string         `json:"session_id,omitempty"`
	Status     string         `json:"status,omitempty"`
	Delta      string         `json:"delta,omitempty"`
	Output     string         `json:"output,omitempty"`
	Error      string         `json:"error,omitempty"`
	ToolID     string         `json:"tool_id,omitempty"`
	ToolName   string         `json:"tool_name,omitempty"`
	Input      map[string]any `json:"input,omitempty"`
	DurationMS int64          `json:"duration_ms,omitempty"`
	Meta       map[string]any `json:"meta,omitempty"`
	Time       time.Time      `json:"time,omitempty"`
}
