package app

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

func loadSysInfoCmd() tea.Cmd {
	return func() tea.Msg {
		return SysInfoMsg{Data: types.SysInfoData{
			OS:		gatherOS(),
			Kernel:		gatherKernel(),
			Arch:		gatherArch(),
			CPU:		gatherCPU(),
			RAM:		gatherRAM(),
			Storage:	gatherStorage(),
			Shell:		gatherShell(),
			ProotVer:	gatherProotVer(),
		}}
	}
}

func gatherOS() string {
	if _, err := os.Stat("/system/build.prop"); err == nil {
		return "Android / Termux"
	}
	out, err := exec.Command("uname", "-o").Output()
	if err != nil {
		return "Linux / Termux"
	}
	return strings.TrimSpace(string(out))
}

func gatherKernel() string {
	out, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func gatherArch() string {
	out, err := exec.Command("uname", "-m").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func gatherCPU() string {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "unknown"
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Hardware") || strings.HasPrefix(line, "model name") || strings.HasPrefix(line, "Processor") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "ARM"
}

func gatherRAM() string {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return "unknown"
	}
	defer f.Close()
	data := map[string]int64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := strings.TrimSuffix(fields[0], ":")
			val, err := strconv.ParseInt(fields[1], 10, 64)
			if err == nil {
				data[key] = val
			}
		}
	}
	total := data["MemTotal"]
	avail := data["MemAvailable"]
	if total == 0 {
		return "unknown"
	}
	used := total - avail
	return fmt.Sprintf("%.1f GB total / %.1f GB used",
		float64(total)/1048576, float64(used)/1048576)
}

func gatherStorage() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	out, err := exec.Command("df", "-h", home).Output()
	if err != nil {
		return "unknown"
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return "unknown"
	}
	fields := strings.Fields(lines[1])
	if len(fields) >= 4 {
		return fmt.Sprintf("%s total / %s used / %s free", fields[1], fields[2], fields[3])
	}
	return "unknown"
}

func gatherShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "bash"
	}
	parts := strings.Split(shell, "/")
	return parts[len(parts)-1]
}

func gatherProotVer() string {
	out, err := exec.Command("proot-distro", "--version").Output()
	if err != nil {
		out, err = exec.Command("proot-distro", "version").Output()
		if err != nil {
			return "not found"
		}
	}
	line := strings.Split(strings.TrimSpace(string(out)), "\n")[0]
	return strings.TrimSpace(line)
}
