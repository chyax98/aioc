package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"aioc/pkg/events"
)

type runHistory struct {
	ID   string
	Path string
	file *os.File
}

func openRunHistory(providerName string) (*runHistory, error) {
	if !saveRunsEnabled() {
		return nil, nil
	}
	dir, err := runsDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	id := newRunID(providerName)
	path := filepath.Join(dir, id+".jsonl")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, err
	}
	return &runHistory{ID: id, Path: path, file: f}, nil
}

func (h *runHistory) Close() error {
	if h == nil || h.file == nil {
		return nil
	}
	return h.file.Close()
}

func saveRunsEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("AIOC_SAVE_RUNS")))
	return v != "0" && v != "false" && v != "no" && v != "off"
}

func runsDir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("AIOC_RUNS_DIR")); v != "" {
		return expandRunHome(v)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".aioc", "runs"), nil
}

func expandRunHome(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func newRunID(providerName string) string {
	providerName = strings.TrimSpace(providerName)
	if providerName == "" {
		providerName = "unknown"
	}
	providerName = sanitizeRunIDPart(providerName)
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102T150405Z") + "-" + providerName
	}
	return time.Now().UTC().Format("20060102T150405Z") + "-" + providerName + "-" + hex.EncodeToString(b[:])
}

func sanitizeRunIDPart(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "unknown"
	}
	return out
}

func runs(args []string) error {
	if len(args) == 0 {
		return listRuns(20)
	}
	switch args[0] {
	case "list", "ls":
		return listRuns(20)
	case "events", "cat", "show":
		id := "latest"
		if len(args) > 1 {
			id = args[1]
		}
		path, err := resolveRunPath(id)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(os.Stdout, f)
		return err
	case "output", "out":
		id := "latest"
		if len(args) > 1 {
			id = args[1]
		}
		path, err := resolveRunPath(id)
		if err != nil {
			return err
		}
		out, ok, err := extractRunOutput(path)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("run has no done output: %s", path)
		}
		fmt.Fprintln(os.Stdout, out)
		return nil
	case "path":
		id := "latest"
		if len(args) > 1 {
			id = args[1]
		}
		path, err := resolveRunPath(id)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, path)
		return nil
	case "help", "--help", "-h":
		printRunsHelp()
		return nil
	default:
		return fmt.Errorf("unknown runs command %q", args[0])
	}
}

func listRuns(limit int) error {
	files, err := runFilesDesc()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("no aioc runs found")
	}
	if limit > 0 && len(files) > limit {
		files = files[:limit]
	}
	for _, path := range files {
		fmt.Fprintln(os.Stdout, strings.TrimSuffix(filepath.Base(path), ".jsonl"))
	}
	return nil
}

func resolveRunPath(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" || id == "latest" {
		files, err := runFilesDesc()
		if err != nil {
			return "", err
		}
		if len(files) == 0 {
			return "", errors.New("no aioc runs found")
		}
		return files[0], nil
	}
	if strings.ContainsRune(id, filepath.Separator) || strings.HasSuffix(id, ".jsonl") {
		return id, nil
	}
	dir, err := runsDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, id+".jsonl")
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func runFilesDesc() ([]string, error) {
	dir, err := runsDir()
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	return files, nil
}

func extractRunOutput(path string) (string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var output string
	found := false
	for {
		var ev events.Event
		err := dec.Decode(&ev)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", false, err
		}
		if ev.Type == events.TypeDone {
			output = ev.Output
			found = true
		}
	}
	return output, found, nil
}

func printRunsHelp() {
	fmt.Println("Usage:")
	fmt.Println("  aioc runs                  # list recent run ids")
	fmt.Println("  aioc runs output [id]      # print done.output, default latest")
	fmt.Println("  aioc runs events [id]      # print raw JSONL, default latest")
	fmt.Println("  aioc runs path [id]        # print JSONL path, default latest")
	fmt.Println("")
	fmt.Println("Env:")
	fmt.Println("  AIOC_RUNS_DIR=~/.aioc/runs")
	fmt.Println("  AIOC_SAVE_RUNS=0           # disable saving")
}
