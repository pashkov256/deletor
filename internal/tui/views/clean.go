package views

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/models"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/tui/tabs"
	"github.com/pashkov256/deletor/internal/utils"
)

type CleanFilesModel struct {
	List            list.Model
	ExtInput        textinput.Model
	SizeInput       textinput.Model
	PathInput       textinput.Model
	ExcludeInput    textinput.Model
	CurrentPath     string
	Extensions      []string
	MinSize         int64
	Exclude         []string
	Options         []string
	OptionState     map[string]bool
	Err             error
	FocusedElement  string // "path", "ext", "size", "button", "option1", "option2", "option3"
	FileToDelete    *models.CleanItem
	ShowDirs        bool
	DirList         list.Model
	DirSize         int64 // Cached directory size
	CalculatingSize bool  // Flag to indicate size calculation in progress
	FilteredSize    int64 // Total size of filtered files
	FilteredCount   int   // Count of filtered files
	ActiveTab       int   // 0 for files, 1 for exclude, 2 for options, 3 for hot keys
	Rules           rules.Rules
	Filemanager     filemanager.FileManager
	TabManager      *tabs.CleanTabManager
}

// Message for directory size updates
type dirSizeMsg struct {
	size int64
}

// Message for filtered files size updates
type filteredSizeMsg struct {
	size  int64
	count int
}

func InitialCleanModel(rules rules.Rules) *CleanFilesModel {
	// Create a temporary model to get rules
	lastestRules, _ := rules.GetRules()
	latestDir := lastestRules.Path
	latestExtensions := lastestRules.Extensions
	latestMinSize := lastestRules.MinSize

	latestExclude := lastestRules.Exclude
	// Initialize inputs
	extInput := textinput.New()
	extInput.Placeholder = "(e.g. js,png,zip)..."
	extInput.SetValue(strings.Join(latestExtensions, ","))
	extInput.PromptStyle = styles.TextInputPromptStyle
	extInput.TextStyle = styles.TextInputTextStyle
	extInput.Cursor.Style = styles.TextInputCursorStyle

	sizeInput := textinput.New()
	sizeInput.Placeholder = "(e.g. 10kb,10mb,10b)..."
	sizeInput.SetValue(latestMinSize)
	minSize, _ := utils.ToBytes(latestMinSize)
	sizeInput.PromptStyle = styles.TextInputPromptStyle
	sizeInput.TextStyle = styles.TextInputTextStyle
	sizeInput.Cursor.Style = styles.TextInputCursorStyle

	pathInput := textinput.New()
	pathInput.SetValue(latestDir)
	pathInput.PromptStyle = styles.TextInputPromptStyle
	pathInput.TextStyle = styles.TextInputTextStyle
	pathInput.Cursor.Style = styles.TextInputCursorStyle

	excludeInput := textinput.New()
	excludeInput.Placeholder = "Exclude specific files/paths (e.g. data,backup)"
	excludeInput.SetValue(strings.Join(latestExclude, ","))
	excludeInput.PromptStyle = styles.TextInputPromptStyle
	excludeInput.TextStyle = styles.TextInputTextStyle
	excludeInput.Cursor.Style = styles.TextInputCursorStyle

	// Create a proper delegate with visible height
	delegate := list.NewDefaultDelegate()

	delegate.SetHeight(1)
	delegate.SetSpacing(1)
	delegate.ShowDescription = false

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#0066ff")).
		Bold(true)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("#dddddd"))

	l := list.New([]list.Item{}, delegate, 30, 10)
	l.SetShowTitle(true)
	l.Title = "Files"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = styles.TitleStyle

	// Create directory list with same delegate
	dirList := list.New([]list.Item{}, delegate, 30, 10)
	dirList.SetShowTitle(true)
	dirList.Title = "Directories"
	dirList.SetShowStatusBar(true)
	dirList.SetFilteringEnabled(false)
	dirList.SetShowHelp(false)
	dirList.Styles.Title = styles.TitleStyle

	return &CleanFilesModel{
		List:            l,
		ExtInput:        extInput,
		SizeInput:       sizeInput,
		PathInput:       pathInput,
		ExcludeInput:    excludeInput,
		CurrentPath:     latestDir,
		Extensions:      latestExtensions,
		MinSize:         minSize,
		Exclude:         latestExclude,
		FocusedElement:  "list",
		ShowDirs:        false,
		DirList:         dirList,
		DirSize:         0,
		CalculatingSize: false,
		FilteredSize:    0,
		FilteredCount:   0,
		ActiveTab:       0,
		Rules:           rules,
	}
}

