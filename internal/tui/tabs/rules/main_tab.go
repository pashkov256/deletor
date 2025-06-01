package rules

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

	pathStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "pathInput" {
		pathStyle = styles.StandardInputFocusedStyle
	}

	content.WriteString(pathStyle.Render("Path: " + t.model.GetPathInput().View()))
	content.WriteString("\n\n")

	saveButtonStyle := styles.StandardButtonStyle
	if t.model.GetFocusedElement() == "saveButton" {
		saveButtonStyle = styles.StandardButtonFocusedStyle
	}

	content.WriteString(saveButtonStyle.Render("💾 Save rules"))

	return content.String()
}
