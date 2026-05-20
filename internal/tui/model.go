package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/educlopez/duck-ai/internal/agents"
	"github.com/educlopez/duck-ai/internal/skills"
)

// screen represents which TUI screen is active.
type screen int

const (
	screenWelcome screen = iota
	screenAgents
	screenSkills
	screenInstalling
	screenDone
)

// Model is the bubbletea model for the installer TUI.
type Model struct {
	screen screen
	repoRoot string

	// Welcome screen
	welcomeCursor int

	// Agent screen
	allAgents      []agents.Agent
	agentSelected  []bool
	agentCursor    int

	// Skills screen
	allSkills     []skills.Skill
	allCommands   []skills.Skill
	skillSelected []bool
	skillCursor   int

	// Installing / Done
	results []skills.LinkResult
	done    bool

	// Terminal size
	width  int
	height int

	// Error
	err error
}

// New creates a fresh Model.
func New(repoRoot string) Model {
	detected := agents.Detected()
	agentSelected := make([]bool, len(detected))
	for i := range agentSelected {
		agentSelected[i] = true
	}

	return Model{
		screen:        screenWelcome,
		repoRoot:      repoRoot,
		allAgents:     detected,
		agentSelected: agentSelected,
	}
}

// -- Init --

func (m Model) Init() tea.Cmd {
	return nil
}

// -- Update --

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case installDoneMsg:
		m.results = msg.results
		m.screen = screenDone
		return m, nil
	}
	return m, nil
}

type installDoneMsg struct {
	results []skills.LinkResult
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenWelcome:
		return m.handleWelcomeKey(msg)
	case screenAgents:
		return m.handleAgentsKey(msg)
	case screenSkills:
		return m.handleSkillsKey(msg)
	case screenDone:
		if msg.String() == "q" || msg.String() == "ctrl+c" || msg.String() == "enter" {
			return m, tea.Quit
		}
	}
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleWelcomeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.welcomeCursor > 0 {
			m.welcomeCursor--
		}
	case "down", "j":
		if m.welcomeCursor < 1 {
			m.welcomeCursor++
		}
	case "enter", " ":
		switch m.welcomeCursor {
		case 0: // Install
			m.screen = screenAgents
		case 1: // Quit
			return m, tea.Quit
		}
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleAgentsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.agentCursor > 0 {
			m.agentCursor--
		}
	case "down", "j":
		if m.agentCursor < len(m.allAgents)-1 {
			m.agentCursor++
		}
	case " ":
		if len(m.agentSelected) > m.agentCursor {
			m.agentSelected[m.agentCursor] = !m.agentSelected[m.agentCursor]
		}
	case "enter":
		return m.loadSkillsScreen()
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) loadSkillsScreen() (tea.Model, tea.Cmd) {
	s, err := skills.DiscoverSkills(m.repoRoot)
	if err != nil {
		m.err = err
		return m, tea.Quit
	}
	c, err := skills.DiscoverCommands(m.repoRoot)
	if err != nil {
		m.err = err
		return m, tea.Quit
	}
	m.allSkills = s
	m.allCommands = c
	m.skillSelected = make([]bool, len(s))
	for i := range m.skillSelected {
		m.skillSelected[i] = true
	}
	m.screen = screenSkills
	return m, nil
}

func (m Model) handleSkillsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.skillCursor > 0 {
			m.skillCursor--
		}
	case "down", "j":
		if m.skillCursor < len(m.allSkills)-1 {
			m.skillCursor++
		}
	case " ":
		if len(m.skillSelected) > m.skillCursor {
			m.skillSelected[m.skillCursor] = !m.skillSelected[m.skillCursor]
		}
	case "a":
		// Toggle all
		allOn := true
		for _, s := range m.skillSelected {
			if !s {
				allOn = false
				break
			}
		}
		for i := range m.skillSelected {
			m.skillSelected[i] = !allOn
		}
	case "enter":
		m.screen = screenInstalling
		return m, m.runInstall()
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) runInstall() tea.Cmd {
	return func() tea.Msg {
		var results []skills.LinkResult

		// Collect selected agents
		var selectedAgents []agents.Agent
		for i, a := range m.allAgents {
			if i < len(m.agentSelected) && m.agentSelected[i] {
				selectedAgents = append(selectedAgents, a)
			}
		}

		// Collect selected skills
		var selectedSkills []skills.Skill
		for i, s := range m.allSkills {
			if i < len(m.skillSelected) && m.skillSelected[i] {
				selectedSkills = append(selectedSkills, s)
			}
		}

		for _, agent := range selectedAgents {
			// Install skills
			for _, skill := range selectedSkills {
				r := skills.Link(agent.Name, skill, agent.SkillsDir)
				results = append(results, r)
			}
			// Install commands (only if agent supports commands dir)
			if agent.CommandsDir != "" {
				for _, cmd := range m.allCommands {
					r := skills.Link(agent.Name, cmd, agent.CommandsDir)
					results = append(results, r)
				}
			}
		}

		return installDoneMsg{results: results}
	}
}

