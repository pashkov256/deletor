package rules

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

type FiltersTab struct {
	model interfaces.RulesModel
}

func (t *FiltersTab) Init() tea.Cmd              { return nil }
func (t *FiltersTab) Update(msg tea.Msg) tea.Cmd { return nil }

func (t *FiltersTab) View() string {
	var content strings.Builder

	inputs := []struct {
		name  string
		input textinput.Model
		key   string
	}{
		{"Extensions", t.model.GetExtInput(), "extensionsInput"},
		{"Min Size", t.model.GetMinSizeInput(), "minSizeInput"},
		{"Max Size", t.model.GetMaxSizeInput(), "maxSizeInput"},
		{"Exclude", t.model.GetExcludeInput(), "excludeInput"},
		{"Older Than", t.model.GetOlderInput(), "olderInput"},
		{"Newer Than", t.model.GetNewerInput(), "newerInput"},
	}

	for _, input := range inputs {
		style := styles.StandardInputStyle
		if t.model.GetFocusedElement() == input.key {
			style = styles.StandardInputFocusedStyle
		}
		content.WriteString(zone.Mark(fmt.Sprintf("rules_%s", input.key), style.Render(input.name+": "+input.input.View())))
		content.WriteString("\n")
	}

	return content.String()
}
