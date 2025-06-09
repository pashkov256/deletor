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
	zone "github.com/lrstanley/bubblezone"
	rules "github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/errors"
	"github.com/pashkov256/deletor/internal/tui/help"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/styles"
	rulesTab "github.com/pashkov256/deletor/internal/tui/tabs/rules"
	"github.com/pashkov256/deletor/internal/utils"
	"github.com/pashkov256/deletor/internal/validation"
)

// RulesModel represents the rules management page
type RulesModel struct {
	// Main tab fields
	LocationInput textinput.Model

	// Filters tab fields
	ExtensionsInput textinput.Model
	MinSizeInput    textinput.Model
	MaxSizeInput    textinput.Model
	ExcludeInput    textinput.Model
	OlderInput      textinput.Model
	NewerInput      textinput.Model

	// Options tab fields
	OptionState map[string]bool

	// Common fields
	rules          rules.Rules
	FocusedElement string // "locationInput", "saveButton", "extensionsInput", "minSizeInput", "maxSizeInput", "excludeInput", "olderInput", "newerInput", "option1", "option2", etc.
	rulesPath      string
	Error          *errors.Error
	TabManager     *rulesTab.RulesTabManager
	Validator      *validation.Validator
}

// NewRulesModel creates a new rules management model
func NewRulesModel(rules rules.Rules, validator *validation.Validator) *RulesModel {
	// Initialize inputs
	lastestRules, _ := rules.GetRules()

	// Main tab inputs
	locationInput := textinput.New()
	locationInput.Placeholder = "Target location (e.g. C:\\Users\\Downloads)"
	locationInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	locationInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	locationInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	locationInput.SetValue(lastestRules.Path)

	// Filters tab inputs
	extensionsInput := textinput.New()
	extensionsInput.Placeholder = "File extensions (e.g. tmp,log,bak)"
	extensionsInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	extensionsInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	extensionsInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	extensionsInput.SetValue(strings.Join(lastestRules.Extensions, ","))

	minSizeInput := textinput.New()
	minSizeInput.Placeholder = "Minimum file size (e.g. 10kb)"
	minSizeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	minSizeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	minSizeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	minSizeInput.SetValue(lastestRules.MinSize)

	maxSizeInput := textinput.New()
	maxSizeInput.Placeholder = "Maximum file size (e.g. 1gb)"
	maxSizeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	maxSizeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	maxSizeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	maxSizeInput.SetValue(lastestRules.MaxSize)

	excludeInput := textinput.New()
	excludeInput.Placeholder = "Exclude specific files/paths (e.g. data,backup)"
	excludeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	excludeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	excludeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	excludeInput.SetValue(strings.Join(lastestRules.Exclude, ","))
	olderInput := textinput.New()
	olderInput.Placeholder = "Older than (e.g. 60 min, 1 hour, 7 days, 1 month)"

	olderInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	olderInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	olderInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	olderInput.SetValue(lastestRules.OlderThan)

	newerInput := textinput.New()
	newerInput.Placeholder = "Newer than (e.g. 60 min, 1 hour, 7 days, 1 month)"
	newerInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	newerInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	newerInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	newerInput.SetValue(lastestRules.NewerThan)

	// Get AppData path
	rulesPath := filepath.Join(os.Getenv("APPDATA"), rules.GetRulesPath())

	return &RulesModel{
		LocationInput:   locationInput,
		ExtensionsInput: extensionsInput,
		MinSizeInput:    minSizeInput,
		MaxSizeInput:    maxSizeInput,
		ExcludeInput:    excludeInput,
		OlderInput:      olderInput,
		NewerInput:      newerInput,
		OptionState: map[string]bool{
			options.ShowHiddenFiles:       lastestRules.ShowHiddenFiles,
			options.ConfirmDeletion:       lastestRules.ConfirmDeletion,
			options.IncludeSubfolders:     lastestRules.IncludeSubfolders,
			options.DeleteEmptySubfolders: lastestRules.DeleteEmptySubfolders,
			options.SendFilesToTrash:      lastestRules.SendFilesToTrash,
			options.LogOperations:         lastestRules.LogOperations,
			options.LogToFile:             lastestRules.LogToFile,
			options.ShowStatistics:        lastestRules.ShowStatistics,
			options.ExitAfterDeletion:     lastestRules.ExitAfterDeletion,
		},
		rules:          rules,
		rulesPath:      rulesPath,
		FocusedElement: "locationInput",
		Validator:      validator,
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

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			// Handle tab clicks
			for i := 0; i < 3; i++ {
				if zone.Get(fmt.Sprintf("tab_%d", i)).InBounds(msg) {
					m.TabManager.SetActiveTabIndex(i)
					// Blur all inputs
					m.LocationInput.Blur()
					m.ExtensionsInput.Blur()
					m.MinSizeInput.Blur()
					m.MaxSizeInput.Blur()
					m.ExcludeInput.Blur()
					m.OlderInput.Blur()
					m.NewerInput.Blur()

					switch i {
					case 0:
						m.FocusedElement = "locationInput"
						m.LocationInput.Focus()
					case 1:
						m.FocusedElement = "extensionsInput"
						m.ExtensionsInput.Focus()
					case 2:
						m.FocusedElement = "option1"
					}
					return m, nil
				}
			}

			// Handle main tab elements
			if zone.Get("rules_location_input").InBounds(msg) {
				// Blur all other inputs
				m.ExtensionsInput.Blur()
				m.MinSizeInput.Blur()
				m.MaxSizeInput.Blur()
				m.ExcludeInput.Blur()
				m.OlderInput.Blur()
				m.NewerInput.Blur()

				m.FocusedElement = "locationInput"
				m.LocationInput.Focus()
			} else if zone.Get("rules_save_button").InBounds(msg) {
				// Blur all inputs
				m.LocationInput.Blur()
				m.ExtensionsInput.Blur()
				m.MinSizeInput.Blur()
				m.MaxSizeInput.Blur()
				m.ExcludeInput.Blur()
				m.OlderInput.Blur()
				m.NewerInput.Blur()

				m.FocusedElement = "saveButton"
				return m.handleEnter()
			}

			// Handle filters tab elements
			if m.TabManager.GetActiveTabIndex() == 1 {
				for _, key := range []string{"extensionsInput", "minSizeInput", "maxSizeInput", "excludeInput", "olderInput", "newerInput"} {
					if zone.Get(fmt.Sprintf("rules_%s", key)).InBounds(msg) {
						// Blur all inputs
						m.LocationInput.Blur()
						m.ExtensionsInput.Blur()
						m.MinSizeInput.Blur()
						m.MaxSizeInput.Blur()
						m.ExcludeInput.Blur()
						m.OlderInput.Blur()
						m.NewerInput.Blur()

						m.FocusedElement = key
						switch key {
						case "extensionsInput":
							m.ExtensionsInput.Focus()
						case "minSizeInput":
							m.MinSizeInput.Focus()
						case "maxSizeInput":
							m.MaxSizeInput.Focus()
						case "excludeInput":
							m.ExcludeInput.Focus()
						case "olderInput":
							m.OlderInput.Focus()
						case "newerInput":
							m.NewerInput.Focus()
						}
						return m, nil
					}
				}
			}

			// Handle options tab elements
			if m.TabManager.GetActiveTabIndex() == 2 {
				for i := 1; i <= len(options.DefaultCleanOption); i++ {
					if zone.Get(fmt.Sprintf("rules_option_%d", i)).InBounds(msg) {
						// Blur all inputs
						m.LocationInput.Blur()
						m.ExtensionsInput.Blur()
						m.MinSizeInput.Blur()
						m.MaxSizeInput.Blur()
						m.ExcludeInput.Blur()
						m.OlderInput.Blur()
						m.NewerInput.Blur()

						m.FocusedElement = fmt.Sprintf("option%d", i)
						name := options.DefaultCleanOption[i-1]
						m.OptionState[name] = !m.OptionState[name]
						return m, nil
					}
				}
			}
		}
		return m, nil

	case *errors.Error:
		m.Error = msg
		return m, nil
	}

	// Handle input updates based on active tab
	activeTab := m.TabManager.GetActiveTabIndex()
	var cmd tea.Cmd

	switch activeTab {
	case 0: // Main tab
		m.LocationInput, cmd = m.LocationInput.Update(msg)
		cmds = append(cmds, cmd)
	case 1: // Filters tab
		switch m.FocusedElement {
		case "extensionsInput":
			m.ExtensionsInput, cmd = m.ExtensionsInput.Update(msg)
		case "minSizeInput":
			m.MinSizeInput, cmd = m.MinSizeInput.Update(msg)
		case "maxSizeInput":
			m.MaxSizeInput, cmd = m.MaxSizeInput.Update(msg)
		case "excludeInput":
			m.ExcludeInput, cmd = m.ExcludeInput.Update(msg)
		case "olderInput":
			m.OlderInput, cmd = m.OlderInput.Update(msg)
		case "newerInput":
			m.NewerInput, cmd = m.NewerInput.Update(msg)
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
		tabs[i] = zone.Mark(fmt.Sprintf("tab_%d", i), style.Render(name))
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
		if m.FocusedElement == "locationInput" {
			pathStyle = styles.StandardInputFocusedStyle
		}
		content.WriteString(zone.Mark("rules_location_input", pathStyle.Render("Path: "+m.LocationInput.View())))
		content.WriteString("\n\n")

		saveButtonStyle := styles.StandardButtonStyle
		if m.FocusedElement == "saveButton" {
			saveButtonStyle = styles.StandardButtonFocusedStyle
		}
		content.WriteString(zone.Mark("rules_save_button", saveButtonStyle.Render("💾 Save rules")))
		content.WriteString("\n\n\n")
		content.WriteString(styles.PathStyle.Render(fmt.Sprintf("Rules are stored in: %s", m.rulesPath)))
		content.WriteString("\n\n" + help.NavigateHelpText)
	case 1: // Filters tab
		inputs := []struct {
			name  string
			input textinput.Model
			key   string
		}{
			{"Extensions", m.ExtensionsInput, "extensionsInput"},
			{"Min Size", m.MinSizeInput, "minSizeInput"},
			{"Max Size", m.MaxSizeInput, "maxSizeInput"},
			{"Exclude", m.ExcludeInput, "excludeInput"},
			{"Older Than", m.OlderInput, "olderInput"},
			{"Newer Than", m.NewerInput, "newerInput"},
		}

		for _, input := range inputs {
			style := styles.StandardInputStyle
			if m.FocusedElement == input.key {
				style = styles.StandardInputFocusedStyle
			}
			content.WriteString(zone.Mark(fmt.Sprintf("rules_%s", input.key), style.Render(input.name+": "+input.input.View())))
			content.WriteString("\n")
		}

	case 2: // Options tab
		for i, name := range options.DefaultCleanOption {
			style := styles.OptionStyle
			if m.OptionState[name] {
				style = styles.SelectedOptionStyle
			}
			if m.FocusedElement == fmt.Sprintf("option%d", i+1) {
				style = styles.OptionFocusedStyle
			}

			emoji := options.GetEmojiByCleanOption(name)

			content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", i+1)))
			content.WriteString(zone.Mark(fmt.Sprintf("rules_option_%d", i+1), style.Render(fmt.Sprintf("[%s] %s %-20s",
				map[bool]string{true: "✓", false: "○"}[m.OptionState[name]],
				emoji, name))))
			content.WriteString("\n")
		}
	}

	// Add error message if there is one
	if m.Error != nil && m.Error.IsVisible() {
		errorStyle := errors.GetStyle(m.Error.GetType())
		content.WriteString("\n")
		content.WriteString(errorStyle.Render(m.Error.GetMessage()))
	}

	return zone.Scan(styles.AppStyle.Render(content.String()))
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
			if strings.HasPrefix(m.FocusedElement, "option") {
				optionNum := strings.TrimPrefix(m.FocusedElement, "option")
				idx, err := strconv.Atoi(optionNum)
				if err == nil && idx > 0 && idx <= len(options.DefaultCleanOption) {
					name := options.DefaultCleanOption[idx-1]
					m.OptionState[name] = !m.OptionState[name]
				}
			}
		}
	}

	// Handle text input for focused element
	var cmd tea.Cmd
	switch activeTab {
	case 0: // Main tab
		m.LocationInput, cmd = m.LocationInput.Update(msg)
	case 1: // Filters tab
		switch m.FocusedElement {
		case "extensionsInput":
			m.ExtensionsInput, cmd = m.ExtensionsInput.Update(msg)
		case "minSizeInput":
			m.MinSizeInput, cmd = m.MinSizeInput.Update(msg)
		case "maxSizeInput":
			m.MaxSizeInput, cmd = m.MaxSizeInput.Update(msg)
		case "excludeInput":
			m.ExcludeInput, cmd = m.ExcludeInput.Update(msg)
		case "olderInput":
			m.OlderInput, cmd = m.OlderInput.Update(msg)
		case "newerInput":
			m.NewerInput, cmd = m.NewerInput.Update(msg)
		}
	}

	return m, cmd
}

