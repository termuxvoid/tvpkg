package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tvpkg/internal/detect"
	"tvpkg/internal/pkgmanager"
)

const taskTickInterval = 120 * time.Millisecond

type state int

const (
	stateLoading state = iota
	stateBrowse
	stateBusy
	stateResult
	stateDetail
	stateError
)

type action int

const (
	actInstall action = iota
	actRemove
	actInfo
	actCount
)

func (a action) String() string {
	switch a {
	case actInstall:
		return "Install"
	case actRemove:
		return "Remove"
	case actInfo:
		return "Info"
	}
	return ""
}

// Messages.
type loadPkgsMsg struct {
	pkgs      []pkgmanager.Package
	installed int
	err       error
}

type infoLoadedMsg struct {
	name string
	out  string
	err  error
}

type taskDoneMsg struct {
	err error
	out string
}

type taskTickMsg time.Time

// tap records the last touch/click position so a quick second tap on the
// same row triggers the focused action (mouse/touchscreen support).
type tap struct {
	x, y int
	at   time.Time
}

// Model is the tvpkg TUI.
type Model struct {
	info detect.Info
	mgr  *pkgmanager.Manager

	width, height int
	state         state
	action        action

	search    textinput.Model
	searching bool
	spinner   spinner.Model
	view      viewport.Model

	pkgs      []pkgmanager.Package
	results   []pkgmanager.Package
	installed int
	cursor    int

	lastTap tap

	task    string
	live    *pkgmanager.Live
	result  string
	errMsg  string
	infoOut string
}

// New creates the TUI model.
func New(info detect.Info, mgr *pkgmanager.Manager) Model {
	search := textinput.New()
	search.Prompt = ""
	search.PromptStyle = lipgloss.NewStyle()
	search.Placeholder = "search packages…"
	search.PlaceholderStyle = lipgloss.NewStyle().Foreground(colOverlay)
	search.CursorStyle = lipgloss.NewStyle().Foreground(colMauve)
	search.CharLimit = 80
	search.Focus()

	m := Model{
		info:    info,
		mgr:     mgr,
		state:   stateLoading,
		action:  actInstall,
		search:  search,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(spinnerStyle)),
		view:    viewport.New(40, 10),
	}
	return m
}

