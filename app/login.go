package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func RunLogin(distroID string) error {
	if !distroIDRe.MatchString(distroID) {
		return fmt.Errorf("invalid distro: %s", distroID)
	}

	out, err := exec.Command("proot-distro", "login", distroID, "--",
		"awk", "-F:", "$3 >= 1000 && $3 != 65534 && $1 !~ /^aid_/ {print $1}", "/etc/passwd").Output()
	if err != nil {
		return loginAsRoot(distroID)
	}
	users := strings.Fields(string(out))
	if len(users) == 0 {
		return loginAsRoot(distroID)
	}

	p := tea.NewProgram(loginModel{distro: distroID, users: users})
	m, err := p.Run()
	if err != nil {
		return err
	}
	lm := m.(loginModel)
	if lm.chosen == "" {
		return nil
	}

	return loginAsUser(distroID, lm.chosen)
}

func loginAsUser(distroID, user string) error {

	cmd := exec.Command("proot-distro", "login", distroID, "--", "su", "-", user)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func loginAsRoot(distroID string) error {
	cmd := exec.Command("proot-distro", "login", distroID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type loginModel struct {
	distro	string
	users	[]string
	cursor	int
	chosen	string
}

func (m loginModel) Init() tea.Cmd	{ return nil }

func (m loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.users)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			m.chosen = m.users[m.cursor]
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m loginModel) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n  Select user for %s:\n\n", m.distro))
	for i, u := range m.users {
		if i == m.cursor {
			b.WriteString(fmt.Sprintf("  > %s\n", u))
		} else {
			b.WriteString(fmt.Sprintf("    %s\n", u))
		}
	}
	b.WriteString("\n  [up/down] navigate  [enter] confirm  [esc] cancel\n")
	return b.String()
}
