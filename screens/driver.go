package screens

import (
	"strings"

	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

func RenderDriverScreen(options []string, icons []string, cursor int, width, height int) string {
	header := ui.StyleTitle.Render("󰾲  GPU Driver")
	divider := ui.StyleDim.Render(strings.Repeat("─", 44))

	infoLines := []string{""}
	switch cursor {
	case 0:
		infoLines = []string{
			ui.StyleDim.Render("  Software rendering via CPU (LLVMpipe)"),
			ui.StyleDim.Render("  Works on all devices, no extra setup"),
		}
	case 1:
		infoLines = []string{
			ui.StyleDim.Render("  VirGL: virtual OpenGL acceleration"),
			ui.StyleDim.Render("  Install: virglrenderer-android (host)"),
			ui.StyleDim.Render("  Server: virgl_test_server_android"),
			ui.StyleDim.Render("  Proot:  GALLIUM_DRIVER=virpipe"),
		}
	case 2:
		infoLines = []string{
			ui.StyleDim.Render("  Zink: OpenGL → Vulkan translation"),
			ui.StyleDim.Render("  Install: mesa-zink + vulkan-loader (host)"),
			ui.StyleDim.Render("  Server: virgl_test_server (Zink mode)"),
			ui.StyleDim.Render("  Proot:  GALLIUM_DRIVER=virpipe"),
			"",
			ui.StyleYellow.Render("  Requires Vulkan support on your GPU."),
			ui.StyleYellow.Render("  Not all devices are compatible."),
		}
	}
	infoLines = append([]string{ui.StyleCyan.Render("  " + options[cursor] + ":")}, infoLines...)
	infoBox := ui.StyleBoxBlue.Padding(0, 2).Render(ui.JoinRows(infoLines...))

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
			if i == 2 {
				color = ui.ColYellow
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
