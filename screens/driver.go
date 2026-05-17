package screens

import (
	"strings"

	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

func RenderDriverScreen(options []string, icons []string, cursor int, width, height int) string {
	header := ui.StyleTitle.Render("󰾲  GPU Driver / VirGL")
	divider := ui.StyleDim.Render(strings.Repeat("─", 44))

	infoBox := ui.StyleBoxBlue.Padding(0, 2).Render(ui.JoinRows(
		ui.StyleCyan.Render("  About VirGL:"),
		ui.StyleDim.Render("  VirGL provides GPU acceleration inside proot"),
		ui.StyleDim.Render("  Provides OpenGL acceleration for desktop applications"),
		ui.StyleDim.Render("  Optional feature (does not affect basic desktop functionality)"),
		"",
		ui.StyleYellow.Render("  What Will Installed:"),
		ui.StyleDim.Render("  • virglrenderer (Termux host GPU bridge)"),
		ui.StyleDim.Render("  • Mesa OpenGL drivers inside the distro"),
	))

	var rows strings.Builder
	for i, opt := range options {
		isActive := i == cursor
		pointer := "  "
		if isActive {
			pointer = ui.StyleCursor.Render("❯ ")
		}
		icon := icons[i] + " "
		var lbl string
		if isActive {
			color := ui.ColGreen
			if i == 1 {
				color = ui.ColRed
			}
			lbl = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColBg)).
				Background(lipgloss.Color(color)).
				Bold(true).Padding(0, 1).Render(opt)
			icon = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(icon)
		} else {
			lbl = ui.StyleNormal.Render(opt)
			icon = ui.StyleBlue.Render(icon)
		}
		rows.WriteString(pointer + icon + lbl + "\n")
	}

	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Enter", "select"),
		ui.KeyHint("Esc", "back"),
	)

	panel := ui.StyleBoxCyan.Padding(1, 3).
		Render(ui.JoinRows(
			header,
			divider,
			infoBox,
			"",
			strings.TrimRight(rows.String(), "\n"),
			"",
			footer,
		))

	return ui.Center(width, height, panel)
}
