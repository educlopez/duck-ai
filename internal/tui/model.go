package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/educlopez/duck-ai/internal/agents"
	"github.com/educlopez/duck-ai/internal/skills"
	"github.com/educlopez/duck-ai/internal/updater"
)

// screen represents which TUI screen is active.
type screen int

const (
	screenWelcome screen = iota
	screenAgents
	screenSkills
	screenInstalling
	screenDone
	screenUpdatePlanning
	screenUpdateConfirm
	screenUpdateNoop
	screenUpdating
	screenUpdateDone
)

// Model is the bubbletea model for the installer TUI.
type Model struct {
	screen   screen
	repoRoot string
	version  string

	// Welcome screen
	welcomeCursor int

	// Agent screen
	allAgents     []agents.Adapter
	agentSelected []bool
	agentCursor   int

	// Skills screen
	allSkills     []skills.Skill
	allCommands   []skills.Skill
	skillSelected []bool
	skillCursor   int

	// Installing / Done
	results []skills.LinkResult
	done    bool

	// Update flow
	updatePlan    *updater.Report
	updatePlanErr error
	updateReport  *updater.Report
	updateErr     error
	spinnerFrame  int
	spinnerActive bool

	// Terminal size
	width  int
	height int

	// Error
	err error
}

// New creates a fresh Model. version is shown in the welcome tagline; pass
// an empty string to hide the version suffix.
func New(repoRoot, version string) Model {
	detected := agents.Detected()
	agentSelected := make([]bool, len(detected))
	for i := range agentSelected {
		agentSelected[i] = true
	}

	return Model{
		screen:        screenWelcome,
		repoRoot:      repoRoot,
		version:       version,
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

	case updatePlanMsg:
		m.updatePlan = msg.report
		m.updatePlanErr = msg.err
		m.spinnerActive = false
		if msg.err != nil {
			m.screen = screenUpdateDone
			m.updateErr = msg.err
			return m, nil
		}
		if planIsNoop(msg.report) {
			m.screen = screenUpdateNoop
			return m, nil
		}
		m.screen = screenUpdateConfirm
		return m, nil

	case updateDoneMsg:
		m.updateReport = msg.report
		m.updateErr = msg.err
		m.spinnerActive = false
		m.screen = screenUpdateDone
		return m, nil

	case spinnerTickMsg:
		if !m.spinnerActive {
			return m, nil
		}
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		return m, tickSpinner()
	}
	return m, nil
}

type installDoneMsg struct {
	results []skills.LinkResult
}

type updatePlanMsg struct {
	report *updater.Report
	err    error
}

type updateDoneMsg struct {
	report *updater.Report
	err    error
}

type spinnerTickMsg struct{}

func tickSpinner() tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenWelcome:
		return m.handleWelcomeKey(msg)
	case screenAgents:
		return m.handleAgentsKey(msg)
	case screenSkills:
		return m.handleSkillsKey(msg)
	case screenUpdateConfirm:
		return m.handleUpdateConfirmKey(msg)
	case screenUpdateNoop:
		return m.handleUpdateNoopKey(msg)
	case screenDone:
		if msg.String() == "q" || msg.String() == "ctrl+c" || msg.String() == "enter" {
			return m, tea.Quit
		}
	case screenUpdateDone:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.resetToWelcome(), nil
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
		if m.welcomeCursor < 2 {
			m.welcomeCursor++
		}
	case "enter", " ":
		switch m.welcomeCursor {
		case 0:
			m.screen = screenAgents
		case 1:
			m.screen = screenUpdatePlanning
			m.spinnerActive = true
			m.spinnerFrame = 0
			return m, tea.Batch(m.runUpdatePlan(), tickSpinner())
		case 2:
			return m, tea.Quit
		}
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleUpdateConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.screen = screenUpdating
		m.spinnerActive = true
		m.spinnerFrame = 0
		return m, tea.Batch(m.runUpdate(), tickSpinner())
	case "n", "N", "esc":
		return m.resetToWelcome(), nil
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleUpdateNoopKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "enter", "esc", " ":
		return m.resetToWelcome(), nil
	}
	return m, nil
}

func (m Model) resetToWelcome() Model {
	m.screen = screenWelcome
	m.updatePlan = nil
	m.updatePlanErr = nil
	m.updateReport = nil
	m.updateErr = nil
	m.spinnerActive = false
	m.spinnerFrame = 0
	return m
}

func (m Model) runUpdatePlan() tea.Cmd {
	repoRoot := m.repoRoot
	return func() tea.Msg {
		rpt, err := updater.Run(updater.Options{RepoRoot: repoRoot, DryRun: true})
		return updatePlanMsg{report: rpt, err: err}
	}
}