// Init starts package loading and the spinner.
func (m Model) Init() tea.Cmd {
	return tea.Batch(loadPkgsCmd(m.mgr), spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.search.Width = m.width - 4
		if m.search.Width < 10 {
			m.search.Width = 10
		}
		w, h := m.width-4, m.height-6
		if w < 10 {
			w = 10
		}
		if h < 3 {
			h = 3
		}
		m.view.Width, m.view.Height = w, h
		return m, nil

	case loadPkgsMsg:
		return m.handleLoaded(msg)

	case infoLoadedMsg:
		m.state = stateDetail
		if msg.err != nil {
			m.infoOut = errStyle.Render("Failed to fetch info: " + msg.err.Error())
			m.view.SetContent(m.infoOut)
		} else {
			m.infoOut = msg.out
			m.view.SetContent(msg.out)
		}
		return m, nil

	case taskDoneMsg:
		m.live = nil
		m.state = stateResult
		m.result = msg.out
		m.errMsg = ""
		if msg.err != nil {
			m.errMsg = msg.err.Error()
		}
		if msg.out == "" {
			msg.out = "(no output)"
		}
		m.view.SetContent(msg.out)
		if msg.err != nil {
			return m, nil
		}
		// Refresh installed state after a successful operation.
		return m, loadPkgsCmd(m.mgr)

	case taskTickMsg:
		if m.state == stateBusy {
			return m, tea.Tick(taskTickInterval, func(t time.Time) tea.Msg { return taskTickMsg(t) })
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.state == stateLoading || m.state == stateBusy {
			return m, cmd
		}
		return m, nil

	case tea.MouseMsg:
		switch m.state {
		case stateBrowse:
			return m.updateBrowseMouse(msg)
		case stateResult, stateDetail:
			var cmd tea.Cmd
			m.view, cmd = m.view.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		switch m.state {
		case stateBrowse:
			return m.updateBrowse(msg)
		case stateDetail:
			return m.updateDetail(msg)
		case stateResult:
			return m.updateResult(msg)
		case stateLoading, stateBusy:
			return m.updateBusy(msg)
		}
	}

	return m, nil
}

func (m Model) handleLoaded(msg loadPkgsMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.state = stateError
		m.errMsg = msg.err.Error()
		return m, nil
	}
	m.pkgs = msg.pkgs
	m.installed = msg.installed
	if m.state != stateLoading && m.state != stateResult {
		return m, nil
	}
	m.refilter()
	if len(m.results) > 0 {
		if m.cursor >= len(m.results) {
			m.cursor = len(m.results) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
	m.state = stateBrowse
	if m.searching {
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		// Close the popup back to the home screen; never dump the full list.
		if m.searching {
			m.searching = false
			m.search.SetValue("")
			m.results = nil
			m.cursor = 0
		}
		return m, nil

	case "enter":
		// Enter runs the focused action on the selected result.
		if m.searching && m.search.Value() != "" && len(m.results) > 0 {
			return m, m.runAction()
		}
		return m, nil

	case "tab", "shift+tab":
		if m.searching {
			d := 1
			if msg.String() == "shift+tab" {
				d = -1
			}
			m.cycleAction(d)
		}
		return m, nil

	case "up", "down", "left", "right", "pgup", "pgdown", "home", "end":
		// Arrow keys navigate the results, never the search text cursor.
		if m.searching && len(m.results) > 0 {
			m.moveCursorKey(msg.String())
		}
		return m, nil
	}

	// Everything else is typed into the search box; the first printable
	// key opens the popup (telescope-style live search).
	if !m.searching {
		m.searching = true
	}
	old := m.search.Value()
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	if m.search.Value() != old {
		m.refilter()
		m.cursor = 0
	}
	return m, cmd
}

func (m Model) updateBrowseMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if !m.searching || len(m.results) == 0 {
		return m, nil
	}
	_, ph, _, padY := m.popupRect()
	innerY := msg.Y - (padY + 1)
	const resultsTop = 3
	if innerY >= resultsTop && msg.Action == tea.MouseActionPress &&
		msg.Button == tea.MouseButtonLeft {
		k := m.resultsRows(ph)
		row := innerY - resultsTop
		if row >= 0 && row < k {
			pageStart := (m.cursor / k) * k
			idx := pageStart + row
			if idx >= 0 && idx < len(m.results) {
				now := time.Now()
				if m.lastTap.at.After(time.Time{}) &&
					now.Sub(m.lastTap.at) < 400*time.Millisecond &&
					msg.X == m.lastTap.x && msg.Y == m.lastTap.y {
					// Double tap: run the focused action.
					m.cursor = idx
					m.lastTap = tap{}
					return m, m.runAction()
				}
				m.cursor = idx
				m.lastTap = tap{x: msg.X, y: msg.Y, at: now}
			}
		}
	}
	if msg.Action == tea.MouseActionPress {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.moveCursor(-3)
		case tea.MouseButtonWheelDown:
			m.moveCursor(3)
		}
	}
	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "enter":
		m.state = stateBrowse
		return m, textinput.Blink
	case "ctrl+c":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.view, cmd = m.view.Update(msg)
	return m, cmd
}

func (m Model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.state = stateBrowse
		return m, textinput.Blink
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.view, cmd = m.view.Update(msg)
	return m, cmd
}

func (m Model) updateBusy(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) cycleAction(d int) {
	m.action = action((int(m.action) + d + int(actCount)) % int(actCount))
}

func (m Model) selected() (pkgmanager.Package, bool) {
	if m.cursor < 0 || m.cursor >= len(m.results) {
		return pkgmanager.Package{}, false
	}
	return m.results[m.cursor], true
}

// moveCursorKey moves the result cursor. up/down step one row; left/right
// and pgup/pgdown page through the results; home/end jump to the edges.
func (m *Model) moveCursorKey(k string) {
	n := len(m.results)
	if n == 0 {
		return
	}
	page := m.resultsRows(m.popupHeight())
	switch k {
	case "up":
		m.cursor--
	case "down":
		m.cursor++
	case "left", "pgup":
		m.cursor -= page
	case "right", "pgdown":
		m.cursor += page
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
}

func (m *Model) moveCursor(d int) {
	if len(m.results) == 0 {
		return
	}
	m.cursor += d
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.results) {
		m.cursor = len(m.results) - 1
	}
}

// refilter rebuilds the result slice from the current search query using a
// case-insensitive substring match on the name (preferred) and description.
func (m *Model) refilter() {
	q := strings.ToLower(m.search.Value())
	m.results = m.results[:0]
	for _, p := range m.pkgs {
		if q == "" ||
			strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Desc), q) {
			m.results = append(m.results, p)
		}
	}
}

