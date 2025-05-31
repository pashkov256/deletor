package clean

import (
	"fmt"

	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/tabs/base"
)

// CleanTabManager manages the tabs for the clean view
type CleanTabManager struct {
	model     interfaces.CleanModel
	tabs      []base.Tab
	activeTab int
}

// NewCleanTabManager creates a new CleanTabManager
func NewCleanTabManager(model interfaces.CleanModel, factory *CleanTabFactory) *CleanTabManager {
	// Create tabs
	tabs := factory.CreateTabs(model)

	// Initialize each tab
	for _, tab := range tabs {
		if err := tab.Init(); err != nil {
			fmt.Printf("Error initializing tab: %v\n", err)
		}
	}

	return &CleanTabManager{
		model:     model,
		tabs:      tabs,
		activeTab: 0,
	}
}

// GetActiveTab returns the currently active tab
func (m *CleanTabManager) GetActiveTab() base.Tab {
	return m.tabs[m.activeTab]
}

// GetActiveTabIndex returns the index of the currently active tab
func (m *CleanTabManager) GetActiveTabIndex() int {
	return m.activeTab
}

// SetActiveTabIndex sets the active tab index
func (m *CleanTabManager) SetActiveTabIndex(index int) {
	if index >= 0 && index < len(m.tabs) {
		m.activeTab = index
	}
}

// GetAllTabs returns all tabs
func (m *CleanTabManager) GetAllTabs() []base.Tab {
	return m.tabs
}
