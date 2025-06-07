package views_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/models"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/errors"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/views"
	"github.com/pashkov256/deletor/internal/validation"
)

func setupCleanTestModel(t *testing.T) *views.CleanFilesModel {
	tempDir := t.TempDir()

	rulesObj := rules.NewRules()
	_ = rulesObj.SetupRulesConfig()
	fileManager := filemanager.NewFileManager()
	validator := validation.NewValidator()

	model := views.InitialCleanModel(rulesObj, fileManager, validator)
	if model == nil {
		t.Fatal("Failed to create model")
	}

	model.CurrentPath = tempDir
	model.PathInput.SetValue(tempDir)
	model.IsLaunched = true

	model.OptionState = map[string]bool{
		options.ShowHiddenFiles:       false,
		options.ConfirmDeletion:       false,
		options.IncludeSubfolders:     false,
		options.DeleteEmptySubfolders: false,
		options.SendFilesToTrash:      false,
		options.LogOperations:         false,
		options.LogToFile:             false,
		options.ShowStatistics:        false,
		options.ExitAfterDeletion:     false,
	}

	return model
}

func TestCleanFilesModel_Init(t *testing.T) {
	model := setupCleanTestModel(t)
	if model == nil {
		t.Fatal("Failed to setup test model")
	}

	cmd := model.Init()

	if cmd == nil {
		t.Error("Init() should return a non-nil command")
	}

	if model.TabManager == nil {
		t.Error("TabManager should be initialized after Init()")
	}

	// Test initial focus
	if model.FocusedElement != "pathInput" {
		t.Errorf("Expected initial focus to be 'pathInput', got '%s'", model.FocusedElement)
	}

	// Test that path input is focused
	if !model.PathInput.Focused() {
		t.Error("Path input should be focused after Init()")
	}
}

func TestCleanFilesModel_InitialState(t *testing.T) {
	model := setupCleanTestModel(t)

	// Test initial values
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
		compare  func(a, b interface{}) bool
	}{
		{"CurrentPath", model.CurrentPath, model.PathInput.Value(), func(a, b interface{}) bool {
			// Compare only the base names since the full paths will be different
			return filepath.Base(a.(string)) == filepath.Base(b.(string))
		}},
		{"Extensions", model.Extensions, []string{}, func(a, b interface{}) bool { return compareSlices(a.([]string), b.([]string)) }},
		{"MinSize", model.MinSize, int64(0), func(a, b interface{}) bool { return a == b }},
		{"MaxSize", model.MaxSize, int64(0), func(a, b interface{}) bool { return a == b }},
		{"Exclude", model.Exclude, []string{}, func(a, b interface{}) bool { return compareSlices(a.([]string), b.([]string)) }},
		{"ShowDirs", model.ShowDirs, false, func(a, b interface{}) bool { return a == b }},
		{"CalculatingSize", model.CalculatingSize, false, func(a, b interface{}) bool { return a == b }},
		{"FilteredSize", model.FilteredSize, int64(0), func(a, b interface{}) bool { return a == b }},
		{"FilteredCount", model.FilteredCount, 0, func(a, b interface{}) bool { return a == b }},
		{"IsLaunched", model.IsLaunched, true, func(a, b interface{}) bool { return a == b }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.compare(tt.got, tt.expected) {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	expectedOptions := map[string]bool{
		options.ShowHiddenFiles:       false,
		options.ConfirmDeletion:       false,
		options.IncludeSubfolders:     false,
		options.DeleteEmptySubfolders: false,
		options.SendFilesToTrash:      false,
		options.LogOperations:         false,
		options.LogToFile:             false,
		options.ShowStatistics:        false,
		options.ExitAfterDeletion:     false,
	}

	for opt, expected := range expectedOptions {
		if got := model.OptionState[opt]; got != expected {
			t.Errorf("OptionState[%s] = %v, want %v", opt, got, expected)
		}
	}
}

func TestCleanFilesModel_InputInitialization(t *testing.T) {
	model := setupCleanTestModel(t)

	// Test input placeholders
	tests := []struct {
		name     string
		input    textinput.Model
		expected string
	}{
		{"ExtInput", model.ExtInput, "e.g. js,png,zip"},
		{"MinSizeInput", model.MinSizeInput, "e.g. 10b,10kb,10mb,10gb,10tb"},
		{"MaxSizeInput", model.MaxSizeInput, "e.g. 10b,10kb,10mb,10gb,10tb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.Placeholder != tt.expected {
				t.Errorf("%s placeholder = %q, want %q", tt.name, tt.input.Placeholder, tt.expected)
			}
		})
	}

	// Test input initialization
	inputs := []struct {
		name        string
		input       textinput.Model
		prompt      string
		placeholder string
	}{
		{"PathInput", model.PathInput, "Path", ""},
		{"ExtInput", model.ExtInput, "Extensions", "e.g. js,png,zip"},
		{"MinSizeInput", model.MinSizeInput, "Min Size", "e.g. 10b,10kb,10mb,10gb,10tb"},
		{"MaxSizeInput", model.MaxSizeInput, "Max Size", "e.g. 10b,10kb,10mb,10gb,10tb"},
		{"ExcludeInput", model.ExcludeInput, "Exclude", ""},
		{"OlderInput", model.OlderInput, "Older Than", ""},
		{"NewerInput", model.NewerInput, "Newer Than", ""},
	}

	for _, tt := range inputs {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.Prompt == "" {
				t.Errorf("%s has empty prompt", tt.name)
			}
			if tt.input.Placeholder != tt.placeholder {
				t.Errorf("%s placeholder = %q, want %q", tt.name, tt.input.Placeholder, tt.placeholder)
			}
		})
	}
}