func (m *RulesModel) handleTab() (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	switch activeTab {
	case 0: // Main tab
		switch m.FocusedElement {
		case "locationInput":
			m.LocationInput.Blur()
			m.FocusedElement = "saveButton"
		case "saveButton":
			m.FocusedElement = "locationInput"
			m.LocationInput.Focus()
		}
	case 1: // Filters tab
		switch m.FocusedElement {
		case "extensionsInput":
			m.ExtensionsInput.Blur()
			m.FocusedElement = "minSizeInput"
			m.MinSizeInput.Focus()
		case "minSizeInput":
			m.MinSizeInput.Blur()
			m.FocusedElement = "maxSizeInput"
			m.MaxSizeInput.Focus()
		case "maxSizeInput":
			m.MaxSizeInput.Blur()
			m.FocusedElement = "excludeInput"
			m.ExcludeInput.Focus()
		case "excludeInput":
			m.ExcludeInput.Blur()
			m.FocusedElement = "olderInput"
			m.OlderInput.Focus()
		case "olderInput":
			m.OlderInput.Blur()
			m.FocusedElement = "newerInput"
			m.NewerInput.Focus()
		case "newerInput":
			m.NewerInput.Blur()
			m.FocusedElement = "extensionsInput"
			m.ExtensionsInput.Focus()
		}
	case 2: // Options tab
		m.FocusedElement = options.GetNextOption(m.FocusedElement, len(options.DefaultCleanOption), true)
	}

	return m, nil
}

