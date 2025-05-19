package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

type FiltersTab struct {
	model interfaces.CleanModel
}

func (t *FiltersTab) Init() tea.Cmd              { return nil }
func (t *FiltersTab) Update(msg tea.Msg) tea.Cmd { return nil }

func (t *FiltersTab) View() string {
	var content strings.Builder
	excludeStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "exclude" {
		excludeStyle = styles.StandardInputFocusedStyle
	}
	excludeInput := t.model.GetExcludeInput()
	excludeInput.Placeholder = "specific files/paths (e.g. data,backup)"
	content.WriteString(excludeStyle.Render("Exclude: " + excludeInput.View()))
	content.WriteString("\n")
	sizeStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "size" {
		sizeStyle = styles.StandardInputFocusedStyle
	}
	sizeInput := t.model.GetSizeInput()
	content.WriteString(sizeStyle.Render("Min size: " + sizeInput.View()))
	return content.String()
}
