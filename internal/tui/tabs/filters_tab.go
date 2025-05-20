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

	excludeInput := t.model.GetExcludeInput()
	minSizeInput := t.model.GetMinSizeInput()
	maxSizeInput := t.model.GetMaxSizeInput()
	olderInput := t.model.GetOlderInput()
	newerInput := t.model.GetNewerInput()

	excludeStyle := styles.StandardInputStyle
	minSizeStyle := styles.StandardInputStyle
	maxSizeStyle := styles.StandardInputStyle
	olderStyle := styles.StandardInputStyle
	newerStyle := styles.StandardInputStyle

	excludeInput.Placeholder = "specific files/paths (e.g. data,backup)"
	olderInput.Placeholder = "e.g. 60 min, 1 hour, 7 days, 1 month"
	newerInput.Placeholder = "e.g. 60 min, 1 hour, 7 days, 1 month"

	switch t.model.GetFocusedElement() {
	case "excludeInput":
		excludeStyle = styles.StandardInputFocusedStyle
	case "minSizeInput":
		minSizeStyle = styles.StandardInputFocusedStyle
	case "maxSizeInput":
		maxSizeStyle = styles.StandardInputFocusedStyle
	case "olderInput":
		olderStyle = styles.StandardInputFocusedStyle
	case "newerInput":
		newerStyle = styles.StandardInputFocusedStyle
	}

	content.WriteString(excludeStyle.Render("Exclude: " + excludeInput.View()))
	content.WriteString("\n")
	content.WriteString(minSizeStyle.Render("Min size: " + minSizeInput.View()))
	content.WriteString("\n")
	content.WriteString(maxSizeStyle.Render("Max size: " + maxSizeInput.View()))
	content.WriteString("\n")
	content.WriteString(olderStyle.Render("Modified more: " + olderInput.View()))
	content.WriteString("\n")
	content.WriteString(newerStyle.Render("Modified less: " + newerInput.View()))
	content.WriteString("\n\n")
	content.WriteString("Press on any input in focus to update the file list, or CRTL+R")
	return content.String()
}
