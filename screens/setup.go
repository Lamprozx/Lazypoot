package screens

import (
	"strings"

	"lazypoot/types"
	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

var setupLabels = []string{
	"GUI Selection",
	"Display Server",
	"GPU Driver",
	"Profile / Users",
	"Extra Packages",
	"Timezone",
	"  Install  ",
}
var setupIcons = []string{"󰍹", "󰹑", "󰾲", "󰟀", "󰏗", "󰌎", "󰁔"}

func RenderSetup(cfg types.InstallConfig, cursor, width, height int) string {
	header := ui.StyleTitle.Render("󰒋  Configure Install  ") +
		ui.StyleDim.Render("·  distro: ") +
		ui.StyleCyan.Render(cfg.Distro)

	var menuRows strings.Builder
	for i, label := range setupLabels {
		isActive := i == cursor
		isInstall := i == len(setupLabels)-1

		pointer := "  "
		if isActive {
			pointer = ui.StyleCursor.Render("❯ ")
		}

		icon := setupIcons[i] + " "
		switch {
		case isInstall && isActive:
			icon = ui.StyleGreen.Render("󰁔 ")
		case isInstall:
			icon = ui.StyleDim.Render("󰁔 ")
		case isActive:
			icon = ui.StyleCyan.Render(icon)
		default:
			icon = ui.StyleBlue.Render(icon)
		}

		var lbl string
		switch {
		case isInstall && isActive:
			lbl = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColBg)).
				Background(lipgloss.Color(ui.ColGreen)).
				Bold(true).Padding(0, 2).Render(label)
		case isInstall:
			lbl = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColGreen)).Bold(true).Render(label)
		case isActive:
			lbl = ui.StyleSelected.Render(label)
		default:
			lbl = ui.StyleNormal.Render(label)
		}

		menuRows.WriteString(pointer + icon + lbl + "\n")
	}

	leftPane := ui.StyleBox.Width(34).Padding(1, 2).
		Render(strings.TrimRight(menuRows.String(), "\n"))

	cartPane := ui.RenderCart(cfg, 32)
	cols := ui.JoinCols(leftPane, "  ", cartPane)

	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Enter", "select"),
		ui.KeyHint("Esc", "back"),
	)

	panel := ui.StyleBoxBlue.Padding(0, 2).
		Render(ui.JoinRows("", header, "", cols, "", footer))

	return ui.Center(width, height, panel)
}

func RenderSetupSubscreen(title string, options []string, icons []string, cursor int, width, height int) string {
	header := ui.StyleTitle.Render(title)
	divider := ui.StyleDim.Render(strings.Repeat("─", 36))
	list := ui.RenderMenuList(options, icons, cursor)
	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Enter", "select"),
		ui.KeyHint("Esc", "back"),
	)
	panel := ui.StyleBoxCyan.Padding(1, 3).
		Render(ui.JoinRows(header, divider, list, "", footer))
	return ui.Center(width, height, panel)
}

func RenderInputScreen(title, label, value string, mask bool, width, height int) string {
	header := ui.StyleTitle.Render(title)
	labelLine := ui.StyleBlue.Render("  " + label + ":")
	displayVal := value
	if mask {
		displayVal = strings.Repeat("•", len([]rune(value)))
	}
	inputBox := ui.StyleBoxBlue.Padding(0, 2).
		Render(ui.StyleFg(displayVal) + ui.StyleCyan.Render("▋"))
	footer := ui.Footer(
		ui.KeyHint("Enter", "confirm"),
		ui.KeyHint("Esc", "cancel"),
	)
	panel := ui.StyleBoxCyan.Padding(1, 4).
		Render(ui.JoinRows(header, "", labelLine, inputBox, "", footer))
	return ui.Center(width, height, panel)
}
