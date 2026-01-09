package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	rules "github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/help"
	"github.com/pashkov256/deletor/internal/tui/menu"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/utils"
)

type MainMenu struct {
	SelectedIndex int
	OptionState   map[string]bool
}

func NewMainMenu(rules rules.Rules) *MainMenu {
	latestRules, _ := rules.GetRules()
	return &MainMenu{
		SelectedIndex: 0,
		OptionState: map[string]bool{
			options.ShowHiddenFiles:       latestRules.ShowHiddenFiles,
			options.ConfirmDeletion:       latestRules.ConfirmDeletion,
			options.IncludeSubfolders:     latestRules.IncludeSubfolders,
			options.DeleteEmptySubfolders: latestRules.DeleteEmptySubfolders,
			options.SendFilesToTrash:      latestRules.SendFilesToTrash,
			options.LogOperations:         latestRules.LogOperations,
			options.LogToFile:             latestRules.LogToFile,
			options.ShowStatistics:        latestRules.ShowStatistics,
			options.DisableEmoji:          latestRules.DisableEmoji,
			options.ExitAfterDeletion:     latestRules.ExitAfterDeletion,
		},
	}
}

func (m *MainMenu) Init() tea.Cmd {
	return nil
}

func (m *MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			return m.HandleFocusBottom()
		case "shift+tab", "up":
			return m.HandleFocusTop()
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

	disableEmoji := m.GetOptionState()[options.DisableEmoji]

	// Title
	title := "🗑️ Deletor v1.5.0"
	if disableEmoji {
		newTitle, err := utils.RemoveEmoji(title)
		if err == nil {
			title = newTitle
		}
	}
	content.WriteString(styles.TitleStyle.Render(title))
	content.WriteString("\n\n")

	// Menu items from constants

	// Render buttons
	for i, item := range menu.MenuItems {
		style := styles.MenuItem
		if i == m.SelectedIndex {
			style = styles.SelectedMenuItemStyle
		}

		if disableEmoji { // removing emoji if disabled
			newItem, err := utils.RemoveEmoji(item)
			if err == nil {
				item = newItem
			}
		}

		button := style.Render(item)
		content.WriteString(zone.Mark(fmt.Sprintf("menu_button_%d", i), button))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(help.NavigateHelpText)

	return zone.Scan(styles.AppStyle.Render(styles.DocStyle.Render(content.String())))
}

// HandleFocusBottom moves focus down
func (m *MainMenu) HandleFocusBottom() (tea.Model, tea.Cmd) {
	if m.SelectedIndex < len(menu.MenuItems)-1 {
		m.SelectedIndex++
	} else {
		m.SelectedIndex = 0
	}
	return m, nil
}

// HandleFocusTop moves focus up
func (m *MainMenu) HandleFocusTop() (tea.Model, tea.Cmd) {
	if m.SelectedIndex > 0 {
		m.SelectedIndex--
	} else {
		m.SelectedIndex = len(menu.MenuItems) - 1
	}
	return m, nil
}

func (m *MainMenu) GetOptionState() map[string]bool {
	return m.OptionState
}
