package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Run starts the TUI and blocks until it exits.
func Run(m Model) (tea.Model, error) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	return p.Run()
}

// View renders the current state.
func (m Model) View() string {
	switch m.state {
	case stateLoading:
		return m.header() + "\n\n  " + m.spinner.View() + " " +
			lipgloss.NewStyle().Foreground(colLav).Bold(true).Render("Loading package list…")

	case stateBrowse:
		var body string
		switch {
		case m.list.FilterValue() == "":
			body = hintStyle.Render("  Type to search packages…")
		case len(m.list.VisibleItems()) == 0:
			body = hintStyle.Render("  No packages match \u201c" + m.list.FilterValue() + "\u201d")
		default:
			body = m.list.View()
		}
		return m.header() + "\n" + m.searchBox() + "\n" + m.statusLine() +
			"\n" + body + "\n" + m.actionBar() + "\n" + m.helpBar()

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

func (m Model) searchBox() string {
	w := m.width - 4
	if w < 12 {
		w = 12
	}
	inner := lipgloss.NewStyle().Width(w - 4).Render(m.list.FilterInput.View())
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colSurface1).
		Padding(0, 1).
		Render(inner)
	label := lipgloss.NewStyle().Foreground(colMauve).Bold(true).PaddingRight(1).Render("search")
	return "  " + label + box
}

func (m Model) statusLine() string {
	q := m.list.FilterValue()
	if q == "" {
		return statStyle.Render("  " + statNumStyle.Render(fmt.Sprintf("%d", len(m.pkgs))) +
			" packages available" +
			statStyle.Render("  ·  ") + statNumStyle.Render(fmt.Sprintf("%d", m.installed)) +
			" installed")
	}
	return statStyle.Render("  " + statNumStyle.Render(fmt.Sprintf("%d", len(m.list.VisibleItems()))) +
		" matches for \u201c" + q + "\u201d")
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
	return "  " + sb.String()
}

func (m Model) helpBar() string {
	sep := "   "
	var sb strings.Builder
	switch m.state {
	case stateBrowse:
		sb.WriteString(help("type", "search") + sep + help("↑/↓", "navigate") + sep +
			help("tab", "action") + sep + help("enter", "run") + sep +
			help("esc", "clear") + sep + help("ctrl+c", "quit"))
	case stateBusy:
		sb.WriteString(help("ctrl+c", "abort"))
	case stateResult, stateDetail:
		sb.WriteString(help("esc", "back") + sep + help("↑/↓", "scroll") + sep + help("q", "quit"))
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
