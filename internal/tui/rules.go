package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rules "github.com/pashkov256/deletor/internal/rules"
)

// RulesModel represents the rules management page
type RulesModel struct {
	extensionsInput textinput.Model
	sizeInput       textinput.Model
	locationInput   textinput.Model
	excludeInput    textinput.Model
	rules           rules.Rules
	focusIndex      int
	rulesPath       string
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

	sizeInput := textinput.New()
	sizeInput.Placeholder = "Minimum file size (e.g. 10kb)"
	sizeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	sizeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	sizeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	sizeInput.SetValue(currentRules.MinSize)

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
		sizeInput:       sizeInput,
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
				&m.sizeInput,
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
				// m.rules.Extensions = strings.Split(m.extensionsInput.Value(), ",")
				// m.rules.Path = m.locationInput.Value()
				// m.rules.MinSize = m.sizeInput.Value()
				// m.rules.Exclude = strings.Split(m.excludeInput.Value(), ",")

				// m.excludeInput.SetValue(strings.Join(m.rules.Exclude, ","))
				// m.locationInput.SetValue(m.rules.Path)
				// m.sizeInput.SetValue(m.rules.MinSize)
				// m.extensionsInput.SetValue(strings.Join(m.rules.Extensions, ","))

				m.rules.UpdateRules(
					m.locationInput.Value(),
					m.sizeInput.Value(),
					strings.Split(m.extensionsInput.Value(), ","),
					strings.Split(m.excludeInput.Value(), ","),
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
					&m.sizeInput,
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
	}

	// Handle input updates
	var cmd tea.Cmd

	// Update the currently focused input
	switch m.focusIndex {
	case 0:
		m.extensionsInput, cmd = m.extensionsInput.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		m.sizeInput, cmd = m.sizeInput.Update(msg)
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

func (m *RulesModel) View() string {
	var s strings.Builder

	// Title
	s.WriteString(TitleStyle.Render(" Rule Management "))
	s.WriteString("\n\n")

	// Extensions input
	extStyle := StandardInputStyle
	if m.focusIndex == 0 {
		extStyle = StandardInputFocusedStyle
	}
	s.WriteString(extStyle.Render("Extensions: " + m.extensionsInput.View()))
	s.WriteString("\n")

	// Size input
	sizeStyle := StandardInputStyle
	if m.focusIndex == 1 {
		sizeStyle = StandardInputFocusedStyle
	}
	s.WriteString(sizeStyle.Render("Min Size: " + m.sizeInput.View()))
	s.WriteString("\n")

	// Location input
	locStyle := StandardInputStyle
	if m.focusIndex == 2 {
		locStyle = StandardInputFocusedStyle
	}
	s.WriteString(locStyle.Render("Default path: " + m.locationInput.View()))
	s.WriteString("\n")

	excludeStyle := StandardInputStyle
	if m.focusIndex == 3 {
		excludeStyle = StandardInputFocusedStyle
	}
	s.WriteString(excludeStyle.Render("Exclude: " + m.excludeInput.View()))
	s.WriteString("\n\n")

	// Save button
	saveButtonStyle := StandardButtonStyle
	if m.focusIndex == 4 {
		saveButtonStyle = StandardButtonFocusedStyle
	}

	s.WriteString(saveButtonStyle.Render("💾 Save rules"))
	s.WriteString("\n\n")

	// Help text
	s.WriteString("Tab: cycle fields • Enter: save rule • Esc: return to menu\n\n")

	// AppData path
	s.WriteString(PathStyle.Render(fmt.Sprintf("Rules are stored in: %s", m.rulesPath)))

	return AppStyle.Render(s.String())
}
