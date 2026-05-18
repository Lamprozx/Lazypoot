package ui

import "strings"

func RenderMenuList(labels []string, icons []string, cursor int) string {
	var sb strings.Builder
	for i, label := range labels {
		isActive := i == cursor
		pointer := "  "
		if isActive {
			pointer = StyleCursor.Render("❯ ")
		}
		icon := "  "
		if i < len(icons) {
			icon = icons[i] + " "
		}
		if isActive {
			icon = StyleCyan.Render(icon)
			sb.WriteString(pointer + icon + StyleSelected.Render(" "+label+" ") + "\n")
		} else {
			icon = StyleBlue.Render(icon)
			sb.WriteString(pointer + icon + StyleNormal.Render(label) + "\n")
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}
