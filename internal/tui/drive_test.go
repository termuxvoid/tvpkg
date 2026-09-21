package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"tvpkg/internal/detect"
	"tvpkg/internal/pkgmanager"
)

func TestDrive(t *testing.T) {
	pkgs := []pkgmanager.Package{
		{Name: "git", Version: "2.47.0", Desc: "distributed version control"},
		{Name: "gitsome", Version: "1.0", Desc: "git helper"},
		{Name: "gitg", Version: "0.1", Desc: "git gui"},
		{Name: "nano", Version: "7.2", Desc: "editor"},
	}
	var m Model = New(detect.Info{Kind: detect.APT}, &pkgmanager.Manager{})
	m = upd(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = upd(m, loadPkgsMsg{pkgs: pkgs, installed: 1})

	step := func(label string) {
		t.Logf("--- %s ---\n%s", label, dumpLines(m.View()))
	}
	key := func(r rune) {
		m = upd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(string(r))})
	}
	step("home")
	key('g')
	step("g")
	key('i')
	step("gi")
	key('t')
	step("git")
	m = upd(m, tea.KeyMsg{Type: tea.KeyDown})
	step("git+down")
	// down moved the cursor; left/right cycle the action, never the cursor.
	// gitsome is not installed, so Remove is unavailable and right advances
	// straight from Install to Info (skipping the disabled remove).
	m = upd(m, tea.KeyMsg{Type: tea.KeyRight})
	step("git+right (remove skipped)")
	if m.action != actInfo {
		t.Fatalf("after right action=%d, want actInfo (remove skipped)", m.action)
	}
	m = upd(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.action != actInstall {
		t.Fatalf("after right again action=%d, want actInstall", m.action)
	}
	m = upd(m, tea.KeyMsg{Type: tea.KeyLeft})
	m = upd(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.action != actInstall {
		t.Fatalf("after left x2 action=%d, want actInstall", m.action)
	}
	if got := m.results[m.cursor].Name; got != "gitsome" {
		t.Fatalf("cursor moved by left/right = %s, want gitsome", got)
	}
	m = upd(m, tea.KeyMsg{Type: tea.KeyEsc})
	step("esc->home")
	if m.searching {
		t.Fatal("esc should close the popup")
	}
	if len(m.results) != 0 {
		t.Fatal("results should be cleared on close")
	}
	key('z')
	step("z (no match)")
	if len(m.results) != 0 {
		t.Fatal("expected no matches for z")
	}
	if !m.searching {
		t.Fatal("expected popup open after typing z")
	}
	m = upd(m, tea.KeyMsg{Type: tea.KeyEsc})
	key('n')
	key('a')
	step("na -> nano")
	if m.cursor != 0 || len(m.results) != 1 || m.results[0].Name != "nano" {
		t.Fatalf("want only nano selected, got cursor=%d results=%+v", m.cursor, m.results)
	}

	// Single tap selects the row; a double tap runs the focused action
	// directly — no confirmation (confirmation is only tied to pill clicks).
	m = upd(m, tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 9, Y: 7})
	m = upd(m, tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 9, Y: 7})
	if got := m.results[0].Name; got != "nano" {
		t.Fatalf("cursor after tap = %s, want nano", got)
	}
	if m.confirm {
		t.Fatal("double-tap must not open a confirmation prompt (pill click only)")
	}
	if m.state != stateBusy && m.state != stateResult {
		t.Fatalf("double-tap should run the action directly, got state=%d", m.state)
	}

	// A successful operation must stay on the Result window until esc,
	// even though the installed-state refresh lands right after it.
	m = upd(m, tea.Msg(taskDoneMsg{})) // the earlier fake-manager run finished
	m = upd(m, loadPkgsMsg{pkgs: pkgs, installed: 2})
	if m.state != stateResult {
		t.Fatalf("result window should persist after refresh, got state=%d", m.state)
	}
	m = upd(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.state != stateBrowse {
		t.Fatalf("esc from result should reopen the search, got state=%d", m.state)
	}
}

