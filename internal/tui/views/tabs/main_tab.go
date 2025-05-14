package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/views"
)

type MainTab struct {
	model *views.CleanFilesModel
}

func NewMainTab(model *views.CleanFilesModel) *MainTab {
	return &MainTab{
		model: model,
	}
}

func (t *MainTab) View() string {

	return ""
}

func (t *MainTab) Init() tea.Cmd {
	return nil
}

func (t *MainTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}
