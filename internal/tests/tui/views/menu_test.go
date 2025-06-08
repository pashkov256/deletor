package views_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/views"
)

func setupMenuTestModel() *views.MainMenu {
	return views.NewMainMenu()
}

func TestMainMenu_Init(t *testing.T) {
	model := setupMenuTestModel()
	cmd := model.Init()

	if cmd != nil {
		t.Error("Init() should return nil command")
	}
}

func TestMainMenu_Update(t *testing.T) {
	tests := []struct {
		name          string
		msg           tea.Msg
		expectedState func(*views.MainMenu) bool
	}{
		{
			name: "Tab key navigation",
			msg:  tea.KeyMsg{Type: tea.KeyTab},
			expectedState: func(m *views.MainMenu) bool {
				return m.List.Index() == 1
			},
		},
		{
			name: "Shift+Tab key navigation",
			msg:  tea.KeyMsg{Type: tea.KeyShiftTab},
			expectedState: func(m *views.MainMenu) bool {
				return m.List.Index() == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := setupMenuTestModel()
			updatedModel, _ := model.Update(tt.msg)
			updatedMenu := updatedModel.(*views.MainMenu)

			if !tt.expectedState(updatedMenu) {
				t.Errorf("Model state after update does not match expected state for test case: %s", tt.name)
			}
		})
	}
}

func TestMainMenu_ListNavigation(t *testing.T) {
	model := setupMenuTestModel()

	if model.List.Index() != 0 {
		t.Error("Initial list index should be 0")
	}

	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedMenu := updatedModel.(*views.MainMenu)
	if updatedMenu.List.Index() != 1 {
		t.Error("Tab key should move cursor down")
	}

	updatedModel, _ = updatedMenu.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	updatedMenu = updatedModel.(*views.MainMenu)
	if updatedMenu.List.Index() != 0 {
		t.Error("Shift+Tab key should move cursor up")
	}
}
