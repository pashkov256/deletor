package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle = lipgloss.NewStyle().
			Margin(2).
			Padding(1, 2).
			Align(lipgloss.Center)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			Align(lipgloss.Center)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				Padding(0, 1)
)

type item struct {
	title string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

type MainMenu struct {
	list list.Model
}

func NewMainMenu() *MainMenu {
	items := []list.Item{
		item{title: "🧹 Clean Files"},
		item{title: "⚙️ Manage Rules"},
		item{title: "📊 Statistics"},
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
		list: l,
	}
}

func (m *MainMenu) Init() tea.Cmd {
	return nil
}

func (m *MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width-2, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MainMenu) View() string {
	return appStyle.Render(m.list.View())
}
