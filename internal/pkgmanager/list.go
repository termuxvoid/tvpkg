package pkgmanager

import (
	"strings"
)

// availableApt builds the package list from `apt-cache search .` (name +
// description) and marks installed ones via dpkg-query.
func (m *Manager) availableApt() ([]Package, error) {
	out, err := m.run.Capture("apt-cache", "search", ".")
	if err != nil {
		return nil, err
	}

	installed, _ := m.InstalledSet()
	pkgs := make([]Package, 0, 4096)
	for _, line := range splitLines(out) {
		if line == "" {
			continue
		}
		name, desc := line, ""
		if idx := strings.Index(line, " - "); idx >= 0 {
			name = line[:idx]
			desc = line[idx+3:]
		}
		if name == "" {
			continue
		}
		pkgs = append(pkgs, Package{
			Name:      name,
			Desc:      desc,
			Installed: installed[name],
		})
	}
	return pkgs, nil
}

// availablePacman builds the package list from `pacman -Sl` (repo, name,
// version). Descriptions are not available in a single bulk listing, so they
// are left empty and the detail view fetches them on demand.
func (m *Manager) availablePacman() ([]Package, error) {
	out, err := m.run.Capture("pacman", "-Sl")
	if err != nil {
		return nil, err
	}

	installed, _ := m.InstalledSet()
	pkgs := make([]Package, 0, 4096)
	for _, line := range splitLines(out) {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		pkg := Package{
			Name:      parts[1],
			Installed: installed[parts[1]],
		}
		if len(parts) >= 3 {
			pkg.Version = parts[2]
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r", "")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
