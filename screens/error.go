package screens

import (
	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

func RenderErrorPopup(msg string, w, h int) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(msg)

	return ui.Center(w, h, box)
}

func RenderInstallError(msg string, w, h int) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(msg)

	return ui.Center(w, h, box)
}
