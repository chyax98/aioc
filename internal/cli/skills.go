package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"aioc/pkg/skill"
)

func skills(args []string) error {
	if len(args) == 0 {
		printSkillsHelp()
		return nil
	}
	switch args[0] {
	case "get", "show", "cat":
		return skillsGet(args[1:])
	case "install":
		return skillsInstall(args[1:])
	case "path":
		fmt.Println(defaultSkillInstallPath())
		return nil
	case "help", "--help", "-h":
		printSkillsHelp()
		return nil
	default:
		return fmt.Errorf("unknown skills command: %s", args[0])
	}
}

func skillsGet(args []string) error {
	fs := flag.NewFlagSet("skills get", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	output := fs.String("o", "", "write skill to file instead of stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output == "" {
		fmt.Print(skill.AIOCCLI)
		return nil
	}
	return writeSkillFile(*output)
}

func skillsInstall(args []string) error {
	fs := flag.NewFlagSet("skills install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", defaultSkillRoot(), "skills root directory")
	name := fs.String("name", skill.Name, "skill directory name")
	printPath := fs.Bool("print-path", false, "print installed SKILL.md path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dir == "" {
		return errors.New("missing install dir")
	}
	path := filepath.Join(expandHome(*dir), *name, skill.FileName)
	if err := writeSkillFile(path); err != nil {
		return err
	}
	if *printPath {
		fmt.Println(path)
	} else {
		fmt.Printf("installed usage shadow %s\n", path)
	}
	return nil
}

func writeSkillFile(path string) error {
	path = expandHome(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(skill.AIOCCLI), 0o644)
}

func defaultSkillRoot() string {
	return filepath.Join(userHome(), ".pi", "agent", "skills")
}

func defaultSkillInstallPath() string {
	return filepath.Join(defaultSkillRoot(), skill.Name, skill.FileName)
}

func expandHome(path string) string {
	if path == "~" {
		return userHome()
	}
	if len(path) >= 2 && path[:2] == "~/" {
		return filepath.Join(userHome(), path[2:])
	}
	return path
}

func userHome() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return home
}

func printSkillsHelp() {
	fmt.Println("Usage:")
	fmt.Println("  aioc skills get                  # print minimal discovery shadow")
	fmt.Println("  aioc skills get -o ./SKILL.md    # write shadow to file")
	fmt.Println("  aioc skills install              # copy shadow to ~/.pi/agent/skills")
	fmt.Println("  aioc skills install --dir ~/.agents/skills")
	fmt.Println("  aioc skills path")
	fmt.Println("")
	fmt.Println("Note: install writes only SKILL.md usage shadow, not aioc binary.")
}
