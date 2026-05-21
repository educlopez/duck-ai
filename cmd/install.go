package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/educlopez/duck-ai/internal/agents"
	"github.com/educlopez/duck-ai/internal/skills"
	"github.com/educlopez/duck-ai/internal/tui"
)

func RunInstallTUI(repoRoot, version string) error {
	m := tui.New(repoRoot, version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func RunInstallAgent(repoRoot, agentID string) error {
	a, ok := agents.ByID(agentID)
	if !ok {
		return fmt.Errorf("unknown agent %q (supported: claude, codex, opencode, agents)", agentID)
	}
	if !a.Detect() {
		fmt.Fprintf(os.Stderr, "agent %q not detected on this system\n", a.ID())
		return nil
	}
	return installToAgents(repoRoot, []agents.Adapter{a})
}

func RunInstallAll(repoRoot string) error {
	detected := agents.Detected()
	if len(detected) == 0 {
		fmt.Fprintln(os.Stderr, "No agents detected.")
		return nil
	}
	return installToAgents(repoRoot, detected)
}

func installToAgents(repoRoot string, adapters []agents.Adapter) error {
	allSkills, err := skills.DiscoverSkills(repoRoot)
	if err != nil {
		return err
	}
	allCommands, err := skills.DiscoverCommands(repoRoot)
	if err != nil {
		return err
	}

	linked, already, errored, skipped := 0, 0, 0, 0

	for _, a := range adapters {
		fmt.Printf("\n  Agent: %s\n", a.ID())

		if !a.Detect() {
			fmt.Printf("    -  not detected, skipping\n")
			continue
		}

		skillsDir := a.SkillsDir()
		if skillsDir == "" {
			fmt.Printf("    -  no skills dir resolved, skipping\n")
		} else {
			for _, sk := range allSkills {
				r := skills.Link(a.ID(), sk, skillsDir)
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

		commandsDir := a.CommandsDir()
		if commandsDir != "" {
			for _, c := range allCommands {
				r := skills.Link(a.ID(), c, commandsDir)
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
