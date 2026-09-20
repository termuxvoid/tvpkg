package pkgmanager

import (
	"io"
	"os"
	"strings"
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

// captureStdout runs fn with os.Stdout redirected and returns everything printed.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	r.Close()
	return string(out)
}

func TestMaintenanceCommands(t *testing.T) {
	os.Setenv(DryRunEnv, "1")
	defer os.Unsetenv(DryRunEnv)

	cases := []struct {
		name  string
		kind  detect.Kind
		op    func(m *Manager) error
		wants []string
	}{
		{"upgrade", detect.APT, (*Manager).Upgrade, []string{"apt update", "apt full-upgrade -y"}},
		{"upgrade", detect.Pacman, (*Manager).Upgrade, []string{"pacman -Syu --noconfirm"}},
		{"clean", detect.APT, (*Manager).Clean, []string{"apt clean"}},
		{"clean", detect.Pacman, (*Manager).Clean, []string{"pacman -Scc"}},
		{"autoclean", detect.APT, (*Manager).AutoClean, []string{"apt autoclean"}},
		{"autoclean", detect.Pacman, (*Manager).AutoClean, []string{"pacman -Sc"}},
		{"files", detect.APT, func(m *Manager) error { return m.Files("git", "nano") },
			[]string{"dpkg -L git nano"}},
		{"files", detect.Pacman, func(m *Manager) error { return m.Files("git") },
			[]string{"pacman -Ql git"}},
	}

	for _, tc := range cases {
		m, err := New(tc.kind, "/data/data/com.termux/files/usr")
		if err != nil {
			t.Fatalf("New(%s): %v", tc.kind, err)
		}
		out := captureStdout(t, func() { _ = tc.op(m) })
		for _, want := range tc.wants {
			if !strings.Contains(out, want) {
				t.Errorf("op=%s kind=%s: did not print %q (got %q)", tc.name, tc.kind, want, out)
			}
		}
	}
}