func (m *CleanFilesModel) Init() tea.Cmd {
	// Set initial focus to path input
	m.FocusedElement = "path"
	m.PathInput.Focus()
	m.TabManager = tabs.NewCleanTabManager(m, tabs.NewCleanTabFactory())
	// If we have a path, load files and calculate size
	if m.CurrentPath != "" {
		return tea.Batch(m.LoadFiles(), m.calculateDirSizeAsync())
	}

	// Otherwise just return the blink command for the path input
	return textinput.Blink
}

func Run(filemanager filemanager.FileManager, rules rules.Rules) error {
	p := tea.NewProgram(InitialCleanModel(rules),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithFPS(30),
		tea.WithInputTTY(),
		tea.WithOutput(os.Stderr),
	)
	_, err := p.Run()
	return err
}

func (m *CleanFilesModel) View() string {
	// --- Tabs rendering ---
	activeTab := m.TabManager.GetActiveTabIndex()
	tabNames := []string{"🗂️ [F1] Main", "🧹 [F2] Filters", "⚙️ [F3] Options", "❔ [F4] Help"}
	tabs := make([]string, 4)
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

	content.WriteString(m.TabManager.GetActiveTab().View())

	// Combine everything
	var ui string
	if activeTab != 0 {
		// For Main tab, show only the content
		ui = content.String()
	} else {
		// For other tabs, show content with hot keys
		ui = lipgloss.JoinVertical(lipgloss.Left,
			content.String(),
			"Arrow keys: navigate in list • Tab: cycle focus • Shift+Tab: focus back • Enter: select/confirm • Esc: back to list",
			"Ctrl+R: refresh • Ctrl+D: delete files • Ctrl+O: open in explorer • Ctrl+C: quit",
			"Left/Right arrow keys: switch tabs",
		)
	}

	if m.Err != nil {
		ui += "\n" + styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.Err))
	}

	return styles.AppStyle.Render(ui)
}

