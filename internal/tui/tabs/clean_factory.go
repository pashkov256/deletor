package tabs

import (
	"github.com/pashkov256/deletor/internal/tui/interfaces"
)

type CleanTabFactory struct{}

func NewCleanTabFactory() *CleanTabFactory {
	return &CleanTabFactory{}
}

func (f *CleanTabFactory) CreateTabs(model interfaces.CleanModel) []Tab {
	return []Tab{
		&MainTab{model: model},
		&FiltersTab{model: model},
		&OptionsTab{model: model},
		&HelpTab{model: model},
	}
}
