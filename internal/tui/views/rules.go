package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rules "github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/errors"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/styles"
	rulesTab "github.com/pashkov256/deletor/internal/tui/tabs/rules"
	"github.com/pashkov256/deletor/internal/utils"
)

// RulesModel represents the rules management page
type RulesModel struct {
	// Main tab fields
	locationInput textinput.Model

	// Filters tab fields
	extensionsInput textinput.Model
	minSizeInput    textinput.Model
	maxSizeInput    textinput.Model
	excludeInput    textinput.Model
	olderInput      textinput.Model
	newerInput      textinput.Model

	// Options tab fields
	optionState      map[string]bool
	optionFocusIndex int

	// Common fields
	rules      rules.Rules
	focusIndex int
	rulesPath  string
	Error      *errors.Error
	TabManager *rulesTab.RulesTabManager
}

// NewRulesModel creates a new rules management model
func NewRulesModel(rules rules.Rules) *RulesModel {
	// Initialize inputs
	currentRules, _ := rules.GetRules()

	// Main tab inputs
	locationInput := textinput.New()
	locationInput.Placeholder = "Target location (e.g. C:\\Users\\Downloads)"
	locationInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	locationInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	locationInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	locationInput.SetValue(currentRules.Path)

	// Filters tab inputs
	extensionsInput := textinput.New()
	extensionsInput.Placeholder = "File extensions (e.g. tmp,log,bak)"
	extensionsInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	extensionsInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	extensionsInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	extensionsInput.SetValue(strings.Join(currentRules.Extensions, ","))

	minSizeInput := textinput.New()
	minSizeInput.Placeholder = "Minimum file size (e.g. 10kb)"
	minSizeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	minSizeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	minSizeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	minSizeInput.SetValue(currentRules.MinSize)

	maxSizeInput := textinput.New()
	maxSizeInput.Placeholder = "Maximum file size (e.g. 1gb)"
	maxSizeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	maxSizeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	maxSizeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	maxSizeInput.SetValue(currentRules.MaxSize)

	excludeInput := textinput.New()
	excludeInput.Placeholder = "Exclude specific files/paths (e.g. data,backup)"
	excludeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	excludeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	excludeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	excludeInput.SetValue(strings.Join(currentRules.Exclude, ","))
	olderInput := textinput.New()
	olderInput.Placeholder = "Older than (e.g. 60 min, 1 hour, 7 days, 1 month)"

	olderInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	olderInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	olderInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	olderInput.SetValue(currentRules.OlderThan)

	newerInput := textinput.New()
	newerInput.Placeholder = "Newer than (e.g. 60 min, 1 hour, 7 days, 1 month)"
	newerInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	newerInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	newerInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	newerInput.SetValue(currentRules.NewerThan)

	// Initialize options state
	optionState := make(map[string]bool)
	for _, name := range options.DefaultCleanOption {
		optionState[name] = false
	}

	// Get AppData path
	rulesPath := filepath.Join(os.Getenv("APPDATA"), rules.GetRulesPath())

	return &RulesModel{
		locationInput:    locationInput,
		extensionsInput:  extensionsInput,
		minSizeInput:     minSizeInput,
		maxSizeInput:     maxSizeInput,
		excludeInput:     excludeInput,
		olderInput:       olderInput,
		newerInput:       newerInput,
		optionState:      optionState,
		optionFocusIndex: 0,
		focusIndex:       0,
		rulesPath:        rulesPath,
		rules:            rules,
	}
}

func (m *RulesModel) Init() tea.Cmd {
	// Initialize tab manager
	m.TabManager = rulesTab.NewRulesTabManager(m, rulesTab.NewRulesTabFactory())
	return textinput.Blink
}

