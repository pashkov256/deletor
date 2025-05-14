package tabs

import (
	"github.com/pashkov256/deletor/internal/tui/components/tabs"
	"github.com/pashkov256/deletor/internal/tui/models"
)

type TabFactory struct{}

func (f *TabFactory) NewMainTab(model interface{}) tabs.Tab {
	return NewMainTab(model.(*models.CleanFilesModel))
}

func (f *TabFactory) NewFiltersTab(model interface{}) tabs.Tab {
	return NewFiltersTab(model.(*models.CleanFilesModel))
}

func (f *TabFactory) NewOptionsTab(model interface{}) tabs.Tab {
	return NewOptionsTab(model.(*models.CleanFilesModel))
}

func (f *TabFactory) NewHelpTab(model interface{}) tabs.Tab {
	return NewHelpTab(model.(*models.CleanFilesModel))
}
