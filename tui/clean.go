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
	rules "github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/utils"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	cleanTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#1E90FF")).
			Padding(0, 1)

	sizeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E90FF"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(0, 0).
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

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
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
	sizeStr := formatSize(i.size)

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
	list                list.Model
	extInput            textinput.Model
	sizeInput           textinput.Model
	pathInput           textinput.Model
	excludeInput        textinput.Model
	currentPath         string
	extensions          []string
	minSize             int64
	exclude             []string
	options             []string
	optionState         map[string]bool
	err                 error
	focusedElement      string // "path", "ext", "size", "button", "option1", "option2", "option3"
	waitingConfirmation bool
	fileToDelete        *cleanItem
	showDirs            bool
	dirList             list.Model
	dirSize             int64 // Cached directory size
	calculatingSize     bool  // Flag to indicate size calculation in progress
	filteredSize        int64 // Total size of filtered files
	filteredCount       int   // Count of filtered files
	activeTab           int   // 0 for files, 1 for exclude
}

func initialModel(startDir string, extensions []string, minSize int64, exclude []string) *model {
	// Fetch the latest rules
	latestDir, latestExtensions, latestMinSize, latestExclude := getLatestRules()

	// Update the parameters with the latest rules
	if latestDir != "" {

		startDir = latestDir
	}
	if len(latestExtensions) > 0 {
		extensions = latestExtensions
	}
	if latestMinSize > 0 {
		minSize = latestMinSize
	}
	if len(latestExclude) > 0 {
		exclude = latestExclude
	}
	// Initialize inputs
	extInput := textinput.New()
	extInput.Placeholder = "(e.g. js,png,zip)..."
	extInput.SetValue(strings.Join(extensions, ","))
	extInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	extInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	extInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))

	sizeInput := textinput.New()
	sizeInput.Placeholder = "(e.g. 10kb,10mb,10b)..."
	sizeInput.SetValue(utils.FormatSize(minSize))
	sizeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	sizeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	sizeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))

	pathInput := textinput.New()
	pathInput.SetValue(startDir)
	pathInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	pathInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	pathInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))

	excludeInput := textinput.New()
	excludeInput.Placeholder = "Exclude specific files/paths (e.g. data,backup)"
	excludeInput.SetValue(strings.Join(exclude, ","))
	excludeInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	excludeInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	excludeInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))

	// Create a proper delegate with visible height
	delegate := list.NewDefaultDelegate()

	delegate.SetHeight(1)
	delegate.SetSpacing(1)
	delegate.ShowDescription = false

	// Стили элементов
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
	l.Styles.Title = cleanTitleStyle

	// Create directory list with same delegate
	dirList := list.New([]list.Item{}, delegate, 30, 10)
	dirList.SetShowTitle(true)
	dirList.Title = "Directories"
	dirList.SetShowStatusBar(true)
	dirList.SetFilteringEnabled(false)
	dirList.SetShowHelp(false)
	dirList.Styles.Title = cleanTitleStyle

	// Define options in fixed order
	options := []string{
		"Show hidden files",
		"Confirm deletion",
	}

	optionState := map[string]bool{
		"Show hidden files": false,
		"Confirm deletion":  false,
	}

	return &model{
		list:                l,
		extInput:            extInput,
		sizeInput:           sizeInput,
		pathInput:           pathInput,
		excludeInput:        excludeInput,
		currentPath:         startDir,
		extensions:          extensions,
		minSize:             minSize,
		exclude:             exclude,
		options:             options,
		optionState:         optionState,
		focusedElement:      "list",
		waitingConfirmation: false,
		fileToDelete:        nil,
		showDirs:            false,
		dirList:             dirList,
		dirSize:             0,
		calculatingSize:     false,
		filteredSize:        0,
		filteredCount:       0,
		activeTab:           0,
	}
}

