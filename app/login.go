package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func imageDistroLoginArgs(rootfs string, args ...string) *exec.Cmd {
	base := []string{
		"--link2symlink", "-0",
		"-r", rootfs,
		"/usr/bin/env", "-i",
		"HOME=/root",
		"PATH=/usr/local/sbin:/usr/local/bin:/bin:/usr/bin:/sbin:/usr/sbin",
	}
	base = append(base, args...)
	return exec.Command("proot", base...)
}

func RunLogin(distroID string) error {
	if !distroIDRe.MatchString(distroID) {
		return fmt.Errorf("invalid distro: %s", distroID)
	}

	if isImageBasedDistro(distroID) {
		home := os.Getenv("HOME")
		if home == "" {
			return fmt.Errorf("HOME not set")
		}
		rootfs := filepath.Join(home, ".lazypoot", "distros", distroID, "rootfs")
		if _, err := os.Stat(filepath.Join(rootfs, "etc", "passwd")); err != nil {
			return fmt.Errorf("rootfs not found for %s", distroID)
		}

		out, err := imageDistroLoginArgs(rootfs,
			"awk", "-F:", "$3 >= 1000 && $3 != 65534 {print $1}", "/etc/passwd",
		).Output()
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
	if isImageBasedDistro(distroID) {
		home := os.Getenv("HOME")
		if home == "" {
			return fmt.Errorf("HOME not set")
		}
		launcher := filepath.Join(home, ".lazypoot", "distros", distroID, "launcher.sh")
		if _, err := os.Stat(launcher); err != nil {
			return fmt.Errorf("launcher not found for %s", distroID)
		}
		cmd := exec.Command(launcher, user)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command("proot-distro", "login", distroID, "--", "su", "-", user)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func loginAsRoot(distroID string) error {
	if isImageBasedDistro(distroID) {
		home := os.Getenv("HOME")
		if home == "" {
			return fmt.Errorf("HOME not set")
		}
		portal := filepath.Join(home, ".lazypoot", "portal", distroID+".sh")
		if _, err := os.Stat(portal); err != nil {
			return fmt.Errorf("portal script not found for %s — run 'lazypoot sync' first", distroID)
		}
		cmd := exec.Command(portal, "root")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command("proot-distro", "login", distroID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type loginModel struct {
	distro  string
	users   []string
	cursor  int
	chosen  string
}

func (m loginModel) Init() tea.Cmd   { return nil }

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
