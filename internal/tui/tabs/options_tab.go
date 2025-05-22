package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

// Define options in fixed order

type OptionsTab struct {
	model interfaces.CleanModel
}

var DefaultOptionState = map[string]bool{
	"Show hidden files":       false,
	"Confirm deletion":        false,
	"Include subfolders":      false,
	"Delete empty subfolders": false,
	"Send files to trash":     false,
}

var DefaultOption = []string{
	"Show hidden files",
	"Confirm deletion",
	"Include subfolders",
	"Delete empty subfolders",
	"Send files to trash",
}

func (t *OptionsTab) View() string {
	var content strings.Builder

	for optionIndex, name := range DefaultOption {

		style := styles.OptionStyle
		if DefaultOptionState[name] {
			style = styles.SelectedOptionStyle
		}
		if t.model.GetFocusedElement() == fmt.Sprintf("option%d", optionIndex+1) {
			style = styles.OptionFocusedStyle
		}
		content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", optionIndex+1)))
		content.WriteString(style.Render(fmt.Sprintf("[%s] %-20s", map[bool]string{true: "✓", false: "○"}[DefaultOptionState[name]], name)))
		content.WriteString("\n")
		optionIndex++
	}
	return content.String()

}

func (t *OptionsTab) Init() tea.Cmd {
	return nil
}

func (t *OptionsTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}
