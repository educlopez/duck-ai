// Package reports holds the writer-based renderers for `duck-ai doctor` and
// `duck-ai registry`. Both the cmd CLI wrappers and the TUI screens consume
// these functions. Keeping them here breaks the cycle between cmd and
// internal/tui (cmd imports internal/tui via install.go).
package reports

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/educlopez/duck-ai/internal/agents"
)

// Doctor writes the duck-ai doctor report to w. repoRoot is the absolute or
// relative path to the duck-ai source repo.
func Doctor(w io.Writer, repoRoot string) error {
	absRepo, err := filepath.Abs(repoRoot)
	if err != nil {
		absRepo = repoRoot
	}

	for _, a := range agents.All() {
		detected := a.Detect()
		skillsDir := a.SkillsDir()
		commandsDir := a.CommandsDir()

		fmt.Fprintf(w, "\n  %s (%s)\n", a.DisplayName(), a.ID())
		fmt.Fprintf(w, "    detected:     %s\n", yesNo(detected))
		fmt.Fprintf(w, "    skills dir:   %s\n", displayPath(skillsDir))
		fmt.Fprintf(w, "    commands dir: %s\n", displayPath(commandsDir))

		if !detected {
			continue
		}

		skillsManaged, skillsUnmanaged := scanDir(skillsDir, absRepo)
		commandsManaged, commandsUnmanaged := scanDir(commandsDir, absRepo)
		managed := skillsManaged + commandsManaged
		fmt.Fprintf(w, "    managed:      %d duck-ai symlinks\n", managed)

		unmanaged := append([]driftEntry{}, skillsUnmanaged...)
		unmanaged = append(unmanaged, commandsUnmanaged...)
		if len(unmanaged) > 0 {
			fmt.Fprintf(w, "    unmanaged:    %d entries (not managed by duck-ai)\n", len(unmanaged))
			for _, u := range unmanaged {
				fmt.Fprintf(w, "      - %s (%s)\n", u.relPath, u.kind)
			}
			fmt.Fprintf(w, "    hint: run `duck-ai update` to absorb colliding entries; non-colliding ones will be left alone.\n")
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
// directory.
type driftEntry struct {
	relPath string
	kind    string // "file" or "dir"
}

// scanDir walks dir one level deep and returns the count of duck-ai-managed
// symlinks plus a slice of unmanaged entries (real files/dirs, or symlinks
// pointing outside repoRoot). Hidden entries are skipped.
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
