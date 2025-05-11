package tui

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
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/utils"
)

type cleanItem struct {
	path string
	size int64
}

func (i cleanItem) Title() string {
	if i.size == -1 {
		return "📂 .." // Parent directory
	}

	if i.size == 0 {
		return "📁 " + filepath.Base(i.path) // Directory
	}

	// Regular file
	filename := filepath.Base(i.path)
	ext := filepath.Ext(filename)

	// Choose icon based on file extension
	icon := "📄 " // Default file icon
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		icon = "🖼️ " // Image
	case ".mp3", ".wav", ".flac", ".ogg":
		icon = "🎵 " // Audio
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv":
		icon = "🎬 " // Video
	case ".pdf":
		icon = "📕 " // PDF
	case ".doc", ".docx", ".txt", ".rtf":
		icon = "📝 " // Document
	case ".zip", ".rar", ".tar", ".gz", ".7z":
		icon = "🗜️ " // Archive
	case ".exe", ".msi", ".bat":
		icon = "⚙️ " // Executable
	}

	// Format the size with unit
	sizeStr := utils.FormatSize(i.size)

	// Calculate padding for alignment
	padding := 50 - len(filename)
	if padding < 0 {
		padding = 0
	}

	return fmt.Sprintf("%s%s%s%s", icon, filename, strings.Repeat(" ", padding), sizeStr)
}

func (i cleanItem) Description() string { return i.path }
func (i cleanItem) FilterValue() string { return i.path }

// Message for directory size updates
type dirSizeMsg struct {
	size int64
}

// Message for filtered files size updates
type filteredSizeMsg struct {
	size  int64
	count int
}

type model struct {
	list            list.Model
	extInput        textinput.Model
	sizeInput       textinput.Model
	pathInput       textinput.Model
	excludeInput    textinput.Model
	currentPath     string
	extensions      []string
	minSize         int64
	exclude         []string
	options         []string
	optionState     map[string]bool
	err             error
	focusedElement  string // "path", "ext", "size", "button", "option1", "option2", "option3"
	fileToDelete    *cleanItem
	showDirs        bool
	dirList         list.Model
	dirSize         int64 // Cached directory size
	calculatingSize bool  // Flag to indicate size calculation in progress
	filteredSize    int64 // Total size of filtered files
	filteredCount   int   // Count of filtered files
	activeTab       int   // 0 for files, 1 for exclude, 2 for options, 3 for hot keys
	rules           rules.Rules
	filemanager     filemanager.FileManager
}

func initialModel(rules rules.Rules) *model {
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
	extInput.PromptStyle = TextInputPromptStyle
	extInput.TextStyle = TextInputTextStyle
	extInput.Cursor.Style = TextInputCursorStyle

	sizeInput := textinput.New()
	sizeInput.Placeholder = "(e.g. 10kb,10mb,10b)..."
	sizeInput.SetValue(latestMinSize)
	minSize, _ := utils.ToBytes(latestMinSize)
	sizeInput.PromptStyle = TextInputPromptStyle
	sizeInput.TextStyle = TextInputTextStyle
	sizeInput.Cursor.Style = TextInputCursorStyle

	pathInput := textinput.New()
	pathInput.SetValue(latestDir)
	pathInput.PromptStyle = TextInputPromptStyle
	pathInput.TextStyle = TextInputTextStyle
	pathInput.Cursor.Style = TextInputCursorStyle

	excludeInput := textinput.New()
	excludeInput.Placeholder = "Exclude specific files/paths (e.g. data,backup)"
	excludeInput.SetValue(strings.Join(latestExclude, ","))
	excludeInput.PromptStyle = TextInputPromptStyle
	excludeInput.TextStyle = TextInputTextStyle
	excludeInput.Cursor.Style = TextInputCursorStyle

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
	l.Styles.Title = TitleStyle

	// Create directory list with same delegate
	dirList := list.New([]list.Item{}, delegate, 30, 10)
	dirList.SetShowTitle(true)
	dirList.Title = "Directories"
	dirList.SetShowStatusBar(true)
	dirList.SetFilteringEnabled(false)
	dirList.SetShowHelp(false)
	dirList.Styles.Title = TitleStyle

	// Define options in fixed order
	options := []string{
		"Show hidden files",
		"Confirm deletion",
		"Include subfolders",
		"Delete empty subfolders",
	}

	optionState := map[string]bool{
		"Show hidden files":       false,
		"Confirm deletion":        false,
		"Include subfolders":      false,
		"Delete empty subfolders": false,
	}

	return &model{
		list:            l,
		extInput:        extInput,
		sizeInput:       sizeInput,
		pathInput:       pathInput,
		excludeInput:    excludeInput,
		currentPath:     latestDir,
		extensions:      latestExtensions,
		minSize:         minSize,
		exclude:         latestExclude,
		options:         options,
		optionState:     optionState,
		focusedElement:  "list",
		showDirs:        false,
		dirList:         dirList,
		dirSize:         0,
		calculatingSize: false,
		filteredSize:    0,
		filteredCount:   0,
		activeTab:       0,
		rules:           rules,
	}
}

