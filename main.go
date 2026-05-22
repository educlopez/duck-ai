package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/educlopez/duck-ai/cmd"
)

// version is the duck-ai release version.
// "dev" is overridden at build time by goreleaser via:
//
//	-ldflags "-X main.version={{.Version}}"
var version = "dev"

func main() {
	repoRoot := repoRootFromEnvOrBinary(version)

	args := os.Args[1:]

	// No args or "install" → TUI
	if len(args) == 0 || args[0] == "install" {
		installArgs := args
		if len(args) > 0 {
			installArgs = args[1:]
		}
		if err := handleInstall(repoRoot, installArgs, version); err != nil {
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

	case "update":
		uargs, err := cmd.ParseUpdateArgs(args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := cmd.RunUpdate(repoRoot, uargs); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "registry":
		rargs, err := cmd.ParseRegistryArgs(args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := cmd.RunRegistry(repoRoot, rargs); err != nil {
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

func handleInstall(repoRoot string, args []string, version string) error {
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
	return cmd.RunInstallTUI(repoRoot, version)
}

// repoRootFromEnvOrBinary resolves the duck-ai repo root.
//
// Priority:
//  1. DUCK_AI_DIR env var (explicit override)
//  2. Walk up from the binary looking for a skills/ sibling (dev mode:
//     running `go run .` or `./duck-ai` from the cloned repo)
//  3. Materialize the embedded source tree into ~/.duck-ai/source/<version>/
//     (release mode: binary installed via curl-pipe with no sibling repo)
//  4. Fallback to cwd if all of the above fail
func repoRootFromEnvOrBinary(version string) string {
	if dir := os.Getenv("DUCK_AI_DIR"); dir != "" {
		return dir
	}
	if exe, err := os.Executable(); err == nil {
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
	// No sibling repo — materialize the embedded source.
	if dir, err := materializeEmbeddedSource(version); err == nil {
		return dir
	} else {
		fmt.Fprintf(os.Stderr, "warning: could not materialize embedded source: %v\n", err)
	}
	// Last-resort fallback: cwd.
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
  duck-ai update                 Re-link skills/commands, backing up any conflicting files
  duck-ai update --dry-run       Show what update would change without touching disk
  duck-ai update --agent NAME    Update only NAME
  duck-ai update --yes           Skip confirmation prompts
  duck-ai update --list-backups  List backup batches under ~/.duck-ai/backups
  duck-ai update --restore TS    Restore files from backup TS (full stamp or unique prefix)
  duck-ai doctor                 Check symlink health per detected agent
  duck-ai registry               List skills/commands with versions per agent
  duck-ai registry --source      List source entries from the repo
  duck-ai registry --json        Emit machine-readable JSON
  duck-ai version                Print version

Environment:
  DUCK_AI_DIR   Override repo root (defaults to binary location)
`)
}
