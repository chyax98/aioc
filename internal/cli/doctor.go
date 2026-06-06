package cli

import (
	"context"
	"fmt"
	"runtime"

	"aioc/pkg/provider"
)

func doctor(args []string) error {
	_ = args
	detections := provider.DetectAllWithVersions(context.Background())
	available := 0
	unsupported := 0
	missing := 0
	for _, d := range detections {
		switch d.Status {
		case "available":
			available++
		case "unsupported":
			unsupported++
		default:
			missing++
		}
	}
	fmt.Printf("aioc_go=%s\n", runtime.Version())
	fmt.Printf("aioc_os_arch=%s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("providers_available=%d\n", available)
	fmt.Printf("providers_unsupported=%d\n", unsupported)
	fmt.Printf("providers_missing=%d\n", missing)
	for _, d := range detections {
		path := d.Path
		if path == "" {
			path = "-"
		}
		version := d.Version
		if version == "" {
			version = "-"
		}
		fmt.Printf("provider=%s status=%s protocol=%s version=%q path=%q", d.Provider, d.Status, d.Protocol, version, path)
		if d.Error != "" {
			fmt.Printf(" error=%q", d.Error)
		}
		fmt.Println()
	}
	if available == 0 {
		return fmt.Errorf("no provider found on PATH")
	}
	return nil
}
