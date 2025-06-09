package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/tui/help"
	"github.com/pashkov256/deletor/internal/tui/menu"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

var (
	docStyle = lipgloss.NewStyle().
			Padding(1, 1).
			Align(lipgloss.Center)

	buttonStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2).
			Foreground(lipgloss.Color("#FFFFFF"))

	selectedButtonStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("#1E90FF"))
)

type MainMenu struct {
	SelectedIndex int
}

func NewMainMenu() *MainMenu {
	return &MainMenu{
		SelectedIndex: 0,
	}
}

func (m *MainMenu) Init() tea.Cmd {
	return nil
}

func (m *MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			return m.handleTab()
		case "shift+tab":
			return m.handleShiftTab()
		case "enter":
			return m, func() tea.Msg {
				return tea.KeyMsg{
					Type: tea.KeyEnter,
				}
			}
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			// Check each menu item for click
			for i := 0; i < 5; i++ {
				if zone.Get(fmt.Sprintf("menu_button_%d", i)).InBounds(msg) {
					m.SelectedIndex = i
					// Emulate Enter key press
					return m, func() tea.Msg {
						return tea.KeyMsg{
							Type: tea.KeyEnter,
						}
					}
				}
			}
		}
	}

	return m, nil
}

func (m *MainMenu) View() string {
	var content strings.Builder

	// Title
	content.WriteString(styles.TitleStyle.Render("🗑️  Deletor v1.4.0"))
	content.WriteString("\n\n")

	// Menu items from constants
	items := []string{
		menu.CleanFIlesTitle,
		menu.CleanCacheTitle,
		menu.ManageRulesTitle,
		menu.StatisticsTitle,
		menu.ExitTitle,
	}

	// Render buttons
	for i, item := range items {
		style := buttonStyle
		if i == m.SelectedIndex {
			style = selectedButtonStyle
		}

		button := style.Render(item)
		content.WriteString(zone.Mark(fmt.Sprintf("menu_button_%d", i), button))
		content.WriteString("\n")
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(help.NavigateHelpText)

	return zone.Scan(styles.AppStyle.Render(docStyle.Render(content.String())))
}

// handleTab moves the cursor down in the list
func (m *MainMenu) handleTab() (tea.Model, tea.Cmd) {
	m.SelectedIndex = (m.SelectedIndex + 1) % 5
	return m, nil
}

// handleShiftTab moves the cursor up in the list
func (m *MainMenu) handleShiftTab() (tea.Model, tea.Cmd) {
	m.SelectedIndex = (m.SelectedIndex - 1 + 5) % 5
	return m, nil
}
