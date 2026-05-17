package screens

import (
	"fmt"
	"strings"

	"lazypoot/types"
	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

const logViewHeight = 20

func RenderInstalling(logs []types.LogEntry, scroll, logTotal int, label string, spinnerChar string, width, height int, installPct int) string {
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColGreen)).
		Bold(true).
		Render(fmt.Sprintf("  %s  ", label))

	spinner := ui.StyleCyan.Render(spinnerChar)
	subhdr := ui.StyleDim.Render("  Streaming log — please wait   ctrl+c to cancel")

	divider := ui.StyleDim.Render(strings.Repeat("─", 62))

	start := 0
	if scroll > logViewHeight {
		start = scroll - logViewHeight
	}
	end := scroll
	if end > len(logs) {
		end = len(logs)
	}
	if start > end {
		start = end
	}

	var logLines strings.Builder
	for _, entry := range logs[start:end] {
		logLines.WriteString(renderLogEntry(entry) + "\n")
	}

	lines := end - start
	for i := lines; i < logViewHeight; i++ {
		logLines.WriteString("\n")
	}

	pct := installPct
	if pct <= 0 && len(logs) > 0 {
		total := logTotal
		if total <= 0 {
			total = 80
		}
		pct = (len(logs) * 100) / total
		if pct > 100 {
			pct = 100
		}
	}
	progressBar := renderProgressBar(pct, 54)
	countLabel := ui.StyleDim.Render(fmt.Sprintf("  %d log lines", len(logs)))

	panel := ui.StyleBoxGreen.
		Padding(1, 2).
		Render(ui.JoinRows(
			header+"  "+spinner,
			subhdr,
			divider,
			logLines.String(),
			divider,
			progressBar,
			countLabel,
		))

	return ui.Center(width, height, panel)
}

func renderLogEntry(entry types.LogEntry) string {
	icons := map[types.LogLevel]string{
		types.LogInfo:	"  ",
		types.LogOK:	"  ",
		types.LogWarn:	"  ",
		types.LogError:	"  ",
		types.LogStep:	"  ",
	}

	levelStyle, ok := ui.LogStyles[string(entry.Level)]
	if !ok {
		levelStyle = ui.StyleNormal
	}

	icon := icons[entry.Level]
	if icon == "" {
		icon = "  "
	}

	badge := levelStyle.
		Bold(true).
		Render(fmt.Sprintf("[%-5s]", string(entry.Level)))

	return icon + badge + "  " + levelStyle.Render(entry.Message)
}

func renderProgressBar(pct, width int) string {
	filled := (pct * width) / 100
	if filled > width {
		filled = width
	}
	filledStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("░", width-filled)
	label := fmt.Sprintf(" %3d%%", pct)
	colorBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColGreen)).
		Render(filledStr) +
		ui.StyleDim.Render(emptyStr)
	return "  " + colorBar + ui.StyleCyan.Render(label)
}
