package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tvpkg/internal/pkgmanager"
)

// Run starts the TUI and blocks until it exits.
func Run(m Model) (tea.Model, error) {
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	return p.Run()
}

// View renders the current state.
func (m Model) View() string {
	switch m.state {
	case stateLoading:
		return m.header() + "\n\n  " + m.spinner.View() + " " +
			lipgloss.NewStyle().Foreground(colLav).Bold(true).Render("Loading package list…")

	case stateBrowse:
		if !m.searching {
			return m.homeView()
		}
		return m.popupView()

	case stateBusy:
		pane := m.truncate(lipgloss.NewStyle().Foreground(colSubtext).Render(m.liveOutput()), m.height-8)
		return m.header() + "\n\n  " + m.spinner.View() + " " + taskStyle.Render(m.task) +
			"\n\n" + outputPane.Render(pane) + "\n" + m.helpBar()

	case stateResult:
		heading := okStyle.Render("✓ Operation completed")
		if m.errMsg != "" {
			heading = errStyle.Render("✗ Operation failed: " + m.errMsg)
		}
		return m.header() + "\n\n  " + heading + "\n" + outputPane.Render(m.view.View()) + "\n" + m.helpBar()

	case stateDetail:
		return m.header() + "\n\n  " + infoTitleStyle.Render("Package info") +
			"\n" + outputPane.Render(m.view.View()) + "\n" + m.helpBar()

	case stateError:
		return m.header() + "\n\n" + errStyle.Render("✗ "+m.errMsg) + "\n\n  " +
			helpDescStyle.Render("press") + " " + helpKeyStyle.Render("q") + " " +
			helpDescStyle.Render("to quit")
	}
	return ""
}

// homeView is the launcher start screen shown before the search popup opens.
func (m Model) homeView() string {
	h, w := m.height, m.width
	if h < 1 {
		h = 1
	}
	if w < 1 {
		w = 1
	}

	var sb strings.Builder
	sb.WriteString("\n\n")
	sb.WriteString(m.centered(lipgloss.NewStyle().Foreground(colLav).Bold(true).Render("tvpkg"), w) + "\n")
	sb.WriteString(m.centered(lipgloss.NewStyle().Foreground(colSubtext).Render("Termux package launcher"), w) + "\n")
	sb.WriteString(m.centered(m.statLine(), w) + "\n\n")

	rows := m.resultsRows(m.popupHeight())
	hint1 := help("type", "search") + sepHelp + help("tab", "action") +
		sepHelp + help("enter", "run")
	hint2 := help("esc", "clear") + sepHelp + help("ctrl+c", "quit") + sepHelp +
		help("touch", "select / double-tap run")
	sb.WriteString(m.centered(hint1, w) + "\n")
	sb.WriteString(m.centered(hint2, w) + "\n\n")
	sb.WriteString(m.centered(lipgloss.NewStyle().Foreground(colOverlay).Italic(true).
		Render(fmt.Sprintf("hint: ~%d matches per page, ←/→ pages too", rows)), w) + "\n")

	return sb.String()
}

func (m Model) centered(s string, w int) string {
	if w < 1 {
		return s
	}
	sw := lipgloss.Width(s)
	if sw >= w {
		return s
	}
	return strings.Repeat(" ", (w-sw)/2) + s
}

func (m Model) statLine() string {
	return statStyle.Render(
		fmt.Sprintf("%d packages available", len(m.pkgs)) + "  ·  " +
			fmt.Sprintf("%d installed", m.installed))
}

// popupView renders the centered search window (vim telescope style).
func (m Model) popupView() string {
	w, h := m.width, m.height
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}

	pw, ph, padX, padY := m.popupRect()
	innerW := pw - 2
	k := m.resultsRows(ph)

	// Build the inner lines.
	var lines []string

	// Title.
	badge := "APT"
	if m.info.IsPacman() {
		badge = "Pacman"
	}
	title := lipgloss.NewStyle().Foreground(colLav).Bold(true).
		Render("tvpkg · " + badge + " launcher")
	lines = append(lines, " "+title)

	// Search input.
	searchLabel := lipgloss.NewStyle().Foreground(colMauve).Bold(true).Render("⌕ search")
	input := m.search.View()
	lines = append(lines, "  "+searchLabel+" "+input)

	// Status / match count.
	q := m.search.Value()
	if q == "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(colSubtext).
			Render("   "+statStyle.Render(fmt.Sprintf("%d", len(m.pkgs))+" packages available · "+
				fmt.Sprintf("%d", m.installed)+" installed ")))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(colSubtext).
			Render("   "+fmt.Sprintf("%d", len(m.results))+" matches for \u201c"+
				truncateRunes(q, 28)+"\u201d"))
	}

	// Results.
	if q == "" {
		for i := 0; i < k; i++ {
			lines = append(lines, firstRow(i, innerW, "⌕  "+lipgloss.NewStyle().
				Foreground(colOverlay).Italic(true).Render("type to search packages…")))
		}
	} else if len(m.results) == 0 {
		for i := 0; i < k; i++ {
			lines = append(lines, firstRow(i, innerW, "  "+lipgloss.NewStyle().
				Foreground(colOverlay).Italic(true).Render("no packages match \u201c"+truncateRunes(q, 24)+"\u201d")))
		}
	} else {
		pageStart := (m.cursor / k) * k
		if pageStart >= len(m.results) {
			pageStart = 0
		}
		end := pageStart + k
		if end > len(m.results) {
			end = len(m.results)
		}
		for i := pageStart; i < end; i++ {
			lines = append(lines, m.resultRow(m.results[i], i == m.cursor, innerW, q))
		}
		for i := end; i < pageStart+k; i++ {
			lines = append(lines, strings.Repeat(" ", innerW))
		}
	}

	// Action pills.
	lines = append(lines, "  "+m.actionBar())
	lines = append(lines, lipgloss.NewStyle().Foreground(colSubtext).Render(
		"  "+help("↑/↓", "move")+sepHelp+help("←/→", "page")+sepHelp+help("enter", "run")+
			sepHelp+help("tab", "action")+sepHelp+help("esc", "close"))+"  "+lipgloss.NewStyle().
		Foreground(colOverlay).Render("[tap select · double-tap run · scroll navigate]"))

	// Normalize line count to the inner height.
	for len(lines) < ph-2 {
		lines = append(lines, "")
	}
	lines = lines[:ph-2]

	// Render each inner line to a fixed width, then frame with a rounded border.
	inner := make([]string, len(lines))
	for i, ln := range lines {
		inner[i] = lipgloss.NewStyle().Width(innerW).Render(ln)
	}
	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colSurface1).
		Render(strings.Join(inner, "\n"))
	frameLines := strings.Split(frame, "\n")

	// Overlay the frame onto the screen, centered.
	screen := make([]string, h)
	for y := 0; y < h; y++ {
		var s string
		ly := y - padY
		if ly >= 0 && ly < len(frameLines) {
			s = strings.Repeat(" ", padX) + frameLines[ly]
		}
		if w > lipgloss.Width(s) {
			s += strings.Repeat(" ", w-lipgloss.Width(s))
		}
		screen[y] = s
	}
	return strings.Join(screen, "\n")
}

