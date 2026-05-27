package screens

import (
	"fmt"
	"strings"

	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

func RenderInstallSuccess(distro, distroID, de, display, vncPassword string, width, height int) string {
	banner := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColGreen)).Bold(true).
		Render(
			"  ╔═══════════════════════════════════════╗\n" +
				"  ║    ✔  INSTALLATION COMPLETE!          ║\n" +
				"  ╚═══════════════════════════════════════╝",
		)

	distroLine := ui.StyleCyan.Render(fmt.Sprintf("  ❱  %s", distro)) +
		ui.StyleGreen.Render("  is ready!")

	alias := ui.StyleKey.Render(distroID)

	var vncHint string
	if display == "vnc" && vncPassword != "" {
		vncHint = ui.StyleBoxPurple.Padding(0, 2).Render(ui.JoinRows(
			ui.StyleMagenta.Render("  [ 🔐 ] VNC Password: ")+
				lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColYellow)).Bold(true).Render(vncPassword),
			ui.StyleDim.Render("  Change with: vncpasswd inside cli mode"),
		))
	}

	guide := ui.StyleBoxBlue.Padding(1, 3).Render(ui.JoinRows(
		ui.StyleTitle.Render("  Quick Start"),
		ui.StyleDim.Render(strings.Repeat("─", 36)),
		"",
		ui.StyleGreen.Render("  Run: ")+alias,
		"",
		ui.StyleBlue.Render("  Shell reload:"),
		ui.StyleDim.Render("    bash → source ~/.bashrc"),
		ui.StyleDim.Render("    zsh  → source ~/.zshrc"),
		ui.StyleDim.Render("    fish → source ~/.config/fish/config.fish"),
		"",
		ui.StyleBlue.Render("  Launcher:"),
		ui.StyleDim.Render("    "+distroID),
		"",
		ui.StyleMagenta.Render("  DE: ")+ui.StyleCyan.Render(de),
		ui.StyleMagenta.Render("  Display: ")+ui.StyleCyan.Render(display),
	))

	footer := ui.Footer(ui.KeyHint("Enter / q", "back"))

	rows := []string{banner, "", distroLine, ""}
	if vncHint != "" {
		rows = append(rows, vncHint, "")
	}
	rows = append(rows, guide, "", footer)

	panel := ui.StyleBoxGreen.Padding(1, 3).Render(ui.JoinRows(rows...))
	return ui.Center(width, height, panel)
}