func (m *CleanFilesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	activeTab := m.TabManager.GetActiveTabIndex()
	switch msg := msg.(type) {
	case dirSizeMsg:
		// Update the directory size
		m.DirSize = msg.size
		return m, nil

	case tea.WindowSizeMsg:
		// Properly set both width and height
		h, v := styles.AppStyle.GetFrameSize()
		// Further reduce listHeight by another 10% (now at 65% of original)
		listHeight := (msg.Height - v - 15) * 65 / 100 // Reserve space for other UI elements and reduce by 35%
		if listHeight < 5 {
			listHeight = 5 // Minimum height to show something
		}
		m.List.SetSize(msg.Width-h, listHeight)
		m.DirList.SetSize(msg.Width-h, listHeight)

		cmds = append(cmds, m.LoadFiles())
		// Trigger directory size calculation when changing directory
		cmds = append(cmds, m.calculateDirSizeAsync())
		return m, tea.Batch(cmds...)

	// Handle message for setting items in the list
	case []list.Item:
		if m.ShowDirs {
			m.DirList.SetItems(msg)
		} else {
			// Preserve selection when updating items
			selectedIdx := m.List.Index()
			m.List.SetItems(msg)
			if selectedIdx < len(msg) {
				m.List.Select(selectedIdx)
			}
		}
		return m, nil

	case error:
		m.Err = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+d":
			return m.OnDelete()
		case "tab", "shift+tab":
			// Handle tab navigation based on active tab
			if activeTab == 0 {
				if m.CurrentPath == "" {
					// When no path is set, only allow focusing between path input and start button
					if msg.String() == "tab" {
						if m.FocusedElement == "path" {
							m.FocusedElement = "startButton"
							m.PathInput.Blur()
						} else {
							m.FocusedElement = "path"
							m.PathInput.Focus()
						}
					} else {
						if m.FocusedElement == "path" {
							m.FocusedElement = "startButton"
							m.PathInput.Blur()
						} else {
							m.FocusedElement = "path"
							m.PathInput.Focus()
						}
					}
					return m, nil
				}
				// Normal tab behavior when path is set
				if msg.String() == "tab" {
					switch m.FocusedElement {
					case "path":
						m.FocusedElement = "ext"
						m.PathInput.Blur()
						m.ExtInput.Focus()
					case "ext":
						m.ExtInput.Blur()
						m.FocusedElement = "size"
						m.SizeInput.Focus()
					case "size":
						m.SizeInput.Blur()
						m.FocusedElement = "list"
					case "list":
						m.FocusedElement = "dirButton"
					case "dirButton":
						m.FocusedElement = "button"
					case "button":
						m.FocusedElement = "path"
						m.PathInput.Focus()
					}
				} else {
					switch m.FocusedElement {
					case "path":
						m.PathInput.Blur()
						m.FocusedElement = "button"
					case "button":
						m.FocusedElement = "dirButton"
					case "dirButton":
						m.FocusedElement = "list"
					case "list":
						m.FocusedElement = "size"
						m.SizeInput.Focus()
					case "size":
						m.SizeInput.Blur()
						m.FocusedElement = "ext"
						m.ExtInput.Focus()
					case "ext":
						m.ExtInput.Blur()
						m.FocusedElement = "path"
						m.PathInput.Focus()
					}
				}
			} else if activeTab == 1 {
				// Tab navigation for Filters tab
				if msg.String() == "tab" {
					switch m.FocusedElement {
					case "exclude":
						m.ExcludeInput.Blur()
						m.FocusedElement = "size"
						m.SizeInput.Focus()
					case "size":
						m.SizeInput.Blur()
						m.FocusedElement = "exclude"
						m.ExcludeInput.Focus()
					}
				} else {
					switch m.FocusedElement {
					case "exclude":
						m.ExcludeInput.Blur()
						m.FocusedElement = "size"
						m.SizeInput.Focus()
					case "size":
						m.SizeInput.Blur()
						m.FocusedElement = "exclude"
						m.ExcludeInput.Focus()
					}
				}
			} else if activeTab == 2 {
				// Tab navigation for Options tab
				if msg.String() == "tab" {
					switch m.FocusedElement {
					case "option1":
						m.FocusedElement = "option2"
					case "option2":
						m.FocusedElement = "option3"
					case "option3":
						m.FocusedElement = "option4"
					case "option4":
						m.FocusedElement = "option1"
					}
				} else {
					switch m.FocusedElement {
					case "option1":
						m.FocusedElement = "option4"
					case "option2":
						m.FocusedElement = "option1"
					case "option3":
						m.FocusedElement = "option2"
					case "option4":
						m.FocusedElement = "option3"
					}
				}
			} else if activeTab == 3 {
				// Hot Keys tab navigation
				if msg.String() == "tab" {
					switch m.FocusedElement {
					case "hotKeys":
						m.FocusedElement = "navigation"
					case "navigation":
						m.FocusedElement = "fileOperations"
					case "fileOperations":
						m.FocusedElement = "filterOperations"
					case "filterOperations":
						m.FocusedElement = "directoryNavigation"
					case "directoryNavigation":
						m.FocusedElement = "options"
					case "options":
						m.FocusedElement = "hotKeys"
					}
				} else {
					switch m.FocusedElement {
					case "hotKeys":
						m.FocusedElement = "options"
					case "options":
						m.FocusedElement = "directoryNavigation"
					case "directoryNavigation":
						m.FocusedElement = "fileOperations"
					case "fileOperations":
						m.FocusedElement = "navigation"
					case "navigation":
						m.FocusedElement = "hotKeys"
					}
				}
			}
			return m, nil
		case "enter":
			if m.CurrentPath == "" {
				if m.FocusedElement == "startButton" {
					// Validate and set the path
					path := m.PathInput.Value()
					if path != "" {
						// Expand tilde in path
						expandedPath := utils.ExpandTilde(path)
						if _, err := os.Stat(expandedPath); err == nil {
							m.CurrentPath = expandedPath

							// Load files for the new path
							cmds = append(cmds, m.LoadFiles(), m.calculateDirSizeAsync())

							// Set focus to path input
							m.FocusedElement = "path"
							m.PathInput.Focus()
						} else {
							m.Err = fmt.Errorf("invalid path: %s", path)
						}
					}
				}
			} else {
				switch m.FocusedElement {
				case "ext", "size", "exclude":
					// Обновляем список файлов при нажатии Enter на полях ввода
					return m, m.LoadFiles()
				case "list":
					if !m.ShowDirs && m.List.SelectedItem() != nil {
						selectedItem := m.List.SelectedItem().(models.CleanItem)
						if selectedItem.Size == -1 {
							// Handle parent directory selection
							m.CurrentPath = selectedItem.Path
							m.PathInput.SetValue(selectedItem.Path)
							// Recalculate directory size when changing directory
							cmds = append(cmds, m.LoadFiles(), m.calculateDirSizeAsync())
							return m, tea.Batch(cmds...)
						}
						// If it's a directory, navigate into it
						info, err := os.Stat(selectedItem.Path)
						if err == nil && info.IsDir() {
							m.CurrentPath = selectedItem.Path
							m.PathInput.SetValue(selectedItem.Path)
							// Recalculate directory size when changing directory
							cmds = append(cmds, m.LoadFiles(), m.calculateDirSizeAsync())
							return m, tea.Batch(cmds...)
						}
					} else if m.ShowDirs && m.DirList.SelectedItem() != nil {
						selectedDir := m.DirList.SelectedItem().(models.CleanItem)
						m.CurrentPath = selectedDir.Path
						m.PathInput.SetValue(selectedDir.Path)
						m.ShowDirs = false
						// Recalculate directory size when changing directory
						cmds = append(cmds, m.LoadFiles(), m.calculateDirSizeAsync())
						return m, tea.Batch(cmds...)
					}
				case "dirButton":
					m.ShowDirs = true
					return m, m.loadDirs()
				case "button":
					if activeTab == 0 {
						return m.OnDelete()
					}
					return m, nil
				case "option1", "option2", "option3", "option4":
					idx := 0
					if m.FocusedElement == "option2" {
						idx = 1
					}
					if m.FocusedElement == "option3" {
						idx = 2
					}
					if m.FocusedElement == "option4" {
						idx = 3
					}
					if idx < len(m.Options) {
						optName := m.Options[idx]
						m.OptionState[optName] = !m.OptionState[optName]
						m.FocusedElement = "option" + fmt.Sprintf("%d", idx+1)

						return m, nil
					}
				}
			}
		}

		// Global hotkeys that work regardless of focus
		switch msg.String() {
		case "ctrl+r": // Refresh files
			return m, m.LoadFiles()
		case "ctrl+d": // Delete files
			return m.OnDelete()
		case "ctrl+o": // Open current directory in file explorer
			cmd := openFileExplorer(m.CurrentPath)
			return m, cmd
		case "alt+c": // Clear filters
			m.SizeInput.SetValue("")
			m.ExcludeInput.SetValue("")
			return m, m.LoadFiles()
		case "alt+1": // Toggle hidden files
			m.OptionState["Show hidden files"] = !m.OptionState["Show hidden files"]
			return m, m.LoadFiles()
		case "alt+2": // Toggle confirm deletion
			m.OptionState["Confirm deletion"] = !m.OptionState["Confirm deletion"]
			return m, nil
		case "alt+3": // Toggle include subfolders
			m.OptionState["Include subfolders"] = !m.OptionState["Include subfolders"]
			return m, nil
		case "alt+4": // Toggle delete empty subfolders
			m.OptionState["Delete empty subfolders"] = !m.OptionState["Delete empty subfolders"]
			return m, nil
		case "left", "right": // Tab switching
			if msg.String() == "left" && m.ActiveTab > 0 {
				m.ActiveTab--
				m.TabManager.SetActiveTabIndex(m.ActiveTab - 1)
				if m.ActiveTab == 1 {
					m.ExcludeInput.Focus()
					m.FocusedElement = "exclude"
				}
			}
			if msg.String() == "right" && m.ActiveTab < 3 {
				m.ActiveTab++
				m.TabManager.SetActiveTabIndex(m.ActiveTab + 1)
				if m.ActiveTab == 1 {
					m.ExcludeInput.Focus()
					m.FocusedElement = "exclude"
				}
			}
			return m, nil
		case "f1":
			m.ActiveTab = 0
			m.TabManager.SetActiveTabIndex(0)
			m.FocusedElement = "path"
			m.PathInput.Focus()
			return m, nil
		case "f2":
			m.ActiveTab = 1
			m.TabManager.SetActiveTabIndex(1)
			m.FocusedElement = "exclude"
			m.ExcludeInput.Focus()
			m.PathInput.Blur()
			m.ExtInput.Blur()
			m.SizeInput.Blur()
			return m, nil
		case "f3":
			m.ActiveTab = 2
			m.TabManager.SetActiveTabIndex(2)
			m.FocusedElement = "option1"
			return m, nil
		case "f4":
			m.TabManager.SetActiveTabIndex(3)
			m.ActiveTab = 3
			return m, nil
		case "up", "down": // Always handle arrow keys for list navigation regardless of focus
			// Make list navigation global - arrow keys always navigate the list
			if !m.ShowDirs {
				m.List, cmd = m.List.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.DirList, cmd = m.DirList.Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Handle space key for options
		if (msg.String() == " " || msg.String() == "enter") && m.ActiveTab == 2 {
			if m.FocusedElement == "option1" || m.FocusedElement == "option2" || m.FocusedElement == "option3" || m.FocusedElement == "option4" {
				idx := int(m.FocusedElement[len(m.FocusedElement)-1] - '1')
				if idx >= 0 && idx < len(m.Options) {
					optName := m.Options[idx]
					m.OptionState[optName] = !m.OptionState[optName]
					return m, nil
				}
			}
		}

		// Handle escape key
		if msg.String() == "esc" {
			// When in directories view, go back to files
			if m.ShowDirs {
				m.ShowDirs = false
				return m, nil
			}

			// Remove focus from inputs, set focus to list
			m.PathInput.Blur()
			m.ExtInput.Blur()
			m.SizeInput.Blur()
			m.FocusedElement = "list"
			return m, nil
		}

		// Number keys for options
		if msg.String() == "1" || msg.String() == "2" {
			if !m.PathInput.Focused() && !m.ExtInput.Focused() && !m.SizeInput.Focused() {
				idx := int(msg.String()[0] - '1')
				if idx >= 0 && idx < len(m.Options) {
					optName := m.Options[idx]
					m.OptionState[optName] = !m.OptionState[optName]
					return m, m.LoadFiles()
				}
			}
		}
	}

	// Handle input updates
	switch m.FocusedElement {
	case "path":
		m.PathInput, cmd = m.PathInput.Update(msg)
		cmds = append(cmds, cmd)
	case "ext":
		if m.CurrentPath != "" {
			m.ExtInput, cmd = m.ExtInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	case "size":
		if m.CurrentPath != "" {
			m.SizeInput, cmd = m.SizeInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	case "exclude":
		if m.CurrentPath != "" {
			m.ExcludeInput, cmd = m.ExcludeInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	case "list":
		if m.CurrentPath != "" {
			if m.ShowDirs {
				m.DirList, cmd = m.DirList.Update(msg)
			} else {
				m.List, cmd = m.List.Update(msg)
			}
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *CleanFilesModel) LoadFiles() tea.Cmd {
	return func() tea.Msg {
		var items []list.Item
		var totalFilteredSize int64 = 0
		var filteredCount int = 0

		currentDir := m.CurrentPath

		// Get user-specified extensions
		extStr := m.ExtInput.Value()
		if extStr != "" {
			// Parse extensions from input
			m.Extensions = []string{}
			for _, ext := range strings.Split(extStr, ",") {
				ext = strings.TrimSpace(ext)
				if ext != "" {
					// Add dot prefix if needed
					if !strings.HasPrefix(ext, ".") {
						ext = "." + ext
					}
					m.Extensions = append(m.Extensions, strings.ToLower(ext))
				}
			}
		} else {
			// If no extensions specified, show all files
			m.Extensions = []string{}
		}

		excludeStr := m.ExcludeInput.Value()
		if excludeStr != "" {
			// Parse extensions from input
			m.Exclude = []string{}
			for _, exc := range strings.Split(excludeStr, ",") {
				exc = strings.TrimSpace(exc)
				if exc != "" {
					m.Exclude = append(m.Exclude, exc)
				}
			}
		} else {
			// If no extensions specified, show all files
			m.Exclude = []string{}
		}

		// Get user-specified min size
		sizeStr := m.SizeInput.Value()
		if sizeStr != "" {
			minSize, err := utils.ToBytes(sizeStr)
			if err == nil {
				m.MinSize = minSize
			} else {
				// If invalid size, reset to 0
				m.MinSize = 0
			}
		} else {
			// If no size specified, show all files regardless of size
			m.MinSize = 0
		}

		fileInfos, err := os.ReadDir(currentDir)
		if err != nil {
			return nil
		}

		// Add to parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir != currentDir {
			items = append(items, models.CleanItem{
				Path: parentDir,
				Size: -1, // Special value for parent directory
			})
		}

		// First collect directories
	dirLoop:
		for _, fileInfo := range fileInfos {
			if !fileInfo.IsDir() {
				continue
			}

			// Skip hidden directories unless enabled
			if !m.OptionState["Show hidden files"] && strings.HasPrefix(fileInfo.Name(), ".") {
				continue
			}

			path := filepath.Join(currentDir, fileInfo.Name())

			if len(m.Exclude) > 0 && fileInfo.IsDir() {
				for _, excludePattern := range m.Exclude {
					if strings.Contains(filepath.ToSlash(path+"/"), excludePattern+"/") {
						continue dirLoop
					}
				}
			}

			items = append(items, models.CleanItem{
				Path: path,
				Size: 0, // Directory
			})
		}

		// Then collect files
	fileLoop:
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				continue
			}

			// Skip hidden files unless enabled
			if !m.OptionState["Show hidden files"] && strings.HasPrefix(fileInfo.Name(), ".") {
				continue
			}

			path := filepath.Join(currentDir, fileInfo.Name())
			info, err := fileInfo.Info()
			if err != nil {
				continue
			}

			size := info.Size()

			// Apply extension filter if specified
			if len(m.Extensions) > 0 {
				ext := strings.ToLower(filepath.Ext(path))
				matched := false
				for _, allowedExt := range m.Extensions {
					if ext == allowedExt {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			// Apply size filter if specified
			if m.MinSize > 0 && size < m.MinSize {
				continue
			}

			if len(m.Exclude) > 0 && !fileInfo.IsDir() {
				for _, excludePattern := range m.Exclude {
					if strings.HasPrefix(fileInfo.Name(), excludePattern) {
						continue fileLoop
					}
				}
			}

			// Add to filtered size and count
			totalFilteredSize += size
			filteredCount++

			items = append(items, models.CleanItem{
				Path: path,
				Size: size,
			})
		}

		// Return both the items and the size info
		m.FilteredSize = totalFilteredSize
		m.FilteredCount = filteredCount
		return items
	}
}

func (m *CleanFilesModel) loadDirs() tea.Cmd {
	return func() tea.Msg {
		var items []list.Item

		// Add parent directory with special display
		parentDir := filepath.Dir(m.CurrentPath)
		if parentDir != m.CurrentPath {
			items = append(items, models.CleanItem{
				Path: parentDir,
				Size: -1, // Special value for parent directory
			})
		}

		// Read current directory
		entries, err := os.ReadDir(m.CurrentPath)
		if err != nil {
			return err
		}

		// Create a channel for results
		results := make(chan models.CleanItem, 100)
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
					if !m.OptionState["Show hidden files"] && strings.HasPrefix(entry.Name(), ".") {
						continue
					}
					results <- models.CleanItem{
						Path: filepath.Join(m.CurrentPath, entry.Name()),
						Size: 0,
					}
				}
			}
			close(results)
		}()

		// Wait for collection to complete
		<-done

		// Sort directories by name
		sort.Slice(items, func(i, j int) bool {
			return items[i].(models.CleanItem).Path < items[j].(models.CleanItem).Path
		})

		// Update path input with current path
		m.PathInput.SetValue(m.CurrentPath)

		return items
	}
}

// Asynchronous directory size calculation
func (m *CleanFilesModel) calculateDirSizeAsync() tea.Cmd {
	return func() tea.Msg {
		m.CalculatingSize = true
		size := calculateDirSize(m.CurrentPath)
		m.CalculatingSize = false
		return dirSizeMsg{size: size}
	}
}

// Function to calculate directory size recursively with option to cancel
func calculateDirSize(path string) int64 {
	// For very large directories, return a placeholder value immediately
	// to avoid blocking the UI
	_, err := os.Stat(path)
	if err != nil {
		return 0
	}

	// If it's a very large directory (like C: or Program Files)
	// just return 0 immediately to prevent lag
	if strings.HasSuffix(path, ":\\") || strings.Contains(path, "Program Files") {
		return 0
	}

	var totalSize int64 = 0

	// Use a channel to limit concurrency
	semaphore := make(chan struct{}, 10)
	var wg sync.WaitGroup

	// Create a function to process a directory
	var processDir func(string) int64
	processDir = func(dirPath string) int64 {
		var size int64 = 0
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return 0
		}

		for _, entry := range entries {
			// Skip hidden files and directories unless enabled
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(dirPath, entry.Name())
			if entry.IsDir() {
				// Process directories with concurrency limits
				wg.Add(1)
				go func(p string) {
					semaphore <- struct{}{}
					defer func() {
						<-semaphore
						wg.Done()
					}()
					dirSize := processDir(p)
					atomic.AddInt64(&totalSize, dirSize)
				}(fullPath)
			} else {
				// Process files directly
				info, err := entry.Info()
				if err == nil {
					fileSize := info.Size()
					atomic.AddInt64(&totalSize, fileSize)
					size += fileSize
				}
			}
		}
		return size
	}

	// Start processing
	processDir(path)

	wg.Wait()

	return totalSize
}

// Helper function to open directory in file explorer
func openFileExplorer(path string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd

		if runtime.GOOS == "windows" {
			cmd = exec.Command("explorer", path)
		} else if runtime.GOOS == "darwin" {
			cmd = exec.Command("open", path)
		} else {
			cmd = exec.Command("xdg-open", path)
		}

		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("could not open file explorer: %v", err)
		}

		return nil
	}
}

