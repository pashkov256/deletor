package views_test

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/views"
)

func setupMenuTestModel() *views.MainMenu {
	rm := newTestingRulesModel()
	return views.NewMainMenu(rm)
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
		key           string
		initialIndex  int
		expectedIndex int
	}{
		{
			name:          "Tab key navigation",
			key:           "tab",
			initialIndex:  0,
			expectedIndex: 1,
		},
		{
			name:          "Shift+Tab key navigation",
			key:           "shift+tab",
			initialIndex:  1,
			expectedIndex: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := views.NewMainMenu(&testingRulesModel{})
			model.SelectedIndex = tt.initialIndex

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, _ := model.Update(msg)
			if m, ok := newModel.(*views.MainMenu); ok {
				if m.SelectedIndex != tt.expectedIndex {
					t.Errorf("Model state after update does not match expected state for test case: %s\nGot: %d, Expected: %d",
						tt.name, m.SelectedIndex, tt.expectedIndex)
				}
			} else {
				t.Errorf("Failed to convert model to MainMenu")
			}
		})
	}
}

func TestMainMenu_ListNavigation(t *testing.T) {
	model := setupMenuTestModel()

	if model.SelectedIndex != 0 {
		t.Error("Initial list index should be 0")
	}

	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedMenu := updatedModel.(*views.MainMenu)
	if model.SelectedIndex != 1 {
		t.Error("Tab key should move cursor down")
	}

	updatedModel, _ = updatedMenu.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	// nolint:staticcheck
	updatedMenu = updatedModel.(*views.MainMenu)
	if model.SelectedIndex != 0 {
		t.Error("Shift+Tab key should move cursor up")
	}
}

// mock model that minimially implements interfaces.RulesModel
func newTestingRulesModel() *testingRulesModel {
	return &testingRulesModel{
		options: map[string]bool{
			"DisableEmoji": false,
		},
	}
}

type testingRulesModel struct {
	options map[string]bool
}

func (t *testingRulesModel) GetOptionState() map[string]bool {
	return t.options
}

func (t *testingRulesModel) GetPathInput() textinput.Model    { return textinput.Model{} }
func (t *testingRulesModel) GetExtInput() textinput.Model     { return textinput.Model{} }
func (t *testingRulesModel) GetMinSizeInput() textinput.Model { return textinput.Model{} }
func (t *testingRulesModel) GetMaxSizeInput() textinput.Model { return textinput.Model{} }
func (t *testingRulesModel) GetExcludeInput() textinput.Model { return textinput.Model{} }
func (t *testingRulesModel) GetOlderInput() textinput.Model   { return textinput.Model{} }
func (t *testingRulesModel) GetNewerInput() textinput.Model   { return textinput.Model{} }
func (t *testingRulesModel) GetFocusedElement() string        { return "" }
func (t *testingRulesModel) GetRulesPath() string             { return "" }

func (t *testingRulesModel) SetFocusedElement(string)    {}
func (t *testingRulesModel) SetOptionState(string, bool) {}

func (t *testingRulesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}
