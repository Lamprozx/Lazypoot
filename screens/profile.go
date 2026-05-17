package screens

import (
	"fmt"
	"strings"

	"lazypoot/types"
	"lazypoot/ui"
)

func RenderProfileMenu(labels []string, icons []string, cursor int, cfg types.InstallConfig, width, height int) string {
	header := ui.StyleTitle.Render("󰟀  Profile & Users")
	divider := ui.StyleDim.Render(strings.Repeat("─", 40))

	summary := ui.StyleBoxBlue.Padding(0, 2).Render(
		ui.StyleBlue.Render("󰟀  Host: ") + ui.StyleCyan.Render(orDash(cfg.Hostname)) + "\n" +
			ui.StyleBlue.Render("󰀄  Users: ") + ui.StyleCyan.Render(fmt.Sprintf("%d configured", len(cfg.Users))),
	)

	list := ui.RenderMenuList(labels, icons, cursor)

	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Enter", "select"),
		ui.KeyHint("Esc", "back"),
	)

	panel := ui.StyleBoxCyan.Padding(1, 3).
		Render(ui.JoinRows(header, divider, summary, "", list, "", footer))

	return ui.Center(width, height, panel)
}

func RenderDeleteUserList(users []types.UserAccount, cursor int, width, height int) string {
	header := ui.StyleTitle.Render("󰀅  Remove User")
	divider := ui.StyleDim.Render(strings.Repeat("─", 36))

	var rows strings.Builder
	for i, u := range users {
		isActive := i == cursor
		pointer := "  "
		if isActive {
			pointer = ui.StyleCursor.Render("❯ ")
		}
		icon := ui.StyleBlue.Render("󰀄 ")
		if isActive {
			icon = ui.StyleCyan.Render("󰀄 ")
		}
		label := ""
		if isActive {
			label = ui.StyleSelected.Render(fmt.Sprintf("%-20s", u.Username))
		} else {
			label = ui.StyleNormal.Render(fmt.Sprintf("%-20s", u.Username))
		}
		rows.WriteString(pointer + icon + label + "\n")
	}

	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Enter", "delete"),
		ui.KeyHint("Esc", "back"),
	)

	panel := ui.StyleBoxRed.Padding(1, 3).
		Render(ui.JoinRows(header, divider, strings.TrimRight(rows.String(), "\n"), "", footer))

	return ui.Center(width, height, panel)
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