func (m *CleanFilesModel) OnDelete() (tea.Model, tea.Cmd) {
	if m.List.SelectedItem() != nil && !m.OptionState["Include subfolders"] {
		selectedItem := m.List.SelectedItem().(models.CleanItem)
		if selectedItem.Size > 0 { // Only delete files, not directories
			if !m.OptionState["Confirm deletion"] {
				// If confirm deletion is disabled, delete all files
				for _, listItem := range m.List.Items() {
					if fileItem, ok := listItem.(models.CleanItem); ok && fileItem.Size > 0 {
						err := os.Remove(fileItem.Path)
						if err != nil {
							m.Err = err
						}
					}
				}
			} else {
				// Delete just the selected file
				err := os.Remove(selectedItem.Path)
				if err != nil {
					m.Err = err
				}
			}
			return m, m.LoadFiles()
		}
	} else if m.OptionState["Include subfolders"] {
		// Delete all files in the current directory and all subfolders
		m.Filemanager.DeleteFiles(m.CurrentPath, m.Extensions, m.Exclude, utils.ToBytesOrDefault(m.SizeInput.Value()))

		if m.OptionState["Delete empty subfolders"] {
			m.Filemanager.DeleteEmptySubfolders(m.CurrentPath)
		}

		return m, m.LoadFiles()
	}
	return m, nil
}

