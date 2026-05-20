package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/educlopez/duck-ai/cmd"
)

const version = "0.1.0"

func main() {
	repoRoot := repoRootFromEnvOrBinary()

	args := os.Args[1:]

	// No args or "install" → TUI
	if len(args) == 0 || args[0] == "install" {
		installArgs := args
		if len(args) > 0 {
			installArgs = args[1:]
		}
		if err := handleInstall(repoRoot, installArgs); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch args[0] {
	case "doctor":
		if err := cmd.RunDoctor(repoRoot); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "version", "--version", "-v":
		fmt.Printf("duck-ai %s\n", version)

	case "help", "--help", "-h":
		printHelp()

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printHelp()
		os.Exit(1)
	}
}

func handleInstall(repoRoot string, args []string) error {
	// Parse flags
	agentFlag := ""
	allFlag := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--agent":
			if i+1 < len(args) {
				agentFlag = args[i+1]
				i++
			} else {
				return fmt.Errorf("--agent requires a value")
			}
		case "--all":
			allFlag = true
		}
	}

	if agentFlag != "" {
		return cmd.RunInstallAgent(repoRoot, agentFlag)
	}
	if allFlag {
		return cmd.RunInstallAll(repoRoot)
	}
	// Default: TUI
	return cmd.RunInstallTUI(repoRoot)
}

// repoRootFromEnvOrBinary resolves the duck-ai repo root.
// Priority: DUCK_AI_DIR env var → directory of the running binary.
func repoRootFromEnvOrBinary() string {
	if dir := os.Getenv("DUCK_AI_DIR"); dir != "" {
		return dir
	}
	exe, err := os.Executable()
	if err == nil {
		// Walk up from the binary looking for a skills/ directory.
		dir := filepath.Dir(exe)
		for {
			if _, err := os.Stat(filepath.Join(dir, "skills")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	// Fallback: cwd
	cwd, _ := os.Getwd()
	return cwd
}

func printHelp() {
	fmt.Print(`duck-ai — personal Claude Code toolkit

Usage:
  duck-ai                        Launch interactive TUI installer
  duck-ai install                Launch interactive TUI installer
  duck-ai install --agent NAME   Install only to NAME (claude|agents|codex|opencode)
  duck-ai install --all          Install to all detected agents non-interactively
  duck-ai doctor                 Check symlink health per detected agent
  duck-ai version                Print version

Environment:
  DUCK_AI_DIR   Override repo root (defaults to binary location)
`)
}
