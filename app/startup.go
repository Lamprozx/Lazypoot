package app

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

func checkFontCmd() tea.Cmd {
	return func() tea.Msg {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		dir := filepath.Join(home, ".termux")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return FontCheckMsg{NerdFontFound: false}
		}

		if ok := verifyFontHash(dir); ok {
			return FontCheckMsg{NerdFontFound: true}
		}

		ents, err := os.ReadDir(dir)
		if err == nil {
			for _, e := range ents {
				if strings.Contains(strings.ToLower(e.Name()), "nerd") {
					return FontCheckMsg{NerdFontFound: true}
				}
			}
		}

		return FontCheckMsg{NerdFontFound: false}
	}
}

func verifyFontHash(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "font.ttf"))
	if err != nil {
		return false
	}

	hashData, err := os.ReadFile(filepath.Join(dir, "font.ttf.sha256"))
	if err != nil {
		return false
	}

	expected := strings.TrimSpace(string(hashData))
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])

	return got == expected
}

func loadInstalledDistrosCmd() tea.Cmd {
	return func() tea.Msg {
		fs := fsScanDistros()
		pd := pdListDistros()
		img := imageScanDistros()

		seen := map[string]bool{}
		var merged []types.InstalledDistro
		for _, d := range fs {
			if !seen[d.ID] {
				seen[d.ID] = true
				merged = append(merged, d)
			}
		}
		for _, d := range pd {
			if !seen[d.ID] {
				seen[d.ID] = true
				merged = append(merged, d)
			}
		}
		for _, d := range img {
			if !seen[d.ID] {
				seen[d.ID] = true
				merged = append(merged, d)
			}
		}

		return InstalledDistrosLoadedMsg{Distros: merged}
	}
}

func imageScanDistros() []types.InstalledDistro {
	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	ids := ScanImageDistros(home)
	if len(ids) == 0 {
		return nil
	}
	lookup := buildDistroLookup()
	var out []types.InstalledDistro
	for _, id := range ids {
		name, icon := id, ""
		if d, ok := lookup[id]; ok {
			name, icon = d.Name, d.Icon
		}
		out = append(out, types.InstalledDistro{
			Name: name,
			ID:   id,
			Icon: icon,
		})
	}
	return out
}

func fsScanDistros() []types.InstalledDistro {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	root := filepath.Join(prefix, "var/lib/proot-distro/installed-rootfs")

	ents, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	lookup := buildDistroLookup()
	var out []types.InstalledDistro

	for _, e := range ents {
		if !e.IsDir() {
			continue
		}

		id := e.Name()
		name, icon := id, ""

		if d, ok := lookup[id]; ok {
			name, icon = d.Name, d.Icon
		}

		out = append(out, types.InstalledDistro{
			Name:	name,
			ID:	id,
			Icon:	icon,
		})
	}

	return out
}

func pdListDistros() []types.InstalledDistro {
	out, err := exec.Command("proot-distro", "list", "--verbose").Output()
	if err != nil {
		return nil
	}

	lookup := buildDistroLookup()

	var res []types.InstalledDistro
	var alias string

	sc := bufio.NewScanner(strings.NewReader(string(out)))

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if strings.HasPrefix(line, "Alias:") {
			alias = strings.TrimSpace(strings.TrimPrefix(line, "Alias:"))
			continue
		}

		if strings.Contains(line, "Installed: yes") && alias != "" {
			name, icon := alias, ""

			if d, ok := lookup[alias]; ok {
				name, icon = d.Name, d.Icon
			}

			res = append(res, types.InstalledDistro{
				Name:	name,
				ID:	alias,
				Icon:	icon,
			})
		}
	}

	return res
}

func buildDistroLookup() map[string]types.DistroItem {
	all := defaultDistroItems()
	m := make(map[string]types.DistroItem)

	for _, d := range all {
		m[d.ID] = d
	}

	return m
}

func generatePortalScript(home, distroID string, usernames []string) (string, error) {
	if !distroIDRe.MatchString(distroID) {
		return "", fmt.Errorf("invalid distroID: %s", distroID)
	}

	dir := filepath.Join(home, ".lazypoot", "portal")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	path := filepath.Join(dir, distroID+".sh")

	var userList strings.Builder
	userList.WriteString("users=(\"root\"")
	for _, u := range usernames {
		if types.SafeNameRe.MatchString(u) {
			userList.WriteString(fmt.Sprintf(" %q", u))
		}
	}
	userList.WriteString(")")

	script := fmt.Sprintf(`#!/usr/bin/env bash
# LazyPoot portal: %s
# Usage: %s [username]

DISTRO="%s"
PORTAL_PATH="%s"
PREFIX="${PREFIX:-/data/data/com.termux/files/usr}"

if [ ! -d "$PREFIX/var/lib/proot-distro/installed-rootfs/$DISTRO" ]; then
  echo "Error: distribution '$DISTRO' is not installed."
  rm -f "$PORTAL_PATH"
  echo "Portal script removed."
  exit 1
fi

if [ -n "$1" ]; then
  proot-distro login %s -- su - "$1"
  exit $?
fi

%s

while true; do
  echo
  echo "  LazyPoot Login Manager --"
  echo
  for i in "${!users[@]}"; do
    printf "    %%d. %%s\n" "$i" "${users[$i]}"
  done

  while true; do
    echo
    printf "  ==> Input Number of User (or press Enter/q to quit): "
    read -r choice

    if [ -z "$choice" ] || [ "$choice" = "q" ] || [ "$choice" = "Q" ]; then
      echo
      exit 0
    fi

    if ! [[ "$choice" =~ ^[0-9]+$ ]]; then
      echo "  [ERROR] '$choice' is not a valid number."
      continue
    fi

    if [ "$choice" -lt 0 ] || [ "$choice" -ge "${#users[@]}" ]; then
      echo "  [ERROR] '$choice' is not in the list."
      continue
    fi

    chosen="${users[$choice]}"
    break
  done

  if [ "$chosen" = "root" ]; then
    proot-distro login %s
  else
    proot-distro login %s -- su - "$chosen"
  fi

  echo
  echo "  Session ended. Returning to login manager..."
done
`, distroID, distroID, distroID, path, distroID, userList.String(), distroID, distroID)

	return path, os.WriteFile(path, []byte(script), 0755)
}

func generateStartupScript(cfg types.InstallConfig) error {
	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("home not set")
	}

	path := filepath.Join(home, ".lazypoot", "launchers", cfg.DistroID+".sh")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	script := fmt.Sprintf(`#!/bin/bash
echo "starting %s..."
proot-distro login '%s'
`, cfg.DistroID, cfg.DistroID)

	return os.WriteFile(path, []byte(script), 0755)
}


