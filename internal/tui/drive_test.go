package tui

import (
	"strings"
	"testing"

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
	m = upd(m, tea.KeyMsg{Type: tea.KeyRight})
	step("git+right(page)")
	m = upd(m, tea.KeyMsg{Type: tea.KeyLeft})
	step("git+left")
	// after down: cursor 1; right(clamps to last): gitg; left(clamps to 0): git
	if got := m.results[m.cursor].Name; got != "git" {
		t.Fatalf("after arrows cursor = %s, want git", got)
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

	// Single tap selects the row; a double tap triggers the focused action.
	m = upd(m, tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 9, Y: 7})
	m = upd(m, tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 9, Y: 7})
	if got := m.results[0].Name; got != "nano" {
		t.Fatalf("cursor after tap = %s, want nano", got)
	}
	if m.state != stateBusy && m.state != stateResult {
		t.Fatalf("double-tap should start the action, got state=%d", m.state)
	}
}

func upd(m Model, msg tea.Msg) Model {
	mm, _ := m.Update(msg)
	return mm.(Model)
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
