package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpRenders(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	defer func() {
		rootCmd.SetOut(rootCmd.OutOrStdout())
		rootCmd.SetErr(rootCmd.OutOrStderr())
	}()

	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	s := out.String()
	for _, want := range []string{
		"tvpkg manages packages through Termux's package manager",
		"Usage:",
		"Available Commands:",
		"install",
		"tvpkg is also available as 'tvp'",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("help output missing %q:\n%s", want, s)
		}
	}
}

func TestSubcommandHelpRenders(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	defer func() {
		rootCmd.SetOut(rootCmd.OutOrStdout())
		rootCmd.SetErr(rootCmd.OutOrStderr())
	}()

	rootCmd.SetArgs([]string{"help", "install"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	s := out.String()
	for _, want := range []string{"Usage:", "tvpkg install <pkgs...>", "Aliases:", "Flags:"} {
		if !strings.Contains(s, want) {
			t.Fatalf("install help missing %q:\n%s", want, s)
		}
	}
}