// firstRow fills blank result slots so the frame keeps a stable height.
func firstRow(i int, innerW int, msg string) string {
	if i == 0 {
		return lipgloss.NewStyle().Width(innerW).Render(msg)
	}
	return strings.Repeat(" ", innerW)
}

// resultRow renders one package row in the popup with selection marker,
// installed indicator, fuzzy-hit highlight and version.
func (m Model) resultRow(pkg pkgmanager.Package, selected bool, innerW int, q string) string {
	dot := dotStyle.Foreground(colOverlay).Render("○")
	if pkg.Installed {
		dot = dotStyle.Foreground(colGreen).Render("●")
	}
	marker := "  "
	if selected {
		marker = "❯ "
	}
	head := marker + dot + " "

	avail := innerW - 2 - lipgloss.Width(head)
	var meta string
	if pkg.Version != "" {
		meta = "  " + versionStyle.Render(pkg.Version)
		avail -= lipgloss.Width(meta)
	}

	// Truncate the plain (unstyled) name so no escape sequence is ever cut
	// in half; highlighting/style are applied afterwards.
	name := pkg.Name
	if lipgloss.Width(name) > avail {
		if avail < 1 {
			avail = 1
		}
		name = truncateRunes(name, avail-1) + "…"
	}
	name = m.queryHighlight(name, q)
	if selected {
		name = selectedNameStyle.Render(name)
	}

	st := lipgloss.NewStyle().Width(innerW)
	if selected {
		st = st.Background(colSurface0)
	}
	return st.Render(head + name + meta)
}

// popupRect returns the centered popup window geometry.
func (m Model) popupRect() (pw, ph, padX, padY int) {
	w, h := m.width, m.height
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}
	pw = w - 4
	if pw > 64 {
		pw = 64
	}
	if pw < 28 {
		pw = 28
	}
	if pw > w {
		pw = w
	}
	ph = h - 6
	if ph > 22 {
		ph = 22
	}
	if ph < 12 {
		ph = 12
	}
	if ph > h {
		ph = h
	}
	padX = (w - pw) / 2
	if padX < 0 {
		padX = 0
	}
	padY = (h - ph) / 2
	if padY < 0 {
		padY = 0
	}
	return pw, ph, padX, padY
}

// popupHeight returns just the height of the popup window.
func (m Model) popupHeight() int {
	_, ph, _, _ := m.popupRect()
	return ph
}

// resultsRows is how many result rows the popup shows (fixed chrome above).
func (m Model) resultsRows(ph int) int {
	k := ph - 8
	if k < 3 {
		k = 3
	}
	return k
}

func (m Model) header() string {
	badge := "APT"
	style := badgeAptStyle
	if m.info.IsPacman() {
		badge = "Pacman"
		style = badgePacmanStyle
	}
	return logoStyle.Render("tvpkg") + style.Render(badge) +
		statStyle.Render("  "+statNumStyle.Render(fmt.Sprintf("%d", m.installed))+
			" / "+statNumStyle.Render(fmt.Sprintf("%d", len(m.pkgs)))+" installed")
}

func (m Model) actionBar() string {
	var sb strings.Builder
	for a := actInstall; a < actCount; a++ {
		style := actionInactiveStyle
		if m.action == a {
			style = actionActiveStyle
		}
		sb.WriteString(style.Render(a.String()))
	}
	return strings.TrimSpace(sb.String())
}

const sepHelp = "   "

func (m Model) helpBar() string {
	var sb strings.Builder
	switch m.state {
	case stateBusy:
		sb.WriteString(help("ctrl+c", "abort"))
	case stateResult, stateDetail:
		sb.WriteString(help("esc", "back") + sepHelp + help("↑/↓", "scroll") + sepHelp + help("q", "quit"))
	}
	return "  " + sb.String()
}

func help(k, d string) string {
	return helpKeyStyle.Render(k) + " " + helpDescStyle.Render(d)
}

func (m Model) liveOutput() string {
	if m.live == nil {
		return ""
	}
	return m.truncate(m.live.Output(), m.height-8)
}

// truncate keeps only the last n lines of s.
func (m Model) truncate(s string, n int) string {
	if n < 1 {
		n = 1
	}
	s = strings.ReplaceAll(s, "\r", "")
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}
