package provider

import (
	"context"
	"time"

	"aioc/pkg/events"
)

type Protocol string

const (
	ProtocolJSONL  Protocol = "jsonl"
	ProtocolACP    Protocol = "acp"
	ProtocolNative Protocol = "native"
	ProtocolPlain  Protocol = "plain"
)

type Detection struct {
	Provider                string   `json:"provider"`
	Command                 string   `json:"command"`
	Path                    string   `json:"path,omitempty"`
	Protocol                Protocol `json:"protocol"`
	Status                  string   `json:"status"`
	Launch                  string   `json:"launch,omitempty"`
	ModelSelectionSupported bool     `json:"model_selection_supported"`
	EnvPath                 string   `json:"env_path"`
	Version                 string   `json:"version,omitempty"`
	Error                   string   `json:"error,omitempty"`
}

type RunRequest struct {
	Cwd             string
	Prompt          string
	Model           string
	SystemPrompt    string
	ResumeSessionID string
	Env             map[string]string
	Args            []string
	Timeout         time.Duration
}

type Result struct {
	Status     string
	Output     string
	Error      string
	DurationMS int64
	SessionID  string
}

type Session struct {
	Events <-chan events.Event
	Done   <-chan Result
}

type Provider interface {
	Name() string
	Detect(ctx context.Context) Detection
	Run(ctx context.Context, req RunRequest) (*Session, error)
}
