package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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

// pkgItem adapts a pkgmanager.Package for the bubbles/list component.
type pkgItem struct {
	pkg pkgmanager.Package
}

func (p pkgItem) Title() string { return p.pkg.Name }

func (p pkgItem) Description() string {
	var sb strings.Builder
	if p.pkg.Installed {
		sb.WriteString("● ")
	} else {
		sb.WriteString("○ ")
	}
	if p.pkg.Version != "" {
		sb.WriteString(p.pkg.Version + " ")
	}
	if p.pkg.Desc != "" {
		sb.WriteString(p.pkg.Desc)
	}
	return sb.String()
}

func (p pkgItem) FilterValue() string {
	return p.pkg.Name + " " + p.pkg.Desc
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

// Model is the tvpkg TUI.
type Model struct {
	info detect.Info
	mgr  *pkgmanager.Manager

	width, height int
	state         state
	action        action

	list    list.Model
	spinner spinner.Model
	view    viewport.Model

	pkgs      []pkgmanager.Package
	installed int
	selector  int

	task    string
	live    *pkgmanager.Live
	result  string
	errMsg  string
	infoOut string
}

// New creates the TUI model.
func New(info detect.Info, mgr *pkgmanager.Manager) Model {
	d := pkgDelegate{}
	m := Model{
		info:    info,
		mgr:     mgr,
		state:   stateLoading,
		action:  actInstall,
		list:    list.New([]list.Item{}, d, 40, 20),
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(spinnerStyle)),
		view:    viewport.New(40, 10),
	}
	m.list.SetShowTitle(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	m.list.SetShowPagination(false)
	// The search box and pagination/status line are rendered by us (see views.go)
	// so the TUI behaves like a search-first (fzf-style) launcher.
	m.list.SetShowFilter(false)
	m.list.SetFilteringEnabled(true)
	styles := list.DefaultStyles()
	styles.FilterPrompt = lipgloss.NewStyle().Foreground(colLav).Bold(true).Padding(0, 1)
	styles.FilterCursor = lipgloss.NewStyle().Foreground(colMauve)
	styles.PaginationStyle = paginationStyle
	m.list.Styles = styles
	// We draw our own "search" label, so drop the built-in prompt.
	m.list.FilterInput.Prompt = ""
	m.list.FilterInput.PromptStyle = lipgloss.NewStyle()
	m.list.FilterInput.Placeholder = "search packages…"
	m.list.FilterInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(colOverlay)
	m.list.FilterInput.CursorStyle = lipgloss.NewStyle().Foreground(colMauve)
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
		h := m.height - 5
		if h < 3 {
			h = 3
		}
		m.list.SetSize(m.width-2, h)
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

	case list.FilterMatchesMsg:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd

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
	items := make([]list.Item, len(msg.pkgs))
	for i, p := range msg.pkgs {
		items[i] = pkgItem{pkg: p}
	}
	m.state = stateBrowse
	cmd := m.list.SetItems(items)
	// Open straight into the fuzzy search so the user starts typing.
	m.list.SetFilterText("")
	m.list.SetFilterState(list.Filtering)
	return m, tea.Batch(cmd, textinput.Blink)
}

func (m Model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Keep the fuzzy search input active at all times while browsing.
	if m.list.FilterState() != list.Filtering {
		m.list.SetFilterState(list.Filtering)
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "tab":
		m.cycleAction(1)
		return m, nil
	case "shift+tab":
		m.cycleAction(-1)
		return m, nil

	case "enter":
		// Enter runs the focused action on the selected search result.
		if m.list.FilterValue() != "" && len(m.list.VisibleItems()) > 0 {
			return m, m.runAction()
		}
		return m, nil

	case "esc":
		// Clear the query but stay in search mode (never dump the full list).
		m.list.SetFilterText("")
		m.list.SetFilterState(list.Filtering)
		return m, textinput.Blink
	}

	// Navigation keys are consumed here so they move the result cursor
	// instead of being typed into the search box.
	switch msg.String() {
	case "up":
		m.list.CursorUp()
		return m, nil
	case "down":
		m.list.CursorDown()
		return m, nil
	case "pgup":
		m.list.PrevPage()
		return m, nil
	case "pgdown":
		m.list.NextPage()
		return m, nil
	}

	// Everything else goes to the list, i.e. the fuzzy search input.
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "enter":
		m.state = stateBrowse
		return m, nil
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
		return m, nil
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
	item, ok := m.list.SelectedItem().(pkgItem)
	if !ok {
		return pkgmanager.Package{}, false
	}
	return item.pkg, true
}

// runAction starts the selected package's focused action.
func (m Model) runAction() tea.Cmd {
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

func (m Model) startTask(live *pkgmanager.Live) tea.Cmd {
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

// pkgDelegate renders package rows with an installed indicator and fuzzy
// match highlighting.
type pkgDelegate struct{}

func (d pkgDelegate) Height() int                               { return 2 }
func (d pkgDelegate) Spacing() int                              { return 0 }
func (d pkgDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d pkgDelegate) ShortHelp() []key.Binding                  { return nil }
func (d pkgDelegate) FullHelp() []key.Binding                   { return nil }

func (d pkgDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	pi := item.(pkgItem)
	pkg := pi.pkg
	selected := m.Index() == index

	dot := lipgloss.NewStyle().Foreground(colOverlay).Render("○")
	if pkg.Installed {
		dot = lipgloss.NewStyle().Foreground(colGreen).Render("●")
	}

	marker := "  "
	if selected {
		marker = "❯ "
	}

	var title string
	if selected {
		title = selectedTitleStyle.Render(pkg.Name)
	} else {
		title = titleStyle.Render(pkg.Name)
	}

	var meta strings.Builder
	if pkg.Version != "" {
		if selected {
			meta.WriteString(versionStyle.Render(pkg.Version) + "  ")
		} else {
			meta.WriteString(versionStyle.Render(pkg.Version) + "  ")
		}
	}
	if pkg.Desc != "" {
		if selected {
			meta.WriteString(selectedDescStyle.Render(pkg.Desc))
		} else {
			meta.WriteString(descStyle.Render(pkg.Desc))
		}
	}

	fmt.Fprintf(w, "%s%s %s\n      %s\n", marker, dot, title, meta.String())
}
