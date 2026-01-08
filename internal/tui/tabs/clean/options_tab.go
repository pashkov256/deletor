package clean

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
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

	for i, name := range options.DefaultCleanOption {
		optionStyle := styles.OptionStyle
		if t.model.GetFocusedElement() == fmt.Sprintf("clean_option_%d", i+1) {
			optionStyle = styles.OptionFocusedStyle
		} else {
			if t.model.GetOptionState()[name] {
				optionStyle = styles.SelectedOptionStyle
			}
		}

		emoji := ""
		if !t.model.GetOptionState()[options.DisableEmoji] { // Selects an emoji if not disabled
			emoji = options.GetEmojiByCleanOption(name)
		}

		content.WriteString(zone.Mark(fmt.Sprintf("clean_option_%d", i+1), optionStyle.Render(fmt.Sprintf("[%s] %s %-20s", map[bool]string{true: "✓", false: "○"}[t.model.GetOptionState()[name]], emoji, name))))

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
