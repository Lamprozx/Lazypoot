package screens

import (
	"fmt"
	"strings"

	"lazypoot/types"
	"lazypoot/ui"

	"github.com/charmbracelet/lipgloss"
)

func RenderConfirmInstall(cfg types.InstallConfig, width, height int) string {

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColGreen)).
		Bold(true).
		Render("  Ready to Install  ")

	divider := ui.StyleDim.Render(strings.Repeat("━", 44))

	row := func(icon, key, val, col string) string {
		k := lipgloss.NewStyle().
			Foreground(lipgloss.Color(col)).
			Bold(true).
			Width(16).
			Render(icon + "  " + key)

		v := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColFg)).
			Render(val)

		return k + v
	}

	pkgs := strings.Join(cfg.Packages, " ")
	if pkgs == "" {
		pkgs = "—"
	}

	summary := ui.JoinRows(
		row("󰣇", "Distro", cfg.Distro, ui.ColCyan),
		row("󰍹", "Desktop", cfg.DE, ui.ColMagenta),
		row("󰹑", "Display", cfg.Display, ui.ColBlue),
		row("󰟀", "Hostname", cfg.Hostname, ui.ColYellow),
		row("󰏗", "Packages", pkgs, ui.ColGreen),
		row("󰌎", "Timezone", cfg.Timezone, ui.ColOrange),
	)

	steps := []struct{ icon, step string }{
		{"\uf487", "Install base distro via proot-distro"},
		{"\uf013", "Update system packages"},
		{"\uf108", "Install desktop environment"},
		{"\uf879", "Configure display server"},
		{"\uf0c0", "Setup user & hostname"},
		{"\uf1b6", "Configure OpenGL / Mesa"},
		{"\uf48e", "Generate startup script + alias"},
	}

	var pipelineLines []string

	pipelineLines = append(pipelineLines,
		ui.StyleDim.Render("Pipeline preview:"),
	)

	for i, s := range steps {

		num := fmt.Sprintf("%2d.", i+1)

		numStyled := ui.StyleDim.Render(num)
		iconStyled := ui.StyleBlue.Render(fmt.Sprintf("%-2s", s.icon))
		textStyled := ui.StyleNormal.Render(s.step)

		line := fmt.Sprintf("%s %s %s",
			numStyled,
			iconStyled,
			textStyled,
		)

		pipelineLines = append(pipelineLines, "  "+line)
	}

	pipelineBlock := ui.StyleBoxBlue.
		Padding(0, 2).
		Render(ui.JoinRows(pipelineLines...))

	confirmBtn := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColBg)).
		Background(lipgloss.Color(ui.ColGreen)).
		Bold(true).
		Padding(0, 3).
		Render(" Enter  Confirm & Install ")

	cancelBtn := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColBg)).
		Background(lipgloss.Color(ui.ColRed)).
		Bold(true).
		Padding(0, 3).
		Render(" Esc  Cancel ")

	actions := confirmBtn + "    " + cancelBtn

	panel := ui.StyleBoxGreen.
		Padding(1, 3).
		Render(ui.JoinRows(
			title,
			divider,
			"",
			summary,
			"",
			pipelineBlock,
			"",
			actions,
		))

	return ui.Center(width, height, panel)
}
