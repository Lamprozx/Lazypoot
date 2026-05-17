package screens

import (
	"fmt"
	"strings"

	"lazypoot/types"
	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

var doctorMenuLabels = []string{
	"View Install Logs",
	"System Information",
	"Back to Main Menu",
}
var doctorMenuIcons = []string{"󰷐", "󰍛", "󰋚"}

func RenderDoctor(cursor int, loading bool, checks []types.CheckResult, spinner string, width, height int) string {
	header := ui.StyleTitle.Render("󰍛  Proot Doctor")
	divider := ui.StyleDim.Render(strings.Repeat("─", 38))

	var statusBlock string
	if loading {
		statusBlock = ui.StyleBox.Padding(0, 2).Align(lipgloss.Left).Render(
			ui.StyleCyan.Render(spinner + "  Scanning environment..."),
		)
	} else if len(checks) == 0 {
		statusBlock = ui.StyleBox.Padding(0, 2).Align(lipgloss.Left).Render(
			ui.StyleDim.Render("  Just A Simple Debugger"),
		)
	} else {
		var rows []string

		iconStyle := lipgloss.NewStyle().Width(4).Align(lipgloss.Left)
		nameStyle := lipgloss.NewStyle().Width(15).Align(lipgloss.Left)

		for _, c := range checks {
			toolCell := lipgloss.JoinHorizontal(
				lipgloss.Left,
				iconStyle.Render(c.Icon),
				nameStyle.Render(c.Tool),
			)

			dotCell := ui.StyleDim.Render(" ···· ")

			var statusStr string
			if c.Found {
				statusStr = ui.StyleGreen.Render("✔  " + doctorFormatPath(c.Path))
			} else {
				statusStr = ui.StyleRed.Render("✘  not found")
			}

			row := lipgloss.JoinHorizontal(
				lipgloss.Left,
				"  ",
				toolCell,
				dotCell,
				statusStr,
			)
			rows = append(rows, row)
		}

		rowsContent := lipgloss.JoinVertical(lipgloss.Left, rows...)

		statusBlock = ui.StyleBox.
			Padding(0, 2).
			Align(lipgloss.Left).
			Render(
				ui.StyleDim.Render("Environment checks:") + "\n" + rowsContent,
			)
	}

	list := ui.RenderMenuList(doctorMenuLabels, doctorMenuIcons, cursor)

	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Enter", "select"),
		ui.KeyHint("Esc", "back"),
	)

	content := ui.JoinRows(
		header,
		divider,
		statusBlock,
		"",
		list,
		"",
		footer,
	)

	panel := ui.StyleBoxBlue.Padding(1, 3).Render(content)

	return ui.Center(width, height, panel)
}

func RenderDoctorLogs(logs []string, width, height int) string {
	header := ui.StyleTitle.Render("󰷐  Install Logs")
	divider := ui.StyleDim.Render(strings.Repeat("─", 52))

	var body strings.Builder
	if len(logs) == 0 {
		body.WriteString(ui.StyleDim.Render("  No logs yet. Complete an install to see output here."))
	} else {
		const maxLines = 24
		start := 0
		if len(logs) > maxLines {
			start = len(logs) - maxLines
		}
		for _, l := range logs[start:] {
			if l == "" {
				continue
			}
			body.WriteString("  " + ui.StyleNormal.Render(l) + "\n")
		}
	}

	footer := ui.Footer(ui.KeyHint("Esc", "back"))

	panel := ui.StyleBoxBlue.Padding(1, 3).
		Render(ui.JoinRows(
			header,
			divider,
			body.String(),
			"",
			footer,
		))

	return ui.Center(width, height, panel)
}

func RenderDoctorSysInfo(data types.SysInfoData, loading bool, spinner string, width, height int) string {
	header := ui.StyleTitle.Render("󰍛  System Information")
	divider := ui.StyleDim.Render(strings.Repeat("─", 44))

	if loading {
		panel := ui.StyleBoxBlue.Padding(1, 3).Render(ui.JoinRows(
			header,
			divider,
			"",
			ui.StyleCyan.Render("  "+spinner+"  Gathering system information..."),
			"",
			ui.Footer(ui.KeyHint("Esc", "back")),
		))
		return ui.Center(width, height, panel)
	}

	sysRow := func(icon, key, val string) string {
		k := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColBlue)).
			Width(22).
			Render(icon + "  " + key + ":")
		v := ui.StyleCyan.Render(doctorTruncate(doctorOrFallback(val, "unknown"), 36))
		return "  " + k + v
	}

	info := []string{
		sysRow("󰻡", "OS", data.OS),
		sysRow("\uf17c", "Kernel", data.Kernel),
		sysRow("󰍛", "Arch", data.Arch),
		sysRow("󰍛", "CPU", data.CPU),
		sysRow("󰘚", "RAM", data.RAM),
		sysRow("󱡶", "Storage", data.Storage),
		sysRow("󰀀", "Shell", data.Shell),
		sysRow("󰰔", "proot-distro", data.ProotVer),
	}

	footer := ui.Footer(ui.KeyHint("Esc", "back"))

	allRows := append([]string{header, divider, ""}, append(info, "", footer)...)
	panel := ui.StyleBoxBlue.Padding(1, 3).Render(ui.JoinRows(allRows...))

	return ui.Center(width, height, panel)
}

func doctorFormatPath(path string) string {
	if path == "" {
		return "not found"
	}
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return fmt.Sprintf("…/%s", strings.Join(parts[len(parts)-2:], "/"))
	}
	return path
}

func doctorOrFallback(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

func doctorTruncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