func (m Model) runUpdate() tea.Cmd {
	repoRoot := m.repoRoot
	return func() tea.Msg {
		rpt, err := updater.Run(updater.Options{RepoRoot: repoRoot})
		return updateDoneMsg{report: rpt, err: err}
	}
}

func planIsNoop(rpt *updater.Report) bool {
	if rpt == nil || len(rpt.Agents) == 0 {
		return true
	}
	for _, ar := range rpt.Agents {
		if ar.Updated > 0 || ar.Replaced > 0 || ar.Missing > 0 || ar.Failed > 0 {
			return false
		}
	}
	return true
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

		var selectedAgents []agents.Adapter
		for i, a := range m.allAgents {
			if i < len(m.agentSelected) && m.agentSelected[i] {
				selectedAgents = append(selectedAgents, a)
			}
		}

		var selectedSkills []skills.Skill
		for i, s := range m.allSkills {
			if i < len(m.skillSelected) && m.skillSelected[i] {
				selectedSkills = append(selectedSkills, s)
			}
		}

		for _, agent := range selectedAgents {
			if !agent.Detect() {
				continue
			}
			skillsDir := agent.SkillsDir()
			if skillsDir != "" {
				for _, skill := range selectedSkills {
					r := skills.Link(agent.ID(), skill, skillsDir)
					results = append(results, r)
				}
			}
			commandsDir := agent.CommandsDir()
			if commandsDir != "" {
				for _, cmd := range m.allCommands {
					r := skills.Link(agent.ID(), cmd, commandsDir)
					results = append(results, r)
				}
			}
		}

		return installDoneMsg{results: results}
	}
}

// -- View --