func (m *RulesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keyboard events directly
		return m.Handle(msg)
	case *errors.Error:
		m.Error = msg
		return m, nil
	}

	// Handle input updates based on active tab
	activeTab := m.TabManager.GetActiveTabIndex()
	var cmd tea.Cmd

	switch activeTab {
	case 0: // Main tab
		m.locationInput, cmd = m.locationInput.Update(msg)
		cmds = append(cmds, cmd)
	case 1: // Filters tab
		switch m.focusIndex {
		case 0:
			m.extensionsInput, cmd = m.extensionsInput.Update(msg)
		case 1:
			m.minSizeInput, cmd = m.minSizeInput.Update(msg)
		case 2:
			m.maxSizeInput, cmd = m.maxSizeInput.Update(msg)
		case 3:
			m.excludeInput, cmd = m.excludeInput.Update(msg)
		case 4:
			m.olderInput, cmd = m.olderInput.Update(msg)
		case 5:
			m.newerInput, cmd = m.newerInput.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *RulesModel) View() string {
	activeTab := m.TabManager.GetActiveTabIndex()
	tabNames := []string{"🗂️ [F1] Main", "🧹 [F2] Filters", "⚙️ [F3] Options"}
	tabs := make([]string, 3)
	for i, name := range tabNames {
		style := styles.TabStyle
		if activeTab == i {
			style = styles.ActiveTabStyle
		}
		tabs[i] = style.Render(name)
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

	// --- Content rendering ---
	var content strings.Builder
	content.WriteString(tabsRow)
	content.WriteString("\n")

	// Render active tab content
	switch activeTab {
	case 0: // Main tab
		pathStyle := styles.StandardInputStyle
		if m.focusIndex == 0 {
			pathStyle = styles.StandardInputFocusedStyle
		}
		content.WriteString(pathStyle.Render("Path: " + m.locationInput.View()))
		content.WriteString("\n\n")

		saveButtonStyle := styles.StandardButtonStyle
		if m.focusIndex == 1 {
			saveButtonStyle = styles.StandardButtonFocusedStyle
		}
		content.WriteString(saveButtonStyle.Render("💾 Save rules"))

	case 1: // Filters tab
		inputs := []struct {
			name  string
			input textinput.Model
		}{
			{"Extensions", m.extensionsInput},
			{"Min Size", m.minSizeInput},
			{"Max Size", m.maxSizeInput},
			{"Exclude", m.excludeInput},
			{"Older Than", m.olderInput},
			{"Newer Than", m.newerInput},
		}

		for i, input := range inputs {
			style := styles.StandardInputStyle
			if m.focusIndex == i {
				style = styles.StandardInputFocusedStyle
			}
			content.WriteString(style.Render(input.name + ": " + input.input.View()))
			content.WriteString("\n")
		}

	case 2: // Options tab
		for i, name := range options.DefaultCleanOption {
			style := styles.OptionStyle
			if m.optionState[name] {
				style = styles.SelectedOptionStyle
			}
			if m.optionFocusIndex == i {
				style = styles.OptionFocusedStyle
			}

			emoji := ""
			switch name {
			case options.ShowHiddenFiles:
				emoji = "👁️"
			case options.ConfirmDeletion:
				emoji = "⚠️"
			case options.IncludeSubfolders:
				emoji = "📁"
			case options.DeleteEmptySubfolders:
				emoji = "🗑️"
			case options.SendFilesToTrash:
				emoji = "♻️"
			case options.LogOperations:
				emoji = "📝"
			case options.LogToFile:
				emoji = "📄"
			case options.ShowStatistics:
				emoji = "📊"
			case options.ExitAfterDeletion:
				emoji = "🚪"
			}

			content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", i+1)))
			content.WriteString(style.Render(fmt.Sprintf("[%s] %s %-20s",
				map[bool]string{true: "✓", false: "○"}[m.optionState[name]],
				emoji, name)))
			content.WriteString("\n")
		}
	}

	// Add error message if there is one
	if m.Error != nil && m.Error.IsVisible() {
		errorStyle := errors.GetStyle(m.Error.GetType())
		content.WriteString("\n")
		content.WriteString(errorStyle.Render(m.Error.GetMessage()))
	}

	return styles.AppStyle.Render(content.String())
}

func (m *RulesModel) Handle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	// Handle special keys first
	switch msg.String() {
	case "tab":
		return m.handleTab()
	case "shift+tab":
		return m.handleShiftTab()
	case "up", "down":
		return m.handleUpDown(msg.String())
	case "right":
		return m.handleArrowRight()
	case "left":
		return m.handleArrowLeft()
	case "f1":
		return m.handleF1()
	case "f2":
		return m.handleF2()
	case "f3":
		return m.handleF3()
	case "enter":
		return m.handleEnter()
	case "ctrl+s":
		return m.handleSave()
	case "alt+c":
		return m.handleAltC()
	case " ":
		if activeTab == 2 { // Options tab
			name := options.DefaultCleanOption[m.optionFocusIndex]
			m.optionState[name] = !m.optionState[name]
			return m, nil
		}
	}

	// Handle text input for focused element
	var cmd tea.Cmd
	switch activeTab {
	case 0: // Main tab
		m.locationInput, cmd = m.locationInput.Update(msg)
	case 1: // Filters tab
		switch m.focusIndex {
		case 0:
			m.extensionsInput, cmd = m.extensionsInput.Update(msg)
		case 1:
			m.minSizeInput, cmd = m.minSizeInput.Update(msg)
		case 2:
			m.maxSizeInput, cmd = m.maxSizeInput.Update(msg)
		case 3:
			m.excludeInput, cmd = m.excludeInput.Update(msg)
		case 4:
			m.olderInput, cmd = m.olderInput.Update(msg)
		case 5:
			m.newerInput, cmd = m.newerInput.Update(msg)
		}
	}

	return m, cmd
}

func (m *RulesModel) handleTab() (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	switch activeTab {
	case 0: // Main tab
		switch m.focusIndex {
		case 0:
			m.locationInput.Blur()
			m.focusIndex = 1 // Save button
		case 1:
			m.focusIndex = 0
			m.locationInput.Focus()
		}
	case 1: // Filters tab
		switch m.focusIndex {
		case 0:
			m.extensionsInput.Blur()
			m.focusIndex = 1
			m.minSizeInput.Focus()
		case 1:
			m.minSizeInput.Blur()
			m.focusIndex = 2
			m.maxSizeInput.Focus()
		case 2:
			m.maxSizeInput.Blur()
			m.focusIndex = 3
			m.excludeInput.Focus()
		case 3:
			m.excludeInput.Blur()
			m.focusIndex = 4
			m.olderInput.Focus()
		case 4:
			m.olderInput.Blur()
			m.focusIndex = 5
			m.newerInput.Focus()
		case 5:
			m.newerInput.Blur()
			m.focusIndex = 0
			m.extensionsInput.Focus()
		}
	case 2: // Options tab
		if m.optionFocusIndex < len(options.DefaultCleanOption)-1 {
			m.optionFocusIndex++
		} else {
			m.optionFocusIndex = 0
		}
	}

	return m, nil
}

func (m *RulesModel) handleShiftTab() (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	switch activeTab {
	case 0: // Main tab
		switch m.focusIndex {
		case 0:
			m.locationInput.Blur()
			m.focusIndex = 1 // Save button
		case 1:
			m.focusIndex = 0
			m.locationInput.Focus()
		}
	case 1: // Filters tab
		switch m.focusIndex {
		case 0:
			m.extensionsInput.Blur()
			m.focusIndex = 5
			m.newerInput.Focus()
		case 1:
			m.minSizeInput.Blur()
			m.focusIndex = 0
			m.extensionsInput.Focus()
		case 2:
			m.maxSizeInput.Blur()
			m.focusIndex = 1
			m.minSizeInput.Focus()
		case 3:
			m.excludeInput.Blur()
			m.focusIndex = 2
			m.maxSizeInput.Focus()
		case 4:
			m.olderInput.Blur()
			m.focusIndex = 3
			m.excludeInput.Focus()
		case 5:
			m.newerInput.Blur()
			m.focusIndex = 4
			m.olderInput.Focus()
		}
	case 2: // Options tab
		if m.optionFocusIndex > 0 {
			m.optionFocusIndex--
		} else {
			m.optionFocusIndex = len(options.DefaultCleanOption) - 1
		}
	}

	return m, nil
}

func (m *RulesModel) handleUpDown(key string) (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	if activeTab == 2 { // Options tab
		if key == "up" {
			if m.optionFocusIndex > 0 {
				m.optionFocusIndex--
			} else {
				m.optionFocusIndex = len(options.DefaultCleanOption) - 1
			}
		} else {
			if m.optionFocusIndex < len(options.DefaultCleanOption)-1 {
				m.optionFocusIndex++
			} else {
				m.optionFocusIndex = 0
			}
		}
		return m, nil
	}

	if key == "up" {
		return m.handleShiftTab()
	}
	return m.handleTab()
}

func (m *RulesModel) handleArrowRight() (tea.Model, tea.Cmd) {
	tabLength := len(m.TabManager.GetAllTabs())
	activeTabIndex := m.TabManager.GetActiveTabIndex()

	if tabLength-1 == activeTabIndex {
		m.TabManager.SetActiveTabIndex(0)
	} else {
		m.TabManager.SetActiveTabIndex(activeTabIndex + 1)
	}

	return m, nil
}

func (m *RulesModel) handleArrowLeft() (tea.Model, tea.Cmd) {
	tabLength := len(m.TabManager.GetAllTabs())
	activeTabIndex := m.TabManager.GetActiveTabIndex()

	if activeTabIndex == 0 {
		m.TabManager.SetActiveTabIndex(tabLength - 1)
	} else {
		m.TabManager.SetActiveTabIndex(activeTabIndex - 1)
	}

	return m, nil
}

func (m *RulesModel) handleF1() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(0)
	m.focusIndex = 0
	m.locationInput.Focus()
	return m, nil
}

