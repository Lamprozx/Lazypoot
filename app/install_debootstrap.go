package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"lazypoot/plugins"
	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

func EnsureHostDependencies(ctx context.Context, ch chan<- tea.Msg) error {
	pkgs := []string{"debootstrap", "proot", "tar", "xz-utils"}
	var missing []string
	for _, pkg := range pkgs {
		switch pkg {
		case "proot":
			if _, err := exec.LookPath("proot"); err != nil {
				missing = append(missing, pkg)
			}
		case "debootstrap":
			if _, err := exec.LookPath("debootstrap"); err != nil {
				missing = append(missing, pkg)
			}
		default:
			if _, err := exec.LookPath(pkg); err != nil {
				missing = append(missing, pkg)
			}
		}
	}
	if len(missing) == 0 {
		sendLog(ch, types.LogOK, "Host dependencies already present")
		return nil
	}
	sendLog(ch, types.LogWarn, fmt.Sprintf("Installing missing dependencies: %s", strings.Join(missing, ", ")))
	if err := streamExec(ctx, ch, "pkg", append([]string{"install", "-y"}, missing...)...); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}
	sendLog(ch, types.LogOK, "Host dependencies installed")
	return nil
}

func RunDebootstrapPipeline(ctx context.Context, cfg types.InstallConfig, ch chan tea.Msg) {
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
	sendLog(ch, types.LogStep, " LazyPoot Bootstrap Installer")
	sendLog(ch, types.LogStep, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	sendLog(ch, types.LogInfo, fmt.Sprintf("Distro : %s (%s)", cfg.Distro, cfg.DistroID))
	sendLog(ch, types.LogInfo, fmt.Sprintf("Arch   : %s", arch))
	sendLog(ch, types.LogInfo, fmt.Sprintf("Variant: %s", cfg.ParrotVariant))

	p, err := plugins.Get(cfg.DistroID)
	if err != nil {
		ch <- InstallErrorMsg{Err: "unsupported distro: " + cfg.DistroID}
		return
	}
	bp, ok := p.(plugins.MutableRootfsDistro)
	if !ok {
		ch <- InstallErrorMsg{Err: "distro does not support bootstrap install"}
		return
	}
	variants, verr := bp.BootstrapVariants(arch)
	if verr != nil || len(variants) == 0 {
		ch <- InstallErrorMsg{Err: "no variants available for " + arch}
		return
	}

	var variantInfo *plugins.BootstrapVariant
	for i := range variants {
		if variants[i].ID == cfg.ParrotVariant {
			variantInfo = &variants[i]
			break
		}
	}
	if variantInfo == nil {
		variantInfo = &variants[0]
	}

	free, storErr := checkAvailableStorage(home)
	if storErr == nil {
		need := int64(variantInfo.EstimatedSizeGB) * 1024 * 1024 * 1024
		sendLog(ch, types.LogInfo, fmt.Sprintf("Free  : %.1f GB", float64(free)/(1024*1024*1024)))
		sendLog(ch, types.LogInfo, fmt.Sprintf("Need  : ~%d GB", variantInfo.EstimatedSizeGB))
		if free < need {
			ch <- InstallErrorMsg{
				Err: fmt.Sprintf("Not enough storage: need ~%d GB, have %.1f GB",
					variantInfo.EstimatedSizeGB, float64(free)/(1024*1024*1024)),
			}
			return
		}
	} else {
		sendLog(ch, types.LogWarn, "Could not check storage: "+storErr.Error())
	}

	if err := EnsureHostDependencies(ctx, ch); err != nil {
		ch <- InstallErrorMsg{Err: "dependencies: " + err.Error()}
		return
	}

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
		ID:      cfg.DistroID,
		Variant: variantInfo.ID,
		Arch:    arch,
		State:   "bootstrapping",
	}
	if err := writeMeta(metaP, meta); err != nil {
		ch <- InstallErrorMsg{Err: "failed to write meta: " + err.Error()}
		return
	}

	sendPct(ch, 5)
	rootfs := rootfsFinalPath(home, cfg.DistroID)

	sendLog(ch, types.LogStep, " [1/6] Bootstrapping Debian base...")
	sendLog(ch, types.LogInfo, "This may take 5-15 minutes depending on your device and network.")

	debootstrapArgs := []string{"-0", "debootstrap", "--arch=" + arch, "--foreign", "stable", rootfs, "http://deb.debian.org/debian"}
	if err := streamExec(ctx, ch, "proot", debootstrapArgs...); err != nil {
		os.RemoveAll(rootfs)
		ch <- InstallErrorMsg{Err: "debootstrap failed: " + err.Error()}
		return
	}
	sendPct(ch, 25)

	sendLog(ch, types.LogStep, " [2/6] Running debootstrap second-stage...")
	secondStageCtx, secondStageCancel := context.WithTimeout(ctx, 30*time.Minute)
	defer secondStageCancel()
	ssCmd := exec.CommandContext(secondStageCtx, "proot", "-0", "-r", rootfs, "-b", "/dev", "-b", "/proc", "-b", "/sys", "/debootstrap/debootstrap", "--second-stage")
	ssCmd.Env = append(os.Environ(), "LD_PRELOAD=", "DEBOOTSTRAP_DIR=/debootstrap")
	ssOut, err := ssCmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(rootfs)
		ch <- InstallErrorMsg{Err: "debootstrap second-stage failed: " + err.Error() + "\n" + string(ssOut)}
		return
	}
	sendLog(ch, types.LogInfo, strings.TrimSpace(string(ssOut)))
	sendPct(ch, 40)

	sendLog(ch, types.LogStep, " [3/6] Installing base packages (wget, gnupg)...")
	cmds := []string{
		"DEBIAN_FRONTEND=noninteractive apt-get update -y",
		"DEBIAN_FRONTEND=noninteractive apt-get install -y wget gnupg",
	}
	for _, cmd := range cmds {
		if err := debootstrapProotExec(ctx, ch, rootfs, cmd); err != nil {
			os.RemoveAll(rootfs)
			ch <- InstallErrorMsg{Err: "base package install failed: " + err.Error()}
			return
		}
	}
	sendPct(ch, 50)

	sendLog(ch, types.LogStep, " [4/6] Switching repositories to Parrot OS...")
	repoCmds := []string{
		fmt.Sprintf(`cat > /etc/apt/sources.list.d/parrot.list << 'EOF'
deb https://deb.parrot.sh/parrot echo main contrib non-free non-free-firmware
deb https://deb.parrot.sh/direct/parrot echo-security main contrib non-free non-free-firmware
deb https://deb.parrot.sh/parrot echo-backports main contrib non-free non-free-firmware
EOF`),
		"wget -q -O /tmp/parrot-archive-keyring.deb https://deb.parrot.sh/parrot/pool/main/p/parrot-archive-keyring/parrot-archive-keyring_2024.12_all.deb",
		"DEBIAN_FRONTEND=noninteractive apt-get install -y /tmp/parrot-archive-keyring.deb",
		"rm -f /tmp/parrot-archive-keyring.deb",
	}
	for _, cmd := range repoCmds {
		if err := debootstrapProotExec(ctx, ch, rootfs, cmd); err != nil {
			os.RemoveAll(rootfs)
			ch <- InstallErrorMsg{Err: "repo switch failed: " + err.Error()}
			return
		}
	}

	sendLog(ch, types.LogStep, " [5/6] Upgrading to Parrot OS (first pass)...")
	sendLog(ch, types.LogInfo, "This will convert Debian into Parrot OS — may take 10-20 minutes.")
	if err := debootstrapProotExec(ctx, ch, rootfs, "DEBIAN_FRONTEND=noninteractive apt-get update -y"); err != nil {
		os.RemoveAll(rootfs)
		ch <- InstallErrorMsg{Err: "apt update failed: " + err.Error()}
		return
	}
	if err := debootstrapProotExec(ctx, ch, rootfs, "DEBIAN_FRONTEND=noninteractive apt-get full-upgrade -y"); err != nil {
		os.RemoveAll(rootfs)
		ch <- InstallErrorMsg{Err: "first full-upgrade failed: " + err.Error()}
		return
	}
	sendPct(ch, 70)

	sendLog(ch, types.LogInfo, fmt.Sprintf("Installing variant packages: %s", strings.Join(variantInfo.PackageSets, ", ")))
	for _, pkgSet := range variantInfo.PackageSets {
		if err := debootstrapProotExec(ctx, ch, rootfs, fmt.Sprintf("apt-get install -y -o Dpkg::Options::='--force-confdef' -o Dpkg::Options::='--force-confold' %s", pkgSet)); err != nil {
			os.RemoveAll(rootfs)
			ch <- InstallErrorMsg{Err: fmt.Sprintf("failed to install %s: %s", pkgSet, err.Error())}
			return
		}
	}
	sendPct(ch, 85)

	sendLog(ch, types.LogInfo, "Second full-upgrade to ensure complete migration...")
	if err := debootstrapProotExec(ctx, ch, rootfs, "DEBIAN_FRONTEND=noninteractive apt-get full-upgrade -y"); err != nil {
		os.RemoveAll(rootfs)
		ch <- InstallErrorMsg{Err: "second full-upgrade failed: " + err.Error()}
		return
	}

	sendLog(ch, types.LogInfo, "Cleaning package cache...")
	if err := debootstrapProotExec(ctx, ch, rootfs, "apt-get clean"); err != nil {
		sendLog(ch, types.LogWarn, "apt clean warning: "+err.Error())
	}
	sendPct(ch, 92)

	os.MkdirAll(filepath.Join(rootfs, "tmp"), 01777)

	sendLog(ch, types.LogStep, " [6/6] Generating launcher and portal scripts...")
	if err := generateImageLauncher(cfg.DistroID, rootfs, home); err != nil {
		os.RemoveAll(rootfs)
		ch <- InstallErrorMsg{Err: "launcher generation failed: " + err.Error()}
		return
	}
	if err := SetupPortalLogin(home, cfg.DistroID, nil); err != nil {
		os.RemoveAll(rootfs)
		ch <- InstallErrorMsg{Err: "portal setup failed: " + err.Error()}
		return
	}
	sendPct(ch, 100)

	meta.State = "ready"
	writeMeta(metaP, meta)

	sendLog(ch, types.LogOK, fmt.Sprintf("%s (%s) installed successfully!", cfg.Distro, variantInfo.Label))
	ch <- ImageInstallDoneMsg{VNCPassword: cfg.VNCPassword}
}

