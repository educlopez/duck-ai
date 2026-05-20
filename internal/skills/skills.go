package skills

import (
	"fmt"
	"os"
	"path/filepath"
)

// Skill represents a single duck-ai skill or command.
type Skill struct {
	Name    string // directory/file name
	SrcPath string // absolute source path
}

// LinkResult holds the outcome of a single symlink operation.
type LinkResult struct {
	Agent  string
	Skill  string
	Status string // "linked", "already_linked", "skipped", "error"
	Err    error
}

// DiscoverSkills scans the skills/ dir and returns all skills (subdirs with SKILL.md).
func DiscoverSkills(repoRoot string) ([]Skill, error) {
	skillsDir := filepath.Join(repoRoot, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("reading skills dir %s: %w", skillsDir, err)
	}

	var skills []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillMD := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		if _, err := os.Stat(skillMD); err != nil {
			continue // no SKILL.md → not a skill
		}
		skills = append(skills, Skill{
			Name:    e.Name(),
			SrcPath: filepath.Join(skillsDir, e.Name()),
		})
	}
	return skills, nil
}

// DiscoverCommands scans claude/commands/*.md and returns all commands.
func DiscoverCommands(repoRoot string) ([]Skill, error) {
	commandsDir := filepath.Join(repoRoot, "claude", "commands")
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading commands dir %s: %w", commandsDir, err)
	}

	var commands []Skill
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		commands = append(commands, Skill{
			Name:    e.Name(),
			SrcPath: filepath.Join(commandsDir, e.Name()),
		})
	}
	return commands, nil
}

// Link creates a symlink at dstDir/skill.Name → skill.SrcPath.
// Returns a LinkResult describing the outcome.
func Link(agentName string, skill Skill, dstDir string) LinkResult {
	result := LinkResult{Agent: agentName, Skill: skill.Name}

	dst := filepath.Join(dstDir, skill.Name)
	info, err := os.Lstat(dst)

	if err == nil {
		// Something exists at dst
		if info.Mode()&os.ModeSymlink != 0 {
			// It's a symlink — check target
			target, err := os.Readlink(dst)
			if err == nil && target == skill.SrcPath {
				result.Status = "already_linked"
				return result
			}
			// Wrong target → remove and re-link
			if err := os.Remove(dst); err != nil {
				result.Status = "error"
				result.Err = fmt.Errorf("removing stale symlink %s: %w", dst, err)
				return result
			}
		} else {
			// Real directory/file exists → skip with warning
			result.Status = "skipped"
			result.Err = fmt.Errorf("%s exists and is not a symlink", dst)
			return result
		}
	} else if !os.IsNotExist(err) {
		result.Status = "error"
		result.Err = fmt.Errorf("stat %s: %w", dst, err)
		return result
	}

	// Ensure parent dir exists
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		result.Status = "error"
		result.Err = fmt.Errorf("creating dir %s: %w", dstDir, err)
		return result
	}

	// Create symlink
	if err := os.Symlink(skill.SrcPath, dst); err != nil {
		result.Status = "error"
		result.Err = fmt.Errorf("symlinking %s → %s: %w", dst, skill.SrcPath, err)
		return result
	}

	result.Status = "linked"
	return result
}

// CheckLink checks the state of an existing symlink without modifying anything.
// Returns "ok", "broken", "missing", "not_symlink".
func CheckLink(dstDir, name string) string {
	dst := filepath.Join(dstDir, name)
	info, err := os.Lstat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return "missing"
		}
		return "error"
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return "not_symlink"
	}
	target, err := os.Readlink(dst)
	if err != nil {
		return "broken"
	}
	if _, err := os.Stat(target); err != nil {
		return "broken"
	}
	return "ok"
}
