package agents

import (
	"os"
	"os/exec"
	"path/filepath"
)

type openCodeAdapter struct{}

func (openCodeAdapter) ID() string          { return IDOpenCode }
func (openCodeAdapter) DisplayName() string { return "OpenCode" }

func (openCodeAdapter) Detect() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	if dirExists(filepath.Join(home, ".config", "opencode")) {
		return true
	}
	_, err = exec.LookPath("opencode")
	return err == nil
}

func (openCodeAdapter) SkillsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "opencode", "skills")
}

func (openCodeAdapter) CommandsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "opencode", "commands")
}
