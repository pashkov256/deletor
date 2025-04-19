package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
			Background(lipgloss.Color("#1E90FF")).
			Padding(0, 1)

	sizeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E90FF"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(0, 1).
			Width(100)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6666")).
			Width(100)

	buttonFocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6666")).
			Background(lipgloss.Color("#FF6666")).
				Padding(0, 1).
			Width(100)

	dirButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1E90FF")).
			Width(100).
			Bold(true)

	dirButtonFocusedStyle = lipgloss.NewStyle().
	
	Background(lipgloss.Color("#1E90FF")).
			Foreground(lipgloss.Color("#fff")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1E90FF")).
			Width(100).
			Bold(true)

	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5"))

	selectedOptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ad58b3")).
			Bold(true)

	optionFocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5f5fd7")).
			Background(lipgloss.Color("#333333"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)
)

type item struct {
	path string
	size int64
}

func (i item) Title() string {
	if i.size == -1 {
		return "📂 .."
	}
	if i.size == 0 {
		return "📂 " + filepath.Base(i.path)
	}
	
	// Get filename and extension
	filename := filepath.Base(i.path)
	
	// Format size with fixed width
	sizeStr := formatSize(i.size)
	
	// Calculate padding to align size to the right
	padding := 50 - len(filename) // Adjust this value based on your needs
	if padding < 0 {
		padding = 0
	}
	
	return fmt.Sprintf("%s%s%s", filename, strings.Repeat(" ", padding), sizeStr)
}

func (i item) Description() string { return i.path }
func (i item) FilterValue() string { return i.path }

type model struct {
	list         list.Model
	extInput     textinput.Model
	sizeInput    textinput.Model
	pathInput    textinput.Model
	currentPath  string
	extensions   []string
	minSize     int64
	options      []string
	optionState  map[string]bool
	err          error
	focusedElement string // "path", "ext", "size", "button", "option1", "option2", "option3"
	waitingConfirmation bool
	fileToDelete  *item
	showDirs      bool
	dirList       list.Model
}

func initialModel(startDir string, extensions []string, minSize int64) model {
	extInput := textinput.New()
	extInput.Placeholder = "File extensions (e.g. js,png,zip)..."
	extInput.Focus()

	sizeInput := textinput.New()
	sizeInput.Placeholder = "File sizes (e.g. 10kb,10mb,10b)..."
	
	pathInput := textinput.New()
	pathInput.SetValue(startDir)
	
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	delegate.SetSpacing(0)
	
	items := []list.Item{}
	l := list.New(items, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)

	// Create directory list with same delegate
	dirList := list.New([]list.Item{}, delegate, 0, 0)
	dirList.SetShowHelp(false)
	dirList.SetShowStatusBar(false)
	dirList.SetFilteringEnabled(false)
	dirList.SetShowFilter(false)

	// Define options in fixed order
	options := []string{
		"Show hidden files",
		"Confirm deletion",
	}

	optionState := map[string]bool{
		"Show hidden files": false,
		"Confirm deletion": false,
	}

	return model{
		list:        l,
		extInput:    extInput,
		sizeInput:   sizeInput,
		pathInput:   pathInput,
		currentPath: startDir,
		extensions:  extensions,
		minSize:     minSize,
		options:     options,
		optionState: optionState,
		focusedElement: "path",
		waitingConfirmation: false,
		fileToDelete: nil,
		showDirs:     false,
		dirList:      dirList,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.loadFiles())
}

func (m model) loadFiles() tea.Cmd {
	return func() tea.Msg {
		var items []list.Item
		currentDir := m.currentPath

		// Add parent directory item
		parentDir := filepath.Dir(currentDir)
		if parentDir != currentDir {
			items = append(items, item{
				path: parentDir,
				size: -1, // Special value for parent directory
			})
		}

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

		// Create a channel for results
		results := make(chan item, 1000)
		done := make(chan bool)

		// Start a goroutine to collect results
		go func() {
			for item := range results {
				items = append(items, item)
			}
			done <- true
		}()

		// Walk through directory in a separate goroutine
		go func() {
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

				results <- item{
					path: path,
					size: info.Size(),
				}
				return nil
			})

			if err != nil {
				close(results)
				return
			}

			close(results)
		}()

		// Wait for collection to complete
		<-done

		// Sort items by name for consistent ordering
		sort.Slice(items, func(i, j int) bool {
			return items[i].(item).path < items[j].(item).path
		})

		// Update path input with current path
		m.pathInput.SetValue(m.currentPath)

		return items
	}
}

