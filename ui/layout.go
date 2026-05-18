package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Center(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func Footer(hints ...string) string {
	return StyleDim.Render("  " + strings.Join(hints, "   ") + "  ")
}

func KeyHint(key, desc string) string {
	return StyleKey.Render("["+key+"]") + StyleDim.Render(" "+desc)
}

func JoinRows(rows ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func JoinCols(cols ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func StyleFg(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(ColFg)).Render(s)
}
