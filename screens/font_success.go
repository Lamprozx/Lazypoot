package screens

import (
	"strings"

	"lazypoot/ui"
)

func RenderFontSuccess(width, height int) string {
	header := ui.StyleTitle.Render("  Font Installed Successfully!")
	divider := ui.StyleDim.Render(strings.Repeat("─", 40))

	msg := ui.StyleNormal.Render("Now reopen Termux to apply the font.")

	footer := ui.Footer(
		ui.KeyHint("Esc", "back to Main Menu"),
	)
	footer = ui.StyleRed.Render(footer)

	panel := ui.StyleBoxGreen.Padding(1, 3).
		Render(ui.JoinRows(header, divider, "", msg, "", footer))

	return ui.Center(width, height, panel)
}