func TestCleanFilesModel_Navigation(t *testing.T) {
	model := setupCleanTestModel(t)
	model.Init()

	// Test F1-F5 tab switching
	t.Run("Tab Switching", func(t *testing.T) {
		tabSwitchTests := []struct {
			key      string
			expected int // tab index
		}{
			{"f1", 0}, // Main tab
			{"f2", 1}, // Filters tab
			{"f3", 2}, // Options tab
			{"f4", 3}, // Log tab
			{"f5", 4}, // Help tab
		}

		for _, tt := range tabSwitchTests {
			t.Run(tt.key, func(t *testing.T) {
				msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
				newModel, _ := model.Handle(msg)
				if m, ok := newModel.(*views.CleanFilesModel); ok {
					model = m
					if model.TabManager.GetActiveTabIndex() != tt.expected {
						t.Errorf("After %s, expected tab index %d, got %d", tt.key, tt.expected, model.TabManager.GetActiveTabIndex())
					}
				} else {
					t.Errorf("Failed to convert model to CleanFilesModel")
				}
			})
		}
	})

	t.Run("Element Navigation", func(t *testing.T) {
		// Test Main tab navigation
		t.Run("Main Tab", func(t *testing.T) {
			// Start at pathInput
			model.FocusedElement = "pathInput"
			model.TabManager.SetActiveTabIndex(0)

			// Test Tab navigation
			tabTests := []struct {
				key      string
				expected string
			}{
				{"tab", "extInput"},
				{"tab", "list"},
				{"tab", "dirButton"},
				{"tab", "deleteButton"},
				{"tab", "pathInput"},
			}

			for _, tt := range tabTests {
				t.Run("Tab "+tt.key, func(t *testing.T) {
					msg := tea.KeyMsg{Type: tea.KeyTab}
					newModel, _ := model.Handle(msg)
					if m, ok := newModel.(*views.CleanFilesModel); ok {
						model = m
						if model.FocusedElement != tt.expected {
							t.Errorf("After Tab, expected focus %s, got %s", tt.expected, model.FocusedElement)
						}
					} else {
						t.Errorf("Failed to convert model to CleanFilesModel")
					}
				})
			}

			// Test Shift+Tab navigation
			shiftTabTests := []struct {
				key      string
				expected string
			}{
				{"shift+tab", "deleteButton"},
				{"shift+tab", "dirButton"},
				{"shift+tab", "list"},
				{"shift+tab", "extInput"},
				{"shift+tab", "pathInput"},
			}

			for _, tt := range shiftTabTests {
				t.Run("Shift+Tab "+tt.key, func(t *testing.T) {
					msg := tea.KeyMsg{Type: tea.KeyShiftTab}
					newModel, _ := model.Handle(msg)
					if m, ok := newModel.(*views.CleanFilesModel); ok {
						model = m
						if model.FocusedElement != tt.expected {
							t.Errorf("After Shift+Tab, expected focus %s, got %s", tt.expected, model.FocusedElement)
						}
					} else {
						t.Errorf("Failed to convert model to CleanFilesModel")
					}
				})
			}
		})

		// Test Filters tab navigation
		t.Run("Filters Tab", func(t *testing.T) {
			model.FocusedElement = "excludeInput"
			model.TabManager.SetActiveTabIndex(1)

			// Test Tab navigation
			tabTests := []struct {
				key      string
				expected string
			}{
				{"tab", "minSizeInput"},
				{"tab", "maxSizeInput"},
				{"tab", "olderInput"},
				{"tab", "newerInput"},
				{"tab", "excludeInput"},
			}

			for _, tt := range tabTests {
				t.Run("Tab "+tt.key, func(t *testing.T) {
					msg := tea.KeyMsg{Type: tea.KeyTab}
					newModel, _ := model.Handle(msg)
					if m, ok := newModel.(*views.CleanFilesModel); ok {
						model = m
						if model.FocusedElement != tt.expected {
							t.Errorf("After Tab, expected focus %s, got %s", tt.expected, model.FocusedElement)
						}
					} else {
						t.Errorf("Failed to convert model to CleanFilesModel")
					}
				})
			}

			shiftTabTests := []struct {
				key      string
				expected string
			}{
				{"shift+tab", "newerInput"},
				{"shift+tab", "olderInput"},
				{"shift+tab", "maxSizeInput"},
				{"shift+tab", "minSizeInput"},
				{"shift+tab", "excludeInput"},
			}

			for _, tt := range shiftTabTests {
				t.Run("Shift+Tab "+tt.key, func(t *testing.T) {
					msg := tea.KeyMsg{Type: tea.KeyShiftTab}
					newModel, _ := model.Handle(msg)
					if m, ok := newModel.(*views.CleanFilesModel); ok {
						model = m
						if model.FocusedElement != tt.expected {
							t.Errorf("After Shift+Tab, expected focus %s, got %s", tt.expected, model.FocusedElement)
						}
					} else {
						t.Errorf("Failed to convert model to CleanFilesModel")
					}
				})
			}
		})

		t.Run("Options Tab", func(t *testing.T) {
			model.FocusedElement = "option1"
			model.TabManager.SetActiveTabIndex(2)

			// Test Tab navigation through options
			for i := 1; i <= len(options.DefaultCleanOption); i++ {
				t.Run(fmt.Sprintf("Tab to option%d", i+1), func(t *testing.T) {
					msg := tea.KeyMsg{Type: tea.KeyTab}
					newModel, _ := model.Handle(msg)
					if m, ok := newModel.(*views.CleanFilesModel); ok {
						model = m
						expected := fmt.Sprintf("option%d", i+1)
						if i == len(options.DefaultCleanOption) {
							expected = "option1"
						}
						if model.FocusedElement != expected {
							t.Errorf("After Tab, expected focus %s, got %s", expected, model.FocusedElement)
						}
					} else {
						t.Errorf("Failed to convert model to CleanFilesModel")
					}
				})
			}
		})
	})

	t.Run("List Navigation", func(t *testing.T) {
		// Set up test items
		items := []list.Item{
			models.CleanItem{Path: "item1", Size: 100},
			models.CleanItem{Path: "item2", Size: 200},
			models.CleanItem{Path: "item3", Size: 300},
		}
		model.List.SetItems(items)

		tests := []struct {
			key      string
			expected int //  selected index
		}{
			{"down", 1}, // Move down to second item
			{"down", 2}, // Move down to third item
			{"up", 1},   // Move up to second item
			{"up", 0},   // Move up to first item
		}

		for _, tt := range tests {
			t.Run(tt.key, func(t *testing.T) {
				msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
				newModel, _ := model.Handle(msg)
				if m, ok := newModel.(*views.CleanFilesModel); ok {
					model = m
					if model.List.Index() != tt.expected {
						t.Errorf("After %s, expected index %d, got %d", tt.key, tt.expected, model.List.Index())
					}
				} else {
					t.Errorf("Failed to convert model to CleanFilesModel")
				}
			})
		}
	})

	t.Run("Element Focusing", func(t *testing.T) {
		tests := []struct {
			element  string
			expected bool
		}{
			{"pathInput", true},
			{"extInput", true},
			{"minSizeInput", true},
			{"maxSizeInput", true},
			{"excludeInput", true},
			{"olderInput", true},
			{"newerInput", true},
		}

		for _, tt := range tests {
			t.Run(tt.element, func(t *testing.T) {
				// First blur all inputs
				model.PathInput.Blur()
				model.ExtInput.Blur()
				model.MinSizeInput.Blur()
				model.MaxSizeInput.Blur()
				model.ExcludeInput.Blur()
				model.OlderInput.Blur()
				model.NewerInput.Blur()

				// Set focused element and focus the corresponding input
				model.FocusedElement = tt.element
				switch tt.element {
				case "pathInput":
					model.PathInput.Focus()
				case "extInput":
					model.ExtInput.Focus()
				case "minSizeInput":
					model.MinSizeInput.Focus()
				case "maxSizeInput":
					model.MaxSizeInput.Focus()
				case "excludeInput":
					model.ExcludeInput.Focus()
				case "olderInput":
					model.OlderInput.Focus()
				case "newerInput":
					model.NewerInput.Focus()
				}

				var input textinput.Model
				switch tt.element {
				case "pathInput":
					input = model.PathInput
				case "extInput":
					input = model.ExtInput
				case "minSizeInput":
					input = model.MinSizeInput
				case "maxSizeInput":
					input = model.MaxSizeInput
				case "excludeInput":
					input = model.ExcludeInput
				case "olderInput":
					input = model.OlderInput
				case "newerInput":
					input = model.NewerInput
				}
				if input.Focused() != tt.expected {
					t.Errorf("Expected %s focus to be %v, got %v", tt.element, tt.expected, input.Focused())
				}
			})
		}
	})
}

