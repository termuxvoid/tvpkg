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
		body := m.list.View()
		if body == "" {
			body = " "
		}
		return m.header() + "\n" + body + "\n" + m.actionBar() + "\n" + m.helpBar()

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
		sb.WriteString(help("↑/k", "up") + sep + help("↓/j", "down") + sep + help("/", "filter") +
			sep + help("tab", "action") + sep + help("enter", "run") + sep + help("q", "quit"))
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
