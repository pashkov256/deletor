package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	sizeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#25A065"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(0, 1).
			Width(100)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF4444")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF0000")).
			Padding(0, 1).
			Width(100)

	buttonFocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF6666")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF0000")).
			Padding(0, 1).
			Width(100)

	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5"))

	selectedOptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#25A065")).
			Bold(true)

	optionFocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#25A065")).
			Background(lipgloss.Color("#333333"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)
)

type item struct {
	path string
	size int64
}

func (i item) Title() string       { return fmt.Sprintf("%s (%s)", filepath.Base(i.path), formatSize(i.size)) }
func (i item) Description() string { return i.path }
func (i item) FilterValue() string { return i.path }

type model struct {
	list         list.Model
	extInput     textinput.Model
	sizeInput    textinput.Model
	currentPath  string
	extensions   []string
	minSize     int64
	options      []string
	optionState  map[string]bool
	err          error
	focusedElement string // "ext", "size", "button", "option1", "option2", "option3"
	waitingConfirmation bool
	fileToDelete  *item
}

func initialModel(startDir string, extensions []string, minSize int64) model {
	extInput := textinput.New()
	extInput.Placeholder = "File extensions (e.g. js,png,zip)..."
	extInput.Focus()

	sizeInput := textinput.New()
	sizeInput.Placeholder = "File sizes (e.g. 10kb,10mb,10b)..."
	
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	delegate.SetSpacing(0)
	
	items := []list.Item{}
	l := list.New(items, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)

	// Define options in fixed order
	options := []string{
		"Show hidden files",
		"Confirm deletion",
		"Show progress",
	}

	optionState := map[string]bool{
		"Show hidden files": false,
		"Confirm deletion": false,
		"Show progress":    true,
	}

	return model{
		list:        l,
		extInput:    extInput,
		sizeInput:   sizeInput,
		currentPath: startDir,
		extensions:  extensions,
		minSize:     minSize,
		options:     options,
		optionState: optionState,
		focusedElement: "ext",
		waitingConfirmation: false,
		fileToDelete: nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.loadFiles())
}

