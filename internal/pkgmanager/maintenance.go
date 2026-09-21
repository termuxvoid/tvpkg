package pkgmanager

import (
	"os"
	"path/filepath"
	"strings"

	"tvpkg/internal/detect"
)

// SetSimulate forces (or disables) dry-run mode on an existing Manager. Used
// by the --simulate CLI flag so a script can preview a command without
// touching the system.
func (m *Manager) SetSimulate(v bool) { m.run.DryRun = v }

// Simulate reports whether the manager is in dry-run mode.
func (m *Manager) Simulate() bool { return m != nil && m.run != nil && m.run.DryRun }

// PlanUpgrade returns a human-readable preview of what an upgrade would do,
// without changing anything (apt full-upgrade -s / pacman -Syu --print).
func (m *Manager) PlanUpgrade() (string, error) {
	switch m.Kind {
	case detect.APT:
		return m.run.Capture("apt", "full-upgrade", "-s")
	case detect.Pacman:
		return m.run.Capture("pacman", "-Syu", "--print")
	}
	return "", m.bad()
}

// ReverseDepends returns the names of installed packages that depend on pkg.
// Used to warn before removing something other packages rely on. It is
// best-effort: an empty result means "no known reverse dependencies".
func (m *Manager) ReverseDepends(pkg string) ([]string, error) {
	var out string
	var err error
	switch m.Kind {
	case detect.APT:
		out, err = m.run.Capture("apt-cache", "rdepends", "--installed", pkg)
		if err != nil {
			return nil, err
		}
		return parseRdepApt(out), nil
	case detect.Pacman:
		out, err = m.run.Capture("pacman", "-Qi", pkg)
		if err != nil {
			return nil, err
		}
		return parseRdepPacman(out), nil
	}
	return nil, m.bad()
}

// parseRdepApt extracts reverse dependencies from `apt-cache rdepends` output,
// e.g.
//
//	git
//	Reverse Depends:
//	  git-man
//	  | openssh-client
//
// Only concrete names are returned (alternatives starting with "|" and the
// "Reverse Depends:" header are skipped).
func parseRdepApt(out string) []string {
	var deps []string
	seen := map[string]bool{}
	for _, line := range splitLines(out) {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "|") ||
			strings.HasSuffix(t, ":") || t == "Reverse Depends" {
			continue
		}
		if !strings.HasPrefix(line, " ") {
			continue // the queried package name on the first line
		}
		if !seen[t] {
			seen[t] = true
			deps = append(deps, t)
		}
	}
	return deps
}

// parseRdepPacman extracts the "Required By" list from `pacman -Qi` output.
func parseRdepPacman(out string) []string {
	for _, line := range splitLines(out) {
		if !strings.HasPrefix(line, "Required By") {
			continue
		}
		_, val, ok := strings.Cut(line, ":")
		if !ok {
			return nil
		}
		val = strings.TrimSpace(val)
		if val == "" || val == "None" {
			return nil
		}
		return strings.Fields(val)
	}
	return nil
}

// Fix inspects the package database for an interrupted or inconsistent state
// and returns a human-readable diagnosis. It never changes anything.
func (m *Manager) Fix() (string, error) {
	switch m.Kind {
	case detect.APT:
		out, err := m.run.Capture("dpkg", "--audit")
		if err != nil && strings.TrimSpace(out) == "" {
			return "", err
		}
		if strings.TrimSpace(out) == "" {
			return "dpkg database is consistent — nothing to fix.", nil
		}
		return "dpkg reports packages needing attention:\n\n" + strings.TrimSpace(out), nil
	case detect.Pacman:
		lock := filepath.Join(m.Prefix, "var", "lib", "pacman", "db.lck")
		var b strings.Builder
		if _, err := os.Stat(lock); err == nil {
			b.WriteString("pacman database is locked:\n  " + lock + "\n" +
				"If no pacman is running, remove that file and retry.\n\n")
		}
		out, err := m.run.Capture("pacman", "-Dk")
		if err != nil {
			b.WriteString("pacman database integrity check reported problems:\n\n" +
				strings.TrimSpace(out))
			return b.String(), nil
		}
		if b.Len() == 0 {
			return "pacman database is consistent — nothing to fix.", nil
		}
		b.WriteString("pacman database integrity check passed.")
		return b.String(), nil
	}
	return "", m.bad()
}

// Repair runs the standard recovery commands for the active package manager:
// dpkg --configure -a followed by apt --fix-broken install for APT, or a full
// pacman sync for pacman. Callers should confirm with the user first.
func (m *Manager) Repair() error {
	switch m.Kind {
	case detect.APT:
		if err := m.run.Stream("dpkg", "--configure", "-a"); err != nil {
			return err
		}
		return m.run.Stream("apt", "--fix-broken", "install", "-y")
	case detect.Pacman:
		return m.run.Stream("pacman", "-Syu", "--noconfirm")
	}
	return m.bad()
}
