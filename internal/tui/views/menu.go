package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

var (
	docStyle = lipgloss.NewStyle().
		// Margin(1).
		Padding(1, 1).
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
		Item{title: "🗑️ Clear cache"},
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
		Padding(0, 1).MarginTop(1)

	return &MainMenu{
		List: l,
	}
}

func (m *MainMenu) Init() tea.Cmd {
	return nil
}

func (m *MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.List.SetSize(msg.Width-4, msg.Height-6)

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			return m.handleTab()
		case "shift+tab":
			return m.handleShiftTab()
		}
	}

	// Pass other messages to the list model
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *MainMenu) View() string {
	var content strings.Builder

	content.WriteString(docStyle.Render(m.List.View()))

	content.WriteString(styles.AppStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		content.String(),
		"⬇/⬆: navigate in list • Tab: cycle focus • Shift+Tab: focus back • Enter: select/confirm • Esc: back to list",
	)))

	return content.String()
}

// handleTab moves the cursor down in the list
func (m *MainMenu) handleTab() (tea.Model, tea.Cmd) {
	m.List.CursorDown()
	return m, nil
}

// handleShiftTab moves the cursor up in the list
func (m *MainMenu) handleShiftTab() (tea.Model, tea.Cmd) {
	m.List.CursorUp()
	return m, nil
}