func (m model) loadDirs() tea.Cmd {
	return func() tea.Msg {
		var items []list.Item
		
		// Add parent directory with special display
		parentDir := filepath.Dir(m.currentPath)
		if parentDir != m.currentPath {
			items = append(items, item{
				path: parentDir,
				size: -1, // Special value for parent directory
			})
		}

		// Read current directory
		entries, err := os.ReadDir(m.currentPath)
		if err != nil {
			return err
		}

		// Create a channel for results
		results := make(chan item, 100)
		done := make(chan bool)

		// Start a goroutine to collect results
		go func() {
			for item := range results {
				items = append(items, item)
			}
			done <- true
		}()

		// Process entries in a separate goroutine
		go func() {
			for _, entry := range entries {
				if entry.IsDir() {
					// Skip hidden directories unless enabled
					if !m.optionState["Show hidden files"] && strings.HasPrefix(entry.Name(), ".") {
						continue
					}
					results <- item{
						path: filepath.Join(m.currentPath, entry.Name()),
						size: 0,
					}
				}
			}
			close(results)
		}()

		// Wait for collection to complete
		<-done

		// Sort directories by name
		sort.Slice(items, func(i, j int) bool {
			return items[i].(item).path < items[j].(item).path
		})

		// Update path input with current path
		m.pathInput.SetValue(m.currentPath)

		return items
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Fixed height for file/directory list
		listHeight := 8
		m.list.SetSize(msg.Width-2, listHeight)
		m.dirList.SetSize(msg.Width-2, listHeight)
		return m, nil

	case tea.KeyMsg:
		// Handle arrow keys first
		if msg.String() == "up" || msg.String() == "down" {
			if !m.showDirs {
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		// Handle Enter on file list first
		if msg.String() == "enter" && !m.showDirs && m.list.SelectedItem() != nil && m.focusedElement != "dirButton" && m.focusedElement != "button" {
			selectedItem := m.list.SelectedItem().(item)
			if selectedItem.size == -1 {
				// Handle parent directory selection
				m.currentPath = selectedItem.path
				m.pathInput.SetValue(selectedItem.path)
				return m, m.loadFiles()
			}
		}

		// Handle Ctrl+C first
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		// Handle directory selection first
		if m.showDirs {
			m.dirList, cmd = m.dirList.Update(msg)
			cmds = append(cmds, cmd)

			if msg.String() == "enter" && m.dirList.SelectedItem() != nil {
				selectedDir := m.dirList.SelectedItem().(item)
				m.currentPath = selectedDir.path
				m.pathInput.SetValue(selectedDir.path)
				m.showDirs = false
				return m, m.loadFiles()
			}
			if msg.String() == "esc" {
				m.showDirs = false
				return m, nil
			}
			return m, tea.Batch(cmds...)
		}

		// Handle input fields first to prevent unnecessary updates
		if m.pathInput.Focused() {
			switch msg.String() {
			case "tab", "esc":
				m.pathInput.Blur()
				m.extInput.Focus()
				m.focusedElement = "ext"
				return m, nil
			case "enter":
				newPath := m.pathInput.Value()
				if _, err := os.Stat(newPath); err == nil {
					m.currentPath = newPath
					return m, m.loadFiles()
				}
				m.err = fmt.Errorf("invalid path: %s", newPath)
				return m, nil
			default:
				m.pathInput, cmd = m.pathInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		if m.extInput.Focused() {
			switch msg.String() {
			case "tab", "esc":
				m.extInput.Blur()
				m.sizeInput.Focus()
				m.focusedElement = "size"
				return m, nil
			default:
				m.extInput, cmd = m.extInput.Update(msg)
				cmds = append(cmds, cmd)
				// Only reload files when input actually changes
				if msg.String() != "tab" && msg.String() != "esc" {
					cmds = append(cmds, m.loadFiles())
				}
				return m, tea.Batch(cmds...)
			}
		}

		if m.sizeInput.Focused() {
			switch msg.String() {
			case "tab":
				m.sizeInput.Blur()
				m.focusedElement = "option1"
				return m, nil
			case "esc":
				m.sizeInput.Blur()
				m.focusedElement = "path"
				m.pathInput.Focus()
				return m, nil
			default:
				m.sizeInput, cmd = m.sizeInput.Update(msg)
				cmds = append(cmds, cmd)
				// Only reload files when input actually changes
				if msg.String() != "tab" && msg.String() != "esc" {
					cmds = append(cmds, m.loadFiles())
				}
				return m, tea.Batch(cmds...)
			}
		}

		switch msg.String() {
		case "tab":
			// Cycle through focusable elements without reloading
			switch m.focusedElement {
			case "path":
				m.pathInput.Blur()
				m.extInput.Focus()
				m.focusedElement = "ext"
			case "ext":
				m.extInput.Blur()
				m.sizeInput.Focus()
				m.focusedElement = "size"
			case "size":
				m.sizeInput.Blur()
				m.focusedElement = "option1"
			case "option1":
				m.focusedElement = "option2"
			case "option2":
				m.focusedElement = "dirButton"
			case "dirButton":
				m.focusedElement = "button"
			case "button":
				m.pathInput.Focus()
				m.focusedElement = "path"
			}
			return m, nil
		case "enter":
			switch m.focusedElement {
			case "dirButton":
				m.showDirs = true
				return m, m.loadDirs()
			case "button":
				if m.list.SelectedItem() != nil {
					selectedItem := m.list.SelectedItem().(item)
					if selectedItem.size == -1 {
						// Handle parent directory selection
						m.currentPath = selectedItem.path
						m.pathInput.SetValue(selectedItem.path)
						return m, m.loadFiles()
					}
					if !m.optionState["Confirm deletion"] {
						// If confirm deletion is disabled, delete all files
						for _, listItem := range m.list.Items() {
							if fileItem, ok := listItem.(item); ok && fileItem.size != -1 {
								err := os.Remove(fileItem.path)
								if err != nil {
									m.err = err
								}
							}
						}
						return m, m.loadFiles()
					}
					err := os.Remove(selectedItem.path)
					if err != nil {
						m.err = err
					}
					return m, m.loadFiles()
				}
			case "option1", "option2":
				idx := int(m.focusedElement[len(m.focusedElement)-1] - '1')
				if idx >= 0 && idx < len(m.options) {
					optName := m.options[idx]
					m.optionState[optName] = !m.optionState[optName]
					return m, m.loadFiles()
				}
			}
		case "esc":
			// Remove focus from inputs
			m.pathInput.Blur()
			m.extInput.Blur()
			m.sizeInput.Blur()
			m.focusedElement = "path"
			m.pathInput.Focus()
			return m, nil
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			if m.focusedElement == "dirButton" {
				m.showDirs = true
				return m, m.loadDirs()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			if m.focusedElement == "button" {
				if m.list.SelectedItem() != nil {
					selectedItem := m.list.SelectedItem().(item)
					if selectedItem.size == -1 {
						// Handle parent directory selection
						m.currentPath = selectedItem.path
						m.pathInput.SetValue(selectedItem.path)
						return m, m.loadFiles()
					}
					if !m.optionState["Confirm deletion"] {
						// If confirm deletion is disabled, delete all files
						for _, listItem := range m.list.Items() {
							if fileItem, ok := listItem.(item); ok && fileItem.size != -1 {
								err := os.Remove(fileItem.path)
								if err != nil {
									m.err = err
								}
							}
						}
						return m, m.loadFiles()
					}
					err := os.Remove(selectedItem.path)
					if err != nil {
						m.err = err
					}
					return m, m.loadFiles()
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("1", "2"))):
			// Toggle options with number keys only when inputs are not focused
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(m.options) {
				optName := m.options[idx]
				m.optionState[optName] = !m.optionState[optName]
				return m, m.loadFiles()
			}
		}
	
	case []list.Item:
		if m.showDirs {
			m.dirList.SetItems(msg)
		} else {
			// Preserve selection when updating items
			selectedIdx := m.list.Index()
			m.list.SetItems(msg)
			if selectedIdx < len(msg) {
				m.list.Select(selectedIdx)
			}
		}
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s strings.Builder

	// Path input field
	pathStyle := borderStyle.Copy()
	if m.focusedElement == "path" {
		s.WriteString(pathStyle.Render(m.pathInput.View()))
	} else {
		s.WriteString(pathStyle.Render(m.pathInput.View()))
	}
	s.WriteString("\n\n")

	// File or directory list with border
	if m.showDirs {
		s.WriteString(titleStyle.Render(" Available Directories "))
		s.WriteString("\n")
		s.WriteString(borderStyle.Render(m.dirList.View()))
	} else {
		s.WriteString(borderStyle.Render(m.list.View()))
	}
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
		s.WriteString(style.Render(fmt.Sprintf("[%s] %-20s", map[bool]string{true: "✓", false: "○"}[m.optionState[name]], name)))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Change Directory button
	dirButtonText := "➡️ Change Directory"
	if m.focusedElement == "dirButton" {
		s.WriteString(dirButtonFocusedStyle.Copy().Width(100).Render(dirButtonText))
	} else {
		s.WriteString(dirButtonStyle.Copy().Width(100).Render(dirButtonText))
	}
	s.WriteString("\n")

	// Delete button
	buttonText := "🗑️ Delete Selected File"
	if m.focusedElement == "button" {
		s.WriteString(buttonFocusedStyle.Copy().Width(100).Render(buttonText))
	} else {
		s.WriteString(buttonStyle.Copy().Width(100).Render(buttonText))
	}
	s.WriteString("\n\n")

	// Help
	s.WriteString("Press 'tab' to navigate • 'enter' to select • '↑↓' move through files • 'esc' to reset focus • 'ctrl+c' to quit")

	// Error
	if m.err != nil {
		s.WriteString(fmt.Sprintf("\nError: %v", m.err))
	}

	return appStyle.Render(s.String())
}

func startTUI(startDir string, extensions []string, minSize int64) error {
	p := tea.NewProgram(initialModel(startDir, extensions, minSize), 
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithFPS(30),
		tea.WithInputTTY(),
	)
	_, err := p.Run()
	return err
} 