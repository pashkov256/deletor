package clean

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
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

	// Exclude patterns
	excludeStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "excludeInput" {
		excludeStyle = styles.StandardInputFocusedStyle
	}
	content.WriteString(zone.Mark("filters_exclude_input", excludeStyle.Render("Exclude: "+t.model.GetExcludeInput().View())))
	content.WriteString("\n")

	// Size filters
	minSizeStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "minSizeInput" {
		minSizeStyle = styles.StandardInputFocusedStyle
	}
	content.WriteString(zone.Mark("filters_min_size_input", minSizeStyle.Render("Min size: "+t.model.GetMinSizeInput().View())))
	content.WriteString("\n")

	maxSizeStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "maxSizeInput" {
		maxSizeStyle = styles.StandardInputFocusedStyle
	}
	content.WriteString(zone.Mark("filters_max_size_input", maxSizeStyle.Render("Max size: "+t.model.GetMaxSizeInput().View())))
	content.WriteString("\n")

	// Date filters
	olderStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "olderInput" {
		olderStyle = styles.StandardInputFocusedStyle
	}
	content.WriteString(zone.Mark("filters_older_input", olderStyle.Render("Older than: "+t.model.GetOlderInput().View())))
	content.WriteString("\n")

	newerStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "newerInput" {
		newerStyle = styles.StandardInputFocusedStyle
	}
	content.WriteString(zone.Mark("filters_newer_input", newerStyle.Render("Newer than: "+t.model.GetNewerInput().View())))

	return content.String()
}
