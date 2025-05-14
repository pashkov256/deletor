package tabs

import (
	"github.com/pashkov256/deletor/internal/tui/models"
)

type CleanTabManager struct {
	*TabManager[models.CleanFilesModel]
	factory TabFactory
}

func NewCleanTabManager(model *models.CleanFilesModel, factory TabFactory) *CleanTabManager {
	cleanTabs := []Tab{
		factory.NewMainTab(model),
		factory.NewFiltersTab(model),
		factory.NewOptionsTab(model),
		factory.NewHelpTab(model),
	}

	return &CleanTabManager{
		TabManager: NewTabManager[models.CleanFilesModel](cleanTabs, model),
		factory:    factory,
	}
}