func (m *model) Init() tea.Cmd {
	m.focusedElement = "path"
	m.pathInput.Focus()
	return tea.Batch(textinput.Blink, m.loadFiles(), m.calculateDirSizeAsync())
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
			minSize, err := toBytes(sizeStr)
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
			return fmt.Errorf("error reading directory: %v", err)
		}

		// Добавим родительскую директорию
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

	// Wait for all goroutines to finish
	wg.Wait()

	return totalSize
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
		h, v := appStyle.GetFrameSize()
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
		// Global hotkeys that work regardless of focus
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+r": // Refresh files
			return m, m.loadFiles()
		case "ctrl+d": // Toggle directory view
			m.showDirs = !m.showDirs
			if m.showDirs {
				return m, m.loadDirs()
			}
			return m, m.loadFiles()
		case "ctrl+o": // Open current directory in file explorer
			cmd := openFileExplorer(m.currentPath)
			return m, cmd
		case "left", "right": // Tab switching
			if msg.String() == "left" && m.activeTab > 0 {
				m.activeTab--
				if m.activeTab == 1 {
					m.excludeInput.Focus()
					m.focusedElement = "exclude"
				}
			}
			if msg.String() == "right" && m.activeTab < 1 {
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
			return m, nil
		case "f3":
			m.activeTab = 2
			m.focusedElement = "option1"
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

		// Handle tab key to cycle through elements
		if msg.String() == "tab" {
			switch m.activeTab {
			case 0: // Main tab
				switch m.focusedElement {
				case "path":
					m.pathInput.Blur()
					m.focusedElement = "ext"
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
			case 1: // Filters tab
				switch m.focusedElement {
				case "exclude":
					m.excludeInput.Blur()
					m.focusedElement = "size"
					m.sizeInput.Focus()
				case "size":
					m.sizeInput.Blur()
					m.focusedElement = "exclude"
					m.excludeInput.Focus()
				default:
					m.focusedElement = "exclude"
					m.excludeInput.Focus()
				}
			case 2: // Options tab
				if m.focusedElement == "" {
					m.focusedElement = "option1"
				} else if m.focusedElement == "option1" {
					m.focusedElement = "option2"
				} else {
					m.focusedElement = "option1"
				}
			}
			return m, nil
		}

		if msg.String() == "shift+tab" {
			switch m.activeTab {
			case 0: // Main tab
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
			case 1: // Filters tab
				switch m.focusedElement {
				case "exclude":
					m.excludeInput.Blur()
					m.focusedElement = "size"
					m.sizeInput.Focus()
				case "size":
					m.sizeInput.Blur()
					m.focusedElement = "exclude"
					m.excludeInput.Focus()
				default:
					m.focusedElement = "exclude"
					m.excludeInput.Focus()
				}
			case 2: // Options tab
				if m.focusedElement == "" {
					m.focusedElement = "option2"
				} else if m.focusedElement == "option2" {
					m.focusedElement = "option1"
				} else {
					m.focusedElement = "option2"
				}
			}
			return m, nil
		}

		// Handle inputs based on current focus
		if m.pathInput.Focused() {
			switch msg.String() {
			case "tab", "enter":
				m.pathInput.Blur()
				m.extInput.Focus()
				m.focusedElement = "ext"
				return m, nil
			case "esc":
				m.pathInput.Blur()
				m.focusedElement = "list"
				return m, nil
			default:
				m.pathInput, cmd = m.pathInput.Update(msg)
				cmds = append(cmds, cmd)
				// Reload files if enter is pressed
				if msg.String() == "enter" {
					// Update path if valid
					newPath := m.pathInput.Value()
					if _, err := os.Stat(newPath); err == nil {
						m.currentPath = newPath
						cmds = append(cmds, m.loadFiles(), m.calculateDirSizeAsync())
					} else {
						m.err = fmt.Errorf("invalid path: %s", newPath)
					}
				}
				return m, tea.Batch(cmds...)
			}
		}

		if m.extInput.Focused() {
			switch msg.String() {
			case "tab":
				m.extInput.Blur()
				m.sizeInput.Focus()
				m.focusedElement = "size"
				return m, nil
			case "enter":
				m.extInput.Blur()
				m.focusedElement = "list"
				// Parse extensions and reload files
				cmds = append(cmds, m.loadFiles())
				return m, tea.Batch(cmds...)
			case "esc":
				m.extInput.Blur()
				m.focusedElement = "list"
				return m, nil
			default:
				m.extInput, cmd = m.extInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		if m.sizeInput.Focused() {
			switch msg.String() {
			case "tab":
				m.sizeInput.Blur()
				m.focusedElement = "exclude"
				m.excludeInput.Focus()
				return m, nil
			case "enter":
				m.sizeInput.Blur()
				m.focusedElement = "exclude"
				m.excludeInput.Focus()
				// Parse size and reload files
				cmds = append(cmds, m.loadFiles())
				return m, tea.Batch(cmds...)
			case "esc":
				m.sizeInput.Blur()
				m.focusedElement = "exclude"
				m.excludeInput.Focus()
				return m, nil
			default:
				m.sizeInput, cmd = m.sizeInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		if m.excludeInput.Focused() {
			switch msg.String() {
			case "tab":
				m.excludeInput.Blur()
				m.focusedElement = "size"
				m.sizeInput.Focus()
				return m, nil
			case "enter":
				m.excludeInput.Blur()
				m.focusedElement = "size"
				m.sizeInput.Focus()
				// Parse exclude and reload files
				cmds = append(cmds, m.loadFiles())
				return m, tea.Batch(cmds...)
			case "esc":
				m.excludeInput.Blur()
				m.focusedElement = "size"
				m.sizeInput.Focus()
				return m, nil
			default:
				m.excludeInput, cmd = m.excludeInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		// Handle enter key for actions
		if msg.String() == "enter" {
			switch m.focusedElement {
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
				if m.list.SelectedItem() != nil {
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
				}
			case "option1", "option2":
				idx := 0
				if m.focusedElement == "option2" {
					idx = 1
				}
				if idx < len(m.options) {
					optName := m.options[idx]
					m.optionState[optName] = !m.optionState[optName]
					m.focusedElement = "option" + fmt.Sprintf("%d", idx+1)
					return m, nil
				}
			}
		}

		// Handle space key for options
		if msg.String() == " " && m.activeTab == 2 {
			if m.focusedElement == "option1" || m.focusedElement == "option2" {
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

		if m.activeTab == 2 && msg.String() == "Enter" {
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(m.options) {
				optName := m.options[idx]
				m.optionState[optName] = !m.optionState[optName]
				return m, nil
			}
		}

	}

	return m, tea.Batch(cmds...)
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

func (m *model) View() string {
	// --- Tabs styles ---
	tabStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(0, 1)

	activeTabStyle := tabStyle.Copy().
		BorderForeground(lipgloss.Color("#1E90FF")).
		Foreground(lipgloss.Color("#1E90FF")).
		Bold(true)

	// --- Tabs rendering ---
	tabNames := []string{"🗂️ [F1] Main", "🧹 [F2] Filters", "⚙️ [F3] Options"}
	tabs := make([]string, 3)
	for i, name := range tabNames {
		style := tabStyle
		if m.activeTab == i {
			style = activeTabStyle
		}
		tabs[i] = style.Render(name)
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// --- Content rendering ---
	var content strings.Builder
	if m.activeTab == 0 {
		pathStyle := borderStyle.Copy()
		if m.focusedElement == "path" {
			pathStyle = pathStyle.BorderForeground(lipgloss.Color("#1E90FF"))
		}
		content.WriteString(pathStyle.Render("Current Path: " + m.pathInput.View()))
		content.WriteString("\n")

		extStyle := borderStyle.Copy()
		if m.focusedElement == "ext" {
			extStyle = extStyle.BorderForeground(lipgloss.Color("#1E90FF"))
		}
		content.WriteString(extStyle.Render("Extensions: " + m.extInput.View()))
		content.WriteString("\n")

		var activeList list.Model
		if m.showDirs {
			activeList = m.dirList
		} else {
			activeList = m.list
		}
		fileCount := len(activeList.Items())
		filteredSizeText := formatSize(m.filteredSize)
		content.WriteString("\n")
		if !m.showDirs {
			content.WriteString(cleanTitleStyle.Render(fmt.Sprintf("Selected files (%d) • Size of selected files: %s",
				m.filteredCount, filteredSizeText)))
		} else {
			content.WriteString(cleanTitleStyle.Render(fmt.Sprintf("Directories in %s (%d)",
				filepath.Base(m.currentPath), fileCount)))
		}
		content.WriteString("\n")
		listStyle := borderStyle.Copy()
		if m.focusedElement == "list" {
			listStyle = listStyle.BorderForeground(lipgloss.Color("#1E90FF"))
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
					sizeStr = formatSize(item.size)
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
		content.WriteString("\n")

		if m.focusedElement == "dirButton" {
			content.WriteString(dirButtonFocusedStyle.Render("➡️ Change Directory"))
		} else {
			content.WriteString(dirButtonStyle.Render("➡️ Change Directory"))
		}
		content.WriteString("\n")
		if m.focusedElement == "button" {
			content.WriteString(buttonFocusedStyle.Render("🗑️ Delete Selected File"))
		} else {
			content.WriteString(buttonStyle.Render("🗑️ Delete Selected File"))
		}
		content.WriteString("\n")
	} else if m.activeTab == 1 {
		// Filters tab: excludeInput и min size, оба с подписью слева
		excludeStyle := borderStyle.Copy()
		if m.focusedElement == "exclude" {
			excludeStyle = excludeStyle.BorderForeground(lipgloss.Color("#1E90FF"))
		}
		m.excludeInput.Placeholder = "specific files/paths (e.g. data,backup)"
		content.WriteString(excludeStyle.Render("exclude: " + m.excludeInput.View()))
		content.WriteString("\n")
		sizeStyle := borderStyle.Copy()
		if m.focusedElement == "size" {
			sizeStyle = sizeStyle.BorderForeground(lipgloss.Color("#1E90FF"))
		}
		content.WriteString(sizeStyle.Render("Min size: " + m.sizeInput.View()))
		content.WriteString("\n")
	} else if m.activeTab == 2 {
		// Options tab: только options, без надписи Options
		for i, name := range m.options {
			style := optionStyle
			if m.optionState[name] {
				style = selectedOptionStyle
			}
			if m.focusedElement == fmt.Sprintf("option%d", i+1) {
				style = optionFocusedStyle
			}
			content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", i+1)))
			content.WriteString(style.Render(fmt.Sprintf("[%s] %-20s", map[bool]string{true: "✓", false: "○"}[m.optionState[name]], name)))
			content.WriteString("\n")
		}
	}

	ui := tabsRow + "\n" + content.String()

	// Help
	ui += "\nArrow keys: navigate • Tab: cycle focus • Enter: select/confirm • Esc: back to list\n"
	ui += "Ctrl+R: refresh • Ctrl+D: toggle dirs • Ctrl+O: open in explorer • Ctrl+C: quit\n"
	ui += "Left/Right: switch tabs"

	if m.err != nil {
		ui += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	return appStyle.Render(ui)
}

func Run(startDir string, extensions []string, minSize int64, exclude []string) error {
	p := tea.NewProgram(initialModel(startDir, extensions, minSize, exclude),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithFPS(30),
		tea.WithInputTTY(),
		tea.WithOutput(os.Stderr),
	)
	_, err := p.Run()
	return err
}

func toBytes(sizeStr string) (int64, error) {
	var value float64
	var unit string

	_, err := fmt.Sscanf(sizeStr, "%f%s", &value, &unit)
	if err != nil {
		return 0, fmt.Errorf("invalid format")
	}

	unit = strings.ToLower(unit)
	multiplier := int64(1)

	switch unit {
	case "b":
		multiplier = 1
	case "kb":
		multiplier = 1024
	case "mb":
		multiplier = 1024 * 1024
	case "gb":
		multiplier = 1024 * 1024 * 1024
	case "tb":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return int64(value * float64(multiplier)), nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getLatestRules() (string, []string, int64, []string) {
	// Load saved rules
	savedRules := rules.GetRules()

	// Initialize default values
	startDir := ""
	extensions := []string{}
	minSize := int64(0)
	exclude := []string{}

	// Use saved directory if provided and valid
	if savedRules.Path != "" {
		if _, err := os.Stat(savedRules.Path); err == nil {
			startDir = savedRules.Path
		}
	}

	// Use saved extensions if provided
	if len(savedRules.Extensions) > 0 {
		extensions = savedRules.Extensions
	}

	// Use saved minimum size if provided
	if savedRules.MinSize != "" {
		if size, err := utils.ToBytes(savedRules.MinSize); err == nil {
			minSize = size
		}
	}

	// Use saved extensions if provided
	if len(savedRules.Exclude) > 0 {
		exclude = savedRules.Exclude
	}

	return startDir, extensions, minSize, exclude
}
