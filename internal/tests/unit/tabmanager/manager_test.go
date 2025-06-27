package tabmanager

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/tabs/base"
	"github.com/stretchr/testify/assert"
)

// mockTab реализует интерфейс base.Tab
type mockTab struct {
	id string
}

func (m mockTab) View() string {
	return "view of " + m.id
}

func (m mockTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (m mockTab) Init() tea.Cmd {
	return nil
}

func TestTabManagerInitialization(t *testing.T) {
	tabs := []base.Tab{
		mockTab{id: "tab1"},
		mockTab{id: "tab2"},
	}
	model := "mockModel"

	manager := base.NewTabManager(tabs, &model)

	assert.NotNil(t, manager)
	assert.Equal(t, 0, manager.GetActiveTabIndex())
	assert.Equal(t, tabs[0], manager.GetActiveTab())
}

func TestGetActiveTab(t *testing.T) {
	tabs := []base.Tab{
		mockTab{id: "A"},
		mockTab{id: "B"},
	}
	manager := base.NewTabManager(tabs, new(string))

	assert.Equal(t, tabs[0], manager.GetActiveTab())

	manager.SetActiveTabIndex(1)
	assert.Equal(t, tabs[1], manager.GetActiveTab())
}

func TestGetActiveTabIndex(t *testing.T) {
	tabs := []base.Tab{
		mockTab{id: "One"},
		mockTab{id: "Two"},
	}
	manager := base.NewTabManager(tabs, new(string))

	assert.Equal(t, 0, manager.GetActiveTabIndex())

	manager.SetActiveTabIndex(1)
	assert.Equal(t, 1, manager.GetActiveTabIndex())
}

func TestSetActiveTabIndex(t *testing.T) {
	tabs := []base.Tab{
		mockTab{id: "X"},
		mockTab{id: "Y"},
		mockTab{id: "Z"},
	}
	manager := base.NewTabManager(tabs, new(string))

	manager.SetActiveTabIndex(2)
	assert.Equal(t, 2, manager.GetActiveTabIndex())
	assert.Equal(t, tabs[2], manager.GetActiveTab())
}
