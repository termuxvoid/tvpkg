// Package cmd implements the tvpkg command-line interface.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tvpkg/internal/detect"
	"tvpkg/internal/pkgmanager"
	"tvpkg/internal/tui"
)

// version is replaced at build time via -ldflags -X.
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "tvpkg",
	Short: "A fast APT/Pacman package-manager wrapper for Termux with a TUI",
	Long: `tvpkg manages packages through Termux's package manager (APT or Pacman).

Run without arguments to open the interactive TUI, or use a subcommand.

tvpkg is also available as 'tvp' (same binary, symlink).`,
	SilenceUsage: true,
	Version:      version,
	Args:         cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unknown command %q (see 'tvpkg help')", args[0])
		}
		return runTUI()
	},
}

var (
	pkgmgrFlag  string
	yesFlag     bool
	simulateFlag bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&pkgmgrFlag, "pkgmgr", "",
		"package manager to use: apt or pacman (default: auto-detect)")
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false,
		"skip all confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&simulateFlag, "simulate", false,
		"dry-run: print the commands that would be executed without changing the system")
	rootCmd.AddCommand(installCmd, removeCmd, searchCmd, updateCmd, listCmd, infoCmd,
		upgradeCmd, cleanCmd, autocleanCmd, listInstalledCmd, filesCmd, fixCmd)
}

// Execute runs the CLI.
func Execute() error {
	return rootCmd.Execute()
}

// confirm prints msg and returns true when the user answers y/yes (case-
// insensitive). If --yes is set the prompt is skipped. An empty or invalid
// answer (including EOF / piped input) is treated as "no".
func confirm(msg string) bool {
	if yesFlag {
		return true
	}
	fmt.Fprintf(os.Stderr, "%s [y/N] ", msg)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	}
	return false
}

// resolve determines which package manager to use, honoring the --pkgmgr
// override, applying --simulate if requested, and refusing to run as root.
func resolve() (detect.Info, *pkgmanager.Manager, error) {
	var kind detect.Kind
	if pkgmgrFlag != "" {
		k, err := detect.Parse(pkgmgrFlag)
		if err != nil {
			return detect.Info{}, nil, err
		}
		kind = k
	}

	info, err := detect.DetectWithOverride("", kind)
	if err != nil {
		return detect.Info{}, nil, err
	}
	if detect.IsRoot() {
		return info, nil, fmt.Errorf("cannot run 'tvpkg' as root")
	}

	mgr, err := pkgmanager.New(info.Kind, info.Prefix)
	if err != nil {
		return detect.Info{}, nil, err
	}
	if simulateFlag {
		mgr.SetSimulate(true)
	}
	return info, mgr, nil
}

func runTUI() error {
	info, mgr, err := resolve()
	if err != nil {
		return err
	}
	_, err = tui.Run(tui.New(info, mgr))
	return err
}