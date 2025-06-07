package views_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/views"
	"github.com/pashkov256/deletor/internal/validation"
)

// setupTestModel creates a new RulesModel with test configuration
func setupTestModel() *views.RulesModel {
	rulesInstance := rules.NewRules()
	validator := validation.NewValidator()
	return views.NewRulesModel(rulesInstance, validator)
}

func TestRulesModel_Init(t *testing.T) {
	model := setupTestModel()
	cmd := model.Init()

	// Test that Init returns a command
	if cmd == nil {
		t.Error("Init() should return a non-nil command")
	}

	// Test that TabManager is initialized
	if model.TabManager == nil {
		t.Error("TabManager should be initialized after Init()")
	}
}

func TestRulesModel_View(t *testing.T) {
	tests := []struct {
		name           string
		setupModel     func(*views.RulesModel)
		expectedOutput string
	}{
		{
			name: "Main tab view",
			setupModel: func(m *views.RulesModel) {
				m.Init()
				m.TabManager.SetActiveTabIndex(0)
				m.FocusedElement = "locationInput"
			},
			expectedOutput: "🗂️ [F1] Main",
		},
		{
			name: "Filters tab view",
			setupModel: func(m *views.RulesModel) {
				m.Init()
				m.TabManager.SetActiveTabIndex(1)
				m.FocusedElement = "extensionsInput"
			},
			expectedOutput: "🧹 [F2] Filters",
		},
		{
			name: "Options tab view",
			setupModel: func(m *views.RulesModel) {
				m.Init()
				m.TabManager.SetActiveTabIndex(2)
				m.FocusedElement = "option1"
			},
			expectedOutput: "⚙️ [F3] Options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := setupTestModel()
			tt.setupModel(model)
			view := model.View()
			if !strings.Contains(view, tt.expectedOutput) {
				t.Errorf("View() output does not contain expected text: %s", tt.expectedOutput)
			}
		})
	}
}

func TestRulesModel_Update(t *testing.T) {
	tests := []struct {
		name          string
		msg           tea.Msg
		setupModel    func(*views.RulesModel)
		expectedState func(*views.RulesModel) bool
	}{
		{
			name: "Update location input",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'e', 's', 't'}},
			setupModel: func(m *views.RulesModel) {
				m.Init()
				m.TabManager.SetActiveTabIndex(0)
				m.FocusedElement = "locationInput"
				m.LocationInput.Focus()
			},
			expectedState: func(m *views.RulesModel) bool {
				return m.LocationInput.Value() == "test"
			},
		},
		{
			name: "Update extensions input",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.', 't', 'x', 't'}},
			setupModel: func(m *views.RulesModel) {
				m.Init()
				m.TabManager.SetActiveTabIndex(1)
				m.FocusedElement = "extensionsInput"
				m.ExtensionsInput.Focus()
			},
			expectedState: func(m *views.RulesModel) bool {
				return m.ExtensionsInput.Value() == ".txt"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := setupTestModel()
			tt.setupModel(model)

			// Send each rune individually to simulate typing
			for _, r := range tt.msg.(tea.KeyMsg).Runes {
				updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
				model = updatedModel.(*views.RulesModel)
			}

			if !tt.expectedState(model) {
				t.Error("Model state after update does not match expected state")
			}
		})
	}
}

