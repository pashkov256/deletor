package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/models"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

type OptionsTab struct {
	model *models.CleanFilesModel
}

func NewOptionsTab(model *models.CleanFilesModel) *OptionsTab {
	return &OptionsTab{
		model: model,
	}
}

func (t *OptionsTab) View() string {

	var content strings.Builder
	for i, name := range t.model.Options {
		style := styles.OptionStyle
		if t.model.OptionState[name] {
			style = styles.SelectedOptionStyle
		}
		if t.model.FocusedElement == fmt.Sprintf("option%d", i+1) {
			style = styles.OptionFocusedStyle
		}
		content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", i+1)))
		content.WriteString(style.Render(fmt.Sprintf("[%s] %-20s", map[bool]string{true: "✓", false: "○"}[t.model.OptionState[name]], name)))
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
