package agents

import (
	"os"
	"os/exec"
	"path/filepath"
)

type codexAdapter struct{}

func (codexAdapter) ID() string          { return IDCodex }
func (codexAdapter) DisplayName() string { return "Codex" }

func (codexAdapter) Detect() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	if dirExists(filepath.Join(home, ".codex")) {
		return true
	}
	_, err = exec.LookPath("codex")
	return err == nil
}

func (codexAdapter) SkillsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".codex", "skills")
}

func (codexAdapter) CommandsDir() string { return "" }
