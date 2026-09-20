package tui

import "github.com/charmbracelet/lipgloss"

// Catppuccin Mocha palette.
const (
	cBase     = "#1e1e2e"
	cMantle   = "#181825"
	cSurface0 = "#313244"
	cSurface1 = "#45475a"
	cOverlay0 = "#6c7086"
	cText     = "#cdd6f4"
	cSubtext0 = "#a6adc8"
	cBlue     = "#89b4fa"
	cLavender = "#b4befe"
	cGreen    = "#a6e3a1"
	cYellow   = "#f9e2af"
	cRed      = "#f38ba8"
	cMauve    = "#cba6f7"
	cPeach    = "#fab387"
)

var (
	colText     = lipgloss.Color(cText)
	colSubtext  = lipgloss.Color(cSubtext0)
	colBlue     = lipgloss.Color(cBlue)
	colLav      = lipgloss.Color(cLavender)
	colGreen    = lipgloss.Color(cGreen)
	colYellow   = lipgloss.Color(cYellow)
	colRed      = lipgloss.Color(cRed)
	colMauve    = lipgloss.Color(cMauve)
	colOverlay  = lipgloss.Color(cOverlay0)
	colSurface  = lipgloss.Color(cSurface0)
	colSurface1 = lipgloss.Color(cSurface1)

	appStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MaxWidth(60)

	logoStyle = lipgloss.NewStyle().
			Foreground(colLav).
			Bold(true)

	headerTextStyle = lipgloss.NewStyle().
			Foreground(colSubtext)

	statStyle = lipgloss.NewStyle().
			Foreground(colSubtext)

	statNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(cPeach))

	badgeAptStyle = lipgloss.NewStyle().
			Foreground(colText).
			Background(colBlue).
			Bold(true).
			Padding(0, 1).
			MarginLeft(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBlue)

	badgePacmanStyle = lipgloss.NewStyle().
				Foreground(colText).
				Background(colGreen).
				Bold(true).
				Padding(0, 1).
				MarginLeft(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colGreen)

	titleStyle = lipgloss.NewStyle().
			Foreground(colLav).
			Bold(true)

	selectedTitleStyle = lipgloss.NewStyle().
				Foreground(colMauve).
				Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(colSubtext)

	selectedDescStyle = lipgloss.NewStyle().
				Foreground(colText)

	versionStyle = lipgloss.NewStyle().
			Foreground(colYellow)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(colLav)

	taskStyle = lipgloss.NewStyle().
			Foreground(colText).
			Bold(true)

	paginationStyle = lipgloss.NewStyle().
			Foreground(colSubtext)

	hintStyle = lipgloss.NewStyle().
			Foreground(colSubtext).Italic(true)

	actionActiveStyle = lipgloss.NewStyle().
				Background(colMauve).
				Foreground(lipgloss.Color(cBase)).
				Bold(true).
				Padding(0, 1).
				MarginRight(1)

	actionInactiveStyle = lipgloss.NewStyle().
				Background(colSurface1).
				Foreground(colSubtext).
				Padding(0, 1).
				MarginRight(1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colMauve).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colSubtext)

	okStyle = lipgloss.NewStyle().
		Foreground(colGreen).
		Bold(true)

	errStyle = lipgloss.NewStyle().
			Foreground(colRed).
			Bold(true)

	infoTitleStyle = lipgloss.NewStyle().
			Foreground(colLav).
			Bold(true)

	outputPane = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colSurface1).
			Padding(0, 1)
)
