package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/educlopez/duck-ai/internal/agents"
)

func RunDoctor(repoRoot string) error {
	absRepo, err := filepath.Abs(repoRoot)
	if err != nil {
		absRepo = repoRoot
	}

	for _, a := range agents.All() {
		detected := a.Detect()
		skillsDir := a.SkillsDir()
		commandsDir := a.CommandsDir()

		fmt.Printf("\n  %s (%s)\n", a.DisplayName(), a.ID())
		fmt.Printf("    detected:     %s\n", yesNo(detected))
		fmt.Printf("    skills dir:   %s\n", displayPath(skillsDir))
		fmt.Printf("    commands dir: %s\n", displayPath(commandsDir))

		if !detected {
			continue
		}

		skillsManaged, skillsUnmanaged := scanDir(skillsDir, absRepo)
		commandsManaged, commandsUnmanaged := scanDir(commandsDir, absRepo)
		managed := skillsManaged + commandsManaged
		fmt.Printf("    managed:      %d duck-ai symlinks\n", managed)

		unmanaged := append([]driftEntry{}, skillsUnmanaged...)
		unmanaged = append(unmanaged, commandsUnmanaged...)
		if len(unmanaged) > 0 {
			fmt.Printf("    unmanaged:    %d entries (not managed by duck-ai)\n", len(unmanaged))
			for _, u := range unmanaged {
				fmt.Printf("      - %s (%s)\n", u.relPath, u.kind)
			}
			fmt.Printf("    hint: run `duck-ai update` to absorb colliding entries; non-colliding ones will be left alone.\n")
		}
	}

	return nil
}

func yesNo(b bool) string {
	if b {
		return "y"
	}
	return "n"
}

func displayPath(p string) string {
	if p == "" {
		return "(none)"
	}
	return p
}

// driftEntry describes a single unmanaged file or directory found in an agent
// directory. relPath is rendered relative to the agent dir's parent so output
// is unambiguous (e.g. "skills/foo" rather than just "foo").
type driftEntry struct {
	relPath string
	kind    string // "file" or "dir"
}

// scanDir walks dir one level deep and returns:
//   - managed: count of symlinks pointing into repoRoot
//   - unmanaged: entries that are NOT such symlinks (real files/dirs, or
//     symlinks pointing elsewhere)
//
// Hidden entries (names starting with ".") are skipped to match what install
// currently ignores by convention (.DS_Store and other dotfiles).
func scanDir(dir, repoRoot string) (managed int, unmanaged []driftEntry) {
	if dir == "" {
		return 0, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, nil
	}
	prefix := filepath.Base(dir) // "skills" or "commands"
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(full)
			if err == nil {
				absTarget := target
				if !filepath.IsAbs(absTarget) {
					absTarget = filepath.Join(filepath.Dir(full), absTarget)
				}
				if strings.HasPrefix(absTarget, repoRoot+string(filepath.Separator)) || absTarget == repoRoot {
					managed++
					continue
				}
			}
		}
		kind := "file"
		if info.IsDir() {
			kind = "dir"
		}
		unmanaged = append(unmanaged, driftEntry{
			relPath: filepath.Join(prefix, name),
			kind:    kind,
		})
	}
	return managed, unmanaged
}