func (m model) loadFiles() tea.Cmd {
	return func() tea.Msg {
		var items []list.Item
		currentDir := m.currentPath

		// Parse extensions from input
		extensions := strings.Split(m.extInput.Value(), ",")
		for i := range extensions {
			extensions[i] = strings.TrimSpace(extensions[i])
		}

		// Parse size from input
		var minSize int64
		if m.sizeInput.Value() != "" {
			sizeStr := strings.TrimSpace(m.sizeInput.Value())
			sizeBytes, err := toBytes(sizeStr)
			if err == nil {
				minSize = sizeBytes
			}
		}

		err := filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Skip hidden files unless enabled
			if !m.optionState["Show hidden files"] && strings.HasPrefix(filepath.Base(path), ".") {
				return nil
			}

			// Check file size if specified
			if minSize > 0 && info.Size() < minSize {
				return nil
			}

			// Check file extension
			if len(extensions) > 0 && extensions[0] != "" {
				ext := strings.TrimPrefix(filepath.Ext(path), ".")
				found := false
				for _, allowedExt := range extensions {
					if strings.EqualFold(ext, strings.TrimSpace(allowedExt)) {
						found = true
						break
					}
				}
				if !found {
					return nil
				}
			}

			items = append(items, item{
				path: path,
				size: info.Size(),
			})
			return nil
		})

		if err != nil {
			return err
		}

		return items
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Set list height to 40% of terminal height or less if there's not enough space
		listHeight := (msg.Height - 15) / 3 // Leave space for inputs and options
		if listHeight < 5 {
			listHeight = 5 // Minimum height for list
		}
		m.list.SetSize(msg.Width-2, listHeight) // -2 for minimal padding
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			// Cycle through focusable elements
			switch m.focusedElement {
			case "ext":
				m.extInput.Blur()
				m.sizeInput.Focus()
				m.focusedElement = "size"
			case "size":
				m.sizeInput.Blur()
				m.focusedElement = "button"
			case "button":
				m.focusedElement = "option1"
			case "option1":
				m.focusedElement = "option2"
			case "option2":
				m.focusedElement = "option3"
			case "option3":
				m.extInput.Focus()
				m.focusedElement = "ext"
			}
			return m, nil
		case "enter":
			switch m.focusedElement {
			case "button":
				if m.list.SelectedItem() != nil {
					selectedItem := m.list.SelectedItem().(item)
					err := os.Remove(selectedItem.path)
					if err != nil {
						m.err = err
					}
					return m, m.loadFiles()
				}
			case "option1", "option2", "option3":
				idx := int(m.focusedElement[len(m.focusedElement)-1] - '1')
				if idx >= 0 && idx < len(m.options) {
					optName := m.options[idx]
					wasEnabled := m.optionState[optName]
					m.optionState[optName] = !wasEnabled
					
					// If we're toggling Confirm deletion, reset all confirmation state
					if optName == "Confirm deletion" {
						m.waitingConfirmation = false
						m.fileToDelete = nil
						m.err = nil
					}
					
					return m, m.loadFiles()
				}
			}
		}

		// Handle input fields first
		if m.extInput.Focused() {
			m.extInput, cmd = m.extInput.Update(msg)
			cmds = append(cmds, cmd)
			cmds = append(cmds, m.loadFiles())
			return m, tea.Batch(cmds...)
		}
		if m.sizeInput.Focused() {
			m.sizeInput, cmd = m.sizeInput.Update(msg)
			cmds = append(cmds, cmd)
			cmds = append(cmds, m.loadFiles())
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			if m.focusedElement == "button" {
				if m.list.SelectedItem() != nil {
					selectedItem := m.list.SelectedItem().(item)
					err := os.Remove(selectedItem.path)
					if err != nil {
						m.err = err
					}
					return m, m.loadFiles()
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("1", "2", "3"))):
			// Toggle options with number keys only when inputs are not focused
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(m.options) {
				optName := m.options[idx]
				m.optionState[optName] = !m.optionState[optName]
				m.waitingConfirmation = false
				m.fileToDelete = nil
				return m, m.loadFiles()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			// Remove focus from inputs
			m.extInput.Blur()
			m.sizeInput.Blur()
			m.focusedElement = "ext"
			m.extInput.Focus()
			return m, nil
		}
	
	case []list.Item:
		m.list.SetItems(msg)
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s strings.Builder

	// Current path info
	s.WriteString(titleStyle.Render(fmt.Sprintf(" Current path: %s ", m.currentPath)))
	s.WriteString("\n\n")

	// File list with border
	s.WriteString(borderStyle.Render(m.list.View()))
	s.WriteString("\n\n")

	// Input fields
	s.WriteString(borderStyle.Render(m.extInput.View()))
	s.WriteString("\n")
	s.WriteString(borderStyle.Render(m.sizeInput.View()))
	s.WriteString("\n\n")

	// Options at the bottom
	s.WriteString("Options:\n")
	for i, name := range m.options {
		style := optionStyle
		if m.optionState[name] {
			style = selectedOptionStyle
		}
		if m.focusedElement == fmt.Sprintf("option%d", i+1) {
			style = optionFocusedStyle
		}
		s.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", i+1)))
		s.WriteString(style.Render(fmt.Sprintf("[%s] %-20s", map[bool]string{true: "x", false: " "}[m.optionState[name]], name)))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Delete button
	if m.list.SelectedItem() != nil {
		buttonText := "Delete Selected File"
		if m.focusedElement == "button" {
			s.WriteString(buttonFocusedStyle.Render(buttonText))
		} else {
			s.WriteString(buttonStyle.Render(buttonText))
		}
		s.WriteString("\n\n")
	}

	// Help
	s.WriteString("Press 'tab' to navigate • 'enter' to select • 'esc' to reset focus • 'ctrl+c' or 'q' to quit")

	// Error
	if m.err != nil {
		s.WriteString(fmt.Sprintf("\nError: %v", m.err))
	}

	return appStyle.Render(s.String())
}

func startTUI(startDir string, extensions []string, minSize int64) error {
	p := tea.NewProgram(initialModel(startDir, extensions, minSize), tea.WithAltScreen())
	_, err := p.Run()
	return err
} 