func (m *RulesModel) handleShiftTab() (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	switch activeTab {
	case 0: // Main tab
		switch m.FocusedElement {
		case "locationInput":
			m.LocationInput.Blur()
			m.FocusedElement = "saveButton"
		case "saveButton":
			m.FocusedElement = "locationInput"
			m.LocationInput.Focus()
		}
	case 1: // Filters tab
		switch m.FocusedElement {
		case "extensionsInput":
			m.ExtensionsInput.Blur()
			m.FocusedElement = "newerInput"
			m.NewerInput.Focus()
		case "minSizeInput":
			m.MinSizeInput.Blur()
			m.FocusedElement = "extensionsInput"
			m.ExtensionsInput.Focus()
		case "maxSizeInput":
			m.MaxSizeInput.Blur()
			m.FocusedElement = "minSizeInput"
			m.MinSizeInput.Focus()
		case "excludeInput":
			m.ExcludeInput.Blur()
			m.FocusedElement = "maxSizeInput"
			m.MaxSizeInput.Focus()
		case "olderInput":
			m.OlderInput.Blur()
			m.FocusedElement = "excludeInput"
			m.ExcludeInput.Focus()
		case "newerInput":
			m.NewerInput.Blur()
			m.FocusedElement = "olderInput"
			m.OlderInput.Focus()
		}
	case 2: // Options tab
		m.FocusedElement = options.GetNextOption(m.FocusedElement, len(options.DefaultCleanOption), false)
	}

	return m, nil
}

