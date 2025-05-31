package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rules "github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/errors"
	"github.com/pashkov256/deletor/internal/tui/help"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/utils"
)

// RulesModel represents the rules management page
type RulesModel struct {
	extensionsInput textinput.Model
	minSizeInput    textinput.Model
	locationInput   textinput.Model
	excludeInput    textinput.Model
	rules           rules.Rules
	focusIndex      int
	rulesPath       string
	Error           *errors.Error
}

// NewRulesModel creates a new rules management model
func NewRulesModel(rules rules.Rules) *RulesModel {
	// Initialize inputs
	currentRules, _ := rules.GetRules()
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

	locationInput := textinput.New()
	locationInput.Placeholder = "Target location (e.g.rules C:\\Users\\Downloads)"
	locationInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	locationInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	locationInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	locationInput.SetValue(currentRules.Path)

	excludeInput := textinput.New()
	excludeInput.Placeholder = "Exclude specific files/paths (e.g. data,backup)"
	excludeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	excludeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	excludeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	excludeInput.SetValue(strings.Join(currentRules.Exclude, ","))

	// Get AppData path
	rulesPath := filepath.Join(os.Getenv("APPDATA"), "deletor")

	return &RulesModel{
		extensionsInput: extensionsInput,
		minSizeInput:    minSizeInput,
		locationInput:   locationInput,
		excludeInput:    excludeInput,
		focusIndex:      0,
		rulesPath:       rulesPath,
		rules:           rules,
	}
}

func (m *RulesModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *RulesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, nil
		case "tab", "shift+tab", "up", "down":
			// Handle input focus cycling
			if msg.String() == "tab" || msg.String() == "down" {
				m.focusIndex = (m.focusIndex + 1) % 5 // Now includes save button
			} else if msg.String() == "shift+tab" || msg.String() == "up" {
				m.focusIndex = (m.focusIndex - 1 + 5) % 5 // Now includes save button
			}

			// Update input focus
			for i, input := range []*textinput.Model{
				&m.extensionsInput,
				&m.minSizeInput,
				&m.locationInput,
				&m.excludeInput,
			} {
				if i == m.focusIndex {
					input.Focus()
				} else {
					input.Blur()
				}
			}

			return m, nil
		case "enter":
			// Save button is focused
			if m.focusIndex == 4 {
				// Validate inputs before saving
				if err := m.validateInputs(); err != nil {
					return m, func() tea.Msg {
						return err
					}
				}

				// Clear error if validation passed
				m.Error = nil

				//save rules
				m.rules.UpdateRules(
					rules.WithPath(m.locationInput.Value()),
					rules.WithMinSize(m.minSizeInput.Value()),
					rules.WithMaxSize(""),
					rules.WithExtensions(utils.ParseExtToSlice(m.extensionsInput.Value())),
					rules.WithExclude(utils.ParseExcludeToSlice(m.excludeInput.Value())),
					rules.WithOlderThan(""),
					rules.WithNewerThan(""),
					rules.WithOptions(
						false, // showHidden
						false, // confirmDeletion
						false, // includeSubfolders
						false, // deleteEmptySubfolders
						false, // sendToTrash
						false, // logOps
						false, // logToFile
						false, // showStats
						false, // exitAfterDeletion
					),
				)

				m.rules.GetRulesPath()
				return m, nil
			}

			// Otherwise, move to next field
			if m.focusIndex < 4 {
				m.focusIndex++
				// Update input focus
				for i, input := range []*textinput.Model{
					&m.extensionsInput,
					&m.minSizeInput,
					&m.locationInput,
					&m.excludeInput,
				} {
					if i == m.focusIndex {
						input.Focus()
					} else {
						input.Blur()
					}
				}
			}

			return m, nil
		}
	case *errors.Error:
		m.Error = msg
		return m, nil
	}

	// Handle input updates
	var cmd tea.Cmd

	// Update the currently focused input
	switch m.focusIndex {
	case 0:
		m.extensionsInput, cmd = m.extensionsInput.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		m.minSizeInput, cmd = m.minSizeInput.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		m.locationInput, cmd = m.locationInput.Update(msg)
		cmds = append(cmds, cmd)
	case 3:
		m.excludeInput, cmd = m.excludeInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *RulesModel) validateInputs() *errors.Error {
	// Validate size input
	if m.minSizeInput.Value() != "" {
		if _, err := utils.ToBytes(m.minSizeInput.Value()); err != nil {
			return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid size format: %v", err))
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

func (m *RulesModel) View() string {
	var s strings.Builder

	// Title
	s.WriteString(styles.TitleStyle.Render(" Rule Management "))
	s.WriteString("\n\n")

	// Extensions input
	extStyle := styles.StandardInputStyle
	if m.focusIndex == 0 {
		extStyle = styles.StandardInputFocusedStyle
	}
	s.WriteString(extStyle.Render("Extensions: " + m.extensionsInput.View()))
	s.WriteString("\n")

	// Size input
	sizeStyle := styles.StandardInputStyle
	if m.focusIndex == 1 {
		sizeStyle = styles.StandardInputFocusedStyle
	}
	s.WriteString(sizeStyle.Render("Min Size: " + m.minSizeInput.View()))
	s.WriteString("\n")

	// Location input
	locStyle := styles.StandardInputStyle
	if m.focusIndex == 2 {
		locStyle = styles.StandardInputFocusedStyle
	}
	s.WriteString(locStyle.Render("Default path: " + m.locationInput.View()))
	s.WriteString("\n")

	excludeStyle := styles.StandardInputStyle
	if m.focusIndex == 3 {
		excludeStyle = styles.StandardInputFocusedStyle
	}
	s.WriteString(excludeStyle.Render("Exclude: " + m.excludeInput.View()))
	s.WriteString("\n\n")

	// Save button
	saveButtonStyle := styles.StandardButtonStyle
	if m.focusIndex == 4 {
		saveButtonStyle = styles.StandardButtonFocusedStyle
	}

	s.WriteString(saveButtonStyle.Render("💾 Save rules"))
	s.WriteString("\n\n")

	// AppData path
	s.WriteString(styles.PathStyle.Render(fmt.Sprintf("Rules are stored in: %s", m.rulesPath)))
	s.WriteString("\n\n")

	// Add error message if there is one
	if m.Error != nil && m.Error.IsVisible() {
		errorStyle := errors.GetStyle(m.Error.GetType())
		s.WriteString("\n")
		s.WriteString(errorStyle.Render(m.Error.GetMessage()))
	}

	s.WriteString("\n\n" + help.NavigateHelpText)
	return styles.AppStyle.Render(s.String())
}
