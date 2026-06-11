// Package tui is an interactive repository list built on Bubble Tea.
// Search: "/" key, select: Enter, quit: Esc or Ctrl+C.
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/wiremeusd/gitty/internal/github"
)

type FetchFunc func() ([]github.Repo, error)

type reposMsg []github.Repo

type fetchErrMsg struct{ err error }

type item struct{ repo github.Repo }

func (i item) Title() string { return i.repo.FullName }

func (i item) Description() string {
	lang := i.repo.Language
	if lang == "" {
		lang = "—"
	}
	return fmt.Sprintf("★ %d · %s", i.repo.Stars, lang)
}

func (i item) FilterValue() string { return i.repo.FullName }

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Model struct {
	list     list.Model
	fetch    FetchFunc
	selected *github.Repo
	offline  bool
	title    string
}

// New creates the model: initial is the cached list (may be empty),
// fetch is the background refresh function (nil if data is already fresh).
func New(account string, initial []github.Repo, fetch FetchFunc) Model {
	l := list.New(toItems(initial), list.NewDefaultDelegate(), 0, 0)
	title := "gitty · " + account
	l.Title = title
	// Our wrapper controls quitting (Esc/Ctrl+C): disable the list's built-in
	// quit keybindings, otherwise "q" silently exits without a selection.
	l.DisableQuitKeybindings()
	return Model{list: l, fetch: fetch, title: title}
}

func toItems(repos []github.Repo) []list.Item {
	items := make([]list.Item, len(repos))
	for i, r := range repos {
		items[i] = item{repo: r}
	}
	return items
}

// Selected returns the chosen repository; nil if the user exited without selecting.
func (m Model) Selected() *github.Repo { return m.selected }

func (m Model) Offline() bool { return m.offline }

func (m Model) Init() tea.Cmd {
	if m.fetch == nil {
		return nil
	}
	fetch := m.fetch
	return func() tea.Msg {
		repos, err := fetch()
		if err != nil {
			return fetchErrMsg{err}
		}
		return reposMsg(repos)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case reposMsg:
		m.offline = false
		m.list.Title = m.title
		return m, m.list.SetItems(toItems(msg))
	case fetchErrMsg:
		m.offline = true
		m.list.Title = m.title + " · offline, showing cached data"
		return m, nil
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			// While a filter is being typed, Enter confirms the filter — delegate to the list.
			if m.list.FilterState() != list.Filtering {
				if it, ok := m.list.SelectedItem().(item); ok {
					m.selected = &it.repo
					return m, tea.Quit
				}
			}
		case "esc":
			// Esc with an active filter clears the filter (handled by the list);
			// without a filter — quit.
			if m.list.FilterState() == list.Unfiltered {
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return docStyle.Render(m.list.View())
}
