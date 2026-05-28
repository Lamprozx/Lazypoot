package app

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"lazypoot/plugins"
	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "2.1.0"

func generateVNCPassword() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate VNC password: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func sendPct(ch chan<- tea.Msg, pct int) {
	ch <- InstallProgressPctMsg{Percent: pct}
}

func RunInstallPipeline(ctx context.Context, cfg types.InstallConfig, ch chan tea.Msg) {
	defer func() {
		if r := recover(); r != nil {
			ch <- InstallErrorMsg{Err: fmt.Sprintf("internal panic: %v", r)}
		}
	}()

	sendLog(ch, types.LogStep, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	sendLog(ch, types.LogStep, " LazyPoot Installer "+version)
	sendLog(ch, types.LogStep, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	sendLog(ch, types.LogInfo, fmt.Sprintf("Distro : %s (%s)", cfg.Distro, cfg.DistroID))
	sendLog(ch, types.LogInfo, fmt.Sprintf("Mode : %s", deModeName(cfg.DE)))
	if cfg.DE != "cli" {
		sendLog(ch, types.LogInfo, fmt.Sprintf("Display : %s", cfg.Display))
		sendLog(ch, types.LogInfo, fmt.Sprintf("Driver: %s", cfg.AccelMode))
		sendLog(ch, types.LogInfo, fmt.Sprintf("Audio : %v", cfg.InstallAudio))
	}
	sendLog(ch, types.LogInfo, fmt.Sprintf("Hostname : %s", cfg.Hostname))
	sendLog(ch, types.LogInfo, fmt.Sprintf("Users : %d additional", len(cfg.Users)))
	sendPct(ch, 0)

	sendLog(ch, types.LogStep, " [1/8] Checking proot-distro...")
	if _, err := exec.LookPath("proot-distro"); err != nil {
		sendLog(ch, types.LogWarn, "proot-distro not found — installing via pkg...")
		if err := streamExec(ctx, ch, "pkg", "install", "-y", "proot-distro"); err != nil {
			ch <- InstallErrorMsg{Err: "failed to install proot-distro: " + err.Error()}
			return
		}
		sendLog(ch, types.LogOK, "proot-distro installed")
	} else {
		sendLog(ch, types.LogOK, "proot-distro already present")
	}
	sendPct(ch, 12)

	sendLog(ch, types.LogStep, fmt.Sprintf(" [2/8] Installing %s base rootfs...", cfg.Distro))
	installErr := streamExec(ctx, ch, "proot-distro", "install", cfg.DistroID)
	verify := exec.CommandContext(ctx, "proot-distro", "login", cfg.DistroID, "--", "true")
	if verify.Run() != nil {
		detail := ""
		if installErr != nil {
			detail = installErr.Error()
		}
		ch <- InstallErrorMsg{Err: fmt.Sprintf("failed to install %s rootfs: %s", cfg.Distro, detail)}
		return
	}
	if installErr != nil {
		sendLog(ch, types.LogOK, fmt.Sprintf("%s rootfs already exists", cfg.Distro))
	} else {
		sendLog(ch, types.LogOK, fmt.Sprintf("%s rootfs installed", cfg.Distro))
	}

	plugin, err := plugins.Get(cfg.DistroID)
	if err != nil {
		ch <- InstallErrorMsg{Err: "unsupported distro: " + cfg.DistroID}
		return
	}
	sendLog(ch, types.LogOK, fmt.Sprintf("Plugin: %s (%s)", plugin.Name(), plugin.PackageManager()))
	sendPct(ch, 25)

	if cfg.Display == "x11" {
		sendLog(ch, types.LogStep, " [3/8] Setting up Termux X11...")
		if err := streamExec(ctx, ch, "pkg", "install", "-y", "x11-repo"); err != nil {
			sendLog(ch, types.LogWarn, "x11-repo install failed — continuing")
		}
		if err := streamExec(ctx, ch, "pkg", "install", "-y", "termux-x11-nightly"); err != nil {
			sendLog(ch, types.LogWarn, "termux-x11-nightly not available — install manually")
		} else {
			sendLog(ch, types.LogOK, "Termux X11 ready")
		}
	} else {
		sendLog(ch, types.LogInfo, " [3/8] Skipping Termux X11")
	}
	sendPct(ch, 37)

	sendLog(ch, types.LogStep, " [4/8] Building setup command list...")
	var cmds []string
	cmds = append(cmds, plugin.UpdateSystem()...)
	if cfg.DE != "cli" {
		cmds = append(cmds, plugin.InstallDesktop(cfg.DE)...)
		cmds = append(cmds, plugin.InstallDisplay(cfg.Display)...)
	}
	cmds = append(cmds, plugin.InstallPackages(cfg.Packages)...)
	cmds = append(cmds, plugin.SetupUser(cfg.Hostname)...)
	if cfg.DE != "cli" && cfg.AccelMode != "software" {
		cmds = append(cmds, plugin.SetupOpenGL()...)
	}
	cmds = append(cmds, plugin.PostInstall()...)
	cmds = append(cmds, plugin.InstallPackages([]string{"expect"})...)
	cmds = append(cmds, buildProfileCommands(cfg)...)
	cmds = append(cmds, buildGreetingCmds(cfg.DistroID)...)
	if cfg.Display == "vnc" && cfg.VNCPassword != "" {
		cmds = append(cmds, buildVNCPasswordCmds(cfg.VNCPassword)...)
	}
	sendLog(ch, types.LogInfo, fmt.Sprintf(" %d commands queued", len(cmds)))
	sendPct(ch, 50)

	sendLog(ch, types.LogStep, fmt.Sprintf(" [5/8] Executing %d commands in chroot...", len(cmds)))
	step5Start := 50
	step5Range := 12
	for i, cmd := range cmds {
		label := truncate(cmd, 60)
		switch {
		case strings.Contains(cmd, "sudoers"):
			label = "sudoers: configuring sudo access"
		case strings.Contains(cmd, "usermod -aG"):
			label = "groups: adding user to sudo/wheel"
		case strings.Contains(cmd, "chpasswd"):
			label = "account: setting password"
		case strings.Contains(cmd, "vncpasswd"):
			label = "vnc: setting VNC password"
		case strings.Contains(cmd, "lazypoot-greeting"):
			label = "greeting: installing welcome message"
		case strings.Contains(cmd, "isablegreeting"):
			label = "greeting: installing toggle script"
		case strings.Contains(cmd, "chmod +x") && (strings.Contains(cmd, "isablegreeting") || strings.Contains(cmd, "nablegreeting")):
			label = "greeting: making script executable"
		}
		sendLog(ch, types.LogInfo, fmt.Sprintf("[%2d/%d] $ %s", i+1, len(cmds), label))
		if err := streamExec(ctx, ch, "proot-distro", "login", cfg.DistroID, "--", "bash", "-lc", cmd); err != nil {
			sendLog(ch, types.LogWarn, fmt.Sprintf(" ⚠ Non-fatal: %s", err.Error()))
		}
		sendPct(ch, step5Start+(i+1)*step5Range/len(cmds))
	}
	sendLog(ch, types.LogOK, "Chroot setup complete")

	needsVirGL := cfg.DE != "cli" && cfg.AccelMode != "software"
	hasAudio := cfg.DE != "cli" && cfg.InstallAudio

	if needsVirGL || hasAudio {
		sendLog(ch, types.LogStep, " [6/8] Installing host packages...")
		if needsVirGL {
			switch cfg.AccelMode {
			case "virgl":
				sendLog(ch, types.LogInfo, "  VirGL: installing virglrenderer-android")
				if err := streamExec(ctx, ch, "pkg", "install", "-y", "virglrenderer-android"); err != nil {
					sendLog(ch, types.LogWarn, "virglrenderer-android install failed")
				}
			case "zink":
				sendLog(ch, types.LogInfo, "  Zink: installing mesa-zink + virglrenderer-mesa-zink + vulkan-loader")
				if err := streamExec(ctx, ch, "pkg", "install", "-y", "mesa-zink", "virglrenderer-mesa-zink", "vulkan-loader-android"); err != nil {
					sendLog(ch, types.LogWarn, "Zink packages install failed — HW accel may not work")
				}
			}
		}
		if hasAudio {
			sendLog(ch, types.LogInfo, "  Audio: installing pulseaudio")
			if err := streamExec(ctx, ch, "pkg", "install", "-y", "pulseaudio"); err != nil {
				sendLog(ch, types.LogWarn, "pulseaudio install failed — audio may not work")
			}
		}
		sendLog(ch, types.LogOK, "Host packages ready")
	} else {
		sendLog(ch, types.LogInfo, " [6/8] Skipping host packages")
	}
	sendPct(ch, 75)

	sendLog(ch, types.LogStep, " [7/8] Generating startup script, portal, and shell hooks...")
	if err := generateStartupScript(cfg); err != nil {
		sendLog(ch, types.LogWarn, "could not write startup script: "+err.Error())
	} else {
		sendLog(ch, types.LogOK, fmt.Sprintf("Launcher  ~/start-%s.sh", cfg.DistroID))
	}
	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	var usernames []string
	for _, u := range cfg.Users {
		usernames = append(usernames, u.Username)
	}
	if err := SetupPortalLogin(home, cfg.DistroID, usernames); err != nil {
		sendLog(ch, types.LogWarn, "could not write portal script: "+err.Error())
	} else {
		sendLog(ch, types.LogOK, fmt.Sprintf("Portal    ~/ .lazypoot/portal/%s.sh", cfg.DistroID))
		sendLog(ch, types.LogOK, fmt.Sprintf("Alias     %s", cfg.DistroID))
	}
	sendPct(ch, 87)

	sendLog(ch, types.LogStep, " [8/8] Finalizing installation environment...")
	sendLog(ch, types.LogInfo, "Registering shell hooks...")
	_ = streamExec(ctx, ch, "fish", "-c", "source ~/.config/fish/config.fish >/dev/null 2>&1")
	_ = streamExec(ctx, ch, "bash", "-c", "source ~/.bashrc >/dev/null 2>&1")
	_ = streamExec(ctx, ch, "zsh", "-c", "source ~/.zshrc >/dev/null 2>&1")

	sendLog(ch, types.LogInfo, "Syncing runtime environment...")
	sendLog(ch, types.LogInfo, "Preparing portal command...")
	sendLog(ch, types.LogInfo, "Entering final state...")
	sendPct(ch, 100)
	sendLog(ch, types.LogOK, "Environment fully initialized")
	sendLog(ch, types.LogWarn, "If alias not found, restart Termux")

	sendLog(ch, types.LogStep, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	sendLog(ch, types.LogOK, fmt.Sprintf(" %s installed & configured!", cfg.Distro))
	sendLog(ch, types.LogOK, fmt.Sprintf(" Launch command: %s", cfg.DistroID))
	sendLog(ch, types.LogStep, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	ch <- InstallFinishedMsg{
		Distro:      cfg.Distro,
		VNCPassword: cfg.VNCPassword,
	}
}

func buildVNCPasswordCmds(vncPass string) []string {
	q := shellQuote(vncPass)
	return []string{
		"mkdir -p ~/.vnc",
		fmt.Sprintf("printf '%%s\\n' %s | vncpasswd -f > ~/.vnc/passwd", q),
		"chmod 600 ~/.vnc/passwd",
	}
}

func alpineFamily(distroID string) bool {
	switch distroID {
	case "alpine", "adelie":
		return true
	default:
		return false
	}
}

func noBashFamily(distroID string) bool {
	switch distroID {
	case "alpine", "adelie", "chimera":
		return true
	default:
		return false
	}
}

func userRcFile(distroID string) string {
	if noBashFamily(distroID) {
		return ".profile"
	}
	return ".bashrc"
}

func rootRcFile(distroID string) string {
	if noBashFamily(distroID) {
		return "/root/.profile"
	}
	return "/root/.bashrc"
}

func buildUserAddCmd(distroID, username string) string {
	q := shellQuote(username)
	switch {
	case alpineFamily(distroID):
		return fmt.Sprintf("id -u %s >/dev/null 2>&1 || adduser -D -h /home/%s -s /bin/sh %s", q, username, q)
	case noBashFamily(distroID):
		return fmt.Sprintf("id -u %s >/dev/null 2>&1 || useradd -m -s /bin/sh %s", q, q)
	default:
		return fmt.Sprintf("id -u %s >/dev/null 2>&1 || useradd -m -s /bin/bash %s", q, q)
	}
}

func buildProfileCommands(cfg types.InstallConfig) []string {
	var cmds []string

	if len(cfg.Users) > 0 {
		cmds = append(cmds, ensureSudoCmd())
	}

	cmds = append(cmds, fmt.Sprintf(`echo 'PS1="root@%s:\w# "' >> %s`, cfg.Hostname, rootRcFile(cfg.DistroID)))
	rc := userRcFile(cfg.DistroID)

	for _, u := range cfg.Users {
		if !types.SafeNameRe.MatchString(u.Username) {
			continue
		}

		sn := u.Username
		sp := shellQuote(u.Password)

		cmds = append(cmds,
			buildUserAddCmd(cfg.DistroID, sn),
			fmt.Sprintf("printf '%%s:%%s\\n' %s %s | chpasswd", shellQuote(sn), sp),
			grantSudoWithPassword(sn),
			fmt.Sprintf(`echo 'PS1="%s@localhost:\w\$ "' >> /home/%s/%s`, sn, sn, rc),
		)

		if cfg.DE != "cli" {
			switch cfg.Display {
			case "vnc":
				cmds = append(cmds, buildVncUserCmds(cfg.VNCPassword, sn, cfg.DistroID)...)
			case "x11":
				cmds = append(cmds, buildX11UserCmds(cfg.DE, sn, cfg.DistroID)...)
			}
		}
	}

	return cmds
}

func ensureSudoCmd() string {
	return strings.Join([]string{
		`command -v sudo >/dev/null 2>&1 || (`,
		`  (command -v apt-get >/dev/null 2>&1 && DEBIAN_FRONTEND=noninteractive apt-get install -y sudo) ||`,
		`  (command -v pacman >/dev/null 2>&1 && pacman -S --noconfirm sudo) ||`,
		`  (command -v apk >/dev/null 2>&1 && apk add sudo) ||`,
		`  (command -v dnf >/dev/null 2>&1 && dnf install -y sudo) ||`,
		`  (command -v zypper >/dev/null 2>&1 && zypper install -y sudo) ||`,
		`  (command -v xbps-install >/dev/null 2>&1 && xbps-install -y sudo)`,
		`)`,
	}, "\n")
}

func grantSudoWithPassword(username string) string {
	qUser := shellQuote(username)
	entry := fmt.Sprintf("%s ALL=(ALL:ALL) ALL", username)

	return strings.Join([]string{
		fmt.Sprintf(`command -v usermod >/dev/null 2>&1 && getent group sudo >/dev/null 2>&1 && usermod -aG sudo %s || true`, qUser),
		fmt.Sprintf(`command -v usermod >/dev/null 2>&1 && getent group wheel >/dev/null 2>&1 && usermod -aG wheel %s || true`, qUser),
		fmt.Sprintf(`echo '%s' >> /etc/sudoers`, entry),
		fmt.Sprintf(`install -d -m 755 /etc/sudoers.d && printf '%%s\n' %s > /etc/sudoers.d/lazy-%s && chmod 440 /etc/sudoers.d/lazy-%s`, shellQuote(entry), username, username),
	}, "\n")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func streamExec(ctx context.Context, ch chan<- tea.Msg, name string, args ...string) error {
	cmdCtx, cmdCancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cmdCancel()
	cmd := exec.CommandContext(cmdCtx, name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup

	filterLine := func(line string) (string, bool) {
		line = strings.TrimSpace(line)
		if line == "" {
			return "", false
		}

		lower := strings.ToLower(line)

		if strings.Contains(lower, "can't set permissions to 0777") {
			return "", false
		}
		if strings.Contains(lower, "warning given when extracting") {
			return "", false
		}
		if strings.Contains(lower, "operation not permitted") && (strings.Contains(lower, "chmod") || strings.Contains(lower, "chown")) {
			return "", false
		}
		if strings.Contains(lower, "cannot change ownership") {
			return "", false
		}
		if strings.Contains(lower, "failed to set permissions") {
			return "", false
		}
		if strings.Contains(lower, "permission denied") && (strings.Contains(lower, "fakeroot") || strings.Contains(lower, "proot")) {
			return "", false
		}
		if strings.Contains(lower, "can't sanitize binding") {
			return "", false
		}

		if strings.HasPrefix(line, "[*]") {
			return "", false
		}
		if strings.HasPrefix(line, "Get:") || strings.HasPrefix(line, "Hit:") || strings.HasPrefix(line, "Ign:") {
			return "", false
		}
		if strings.Contains(lower, "no mirror or mirror group selected") {
			return "", false
		}
		if strings.Contains(lower, "testing the available mirrors") {
			return "", false
		}
		if strings.Contains(lower, "picking mirror:") {
			return "", false
		}
		if strings.HasPrefix(line, "Fetched ") {
			return "", false
		}

		return line, true
	}

	fanIn := func(r io.Reader, lvl types.LogLevel) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		for scanner.Scan() {
			raw := scanner.Text()

			line, ok := filterLine(raw)
			if !ok {
				continue
			}

			sendLvl := lvl
			lower := strings.ToLower(line)
			if strings.Contains(lower, "error") || strings.Contains(lower, "failed") || strings.Contains(lower, "fatal") {
				sendLvl = types.LogWarn
			}

			ch <- InstallProgressMsg{
				Line:  line,
				Level: sendLvl,
			}
		}
	}

	wg.Add(2)
	go fanIn(stdout, types.LogInfo)
	go fanIn(stderr, types.LogWarn)

	wg.Wait()
	return cmd.Wait()
}

func deModeName(de string) string {
	if de == "cli" {
		return "CLI Only"
	}
	return strings.ToUpper(de)
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func buildGreetingCmds(distroID string) []string {
	return []string{
		fmt.Sprintf(`cat > /etc/profile.d/lazypoot-greeting.sh << 'GREETINGEOF'
if [ ! -f "$HOME/.lazypoot-greeting-done" ] && [ ! -f "$HOME/.lazypoot-greeting-disabled" ]; then
    touch "$HOME/.lazypoot-greeting-done"

		CYAN='\033[1;38;2;120;220;255m'
		BLUE='\033[1;38;2;90;160;255m'
		GREEN='\033[1;38;2;120;255;160m'
		YELLOW='\033[1;38;2;255;210;120m'
		RED='\033[1;38;2;255;120;120m'
		RESET='\033[0m'
		DIM='\033[2m'
		BOLD='\033[1m'

		echo ""
		echo -e "➣${BOLD}\033[38;2;255;210;120m${RESET} Login 𝗦𝘂𝗰𝗰𝗲𝘀𝗳𝘂𝗹𝘆${RESET}"
		echo -e "${DIM}Welcome to ${RESET}${CYAN}${BOLD}%s${RESET}"
		echo ""
		echo -e "➣${BOLD}If you have any issues, report them at:${RESET}"
		echo -e "${BLUE}${DIM}https://github.com/Lamprozx/Lazypoot/issues${RESET}"
		echo ""
		echo -e "➣${RED}${BOLD}To disable this message next time:${RESET}"
		echo -e "${CYAN}${BOLD}disablegreeting${RESET}"
		echo -e "\033[97mPres Ctrl + d or type exit to Logout${RESET}"
		echo -e "${DIM}──────────────────────────${RESET}"
fi
GREETINGEOF`, distroID),

		`cat > /usr/local/bin/disablegreeting << 'DISABLEEOF'
#!/bin/sh
touch "$HOME/.lazypoot-greeting-disabled"
echo "Greeting disabled."
DISABLEEOF`,
		"chmod +x /usr/local/bin/disablegreeting",

		`cat > /usr/local/bin/enablegreeting << 'ENABLEEOF'
#!/bin/sh
rm -f "$HOME/.lazypoot-greeting-disabled"
echo "Greeting re-enabled."
ENABLEEOF`,
		"chmod +x /usr/local/bin/enablegreeting",
	}
}
