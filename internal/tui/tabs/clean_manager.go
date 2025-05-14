package tabs

import (
	"github.com/pashkov256/deletor/internal/tui/interfaces"
)

// CleanModel defines the interface that models must implement to work with clean tabs
type CleanModel interface {
	GetCurrentPath() string
	GetExtensions() []string
	GetMinSize() int64
	GetExclude() []string
	GetOptions() []string
	GetOptionState() map[string]bool
	GetFocusedElement() string
	GetShowDirs() bool
	GetDirSize() int64
	GetCalculatingSize() bool
	GetFilteredSize() int64
	GetFilteredCount() int
	GetActiveTab() int
}

type CleanTabManager struct {
	*TabManager[interfaces.CleanModel]
	factory TabFactory
}

func NewCleanTabManager(model interfaces.CleanModel, factory TabFactory) *CleanTabManager {
	cleanTabs := []Tab{
		factory.NewMainTab(model),
		factory.NewFiltersTab(model),
		factory.NewOptionsTab(model),
		factory.NewHelpTab(model),
	}

	return &CleanTabManager{
		TabManager: NewTabManager(cleanTabs, &model),
		factory:    factory,
	}
}
