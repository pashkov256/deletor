package views

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

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
	MinSizeInput    textinput.Model
	MaxSizeInput    textinput.Model
	PathInput       textinput.Model
	ExcludeInput    textinput.Model
	OlderInput      textinput.Model
	NewerInput      textinput.Model
	CurrentPath     string
	Extensions      []string
	MinSize         int64
	MaxSize         int64
	Exclude         []string
	Options         []string
	OptionState     map[string]bool
	Err             error
	FocusedElement  string // "pathInput", "extInput","excludeInput","olderInput","newerInput", "minSizeInput","maxSizeInput", "deleteButton","dirButton", "option1", "option2", "option3"
	FileToDelete    *models.CleanItem
	ShowDirs        bool
	DirList         list.Model
	DirSize         int64 // Cached directory size
	CalculatingSize bool  // Flag to indicate size calculation in progress
	FilteredSize    int64 // Total size of filtered files
	FilteredCount   int   // Count of filtered files
	Rules           rules.Rules
	Filemanager     filemanager.FileManager
	TabManager      *tabs.CleanTabManager
}

// Message for directory size updates
type dirSizeMsg struct {
	size int64
}

func InitialCleanModel(rules rules.Rules, fileManager filemanager.FileManager) *CleanFilesModel {
	// Create a temporary model to get rules
	lastestRules, _ := rules.GetRules()
	latestDir := lastestRules.Path
	latestExtensions := lastestRules.Extensions
	latestMinSize := lastestRules.MinSize

	latestExclude := lastestRules.Exclude
	// Initialize inputs
	extInput := textinput.New()
	extInput.Placeholder = "e.g. js,png,zip"
	extInput.SetValue(strings.Join(latestExtensions, ","))
	extInput.PromptStyle = styles.TextInputPromptStyle
	extInput.TextStyle = styles.TextInputTextStyle
	extInput.Cursor.Style = styles.TextInputCursorStyle

	minSizeInput := textinput.New()
	minSizeInput.Placeholder = "e.g. 10b,10kb,10mb,10gb,10tb"
	minSizeInput.SetValue(latestMinSize)
	minSize, _ := utils.ToBytes(latestMinSize)
	minSizeInput.PromptStyle = styles.TextInputPromptStyle
	minSizeInput.TextStyle = styles.TextInputTextStyle
	minSizeInput.Cursor.Style = styles.TextInputCursorStyle

	maxSizeInput := textinput.New()
	maxSizeInput.Placeholder = "e.g. 10b,10kb,10mb,10gb,10tb"
	maxSizeInput.PromptStyle = styles.TextInputPromptStyle
	maxSizeInput.TextStyle = styles.TextInputTextStyle
	maxSizeInput.Cursor.Style = styles.TextInputCursorStyle

	pathInput := textinput.New()
	pathInput.SetValue(latestDir)
	pathInput.PromptStyle = styles.TextInputPromptStyle
	pathInput.TextStyle = styles.TextInputTextStyle
	pathInput.Cursor.Style = styles.TextInputCursorStyle

	excludeInput := textinput.New()
	excludeInput.SetValue(strings.Join(latestExclude, ","))
	excludeInput.PromptStyle = styles.TextInputPromptStyle
	excludeInput.TextStyle = styles.TextInputTextStyle
	excludeInput.Cursor.Style = styles.TextInputCursorStyle

	olderInput := textinput.New()
	olderInput.PromptStyle = styles.TextInputPromptStyle
	olderInput.TextStyle = styles.TextInputTextStyle
	olderInput.Cursor.Style = styles.TextInputCursorStyle

	newerInput := textinput.New()
	newerInput.PromptStyle = styles.TextInputPromptStyle
	newerInput.TextStyle = styles.TextInputTextStyle
	newerInput.Cursor.Style = styles.TextInputCursorStyle

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
		MinSizeInput:    minSizeInput,
		MaxSizeInput:    maxSizeInput,
		PathInput:       pathInput,
		ExcludeInput:    excludeInput,
		OlderInput:      olderInput,
		NewerInput:      newerInput,
		CurrentPath:     latestDir,
		Extensions:      latestExtensions,
		MinSize:         minSize,
		Exclude:         latestExclude,
		OptionState:     tabs.DefaultOptionState,
		FocusedElement:  "list",
		ShowDirs:        false,
		DirList:         dirList,
		DirSize:         0,
		CalculatingSize: false,
		FilteredSize:    0,
		FilteredCount:   0,
		Rules:           rules,
		Filemanager:     fileManager,
	}
}

