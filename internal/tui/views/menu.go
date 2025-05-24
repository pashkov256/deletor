package views

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle = lipgloss.NewStyle().
		Margin(1).
		Padding(1, 2).
		Align(lipgloss.Center)
)

type Item struct {
	title string
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return "" }
func (i Item) FilterValue() string { return i.title }

type MainMenu struct {
	List list.Model
}

func NewMainMenu() *MainMenu {
	items := []list.Item{
		Item{title: "🧹 Clean files"},
		Item{title: "🗑️ Clear system cache"},
		Item{title: "⚙️ Manage rules"},
		Item{title: "📊 Statistics"},
		Item{title: "🚪 Exit"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	delegate.SetSpacing(0)

	l := list.New(items, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.Title = "🗑️ Deletor"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#1E90FF")).
		Padding(0, 1)

	return &MainMenu{
		List: l,
	}
}

func (m *MainMenu) Init() tea.Cmd {
	return nil
}

func (m *MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.List.SetSize(msg.Width-4, msg.Height-6)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *MainMenu) View() string {
	return docStyle.Render(m.List.View())
}