func (m Model) View() string {
	// The welcome screen is fully self-contained (logo, tagline, menu,
	// help footer) and wrapped in its own FrameStyle. Render it directly
	// and skip the legacy header + footer that the other screens share.
	if m.screen == screenWelcome {
		return RenderWelcome(m.welcomeCursor, m.version)
	}

	var b strings.Builder

	logo := RenderLogo()
	tagline := Tagline(m.version)
	header := lipgloss.JoinVertical(lipgloss.Center,
		logo,
		styleTitle.Render(tagline),
	)
	b.WriteString(header + "\n\n")

	switch m.screen {
	case screenAgents:
		b.WriteString(m.viewAgents())
	case screenSkills:
		b.WriteString(m.viewSkills())
	case screenInstalling:
		b.WriteString(m.viewInstalling())
	case screenDone:
		b.WriteString(m.viewDone())
	case screenUpdatePlanning:
		b.WriteString(m.viewUpdatePlanning())
	case screenUpdateConfirm:
		b.WriteString(m.viewUpdateConfirm())
	case screenUpdateNoop:
		b.WriteString(m.viewUpdateNoop())
	case screenUpdating:
		b.WriteString(m.viewUpdating())
	case screenUpdateDone:
		b.WriteString(m.viewUpdateDone())
	}

	b.WriteString("\n" + styleKey.Render("  ctrl+c / q  quit"))

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
		name := styleMuted.Render(a.DisplayName())
		if i == m.agentCursor {
			cursor = styleCursor.Render("❯ ")
			name = styleSelected.Render(a.DisplayName())
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

func (m Model) spinnerGlyph() string {
	return spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
}

func (m Model) viewUpdatePlanning() string {
	return styleAccent.Render(fmt.Sprintf("  %s Computing update plan...", m.spinnerGlyph())) + "\n\n" +
		styleMuted.Render("  Scanning detected agents for stale or conflicting symlinks.")
}

func (m Model) viewUpdateConfirm() string {
	var b strings.Builder
	b.WriteString(styleAccent.Render("  Review update plan\n"))
	b.WriteString(styleMuted.Render("  Files marked replace will be backed up before being symlinked.\n"))

	rpt := m.updatePlan
	totalNoop, totalUpdated, totalReplaced, totalMissing := 0, 0, 0, 0

	for _, ar := range rpt.Agents {
		totalNoop += ar.Noop
		totalUpdated += ar.Updated
		totalReplaced += ar.Replaced
		totalMissing += ar.Missing

		header := styleSelected.Render(fmt.Sprintf("  %s", ar.Agent))
		rows := []string{
			header,
			styleMuted.Render(fmt.Sprintf("    ~  noop:     %d", ar.Noop)),
			styleSuccess.Render(fmt.Sprintf("    >  updated:  %d", ar.Updated)),
		}
		replacedLine := fmt.Sprintf("    !  replaced: %d", ar.Replaced)
		if ar.Replaced > 0 {
			rows = append(rows, styleWarning.Render(replacedLine+"  (will back up)"))
		} else {
			rows = append(rows, styleMuted.Render(replacedLine))
		}
		rows = append(rows, styleSuccess.Render(fmt.Sprintf("    +  missing:  %d", ar.Missing)))
		b.WriteString(styleAgentBox.Render(lipgloss.JoinVertical(lipgloss.Left, rows...)) + "\n")
	}

	total := fmt.Sprintf("  Total: noop=%d  updated=%d  replaced=%d  missing=%d",
		totalNoop, totalUpdated, totalReplaced, totalMissing)
	b.WriteString("\n" + styleAccent.Render(total) + "\n")

	b.WriteString("\n" + styleKey.Render("  [y] apply  •  [n] cancel  •  [esc] back"))
	return b.String()
}

func (m Model) viewUpdateNoop() string {
	var b strings.Builder
	b.WriteString(styleSuccess.Render("  ✓  Already up to date.\n"))
	b.WriteString(styleMuted.Render("  No symlinks needed to be refreshed or replaced.\n"))
	b.WriteString("\n" + styleKey.Render("  enter  back to menu"))
	return b.String()
}

func (m Model) viewUpdating() string {
	var b strings.Builder
	b.WriteString(styleAccent.Render(fmt.Sprintf("  %s Applying update...", m.spinnerGlyph())) + "\n\n")

	if m.updatePlan != nil {
		for _, ar := range m.updatePlan.Agents {
			work := ar.Updated + ar.Replaced + ar.Missing
			if work == 0 {
				b.WriteString(styleMuted.Render(fmt.Sprintf("  ✓  %s  (no changes)", ar.Agent)) + "\n")
				continue
			}
			b.WriteString(styleAccent.Render(fmt.Sprintf("  %s  %s", m.spinnerGlyph(), ar.Agent)) + "\n")
		}
	}
	b.WriteString("\n" + styleMuted.Render("  Reclassifying symlinks and backing up conflicts."))
	return b.String()
}

func (m Model) viewUpdateDone() string {
	var b strings.Builder
	if m.updateErr != nil {
		b.WriteString(styleError.Render("  Update failed: " + m.updateErr.Error() + "\n"))
		b.WriteString("\n" + styleKey.Render("  enter  back to menu  •  q  quit"))
		return b.String()
	}

	b.WriteString(styleSuccess.Render("  ✓  Update complete!\n"))

	rpt := m.updateReport
	if rpt == nil || len(rpt.Agents) == 0 {
		b.WriteString(styleMuted.Render("  No agents detected.\n"))
		b.WriteString("\n" + styleKey.Render("  enter  back to menu  •  q  quit"))
		return b.String()
	}

	for _, ar := range rpt.Agents {
		rows := []string{
			styleSelected.Render(fmt.Sprintf("  %s", ar.Agent)),
			styleMuted.Render(fmt.Sprintf("    ~  noop:     %d", ar.Noop)),
			styleSuccess.Render(fmt.Sprintf("    >  updated:  %d", ar.Updated)),
		}
		if ar.Replaced > 0 {
			rows = append(rows, styleWarning.Render(fmt.Sprintf("    !  replaced: %d", ar.Replaced)))
		} else {
			rows = append(rows, styleMuted.Render(fmt.Sprintf("    !  replaced: %d", ar.Replaced)))
		}
		rows = append(rows,
			styleSuccess.Render(fmt.Sprintf("    +  missing:  %d", ar.Missing)),
			styleError.Render(fmt.Sprintf("    ✗  failed:   %d", ar.Failed)),
		)
		b.WriteString(styleAgentBox.Render(lipgloss.JoinVertical(lipgloss.Left, rows...)) + "\n")
	}

	if rpt.BackupHits > 0 {
		b.WriteString("\n" + styleWarning.Render(fmt.Sprintf("  Backups saved to %s", rpt.BackupDir)) + "\n")
		stamp := backupStamp(rpt.BackupDir)
		if stamp != "" {
			b.WriteString(styleMuted.Render(fmt.Sprintf("  Restore with: duck-ai update --restore %s", stamp)) + "\n")
		}
	}
	b.WriteString("\n" + styleKey.Render("  enter  back to menu  •  q  quit"))
	return b.String()
}

func backupStamp(dir string) string {
	if dir == "" {
		return ""
	}
	idx := strings.LastIndex(dir, "/")
	if idx < 0 || idx == len(dir)-1 {
		return dir
	}
	return dir[idx+1:]
}
