// Package detect implements automatic APT/Pacman detection for Termux.
//
// Detection order mirrors termux-setup-package-manager:
//  1. $TERMUX_APP_PACKAGE_MANAGER (exported by termux-app v0.119.0+)
//  2. $TERMUX_MAIN_PACKAGE_FORMAT (exported by login script when the app is older; debian -> apt)
//  3. Presence of $PREFIX/bin/pacman or $PREFIX/bin/apt
//  4. Default to apt (Termux's own fallback)
package detect

import (
	"fmt"
	"os"
)

// DefaultPrefix is the standard Termux installation prefix.
const DefaultPrefix = "/data/data/com.termux/files/usr"

// Kind is the package manager type.
type Kind string

const (
	APT    Kind = "apt"
	Pacman Kind = "pacman"
)

// Info describes the detected environment.
type Info struct {
	Kind   Kind
	Prefix string
	Bin    string
}

// IsApt reports whether the manager is APT.
func (i Info) IsApt() bool { return i.Kind == APT }

// IsPacman reports whether the manager is Pacman.
func (i Info) IsPacman() bool { return i.Kind == Pacman }

// Prefix returns the Termux installation prefix, honoring $PREFIX.
func Prefix() string {
	if p := os.Getenv("PREFIX"); p != "" {
		return p
	}
	return DefaultPrefix
}

// Parse validates a user-supplied manager name.
func Parse(kind string) (Kind, error) {
	switch kind {
	case string(APT):
		return APT, nil
	case string(Pacman):
		return Pacman, nil
	case "":
		return "", nil
	}
	return "", fmt.Errorf("unsupported package manager %q (expected 'apt' or 'pacman')", kind)
}

// FromEnv detects the manager from Termux environment variables.
func FromEnv() (Kind, bool) {
	if v := os.Getenv("TERMUX_APP_PACKAGE_MANAGER"); v == string(APT) || v == string(Pacman) {
		return Kind(v), true
	}
	if v := os.Getenv("TERMUX_MAIN_PACKAGE_FORMAT"); v != "" {
		if v == "debian" {
			return APT, true
		}
		return Pacman, true
	}
	return "", false
}

// FromBinary detects the manager by probing the installed binaries.
func FromBinary(prefix string) (Kind, bool) {
	if isFile(prefix + "/bin/pacman") {
		return Pacman, true
	}
	if isFile(prefix + "/bin/apt") {
		return APT, true
	}
	return "", false
}

// Detect resolves the package manager for the current environment.
func Detect() (Info, error) {
	var info Info
	info.Prefix = Prefix()

	if kind, ok := FromEnv(); ok {
		info.Kind = kind
	} else if kind, ok := FromBinary(info.Prefix); ok {
		info.Kind = kind
	} else {
		info.Kind = APT
	}

	info.Bin = info.Prefix + "/bin/" + string(info.Kind)
	return info, nil
}

// DetectWithOverride detects the manager, honoring an optional override.
// An empty override means auto-detection. A parsed manager that is not
// present on disk still wins (the underlying tool will fail with a clear
// message when actually invoked).
func DetectWithOverride(prefix string, kind Kind) (Info, error) {
	if prefix == "" {
		prefix = Prefix()
	}
	info := Info{Prefix: prefix}

	switch kind {
	case APT, Pacman:
		info.Kind = kind
	default:
		if k, ok := FromEnv(); ok {
			info.Kind = k
		} else if k, ok := FromBinary(prefix); ok {
			info.Kind = k
		} else {
			info.Kind = APT
		}
	}

	info.Bin = info.Prefix + "/bin/" + string(info.Kind)
	return info, nil
}

// IsRoot reports whether tvpkg is running as root, which Termux forbids.
func IsRoot() bool { return os.Geteuid() == 0 }

func isFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}
