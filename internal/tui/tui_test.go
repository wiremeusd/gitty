package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/wiremeusd/gitty/internal/github"
)

func repos(names ...string) []github.Repo {
	var rs []github.Repo
	for _, n := range names {
		rs = append(rs, github.Repo{FullName: "me/" + n, Name: n})
	}
	return rs
}

func TestNewShowsInitialRepos(t *testing.T) {
	m := New("me", repos("one", "two"), nil)
	if got := len(m.list.Items()); got != 2 {
		t.Fatalf("items = %d, want 2", got)
	}
}

func TestReposMsgReplacesItems(t *testing.T) {
	m := New("me", repos("old"), nil)
	updated, _ := m.Update(reposMsg(repos("a", "b", "c")))
	m = updated.(Model)
	if got := len(m.list.Items()); got != 3 {
		t.Fatalf("items = %d, want 3", got)
	}
	if m.Offline() {
		t.Fatal("fresh data must clear offline flag")
	}
}

func TestFetchErrorSetsOfflineFlag(t *testing.T) {
	m := New("me", repos("cached"), nil)
	updated, _ := m.Update(fetchErrMsg{errors.New("no network")})
	m = updated.(Model)
	if !m.Offline() {
		t.Fatal("expected offline flag")
	}
	if got := len(m.list.Items()); got != 1 {
		t.Fatal("cached items must survive fetch error")
	}
}

func TestEnterSelectsRepo(t *testing.T) {
	m := New("me", repos("one", "two"), nil)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	sel := m.Selected()
	if sel == nil || sel.FullName != "me/one" {
		t.Fatalf("selected = %+v", sel)
	}
	if cmd == nil {
		t.Fatal("enter must produce a quit command")
	}
}

func TestInitWithFetchProducesReposMsg(t *testing.T) {
	fetch := func() ([]github.Repo, error) { return repos("fresh"), nil }
	m := New("me", nil, fetch)
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
	msg := cmd()
	rs, ok := msg.(reposMsg)
	if !ok || len(rs) != 1 {
		t.Fatalf("msg = %#v", msg)
	}
}

func TestInitWithoutFetchIsNil(t *testing.T) {
	m := New("me", repos("one"), nil)
	if m.Init() != nil {
		t.Fatal("nil fetch must produce nil Init cmd")
	}
}

func TestEnterSelectsFromFilteredList(t *testing.T) {
	m := New("me", repos("alpha", "beta", "alphabet"), nil)
	m.list.SetFilterText("beta")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	sel := m.Selected()
	if sel == nil || sel.FullName != "me/beta" {
		t.Fatalf("selected = %+v, want me/beta", sel)
	}
}

func TestQKeyDoesNotQuit(t *testing.T) {
	m := New("me", repos("one"), nil)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)
	if m.Selected() != nil {
		t.Fatal("q must not select")
	}
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, isQuit := msg.(tea.QuitMsg); isQuit {
				t.Fatal("q must not quit the program")
			}
		}
	}
}
