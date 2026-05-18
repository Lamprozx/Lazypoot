package main

import (
	"flag"
	"fmt"
	"os"

	"lazypoot/app"
	"lazypoot/screens"
	_ "lazypoot/plugins"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		runTUI()
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "login":
		runLogin(os.Args[2:])
	case "sync":
		runSync(os.Args[2:])
	case "doctor":
		runDoctor()
	case "debug-replay":
		runDebug()
	case "help", "--help", "-h":
		printUsage()
	case "version", "--version", "-v", "--v":
		printVersion(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "lazypoot: unknown command %q\n", cmd)
		fmt.Fprintf(os.Stderr, "Run 'lazypoot --help' for usage.\n")
		os.Exit(1)
	}
}

func runLogin(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: lazypoot login <distro>")
		os.Exit(1)
	}
	if err := app.RunLogin(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "lazypoot login: %v\n", err)
		os.Exit(1)
	}
}

func runSync(args []string) {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	dryRun := fs.Bool("dry-run", false, "preview hook changes")
	fs.BoolVar(dryRun, "n", false, "shorthand for --dry-run")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "lazypoot sync: unexpected arguments: %v\n", fs.Args())
		os.Exit(1)
	}

	home := resolveHome()
	if *dryRun {
		report, err := app.SyncDryRun(home)
		if err != nil {
			fmt.Fprintf(os.Stderr, "lazypoot sync: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(report)
	} else {
		if err := app.SyncPortalHooks(home); err != nil {
			fmt.Fprintf(os.Stderr, "lazypoot sync: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("lazypoot: hooks synced")
	}
}

func runDoctor() {
	results := app.RunDoctor(resolveHome())
	app.PrintDoctorResults(results)
}

func runDebug() {
	os.Setenv("LAZYPOOT_DEBUG", "1")
	home := resolveHome()
	fmt.Print(screens.FullVersion())
	fmt.Println("\nDebug info:")
	fmt.Println("  GODEBUG:", os.Getenv("GODEBUG"))
	fmt.Println("  SHELL:", os.Getenv("SHELL"))
	fmt.Println("  HOME:", home)
	fmt.Println("  PREFIX:", os.Getenv("PREFIX"))
	fmt.Println("\nDoctor results:")
	app.PrintDoctorResults(app.RunDoctor(home))
}

func resolveHome() string {
	home := os.Getenv("HOME")
	if home == "" {
		fmt.Fprintln(os.Stderr, "lazypoot: HOME not set")
		os.Exit(1)
	}
	return home
}

func printVersion(args []string) {
	for _, a := range args {
		if a == "--json" {
			fmt.Println(screens.VersionJSON())
			return
		}
	}
	fmt.Print(screens.FullVersion())
}

func runTUI() {
	os.Setenv("TERM", "xterm-256color")
	os.Setenv("COLORTERM", "truecolor")
	os.Setenv("GODEBUG", "netdns=go")

	prog := tea.NewProgram(
		app.NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lazypoot: fatal error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(screens.FullVersion())
	fmt.Print(`Usage:
  lazypoot                       Run interactive TUI
  lazypoot login <distro>        Login to a distro
  lazypoot sync                  Sync shell hooks from manifest
  lazypoot sync --dry-run        Preview hook changes
  lazypoot doctor                Run system checks
  lazypoot debug-replay          Show debug info + doctor
  lazypoot --version, -v         Show version
  lazypoot --help, -h            Show this help
`)
}
