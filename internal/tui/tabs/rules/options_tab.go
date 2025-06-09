package rules

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

type OptionsTab struct {
	model interfaces.RulesModel
}

func (t *OptionsTab) Init() tea.Cmd              { return nil }
func (t *OptionsTab) Update(msg tea.Msg) tea.Cmd { return nil }

func (t *OptionsTab) View() string {
	var content strings.Builder

	for i, name := range options.DefaultCleanOption {
		style := styles.OptionStyle
		if t.model.GetFocusedElement() == fmt.Sprintf("rules_option_%d", i+1) {
			style = styles.OptionFocusedStyle
		} else {
			if t.model.GetOptionState()[name] {
				style = styles.SelectedOptionStyle
			}
		}

		emoji := options.GetEmojiByCleanOption(name)

		content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", i+1)))
		content.WriteString(zone.Mark(fmt.Sprintf("rules_option_%d", i+1), style.Render(fmt.Sprintf("[%s] %s %-20s",
			map[bool]string{true: "✓", false: "○"}[t.model.GetOptionState()[name]],
			emoji, name))))
		content.WriteString("\n")
	}

	return content.String()
}
