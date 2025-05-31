package clean

import (
	"fmt"

	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/tabs/base"
)

type CleanTabFactory struct{}

func NewCleanTabFactory() *CleanTabFactory {
	return &CleanTabFactory{}
}

func (f *CleanTabFactory) CreateTabs(model interfaces.CleanModel) []base.Tab {
	// Create tabs
	tabs := []base.Tab{
		&MainTab{model: model},
		&FiltersTab{model: model},
		&OptionsTab{model: model},
		&LogTab{model: model},
		&HelpTab{model: model},
	}

	// Initialize each tab
	for _, tab := range tabs {
		if err := tab.Init(); err != nil {
			fmt.Printf("Error initializing tab: %v\n", err)
		}
	}

	return tabs
}