func TestTapPillOpensConfirm(t *testing.T) {
	pkgs := []pkgmanager.Package{
		{Name: "nano", Version: "7.2", Desc: "editor"},
	}
	var m Model = New(detect.Info{Kind: detect.APT}, &pkgmanager.Manager{})
	m = upd(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = upd(m, loadPkgsMsg{pkgs: pkgs, installed: 0})
	key := func(r rune) {
		m = upd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(string(r))})
	}
	key('n')

	// A tap on the pill row. At 80x24 the popup pads ph=18, padY=3, so the
	// pill row is absolute row 13+4=17; the pills start at col padX+3=11.
	m = upd(m, tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 12, Y: 17})
	if !m.confirm || m.confirmPkg != "nano" || !m.confirmInstall {
		t.Fatalf("tapping Install pill should open the confirmation dialog, got confirm=%v pkg=%q install=%v",
			m.confirm, m.confirmPkg, m.confirmInstall)
	}
	if m.action != actInstall {
		t.Fatalf("tap should set action to install, got %d", m.action)
	}
	if v := m.View(); !strings.Contains(v, "Install nano?") || !strings.Contains(v, "cancel") {
		t.Fatalf("confirmation dialog should render the question and no-cancel:\n%s", dumpLines(v))
	}

	// 'n' cancels back to the popup.
	m = upd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if m.confirm {
		t.Fatal("n should cancel the confirmation")
	}

	// Tapping the Remove pill (unavailable: nano not installed) must be a
	// no-op — no prompt.
	m = upd(m, tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 21, Y: 17})
	if m.confirm {
		t.Fatal("tapping a disabled pill must not open a prompt")
	}
}

func upd(m Model, msg tea.Msg) Model {
	mm, _ := m.Update(msg)
	return mm.(Model)
}

