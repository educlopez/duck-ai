package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent  = lipgloss.Color("#7C3AED")
	colorSuccess = lipgloss.Color("#10B981")
	colorError   = lipgloss.Color("#EF4444")
	colorWarning = lipgloss.Color("#F59E0B")
	colorMuted   = lipgloss.Color("#6B7280")
	colorWhite   = lipgloss.Color("#F9FAFB")

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			MarginBottom(1)

	styleSubtitle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginBottom(1)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorSuccess)

	styleError = lipgloss.NewStyle().
			Foreground(colorError)

	styleWarning = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	styleMuted = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleAccent = lipgloss.NewStyle().
			Foreground(colorAccent)

	styleSelected = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleCursor = lipgloss.NewStyle().
			Foreground(colorAccent)

	styleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2).
			MarginTop(1)

	styleAgentBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 2).
			MarginTop(1)

	styleKey = lipgloss.NewStyle().
			Foreground(colorMuted)
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const banner = `
 ██████╗ ██╗   ██╗ ██████╗██╗  ██╗      █████╗ ██╗
 ██╔══██╗██║   ██║██╔════╝██║ ██╔╝     ██╔══██╗██║
 ██║  ██║██║   ██║██║     █████╔╝█████╗███████║██║
 ██║  ██║██║   ██║██║     ██╔═██╗╚════╝██╔══██║██║
 ██████╔╝╚██████╔╝╚██████╗██║  ██╗     ██║  ██║██║
 ╚═════╝  ╚═════╝  ╚═════╝╚═╝  ╚═╝     ╚═╝  ╚═╝╚═╝`