func (m *RulesModel) handleF2() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(1)
	m.focusIndex = 0
	m.extensionsInput.Focus()
	return m, nil
}

func (m *RulesModel) handleF3() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(2)
	m.optionFocusIndex = 0
	return m, nil
}

func (m *RulesModel) handleEnter() (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	if activeTab == 0 && m.focusIndex == 1 { // Save button in Main tab
		if err := m.validateInputs(); err != nil {
			return m, func() tea.Msg {
				return err
			}
		}

		// Clear error if validation passed
		m.Error = nil

		// Save rules
		m.rules.UpdateRules(
			rules.WithPath(m.locationInput.Value()),
			rules.WithMinSize(m.minSizeInput.Value()),
			rules.WithMaxSize(m.maxSizeInput.Value()),
			rules.WithExtensions(utils.ParseExtToSlice(m.extensionsInput.Value())),
			rules.WithExclude(utils.ParseExcludeToSlice(m.excludeInput.Value())),
			rules.WithOlderThan(m.olderInput.Value()),
			rules.WithNewerThan(m.newerInput.Value()),
			rules.WithOptions(
				m.optionState[options.ShowHiddenFiles],
				m.optionState[options.ConfirmDeletion],
				m.optionState[options.IncludeSubfolders],
				m.optionState[options.DeleteEmptySubfolders],
				m.optionState[options.SendFilesToTrash],
				m.optionState[options.LogOperations],
				m.optionState[options.LogToFile],
				m.optionState[options.ShowStatistics],
				m.optionState[options.ExitAfterDeletion],
			),
		)
	} else if activeTab == 2 { // Options tab
		name := options.DefaultCleanOption[m.optionFocusIndex]
		m.optionState[name] = !m.optionState[name]
	}

	return m, nil
}

