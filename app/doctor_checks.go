package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

var checkTools = []struct {
	id	string
	icon	string
}{
	{"proot-distro", "󰒋"},
	{"proot", "󰒋"},
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

func ScanImageDistros(home string) []string {
	dd := filepath.Join(home, ".lazypoot", "distros")
	entries, err := os.ReadDir(dd)
	if err != nil {
		return nil
	}
	var installed []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mp := filepath.Join(dd, e.Name(), "meta.json")
		meta, err := ReadMeta(mp)
		if err != nil {
			continue
		}
		if meta.State == "ready" {
			installed = append(installed, e.Name())
		}
	}
	return installed
}

func CheckImageDistroIntegrity(home, distroID string) (string, bool) {
	rf := filepath.Join(home, ".lazypoot", "distros", distroID, "rootfs")
	if _, err := os.Stat(filepath.Join(rf, "etc", "passwd")); err != nil {
		return "rootfs missing or corrupt", false
	}
	if _, err := os.Stat(filepath.Join(rf, "bin", "sh")); err != nil {
		if _, err2 := os.Stat(filepath.Join(rf, "bin", "bash")); err2 != nil {
			return "no shell binary found", false
		}
	}
	mp := filepath.Join(home, ".lazypoot", "distros", distroID, "meta.json")
	if _, err := os.Stat(mp); err != nil {
		return "meta.json missing", false
	}
	return "", true
}