func TestInstalledActionDisabled(t *testing.T) {
	pkgs := []pkgmanager.Package{
		{Name: "nano", Version: "7.2", Desc: "editor", Installed: true},
	}
	var m Model = New(detect.Info{Kind: detect.APT}, &pkgmanager.Manager{})
	m = upd(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = upd(m, loadPkgsMsg{pkgs: pkgs, installed: 1})
	key := func(r rune) {
		m = upd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(string(r))})
	}
	key('n')
	if m.action != actRemove {
		t.Fatalf("selecting an installed pkg should default the action to remove, got action=%d", m.action)
	}

	// Cycling must skip the disabled Install: remove -> info -> remove.
	m = upd(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.action != actInfo {
		t.Fatalf("after right on installed pkg action=%d, want actInfo (install skipped)", m.action)
	}
	m = upd(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.action != actRemove {
		t.Fatalf("after right x2 on installed pkg action=%d, want actRemove", m.action)
	}

	// The Install slot renders as a disabled "installed" tag, not a runnable
	// pill, and even a stale Install action must not trigger anything.
	m.action = actInstall
	v := m.View()
	if !strings.Contains(v, "installed") {
		t.Fatalf("expected an installed tag in the action bar:\n%s", v)
	}
	if strings.Contains(v, " Install ") {
		t.Fatalf("Install pill must not render for an installed package:\n%s", v)
	}
	m = upd(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateBrowse {
		t.Fatalf("Enter on a disabled install must be a no-op, got state=%d", m.state)
	}
}

func TestHeaderSingleLine(t *testing.T) {
	var m Model = New(detect.Info{Kind: detect.Pacman}, &pkgmanager.Manager{})
	m.installed = 117
	m.pkgs = make([]pkgmanager.Package, 5332)
	s := m.header()
	if strings.Contains(s, "\n") {
		t.Fatalf("header must be a single line, got:\n%s", s)
	}
	if strings.Contains(s, "╭") || strings.Contains(s, "╰") {
		t.Fatalf("header should not render border boxes, got: %q", s)
	}
	if !strings.Contains(s, "Pacman") || !strings.Contains(s, "117 / 5332 installed") {
		t.Fatalf("header missing manager/stats: %q", s)
	}

	apt := New(detect.Info{Kind: detect.APT}, &pkgmanager.Manager{})
	apt.installed = 4
	apt.pkgs = make([]pkgmanager.Package, 4944)
	if s := apt.header(); !strings.Contains(s, "APT") {
		t.Fatalf("apt header missing badge: %q", s)
	}
}

func TestInstalledRefreshAfterOp(t *testing.T) {
	pkgs := []pkgmanager.Package{{Name: "nano", Version: "7.2", Desc: "editor"}}
	var m Model = New(detect.Info{Kind: detect.APT}, &pkgmanager.Manager{})
	m = upd(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = upd(m, loadPkgsMsg{pkgs: pkgs, installed: 0})
	key := func(r rune) {
		m = upd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(string(r))})
	}
	key('n')
	if m.action != actInstall {
		t.Fatalf("action on uninstalled pkg=%d, want actInstall", m.action)
	}

	m.opPkg, m.opInstall = "nano", true
	m.applyOpResult()
	if !m.pkgs[0].Installed {
		t.Fatal("install should mark the package installed immediately")
	}
	if m.installed != 1 {
		t.Fatalf("installed count after install=%d, want 1", m.installed)
	}
	if m.action != actRemove {
		t.Fatalf("action after install=%d, want actRemove (resynced)", m.action)
	}
	if v := m.View(); !strings.Contains(v, "installed") {
		t.Fatalf("pills should show the installed tag after install:\n%s", v)
	}

	m.opPkg, m.opInstall = "nano", false
	m.applyOpResult()
	if m.pkgs[0].Installed {
		t.Fatal("remove should mark the package as not installed immediately")
	}
	if m.installed != 0 {
		t.Fatalf("installed count after remove=%d, want 0", m.installed)
	}
	// Remove just became invalid, so the action advances past it to Info.
	if m.action != actInfo {
		t.Fatalf("action after remove=%d, want actInfo (resynced)", m.action)
	}
}

func TestPacmanInfoDisabledForUninstalled(t *testing.T) {
	pkgs := []pkgmanager.Package{{Name: "nano", Version: "7.2", Desc: "editor"}}
	var m Model = New(detect.Info{Kind: detect.Pacman}, &pkgmanager.Manager{})
	m = upd(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = upd(m, loadPkgsMsg{pkgs: pkgs, installed: 0})
	key := func(r rune) {
		m = upd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(string(r))})
	}
	key('n')
	if !m.actionDisabled(pkgs[0], actInfo) {
		t.Fatal("pacman info must be disabled for a package that is not installed")
	}
	if m.action != actInstall {
		t.Fatalf("action on uninstalled pacman pkg=%d, want actInstall", m.action)
	}
	// Only Install is valid here: remove and info are disabled, so cycling
	// must wrap straight back to Install.
	m = upd(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.action != actInstall {
		t.Fatalf("cycle on uninstalled pacman pkg=%d, want actInstall (others disabled)", m.action)
	}
	// Enter must never run a disabled action (no busy state, no install).
	if cmd := m.runDirect(); cmd != nil {
		t.Fatal("runDirect on all-disabled selection should be a no-op")
	}
}

func TestParseProgress(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"preparing", 0},
		{"(Reading database ... 42%", 42},
		{"almost 101%", 100},
	}
	for _, c := range cases {
		if got := parseProgress(c.in); got != c.want {
			t.Fatalf("parseProgress(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestCleanAptWarnings(t *testing.T) {
	in := "WARNING: apt does not have a stable CLI interface. Use with caution in scripts.\nReading package lists...\n"
	if got := cleanAptWarnings(in); strings.Contains(got, "stable CLI") {
		t.Fatalf("apt warning not cleaned: %q", got)
	}
}

func TestBusyView(t *testing.T) {
	var m Model = New(detect.Info{Kind: detect.APT}, &pkgmanager.Manager{})
	m = upd(m, tea.WindowSizeMsg{Width: 60, Height: 24})
	m.state = stateBusy
	m.task = "Installing nano"
	m.ticks = 3
	m.progress = 58
	m.taskStart = time.Now().Add(-90 * time.Second)

	v := m.View()
	for _, want := range []string{"Installing nano", "58% complete"} {
		if !strings.Contains(v, want) {
			t.Fatalf("busy view missing %q:\n%s", want, v)
		}
	}

	m.ticks = 90
	m.progress = 0
	if v := m.View(); !strings.Contains(v, "working ·") {
		t.Fatalf("indeterminate busy view missing elapsed marker:\n%s", v)
	}

	m.showLog = true
	if v := m.View(); !strings.Contains(v, "Installing nano") {
		t.Fatalf("log view missing task:\n%s", v)
	}
}

func dumpLines(s string) string {
	var sb strings.Builder
	for _, ln := range strings.Split(s, "\n") {
		if strings.TrimSpace(ln) != "" {
			sb.WriteString("|" + ln + "\n")
		}
	}
	return sb.String()
}
