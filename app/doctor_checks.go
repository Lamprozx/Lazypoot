package app

import (
	"os/exec"
	"strings"

	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

var checkTools = []struct {
	id	string
	icon	string
}{
	{"proot-distro", "󰒋"},
	{"bash", "\uf489"},
	{"wget", "󰖩"},
	{"curl", "󰛳"},
	{"tar", "󰋚"},
	{"pkg", "󰏗"},
}

func runDoctorChecksCmd() tea.Cmd {
	return func() tea.Msg {
		args := make([]string, len(checkTools))
		for i, t := range checkTools {
			args[i] = t.id
		}
		out, _ := exec.Command("whereis", args...).Output()
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		lineMap := make(map[string]string)
		for _, l := range lines {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) == 2 {
				lineMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
		results := make([]types.CheckResult, len(checkTools))
		for i, t := range checkTools {
			paths := lineMap[t.id]
			first := ""
			if paths != "" {
				fields := strings.Fields(paths)
				if len(fields) > 0 {
					first = fields[0]
				}
			}
			results[i] = types.CheckResult{
				Tool:	t.id,
				Icon:	t.icon,
				Found:	paths != "",
				Path:	first,
			}
		}
		return DoctorChecksMsg{Results: results}
	}
}

func spinnerFrames() []string {
	return []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
}

func spinnerFrame(tick int) string {
	frames := spinnerFrames()
	return frames[tick%len(frames)]
}