func (m *RulesModel) handleUpDown(key string) (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	if activeTab == 2 { // Options tab
		if key == "up" {
			m.FocusedElement = options.GetNextOption(m.FocusedElement, len(options.DefaultCleanOption), false)
		} else {
			m.FocusedElement = options.GetNextOption(m.FocusedElement, len(options.DefaultCleanOption), true)
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
		m.FocusedElement = "locationInput"
		m.LocationInput.Focus()
	} else {
		m.TabManager.SetActiveTabIndex(activeTabIndex + 1)
		switch activeTabIndex + 1 {
		case 1:
			m.FocusedElement = "extensionsInput"
			m.ExtensionsInput.Focus()
		case 2:
			m.FocusedElement = "option1"
		}
	}

	return m, nil
}

func (m *RulesModel) handleArrowLeft() (tea.Model, tea.Cmd) {
	tabLength := len(m.TabManager.GetAllTabs())
	activeTabIndex := m.TabManager.GetActiveTabIndex()

	if activeTabIndex == 0 {
		m.TabManager.SetActiveTabIndex(tabLength - 1)
		m.FocusedElement = "option1"
	} else {
		m.TabManager.SetActiveTabIndex(activeTabIndex - 1)
		switch activeTabIndex - 1 {
		case 0:
			m.FocusedElement = "locationInput"
			m.LocationInput.Focus()
		case 1:
			m.FocusedElement = "extensionsInput"
			m.ExtensionsInput.Focus()
		}
	}

	return m, nil
}

func (m *RulesModel) handleF1() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(0)
	m.FocusedElement = "locationInput"
	m.LocationInput.Focus()
	return m, nil
}

func (m *RulesModel) handleF2() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(1)
	m.FocusedElement = "extensionsInput"
	m.ExtensionsInput.Focus()
	return m, nil
}

func (m *RulesModel) handleF3() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(2)
	m.FocusedElement = "option1"
	return m, nil
}

