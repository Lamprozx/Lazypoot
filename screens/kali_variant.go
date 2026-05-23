package screens

import (
	"fmt"
	"strings"

	"lazypoot/plugins"
	"lazypoot/ui"
)

func RenderKaliVariant(variants []plugins.ImageVariant, cursor int, storage int64, arch string, width, height int) string {
	header := ui.StyleTitle.Render("ﴣ  Kali Linux — Select Variant")

	archLine := ui.StyleCyan.Render("  Arch: ") + ui.StyleBold.Render(arch)
	freeStr := "unknown"
	if storage > 0 {
		freeStr = fmt.Sprintf("%.1f GB", float64(storage)/(1024*1024*1024))
	}
	storageLine := ui.StyleBlue.Render("  Free: ") + ui.StyleBold.Render(freeStr)

	divider := ui.StyleDim.Render(strings.Repeat("─", 52))

	var listRows strings.Builder
	for i, v := range variants {
		isActive := i == cursor
		pointer := "  "
		if isActive {
			pointer = ui.StyleCursor.Render("❯ ")
		}

		icon := "󰋗 "
		if isActive {
			icon = ui.StyleCyan.Render(icon)
		} else {
			icon = ui.StyleBlue.Render(icon)
		}

		sizeStr := fmt.Sprintf("~%.1f GB", float64(v.Size)/(1024*1024*1024))
		if v.Size < 1024*1024*1024 {
			sizeStr = fmt.Sprintf("~%d MB", v.Size/(1024*1024))
		}
		extractStr := fmt.Sprintf("(%.0f MB extracted)", float64(v.Extracted)/(1024*1024))
		if v.Extracted >= 1024*1024*1024 {
			extractStr = fmt.Sprintf("(%.1f GB extracted)", float64(v.Extracted)/(1024*1024*1024))
		}

		var label string
		if isActive {
			label = ui.StyleSelected.Render(" " + v.Label + " ")
		} else {
			label = ui.StyleNormal.Render(v.Label)
		}

		sizeStyle := ui.StyleDim
		if v.Size > storage && storage > 0 {
			sizeStyle = ui.StyleRed
		}

		listRows.WriteString(fmt.Sprintf("%s%s %s  %s %s\n",
			pointer, icon, label,
			sizeStyle.Render(sizeStr),
			ui.StyleDim.Render(extractStr),
		))
	}

	warning := ""
	if storage > 0 {
		for _, v := range variants {
			if v.Extracted > storage {
				warning = ui.StyleRed.Render("  ⚠ Not enough storage for " + v.Label)
				break
			}
		}
	}

	list := strings.TrimRight(listRows.String(), "\n")
	if warning != "" {
		list += "\n\n" + warning
	}

	footer := ui.Footer(
		ui.KeyHint("↑↓ / jk", "navigate"),
		ui.KeyHint("Enter", "select & install"),
		ui.KeyHint("Esc", "back"),
	)

	panel := ui.StyleBoxCyan.
		Padding(1, 3).
		Render(ui.JoinRows(
			header,
			"",
			archLine,
			storageLine,
			"",
			divider,
			list,
			"",
			footer,
		))

	return ui.Center(width, height, panel)
}
