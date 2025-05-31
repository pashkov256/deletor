package rules

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

	// Get all inputs
	excludeInput := t.model.GetExcludeInput()
	minSizeInput := t.model.GetMinSizeInput()
	maxSizeInput := t.model.GetMaxSizeInput()
	olderInput := t.model.GetOlderInput()
	newerInput := t.model.GetNewerInput()
	extInput := t.model.GetExtInput()

	// Set placeholders
	excludeInput.Placeholder = "specific files/paths (e.g. data,backup)"
	olderInput.Placeholder = "e.g. 60 min, 1 hour, 7 days, 1 month"
	newerInput.Placeholder = "e.g. 60 min, 1 hour, 7 days, 1 month"
	minSizeInput.Placeholder = "e.g. 10b,10kb,10mb,10gb,10tb"
	maxSizeInput.Placeholder = "e.g. 10b,10kb,10mb,10gb,10tb"
	extInput.Placeholder = "e.g. js,png,zip"

	// Render inputs with appropriate styles
	inputs := []struct {
		name  string
		input textinput.Model
	}{
		{"excludeInput", excludeInput},
		{"minSizeInput", minSizeInput},
		{"maxSizeInput", maxSizeInput},
		{"olderInput", olderInput},
		{"newerInput", newerInput},
		{"extInput", extInput},
	}

	for _, input := range inputs {
		style := styles.StandardInputStyle
		if t.model.GetFocusedElement() == input.name {
			style = styles.StandardInputFocusedStyle
		}
		content.WriteString(style.Render(input.name + ": " + input.input.View()))
		content.WriteString("\n")
	}

	return content.String()
}
