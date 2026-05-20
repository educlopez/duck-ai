package skillregistry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/educlopez/duck-ai/internal/agents"
)

// Manifest is a minimal record of one skill or command on disk.
type Manifest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Path        string `json:"path"`
	Kind        string `json:"kind"` // "skill" or "command"
}

// ParseSource walks the duck-ai repo and returns one Manifest per source
// skill (skills/<name>/SKILL.md) and command (claude/commands/<name>.md).
func ParseSource(repoRoot string) ([]Manifest, error) {
	var out []Manifest

	skillsDir := filepath.Join(repoRoot, "skills")
	skillEntries, err := os.ReadDir(skillsDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read skills dir: %w", err)
	}
	for _, e := range skillEntries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		skillFile := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue
		}
		m, err := parseFile(skillFile, "skill", e.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}

	commandsDir := filepath.Join(repoRoot, "claude", "commands")
	cmdEntries, err := os.ReadDir(commandsDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read commands dir: %w", err)
	}
	for _, e := range cmdEntries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		m, err := parseFile(filepath.Join(commandsDir, e.Name()), "command", name)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}

	sortManifests(out)
	return out, nil
}

// ParseInstalled walks the adapter's SkillsDir and CommandsDir one level deep
// and parses any symlink targeting a duck-ai source file. Non-symlink entries
// are skipped (doctor surfaces them as drift).
func ParseInstalled(adapter agents.Adapter) ([]Manifest, error) {
	var out []Manifest

	skillsDir := adapter.SkillsDir()
	if skillsDir != "" {
		ms, err := walkInstalledDir(skillsDir, "skill")
		if err != nil {
			return nil, err
		}
		out = append(out, ms...)
	}

	commandsDir := adapter.CommandsDir()
	if commandsDir != "" {
		ms, err := walkInstalledDir(commandsDir, "command")
		if err != nil {
			return nil, err
		}
		out = append(out, ms...)
	}

	sortManifests(out)
	return out, nil
}

func walkInstalledDir(dir, kind string) ([]Manifest, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}

	var out []Manifest
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
		if info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		target, err := os.Readlink(full)
		if err != nil {
			continue
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(full), target)
		}

		var manifestPath, displayName string
		switch kind {
		case "skill":
			manifestPath = filepath.Join(target, "SKILL.md")
			displayName = name
		case "command":
			manifestPath = target
			displayName = strings.TrimSuffix(name, ".md")
		}

		if _, err := os.Stat(manifestPath); err != nil {
			continue
		}
		m, err := parseFile(manifestPath, kind, displayName)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func parseFile(path, kind, fallbackName string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read %s: %w", path, err)
	}
	m := Manifest{Path: path, Kind: kind, Name: fallbackName}
	fields := parseFrontmatter(data)
	if v, ok := fields["name"]; ok && v != "" {
		m.Name = v
	}
	if v, ok := fields["description"]; ok {
		m.Description = v
	}
	if v, ok := fields["version"]; ok {
		m.Version = v
	}
	return m, nil
}

// parseFrontmatter extracts a minimal map of top-level key/value pairs from
// YAML frontmatter delimited by leading and trailing "---" lines. Folded or
// multi-line values join on a single space. Nested keys are ignored.
func parseFrontmatter(data []byte) map[string]string {
	out := map[string]string{}
	text := string(data)
	if !strings.HasPrefix(text, "---") {
		return out
	}
	lines := strings.Split(text, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return out
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return out
	}

	var currentKey string
	var folded []string
	flush := func() {
		if currentKey == "" {
			return
		}
		out[currentKey] = strings.TrimSpace(strings.Join(folded, " "))
		currentKey = ""
		folded = nil
	}

	for i := 1; i < end; i++ {
		line := lines[i]
		trimmed := strings.TrimRight(line, " \t")
		if trimmed == "" {
			continue
		}
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			if currentKey != "" {
				folded = append(folded, strings.TrimSpace(line))
			}
			continue
		}
		idx := strings.Index(trimmed, ":")
		if idx < 0 {
			continue
		}
		flush()
		key := strings.TrimSpace(trimmed[:idx])
		value := strings.TrimSpace(trimmed[idx+1:])
		if value == ">" || value == "|" || value == ">-" || value == "|-" {
			currentKey = key
			continue
		}
		value = unquote(value)
		out[key] = value
	}
	flush()
	return out
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func sortManifests(in []Manifest) {
	sort.Slice(in, func(i, j int) bool {
		if in[i].Kind != in[j].Kind {
			return in[i].Kind < in[j].Kind
		}
		return in[i].Name < in[j].Name
	})
}
