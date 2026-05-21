package tui

import "github.com/charmbracelet/lipgloss"

// Rose Pine base palette + duck-ai yellow accents.
//
// The base/surface/text/subtext colors come straight from Rose Pine; the
// accent colors replace gentle-ai's mauve/lavender with the rubber-duck
// yellow ramp so the TUI feels on-brand for duck-ai.
var (
	ColorBase    = lipgloss.Color("#191724")
	ColorSurface = lipgloss.Color("#1f1d2e")
	ColorOverlay = lipgloss.Color("#6e6a86")
	ColorText    = lipgloss.Color("#e0def4")
	ColorSubtext = lipgloss.Color("#908caa")
	ColorRed     = lipgloss.Color("#eb6f92")
	ColorGreen   = lipgloss.Color("#9ccfd8")
	ColorYellow  = lipgloss.Color("#f1ca93")

	// Duck accents (replace mauve/lavender/peach from gentle-ai).
	ColorDuckHead   = lipgloss.Color("#FFC833") // bright yellow — headings / cursor
	ColorDuckBody   = lipgloss.Color("#F5A623") // amber — selection / brand
	ColorDuckBronze = lipgloss.Color("#B36F00") // bronze — accents
)

// Cursor is the prefix used for the currently focused item.
const Cursor = "▸ "

// Tagline returns the welcome screen tagline with the given version.
func Tagline(version string) string {
	if version == "" {
		return "duck-ai — multi-agent skill manager"
	}
	return "duck-ai " + version + " — multi-agent skill manager"
}

// Pre-built reusable styles (mirror gentle-ai's exported style API).
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorDuckHead).
			Bold(true)

	HeadingStyle = lipgloss.NewStyle().
			Foreground(ColorDuckBody).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	SubtextStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorDuckHead).
			Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	FrameStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorDuckHead).
			Padding(1, 2)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorOverlay).
			Padding(0, 1)
)

// Legacy lowercase aliases — kept so the non-welcome screens (agents,
// skills, install, update flow) continue to compile without churn. They
// reuse the same colors as the exported styles above.
var (
	colorAccent  = ColorDuckBody
	colorSuccess = ColorGreen
	colorError   = ColorRed
	colorWarning = ColorYellow
	colorMuted   = ColorSubtext
	colorTitle   = ColorDuckHead

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTitle).
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