func (m *CleanFilesModel) Init() tea.Cmd {
	// Set initial focus to path input
	m.FocusedElement = "pathInput"
	m.PathInput.Focus()
	m.TabManager = tabs.NewCleanTabManager(m, tabs.NewCleanTabFactory())

	// If we have a path, load files and calculate size
	if m.CurrentPath != "" {
		return tea.Batch(m.LoadFiles(), m.CalculateDirSizeAsync())
	}

	// Otherwise just return the blink command for the path input
	return textinput.Blink
}

func (m *CleanFilesModel) View() string {
	// --- Tabs rendering ---
	activeTab := m.TabManager.GetActiveTabIndex()
	tabNames := []string{"🗂️ [F1] Main", "🧹 [F2] Filters", "⚙️ [F3] Options", "📖 [F4] Log", "❔ [F5] Help"}
	tabs := make([]string, 5)
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

	//Render active tab
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keyboard events directly
		return m.Handle(msg)

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
		cmds = append(cmds, m.CalculateDirSizeAsync())
		return m, tea.Batch(cmds...)

	case dirSizeMsg:
		// Update the directory size
		m.DirSize = msg.size
		return m, nil

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
	}

	return m, tea.Batch(cmd, tea.Batch(cmds...))
}

func (m *CleanFilesModel) LoadFiles() tea.Cmd {
	return func() tea.Msg {
		var items []list.Item
		var totalFilteredSize int64 = 0
		var filteredCount int = 0

		currentDir := m.CurrentPath

		m.Extensions = utils.ParseExtToSlice(m.ExtInput.Value())
		m.Exclude = utils.ParseExcludeToSlice(m.ExcludeInput.Value())

		var olderDuration, newerDuration time.Time
		var err error

		if m.OlderInput.Value() != "" {
			olderDuration, err = utils.ParseTimeDuration(m.OlderInput.Value())
			if err != nil {
				m.Err = fmt.Errorf("invalid older than time: %v", err)
				return nil
			}
		}

		if m.NewerInput.Value() != "" {
			newerDuration, err = utils.ParseTimeDuration(m.NewerInput.Value())
			if err != nil {
				m.Err = fmt.Errorf("invalid newer than time: %v", err)
				return nil
			}
		}

		minSizeStr := m.MinSizeInput.Value()
		if minSizeStr != "" {
			minSize, err := utils.ToBytes(minSizeStr)
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

		maxSizeStr := m.MaxSizeInput.Value()
		if maxSizeStr != "" {
			maxSize, err := utils.ToBytes(maxSizeStr)
			if err == nil {
				m.MaxSize = maxSize
			} else {
				// If invalid size, reset to 0
				m.MaxSize = 0
			}
		} else {
			// If no size specified, show all files regardless of size
			m.MaxSize = 0
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

		filter := m.Filemanager.NewFileFilter(m.MinSize, m.MaxSize, utils.ParseExtToMap(m.Extensions), m.Exclude, olderDuration, newerDuration)

		// Then collect files
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

			fi, err := fileInfo.Info()
			if err != nil {
				continue
			}

			if !filter.MatchesFilters(fi, currentDir) {
				continue
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

func (m *CleanFilesModel) LoadDirs() tea.Cmd {
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
func (m *CleanFilesModel) CalculateDirSizeAsync() tea.Cmd {
	return func() tea.Msg {
		m.CalculatingSize = true
		size := m.Filemanager.CalculateDirSize(m.CurrentPath)
		m.CalculatingSize = false
		return dirSizeMsg{size: size}
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
						// If send files to trash is enabled, move all files to trash
						if m.OptionState["Send files to trash"] {
							m.Filemanager.MoveFileToTrash(fileItem.Path)
						} else {
							err := os.Remove(fileItem.Path)
							if err != nil {
								m.Err = err
							}
						}
					}
				}
			} else {
				// If send files to trash is enabled, move all files to trash
				if m.OptionState["Send files to trash"] {
					m.Filemanager.MoveFileToTrash(selectedItem.Path)
				} else {
					// Delete just the selected file
					err := os.Remove(selectedItem.Path)
					if err != nil {
						m.Err = err
					}
				}
			}
			return m, m.LoadFiles()
		}
	} else if m.OptionState["Include subfolders"] {
		var olderDuration, newerDuration time.Time
		var err error

		if m.OlderInput.Value() != "" {
			olderDuration, err = utils.ParseTimeDuration(m.OlderInput.Value())
			if err != nil {
				m.Err = fmt.Errorf("invalid older than time: %v", err)
				return m, nil
			}
		}

		if m.NewerInput.Value() != "" {
			newerDuration, err = utils.ParseTimeDuration(m.NewerInput.Value())
			if err != nil {
				m.Err = fmt.Errorf("invalid newer than time: %v", err)
				return m, nil
			}
		}

		if m.OptionState["Send files to trash"] {
			m.Filemanager.MoveFilesToTrash(m.CurrentPath, m.Extensions, m.Exclude, utils.ToBytesOrDefault(m.MinSizeInput.Value()), utils.ToBytesOrDefault(m.MaxSizeInput.Value()), olderDuration, newerDuration)
		} else {
			// Delete all files in the current directory and all subfolders
			m.Filemanager.DeleteFiles(m.CurrentPath, m.Extensions, m.Exclude, utils.ToBytesOrDefault(m.MinSizeInput.Value()), utils.ToBytesOrDefault(m.MaxSizeInput.Value()), olderDuration, newerDuration)
		}

		if m.OptionState["Delete empty subfolders"] {
			m.Filemanager.DeleteEmptySubfolders(m.CurrentPath)
		}

		return m, m.LoadFiles()
	}
	return m, nil
}

// opens the system's file explorer at the specified path
func (m *CleanFilesModel) OpenFileExplorer(path string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("explorer", path)
		case "darwin":
			cmd = exec.Command("open", path)
		default: // "linux", "freebsd", "openbsd", "netbsd"
			cmd = exec.Command("xdg-open", path)
		}
		if err := cmd.Start(); err != nil {
			return tea.Msg(err)
		}
		return nil
	}
}

