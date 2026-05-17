package screens

import (
	"fmt"
	"strings"

	"lazypoot/types"
	"lazypoot/ui"
)

func RenderSelectDistro(items []types.DistroItem, cursor, width, height int) string {
	const viewH = 18

	header := ui.StyleTitle.Render("󰒋  Select Distro to Install")

	divider := ui.StyleDim.Render(strings.Repeat("─", 52))

	start := 0
	if cursor >= viewH {
		start = cursor - viewH + 1
	}
	end := start + viewH
	if end > len(items) {
		end = len(items)
	}

	var listRows strings.Builder
	for i := start; i < end; i++ {
		item := items[i]
		isActive := i == cursor

		pointer := "  "
		if isActive {
			pointer = ui.StyleCursor.Render("❯ ")
		}

		check := ui.StyleDim.Render("[ ] ")
		if item.Selected {
			check = ui.StyleGreen.Render("[✓] ")
		}

		icon := ui.StyleBlue.Render(item.Icon + " ")
		if isActive {
			icon = ui.StyleCyan.Render(item.Icon + " ")
		}

		name := ""
		if isActive {
			name = ui.StyleSelected.Render(fmt.Sprintf("%-22s", item.Name))
		} else {
			name = ui.StyleNormal.Render(fmt.Sprintf("%-22s", item.Name))
		}

		tag := ui.StyleDim.Render("‹" + item.ID + "›")

		listRows.WriteString(pointer + check + icon + name + "  " + tag + "\n")
	}

	list := strings.TrimRight(listRows.String(), "\n")

	scrollHint := ""
	if len(items) > viewH {
		scrollHint = "\n" + ui.StyleDim.Render(fmt.Sprintf("  %d/%d  [scroll with j/k]", cursor+1, len(items)))
	}

	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Space", "toggle"),
		ui.KeyHint("Enter", "confirm"),
		ui.KeyHint("Esc", "back"),
	)

	panel := ui.StyleBoxCyan.
		Padding(1, 3).
		Render(ui.JoinRows(
			header,
			divider,
			list+scrollHint,
			"",
			footer,
		))

	return ui.Center(width, height, panel)
}
