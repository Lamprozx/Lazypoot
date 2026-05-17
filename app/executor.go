package app

import (
	"context"

	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

func waitForLog(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func ExecuteInstall(ctx context.Context, m Model) (chan tea.Msg, tea.Cmd) {
	ch := make(chan tea.Msg, 256)

	go RunInstallPipeline(ctx, m.config, ch)

	return ch, func() tea.Msg { return InstallStartedMsg{} }
}

func sendLog(ch chan<- tea.Msg, level types.LogLevel, msg string) {
	ch <- InstallProgressMsg{Line: msg, Level: level}
}