func TestRulesModel_TabNavigation(t *testing.T) {
	tests := []struct {
		name          string
		key           tea.KeyType
		initialTab    int
		expectedTab   int
		expectedFocus string
	}{
		{
			name:          "Tab key navigation",
			key:           tea.KeyTab,
			initialTab:    0,
			expectedTab:   0,
			expectedFocus: "saveButton",
		},
		{
			name:          "Right arrow navigation",
			key:           tea.KeyRight,
			initialTab:    0,
			expectedTab:   1,
			expectedFocus: "extensionsInput",
		},
		{
			name:          "Left arrow navigation",
			key:           tea.KeyLeft,
			initialTab:    1,
			expectedTab:   0,
			expectedFocus: "locationInput",
		},
		{
			name:          "F1 key navigation",
			key:           tea.KeyF1,
			initialTab:    2,
			expectedTab:   0,
			expectedFocus: "locationInput",
		},
		{
			name:          "F2 key navigation",
			key:           tea.KeyF2,
			initialTab:    0,
			expectedTab:   1,
			expectedFocus: "extensionsInput",
		},
		{
			name:          "F3 key navigation",
			key:           tea.KeyF3,
			initialTab:    0,
			expectedTab:   2,
			expectedFocus: "option1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := setupTestModel()
			model.Init()
			model.TabManager.SetActiveTabIndex(tt.initialTab)

			// Set initial focus based on tab
			switch tt.initialTab {
			case 0:
				model.FocusedElement = "locationInput"
				model.LocationInput.Focus()
			case 1:
				model.FocusedElement = "extensionsInput"
				model.ExtensionsInput.Focus()
			case 2:
				model.FocusedElement = "option1"
			}

			// Handle the key press
			updatedModel, _ := model.Update(tea.KeyMsg{Type: tt.key})
			updatedRulesModel := updatedModel.(*views.RulesModel)

			if updatedRulesModel.TabManager.GetActiveTabIndex() != tt.expectedTab {
				t.Errorf("Expected tab %d, got %d", tt.expectedTab, updatedRulesModel.TabManager.GetActiveTabIndex())
			}

			if updatedRulesModel.FocusedElement != tt.expectedFocus {
				t.Errorf("Expected focus %s, got %s", tt.expectedFocus, updatedRulesModel.FocusedElement)
			}
		})
	}
}

func TestRulesModel_OptionToggling(t *testing.T) {
	tests := []struct {
		name          string
		optionKey     string
		initialState  bool
		expectedState bool
	}{
		{
			name:          "Toggle ShowHiddenFiles",
			optionKey:     options.ShowHiddenFiles,
			initialState:  false,
			expectedState: true,
		},
		{
			name:          "Toggle ConfirmDeletion",
			optionKey:     options.ConfirmDeletion,
			initialState:  true,
			expectedState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := setupTestModel()
			model.Init()
			model.TabManager.SetActiveTabIndex(2)
			model.OptionState[tt.optionKey] = tt.initialState

			// Set focus to the option
			optionIndex := -1
			for i, name := range options.DefaultCleanOption {
				if name == tt.optionKey {
					optionIndex = i + 1
					break
				}
			}
			if optionIndex == -1 {
				t.Fatalf("Option %s not found in DefaultCleanOption", tt.optionKey)
			}
			model.FocusedElement = fmt.Sprintf("option%d", optionIndex)

			// Simulate space key press
			updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
			updatedRulesModel := updatedModel.(*views.RulesModel)

			if updatedRulesModel.OptionState[tt.optionKey] != tt.expectedState {
				t.Errorf("Option %s state = %v, want %v", tt.optionKey,
					updatedRulesModel.OptionState[tt.optionKey], tt.expectedState)
			}
		})
	}
}

func TestRulesModel_Validation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		setupModel    func(*views.RulesModel)
		expectedError bool
	}{
		{
			name: "Valid inputs",
			setupModel: func(m *views.RulesModel) {
				m.LocationInput.SetValue(tempDir)
				m.MinSizeInput.SetValue("10mb")
				m.MaxSizeInput.SetValue("1gb")
				m.OlderInput.SetValue("1day")
				m.NewerInput.SetValue("1hour")
			},
			expectedError: false,
		},
		{
			name: "Invalid min size",
			setupModel: func(m *views.RulesModel) {
				m.LocationInput.SetValue(tempDir)
				m.MinSizeInput.SetValue("invalid")
			},
			expectedError: true,
		},
		{
			name: "Invalid max size",
			setupModel: func(m *views.RulesModel) {
				m.LocationInput.SetValue(tempDir)
				m.MaxSizeInput.SetValue("invalid")
			},
			expectedError: true,
		},
		{
			name: "Invalid older time",
			setupModel: func(m *views.RulesModel) {
				m.LocationInput.SetValue(tempDir)
				m.OlderInput.SetValue("invalid")
			},
			expectedError: true,
		},
		{
			name: "Invalid newer time",
			setupModel: func(m *views.RulesModel) {
				m.LocationInput.SetValue(tempDir)
				m.NewerInput.SetValue("invalid")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := setupTestModel()
			model.Init()
			tt.setupModel(model)

			err := model.ValidateInputs()
			if (err != nil) != tt.expectedError {
				t.Errorf("ValidateInputs() error = %v, wantErr %v", err, tt.expectedError)
			}
		})
	}
}
