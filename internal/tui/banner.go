package tui

import "strings"

// termuxvoidRows is the "termuxvoid" banner generated with
// `toilet -f smblock` (the Sam Hocevar small block font, shipped with the
// toilet package). It is embedded so the TUI has no runtime dependency on
// toilet/figlet. The small tvpkgRows variant is used on narrower screens.
var termuxvoidRows = padSplash([]string{
	"▐                        ▗   ▌",
	"▜▀ ▞▀▖▙▀▖▛▚▀▖▌ ▌▚▗▘▌ ▌▞▀▖▄ ▞▀▌",
	"▐ ▖▛▀ ▌  ▌▐ ▌▌ ▌▗▚ ▐▐ ▌ ▌▐ ▌ ▌",
	" ▀ ▝▀▘▘  ▘▝ ▘▝▀▘▘ ▘ ▘ ▝▀ ▀▘▝▀▘",
})

// tvpkgRows is the compact "tvpkg" banner, used when the terminal is wide
// enough for a banner but too narrow for the full termuxvoid mark.
var tvpkgRows = []string{
	"▐        ▌",
	"▜▀ ▌ ▌▛▀▖▌▗▘▞▀▌",
	"▐ ▖▐▐ ▙▄▘▛▚ ▚▄▌",
	" ▀  ▘ ▌  ▘ ▘▗▄▘",
}

// splash returns the banner art when the terminal is big enough to show it
// comfortably, adapting to screen size like other terminal package managers.
func (m Model) splash() ([]string, bool) {
	if m.width >= 34 {
		return termuxvoidRows, true
	}
	if m.width >= 19 {
		return tvpkgRows, true
	}
	return nil, false
}

// padSplash right-pads every row to the widest one so the block glyphs keep
// their shape and the whole banner centers as a single unit.
func padSplash(rows []string) []string {
	w := 0
	for _, r := range rows {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		if n := len([]rune(r)); n < w {
			out[i] = r + strings.Repeat(" ", w-n)
			continue
		}
		out[i] = r
	}
	return out
}
