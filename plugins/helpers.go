package plugins

import (
	"fmt"
	"strings"

	"lazypoot/types"
)

func joinPkgs(pkgs []string) string {
	return strings.Join(pkgs, " ")
}

func ValidatePackages(pkgs []string) error {
	for _, p := range pkgs {
		if !types.SafePkgRe.MatchString(p) {
			return fmt.Errorf("invalid package name: %q", p)
		}
	}
	return nil
}