func (m *RulesModel) handleEnter() (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	if activeTab == 0 && m.FocusedElement == "saveButton" { // Save button in Main tab
		if err := m.ValidateInputs(); err != nil {
			return m, func() tea.Msg {
				return err
			}
		}

		// Clear error if validation passed
		m.Error = nil

		// Save rules
		m.rules.UpdateRules(
			rules.WithPath(m.LocationInput.Value()),
			rules.WithMinSize(m.MinSizeInput.Value()),
			rules.WithMaxSize(m.MaxSizeInput.Value()),
			rules.WithExtensions(utils.ParseExtToSlice(m.ExtensionsInput.Value())),
			rules.WithExclude(utils.ParseExcludeToSlice(m.ExcludeInput.Value())),
			rules.WithOlderThan(m.OlderInput.Value()),
			rules.WithNewerThan(m.NewerInput.Value()),
			rules.WithOptions(
				m.OptionState[options.ShowHiddenFiles],
				m.OptionState[options.ConfirmDeletion],
				m.OptionState[options.IncludeSubfolders],
				m.OptionState[options.DeleteEmptySubfolders],
				m.OptionState[options.SendFilesToTrash],
				m.OptionState[options.LogOperations],
				m.OptionState[options.LogToFile],
				m.OptionState[options.ShowStatistics],
				m.OptionState[options.ExitAfterDeletion],
			),
		)
	} else if activeTab == 2 { // Options tab
		if strings.HasPrefix(m.FocusedElement, "option") {
			optionNum := strings.TrimPrefix(m.FocusedElement, "option")
			idx, err := strconv.Atoi(optionNum)
			if err == nil && idx > 0 && idx <= len(options.DefaultCleanOption) {
				name := options.DefaultCleanOption[idx-1]
				m.OptionState[name] = !m.OptionState[name]
			}
		}
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
		m.LocationInput.SetValue("")
	case 1: // Filters tab
		m.ExtensionsInput.SetValue("")
		m.MinSizeInput.SetValue("")
		m.MaxSizeInput.SetValue("")
		m.ExcludeInput.SetValue("")
		m.OlderInput.SetValue("")
		m.NewerInput.SetValue("")
	case 2: // Options tab
		for name := range m.OptionState {
			m.OptionState[name] = false
		}
	}

	return m, nil
}

// ValidateInputs validates all input fields in the model
func (m *RulesModel) ValidateInputs() *errors.Error {
	// Validate size inputs

	if m.MinSizeInput.Value() != "" {
		if m.Validator.ValidateSize(m.MinSizeInput.Value()) != nil {
			return errors.New(errors.ErrorTypeValidation, "Invalid (min size input) format")
		}
	}

	if m.MaxSizeInput.Value() != "" {
		if m.Validator.ValidateSize(m.MaxSizeInput.Value()) != nil {
			return errors.New(errors.ErrorTypeValidation, "Invalid (max size input) format")
		}
	}

	if m.NewerInput.Value() != "" {
		if m.Validator.ValidateTimeDuration(m.NewerInput.Value()) != nil {
			return errors.New(errors.ErrorTypeValidation, "Invalid (newer input) time format")
		}
	}

	if m.OlderInput.Value() != "" {
		if m.Validator.ValidateTimeDuration(m.OlderInput.Value()) != nil {
			return errors.New(errors.ErrorTypeValidation, "Invalid (older input) time format")
		}
	}

	// Validate location input
	if m.LocationInput.Value() != "" {
		expandedPath := utils.ExpandTilde(m.LocationInput.Value())
		if _, err := os.Stat(expandedPath); err != nil {
			return errors.New(errors.ErrorTypeFileSystem, fmt.Sprintf("Invalid path: %s", m.LocationInput.Value()))
		}
	}

	return nil
}

func (m *RulesModel) GetFocusedElement() string {
	return m.FocusedElement
}

func (m *RulesModel) SetFocusedElement(element string) {
	m.FocusedElement = element

}

func (m *RulesModel) GetOptionState() map[string]bool {
	return m.OptionState
}

func (m *RulesModel) SetOptionState(option string, state bool) {
	m.OptionState[option] = state
}

func (m *RulesModel) GetPathInput() textinput.Model {
	return m.LocationInput
}

func (m *RulesModel) GetExtInput() textinput.Model {
	return m.ExtensionsInput
}

func (m *RulesModel) GetMinSizeInput() textinput.Model {
	return m.MinSizeInput
}

func (m *RulesModel) GetMaxSizeInput() textinput.Model {
	return m.MaxSizeInput
}

func (m *RulesModel) GetExcludeInput() textinput.Model {
	return m.ExcludeInput
}

func (m *RulesModel) GetOlderInput() textinput.Model {
	return m.OlderInput
}

func (m *RulesModel) GetNewerInput() textinput.Model {
	return m.NewerInput
}
