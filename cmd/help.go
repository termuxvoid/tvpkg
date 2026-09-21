package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"tvpkg/internal/style"
)

// Colored help/usage rendering. The ANSI helpers are no-ops when stdout is
// piped or NO_COLOR is set (see internal/style), so `tvpkg --help | cat`
// still produces plain text.
//
// Section headers are bold green, command names cyan, and supporting text
// (descriptions, flags, examples) is dimmed. Colors are applied only here —
// the rest of the CLI output stays as the underlying tool produces it.

// installColorHelp wires the colored help and usage functions onto the root
// command. Both are inherited by every subcommand through cobra's parent
// lookup.
func installColorHelp() {
	rootCmd.SetHelpFunc(colorHelp)
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, err := fmt.Fprint(cmd.OutOrStderr(), colorUsage(cmd))
		return err
	})
}

// colorHelp mirrors cobra's default help text with colors: the Long/Short
// description followed by the usage string.
func colorHelp(cmd *cobra.Command, _ []string) {
	out := cmd.OutOrStdout()
	if desc := strings.TrimRight(strings.TrimLeft(cmd.Long, "\n"), " \n"); desc != "" {
		fmt.Fprintln(out, desc)
	} else if cmd.Short != "" {
		fmt.Fprintln(out, cmd.Short)
	}
	fmt.Fprintln(out)
	fmt.Fprint(out, colorUsage(cmd))
}

// colorUsage renders the usage section for cmd.
func colorUsage(cmd *cobra.Command) string {
	bgreen := func(s string) string { return style.Out.Bold(style.Out.Green(s)) }
	cyan := func(s string) string { return style.Out.Cyan(s) }
	dim := func(s string) string { return style.Out.Dim(s) }

	var b strings.Builder
	b.WriteString(bgreen("Usage:") + "\n")
	if cmd.Runnable() {
		b.WriteString("  " + cmd.UseLine() + "\n")
	}
	if cmd.HasAvailableSubCommands() {
		b.WriteString("  " + cmd.CommandPath() + " [command]\n")
	}
	if len(cmd.Aliases) > 0 {
		b.WriteString("\n" + bgreen("Aliases:") + "\n  " + cyan(strings.Join(cmd.Aliases, ", ")) + "\n")
	}
	if cmd.HasExample() {
		b.WriteString("\n" + bgreen("Examples:") + "\n" + dim(strings.TrimRight(cmd.Example, "\n")) + "\n")
	}
	if cmd.HasAvailableSubCommands() {
		b.WriteString("\n" + bgreen("Available Commands:") + "\n")
		padding := cmd.NamePadding()
		for _, c := range cmd.Commands() {
			if !(c.IsAvailableCommand() || c.Name() == "help") {
				continue
			}
			b.WriteString("  " + cyan(fmt.Sprintf("%-*s", padding, c.Name())) + " " + dim(c.Short) + "\n")
		}
	}
	if cmd.HasAvailableLocalFlags() {
		b.WriteString("\n" + bgreen("Flags:") + "\n" + dim(strings.TrimRight(cmd.LocalFlags().FlagUsages(), " \n")) + "\n")
	}
	if cmd.HasAvailableInheritedFlags() {
		b.WriteString("\n" + bgreen("Global Flags:") + "\n" + dim(strings.TrimRight(cmd.InheritedFlags().FlagUsages(), " \n")) + "\n")
	}
	if cmd.HasHelpSubCommands() {
		b.WriteString("\n" + bgreen("Additional help topics:") + "\n")
		for _, c := range cmd.Commands() {
			if !c.IsAdditionalHelpTopicCommand() {
				continue
			}
			b.WriteString("  " + cyan(fmt.Sprintf("%-*s", cmd.CommandPathPadding(), c.CommandPath())) + " " + dim(c.Short) + "\n")
		}
	}
	if cmd.HasAvailableSubCommands() {
		b.WriteString("\n" + dim(fmt.Sprintf("Use \"%s [command] --help\" for more information about a command.", cmd.CommandPath())) + "\n")
	}
	return b.String()
}