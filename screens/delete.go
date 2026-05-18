package screens

import (
	"fmt"
	"strings"

	"lazypoot/types"
	"lazypoot/ui"
)

func RenderNoDistro(width, height int) string {
	icon := ui.StyleYellow.Render("  ")
	title := ui.StyleTitle.Render("No Distros Installed")

	msg := ui.StyleDim.Render(
		"  You haven't installed any distro yet.\n" +
			"  Go back and choose  Install Distro  to get started.",
	)

	hint := ui.StyleBoxBlue.Padding(0, 2).Render(
		ui.StyleCyan.Render("  ") +
			ui.StyleNormal.Render(" Press ") +
			ui.StyleKey.Render("[i]") +
			ui.StyleNormal.Render(" from the main menu to install a distro"),
	)

	footer := ui.Footer(ui.KeyHint("Esc", "back"))

	panel := ui.StyleBoxRed.
		Padding(1, 4).
		Render(ui.JoinRows(
			icon+"  "+title,
			"",
			msg,
			"",
			hint,
			"",
			footer,
		))

	return ui.Center(width, height, panel)
}

func RenderDeleteList(distros []types.InstalledDistro, cursor int, width, height int) string {
	header := ui.StyleTitle.Render("  Delete Distro")
	warning := ui.StyleRed.Render("  ⚠  This action is irreversible!")
	divider := ui.StyleDim.Render(strings.Repeat("─", 48))

	var listRows strings.Builder
	for i, d := range distros {
		isActive := i == cursor

		pointer := "  "
		if isActive {
			pointer = ui.StyleCursor.Render("❯ ")
		}

		check := ui.StyleDim.Render("[ ] ")
		if d.Selected {
			check = ui.StyleRed.Render("[✗] ")
		}

		icon := ui.StyleBlue.Render(d.Icon + " ")
		if isActive {
			icon = ui.StyleCyan.Render(d.Icon + " ")
		}

		name := ""
		if isActive {
			name = ui.StyleSelected.Render(fmt.Sprintf("%-22s", d.Name))
		} else {
			name = ui.StyleNormal.Render(fmt.Sprintf("%-22s", d.Name))
		}

		id := ui.StyleDim.Render("‹" + d.ID + "›")

		listRows.WriteString(pointer + check + icon + name + "  " + id + "\n")
	}

	list := strings.TrimRight(listRows.String(), "\n")

	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Space", "toggle"),
		ui.KeyHint("Enter", "delete selected"),
		ui.KeyHint("Esc", "back"),
	)

	panel := ui.StyleBoxRed.
		Padding(1, 3).
		Render(ui.JoinRows(
			header,
			warning,
			divider,
			list,
			"",
			footer,
		))

	return ui.Center(width, height, panel)
}