func debootstrapProotExec(ctx context.Context, ch chan<- tea.Msg, rootfs, cmd string) error {
	cmdCtx, cmdCancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cmdCancel()

	args := []string{
		"-0",
		"-r", rootfs,
		"-b", "/dev",
		"-b", "/proc",
		"-b", "/sys",
		"-b", "/sdcard",
		"-b", os.Getenv("HOME"),
		"-w", "/root",
		"/usr/bin/env", "-i",
		"HOME=/root",
		"PATH=/usr/local/sbin:/usr/local/bin:/bin:/usr/bin:/sbin:/usr/sbin",
		"TERM=" + os.Getenv("TERM"),
		"LANG=C.UTF-8",
		"DEBIAN_FRONTEND=noninteractive",
		"DEBCONF_NONINTERACTIVE_SEEN=true",
		"/bin/bash", "-lc", cmd,
	}

	execCmd := exec.CommandContext(cmdCtx, "proot", args...)
	execCmd.Env = append(os.Environ(), "LD_PRELOAD=", "DEBOOTSTRAP_DIR=/debootstrap")

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

	scanOutput := func(r io.Reader, lvl types.LogLevel) {
		scanner := bufio.NewReader(r)
		for {
			line, err := scanner.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimRight(line, "\n\r")
			if line != "" {
				ch <- InstallProgressMsg{Line: line, Level: lvl}
			}
		}
	}

	go scanOutput(stdout, types.LogInfo)
	go scanOutput(stderr, types.LogWarn)

	return execCmd.Wait()
}
