package tui

// splashRows is the "tvpkg" banner generated with `toilet -f smblock`
// (the Sam Hocevar small block font, shipped with the toilet package).
// It is embedded so the TUI has no runtime dependency on toilet/figlet.
var splashRows = []string{
	"▐        ▌",
	"▜▀ ▌ ▌▛▀▖▌▗▘▞▀▌",
	"▐ ▖▐▐ ▙▄▘▛▚ ▚▄▌",
	" ▀  ▘ ▌  ▘ ▘▗▄▘",
}

// splash returns the banner art when the terminal is big enough to show
// it comfortably, adapting to screen size like other launcher tools.
func (m Model) splash() ([]string, bool) {
	if m.width < 19 || m.height < 12 {
		return nil, false
	}
	return splashRows, true
}
