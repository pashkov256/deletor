package rules

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/tui/help"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

type MainTab struct {
	model interfaces.RulesModel
}

func (t *MainTab) Init() tea.Cmd              { return nil }
func (t *MainTab) Update(msg tea.Msg) tea.Cmd { return nil }

func (t *MainTab) View() string {
	var content strings.Builder

	// Location input with label
	pathStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "locationInput" {
		pathStyle = styles.StandardInputFocusedStyle
	}
	inputContent := pathStyle.Render("Path: " + t.model.GetPathInput().View())
	content.WriteString(zone.Mark("rules_location_input", inputContent))
	content.WriteString("\n\n")

	// Save button
	saveButtonStyle := styles.StandardButtonStyle
	if t.model.GetFocusedElement() == "saveButton" {
		saveButtonStyle = styles.StandardButtonFocusedStyle
	}
	buttonContent := saveButtonStyle.Render("💾 Save rules")
	content.WriteString(zone.Mark("rules_save_button", buttonContent))
	content.WriteString("\n\n\n")

	content.WriteString(styles.PathStyle.Render(fmt.Sprintf("Rules are stored in: %s", t.model.GetRulesPath())))
	content.WriteString("\n\n" + help.NavigateHelpText)

	return content.String()
}
