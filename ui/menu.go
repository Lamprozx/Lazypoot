package ui

import (
	"fmt"
	"strings"

	"lazypoot/types"

	"github.com/charmbracelet/lipgloss"
)

func RenderCart(cfg types.InstallConfig, width int) string {
	line := func(icon, key, val string) string {
		k := lipgloss.NewStyle().Foreground(lipgloss.Color(ColBlue)).Render(icon + " " + key + ":")
		var v string
		if val == "" {
			v = StyleDim.Render("  —")
		} else {
			v = StyleCyan.Render("  " + val)
		}
		return k + v
	}

	audioVal := "no"
	if cfg.InstallAudio {
		audioVal = "yes"
	}
	pkgs := strings.Join(cfg.Packages, " ")
	userCount := ""
	if len(cfg.Users) > 0 {
		userCount = fmt.Sprintf("%d user(s)", len(cfg.Users))
	}

	rows := []string{
		StyleMagenta.Render("  \uf1b2 Config Cart"),
		StyleDim.Render(strings.Repeat("─", width-4)),
		line("󰣇", "Distro", cfg.Distro),
		line("󰍹", "DE", cfg.DE),
		line("󰹑", "Display", cfg.Display),
		line("󰾲", "Driver", cfg.AccelMode),
		line("󰓃", "Audio", audioVal),
		line("󰟀", "Host", cfg.Hostname),
		line("󰯄", "Users", userCount),
		line("󰏗", "Pkgs", pkgs),
		line("󰌎", "TZ", cfg.Timezone),
	}

	return StyleBoxPurple.
		Width(width).
		Padding(1, 2).
		Render(strings.Join(rows, "\n"))
}
