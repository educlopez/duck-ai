package agents

import (
	"os"
	"path/filepath"
)

type genericAdapter struct{}

func (genericAdapter) ID() string          { return IDAgents }
func (genericAdapter) DisplayName() string { return "Agents (generic)" }

func (genericAdapter) Detect() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return dirExists(filepath.Join(home, ".agents"))
}

func (genericAdapter) SkillsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".agents", "skills")
}

func (genericAdapter) CommandsDir() string { return "" }
