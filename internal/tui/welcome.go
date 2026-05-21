package tui

import "strings"

// welcomeOptions returns the duck-ai welcome menu options.
func welcomeOptions() []string {
	return []string{
		"Install skills & commands",
		"Update (refresh symlinks, backup conflicts)",
		"Doctor (verify symlink health per agent)",
		"Registry (list managed skills + versions)",
		"Quit",
	}
}

// renderOptions renders a list of menu options with the cursor prefix on the
// selected row and a 2-space indent on the others (matching the cursor width).
func renderOptions(options []string, cursor int) string {
	var out strings.Builder
	for idx, option := range options {
		if idx == cursor {
			out.WriteString(SelectedStyle.Render(Cursor+option) + "\n")
		} else {
			out.WriteString(UnselectedStyle.Render("  "+option) + "\n")
		}
	}
	return out.String()
}

// RenderWelcome renders the welcome screen wrapped in FrameStyle. Layout:
// logo → tagline → Menu heading → options → help footer.
func RenderWelcome(cursor int, version string) string {
	var b strings.Builder

	b.WriteString(RenderLogo())
	b.WriteString("\n\n")
	b.WriteString(SubtextStyle.Render(Tagline(version)))
	b.WriteString("\n\n")
	b.WriteString(HeadingStyle.Render("Menu"))
	b.WriteString("\n\n")
	b.WriteString(renderOptions(welcomeOptions(), cursor))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓ navigate • enter select • q quit"))

	return FrameStyle.Render(b.String())
}
