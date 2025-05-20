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
	if t.model.GetFocusedElement() == "excludeInput" {
		excludeStyle = styles.StandardInputFocusedStyle
	}
	excludeInput := t.model.GetExcludeInput()
	excludeInput.Placeholder = "specific files/paths (e.g. data,backup)"
	content.WriteString(excludeStyle.Render("Exclude: " + excludeInput.View()))
	content.WriteString("\n")

	minSizeStyle := styles.StandardInputStyle
	maxSizeStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "minSize" {
		minSizeStyle = styles.StandardInputFocusedStyle
	}
	if t.model.GetFocusedElement() == "maxSize" {
		maxSizeStyle = styles.StandardInputFocusedStyle
	}

	minSizeInput := t.model.GetMinSizeInput()
	maxSizeInput := t.model.GetMaxSizeInput()

	content.WriteString(minSizeStyle.Render("Min size: " + minSizeInput.View()))
	content.WriteString("\n")
	content.WriteString(maxSizeStyle.Render("Max size: " + maxSizeInput.View()))
	return content.String()
}
