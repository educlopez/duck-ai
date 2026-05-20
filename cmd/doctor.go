package cmd

import (
	"fmt"
	"os"

	"github.com/educlopez/duck-ai/internal/agents"
	"github.com/educlopez/duck-ai/internal/skills"
)

// RunDoctor checks all symlinks per detected agent and reports status.
func RunDoctor(repoRoot string) error {
	allSkills, err := skills.DiscoverSkills(repoRoot)
	if err != nil {
		return err
	}
	allCommands, err := skills.DiscoverCommands(repoRoot)
	if err != nil {
		return err
	}

	detected := agents.Detected()
	if len(detected) == 0 {
		fmt.Fprintln(os.Stderr, "No agents detected.")
		return nil
	}

	hasIssue := false

	for _, agent := range detected {
		fmt.Printf("\n  Agent: %s\n", agent.Name)

		for _, skill := range allSkills {
			status := skills.CheckLink(agent.SkillsDir, skill.Name)
			printDoctorStatus(skill.Name, status)
			if status != "ok" {
				hasIssue = true
			}
		}

		if agent.CommandsDir != "" {
			for _, cmd := range allCommands {
				status := skills.CheckLink(agent.CommandsDir, cmd.Name)
				printDoctorStatus(cmd.Name, status)
				if status != "ok" {
					hasIssue = true
				}
			}
		}
	}

	fmt.Println()
	if hasIssue {
		fmt.Println("  Issues found. Run `duck-ai install` to fix.")
	} else {
		fmt.Println("  All symlinks OK.")
	}
	return nil
}

func printDoctorStatus(name, status string) {
	switch status {
	case "ok":
		fmt.Printf("    ✓  %s\n", name)
	case "missing":
		fmt.Printf("    ✗  %s (missing — not installed)\n", name)
	case "broken":
		fmt.Printf("    ✗  %s (broken symlink)\n", name)
	case "not_symlink":
		fmt.Printf("    ⚠  %s (exists but is not a symlink)\n", name)
	default:
		fmt.Printf("    ?  %s (%s)\n", name, status)
	}
}
