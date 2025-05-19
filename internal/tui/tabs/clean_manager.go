package tabs

import (
	"github.com/pashkov256/deletor/internal/tui/interfaces"
)

// CleanTabManager manages the tabs for the clean view
type CleanTabManager struct {
	model     interfaces.CleanModel
	tabs      []Tab
	activeTab int
}

// NewCleanTabManager creates a new CleanTabManager
func NewCleanTabManager(model interfaces.CleanModel, factory *CleanTabFactory) *CleanTabManager {
	return &CleanTabManager{
		model:     model,
		tabs:      factory.CreateTabs(model),
		activeTab: 0,
	}
}

// GetActiveTab returns the currently active tab
func (m *CleanTabManager) GetActiveTab() Tab {
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