func (m *RulesModel) handleSave() (tea.Model, tea.Cmd) {
	return m.handleEnter()
}

func (m *RulesModel) handleAltC() (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	switch activeTab {
	case 0: // Main tab
		m.locationInput.SetValue("")
	case 1: // Filters tab
		m.extensionsInput.SetValue("")
		m.minSizeInput.SetValue("")
		m.maxSizeInput.SetValue("")
		m.excludeInput.SetValue("")
		m.olderInput.SetValue("")
		m.newerInput.SetValue("")
	case 2: // Options tab
		for name := range m.optionState {
			m.optionState[name] = false
		}
	}

	return m, nil
}

func (m *RulesModel) validateInputs() *errors.Error {
	// Validate size inputs
	if m.minSizeInput.Value() != "" {
		if _, err := utils.ToBytes(m.minSizeInput.Value()); err != nil {
			return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid min size format: %v", err))
		}
	}
	if m.maxSizeInput.Value() != "" {
		if _, err := utils.ToBytes(m.maxSizeInput.Value()); err != nil {
			return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid max size format: %v", err))
		}
	}

	// Validate location input
	if m.locationInput.Value() != "" {
		expandedPath := utils.ExpandTilde(m.locationInput.Value())
		if _, err := os.Stat(expandedPath); err != nil {
			return errors.New(errors.ErrorTypeFileSystem, fmt.Sprintf("Invalid path: %s", m.locationInput.Value()))
		}
	}

	return nil
}

func (m *RulesModel) GetPathInput() textinput.Model {
	return m.locationInput
}

func (m *RulesModel) GetExtInput() textinput.Model {
	return m.extensionsInput
}

func (m *RulesModel) GetMinSizeInput() textinput.Model {
	return m.minSizeInput
}

func (m *RulesModel) GetMaxSizeInput() textinput.Model {
	return m.maxSizeInput
}

func (m *RulesModel) GetExcludeInput() textinput.Model {
	return m.excludeInput
}

func (m *RulesModel) GetOlderInput() textinput.Model {
	return m.olderInput
}

func (m *RulesModel) GetNewerInput() textinput.Model {
	return m.newerInput
}

func (m *RulesModel) GetFocusedElement() string {
	activeTab := m.TabManager.GetActiveTabIndex()
	if activeTab == 2 {
		return fmt.Sprintf("option%d", m.optionFocusIndex+1)
	}
	return fmt.Sprintf("input%d", m.focusIndex)
}

func (m *RulesModel) GetOptionState() map[string]bool {
	return m.optionState
}

func (m *RulesModel) SetFocusedElement(element string) {
	if strings.HasPrefix(element, "option") {
		if num, err := strconv.Atoi(element[6:]); err == nil {
			m.optionFocusIndex = num - 1
		}
	} else if strings.HasPrefix(element, "input") {
		if num, err := strconv.Atoi(element[5:]); err == nil {
			m.focusIndex = num
		}
	}
}

func (m *RulesModel) SetOptionState(option string, state bool) {
	m.optionState[option] = state
}
