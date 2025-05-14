package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/models"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

type FiltersTab struct {
	model *models.CleanFilesModel
}

func NewFiltersTab(model *models.CleanFilesModel) *FiltersTab {
	return &FiltersTab{
		model: model,
	}
}

func (t *FiltersTab) View() string {
	var content strings.Builder

	excludeStyle := styles.StandardInputStyle
	if t.model.FocusedElement == "exclude" {
		excludeStyle = styles.StandardInputFocusedStyle
	}
	t.model.ExcludeInput.Placeholder = "specific files/paths (e.g. data,backup)"
	content.WriteString(excludeStyle.Render("Exclude: " + t.model.ExcludeInput.View()))
	content.WriteString("\n")
	sizeStyle := styles.StandardInputStyle
	if t.model.FocusedElement == "size" {
		sizeStyle = styles.StandardInputFocusedStyle
	}
	content.WriteString(sizeStyle.Render("Min size: " + t.model.SizeInput.View()))

	return content.String()
}

func (t *FiltersTab) Init() tea.Cmd {
	return nil
}

func (t *FiltersTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}