func (m *model) Init() tea.Cmd {
	// Set initial focus to path input
	m.focusedElement = "path"
	m.pathInput.Focus()

	// If we have a path, load files and calculate size
	if m.currentPath != "" {
		return tea.Batch(m.loadFiles(), m.calculateDirSizeAsync())
	}

	// Otherwise just return the blink command for the path input
	return textinput.Blink
}

func Run(filemanager filemanager.FileManager, rules rules.Rules) error {
	p := tea.NewProgram(initialModel(rules),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithFPS(30),
		tea.WithInputTTY(),
		tea.WithOutput(os.Stderr),
	)
	_, err := p.Run()
	return err
}

func (m *model) View() string {
	// --- Tabs rendering ---
	tabNames := []string{"🗂️ [F1] Main", "🧹 [F2] Filters", "⚙️ [F3] Options", "❔ [F4] Help"}
	tabs := make([]string, 4)
	for i, name := range tabNames {
		style := TabStyle
		if m.activeTab == i {
			style = ActiveTabStyle
		}
		tabs[i] = style.Render(name)
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

	// --- Content rendering ---
	var content strings.Builder
	content.WriteString(tabsRow)
	content.WriteString("\n")

	if m.activeTab == 3 {
		// Help tab content

		// Navigation
		content.WriteString(OptionStyle.Render("Navigation:"))
		content.WriteString("\n")
		content.WriteString("  F1-F4    - Switch between tabs\n")
		content.WriteString("  Esc      - Return to main menu\n")
		content.WriteString("  Tab      - Next field\n")
		content.WriteString("  Shift+Tab - Previous field\n")
		content.WriteString("  Ctrl+C   - Exit application\n\n")

		// File Operations
		content.WriteString(OptionStyle.Render("File Operations:"))
		content.WriteString("\n")
		content.WriteString("  Ctrl+R   - Refresh file list\n")
		content.WriteString("  Crtl+O   - Open in explorer\n")
		content.WriteString("  Ctrl+D   - Delete files\n\n")

		// Filter Operations
		content.WriteString(OptionStyle.Render("Filter Operations:"))
		content.WriteString("\n")
		content.WriteString("  Alt+C    - Clear filters\n\n")

		// Options
		content.WriteString(OptionStyle.Render("Options:"))
		content.WriteString("\n")
		content.WriteString("  Alt+1    - Toggle hidden files\n")
		content.WriteString("  Alt+2    - Toggle confirm deletion\n")
		content.WriteString("  Alt+3    - Toggle include subfolders\n")
		content.WriteString("  Alt+4    - Toggle delete empty subfolders\n")
	} else if m.activeTab == 0 {
		pathStyle := StandardInputStyle
		if m.focusedElement == "path" {
			pathStyle = StandardInputFocusedStyle
		}
		content.WriteString(pathStyle.Render("Current Path: " + m.pathInput.View()))

		// If no path is set, show only the start button
		if m.currentPath == "" {
			startButtonStyle := LaunchButtonStyle
			if m.focusedElement == "startButton" {
				startButtonStyle = LaunchButtonFocusedStyle
			}
			content.WriteString("\n")
			content.WriteString(startButtonStyle.Render("📂 Launch"))
		} else {
			// Show full interface when path is set
			extStyle := StandardInputStyle
			if m.focusedElement == "ext" {
				extStyle = StandardInputFocusedStyle
			}
			content.WriteString("\n")
			content.WriteString(extStyle.Render("Extensions: " + m.extInput.View()))
			content.WriteString("\n")
			var activeList list.Model
			if m.showDirs {
				activeList = m.dirList
			} else {
				activeList = m.list
			}
			fileCount := len(activeList.Items())
			filteredSizeText := utils.FormatSize(m.filteredSize)
			content.WriteString("\n")
			if !m.showDirs {
				content.WriteString(TitleStyle.Render(fmt.Sprintf("Selected files (%d) • Size of selected files: %s",
					m.filteredCount, filteredSizeText)))
			} else {
				content.WriteString(TitleStyle.Render(fmt.Sprintf("Directories in %s (%d)",
					filepath.Base(m.currentPath), fileCount)))
			}
			content.WriteString("\n")
			listStyle := ListStyle
			if m.focusedElement == "list" {
				listStyle = ListFocusedStyle
			}

			var listContent strings.Builder
			if len(activeList.Items()) == 0 {
				if !m.showDirs {
					listContent.WriteString("No files match your filters. Try changing extensions or size filters.")
				} else {
					listContent.WriteString("No directories found in this location.")
				}
			} else {
				items := activeList.Items()
				selectedIndex := activeList.Index()
				totalItems := len(items)

				visibleItems := 10
				if visibleItems > totalItems {
					visibleItems = totalItems
				}

				startIdx := 0
				if selectedIndex > visibleItems-3 && totalItems > visibleItems {
					startIdx = selectedIndex - (visibleItems / 2)
					if startIdx+visibleItems > totalItems {
						startIdx = totalItems - visibleItems
					}
				}
				if startIdx < 0 {
					startIdx = 0
				}

				endIdx := startIdx + visibleItems
				if endIdx > totalItems {
					endIdx = totalItems
				}

				for i := startIdx; i < endIdx; i++ {
					item := items[i].(cleanItem)

					icon := "📄 "
					if item.size == -1 {
						icon = "⬆️ "
					} else if item.size == 0 {
						icon = "📁 "
					} else {
						ext := strings.ToLower(filepath.Ext(item.path))
						switch ext {
						case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".apng":
							icon = "🖼️ "
						case ".mp3", ".wav", ".flac", ".ogg":
							icon = "🎵 "
						case ".mp4", ".avi", ".mkv", ".mov":
							icon = "🎬 "
						case ".zip", ".rar", ".7z", ".tar", ".gz":
							icon = "🗜️ "
						case ".exe", ".msi":
							icon = "⚙️ "
						case ".pdf":
							icon = "📕 "
						case ".doc", ".docx", ".txt":
							icon = "📝 "
						}
					}

					filename := filepath.Base(item.path)
					sizeStr := ""
					if item.size > 0 {
						sizeStr = utils.FormatSize(item.size)
					} else if item.size == 0 {
						sizeStr = "DIR"
					} else {
						sizeStr = "UP DIR"
					}

					prefix := "  "
					style := lipgloss.NewStyle()

					if i == selectedIndex {
						prefix = "> "
						style = style.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#0066FF")).Bold(true)
					} else if item.size == -1 || item.size == 0 {
						style = style.Foreground(lipgloss.Color("#4DC4FF"))
					}

					displayName := filename
					if len(displayName) > 40 {
						displayName = displayName[:37] + "..."
					}

					padding := 44 - len(displayName)
					if padding < 1 {
						padding = 1
					}

					fileLine := fmt.Sprintf("%s%s%s%s%s",
						prefix,
						icon,
						displayName,
						strings.Repeat(" ", padding),
						sizeStr)

					listContent.WriteString(style.Render(fileLine))
					listContent.WriteString("\n")
				}

				if totalItems > visibleItems {
					scrollInfo := fmt.Sprintf("\nShowing %d-%d of %d items (%.0f%%)",
						startIdx+1, endIdx, totalItems,
						float64(selectedIndex+1)/float64(totalItems)*100)
					listContent.WriteString(lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#999999")).Render(scrollInfo))
				}
			}
			content.WriteString(listStyle.Render(listContent.String()))

			// Buttons section
			content.WriteString("\n\n")
			if m.focusedElement == "dirButton" {
				content.WriteString(StandardButtonFocusedStyle.Render("➡️ Show directories"))
			} else {
				content.WriteString(StandardButtonStyle.Render("➡️ Show directories"))
			}
			content.WriteString("\n\n")

			if m.focusedElement == "button" {
				content.WriteString(DeleteButtonFocusedStyle.Render("🗑️ Start cleaning"))
			} else {
				content.WriteString(DeleteButtonStyle.Render("🗑️ Start cleaning"))
			}
			content.WriteString("\n")
		}
	} else if m.activeTab == 1 {
		// Filters tab
		excludeStyle := StandardInputStyle
		if m.focusedElement == "exclude" {
			excludeStyle = StandardInputFocusedStyle
		}
		m.excludeInput.Placeholder = "specific files/paths (e.g. data,backup)"
		content.WriteString(excludeStyle.Render("Exclude: " + m.excludeInput.View()))
		content.WriteString("\n")
		sizeStyle := StandardInputStyle
		if m.focusedElement == "size" {
			sizeStyle = StandardInputFocusedStyle
		}
		content.WriteString(sizeStyle.Render("Min size: " + m.sizeInput.View()))
	} else if m.activeTab == 2 {
		// Options tab
		for i, name := range m.options {
			style := OptionStyle
			if m.optionState[name] {
				style = SelectedOptionStyle
			}
			if m.focusedElement == fmt.Sprintf("option%d", i+1) {
				style = OptionFocusedStyle
			}
			content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", i+1)))
			content.WriteString(style.Render(fmt.Sprintf("[%s] %-20s", map[bool]string{true: "✓", false: "○"}[m.optionState[name]], name)))
			content.WriteString("\n")
		}
	}

	// Combine everything
	var ui string
	if m.activeTab != 0 {
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

	if m.err != nil {
		ui += "\n" + ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	return AppStyle.Render(ui)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case dirSizeMsg:
		// Update the directory size
		m.dirSize = msg.size
		return m, nil

	case tea.WindowSizeMsg:
		// Properly set both width and height
		h, v := AppStyle.GetFrameSize()
		// Further reduce listHeight by another 10% (now at 65% of original)
		listHeight := (msg.Height - v - 15) * 65 / 100 // Reserve space for other UI elements and reduce by 35%
		if listHeight < 5 {
			listHeight = 5 // Minimum height to show something
		}
		m.list.SetSize(msg.Width-h, listHeight)
		m.dirList.SetSize(msg.Width-h, listHeight)

		cmds = append(cmds, m.loadFiles())
		// Trigger directory size calculation when changing directory
		cmds = append(cmds, m.calculateDirSizeAsync())
		return m, tea.Batch(cmds...)

	// Handle message for setting items in the list
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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+d":
			return m.OnDelete()
		case "tab", "shift+tab":
			// Handle tab navigation based on active tab
			if m.activeTab == 0 {
				if m.currentPath == "" {
					// When no path is set, only allow focusing between path input and start button
					if msg.String() == "tab" {
						if m.focusedElement == "path" {
							m.focusedElement = "startButton"
							m.pathInput.Blur()
						} else {
							m.focusedElement = "path"
							m.pathInput.Focus()
						}
					} else {
						if m.focusedElement == "path" {
							m.focusedElement = "startButton"
							m.pathInput.Blur()
						} else {
							m.focusedElement = "path"
							m.pathInput.Focus()
						}
					}
					return m, nil
				}
				// Normal tab behavior when path is set
				if msg.String() == "tab" {
					switch m.focusedElement {
					case "path":
						m.focusedElement = "ext"
						m.pathInput.Blur()
						m.extInput.Focus()
					case "ext":
						m.extInput.Blur()
						m.focusedElement = "size"
						m.sizeInput.Focus()
					case "size":
						m.sizeInput.Blur()
						m.focusedElement = "list"
					case "list":
						m.focusedElement = "dirButton"
					case "dirButton":
						m.focusedElement = "button"
					case "button":
						m.focusedElement = "path"
						m.pathInput.Focus()
					}
				} else {
					switch m.focusedElement {
					case "path":
						m.pathInput.Blur()
						m.focusedElement = "button"
					case "button":
						m.focusedElement = "dirButton"
					case "dirButton":
						m.focusedElement = "list"
					case "list":
						m.focusedElement = "size"
						m.sizeInput.Focus()
					case "size":
						m.sizeInput.Blur()
						m.focusedElement = "ext"
						m.extInput.Focus()
					case "ext":
						m.extInput.Blur()
						m.focusedElement = "path"
						m.pathInput.Focus()
					}
				}
			} else if m.activeTab == 1 {
				// Tab navigation for Filters tab
				if msg.String() == "tab" {
					switch m.focusedElement {
					case "exclude":
						m.excludeInput.Blur()
						m.focusedElement = "size"
						m.sizeInput.Focus()
					case "size":
						m.sizeInput.Blur()
						m.focusedElement = "exclude"
						m.excludeInput.Focus()
					}
				} else {
					switch m.focusedElement {
					case "exclude":
						m.excludeInput.Blur()
						m.focusedElement = "size"
						m.sizeInput.Focus()
					case "size":
						m.sizeInput.Blur()
						m.focusedElement = "exclude"
						m.excludeInput.Focus()
					}
				}
			} else if m.activeTab == 2 {
				// Tab navigation for Options tab
				if msg.String() == "tab" {
					switch m.focusedElement {
					case "option1":
						m.focusedElement = "option2"
					case "option2":
						m.focusedElement = "option3"
					case "option3":
						m.focusedElement = "option4"
					case "option4":
						m.focusedElement = "option1"
					}
				} else {
					switch m.focusedElement {
					case "option1":
						m.focusedElement = "option4"
					case "option2":
						m.focusedElement = "option1"
					case "option3":
						m.focusedElement = "option2"
					case "option4":
						m.focusedElement = "option3"
					}
				}
			} else if m.activeTab == 3 {
				// Hot Keys tab navigation
				if msg.String() == "tab" {
					switch m.focusedElement {
					case "hotKeys":
						m.focusedElement = "navigation"
					case "navigation":
						m.focusedElement = "fileOperations"
					case "fileOperations":
						m.focusedElement = "filterOperations"
					case "filterOperations":
						m.focusedElement = "directoryNavigation"
					case "directoryNavigation":
						m.focusedElement = "options"
					case "options":
						m.focusedElement = "hotKeys"
					}
				} else {
					switch m.focusedElement {
					case "hotKeys":
						m.focusedElement = "options"
					case "options":
						m.focusedElement = "directoryNavigation"
					case "directoryNavigation":
						m.focusedElement = "fileOperations"
					case "fileOperations":
						m.focusedElement = "navigation"
					case "navigation":
						m.focusedElement = "hotKeys"
					}
				}
			}
			return m, nil
		case "enter":
			if m.currentPath == "" {
				if m.focusedElement == "startButton" {
					// Validate and set the path
					path := m.pathInput.Value()
					if path != "" {
						// Expand tilde in path
						expandedPath := utils.ExpandTilde(path)
						if _, err := os.Stat(expandedPath); err == nil {
							m.currentPath = expandedPath

							// Load files for the new path
							cmds = append(cmds, m.loadFiles(), m.calculateDirSizeAsync())

							// Set focus to path input
							m.focusedElement = "path"
							m.pathInput.Focus()
						} else {
							m.err = fmt.Errorf("invalid path: %s", path)
						}
					}
				}
			} else {
				switch m.focusedElement {
				case "ext", "size", "exclude":
					// Обновляем список файлов при нажатии Enter на полях ввода
					return m, m.loadFiles()
				case "list":
					if !m.showDirs && m.list.SelectedItem() != nil {
						selectedItem := m.list.SelectedItem().(cleanItem)
						if selectedItem.size == -1 {
							// Handle parent directory selection
							m.currentPath = selectedItem.path
							m.pathInput.SetValue(selectedItem.path)
							// Recalculate directory size when changing directory
							cmds = append(cmds, m.loadFiles(), m.calculateDirSizeAsync())
							return m, tea.Batch(cmds...)
						}
						// If it's a directory, navigate into it
						info, err := os.Stat(selectedItem.path)
						if err == nil && info.IsDir() {
							m.currentPath = selectedItem.path
							m.pathInput.SetValue(selectedItem.path)
							// Recalculate directory size when changing directory
							cmds = append(cmds, m.loadFiles(), m.calculateDirSizeAsync())
							return m, tea.Batch(cmds...)
						}
					} else if m.showDirs && m.dirList.SelectedItem() != nil {
						selectedDir := m.dirList.SelectedItem().(cleanItem)
						m.currentPath = selectedDir.path
						m.pathInput.SetValue(selectedDir.path)
						m.showDirs = false
						// Recalculate directory size when changing directory
						cmds = append(cmds, m.loadFiles(), m.calculateDirSizeAsync())
						return m, tea.Batch(cmds...)
					}
				case "dirButton":
					m.showDirs = true
					return m, m.loadDirs()
				case "button":
					if m.activeTab == 0 {
						return m.OnDelete()
					}
					return m, nil
				case "option1", "option2", "option3", "option4":
					idx := 0
					if m.focusedElement == "option2" {
						idx = 1
					}
					if m.focusedElement == "option3" {
						idx = 2
					}
					if m.focusedElement == "option4" {
						idx = 3
					}
					if idx < len(m.options) {
						optName := m.options[idx]
						m.optionState[optName] = !m.optionState[optName]
						m.focusedElement = "option" + fmt.Sprintf("%d", idx+1)

						return m, nil
					}
				}
			}
		}

		// Global hotkeys that work regardless of focus
		switch msg.String() {
		case "ctrl+r": // Refresh files
			return m, m.loadFiles()
		case "ctrl+d": // Delete files
			return m.OnDelete()
		case "ctrl+o": // Open current directory in file explorer
			cmd := openFileExplorer(m.currentPath)
			return m, cmd
		case "alt+c": // Clear filters
			m.sizeInput.SetValue("")
			m.excludeInput.SetValue("")
			return m, m.loadFiles()
		case "alt+1": // Toggle hidden files
			m.optionState["Show hidden files"] = !m.optionState["Show hidden files"]
			return m, m.loadFiles()
		case "alt+2": // Toggle confirm deletion
			m.optionState["Confirm deletion"] = !m.optionState["Confirm deletion"]
			return m, nil
		case "alt+3": // Toggle include subfolders
			m.optionState["Include subfolders"] = !m.optionState["Include subfolders"]
			return m, nil
		case "alt+4": // Toggle delete empty subfolders
			m.optionState["Delete empty subfolders"] = !m.optionState["Delete empty subfolders"]
			return m, nil
		case "left", "right": // Tab switching
			if msg.String() == "left" && m.activeTab > 0 {
				m.activeTab--
				if m.activeTab == 1 {
					m.excludeInput.Focus()
					m.focusedElement = "exclude"
				}
			}
			if msg.String() == "right" && m.activeTab < 3 {
				m.activeTab++
				if m.activeTab == 1 {
					m.excludeInput.Focus()
					m.focusedElement = "exclude"
				}
			}
			return m, nil
		case "f1":
			m.activeTab = 0
			m.focusedElement = "path"
			m.pathInput.Focus()
			return m, nil
		case "f2":
			m.activeTab = 1
			m.focusedElement = "exclude"
			m.excludeInput.Focus()
			m.pathInput.Blur()
			m.extInput.Blur()
			m.sizeInput.Blur()
			return m, nil
		case "f3":
			m.activeTab = 2
			m.focusedElement = "option1"
			return m, nil
		case "f4":
			m.activeTab = 3
			return m, nil
		case "up", "down": // Always handle arrow keys for list navigation regardless of focus
			// Make list navigation global - arrow keys always navigate the list
			if !m.showDirs {
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.dirList, cmd = m.dirList.Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Handle space key for options
		if (msg.String() == " " || msg.String() == "enter") && m.activeTab == 2 {
			if m.focusedElement == "option1" || m.focusedElement == "option2" || m.focusedElement == "option3" || m.focusedElement == "option4" {
				idx := int(m.focusedElement[len(m.focusedElement)-1] - '1')
				if idx >= 0 && idx < len(m.options) {
					optName := m.options[idx]
					m.optionState[optName] = !m.optionState[optName]
					return m, nil
				}
			}
		}

		// Handle escape key
		if msg.String() == "esc" {
			// When in directories view, go back to files
			if m.showDirs {
				m.showDirs = false
				return m, nil
			}

			// Remove focus from inputs, set focus to list
			m.pathInput.Blur()
			m.extInput.Blur()
			m.sizeInput.Blur()
			m.focusedElement = "list"
			return m, nil
		}

		// Number keys for options
		if msg.String() == "1" || msg.String() == "2" {
			if !m.pathInput.Focused() && !m.extInput.Focused() && !m.sizeInput.Focused() {
				idx := int(msg.String()[0] - '1')
				if idx >= 0 && idx < len(m.options) {
					optName := m.options[idx]
					m.optionState[optName] = !m.optionState[optName]
					return m, m.loadFiles()
				}
			}
		}
	}

	// Handle input updates
	switch m.focusedElement {
	case "path":
		m.pathInput, cmd = m.pathInput.Update(msg)
		cmds = append(cmds, cmd)
	case "ext":
		if m.currentPath != "" {
			m.extInput, cmd = m.extInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	case "size":
		if m.currentPath != "" {
			m.sizeInput, cmd = m.sizeInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	case "exclude":
		if m.currentPath != "" {
			m.excludeInput, cmd = m.excludeInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	case "list":
		if m.currentPath != "" {
			if m.showDirs {
				m.dirList, cmd = m.dirList.Update(msg)
			} else {
				m.list, cmd = m.list.Update(msg)
			}
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) loadFiles() tea.Cmd {
	return func() tea.Msg {
		var items []list.Item
		var totalFilteredSize int64 = 0
		var filteredCount int = 0

		currentDir := m.currentPath

		// Get user-specified extensions
		extStr := m.extInput.Value()
		if extStr != "" {
			// Parse extensions from input
			m.extensions = []string{}
			for _, ext := range strings.Split(extStr, ",") {
				ext = strings.TrimSpace(ext)
				if ext != "" {
					// Add dot prefix if needed
					if !strings.HasPrefix(ext, ".") {
						ext = "." + ext
					}
					m.extensions = append(m.extensions, strings.ToLower(ext))
				}
			}
		} else {
			// If no extensions specified, show all files
			m.extensions = []string{}
		}

		excludeStr := m.excludeInput.Value()
		if excludeStr != "" {
			// Parse extensions from input
			m.exclude = []string{}
			for _, exc := range strings.Split(excludeStr, ",") {
				exc = strings.TrimSpace(exc)
				if exc != "" {
					m.exclude = append(m.exclude, exc)
				}
			}
		} else {
			// If no extensions specified, show all files
			m.exclude = []string{}
		}

		// Get user-specified min size
		sizeStr := m.sizeInput.Value()
		if sizeStr != "" {
			minSize, err := utils.ToBytes(sizeStr)
			if err == nil {
				m.minSize = minSize
			} else {
				// If invalid size, reset to 0
				m.minSize = 0
			}
		} else {
			// If no size specified, show all files regardless of size
			m.minSize = 0
		}

		fileInfos, err := os.ReadDir(currentDir)
		if err != nil {
			return nil
		}

		// Add to parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir != currentDir {
			items = append(items, cleanItem{
				path: parentDir,
				size: -1, // Special value for parent directory
			})
		}

		// First collect directories
	dirLoop:
		for _, fileInfo := range fileInfos {
			if !fileInfo.IsDir() {
				continue
			}

			// Skip hidden directories unless enabled
			if !m.optionState["Show hidden files"] && strings.HasPrefix(fileInfo.Name(), ".") {
				continue
			}

			path := filepath.Join(currentDir, fileInfo.Name())

			if len(m.exclude) > 0 && fileInfo.IsDir() {
				for _, excludePattern := range m.exclude {
					if strings.Contains(filepath.ToSlash(path+"/"), excludePattern+"/") {
						continue dirLoop
					}
				}
			}

			items = append(items, cleanItem{
				path: path,
				size: 0, // Directory
			})
		}

		// Then collect files
	fileLoop:
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				continue
			}

			// Skip hidden files unless enabled
			if !m.optionState["Show hidden files"] && strings.HasPrefix(fileInfo.Name(), ".") {
				continue
			}

			path := filepath.Join(currentDir, fileInfo.Name())
			info, err := fileInfo.Info()
			if err != nil {
				continue
			}

			size := info.Size()

			// Apply extension filter if specified
			if len(m.extensions) > 0 {
				ext := strings.ToLower(filepath.Ext(path))
				matched := false
				for _, allowedExt := range m.extensions {
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
			if m.minSize > 0 && size < m.minSize {
				continue
			}

			if len(m.exclude) > 0 && !fileInfo.IsDir() {
				for _, excludePattern := range m.exclude {
					if strings.HasPrefix(fileInfo.Name(), excludePattern) {
						continue fileLoop
					}
				}
			}

			// Add to filtered size and count
			totalFilteredSize += size
			filteredCount++

			items = append(items, cleanItem{
				path: path,
				size: size,
			})
		}

		// Return both the items and the size info
		m.filteredSize = totalFilteredSize
		m.filteredCount = filteredCount
		return items
	}
}

func (m *model) loadDirs() tea.Cmd {
	return func() tea.Msg {
		var items []list.Item

		// Add parent directory with special display
		parentDir := filepath.Dir(m.currentPath)
		if parentDir != m.currentPath {
			items = append(items, cleanItem{
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
		results := make(chan cleanItem, 100)
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
					results <- cleanItem{
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
			return items[i].(cleanItem).path < items[j].(cleanItem).path
		})

		// Update path input with current path
		m.pathInput.SetValue(m.currentPath)

		return items
	}
}

// Asynchronous directory size calculation
func (m *model) calculateDirSizeAsync() tea.Cmd {
	return func() tea.Msg {
		m.calculatingSize = true
		size := calculateDirSize(m.currentPath)
		m.calculatingSize = false
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

func (m *model) OnDelete() (tea.Model, tea.Cmd) {
	if m.list.SelectedItem() != nil && !m.optionState["Include subfolders"] {
		selectedItem := m.list.SelectedItem().(cleanItem)
		if selectedItem.size > 0 { // Only delete files, not directories
			if !m.optionState["Confirm deletion"] {
				// If confirm deletion is disabled, delete all files
				for _, listItem := range m.list.Items() {
					if fileItem, ok := listItem.(cleanItem); ok && fileItem.size > 0 {
						err := os.Remove(fileItem.path)
						if err != nil {
							m.err = err
						}
					}
				}
			} else {
				// Delete just the selected file
				err := os.Remove(selectedItem.path)
				if err != nil {
					m.err = err
				}
			}
			return m, m.loadFiles()
		}
	} else if m.optionState["Include subfolders"] {
		// Delete all files in the current directory and all subfolders
		m.filemanager.DeleteFiles(m.currentPath, m.extensions, m.exclude, utils.ToBytesOrDefault(m.sizeInput.Value()))

		if m.optionState["Delete empty subfolders"] {
			m.filemanager.DeleteEmptySubfolders(m.currentPath)
		}

		return m, m.loadFiles()
	}
	return m, nil
}

func (m *model) getLatestRules() (string, []string, int64, []string) {
	rules, err := m.rules.GetRules()
	if err != nil {
		return "", []string{}, 0, []string{}
	}

	latestMinSize, _ := utils.ToBytes(rules.MinSize)

	return rules.Path, rules.Extensions, latestMinSize, rules.Exclude
}