// Handle processes a key message and returns appropriate model and command
func (m *CleanFilesModel) Handle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		return m.handleTab()
	case "shift+tab":
		return m.handleShiftTab()
	case "up", "down":
		// Always handle arrow keys for list navigation regardless of focus
		// Make list navigation global - arrow keys always navigate the list
		var cmd tea.Cmd
		var cmds []tea.Cmd
		if !m.ShowDirs {
			m.List, cmd = m.List.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.DirList, cmd = m.DirList.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case "f1":
		return m.handleF1()
	case "f2":
		return m.handleF2()
	case "f3":
		return m.handleF3()
	case "f4":
		return m.handleF4()
	case "f5":
		return m.handleF5()
	case "ctrl+r":
		return m, m.LoadFiles()
	case "ctrl+d":
		return m.OnDelete()
	case "ctrl+o":
		return m, m.OpenFileExplorer(m.CurrentPath)
	case "alt+c":
		return m.handleAltC()
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
	case "enter":
		return m.handleEnter()

	case "list":
		var cmd tea.Cmd
		var cmds []tea.Cmd
		if m.CurrentPath != "" {
			if m.ShowDirs {
				m.DirList, cmd = m.DirList.Update(msg)
			} else {
				m.List, cmd = m.List.Update(msg)
			}
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	default:
		switch m.FocusedElement {
		case "pathInput":
			var cmd tea.Cmd
			var cmds []tea.Cmd
			m.PathInput, cmd = m.PathInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case "extInput":
			var cmd tea.Cmd
			var cmds []tea.Cmd
			m.ExtInput, cmd = m.ExtInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case "minSizeInput":
			var cmd tea.Cmd
			var cmds []tea.Cmd
			m.MinSizeInput, cmd = m.MinSizeInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case "maxSizeInput":
			var cmd tea.Cmd
			var cmds []tea.Cmd
			m.MaxSizeInput, cmd = m.MaxSizeInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case "excludeInput":
			var cmd tea.Cmd
			var cmds []tea.Cmd
			m.ExcludeInput, cmd = m.ExcludeInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case "olderInput":
			var cmd tea.Cmd
			var cmds []tea.Cmd
			m.OlderInput, cmd = m.OlderInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case "newerInput":
			var cmd tea.Cmd
			var cmds []tea.Cmd
			m.NewerInput, cmd = m.NewerInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		//If you put the space handling above, then you will not be able to write a space in input.
		if msg.String() == " " {
			return m.handleSpace()
		}

		return m, nil
	}

}

func (m *CleanFilesModel) handleTab() (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	switch activeTab {
	case 0: //Main tab
		if m.CurrentPath == "" { // When no path is set, only allow focusing between path input and start button
			switch m.FocusedElement {
			case "pathInput":
				m.FocusedElement = "startButton"
				m.PathInput.Blur()
			case "startButton":
				m.FocusedElement = "pathInput"
				m.PathInput.Focus()
			}
			return m, nil
		} else { //When path is set
			switch m.FocusedElement {
			case "pathInput":
				m.FocusedElement = "extInput"
				m.PathInput.Blur()
				m.ExtInput.Focus()
			case "extInput":
				m.ExtInput.Blur()
				m.FocusedElement = "list"
			case "list":
				m.FocusedElement = "dirButton"
			case "dirButton":
				m.FocusedElement = "deleteButton"
			case "deleteButton":
				m.FocusedElement = "pathInput"
				m.PathInput.Focus()
			}
		}
	case 1: // Tab navigation for Filters tab
		switch m.FocusedElement {
		case "excludeInput":
			m.ExcludeInput.Blur()
			m.FocusedElement = "minSizeInput"
			m.MinSizeInput.Focus()
		case "minSizeInput":
			m.MinSizeInput.Blur()
			m.FocusedElement = "maxSizeInput"
			m.MaxSizeInput.Focus()
		case "maxSizeInput":
			m.MaxSizeInput.Blur()
			m.FocusedElement = "olderInput"
			m.OlderInput.Focus()
		case "olderInput":
			m.OlderInput.Blur()
			m.FocusedElement = "newerInput"
			m.NewerInput.Focus()
		case "newerInput":
			m.NewerInput.Blur()
			m.FocusedElement = "excludeInput"
			m.ExcludeInput.Focus()
		}
	case 2: // Tab navigation for Options tab
		switch m.FocusedElement {
		case "option1":
			m.FocusedElement = "option2"
		case "option2":
			m.FocusedElement = "option3"
		case "option3":
			m.FocusedElement = "option4"
		case "option4":
			m.FocusedElement = "option5"
		case "option5":
			m.FocusedElement = "option1"
		}

	}

	return m, nil
}

func (m *CleanFilesModel) handleSpace() (tea.Model, tea.Cmd) {
	// Handle space key for options
	if strings.HasPrefix(m.FocusedElement, "option") {
		// Extract option number from the focused element (e.g. "option1" -> "1")
		optionNum := strings.TrimPrefix(m.FocusedElement, "option")
		idx, err := strconv.Atoi(optionNum)
		if err != nil {
			return m, nil
		}
		if idx < 1 || idx > len(tabs.DefaultOption) {
			return m, nil
		}
		idx-- // Convert to 0-based index

		// Get the option name and toggle its state
		optName := tabs.DefaultOption[idx]
		m.OptionState[optName] = !m.OptionState[optName]

		// Keep focus on the current option
		m.FocusedElement = "option" + optionNum

		// If this is the "Show hidden files" option, reload files
		if optName == "Show hidden files" {
			return m, m.LoadFiles()
		}
		return m, nil
	}
	return m, nil
}

func (m *CleanFilesModel) handleShiftTab() (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTabIndex()

	switch activeTab {
	case 0: //Main tab
		if m.CurrentPath == "" { // When no path is set, only allow focusing between path input and start button
			switch m.FocusedElement {
			case "pathInput":
				m.FocusedElement = "startButton"
				m.PathInput.Blur()

			case "startButton":
				m.FocusedElement = "pathInput"
				m.PathInput.Focus()
			}
		} else {
			switch m.FocusedElement {
			case "pathInput":
				m.PathInput.Blur()
				m.FocusedElement = "deleteButton"
			case "deleteButton":
				m.FocusedElement = "dirButton"
			case "dirButton":
				m.FocusedElement = "list"
			case "list":
				m.FocusedElement = "extInput"
				m.ExtInput.Focus()
			case "extInput":
				m.ExtInput.Blur()
				m.FocusedElement = "pathInput"
				m.PathInput.Focus()
			}
		}
	case 1: // Tab navigation for Filters tab
		switch m.FocusedElement {
		case "excludeInput":
			m.ExcludeInput.Blur()
			m.FocusedElement = "newerInput"
			m.NewerInput.Focus()
		case "minSizeInput":
			m.MinSizeInput.Blur()
			m.FocusedElement = "excludeInput"
			m.ExcludeInput.Focus()
		case "maxSizeInput":
			m.MaxSizeInput.Blur()
			m.FocusedElement = "minSizeInput"
			m.MinSizeInput.Focus()
		case "olderInput":
			m.OlderInput.Blur()
			m.FocusedElement = "maxSizeInput"
			m.MaxSizeInput.Focus()
		case "newerInput":
			m.NewerInput.Blur()
			m.FocusedElement = "olderInput"
			m.OlderInput.Focus()
		}
	case 2: // Tab navigation for Options tab
		switch m.FocusedElement {
		case "option1":
			m.FocusedElement = "option5"
		case "option2":
			m.FocusedElement = "option1"
		case "option3":
			m.FocusedElement = "option2"
		case "option4":
			m.FocusedElement = "option3"
		case "option5":
			m.FocusedElement = "option4"
		}
	}

	return m, nil
}

func (m *CleanFilesModel) handleF1() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(0)
	m.FocusedElement = "pathInput"
	pathInput := m.GetPathInput()
	pathInput.Focus()
	return m, nil
}

