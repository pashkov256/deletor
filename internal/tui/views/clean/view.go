package clean

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
	"github.com/pashkov256/deletor/internal/tui/models"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/tui/views/tabs"
	"github.com/pashkov256/deletor/internal/utils"
)

func (i models.CleanItem) Title() string {
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

func (i CleanItem) Description() string { return i.path }
func (i CleanItem) FilterValue() string { return i.path }

func InitialModel(rules rules.Rules) *CleanFilesModel {
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
		Options:         options,
		OptionState:     optionState,
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
	m.TabManager = tabs.NewCleanTabManager(m)
	// If we have a path, load files and calculate size
	if m.CurrentPath != "" {
		return tea.Batch(m.LoadFiles(), m.calculateDirSizeAsync())
	}

	// Otherwise just return the blink command for the path input
	return textinput.Blink
}

func Run(filemanager filemanager.FileManager, rules rules.Rules) error {
	p := tea.NewProgram(InitialModel(rules),
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

	if activeTab == 3 {
		content.WriteString(m.TabManager.GetActiveTab().View())
	} else if activeTab == 0 {
		pathStyle := styles.StandardInputStyle
		if m.FocusedElement == "path" {
			pathStyle = styles.StandardInputFocusedStyle
		}
		content.WriteString(pathStyle.Render("Current Path: " + m.PathInput.View()))

		// If no path is set, show only the start button
		if m.CurrentPath == "" {
			startButtonStyle := styles.LaunchButtonStyle
			if m.FocusedElement == "startButton" {
				startButtonStyle = styles.LaunchButtonFocusedStyle
			}
			content.WriteString("\n")
			content.WriteString(startButtonStyle.Render("📂 Launch"))
		} else {
			// Show full interface when path is set
			extStyle := styles.StandardInputStyle
			if m.FocusedElement == "ext" {
				extStyle = styles.StandardInputFocusedStyle
			}
			content.WriteString("\n")
			content.WriteString(extStyle.Render("Extensions: " + m.ExtInput.View()))
			content.WriteString("\n")
			var activeList list.Model
			if m.ShowDirs {
				activeList = m.DirList
			} else {
				activeList = m.List
			}
			fileCount := len(activeList.Items())
			filteredSizeText := utils.FormatSize(m.FilteredSize)
			content.WriteString("\n")
			if !m.ShowDirs {
				content.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Selected files (%d) • Size of selected files: %s",
					m.FilteredCount, filteredSizeText)))
			} else {
				content.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Directories in %s (%d)",
					filepath.Base(m.CurrentPath), fileCount)))
			}
			content.WriteString("\n")
			listStyle := styles.ListStyle
			if m.FocusedElement == "list" {
				listStyle = styles.ListFocusedStyle
			}

			var listContent strings.Builder
			if len(activeList.Items()) == 0 {
				if !m.ShowDirs {
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
					item := items[i].(CleanItem)

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
			if m.FocusedElement == "dirButton" {
				content.WriteString(styles.StandardButtonFocusedStyle.Render("➡️ Show directories"))
			} else {
				content.WriteString(styles.StandardButtonStyle.Render("➡️ Show directories"))
			}
			content.WriteString("\n\n")

			if m.FocusedElement == "button" {
				content.WriteString(styles.DeleteButtonFocusedStyle.Render("🗑️ Start cleaning"))
			} else {
				content.WriteString(styles.DeleteButtonStyle.Render("🗑️ Start cleaning"))
			}
			content.WriteString("\n")
		}
	} else if activeTab == 1 {
		// // Filters tab
		// excludeStyle := styles.StandardInputStyle
		// if m.FocusedElement == "exclude" {
		// 	excludeStyle = styles.StandardInputFocusedStyle
		// }
		// m.ExcludeInput.Placeholder = "specific files/paths (e.g. data,backup)"
		// content.WriteString(excludeStyle.Render("Exclude: " + m.ExcludeInput.View()))
		// content.WriteString("\n")
		// sizeStyle := styles.StandardInputStyle
		// if m.FocusedElement == "size" {
		// 	sizeStyle = styles.StandardInputFocusedStyle
		// }
		// content.WriteString(sizeStyle.Render("Min size: " + m.SizeInput.View()))
		content.WriteString(m.TabManager.GetActiveTab().View())
	} else if activeTab == 2 {
		// Options tab
		for i, name := range m.Options {
			style := styles.OptionStyle
			if m.OptionState[name] {
				style = styles.SelectedOptionStyle
			}
			if m.FocusedElement == fmt.Sprintf("option%d", i+1) {
				style = styles.OptionFocusedStyle
			}
			content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", i+1)))
			content.WriteString(style.Render(fmt.Sprintf("[%s] %-20s", map[bool]string{true: "✓", false: "○"}[m.OptionState[name]], name)))
			content.WriteString("\n")
		}
	}

	// Combine everything
	var ui string
	if m.ActiveTab != 0 {
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

	switch msg := msg.(type) {
	case DirSizeMsg:
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
			if m.ActiveTab == 0 {
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
			} else if m.ActiveTab == 1 {
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
			} else if m.ActiveTab == 2 {
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
			} else if m.ActiveTab == 3 {
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
				if m.FocusedElement == "StartButton" {
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
						selectedItem := m.List.SelectedItem().(CleanItem)
						if selectedItem.size == -1 {
							// Handle parent directory selection
							m.CurrentPath = selectedItem.path
							m.PathInput.SetValue(selectedItem.path)
							// Recalculate directory size when changing directory
							cmds = append(cmds, m.LoadFiles(), m.calculateDirSizeAsync())
							return m, tea.Batch(cmds...)
						}
						// If it's a directory, navigate into it
						info, err := os.Stat(selectedItem.path)
						if err == nil && info.IsDir() {
							m.CurrentPath = selectedItem.path
							m.PathInput.SetValue(selectedItem.path)
							// Recalculate directory size when changing directory
							cmds = append(cmds, m.LoadFiles(), m.calculateDirSizeAsync())
							return m, tea.Batch(cmds...)
						}
					} else if m.ShowDirs && m.DirList.SelectedItem() != nil {
						selectedDir := m.DirList.SelectedItem().(CleanItem)
						m.CurrentPath = selectedDir.path
						m.PathInput.SetValue(selectedDir.path)
						m.ShowDirs = false
						// Recalculate directory size when changing directory
						cmds = append(cmds, m.LoadFiles(), m.calculateDirSizeAsync())
						return m, tea.Batch(cmds...)
					}
				case "dirButton":
					m.ShowDirs = true
					return m, m.loadDirs()
				case "button":
					if m.ActiveTab == 0 {
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
			m.FocusedElement = "path"
			m.PathInput.Focus()
			return m, nil
		case "f2":
			m.ActiveTab = 1
			m.FocusedElement = "exclude"
			m.ExcludeInput.Focus()
			m.PathInput.Blur()
			m.ExtInput.Blur()
			m.SizeInput.Blur()
			return m, nil
		case "f3":
			m.ActiveTab = 2
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
			items = append(items, CleanItem{
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

			items = append(items, CleanItem{
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

			items = append(items, CleanItem{
				path: path,
				size: size,
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
			items = append(items, CleanItem{
				path: parentDir,
				size: -1, // Special value for parent directory
			})
		}

		// Read current directory
		entries, err := os.ReadDir(m.CurrentPath)
		if err != nil {
			return err
		}

		// Create a channel for results
		results := make(chan CleanItem, 100)
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
					results <- CleanItem{
						path: filepath.Join(m.CurrentPath, entry.Name()),
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
			return items[i].(CleanItem).path < items[j].(CleanItem).path
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
		return DirSizeMsg{size: size}
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
		selectedItem := m.List.SelectedItem().(CleanItem)
		if selectedItem.size > 0 { // Only delete files, not directories
			if !m.OptionState["Confirm deletion"] {
				// If confirm deletion is disabled, delete all files
				for _, listItem := range m.List.Items() {
					if fileItem, ok := listItem.(CleanItem); ok && fileItem.size > 0 {
						err := os.Remove(fileItem.path)
						if err != nil {
							m.Err = err
						}
					}
				}
			} else {
				// Delete just the selected file
				err := os.Remove(selectedItem.path)
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

func (m *CleanFilesModel) GetLatestRules() (string, []string, int64, []string) {
	rules, err := m.Rules.GetRules()
	if err != nil {
		return "", []string{}, 0, []string{}
	}

	latestMinSize, _ := utils.ToBytes(rules.MinSize)

	return rules.Path, rules.Extensions, latestMinSize, rules.Exclude
}
