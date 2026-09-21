// Package style provides small ANSI escape helpers for CLI output.
//
// Colors are enabled per-file-descriptor only when that descriptor is a
// terminal and the NO_COLOR environment variable is not set, so piped or
// redirected output stays plain text (matches the modern terminal convention).
package style

import (
	"os"
	"strconv"

	"github.com/mattn/go-isatty"
)

const reset = "\x1b[0m"

// Palette wraps the ANSI sequences for one output stream.
type Palette struct {
	on bool
}

// Out is the palette for stdout and Err the palette for stderr — they decide
// independently whether color is enabled, because a command's stdout may be
// piped while its prompts still hit an interactive stderr.
var (
	Out = Palette{on: use(os.Stdout)}
	Err = Palette{on: use(os.Stderr)}
)

func use(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isatty.IsTerminal(f.Fd())
}

func (p Palette) wrap(code string, s string) string {
	if !p.on {
		return s
	}
	return "\x1b[" + code + "m" + s + reset
}

func (p Palette) fg(code int, s string) string { return p.wrap(strconv.Itoa(code), s) }

// On reports whether ANSI codes are active for this palette.
func (p Palette) On() bool { return p.on }

// Text attributes.
func (p Palette) Bold(s string) string      { return p.wrap("1", s) }
func (p Palette) Dim(s string) string       { return p.wrap("2", s) }
func (p Palette) Underline(s string) string { return p.wrap("4", s) }

// Standard foreground colors (30–37).
func (p Palette) Black(s string) string   { return p.fg(30, s) }
func (p Palette) Red(s string) string     { return p.fg(31, s) }
func (p Palette) Green(s string) string   { return p.fg(32, s) }
func (p Palette) Yellow(s string) string  { return p.fg(33, s) }
func (p Palette) Blue(s string) string    { return p.fg(34, s) }
func (p Palette) Magenta(s string) string { return p.fg(35, s) }
func (p Palette) Cyan(s string) string    { return p.fg(36, s) }
func (p Palette) White(s string) string   { return p.fg(37, s) }

// Bright foreground colors (90–97), the punchier shades of the 16-color box.
func (p Palette) HiRed(s string) string     { return p.fg(91, s) }
func (p Palette) HiGreen(s string) string   { return p.fg(92, s) }
func (p Palette) HiYellow(s string) string  { return p.fg(93, s) }
func (p Palette) HiBlue(s string) string    { return p.fg(94, s) }
func (p Palette) HiMagenta(s string) string { return p.fg(95, s) }
func (p Palette) HiCyan(s string) string    { return p.fg(96, s) }
func (p Palette) HiWhite(s string) string   { return p.fg(97, s) }