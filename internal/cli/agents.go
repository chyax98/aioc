package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

import "aioc/pkg/provider"

func agents(args []string) error {
	jsonOut := hasFlag(args, "--json")
	withVersions := hasFlag(args, "--versions")
	var detections []provider.Detection
	if withVersions {
		detections = provider.DetectAllWithVersions(context.Background())
	} else {
		detections = provider.DetectAll(context.Background())
	}
	if jsonOut {
		enc := json.NewEncoder(stdout())
		enc.SetIndent("", "  ")
		return enc.Encode(detections)
	}
	for _, d := range detections {
		path := d.Path
		if path == "" {
			path = "-"
		}
		version := d.Version
		if version == "" {
			version = "-"
		}
		if withVersions {
			fmt.Printf("%-12s %-12s %-8s %-18s %s\n", d.Provider, d.Status, d.Protocol, version, path)
		} else {
			fmt.Printf("%-12s %-12s %-8s %s\n", d.Provider, d.Status, d.Protocol, path)
		}
	}
	return nil
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if strings.TrimSpace(arg) == flag {
			return true
		}
	}
	return false
}
