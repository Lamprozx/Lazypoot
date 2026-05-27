package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"lazypoot/plugins"
	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

func RunImageInstallPipeline(ctx context.Context, cfg types.InstallConfig, ch chan tea.Msg) {
	defer func() {
		if r := recover(); r != nil {
			ch <- InstallErrorMsg{Err: fmt.Sprintf("internal panic: %v", r)}
		}
	}()

	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	arch := detectTermuxArch()

	sendLog(ch, types.LogStep, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	sendLog(ch, types.LogStep, " LazyPoot Image Installer")
	sendLog(ch, types.LogStep, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	sendLog(ch, types.LogInfo, fmt.Sprintf("Distro : %s (%s)", cfg.Distro, cfg.DistroID))
	sendLog(ch, types.LogInfo, fmt.Sprintf("Arch   : %s", arch))
	sendLog(ch, types.LogInfo, fmt.Sprintf("Variant: %s", cfg.KaliVariant))

	p, err := plugins.Get(cfg.DistroID)
	if err != nil {
		ch <- InstallErrorMsg{Err: "unsupported distro: " + cfg.DistroID}
		return
	}

	kp, ok := p.(plugins.ImageBasedDistro)
	if !ok {
		ch <- InstallErrorMsg{Err: "distro does not support image-based install"}
		return
	}
	variants, verr := kp.Variants(arch)
	if verr != nil || len(variants) == 0 {
		ch <- InstallErrorMsg{Err: "no variants available for " + arch}
		return
	}

	var variantInfo *plugins.ImageVariant
	for i := range variants {
		if variants[i].ID == cfg.KaliVariant {
			variantInfo = &variants[i]
			break
		}
	}
	if variantInfo == nil {
		variantInfo = &variants[0]
	}

	free, storErr := checkAvailableStorage(home)
	if storErr == nil {
		sendLog(ch, types.LogInfo, fmt.Sprintf("Free  : %.1f GB", float64(free)/(1024*1024*1024)))
		needed := variantInfo.Extracted
		sendLog(ch, types.LogInfo, fmt.Sprintf("Need  : %.1f GB", float64(needed)/(1024*1024*1024)))
		if free < needed {
			ch <- InstallErrorMsg{
				Err: fmt.Sprintf("Not enough storage: need %.1f GB, have %.1f GB",
					float64(needed)/(1024*1024*1024), float64(free)/(1024*1024*1024)),
			}
			return
		}
	} else {
		sendLog(ch, types.LogWarn, "Could not check storage: "+storErr.Error())
	}

	sendLog(ch, types.LogStep, "Checking proot...")
	if _, err := exec.LookPath("proot"); err != nil {
		sendLog(ch, types.LogWarn, "proot not found — installing via pkg...")
		if err := streamExec(ctx, ch, "pkg", "install", "-y", "proot"); err != nil {
			ch <- InstallErrorMsg{Err: "failed to install proot: " + err.Error()}
			return
		}
		sendLog(ch, types.LogOK, "proot installed")
	} else {
		sendLog(ch, types.LogOK, "proot already present")
	}

	sendPct(ch, 5)
	dd := distroDir(home, cfg.DistroID)
	os.MkdirAll(dd, 0755)

	metaP := metaPath(home, cfg.DistroID)
	if existing, rerr := ReadMeta(metaP); rerr == nil && existing.State == "ready" {
		ch <- InstallErrorMsg{Err: fmt.Sprintf("%s sudah terinstal. Hapus ~/.lazypoot/distros/%s untuk instal ulang", cfg.Distro, cfg.DistroID)}
		return
	}

	lock := lockPath(home, cfg.DistroID)
	if data, lerr := os.ReadFile(lock); lerr == nil {
		var pid int
		fmt.Sscanf(string(data), "%d", &pid)
		proc, _ := os.FindProcess(pid)
		if proc != nil && proc.Signal(syscall.Signal(0)) == nil {
			ch <- InstallErrorMsg{Err: "install lock exists — another install may be in progress"}
			return
		}
		os.Remove(lock)
	}
	os.WriteFile(lock, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644)
	defer os.Remove(lock)

	meta := ImageMeta{
		ID:        cfg.DistroID,
		Variant:   variantInfo.ID,
		Arch:      arch,
		Installed: 0,
		State:     "downloading",
	}
	if err := writeMeta(metaPath(home, cfg.DistroID), meta); err != nil {
		ch <- InstallErrorMsg{Err: "failed to write meta: " + err.Error()}
		return
	}

	partPath := imagePartPath(home, cfg.DistroID)
	urls := []string{variantInfo.ImageURL}
	urls = append(urls, variantInfo.Mirrors...)

	sendLog(ch, types.LogStep, fmt.Sprintf(" [1/4] Downloading %s...", variantInfo.Label))
	var downloadErr error
	for _, u := range urls {
		_, downloadErr = downloadWithResume(ctx, ch, u, partPath)
		if downloadErr == nil {
			break
		}
		sendLog(ch, types.LogWarn, fmt.Sprintf("mirror failed: %v", downloadErr))
	}
	if downloadErr != nil {
		os.Remove(partPath)
		ch <- InstallErrorMsg{Err: fmt.Sprintf("download failed: %v", downloadErr)}
		return
	}
	sendLog(ch, types.LogOK, "Download complete")
	sendPct(ch, 35)

	finalPath := imageFinalPath(home, cfg.DistroID)
	if err := os.Rename(partPath, finalPath); err != nil {
		ch <- InstallErrorMsg{Err: "rename: " + err.Error()}
		return
	}

	if variantInfo.SHAURL != "" {
		sendLog(ch, types.LogStep, " [2/4] Verifying checksum...")
		if err := verifyChecksum(finalPath, variantInfo.SHAURL); err != nil {
			os.Remove(finalPath)
			ch <- InstallErrorMsg{Err: "checksum mismatch: " + err.Error()}
			return
		}
		sendLog(ch, types.LogOK, "Checksum OK")
	} else {
		sendLog(ch, types.LogWarn, "No SHA URL — skipping verification")
	}
	sendPct(ch, 45)

	sendLog(ch, types.LogStep, " [3/4] Extracting rootfs...")
	meta.State = "extracting"
	if err := writeMeta(metaPath(home, cfg.DistroID), meta); err != nil {
		os.Remove(finalPath)
		ch <- InstallErrorMsg{Err: "failed to write meta: " + err.Error()}
		return
	}

	if err := os.RemoveAll(rootfsTmpPath(home, cfg.DistroID)); err != nil {
		os.Remove(finalPath)
		ch <- InstallErrorMsg{Err: "cleanup: " + err.Error()}
		return
	}
	os.MkdirAll(rootfsTmpPath(home, cfg.DistroID), 0755)

	if err := extractTarXZ(ch, finalPath, rootfsTmpPath(home, cfg.DistroID)); err != nil {
		os.RemoveAll(rootfsTmpPath(home, cfg.DistroID))
		os.Remove(finalPath)
		ch <- InstallErrorMsg{Err: "extraction failed: " + err.Error()}
		return
	}

	tmpPath := rootfsTmpPath(home, cfg.DistroID)
	if entries, rerr := os.ReadDir(tmpPath); rerr == nil && len(entries) == 1 && entries[0].IsDir() {
		prefixDir := filepath.Join(tmpPath, entries[0].Name())
		if pe, perr := os.ReadDir(prefixDir); perr == nil {
			for _, e := range pe {
				oldPath := filepath.Join(prefixDir, e.Name())
				newPath := filepath.Join(tmpPath, e.Name())
				_ = os.Rename(oldPath, newPath)
			}
			os.Remove(prefixDir)
		}
	}

	sendPct(ch, 65)

	if err := os.Rename(rootfsTmpPath(home, cfg.DistroID), rootfsFinalPath(home, cfg.DistroID)); err != nil {
		os.RemoveAll(rootfsTmpPath(home, cfg.DistroID))
		ch <- InstallErrorMsg{Err: "rename rootfs: " + err.Error()}
		return
	}
	os.Remove(finalPath)

	rootfs := rootfsFinalPath(home, cfg.DistroID)
	os.MkdirAll(filepath.Join(rootfs, "tmp"), 01777)

	sendLog(ch, types.LogOK, "Rootfs extracted")
	sendPct(ch, 70)

	sendLog(ch, types.LogStep, " [4/4] Configuring system...")
	cmdCtx, cmdCancel := context.WithTimeout(ctx, 120*time.Minute)
	defer cmdCancel()

	if cfg.DistroID == "kali" {
		_ = os.WriteFile(filepath.Join(rootfs, "etc", "resolv.conf"),
			[]byte("nameserver 1.1.1.1\nnameserver 8.8.8.8\n"), 0644)
		sendLog(ch, types.LogInfo, "DNS configured")
	}

	var cmds []string
	cmds = append(cmds, p.UpdateSystem()...)
	if cfg.DE != "cli" {
		cmds = append(cmds, p.InstallDesktop(cfg.DE)...)
		cmds = append(cmds, p.InstallDisplay(cfg.Display)...)
	}
	cmds = append(cmds, p.InstallPackages(cfg.Packages)...)
	cmds = append(cmds, p.SetupUser(cfg.Hostname)...)
	if cfg.DE != "cli" && cfg.InstallVirGL {
		cmds = append(cmds, p.SetupOpenGL()...)
	}
	cmds = append(cmds, p.PostInstall()...)
	cmds = append(cmds, p.InstallPackages([]string{"expect"})...)
	cmds = append(cmds, buildProfileCommands(cfg)...)
	cmds = append(cmds, buildGreetingCmds(cfg.DistroID)...)
	if cfg.Display == "vnc" && cfg.VNCPassword != "" {
		cmds = append(cmds, buildVNCPasswordCmds(cfg.VNCPassword)...)
	}

	var filtered []string
	for _, c := range cmds {
		if strings.Contains(c, "PS1=") {
			continue
		}
		filtered = append(filtered, c)
	}
	cmds = filtered

	for i, cmd := range cmds {
		label := truncate(cmd, 60)
		sendLog(ch, types.LogInfo, fmt.Sprintf("[%2d/%d] $ %s", i+1, len(cmds), label))
		if err := kaliStreamExec(cmdCtx, ch, rootfs, cmd); err != nil {
			sendLog(ch, types.LogWarn, fmt.Sprintf(" ⚠ %v", err))
		}
	}
	sendLog(ch, types.LogOK, "System configured")
	sendPct(ch, 95)

	sendLog(ch, types.LogStep, "Generating launcher and portal hooks...")
	if err := generateImageLauncher(cfg.DistroID, rootfs, home); err != nil {
		sendLog(ch, types.LogWarn, "launcher: "+err.Error())
	} else {
		sendLog(ch, types.LogOK, fmt.Sprintf("Launcher  ~/.lazypoot/distros/%s/launcher.sh", cfg.DistroID))
	}

	var usernames []string
	for _, u := range cfg.Users {
		usernames = append(usernames, u.Username)
	}
	if err := SetupImagePortalLogin(home, cfg.DistroID, rootfs, usernames); err != nil {
		sendLog(ch, types.LogWarn, "portal: "+err.Error())
	} else {
		sendLog(ch, types.LogOK, fmt.Sprintf("Portal    ~/.lazypoot/portal/%s.sh", cfg.DistroID))
	}

	meta.State = "ready"
	meta.Installed = time.Now().Unix()
	if err := writeMeta(metaPath(home, cfg.DistroID), meta); err != nil {
		sendLog(ch, types.LogWarn, "failed to write meta: "+err.Error())
	}

	sendPct(ch, 100)
	sendLog(ch, types.LogOK, fmt.Sprintf("%s (%s) installed!", cfg.Distro, variantInfo.Label))
	sendLog(ch, types.LogOK, fmt.Sprintf(" Launch: %s", cfg.DistroID))

	ch <- ImageInstallDoneMsg{
		Distro:      cfg.Distro,
		DistroID:    cfg.DistroID,
		RootfsDir:   rootfs,
		VNCPassword: cfg.VNCPassword,
	}
}

var ansiRegexp = regexp.MustCompile(
	`\x1b\[[\d;]*[a-zA-Z]` +
		`|\x1b\][^\x1b]*\x1b\\` +
		`|\x1b[PX\\^_][^\x1b]*\x1b\\` +
		`|\x1b[A-OW-Z\\\[\]` + "`" + `]`,
)

func sanitizeLine(s string) string {
	s = ansiRegexp.ReplaceAllString(s, "")
	s = strings.TrimRight(s, "\r")
	parts := strings.Split(s, "\r")
	s = parts[len(parts)-1]
	s = strings.Map(func(r rune) rune {
		if r == '\t' {
			return r
		}
		if r == '\n' || r == '\ufffd' {
			return -1
		}
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

func kaliStreamExec(ctx context.Context, ch chan<- tea.Msg, rootfs, cmd string) error {
	cmdCtx, cmdCancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cmdCancel()

	args := []string{
		"--link2symlink", "-0",
		"-r", rootfs,
		"-b", "/dev",
		"-b", "/proc",
		"-b", "/sdcard",
		"-b", "/system",
		"-b", "/vendor",
		"-b", "/apex",
		"-b", os.Getenv("HOME"),
		"-w", "/root",
		"/usr/bin/env", "-i",
		"HOME=/root",
		"PATH=/usr/local/sbin:/usr/local/bin:/bin:/usr/bin:/sbin:/usr/sbin",
		"TERM=" + os.Getenv("TERM"),
		"LANG=C.UTF-8",
		"/bin/bash", "-lc", cmd,
	}

	execCmd := exec.CommandContext(cmdCtx, "proot", args...)
	execCmd.Env = append(os.Environ(), "LD_PRELOAD=")

	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := execCmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	fanIn := func(r io.Reader, lvl types.LogLevel) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := sanitizeLine(scanner.Text())
			if line == "" {
				continue
			}
			ch <- InstallProgressMsg{Line: line, Level: lvl}
		}
	}

	wg.Add(2)
	go fanIn(stdout, types.LogInfo)
	go fanIn(stderr, types.LogWarn)
	wg.Wait()

	return execCmd.Wait()
}

func generateImageLauncher(distroID, rootfs, home string) error {
	lPath := launcherPath(home, distroID)
	os.MkdirAll(filepath.Dir(lPath), 0755)

	content := fmt.Sprintf(`#!/data/data/com.termux/files/usr/bin/bash
unset LD_PRELOAD
export TMPDIR=/tmp
cd "$HOME"

rootfs=%s

if [ ! -d "$rootfs" ]; then
	echo "LazyPoot: %s rootfs not found at $rootfs"
	exit 1
fi

user="root"
home_dir="/root"
start="/bin/bash --login"

if [ "$#" -gt 0 ] && [ "$1" != "-r" ] && [ "$1" != "-R" ]; then
	user="$1"
	home_dir="/home/$user"
	start="sudo -u $user /bin/bash --login"
	shift
fi

cmdline="proot \\
	--link2symlink -0 \\
	-r $rootfs \\
	-b /dev \\
	-b /proc \\
	-b /sdcard \\
	-b /system \\
	-b /vendor \\
	-b /apex \\
	-b $HOME \\
	-w $home_dir \\
	/usr/bin/env -i \\
	HOME=$home_dir \\
	PATH=/usr/local/sbin:/usr/local/bin:/bin:/usr/bin:/sbin:/usr/sbin \\
	TERM=$TERM \\
	LANG=C.UTF-8 \\
	$start"

if [ "$#" -eq 0 ]; then
	eval exec "$cmdline"
else
	eval exec "$cmdline -c \"$*\""
fi
`, rootfs, distroID)

	return os.WriteFile(lPath, []byte(content), 0700)
}

func SetupImagePortalLogin(home, distroID, rootfs string, usernames []string) error {
	portalDir := filepath.Join(home, ".lazypoot", "portal")
	os.MkdirAll(portalDir, 0755)
	pPath := filepath.Join(portalDir, distroID+".sh")

	userList := ""
	for _, u := range usernames {
		userList += fmt.Sprintf("    \"%s\"\n", u)
	}
	if userList == "" {
		userList = "    \"root\"\n"
	}

	content := fmt.Sprintf(`#!/data/data/com.termux/files/usr/bin/bash
unset LD_PRELOAD
export TMPDIR=/tmp
ROOTFS=%s

if [ ! -d "$ROOTFS" ]; then
	echo "LazyPoot: rootfs not found at $ROOTFS"
	exit 1
fi

login_as_root() {
	exec proot --link2symlink -0 \
		-r "$ROOTFS" \
		-b /dev -b /proc -b /sdcard \
		-b /system -b /vendor -b /apex \
		-b "$HOME" \
		-w /root \
		/usr/bin/env -i \
		HOME=/root \
		PATH=/usr/local/sbin:/usr/local/bin:/bin:/usr/bin:/sbin:/usr/sbin \
		TERM="$TERM" LANG=C.UTF-8 \
		/bin/bash --login
}

if [ -n "$1" ]; then
	exec proot --link2symlink -0 \
		-r "$ROOTFS" \
		-b /dev -b /proc -b /sdcard \
		-b /system -b /vendor -b /apex \
		-b "$HOME" \
		-w "/home/$1" \
		/usr/bin/env -i \
		HOME=/home/$1 \
		PATH=/usr/local/sbin:/usr/local/bin:/bin:/usr/bin:/sbin:/usr/sbin \
		TERM="$TERM" LANG=C.UTF-8 \
		sudo -u "$1" /bin/bash --login
fi

USERS=(%s)

if [ "${#USERS[@]}" -eq 1 ]; then
	login_as_root
fi

echo "Select user:"
select user in "${USERS[@]}"; do
	case $user in
		root) login_as_root ;;
		*) exec proot --link2symlink -0 \
			-r "$ROOTFS" \
			-b /dev -b /proc -b /sdcard \
			-b /system -b /vendor -b /apex \
			-b "$HOME" \
			-w "/home/$user" \
			/usr/bin/env -i \
			HOME=/home/$user \
			PATH=/usr/local/sbin:/usr/local/bin:/bin:/usr/bin:/sbin:/usr/sbin \
			TERM="$TERM" LANG=C.UTF-8 \
			sudo -u "$user" /bin/bash --login ;;
	esac
done
`, rootfs, userList)

	return os.WriteFile(pPath, []byte(content), 0700)
}
