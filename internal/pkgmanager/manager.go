// Package pkgmanager provides a unified interface over Termux's two package
// managers: APT (dpkg) and Pacman. All external commands are executed through
// a Runner that supports dry-run mode for development.
package pkgmanager

import (
	"fmt"
	"strings"

	"tvpkg/internal/detect"
)

// Package is a single package listing.
type Package struct {
	Name      string
	Version   string
	Desc      string
	Installed bool
}

// Title implements list.DefaultItem.
func (p Package) Title() string { return p.Name }

// Description implements list.DefaultItem.
func (p Package) Description() string { return p.Desc }

// FilterValue implements bubbles/list.Item for fuzzy filtering.
func (p Package) FilterValue() string {
	return p.Name + " " + p.Desc
}

// Manager wraps one package manager backend.
type Manager struct {
	Kind   detect.Kind
	Prefix string
	run    *Runner
}

// New creates a Manager for the given kind.
func New(kind detect.Kind, prefix string) (*Manager, error) {
	if prefix == "" {
		prefix = detect.Prefix()
	}
	switch kind {
	case detect.APT, detect.Pacman:
	default:
		return nil, fmt.Errorf("unsupported package manager %q", kind)
	}
	return &Manager{Kind: kind, Prefix: prefix, run: NewRunner(prefix)}, nil
}

// Name returns the manager name ("apt" or "pacman").
func (m *Manager) Name() string { return string(m.Kind) }

// Update refreshes the package database.
func (m *Manager) Update() error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("apt", "update")
	case detect.Pacman:
		return m.run.Stream("pacman", "-Sy")
	}
	return m.bad()
}

// Install installs the given packages.
func (m *Manager) Install(pkgs ...string) error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("apt", append([]string{"install", "-y"}, pkgs...)...)
	case detect.Pacman:
		return m.run.Stream("pacman", append([]string{"-S", "--needed", "--noconfirm"}, pkgs...)...)
	}
	return m.bad()
}

// Remove removes the given packages.
func (m *Manager) Remove(pkgs ...string) error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("apt", append([]string{"remove", "-y"}, pkgs...)...)
	case detect.Pacman:
		return m.run.Stream("pacman", append([]string{"-Rcns", "--noconfirm"}, pkgs...)...)
	}
	return m.bad()
}

// Search searches packages by query (name or description).
func (m *Manager) Search(query string) error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("apt-cache", "search", query)
	case detect.Pacman:
		return m.run.Stream("pacman", "-Ss", query)
	}
	return m.bad()
}

// Info shows metadata for a package.
func (m *Manager) Info(pkg string) error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("apt", "show", pkg)
	case detect.Pacman:
		return m.run.Stream("pacman", "-Si", pkg)
	}
	return m.bad()
}

// List lists all available packages.
func (m *Manager) List() error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("apt", "list")
	case detect.Pacman:
		return m.run.Stream("pacman", "-Sl")
	}
	return m.bad()
}

// ListInstalled lists installed packages.
func (m *Manager) ListInstalled() error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("apt", "list", "--installed")
	case detect.Pacman:
		return m.run.Stream("pacman", "-Qq")
	}
	return m.bad()
}

// Upgrade upgrades all installed packages to their latest versions, mirroring
// pkg: apt refreshes first then full-upgrades; pacman does a single -Syu.
func (m *Manager) Upgrade() error {
	switch m.Kind {
	case detect.APT:
		if err := m.run.Stream("apt", "update"); err != nil {
			return err
		}
		return m.run.Stream("apt", "full-upgrade", "-y")
	case detect.Pacman:
		return m.run.Stream("pacman", "-Syu", "--noconfirm")
	}
	return m.bad()
}

// Clean removes every package from the cache (apt clean / pacman -Scc).
func (m *Manager) Clean() error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("apt", "clean")
	case detect.Pacman:
		return m.run.Stream("pacman", "-Scc")
	}
	return m.bad()
}

// AutoClean removes outdated package files from the cache, keeping the latest
// (apt autoclean / pacman -Sc).
func (m *Manager) AutoClean() error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("apt", "autoclean")
	case detect.Pacman:
		return m.run.Stream("pacman", "-Sc")
	}
	return m.bad()
}