func TestCleanFilesModel_FileOperations(t *testing.T) {
	t.Run("Load Files", func(t *testing.T) {
		model := setupCleanTestModel(t)
		model.Init()

		tempDir := model.CurrentPath
		testFiles := []struct {
			name    string
			content string
			size    int64
		}{
			{"test1.txt", "test content 1", 13},
			{"test2.txt", "test content 2", 13},
			{"test3.log", "test log content", 15},
		}

		for _, file := range testFiles {
			filePath := filepath.Join(tempDir, file.name)
			if err := os.WriteFile(filePath, []byte(file.content), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", file.name, err)
			}
		}

		cmd := model.LoadFiles()
		msg := cmd()
		if err, ok := msg.(*errors.Error); ok {
			t.Fatalf("Failed to load files: %v", err)
		}
		items, ok := msg.([]list.Item)
		if !ok {
			t.Fatalf("LoadFiles() did not return []list.Item")
		}
		model.List.SetItems(items)

		// Verify files are loaded (excluding parent directory entry)
		fileCount := 0
		for _, item := range items {
			cleanItem := item.(models.CleanItem)
			if cleanItem.Size != -1 { // Skip parent directory entry
				fileCount++
			}
		}
		if fileCount != len(testFiles) {
			t.Errorf("Expected %d files, got %d", len(testFiles), fileCount)
		}
	})

	t.Run("Load Directories", func(t *testing.T) {
		model := setupCleanTestModel(t)
		model.Init()

		tempDir := model.CurrentPath
		subDirs := []string{"subdir1", "subdir2", "subdir3"}
		for _, dir := range subDirs {
			dirPath := filepath.Join(tempDir, dir)
			if err := os.Mkdir(dirPath, 0755); err != nil {
				t.Fatalf("Failed to create test directory %s: %v", dir, err)
			}
		}

		cmd := model.LoadDirs()
		msg := cmd()
		if err, ok := msg.(*errors.Error); ok {
			t.Fatalf("Failed to load directories: %v", err)
		}
		items, ok := msg.([]list.Item)
		if !ok {
			t.Fatalf("LoadDirs() did not return []list.Item")
		}
		model.DirList.SetItems(items)

		// Verify directories are loaded (excluding parent directory entry)
		dirCount := 0
		for _, item := range items {
			cleanItem := item.(models.CleanItem)
			if cleanItem.Size != -1 { // Skip parent directory entry
				dirCount++
			}
		}
		if dirCount != len(subDirs) {
			t.Errorf("Expected %d directories, got %d", len(subDirs), dirCount)
		}
	})

	t.Run("Directory Size Calculation", func(t *testing.T) {
		model := setupCleanTestModel(t)
		model.Init()

		tempDir := model.CurrentPath
		testFiles := map[string]int64{
			"file1.txt": 10,
			"file2.txt": 15,
			"file3.txt": 1,
		}
		for name, size := range testFiles {
			filePath := filepath.Join(tempDir, name)
			if err := os.WriteFile(filePath, make([]byte, size), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}

		cmd := model.CalculateDirSizeAsync()
		msg := cmd()
		if dirSizeMsg, ok := msg.(views.DirSizeMsg); !ok {
			t.Fatal("Expected DirSizeMsg")
		} else if dirSizeMsg.Size != 26 { // 10 + 15 + 1
			t.Errorf("Expected total size 26, got %d", dirSizeMsg.Size)
		}
	})

	t.Run("Hidden File Handling", func(t *testing.T) {
		model := setupCleanTestModel(t)
		model.Init()

		tempDir := model.CurrentPath
		testFiles := []struct {
			name    string
			content string
			size    int64
			hidden  bool
		}{
			{"test1.txt", "test content 1", 13, false},
			{".hidden1.txt", "hidden content 1", 14, true},
			{"test2.txt", "test content 2", 13, false},
			{".hidden2.txt", "hidden content 2", 14, true},
		}

		for _, file := range testFiles {
			filePath := filepath.Join(tempDir, file.name)
			if err := os.WriteFile(filePath, []byte(file.content), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", file.name, err)
			}
		}

		model.OptionState[options.ShowHiddenFiles] = false
		cmd := model.LoadFiles()
		msg := cmd()
		if err, ok := msg.(*errors.Error); ok {
			t.Fatalf("Failed to load files: %v", err)
		}
		items, ok := msg.([]list.Item)
		if !ok {
			t.Fatalf("LoadFiles() did not return []list.Item")
		}
		model.List.SetItems(items)

		// Verify hidden files are not shown (excluding parent directory entry)
		fileCount := 0
		for _, item := range items {
			cleanItem := item.(models.CleanItem)
			if cleanItem.Size != -1 { // Skip parent directory entry
				fileCount++
			}
		}
		expectedCount := 2 // Only non-hidden files
		if fileCount != expectedCount {
			t.Errorf("Expected %d files with hidden files disabled, got %d", expectedCount, fileCount)
		}

		model.OptionState[options.ShowHiddenFiles] = true
		cmd = model.LoadFiles()
		msg = cmd()
		if err, ok := msg.(*errors.Error); ok {
			t.Fatalf("Failed to load files: %v", err)
		}
		items, ok = msg.([]list.Item)
		if !ok {
			t.Fatalf("LoadFiles() did not return []list.Item")
		}
		model.List.SetItems(items)

		// Verify all files are shown (excluding parent directory entry)
		fileCount = 0
		for _, item := range items {
			cleanItem := item.(models.CleanItem)
			if cleanItem.Size != -1 { // Skip parent directory entry
				fileCount++
			}
		}
		expectedCount = 4
		if fileCount != expectedCount {
			t.Errorf("Expected %d files with hidden files enabled, got %d", expectedCount, fileCount)
		}
	})
}

func TestCleanFilesModel_DeletionOperations(t *testing.T) {
	t.Run("Batch Deletion", func(t *testing.T) {
		model := setupCleanTestModel(t)
		model.Init()

		tempDir := model.CurrentPath
		testFiles := []struct {
			name    string
			content string
			size    int64
		}{
			{"test1.txt", "test content 1", 13},
			{"test2.txt", "test content 2", 13},
			{"test3.txt", "test content 3", 13},
		}

		for _, file := range testFiles {
			filePath := filepath.Join(tempDir, file.name)
			if err := os.WriteFile(filePath, []byte(file.content), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", file.name, err)
			}
		}

		model.OptionState[options.ConfirmDeletion] = false

		cmd := model.LoadFiles()
		msg := cmd()
		if err, ok := msg.(*errors.Error); ok {
			t.Fatalf("Failed to load files: %v", err)
		}
		items, ok := msg.([]list.Item)
		if !ok {
			t.Fatalf("LoadFiles() did not return []list.Item")
		}
		model.List.SetItems(items)

		newModel, cmd := model.OnDelete()
		if m, ok := newModel.(*views.CleanFilesModel); ok {
			model = m
		} else {
			t.Fatal("OnDelete() did not return CleanFilesModel")
		}
		msg = cmd()
		if err, ok := msg.(*errors.Error); ok {
			t.Fatalf("Failed to execute delete command: %v", err)
		}
		items, ok = msg.([]list.Item)
		if !ok {
			t.Fatalf("OnDelete() did not return []list.Item")
		}
		model.List.SetItems(items)

		// Verify all files are deleted (excluding parent directory entry)
		fileCount := 0
		for _, item := range items {
			cleanItem := item.(models.CleanItem)
			if cleanItem.Size != -1 { // Skip parent directory entry
				fileCount++
			}
		}
		if fileCount != 0 {
			t.Errorf("Expected no files after batch deletion, got %d", fileCount)
		}
	})

	t.Run("Deletion with Confirmation", func(t *testing.T) {
		model := setupCleanTestModel(t)
		model.Init()

		tempDir := model.CurrentPath
		filePath := filepath.Join(tempDir, "test.txt")
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		model.OptionState[options.ConfirmDeletion] = true

		cmd := model.LoadFiles()
		msg := cmd()
		if err, ok := msg.(*errors.Error); ok {
			t.Fatalf("Failed to load files: %v", err)
		}
		items, ok := msg.([]list.Item)
		if !ok {
			t.Fatalf("LoadFiles() did not return []list.Item")
		}
		model.List.SetItems(items)

		// Select first file (excluding parent directory entry)
		for i, item := range items {
			cleanItem := item.(models.CleanItem)
			if cleanItem.Size != -1 { // Skip parent directory entry
				model.List.Select(i)
				break
			}
		}

		newModel, cmd := model.OnDelete()
		if m, ok := newModel.(*views.CleanFilesModel); ok {
			model = m
		} else {
			t.Fatal("OnDelete() did not return CleanFilesModel")
		}
		msg = cmd()
		if err, ok := msg.(*errors.Error); ok {
			t.Fatalf("Failed to execute delete command: %v", err)
		}
		items, ok = msg.([]list.Item)
		if !ok {
			t.Fatalf("OnDelete() did not return []list.Item")
		}
		model.List.SetItems(items)

		// Verify file is deleted (excluding parent directory entry)
		fileCount := 0
		for _, item := range items {
			cleanItem := item.(models.CleanItem)
			if cleanItem.Size != -1 { // Skip parent directory entry
				fileCount++
			}
		}
		if fileCount != 0 {
			t.Errorf("Expected no files after deletion, got %d", fileCount)
		}
	})
}

func TestCleanFilesModel_OptionsAndSettings(t *testing.T) {
	t.Run("Option Toggling", func(t *testing.T) {
		model := setupCleanTestModel(t)
		model.Init()

		for i := 1; i <= len(options.DefaultCleanOption); i++ {
			optionKey := fmt.Sprintf("alt+%d", i)
			initialState := model.OptionState[options.DefaultCleanOption[i-1]]

			model.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(optionKey)})

			if model.OptionState[options.DefaultCleanOption[i-1]] == initialState {
				t.Errorf("Option %s state did not change after toggling", options.DefaultCleanOption[i-1])
			}
		}
	})

}
func compareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
