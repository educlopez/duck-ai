package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/educlopez/duck-ai/internal/agents"
	"github.com/educlopez/duck-ai/internal/skills"
	"github.com/educlopez/duck-ai/internal/tui"
)

// RunInstallTUI launches the interactive TUI installer.
func RunInstallTUI(repoRoot string) error {
	m := tui.New(repoRoot)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunInstallAgent installs to a single named agent non-interactively.
func RunInstallAgent(repoRoot, agentName string) error {
	allAgents := agents.All()
	var target *agents.Agent
	for i, a := range allAgents {
		if a.Name == agentName {
			target = &allAgents[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("unknown agent %q (supported: claude, agents, codex, opencode)", agentName)
	}
	return installToAgents(repoRoot, []agents.Agent{*target})
}

// RunInstallAll installs to all detected agents non-interactively.
func RunInstallAll(repoRoot string) error {
	detected := agents.Detected()
	if len(detected) == 0 {
		fmt.Fprintln(os.Stderr, "No agents detected.")
		return nil
	}
	return installToAgents(repoRoot, detected)
}

func installToAgents(repoRoot string, agentList []agents.Agent) error {
	allSkills, err := skills.DiscoverSkills(repoRoot)
	if err != nil {
		return err
	}
	allCommands, err := skills.DiscoverCommands(repoRoot)
	if err != nil {
		return err
	}

	linked, already, errored, skipped := 0, 0, 0, 0

	for _, agent := range agentList {
		fmt.Printf("\n  Agent: %s\n", agent.Name)

		for _, skill := range allSkills {
			r := skills.Link(agent.Name, skill, agent.SkillsDir)
			printResult(r)
			switch r.Status {
			case "linked":
				linked++
			case "already_linked":
				already++
			case "error":
				errored++
			case "skipped":
				skipped++
			}
		}

		if agent.CommandsDir != "" {
			for _, cmd := range allCommands {
				r := skills.Link(agent.Name, cmd, agent.CommandsDir)
				printResult(r)
				switch r.Status {
				case "linked":
					linked++
				case "already_linked":
					already++
				case "error":
					errored++
				case "skipped":
					skipped++
				}
			}
		}
	}

	fmt.Printf("\n  Done — linked: %d  already: %d  skipped: %d  errors: %d\n",
		linked, already, skipped, errored)
	return nil
}

func printResult(r skills.LinkResult) {
	switch r.Status {
	case "linked":
		fmt.Printf("    ✓  %s\n", r.Skill)
	case "already_linked":
		fmt.Printf("    ~  %s (already linked)\n", r.Skill)
	case "skipped":
		fmt.Printf("    ⚠  %s (skipped: %v)\n", r.Skill, r.Err)
	case "error":
		fmt.Printf("    ✗  %s (error: %v)\n", r.Skill, r.Err)
	}
}
