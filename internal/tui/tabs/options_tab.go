package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

// Define options in fixed order
var options = []string{
	"Show hidden files",
	"Confirm deletion",
	"Include subfolders",
	"Delete empty subfolders",
}

type OptionsTab struct {
	model interfaces.CleanModel
}

func (t *OptionsTab) View() string {
	var content strings.Builder
	optionState := t.model.GetOptionState()

	for i, name := range options {
		style := styles.OptionStyle
		if optionState[name] {
			style = styles.SelectedOptionStyle
		}
		if t.model.GetFocusedElement() == fmt.Sprintf("option%d", i+1) {
			style = styles.OptionFocusedStyle
		}
		content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", i+1)))
		content.WriteString(style.Render(fmt.Sprintf("[%s] %-20s", map[bool]string{true: "✓", false: "○"}[optionState[name]], name)))
		content.WriteString("\n")
	}
	return content.String()

}

func (t *OptionsTab) Init() tea.Cmd {
	return nil
}

func (t *OptionsTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}