// Files shows every file installed by the given packages
// (dpkg -L / pacman -Ql).
func (m *Manager) Files(pkgs ...string) error {
	switch m.Kind {
	case detect.APT:
		return m.run.Stream("dpkg", append([]string{"-L"}, pkgs...)...)
	case detect.Pacman:
		return m.run.Stream("pacman", append([]string{"-Ql"}, pkgs...)...)
	}
	return m.bad()
}

// InstallWithOutput installs packages and returns the captured output.
func (m *Manager) InstallWithOutput(pkgs ...string) (string, error) {
	switch m.Kind {
	case detect.APT:
		return m.run.Capture("apt", append([]string{"install", "-y"}, pkgs...)...)
	case detect.Pacman:
		return m.run.Capture("pacman", append([]string{"-S", "--needed", "--noconfirm"}, pkgs...)...)
	}
	return "", m.bad()
}

// RemoveWithOutput removes packages and returns the captured output.
func (m *Manager) RemoveWithOutput(pkgs ...string) (string, error) {
	switch m.Kind {
	case detect.APT:
		return m.run.Capture("apt", append([]string{"remove", "-y"}, pkgs...)...)
	case detect.Pacman:
		return m.run.Capture("pacman", append([]string{"-Rcns", "--noconfirm"}, pkgs...)...)
	}
	return "", m.bad()
}

// StartInstall runs an install in the background, streaming output live.
func (m *Manager) StartInstall(pkgs ...string) (*Live, error) {
	switch m.Kind {
	case detect.APT:
		return m.run.RunLive("apt", append([]string{"install", "-y"}, pkgs...)...)
	case detect.Pacman:
		return m.run.RunLive("pacman", append([]string{"-S", "--needed", "--noconfirm"}, pkgs...)...)
	}
	return nil, m.bad()
}

// StartRemove runs a remove in the background, streaming output live.
func (m *Manager) StartRemove(pkgs ...string) (*Live, error) {
	switch m.Kind {
	case detect.APT:
		return m.run.RunLive("apt", append([]string{"remove", "-y"}, pkgs...)...)
	case detect.Pacman:
		return m.run.RunLive("pacman", append([]string{"-Rcns", "--noconfirm"}, pkgs...)...)
	}
	return nil, m.bad()
}

// InfoWithOutput returns package metadata as text.
func (m *Manager) InfoWithOutput(pkg string) (string, error) {
	switch m.Kind {
	case detect.APT:
		return m.run.Capture("apt-cache", "show", pkg)
	case detect.Pacman:
		return m.run.Capture("pacman", "-Si", pkg)
	}
	return "", m.bad()
}

// Available fetches the full package list with installed flags. The list is
// fetched once (client-side) so the TUI can filter instantly as you type.
func (m *Manager) Available() ([]Package, error) {
	if m.Kind == detect.Pacman {
		return m.availablePacman()
	}
	return m.availableApt()
}

// InstalledSet returns the set of installed package names.
func (m *Manager) InstalledSet() (map[string]bool, error) {
	set := map[string]bool{}
	var out string
	var err error

	switch m.Kind {
	case detect.APT:
		out, err = m.run.Capture("dpkg-query", "-W", "-f=${db:Status-Abbrev} ${binary:Package}\n")
	case detect.Pacman:
		out, err = m.run.Capture("pacman", "-Qq")
	default:
		return nil, m.bad()
	}
	if err != nil {
		return nil, err
	}
	parseInstalledInto(set, m.Kind, out)
	return set, nil
}

// parseInstalledInto fills set with fully-installed package names. apt's
// dpkg-query lists packages in any dpkg state (for example "rc": removed but
// config files remain), which a remove already failed to purge — only "ii"
// (desired install, status installed) counts as installed.
func parseInstalledInto(set map[string]bool, kind detect.Kind, out string) {
	for _, line := range splitLines(out) {
		if line == "" {
			continue
		}
		if kind == detect.APT {
			f := strings.Fields(line)
			if len(f) >= 2 && f[0] == "ii" && f[1] != "" {
				set[f[1]] = true
			}
			continue
		}
		set[line] = true
	}
}

func (m *Manager) bad() error {
	return fmt.Errorf("unsupported package manager %q", m.Kind)
}
