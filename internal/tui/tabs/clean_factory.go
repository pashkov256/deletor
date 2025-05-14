package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
)

type cleanTabFactory struct {
	styles TabStyles
}

func NewCleanTabFactory() TabFactory {
	return &cleanTabFactory{
		styles: TabStyles{
			TabStyle:       lipgloss.NewStyle().Padding(0, 1),
			ActiveTabStyle: lipgloss.NewStyle().Padding(0, 1).Bold(true),
		},
	}
}

func (f *cleanTabFactory) NewMainTab(model interface{}) Tab {
	return &mainTab{
		model:  model.(interfaces.CleanModel),
		styles: f.styles,
	}
}

func (f *cleanTabFactory) NewFiltersTab(model interface{}) Tab {
	return &filtersTab{
		model:  model.(interfaces.CleanModel),
		styles: f.styles,
	}
}

func (f *cleanTabFactory) NewOptionsTab(model interface{}) Tab {
	return &optionsTab{
		model:  model.(interfaces.CleanModel),
		styles: f.styles,
	}
}

func (f *cleanTabFactory) NewHelpTab(model interface{}) Tab {
	return &helpTab{
		model:  model.(interfaces.CleanModel),
		styles: f.styles,
	}
}

type mainTab struct {
	model  interfaces.CleanModel
	styles TabStyles
}

func (t *mainTab) View() string {
	return "Main Tab"
}

func (t *mainTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (t *mainTab) Init() tea.Cmd {
	return nil
}

type filtersTab struct {
	model  interfaces.CleanModel
	styles TabStyles
}

func (t *filtersTab) View() string {
	return "Filters Tab"
}

func (t *filtersTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (t *filtersTab) Init() tea.Cmd {
	return nil
}

type optionsTab struct {
	model  interfaces.CleanModel
	styles TabStyles
}

func (t *optionsTab) View() string {
	return "Options Tab"
}

func (t *optionsTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (t *optionsTab) Init() tea.Cmd {
	return nil
}

type helpTab struct {
	model  interfaces.CleanModel
	styles TabStyles
}

func (t *helpTab) View() string {
	return "Help Tab"
}

func (t *helpTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (t *helpTab) Init() tea.Cmd {
	return nil
}