// -- View --

func (m Model) View() string {
	var b strings.Builder

	header := styleTitle.Render(banner)
	b.WriteString(header + "\n")
	b.WriteString(styleSubtitle.Render("  Personal Claude Code toolkit for Cinetic/Neven work") + "\n\n")

	switch m.screen {
	case screenWelcome:
		b.WriteString(m.viewWelcome())
	case screenAgents:
		b.WriteString(m.viewAgents())
	case screenSkills:
		b.WriteString(m.viewSkills())
	case screenInstalling:
		b.WriteString(m.viewInstalling())
	case screenDone:
		b.WriteString(m.viewDone())
	}

	b.WriteString("\n" + styleKey.Render("  ctrl+c / q  quit"))

	return b.String()
}

func (m Model) viewWelcome() string {
	var b strings.Builder
	items := []string{"  Install skills & commands", "  Quit"}
	b.WriteString(styleAccent.Render("  What would you like to do?\n\n"))
	for i, item := range items {
		if i == m.welcomeCursor {
			b.WriteString(styleCursor.Render("  ❯ ") + styleSelected.Render(item) + "\n")
		} else {
			b.WriteString(styleMuted.Render("    "+item) + "\n")
		}
	}
	b.WriteString("\n" + styleKey.Render("  ↑/↓ navigate  •  enter select"))
	return b.String()
}

func (m Model) viewAgents() string {
	var b strings.Builder
	b.WriteString(styleAccent.Render("  Detected agents\n\n"))

	if len(m.allAgents) == 0 {
		b.WriteString(styleError.Render("  No agents detected. Install claude, codex, agents, or opencode first.\n"))
		return b.String()
	}

	for i, a := range m.allAgents {
		checked := "[ ]"
		if i < len(m.agentSelected) && m.agentSelected[i] {
			checked = styleSuccess.Render("[✓]")
		}
		cursor := "  "
		name := styleMuted.Render(a.Name)
		if i == m.agentCursor {
			cursor = styleCursor.Render("❯ ")
			name = styleSelected.Render(a.Name)
		}
		b.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, checked, name))
	}

	b.WriteString("\n" + styleKey.Render("  ↑/↓ navigate  •  space toggle  •  enter confirm"))
	return b.String()
}

func (m Model) viewSkills() string {
	var b strings.Builder
	b.WriteString(styleAccent.Render("  Select skills to install\n\n"))

	for i, s := range m.allSkills {
		checked := "[ ]"
		if i < len(m.skillSelected) && m.skillSelected[i] {
			checked = styleSuccess.Render("[✓]")
		}
		cursor := "  "
		name := styleMuted.Render(s.Name)
		if i == m.skillCursor {
			cursor = styleCursor.Render("❯ ")
			name = styleSelected.Render(s.Name)
		}
		b.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, checked, name))
	}

	if len(m.allCommands) > 0 {
		b.WriteString("\n" + styleAccent.Render("  Commands (always installed)\n"))
		for _, c := range m.allCommands {
			b.WriteString(styleMuted.Render(fmt.Sprintf("    • %s", c.Name)) + "\n")
		}
	}

	b.WriteString("\n" + styleKey.Render("  ↑/↓ navigate  •  space toggle  •  a toggle all  •  enter install"))
	return b.String()
}

func (m Model) viewInstalling() string {
	return styleAccent.Render("  Installing...") + "\n\n" +
		styleMuted.Render("  Please wait while skills are being symlinked.")
}

func (m Model) viewDone() string {
	var b strings.Builder
	b.WriteString(styleSuccess.Render("  Installation complete!\n\n"))

	linked, already, skipped, errored := 0, 0, 0, 0
	for _, r := range m.results {
		switch r.Status {
		case "linked":
			linked++
		case "already_linked":
			already++
		case "skipped":
			skipped++
		case "error":
			errored++
		}
	}

	summary := styleBox.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			styleSuccess.Render(fmt.Sprintf("  ✓  Linked:        %d", linked)),
			styleMuted.Render(fmt.Sprintf("  ~  Already linked: %d", already)),
			styleError.Render(fmt.Sprintf("  ✗  Errors:         %d", errored)),
			styleMuted.Render(fmt.Sprintf("  ⚠  Skipped:        %d", skipped)),
		),
	)
	b.WriteString(summary + "\n")

	if errored > 0 {
		b.WriteString("\n" + styleError.Render("  Errors:\n"))
		for _, r := range m.results {
			if r.Status == "error" || r.Status == "skipped" {
				b.WriteString(styleMuted.Render(fmt.Sprintf("    [%s] %s/%s: %v\n", r.Agent, r.Agent, r.Skill, r.Err)))
			}
		}
	}

	b.WriteString("\n" + styleKey.Render("  enter / q  exit"))
	return b.String()
}
