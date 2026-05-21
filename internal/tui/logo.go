package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// logoLines contains the braille-art duck logo for duck-ai.
var logoLines = []string{
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣤⣤⣄⠀⠀⠀⠀⠀⠀",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⣿⣿⡉⢋⣧⣤⣤⡄⠀⠀",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⠈⠻⣿⣿⡿⠋⠉⠁⠀⠀⠀",
	"⠀⠀⣠⣶⣶⣶⣾⣿⣿⣿⣿⣿⣿⣿⣶⣦⡀⠀⠀⠀⠀⠀",
	"⠀⠀⠈⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡆⠀⠀⠀⠀",
	"⠀⠀⠀⠈⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⠀⠀⠀",
	"⠀⠀⠀⠀⠀⠀⠈⠉⠉⠙⠛⠉⠉⠉⠀⠀⠀⠀⠀⠀⠀⠀",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
}

// duckGradient defines the top-to-bottom yellow-to-bronze gradient for the
// braille duck logo (head → body → water/shadow).
var duckGradient = []lipgloss.Color{
	lipgloss.Color("#FFE066"), // band 1 — head highlight
	lipgloss.Color("#FFC833"), // band 2 — head/bill yellow
	lipgloss.Color("#F5A623"), // band 3 — body amber
	lipgloss.Color("#D88A00"), // band 4 — body orange
	lipgloss.Color("#B36F00"), // band 5 — water/shadow bronze
}

// RenderLogo returns the braille duck logo with a top-to-bottom gradient.
func RenderLogo() string {
	total := len(logoLines)
	if total == 0 {
		return ""
	}

	bands := len(duckGradient)
	var b strings.Builder

	for i, line := range logoLines {
		bandIdx := (i * bands) / total
		if bandIdx >= bands {
			bandIdx = bands - 1
		}
		style := lipgloss.NewStyle().Foreground(duckGradient[bandIdx])
		b.WriteString(style.Render(line))
		if i < total-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
