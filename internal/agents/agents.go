package agents

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Agent represents a supported AI agent with its directory paths.
type Agent struct {
	Name        string
	SkillsDir   string
	CommandsDir string // empty if not supported
}

// All returns all known agents with their resolved paths.
func All() []Agent {
	home, _ := os.UserHomeDir()

	return []Agent{
		{
			Name:        "claude",
			SkillsDir:   filepath.Join(home, ".claude", "skills"),
			CommandsDir: filepath.Join(home, ".claude", "commands"),
		},
		{
			Name:      "agents",
			SkillsDir: filepath.Join(home, ".agents", "skills"),
		},
		{
			Name:      "codex",
			SkillsDir: filepath.Join(home, ".codex", "skills"),
		},
		{
			Name:        "opencode",
			SkillsDir:   filepath.Join(home, ".config", "opencode", "skills"),
			CommandsDir: filepath.Join(home, ".config", "opencode", "commands"),
		},
	}
}

// Detected returns agents that appear to be installed on the system.
func Detected() []Agent {
	all := All()
	var detected []Agent
	for _, a := range all {
		if isDetected(a) {
			detected = append(detected, a)
		}
	}
	return detected
}

// isDetected checks whether the agent config directory exists or the binary is in PATH.
func isDetected(a Agent) bool {
	home, _ := os.UserHomeDir()

	switch a.Name {
	case "claude":
		return dirExists(filepath.Join(home, ".claude")) || inPath("claude")
	case "agents":
		return dirExists(filepath.Join(home, ".agents"))
	case "codex":
		return dirExists(filepath.Join(home, ".codex")) || inPath("codex")
	case "opencode":
		return dirExists(filepath.Join(home, ".config", "opencode")) || inPath("opencode")
	}
	return false
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func inPath(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}
