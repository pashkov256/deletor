package clean

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

// Define options in fixed order

type OptionsTab struct {
	model interfaces.CleanModel
}

func (t *OptionsTab) View() string {
	var content strings.Builder

	for optionIndex, name := range options.DefaultCleanOption {
		style := styles.OptionStyle
		if t.model.GetOptionState()[name] {
			style = styles.SelectedOptionStyle
		}
		if t.model.GetFocusedElement() == fmt.Sprintf("option%d", optionIndex+1) {
			style = styles.OptionFocusedStyle
		}
		content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", optionIndex+1)))

		// Add emojis based on option name
		emoji := options.GetEmojiByCleanOption(name)

		content.WriteString(style.Render(fmt.Sprintf("[%s] %s %-20s", map[bool]string{true: "✓", false: "○"}[t.model.GetOptionState()[name]], emoji, name)))
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