func (m *CleanFilesModel) getLatestRules() (string, []string, int64, []string) {
	rules, err := m.Rules.GetRules()
	if err != nil {
		return "", []string{}, 0, []string{}
	}

	latestMinSize, _ := utils.ToBytes(rules.MinSize)

	return rules.Path, rules.Extensions, latestMinSize, rules.Exclude
}

func (m *CleanFilesModel) GetActiveTab() int {
	return m.ActiveTab
}

func (m *CleanFilesModel) GetCurrentPath() string {
	return m.CurrentPath
}

func (m *CleanFilesModel) GetExtensions() []string {
	return m.Extensions
}

func (m *CleanFilesModel) GetMinSize() int64 {
	return m.MinSize
}

func (m *CleanFilesModel) GetExclude() []string {
	return m.Exclude
}

func (m *CleanFilesModel) GetOptions() []string {
	return m.Options
}

func (m *CleanFilesModel) GetOptionState() map[string]bool {
	return m.OptionState
}

func (m *CleanFilesModel) GetFocusedElement() string {
	return m.FocusedElement
}

func (m *CleanFilesModel) GetShowDirs() bool {
	return m.ShowDirs
}

func (m *CleanFilesModel) GetDirSize() int64 {
	return m.DirSize
}

func (m *CleanFilesModel) GetCalculatingSize() bool {
	return m.CalculatingSize
}

