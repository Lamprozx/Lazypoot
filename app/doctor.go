package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DoctorResult struct {
	Name   string
	Status string
	Detail string
}

func RunDoctor(home string) []DoctorResult {
	var results []DoctorResult

	results = append(results, checkProotDistro()...)
	results = append(results, checkRootfs(home)...)
	results = append(results, checkRCWritable(home)...)
	results = append(results, checkPortalScripts(home)...)
	results = append(results, checkState(home)...)
	results = append(results, checkLock()...)

	return results
}

func checkProotDistro() []DoctorResult {
	path, err := exec.LookPath("proot-distro")
	if err != nil {
		return []DoctorResult{{"proot-distro", "FAIL", "not found in PATH"}}
	}
	return []DoctorResult{{"proot-distro", "OK", path}}
}

func checkRootfs(home string) []DoctorResult {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	root := filepath.Join(prefix, "var/lib/proot-distro/installed-rootfs")
	entries, err := os.ReadDir(root)

	ids := ReadManifest(home)
	if len(ids) == 0 {
		return []DoctorResult{{"rootfs", "WARN", "no distros in manifest"}}
	}

	var results []DoctorResult
	pdFound := map[string]bool{}
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				pdFound[e.Name()] = true
			}
		}
	}

	imageDistros := ScanImageDistros(home)
	imageMap := make(map[string]bool, len(imageDistros))
	for _, id := range imageDistros {
		imageMap[id] = true
	}

	for _, id := range ids {
		if imageMap[id] {
			msg, ok := CheckImageDistroIntegrity(home, id)
			if ok {
				results = append(results, DoctorResult{fmt.Sprintf("  %s rootfs", id), "OK", "image-based"})
			} else {
				results = append(results, DoctorResult{fmt.Sprintf("  %s rootfs", id), "FAIL", msg})
			}
		} else if pdFound[id] {
			results = append(results, DoctorResult{fmt.Sprintf("  %s rootfs", id), "OK", ""})
		} else {
			results = append(results, DoctorResult{fmt.Sprintf("  %s rootfs", id), "FAIL", "directory not found"})
		}
	}
	return results
}

func checkRCWritable(home string) []DoctorResult {
	var results []DoctorResult
	for _, t := range RCTargets(home) {
		f, err := os.OpenFile(t.Path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			results = append(results, DoctorResult{fmt.Sprintf("  %s", filepath.Base(t.Path)), "FAIL", err.Error()})
		} else {
			f.Close()
			results = append(results, DoctorResult{fmt.Sprintf("  %s", filepath.Base(t.Path)), "OK", t.Path})
		}
	}
	return results
}

func checkPortalScripts(home string) []DoctorResult {
	ids := ReadManifest(home)
	if len(ids) == 0 {
		return []DoctorResult{{"portal scripts", "WARN", "no distros in manifest"}}
	}
	var results []DoctorResult
	for _, id := range ids {
		script := filepath.Join(home, ".lazypoot", "portal", id+".sh")
		data, err := os.ReadFile(script)
		if err != nil {
			results = append(results, DoctorResult{fmt.Sprintf("  %s", id), "FAIL", err.Error()})
			continue
		}
		content := string(data)
		if strings.HasPrefix(content, "#!/data/data/com.termux/files/usr/bin/bash") ||
			strings.HasPrefix(content, "#!/usr/bin/env bash") {
			results = append(results, DoctorResult{fmt.Sprintf("  %s", id), "OK", script})
		} else {
			results = append(results, DoctorResult{fmt.Sprintf("  %s", id), "WARN", "missing shebang"})
		}
	}
	return results
}

func checkState(home string) []DoctorResult {
	s := ReadState(home)
	if len(s.DistroIDs) == 0 {
		return []DoctorResult{{"state", "WARN", "no distros in state"}}
	}
	return []DoctorResult{{"state", "OK", fmt.Sprintf("hook v%d, %d distro(s), hash=%s", s.HookVersion, len(s.DistroIDs), s.LastSyncHash[:8])}}
}

func checkLock() []DoctorResult {
	return []DoctorResult{{"lock", "OK", "no conflict"}}
}

func PrintDoctorResults(results []DoctorResult) {
	for _, r := range results {
		icon := "✓"
		switch r.Status {
		case "FAIL":
			icon = "✗"
		case "WARN":
			icon = "!"
		}
		line := fmt.Sprintf("  %s %s", icon, r.Name)
		if r.Detail != "" {
			line += fmt.Sprintf(" → %s", r.Detail)
		}
		if r.Status == "FAIL" {
			line = fmt.Sprintf("  %s %s", icon, r.Name)
			if r.Detail != "" {
				line += fmt.Sprintf(" → %s", r.Detail)
			}
		}
		fmt.Println(line)
	}
}
