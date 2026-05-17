package screens

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	Version    = "1.0.0"
	GitCommit  = "unknown"
	BuildDate  = "unknown"
	BuildGoVer = "unknown"
	GitDirty   = "clean"
)

var banner = strings.Join([]string{
	`__                                 ____                __      `,
	`/\ \                               /\  _` + "`" + `\             /\ \__   `,
	`\ \ \         __     ____    __  __\ \ \L\ \___     ___\ \ ,_\  `,
	` \ \ \  __  /'__` + "`" + `\  /\_ ,` + "`" + `\ /\ \/\ \\ \ ,__/ __` + "`" + `\  / __` + "`" + `\ \ \/  `,
	`  \ \ \L\ \/\ \L\.\_\/_/  /_\ \ \_\ \\ \ \/\ \L\ \/\ \L\ \ \ \_ `,
	`   \ \____/\ \__/.\_\ /\____\\/` + "`" + `____ \\ \_\ \____/\ \____/\ \__\`,
	`    \/___/  \/__/\/_/ \/____/ ` + "`" + `/___/> \\/_/\/___/  \/___/  \/__/`,
	`                                 /\___/                         `,
	`                                 \/__/                          `,
}, "\n")

func safeShortCommit() string {
	c := strings.TrimSpace(GitCommit)
	if c == "" || c == "unknown" {
		return "unknown"
	}
	if len(c) > 7 {
		return c[:7]
	}
	return c
}

func VersionLine() string {
	commit := safeShortCommit()
	if strings.EqualFold(strings.TrimSpace(GitDirty), "dirty") {
		commit += " dirty"
	}
	return fmt.Sprintf("LazyPoot v%s (%s)", Version, commit)
}

func resolveGoVersion() string {
	gv := strings.TrimSpace(BuildGoVer)
	if gv == "" || gv == "unknown" {
		gv = runtime.Version()
	}
	return gv
}

func BuildInfo() string {
	return fmt.Sprintf("build: %s | %s/%s | go %s",
		strings.TrimSpace(BuildDate),
		runtime.GOOS, runtime.GOARCH,
		resolveGoVersion(),
	)
}

func BuildSessionID() string {
	h := sha256.Sum256([]byte(GitCommit + BuildDate + runtime.GOOS))
	return hex.EncodeToString(h[:])
}

type versionJSON struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Dirty     bool   `json:"dirty"`
	BuildTime string `json:"build_time"`
	Go        string `json:"go"`
	OS        string `json:"os"`
	Shell     string `json:"shell"`
}

func VersionJSON() string {
	v := versionJSON{
		Version:   Version,
		Commit:    safeShortCommit(),
		Dirty:     strings.EqualFold(strings.TrimSpace(GitDirty), "dirty"),
		BuildTime: strings.TrimSpace(BuildDate),
		Go:        resolveGoVersion(),
		OS:        runtime.GOOS + "/" + runtime.GOARCH,
		Shell:     RuntimeShell(),
	}
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

func RuntimeShell() string {
	ppid := os.Getppid()
	if ppid > 1 {
		data, err := os.ReadFile("/proc/" + strconv.Itoa(ppid) + "/comm")
		if err == nil {
			s := strings.TrimSpace(string(data))
			if s != "" {
				return s
			}
		}
	}

	if shell := strings.TrimSpace(os.Getenv("SHELL")); shell != "" {
		base := filepath.Base(shell)
		if base != "." && base != string(filepath.Separator) {
			return base
		}
	}

	return "unknown"
}

func FullVersion() string {
	var sb strings.Builder
	sb.WriteString(banner)
	sb.WriteString("\n\n")
	sb.WriteString(VersionLine())
	sb.WriteString("\n")
	sb.WriteString(BuildInfo())
	sb.WriteString("\n")
	sb.WriteString("shell:   ")
	sb.WriteString(RuntimeShell())
	sb.WriteString("\n\n")
	sb.WriteString("LazyPoot – proot-distro login manager for Termux\n")
	sb.WriteString("Copyright (C) 2026 Lamprozx, MIT License\n")
	sb.WriteString("https://github.com/lamprozx/lazypoot\n")
	return sb.String()
}