func (m *CleanFilesModel) handleF2() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(1)
	m.FocusedElement = "excludeInput"
	excludeInput := m.GetExcludeInput()
	excludeInput.Focus()
	return m, nil
}

func (m *CleanFilesModel) handleF3() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(2)
	m.FocusedElement = "option1"
	return m, nil
}
func (m *CleanFilesModel) handleF4() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(3)
	return m, nil
}

func (m *CleanFilesModel) handleF5() (tea.Model, tea.Cmd) {
	m.TabManager.SetActiveTabIndex(4)
	return m, nil
}

func (m *CleanFilesModel) handleAltC() (tea.Model, tea.Cmd) {
	m.MinSizeInput.SetValue("")
	m.ExcludeInput.SetValue("")
	return m, m.LoadFiles()
}

func (m *CleanFilesModel) handleEnter() (tea.Model, tea.Cmd) {
	if m.CurrentPath == "" {
		if m.FocusedElement == "startButton" {
			path := m.PathInput.Value()
			if path != "" {
				expandedPath := utils.ExpandTilde(path)
				if _, err := os.Stat(expandedPath); err == nil {
					m.CurrentPath = expandedPath
					m.FocusedElement = "pathInput"
					m.PathInput.Focus()
					return m, tea.Batch(m.LoadFiles(), m.CalculateDirSizeAsync())
				} else {
					m.SetErr(fmt.Errorf("invalid path: %s", path))
				}
			}
		}
	} else {
		switch m.FocusedElement {
		case "extInput", "minSizeInput", "maxSizeInput", "excludeInput", "olderInput", "newerInput":
			// Update the list of files when pressing Enter in the input fields
			return m, m.LoadFiles()
		case "dirButton":
			m.ShowDirs = true
			return m, m.LoadDirs()
		case "deleteButton":
			return m.OnDelete()
		case "list":
			var cmds []tea.Cmd
			if !m.ShowDirs && m.List.SelectedItem() != nil {
				selectedItem := m.List.SelectedItem().(models.CleanItem)
				if selectedItem.Size == -1 {
					// Handle parent directory selection
					m.CurrentPath = selectedItem.Path
					m.PathInput.SetValue(selectedItem.Path)
					// Recalculate directory size when changing directory
					cmds = append(cmds, m.LoadFiles(), m.CalculateDirSizeAsync())
					return m, tea.Batch(cmds...)
				}
				// If it's a directory, navigate into it
				info, err := os.Stat(selectedItem.Path)
				if err == nil && info.IsDir() {
					m.CurrentPath = selectedItem.Path
					m.PathInput.SetValue(selectedItem.Path)
					// Recalculate directory size when changing directory
					cmds = append(cmds, m.LoadFiles(), m.CalculateDirSizeAsync())
					return m, tea.Batch(cmds...)
				}
			} else if m.ShowDirs && m.DirList.SelectedItem() != nil {
				selectedDir := m.DirList.SelectedItem().(models.CleanItem)
				m.CurrentPath = selectedDir.Path
				m.PathInput.SetValue(selectedDir.Path)
				m.ShowDirs = false
				// Recalculate directory size when changing directory
				cmds = append(cmds, m.LoadFiles(), m.CalculateDirSizeAsync())
				return m, tea.Batch(cmds...)
			}
		default:
			// Handle space key for options
			if strings.HasPrefix(m.FocusedElement, "option") {
				// Extract option number from the focused element (e.g. "option1" -> "1")
				optionNum := strings.TrimPrefix(m.FocusedElement, "option")
				idx, err := strconv.Atoi(optionNum)
				if err != nil {
					return m, nil
				}
				if idx < 1 || idx > len(tabs.DefaultOption) {
					return m, nil
				}
				idx-- // Convert to 0-based index

				// Get the option name and toggle its state
				optName := tabs.DefaultOption[idx]
				m.OptionState[optName] = !m.OptionState[optName]

				// Keep focus on the current option
				m.FocusedElement = "option" + optionNum

				// If this is the "Show hidden files" option, reload files
				if optName == "Show hidden files" {
					return m, m.LoadFiles()
				}
				return m, nil
			}
			return m, nil
		}
	}
	return m, nil
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

func (m *CleanFilesModel) GetMinSizeInput() textinput.Model {
	return m.MinSizeInput
}
func (m *CleanFilesModel) GetMaxSizeInput() textinput.Model {
	return m.MaxSizeInput
}

func (m *CleanFilesModel) GetExcludeInput() textinput.Model {
	return m.ExcludeInput
}

func (m *CleanFilesModel) GetOlderInput() textinput.Model {
	return m.OlderInput
}

func (m *CleanFilesModel) GetNewerInput() textinput.Model {
	return m.NewerInput
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

func (m *CleanFilesModel) SetCurrentPath(path string) {
	m.CurrentPath = path
}

func (m *CleanFilesModel) SetPathInput(input textinput.Model) {
	m.PathInput = input
}

func (m *CleanFilesModel) SetExtInput(input textinput.Model) {
	m.ExtInput = input
}

func (m *CleanFilesModel) SetExcludeInput(input textinput.Model) {
	m.ExcludeInput = input
}

func (m *CleanFilesModel) SetSizeInput(input textinput.Model) {
	m.MinSizeInput = input
}

func (m *CleanFilesModel) GetErr() error {
	return m.Err
}

func (m *CleanFilesModel) SetErr(err error) {
	m.Err = err
}