// runAction starts the selected package's focused action. It is a pointer
// receiver so the state changes (busy/result) survive into the returned Model.
func (m *Model) runAction() tea.Cmd {
	pkg, ok := m.selected()
	if !ok {
		return nil
	}

	switch m.action {
	case actInstall:
		live, err := m.mgr.StartInstall(pkg.Name)
		if err != nil {
			m.state = stateResult
			m.errMsg = err.Error()
			return nil
		}
		m.task = "Installing " + pkg.Name
		return m.startTask(live)
	case actRemove:
		live, err := m.mgr.StartRemove(pkg.Name)
		if err != nil {
			m.state = stateResult
			m.errMsg = err.Error()
			return nil
		}
		m.task = "Removing " + pkg.Name
		return m.startTask(live)
	case actInfo:
		m.state = stateDetail
		m.infoOut = ""
		return infoCmd(m.mgr, pkg.Name)
	}
	return nil
}

func (m *Model) startTask(live *pkgmanager.Live) tea.Cmd {
	m.state = stateBusy
	m.live = live
	m.result = ""
	m.errMsg = ""
	cmds := []tea.Cmd{
		waitCmd(live),
		tea.Tick(taskTickInterval, func(t time.Time) tea.Msg { return taskTickMsg(t) }),
	}
	return tea.Batch(cmds...)
}

func waitCmd(live *pkgmanager.Live) tea.Cmd {
	return func() tea.Msg {
		err := live.Wait()
		return taskDoneMsg{err: err, out: live.Output()}
	}
}

func loadPkgsCmd(mgr *pkgmanager.Manager) tea.Cmd {
	return func() tea.Msg {
		pkgs, err := mgr.Available()
		if err != nil {
			return loadPkgsMsg{err: err}
		}
		installed := 0
		for _, p := range pkgs {
			if p.Installed {
				installed++
			}
		}
		return loadPkgsMsg{pkgs: pkgs, installed: installed}
	}
}

func infoCmd(mgr *pkgmanager.Manager, name string) tea.Cmd {
	return func() tea.Msg {
		out, err := mgr.InfoWithOutput(name)
		return infoLoadedMsg{name: name, out: out, err: err}
	}
}

// queryHighlight wraps the matched substring of name in a highlighted style.
func (m Model) queryHighlight(name, q string) string {
	q = strings.ToLower(q)
	if q == "" {
		return name
	}
	i := strings.Index(strings.ToLower(name), q)
	if i < 0 {
		return name
	}
	return name[:i] + hlStyle.Render(name[i:i+len(q)]) + name[i+len(q):]
}

func truncateRunes(s string, max int) string {
	if max < 1 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
