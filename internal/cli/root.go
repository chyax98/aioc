package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	aiocconfig "aioc/internal/config"
	"aioc/pkg/provider"
)

func Execute(args []string, version string) error {
	args, loadOpts := splitGlobalConfigArgs(args)
	if _, err := aiocconfig.LoadAutoWithOptions(loadOpts); err != nil {
		return err
	}
	if len(args) == 0 {
		printHelp(version)
		return nil
	}
	args = normalizeRootArgs(args)
	switch args[0] {
	case "agents":
		return agents(args[1:])
	case "doctor":
		return doctor(args[1:])
	case "models":
		return models(args[1:])
	case "models-all":
		return allModels(args[1:])
	case "config":
		return configCmd(args[1:])
	case "runs", "history":
		return runs(args[1:])
	case "run":
		return run(args[1:])
	case "usage", "guide", "manual":
		return usage(args[1:])
	case "prompt":
		return prompt(args[1:])
	case "skills", "skill":
		return skills(args[1:])
	case "version", "--version", "-v":
		fmt.Println(version)
		return nil
	case "help", "--help", "-h":
		printHelp(version)
		return nil
	default:
		return run(args)
	}
}

func splitGlobalConfigArgs(args []string) ([]string, aiocconfig.LoadOptions) {
	clean := make([]string, 0, len(args))
	opts := aiocconfig.LoadOptions{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--no-config":
			opts.NoConfig = true
		case arg == "--config" && i+1 < len(args):
			opts.Paths = append(opts.Paths, args[i+1])
			i++
		case strings.HasPrefix(arg, "--config="):
			opts.Paths = append(opts.Paths, strings.TrimPrefix(arg, "--config="))
		default:
			clean = append(clean, arg)
		}
	}
	return clean, opts
}

func normalizeRootArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	switch args[0] {
	case "agents", "doctor", "models", "models-all", "config", "runs", "history", "run", "usage", "guide", "manual", "prompt", "skills", "skill", "version", "--version", "-v", "help", "--help", "-h":
		return args
	}
	if _, ok := provider.SpecByName(args[0]); ok {
		return append([]string{"run", "-p", args[0]}, args[1:]...)
	}
	if strings.HasPrefix(args[0], "-") {
		return append([]string{"run"}, args...)
	}
	return append([]string{"run"}, args...)
}

func PrintError(err error) {
	if err == nil || errors.Is(err, errSilent) {
		return
	}
	fmt.Fprintln(os.Stderr, "Error:", err)
}

var errSilent = errors.New("silent")

func printHelp(version string) {
	fmt.Printf("aioc %s\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  aioc agents [--json] [--versions]")
	fmt.Println("  aioc doctor")
	fmt.Println("  aioc models -p <provider> [--json]")
	fmt.Println("  aioc <prompt>                              # default: run with auto provider")
	fmt.Println("  aioc <provider> <prompt>                   # provider shorthand, e.g. aioc claude \"review\"")
	fmt.Println("  aioc [--config path] [--no-config] ...")
	fmt.Println("  aioc config                               # show loaded config files")
	fmt.Println("  aioc runs output [latest|id]              # print saved final output")
	fmt.Println("  aioc runs events [latest|id]              # print saved JSONL events")
	fmt.Println("  aioc run [--cwd dir] [--model model] [--thinking level] <prompt>")
	fmt.Println("  aioc run -p <provider> <prompt>            # explicit provider when needed")
	fmt.Println("  aioc models-all [--json]")
	fmt.Println("  aioc usage")
	fmt.Println("  aioc prompt")
	fmt.Println("  aioc skills get")
	fmt.Println("  aioc skills install   # installs minimal usage shadow, not aioc itself")
}