func (m *CleanFilesModel) GetFilteredSize() int64 {
	return m.FilteredSize
}

func (m *CleanFilesModel) GetFilteredCount() int {
	return m.FilteredCount
}

func (m *CleanFilesModel) GetList() list.Model {
	return m.List
}

func (m *CleanFilesModel) GetDirList() list.Model {
	return m.DirList
}

func (m *CleanFilesModel) GetRules() rules.Rules {
	return m.Rules
}

func (m *CleanFilesModel) GetFilemanager() filemanager.FileManager {
	return m.Filemanager
}

func (m *CleanFilesModel) GetFileToDelete() *models.CleanItem {
	return m.FileToDelete
}

func (m *CleanFilesModel) GetPathInput() textinput.Model {
	return m.PathInput
}

func (m *CleanFilesModel) GetExtInput() textinput.Model {
	return m.ExtInput
}

func (m *CleanFilesModel) GetSizeInput() textinput.Model {
	return m.SizeInput
}

func (m *CleanFilesModel) GetExcludeInput() textinput.Model {
	return m.ExcludeInput
}

func (m *CleanFilesModel) SetFocusedElement(element string) {
	m.FocusedElement = element
}

func (m *CleanFilesModel) SetShowDirs(show bool) {
	m.ShowDirs = show
}

func (m *CleanFilesModel) SetOptionState(option string, state bool) {
	if m.OptionState == nil {
		m.OptionState = make(map[string]bool)
	}
	m.OptionState[option] = state
}

func (m *CleanFilesModel) SetMinSize(size int64) {
	m.MinSize = size
}

func (m *CleanFilesModel) SetExclude(exclude []string) {
	m.Exclude = exclude
}

func (m *CleanFilesModel) SetExtensions(extensions []string) {
	m.Extensions = extensions
}
