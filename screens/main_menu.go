package screens

import (
	"fmt"
	"strings"

	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

const asciiTitle = `██╗      █████╗ ███████╗██╗   ██╗██████╗  ██████╗  ██████╗ ████████╗
██║     ██╔══██╗╚══███╔╝╚██╗ ██╔╝██╔══██╗██╔═══██╗██╔═══██╗╚══██╔══╝
██║     ███████║  ███╔╝  ╚████╔╝ ██████╔╝██║   ██║██║   ██║   ██║   
██║     ██╔══██║ ███╔╝    ╚██╔╝  ██╔═══╝ ██║   ██║██║   ██║   ██║   
███████╗██║  ██║███████╗   ██║   ██║     ╚██████╔╝╚██████╔╝   ██║   
╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝      ╚═════╝  ╚═════╝    ╚═╝   `

type mainMenuItem struct {
	icon  string
	label string
	key   string
}

var mainMenuItems = []mainMenuItem{
	{"󰒋", "Install Distro", "i"},
	{"󰗑", "Delete Distro", "d"},
	{"󰍛", "Proot Doctor", "p"},
	{"󰈆", "Exit", "q"},
}

func RenderMainMenu(cursor, width, height int) string {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColBlue)).Bold(true)
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColCyan)).Bold(true)

	var styledTitle strings.Builder
	for _, line := range strings.Split(asciiTitle, "\n") {
		styledTitle.WriteString(titleStyle.Render(line) + "\n")
	}
	subtitle := subtitleStyle.Render("   ❱  proot distro manager  ❰")
	titleBlock := styledTitle.String() + "\n" + subtitle

	var menuRows strings.Builder
	for i, item := range mainMenuItems {
		isActive := i == cursor
		pointer := "   "
		if isActive {
			pointer = ui.StyleCursor.Render(" ❯ ")
		}
		iconStr := ui.StyleBlue.Render(item.icon + "  ")
		if isActive {
			iconStr = ui.StyleCyan.Render(item.icon + "  ")
		}
		label := ""
		if isActive {
			label = ui.StyleSelected.Render(fmt.Sprintf(" %-20s", item.label))
		} else {
			label = ui.StyleNormal.Render(fmt.Sprintf("%-20s", item.label))
		}
		key := ui.StyleKey.Render("[" + item.key + "]")
		menuRows.WriteString(pointer + iconStr + label + "  " + key + "\n")
	}

	menuBlock := strings.TrimRight(menuRows.String(), "\n")
	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Enter", "select"),
		ui.KeyHint("q", "quit"),
	)
	versionLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColComment)).
		Render("  \uf489  lazypoot v2.2.0  ·  proot-distro manager")

	panel := ui.StyleBoxBlue.
		Padding(1, 4).
		Render(ui.JoinRows(
			titleBlock,
			"",
			menuBlock,
			"",
			footer,
			"",
			versionLine,
		))

	return ui.Center(width, height, panel)
}
