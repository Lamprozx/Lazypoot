package screens

import (
	"strings"

	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

func renderButton(label string, active bool, fg, bg string) string {
	base := lipgloss.NewStyle().Padding(0, 3)

	if active {
		return base.
			Foreground(lipgloss.Color(bg)).
			Background(lipgloss.Color(fg)).
			Bold(true).
			Render("  " + label + "  ")
	}

	return base.
		Foreground(lipgloss.Color(fg)).
		Render(label)
}

func divider(width int) string {
	return ui.StyleDim.Render(strings.Repeat("─", width))
}

func RenderFontWarnInteractive(cursor, width, height int) string {

	banner := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColOrange)).
		Bold(true).
		Render("  󰌎  Nerd Font Not Detected")

	notice := ui.StyleYellow.Render(
		"  This tool uses Nerd Font icons throughout the UI.\n" +
			"  Without one, icons may appear as empty rectangles.",
	)

	infoRows := []string{
		ui.StyleCyan.Render("  Font       : ") + ui.StyleFg("FiraCode Nerd Font"),
		ui.StyleCyan.Render("  Source     : ") + ui.StyleDim.Render("github.com/ryanoasis/nerd-fonts/releases/download/v3.4.0/FiraCode.zip "),
		ui.StyleCyan.Render("  Size       : ") + ui.StyleDim.Render("~2 MB (download,extract and clean automatically)"),
		ui.StyleCyan.Render("  Install to : ") + ui.StyleDim.Render("~/.termux/font.ttf"),
		"",
		ui.StyleGreen.Render("  Auto-install via LazyPoot: press [Yes] below"),
	}

	info := ui.StyleBoxBlue.Padding(1, 2).Render(ui.JoinRows(infoRows...))

	noBtn := renderButton("Nope", cursor == 0, ui.ColRed, ui.ColBg)
	yesBtn := renderButton("Yes — Install Now", cursor == 1, ui.ColGreen, ui.ColBg)

	btnRow := noBtn + "    " + yesBtn

	footer := ui.StyleDim.Render(
		"  j/k or ←/→ to choose   Enter to confirm   Esc to skip",
	)

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ui.ColOrange)).
		Padding(1, 3).
		Render(ui.JoinRows(
			banner,
			divider(54),
			"",
			notice,
			"",
			info,
			"",
			btnRow,
			"",
			footer,
		))

	return ui.Center(width, height, panel)
}
