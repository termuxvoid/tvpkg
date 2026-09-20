package pkgmanager

import (
	"testing"

	"tvpkg/internal/detect"
)

func TestParseInstalledIntoApt(t *testing.T) {
	out := "" +
		"ii  termux-ai\n" +
		"ii  outguess\n" +
		"rc  figlet\n" +
		"ii  nano\n"
	set := map[string]bool{}
	parseInstalledInto(set, detect.APT, out)
	for _, want := range []string{"termux-ai", "outguess", "nano"} {
		if !set[want] {
			t.Fatalf("expected %q marked installed, got %v", want, set)
		}
	}
	if set["figlet"] {
		t.Fatalf("rc-state package must NOT count as installed: %v", set)
	}
	if len(set) != 3 {
		t.Fatalf("installed set size=%d, want 3: %v", len(set), set)
	}
}

func TestParseInstalledIntoPacman(t *testing.T) {
	set := map[string]bool{}
	parseInstalledInto(set, detect.Pacman, "pacman-v6\nfiglet\n\n")
	if !set["pacman-v6"] || !set["figlet"] {
		t.Fatalf("pacman names missing: %v", set)
	}
}
