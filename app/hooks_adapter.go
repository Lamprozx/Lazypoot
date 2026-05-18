package app

import (
	"path/filepath"
)

type RCTarget struct {
	Path    string
	BlockFn func([]string) string
}

func RCTargets(home string) []RCTarget {
	return []RCTarget{
		{filepath.Join(home, ".bashrc"), GenerateSHBlock},
		{filepath.Join(home, ".zshrc"), GenerateSHBlock},
		{filepath.Join(home, ".config", "fish", "config.fish"), GenerateFishBlock},
	}
}
