package screens

import (
	"fmt"
	"strings"

	"lazypoot/types"
	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

var fontSteps = []string{
	"Downloading Nerd Font",
	"Extracting font from archive",
	"Installing to ~/.termux/font.ttf",
	"Reloading Termux",
}

func RenderFontInstalling(logs []types.LogEntry, step, width, height int) string {
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColOrange)).Bold(true).
		Render("  󰌎  Installing Nerd Font")

	divider := ui.StyleDim.Render(strings.Repeat("─", 52))

	var stepsBuilder strings.Builder
	for i, s := range fontSteps {
		num := i + 1
		var icon, lbl string
		switch {
		case num < step:
			icon = ui.StyleGreen.Render("  ")
			lbl = ui.StyleGreen.Render(s)
		case num == step:
			icon = ui.StyleCyan.Render("  ")
			lbl = ui.StyleCyan.Render(s)
		default:
			icon = ui.StyleDim.Render("  ")
			lbl = ui.StyleDim.Render(s)
		}
		stepsBuilder.WriteString(fmt.Sprintf("  %s %s\n", icon, lbl))
	}

	pct := 0
	if step > 0 {
		pct = (step * 100) / len(fontSteps)
		if pct > 100 {
			pct = 100
		}
	}
	progressBar := renderFontProgressBar(pct, 46)

	const viewH = 8
	start := 0
	if len(logs) > viewH {
		start = len(logs) - viewH
	}
	var logLines strings.Builder
	for _, e := range logs[start:] {
		logLines.WriteString(renderFontLogLine(e) + "\n")
	}

	logBox := ui.StyleBox.Padding(0, 2).Render(strings.TrimRight(logLines.String(), "\n"))

	footer := fmt.Sprintf(
		"%s\n%s",
		ui.StyleDim.Render("  Please wait..."),
		ui.Footer(ui.KeyHint("Esc", "close")),
	)

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ui.ColOrange)).
		Padding(1, 3).
		Render(ui.JoinRows(
			header,
			divider,
			stepsBuilder.String(),
			progressBar,
			"",
			logBox,
			"",
			footer,
		))

	return ui.Center(width, height, panel)
}

func renderFontProgressBar(pct, barWidth int) string {
	filled := (pct * barWidth) / 100
	if filled > barWidth {
		filled = barWidth
	}
	filledStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("░", barWidth-filled)
	label := fmt.Sprintf(" %3d%%", pct)
	bar := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColOrange)).Render(filledStr) +
		ui.StyleDim.Render(emptyStr)
	return "  " + bar + ui.StyleOrange.Render(label)
}

func renderFontLogLine(e types.LogEntry) string {
	style, ok := ui.LogStyles[string(e.Level)]
	if !ok {
		style = ui.StyleNormal
	}
	icons := map[types.LogLevel]string{
		types.LogInfo:	"  ", types.LogOK: "  ",
		types.LogWarn:	"  ", types.LogError: "  ", types.LogStep: "  ",
	}
	icon := icons[e.Level]
	badge := style.Bold(true).Render(fmt.Sprintf("[%-5s]", string(e.Level)))
	return icon + badge + "  " + style.Render(e.Message)
}
