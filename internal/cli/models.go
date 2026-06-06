package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"aioc/pkg/agent"
	"aioc/pkg/provider"
)

func models(args []string) error {
	fs := flag.NewFlagSet("models", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	providerName := fs.String("p", "", "provider")
	jsonOut := fs.Bool("json", false, "print JSON")
	timeout := fs.Duration("timeout", 15*time.Second, "discovery timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *providerName == "" {
		return errors.New("missing provider: -p <provider>")
	}
	d := findDetection(*providerName)
	if d.Status != "available" {
		return fmt.Errorf("provider %q unavailable: %s", *providerName, d.Error)
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	models, err := agent.ListModels(ctx, d.Provider, d.Path)
	if err != nil {
		return err
	}
	if *jsonOut {
		enc := json.NewEncoder(stdout())
		enc.SetIndent("", "  ")
		return enc.Encode(models)
	}
	if len(models) == 0 {
		fmt.Printf("%s: no model catalog exposed\n", d.Provider)
		return nil
	}
	for _, m := range models {
		parts := []string{m.ID}
		if m.Label != "" && m.Label != m.ID {
			parts = append(parts, "-", m.Label)
		}
		if m.Provider != "" {
			parts = append(parts, fmt.Sprintf("[%s]", m.Provider))
		}
		if m.Default {
			parts = append(parts, "(default)")
		}
		if m.Thinking != nil && len(m.Thinking.SupportedLevels) > 0 {
			levels := make([]string, 0, len(m.Thinking.SupportedLevels))
			for _, level := range m.Thinking.SupportedLevels {
				levels = append(levels, level.Value)
			}
			parts = append(parts, "thinking="+strings.Join(levels, ","))
		}
		fmt.Println(strings.Join(parts, " "))
	}
	return nil
}

func allModels(args []string) error {
	fs := flag.NewFlagSet("models-all", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	jsonOut := fs.Bool("json", false, "print JSON")
	timeout := fs.Duration("timeout", 20*time.Second, "per-provider discovery timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	type providerModels struct {
		Provider string        `json:"provider"`
		Path     string        `json:"path,omitempty"`
		Status   string        `json:"status"`
		Models   []agent.Model `json:"models,omitempty"`
		Error    string        `json:"error,omitempty"`
	}
	out := []providerModels{}
	for _, d := range provider.DetectAll(context.Background()) {
		entry := providerModels{Provider: d.Provider, Path: d.Path, Status: d.Status, Error: d.Error}
		if d.Status == "available" {
			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			models, err := agent.ListModels(ctx, d.Provider, d.Path)
			cancel()
			if err != nil {
				entry.Error = err.Error()
			} else {
				entry.Models = models
			}
		}
		out = append(out, entry)
	}
	if *jsonOut {
		enc := json.NewEncoder(stdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}
	for _, entry := range out {
		fmt.Printf("%s: %s", entry.Provider, entry.Status)
		if entry.Error != "" {
			fmt.Printf(" (%s)", entry.Error)
		}
		fmt.Println()
		for _, model := range entry.Models {
			fmt.Printf("  %s\n", model.ID)
		}
	}
	return nil
}
