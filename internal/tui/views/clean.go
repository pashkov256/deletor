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
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/logging"
	"github.com/pashkov256/deletor/internal/models"
	rules "github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/errors"
	"github.com/pashkov256/deletor/internal/tui/help"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/tui/tabs/clean"
	"github.com/pashkov256/deletor/internal/utils"
	"github.com/pashkov256/deletor/internal/validation"
)

type CleanFilesModel struct {
	List              list.Model
	ExtInput          textinput.Model
	MinSizeInput      textinput.Model
	MaxSizeInput      textinput.Model
	PathInput         textinput.Model
	ExcludeInput      textinput.Model
	OlderInput        textinput.Model
	NewerInput        textinput.Model
	CurrentPath       string
	Extensions        []string
	MinSize           int64
	MaxSize           int64
	Exclude           []string
	Options           []string
	OptionState       map[string]bool
	FocusedElement    string // "pathInput", "extInput","excludeInput","olderInput","newerInput", "minSizeInput","maxSizeInput", "deleteButton","dirButton", "clean_option_1", "clean_option_2", "clean_option_3"
	FileToDelete      *models.CleanItem
	ShowDirs          bool
	DirList           list.Model
	DirSize           int64 // Cached directory size
	CalculatingSize   bool  // Flag to indicate size calculation in progress
	FilteredSize      int64 // Total size of filtered files
	FilteredCount     int   // Count of filtered files
	Rules             rules.Rules
	Filemanager       filemanager.FileManager
	TabManager        *clean.CleanTabManager
	Validator         *validation.Validator
	Logger            *logging.Logger
	Error             *errors.Error
	IsLaunched        bool            // Track if the app has been launched
	SelectedFiles     map[string]bool // Track selected files by path
	SelectedSize      int64           // Track selected files size
	SelectedCount     int             // Track selected files count
	LastSelectedIndex int             // Track last selected index for range selection
}

// Message for directory size updates
type DirSizeMsg struct {
	Size int64
}

func InitialCleanModel(rules rules.Rules, fileManager filemanager.FileManager, validator *validation.Validator) *CleanFilesModel {
	// Create a temporary model to get rules
	lastestRules, _ := rules.GetRules()
	latestDir := lastestRules.Path
	latestExtensions := lastestRules.Extensions
	latestMinSize := lastestRules.MinSize
	latestMaxSize := lastestRules.MaxSize
	latestExclude := lastestRules.Exclude
	latestOlderThan := lastestRules.OlderThan
	latestNewerThan := lastestRules.NewerThan

	// Initialize inputs
	extInput := textinput.New()
	extInput.SetValue(strings.Join(latestExtensions, ","))
	extInput.PromptStyle = styles.TextInputPromptStyle
	extInput.TextStyle = styles.TextInputTextStyle
	extInput.Cursor.Style = styles.TextInputCursorStyle

	minSizeInput := textinput.New()
	minSizeInput.SetValue(latestMinSize)
	minSize, _ := utils.ToBytes(latestMinSize)
	minSizeInput.PromptStyle = styles.TextInputPromptStyle
	minSizeInput.TextStyle = styles.TextInputTextStyle
	minSizeInput.Cursor.Style = styles.TextInputCursorStyle

	maxSizeInput := textinput.New()
	maxSizeInput.SetValue(latestMaxSize)
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
	olderInput.SetValue(latestOlderThan)
	olderInput.PromptStyle = styles.TextInputPromptStyle
	olderInput.TextStyle = styles.TextInputTextStyle
	olderInput.Cursor.Style = styles.TextInputCursorStyle

	newerInput := textinput.New()
	newerInput.SetValue(latestNewerThan)
	newerInput.PromptStyle = styles.TextInputPromptStyle
	newerInput.TextStyle = styles.TextInputTextStyle
	newerInput.Cursor.Style = styles.TextInputCursorStyle

	extInput.Placeholder = "e.g. js,png,zip"
	minSizeInput.Placeholder = "e.g. 10b,10kb,10mb,10gb,10tb"
	maxSizeInput.Placeholder = "e.g. 10b,10kb,10mb,10gb,10tb"
	excludeInput.Placeholder = "specific files/paths (e.g. data,backup)"
	olderInput.Placeholder = "e.g. 60 min, 1 hour, 7 days, 1 month"
	newerInput.Placeholder = "e.g. 60 min, 1 hour, 7 days, 1 month"

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
	l.Styles.Title = styles.ListTitleStyle

	// Create directory list with same delegate
	dirList := list.New([]list.Item{}, delegate, 30, 10)
	dirList.SetShowTitle(true)
	dirList.Title = "Directories"
	dirList.SetShowStatusBar(true)
	dirList.SetFilteringEnabled(false)
	dirList.SetShowHelp(false)
	dirList.Styles.Title = styles.ListTitleStyle

	// Initialize logger
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Error getting user config dir: %v\n", err)
		return nil
	}

	logPath := filepath.Join(userConfigDir, "deletor", "deletor.log")

	// Expand the path if it contains tilde
	expandedPath := utils.ExpandTilde(latestDir)

	// Create model first
	model := &CleanFilesModel{
		List:         l,
		ExtInput:     extInput,
		MinSizeInput: minSizeInput,
		MaxSizeInput: maxSizeInput,
		PathInput:    pathInput,
		ExcludeInput: excludeInput,
		OlderInput:   olderInput,
		NewerInput:   newerInput,
		CurrentPath:  expandedPath,
		Extensions:   latestExtensions,
		MinSize:      minSize,
		Exclude:      latestExclude,
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
		FocusedElement:    "list",
		ShowDirs:          false,
		DirList:           dirList,
		DirSize:           0,
		CalculatingSize:   false,
		FilteredSize:      0,
		FilteredCount:     0,
		Rules:             rules,
		Filemanager:       fileManager,
		Validator:         validator,
		IsLaunched:        expandedPath != "", // Set IsLaunched to true if path is already set
		SelectedFiles:     make(map[string]bool),
		SelectedSize:      0,
		SelectedCount:     0,
		LastSelectedIndex: -1,
	}

	// Initialize tab manager
	model.TabManager = clean.NewCleanTabManager(model, clean.NewCleanTabFactory())

	// Initialize logger with callback
	logger, err := logging.NewLogger(logPath, func(stats *logging.ScanStatistics) {
		if model.TabManager != nil {
			if logTab, ok := model.TabManager.GetActiveTab().(*clean.LogTab); ok {
				logTab.UpdateStats(stats)
			}
		}
	})
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return nil
	}

	model.Logger = logger

	// Log initial message
	logger.Log(logging.INFO, "Application started")

	return model
}

func (m *CleanFilesModel) Init() tea.Cmd {
	// Set initial focus to path input
	m.FocusedElement = "pathInput"
	m.PathInput.Focus()
	m.TabManager = clean.NewCleanTabManager(m, clean.NewCleanTabFactory())

	// If we have a path, load files and calculate size
	if m.CurrentPath != "" {
		// Ensure path is expanded
		expandedPath := utils.ExpandTilde(m.CurrentPath)
		if _, err := os.Stat(expandedPath); err == nil {
			m.CurrentPath = expandedPath
			return tea.Batch(m.LoadFiles(), m.CalculateDirSizeAsync())
		}
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
		tabs[i] = zone.Mark(fmt.Sprintf("tab_%d", i), style.Render(name))
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

	// --- Content rendering ---
	var content strings.Builder
	content.WriteString(tabsRow)
	content.WriteString("\n")

	// Render active tab content
	content.WriteString(m.TabManager.GetActiveTab().View())

	// Add error message if there is one
	if m.Error != nil && m.Error.IsVisible() {
		errorStyle := errors.GetStyle(m.Error.GetType())
		content.WriteString("\n")
		content.WriteString(errorStyle.Render(m.Error.GetMessage()))
	}

	// Combine everything
	var ui string
	if activeTab != 0 {
		ui = content.String()
	} else {
		ui = lipgloss.JoinVertical(lipgloss.Left,
			content.String(),
			help.NavigateHelpText,
			help.ListHelpText,
			help.CleanHelpText,
		)
	}

	return zone.Scan(styles.AppStyle.Render(ui))
}

func (m *CleanFilesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.Handle(msg)

	case tea.MouseMsg:
		// nolint:staticcheck
		if msg.Type == tea.MouseLeft && msg.Action == tea.MouseActionPress {
			// Handle tab clicks
			for i := 0; i < 5; i++ {
				if zone.Get(fmt.Sprintf("tab_%d", i)).InBounds(msg) {
					m.TabManager.SetActiveTabIndex(i)
					switch i {
					case 0:
						model, cmd := m.handleF1()
						return model, cmd
					case 1:
						model, cmd := m.handleF2()
						return model, cmd
					case 2:
						model, cmd := m.handleF3()
						return model, cmd
					case 3:
						model, cmd := m.handleF4()
						return model, cmd
					case 4:
						model, cmd := m.handleF5()
						return model, cmd
					}
				}
			}

			// Handle path input click
			if zone.Get("main_path_input").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "pathInput"
				m.PathInput.Focus()
				return m, nil
			}

			// Handle start button click
			if zone.Get("main_start_button").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "startButton"
				return m.HandlePressStartButton()
			}

			// Handle extensions input click
			if zone.Get("main_ext_input").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "extInput"
				m.ExtInput.Focus()
				return m, nil
			}

			// Handle directory button click
			if zone.Get("main_dir_button").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "dirButton"
				return m.HandlePressDirButton()
			}

			// Handle delete button click
			if zone.Get("main_delete_button").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "deleteButton"
				return m.OnDelete()
			}

			// Handle filters tab inputs
			if zone.Get("filters_exclude_input").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "excludeInput"
				m.ExcludeInput.Focus()
				return m, nil
			}

			if zone.Get("filters_min_size_input").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "minSizeInput"
				m.MinSizeInput.Focus()
				return m, nil
			}

			if zone.Get("filters_max_size_input").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "maxSizeInput"
				m.MaxSizeInput.Focus()
				return m, nil
			}

			if zone.Get("filters_older_input").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "olderInput"
				m.OlderInput.Focus()
				return m, nil
			}

			if zone.Get("filters_newer_input").InBounds(msg) {
				m.blurAllInputs()
				m.FocusedElement = "newerInput"
				m.NewerInput.Focus()
				return m, nil
			}

			// Handle options tab clicks
			for i, option := range options.DefaultCleanOption {
				if zone.Get(fmt.Sprintf("clean_option_%d", i+1)).InBounds(msg) {
					m.blurAllInputs()
					m.FocusedElement = fmt.Sprintf("clean_option_%d", i+1)
					m.OptionState[option] = !m.OptionState[option]

					// If this is the options.ShowHiddenFiles option, reload files
					if option == options.ShowHiddenFiles {
						return m, m.LoadFiles()
					}
					return m, nil
				}
			}
		}

	case tea.WindowSizeMsg:
		// Properly set both width and height
		h, v := styles.AppStyle.GetFrameSize()
		listHeight := (msg.Height - v - 15) * 65 / 100
		if listHeight < 5 {
			listHeight = 5
		}
		m.List.SetSize(msg.Width-h, listHeight)
		m.DirList.SetSize(msg.Width-h, listHeight)

		cmds = append(cmds, m.LoadFiles())
		cmds = append(cmds, m.CalculateDirSizeAsync())
		return m, tea.Batch(cmds...)

	case DirSizeMsg:
		m.DirSize = msg.Size
		return m, nil

	case []list.Item:
		if m.ShowDirs {
			m.DirList.SetItems(msg)
		} else {
			selectedIdx := m.List.Index()
			m.List.SetItems(msg)
			if selectedIdx < len(msg) {
				m.List.Select(selectedIdx)
			}
		}
		return m, nil

	case *errors.Error:
		m.Error = msg
		return m, nil

	case error:
		return m, func() tea.Msg {
			return errors.New(errors.ErrorTypeValidation, msg.Error())
		}
	}

	switch m.FocusedElement {
	case "pathInput":
		m.PathInput, cmd = m.PathInput.Update(msg)
		cmds = append(cmds, cmd)
	case "extInput":
		m.ExtInput, cmd = m.ExtInput.Update(msg)
		cmds = append(cmds, cmd)
	case "minSizeInput":
		m.MinSizeInput, cmd = m.MinSizeInput.Update(msg)
		cmds = append(cmds, cmd)
	case "maxSizeInput":
		m.MaxSizeInput, cmd = m.MaxSizeInput.Update(msg)
		cmds = append(cmds, cmd)
	case "excludeInput":
		m.ExcludeInput, cmd = m.ExcludeInput.Update(msg)
		cmds = append(cmds, cmd)
	case "olderInput":
		m.OlderInput, cmd = m.OlderInput.Update(msg)
		cmds = append(cmds, cmd)
	case "newerInput":
		m.NewerInput, cmd = m.NewerInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmd, tea.Batch(cmds...))
}

func (m *CleanFilesModel) LoadFiles() tea.Cmd {
	return func() tea.Msg {
		var items []list.Item
		var totalFilteredSize int64 = 0
		var filteredCount = 0

		currentDir := m.CurrentPath

		m.Extensions = utils.ParseExtToSlice(m.ExtInput.Value())
		m.Exclude = utils.ParseExcludeToSlice(m.ExcludeInput.Value())

		var olderDuration, newerDuration time.Time
		var err error

		if m.OlderInput.Value() != "" {
			olderDuration, err = utils.ParseTimeDuration(m.OlderInput.Value())
			if err != nil {
				return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid older than time: %v", err))
			}
		}

		if m.NewerInput.Value() != "" {
			newerDuration, err = utils.ParseTimeDuration(m.NewerInput.Value())
			if err != nil {
				return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid newer than time: %v", err))
			}
		}

		minSizeStr := m.MinSizeInput.Value()
		if minSizeStr != "" {
			minSize, err := utils.ToBytes(minSizeStr)
			if err == nil {
				m.MinSize = minSize
			} else {
				return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid minimum size: %v", err))
			}
		} else {
			m.MinSize = 0
		}

		maxSizeStr := m.MaxSizeInput.Value()
		if maxSizeStr != "" {
			maxSize, err := utils.ToBytes(maxSizeStr)
			if err == nil {
				m.MaxSize = maxSize
			} else {
				return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid maximum size: %v", err))
			}
		} else {
			m.MaxSize = 0
		}

		// Only show directory error if app has been launched
		if m.IsLaunched {
			fileInfos, err := os.ReadDir(currentDir)
			if err != nil {
				return errors.New(errors.ErrorTypeFileSystem, fmt.Sprintf("Failed to read directory: %v", err))
			}

			// Add to parent directory
			parentDir := filepath.Dir(currentDir)
			if parentDir != currentDir {
				items = append(items, models.CleanItem{
					Path:  parentDir,
					Size:  -1, // Special value for parent directory
					IsDir: true,
				})
			}

			filter := m.Filemanager.NewFileFilter(m.MinSize, m.MaxSize, utils.ParseExtToMap(m.Extensions), m.Exclude, olderDuration, newerDuration)

			// Then collect files
			for _, fileInfo := range fileInfos {
				if fileInfo.IsDir() {
					continue
				}

				// Skip hidden files unless enabled
				if !m.OptionState[options.ShowHiddenFiles] && strings.HasPrefix(fileInfo.Name(), ".") {
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
					Path:  path,
					Size:  size,
					IsDir: false,
				})
			}
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
				Path:  parentDir,
				Size:  -1, // Special value for parent directory
				IsDir: true,
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
					if !m.OptionState[options.ShowHiddenFiles] && strings.HasPrefix(entry.Name(), ".") {
						continue
					}
					results <- models.CleanItem{
						Path:  filepath.Join(m.CurrentPath, entry.Name()),
						Size:  0,
						IsDir: true,
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
		return DirSizeMsg{Size: size}
	}
}

func (m *CleanFilesModel) OnDelete() (tea.Model, tea.Cmd) {
	// Create statistics for this operation
	stats := &logging.ScanStatistics{
		StartTime:     time.Now(),
		Directory:     m.CurrentPath,
		OperationType: "delete",
	}

	// Initialize counters
	stats.TotalFiles = 0
	stats.TotalSize = 0
	stats.DeletedFiles = 0
	stats.DeletedSize = 0
	stats.TrashedFiles = 0
	stats.TrashedSize = 0

	if len(m.SelectedFiles) > 0 {
		stats.TotalFiles = int64(m.SelectedCount)
		stats.TotalSize = m.SelectedSize

		if m.OptionState[options.SendFilesToTrash] {
			for filePath := range m.SelectedFiles {
				m.Filemanager.MoveFileToTrash(filePath)
				delete(m.SelectedFiles, filePath)
			}
			stats.TrashedFiles = int64(m.SelectedCount)
			stats.TrashedSize = m.SelectedSize
		} else {
			for filePath := range m.SelectedFiles {
				os.Remove(filePath)
				delete(m.SelectedFiles, filePath)
			}
			stats.DeletedFiles = int64(m.SelectedCount)
			stats.DeletedSize = m.SelectedSize
		}
		// Update end time
		stats.EndTime = time.Now()

		// Log final statistics
		if m.Logger != nil {
			m.Logger.Log(logging.INFO, fmt.Sprintf("Delete operation completed. Statistics: %+v", stats))
			m.Logger.UpdateStats(stats)
		}

		// Update all LogTabs
		if m.TabManager != nil {
			for _, tab := range m.TabManager.GetAllTabs() {
				if logTab, ok := tab.(*clean.LogTab); ok {
					logTab.UpdateStats(stats)
				}
			}
		}
		m.SelectedFiles = make(map[string]bool)
		m.SelectedCount = 0
		m.SelectedSize = 0

		return m, m.LoadFiles()
	}
	if m.OptionState[options.IncludeSubfolders] {
		var olderDuration, newerDuration time.Time
		var err error

		if m.OlderInput.Value() != "" {
			olderDuration, err = utils.ParseTimeDuration(m.OlderInput.Value())
			if err != nil {
				return m, func() tea.Msg {
					return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid older than time: %v", err))
				}
			}
		}

		if m.NewerInput.Value() != "" {
			newerDuration, err = utils.ParseTimeDuration(m.NewerInput.Value())
			if err != nil {
				return m, func() tea.Msg {
					return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid newer than time: %v", err))
				}
			}
		}

		if m.OptionState[options.SendFilesToTrash] {
			m.Filemanager.MoveFilesToTrash(m.CurrentPath, m.Extensions, m.Exclude, utils.ToBytesOrDefault(m.MinSizeInput.Value()), utils.ToBytesOrDefault(m.MaxSizeInput.Value()), olderDuration, newerDuration)
		} else {
			// Delete all files in the current directory and all subfolders
			m.Filemanager.DeleteFiles(m.CurrentPath, m.Extensions, m.Exclude, utils.ToBytesOrDefault(m.MinSizeInput.Value()), utils.ToBytesOrDefault(m.MaxSizeInput.Value()), olderDuration, newerDuration)
		}

		if m.OptionState[options.DeleteEmptySubfolders] {
			m.Filemanager.DeleteEmptySubfolders(m.CurrentPath)
		}

		return m, m.LoadFiles()
	}

	// Process files based on Confirm deletion option
	if m.OptionState[options.ConfirmDeletion] {
		// Single file deletion mode
		selectedItem := m.List.SelectedItem()
		if selectedItem == nil {
			if m.Logger != nil {
				m.Logger.Log(logging.DEBUG, "No file selected for deletion")
			}
			return m, nil
		}

		item := selectedItem.(models.CleanItem)
		// Skip parent directory entry
		if item.Size == -1 {
			return m, nil
		}

		stats.TotalFiles = 1
		stats.TotalSize = item.Size

		if m.OptionState[options.SendFilesToTrash] {
			// Move to trash
			m.Filemanager.MoveFileToTrash(item.Path)
			stats.TrashedFiles = 1
			stats.TrashedSize = item.Size

			if m.Logger != nil {
				m.Logger.Log(logging.DEBUG, fmt.Sprintf("Moved to trash: %s (size: %d)", item.Path, item.Size))
			}
		} else {
			// Permanent deletion
			if err := os.Remove(item.Path); err != nil {
				if m.Logger != nil {
					m.Logger.Log(logging.ERROR, fmt.Sprintf("Failed to delete file: %v", err))
				}
				return m, func() tea.Msg {
					return errors.New(errors.ErrorTypeFileSystem, fmt.Sprintf("Failed to delete file: %v", err))
				}
			}
			stats.DeletedFiles = 1
			stats.DeletedSize = item.Size

			if m.Logger != nil {
				m.Logger.Log(logging.DEBUG, fmt.Sprintf("Deleted: %s (size: %d)", item.Path, item.Size))
			}
		}
	} else {
		// Batch deletion mode - process all selected files
		items := m.List.Items()
		if len(items) == 0 {
			if m.Logger != nil {
				m.Logger.Log(logging.DEBUG, "No files to delete")
			}
			return m, nil
		}

		// Count selected files
		selectedCount := 0
		for _, item := range items {
			cleanItem := item.(models.CleanItem)
			if cleanItem.Size != -1 {
				selectedCount++
			}
		}

		if selectedCount == 0 {
			if m.Logger != nil {
				m.Logger.Log(logging.DEBUG, "No files selected for deletion")
			}
			return m, nil
		}

		stats.TotalFiles = int64(selectedCount)

		for _, item := range items {
			cleanItem := item.(models.CleanItem)
			// Skip parent directory entry and unselected files
			if cleanItem.Size == -1 {
				continue
			}

			stats.TotalSize += cleanItem.Size

			if m.OptionState[options.SendFilesToTrash] {
				// Move to trash
				m.Filemanager.MoveFileToTrash(cleanItem.Path)
				stats.TrashedFiles++
				stats.TrashedSize += cleanItem.Size

				if m.Logger != nil {
					m.Logger.Log(logging.DEBUG, fmt.Sprintf("Moved to trash: %s (size: %d)", cleanItem.Path, cleanItem.Size))
				}
			} else {
				// Permanent deletion
				if err := os.Remove(cleanItem.Path); err != nil {
					if m.Logger != nil {
						m.Logger.Log(logging.ERROR, fmt.Sprintf("Failed to delete file: %v", err))
					}
					return m, func() tea.Msg {
						return errors.New(errors.ErrorTypeFileSystem, fmt.Sprintf("Failed to delete file: %v", err))
					}
				}
				stats.DeletedFiles++
				stats.DeletedSize += cleanItem.Size

				if m.Logger != nil {
					m.Logger.Log(logging.DEBUG, fmt.Sprintf("Deleted: %s (size: %d)", cleanItem.Path, cleanItem.Size))
				}
			}
		}
	}

	// Update end time
	stats.EndTime = time.Now()

	// Log final statistics
	if m.Logger != nil {
		m.Logger.Log(logging.INFO, fmt.Sprintf("Delete operation completed. Statistics: %+v", stats))
		m.Logger.UpdateStats(stats)
	}

	// Update all LogTabs
	if m.TabManager != nil {
		for _, tab := range m.TabManager.GetAllTabs() {
			if logTab, ok := tab.(*clean.LogTab); ok {
				logTab.UpdateStats(stats)
			}
		}
	}

	if m.OptionState[options.ExitAfterDeletion] {
		return m, tea.Quit
	}

	return m, m.LoadFiles()
}

func (m *CleanFilesModel) DeleteUserSelectedFiles(stats *logging.ScanStatistics) (tea.Model, tea.Cmd) {

	if len(m.SelectedFiles) > 0 {
		stats.TotalFiles = int64(m.SelectedCount)
		stats.TotalSize = m.SelectedSize

		if m.OptionState[options.SendFilesToTrash] {
			for filePath := range m.SelectedFiles {
				m.Filemanager.MoveFileToTrash(filePath)
			}
			stats.TrashedFiles = int64(m.SelectedCount)
			stats.TrashedSize = m.SelectedSize
		} else {
			for filePath := range m.SelectedFiles {
				os.Remove(filePath)
			}
			stats.DeletedFiles = int64(m.SelectedCount)
			stats.DeletedSize = m.SelectedSize
		}

		stats.EndTime = time.Now()

		// Log final statistics
		if m.Logger != nil {
			m.Logger.Log(logging.INFO, fmt.Sprintf("Delete user selected files completed. Statistics: %+v", stats))
			m.Logger.UpdateStats(stats)
		}

		// Clear selections after deletion
		m.SelectedFiles = make(map[string]bool)
		m.SelectedSize = 0
		m.SelectedCount = 0
		m.LastSelectedIndex = -1
	}

	return m, m.LoadFiles()
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
	case "shift+up", "shift+down":
		if !m.ShowDirs {
			// Get current item before update
			selectedItem := m.List.SelectedItem()
			if selectedItem != nil {
				item := selectedItem.(models.CleanItem)
				if item.Size != -1 {
					m.SelectedFiles[item.Path] = true
					m.SelectedSize += item.Size
					m.SelectedCount++
				}
			}

			// Create new message with just the direction
			var directionMsg tea.KeyMsg
			if msg.String() == "shift+up" {
				directionMsg = tea.KeyMsg{Type: tea.KeyUp}
			} else {
				directionMsg = tea.KeyMsg{Type: tea.KeyDown}
			}

			// Update list position with just the direction
			var cmd tea.Cmd
			m.List, cmd = m.List.Update(directionMsg)

			return m, cmd
		}
		return m, nil
	case "ctrl+a":
		if !m.ShowDirs {
			items := m.List.Items()

			// Check if any files are selected
			hasSelected := false
			for _, item := range items {
				cleanItem := item.(models.CleanItem)
				if cleanItem.Size != -1 {
					hasSelected = true
					break
				}
			}

			// If any files are selected, deselect all
			if hasSelected {
				m.SelectedFiles = make(map[string]bool)
				m.SelectedSize = 0
				m.SelectedCount = 0
			} else {
				// Otherwise select all files
				m.SelectedFiles = make(map[string]bool)
				m.SelectedSize = 0
				m.SelectedCount = 0
				for _, item := range items {
					cleanItem := item.(models.CleanItem)
					if cleanItem.Size != -1 { // Skip parent directory entry
						m.SelectedFiles[cleanItem.Path] = true
						m.SelectedSize += cleanItem.Size
						m.SelectedCount++
					}
				}
			}
			return m, nil
		}
		return m, nil
	case "right":
		if !strings.HasSuffix(m.FocusedElement, "Input") {
			return m.handleArrowRight()
		} else {
			m.UpdateInputs(msg)
		}
		return m, nil
	case "left":
		if !strings.HasSuffix(m.FocusedElement, "Input") {
			return m.handleArrowLeft()
		} else {
			m.UpdateInputs(msg)
		}
		return m, nil
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
	case "ctrl+s": //toogle dir mode or files mode
		if m.ShowDirs {
			m.ShowDirs = false
			return m, m.LoadFiles()
		} else {
			m.ShowDirs = true
			return m, m.LoadDirs()
		}
	case "ctrl+d":
		return m.OnDelete()
	case "ctrl+o":
		return m, m.OpenFileExplorer(m.CurrentPath)
	case "alt+c":
		return m.handleAltC()
	case "alt+1": // Toggle hidden files
		m.OptionState[options.ShowHiddenFiles] = !m.OptionState[options.ShowHiddenFiles]
		return m, m.LoadFiles()
	case "alt+2": // Toggle confirm deletion
		m.OptionState[options.ConfirmDeletion] = !m.OptionState[options.ConfirmDeletion]
		return m, nil
	case "alt+3": // Toggle include subfolders
		m.OptionState[options.IncludeSubfolders] = !m.OptionState[options.IncludeSubfolders]
		return m, nil
	case "alt+4": // Toggle delete empty subfolders
		m.OptionState[options.DeleteEmptySubfolders] = !m.OptionState[options.DeleteEmptySubfolders]
		return m, nil
	case "alt+5": // Toggle send files to trash
		m.OptionState[options.SendFilesToTrash] = !m.OptionState[options.SendFilesToTrash]
		return m, nil
	case "alt+6": // Toggle log operations
		m.OptionState[options.LogOperations] = !m.OptionState[options.LogOperations]
		return m, nil
	case "alt+7": // Toggle log to file
		m.OptionState[options.LogToFile] = !m.OptionState[options.LogToFile]
		return m, nil
	case "alt+8": // Toggle show statistics
		m.OptionState[options.ShowStatistics] = !m.OptionState[options.ShowStatistics]
		return m, nil
	case "alt+9": // Toggle exit after deletion
		m.OptionState[options.ExitAfterDeletion] = !m.OptionState[options.ExitAfterDeletion]
		return m, nil
	case "enter":
		return m.handleEnter()
	case " ":
		if !m.ShowDirs && m.FocusedElement == "list" {
			selectedItem := m.List.SelectedItem()
			if selectedItem == nil {
				return m, nil
			}

			cleanItem, ok := selectedItem.(models.CleanItem)
			if !ok {
				return m, nil
			}

			// Skip parent directory entry
			if cleanItem.Size == -1 {
				return m, nil
			}

			// Toggle selection
			if m.SelectedFiles == nil {
				m.SelectedFiles = make(map[string]bool)
			}

			if m.SelectedFiles[cleanItem.Path] {
				delete(m.SelectedFiles, cleanItem.Path)
				m.SelectedSize -= cleanItem.Size
				m.SelectedCount--
			} else {
				m.SelectedFiles[cleanItem.Path] = true
				m.SelectedSize += cleanItem.Size
				m.SelectedCount++
			}

			m.LastSelectedIndex = m.List.Index()
			return m, nil
		}
		return m.handleSpace()
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
	case "alt+up", "alt+down":
		if !m.ShowDirs {
			// Get current item before update
			selectedItem := m.List.SelectedItem()
			if selectedItem != nil {
				item := selectedItem.(models.CleanItem)
				if item.Size != -1 && m.SelectedFiles[item.Path] {
					delete(m.SelectedFiles, item.Path)
					m.SelectedSize -= item.Size
					m.SelectedCount--
				}
			}

			// Create new message with just the direction
			var directionMsg tea.KeyMsg
			if msg.String() == "alt+up" {
				directionMsg = tea.KeyMsg{Type: tea.KeyUp}
			} else {
				directionMsg = tea.KeyMsg{Type: tea.KeyDown}
			}

			// Update list position with just the direction
			var cmd tea.Cmd
			m.List, cmd = m.List.Update(directionMsg)

			return m, cmd
		}
		return m, nil
	default:
		m.UpdateInputs(msg)

		//If you put the space handling above, then you will not be able to write a space in input.
		if msg.String() == " " {
			return m.handleSpace()
		}

		return m, nil
	}
}

func (m *CleanFilesModel) UpdateInputs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	default:
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
		m.FocusedElement = options.GetNextOption(m.FocusedElement, "clean_option_", len(options.DefaultCleanOption), true)
	}

	return m, nil
}

func (m *CleanFilesModel) handleSpace() (tea.Model, tea.Cmd) {
	// Handle space key for options
	if strings.HasPrefix(m.FocusedElement, "clean_option_") {
		optionNum := strings.TrimPrefix(m.FocusedElement, "clean_option_")
		idx, err := strconv.Atoi(optionNum)
		if err != nil {
			return m, nil
		}
		if idx < 1 || idx > len(options.DefaultCleanOption) {
			return m, nil
		}
		idx-- // Convert to 0-based index

		// Get the option name and toggle its state
		optName := options.DefaultCleanOption[idx]
		m.OptionState[optName] = !m.OptionState[optName]

		// Debug log when option is toggled
		if m.Logger != nil {
			m.Logger.Log(logging.DEBUG, fmt.Sprintf("Option '%s' toggled to: %v", optName, m.OptionState[optName]))
		}

		// Keep focus on the current option
		m.FocusedElement = "clean_option_" + optionNum

		// If this is the options.ShowHiddenFiles option, reload files
		if optName == options.ShowHiddenFiles {
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
		m.FocusedElement = options.GetNextOption(m.FocusedElement, "clean_option_", len(options.DefaultCleanOption), false)
	}

	return m, nil
}

func (m *CleanFilesModel) handleArrowRight() (tea.Model, tea.Cmd) {
	tabLength := len(m.TabManager.GetAllTabs())
	activeTabIndex := m.TabManager.GetActiveTabIndex()

	if tabLength-1 == activeTabIndex {
		m.TabManager.SetActiveTabIndex(0)
	} else {
		m.TabManager.SetActiveTabIndex(activeTabIndex + 1)
	}

	return m, nil
}

func (m *CleanFilesModel) handleArrowLeft() (tea.Model, tea.Cmd) {
	tabLength := len(m.TabManager.GetAllTabs())
	activeTabIndex := m.TabManager.GetActiveTabIndex()

	if activeTabIndex == 0 {
		m.TabManager.SetActiveTabIndex(tabLength - 1)
	} else {
		m.TabManager.SetActiveTabIndex(activeTabIndex - 1)
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
	m.FocusedElement = "clean_option_1"
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
			return m.HandlePressStartButton()
		}
	} else {
		switch m.FocusedElement {
		case "pathInput":
			path := m.PathInput.Value()
			if path != "" {
				expandedPath := utils.ExpandTilde(path)
				if _, err := os.Stat(expandedPath); err == nil {
					m.CurrentPath = expandedPath
					m.IsLaunched = true
					m.Error = nil
					return m, tea.Batch(m.LoadFiles(), m.CalculateDirSizeAsync())
				} else {
					return m, func() tea.Msg {
						return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid path: %s", path))
					}
				}
			}
		case "extInput", "minSizeInput", "maxSizeInput", "excludeInput", "olderInput", "newerInput":
			// Validate input values before updating
			var err error
			switch m.FocusedElement {
			case "minSizeInput":
				if m.MinSizeInput.Value() != "" {
					err = m.Validator.ValidateSize(m.MinSizeInput.Value())
				}
			case "maxSizeInput":
				if m.MaxSizeInput.Value() != "" {
					err = m.Validator.ValidateSize(m.MaxSizeInput.Value())
				}
			case "olderInput":
				if m.OlderInput.Value() != "" {
					err = m.Validator.ValidateTimeDuration(m.OlderInput.Value())
				}
			case "newerInput":
				if m.NewerInput.Value() != "" {
					err = m.Validator.ValidateTimeDuration(m.NewerInput.Value())
				}

			}

			if err != nil {
				return m, func() tea.Msg {
					return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid %s value: %v", m.FocusedElement, err))
				}
			}

			// Clear error if validation passed
			m.Error = nil
			// Update the list of files when pressing Enter in the input fields
			return m, m.LoadFiles()
		case "dirButton":
			return m.HandlePressDirButton()
		case "deleteButton":
			return m.OnDelete()
		case "list":
			var cmds []tea.Cmd
			if !m.ShowDirs && m.List.SelectedItem() != nil {
				selectedItem := m.List.SelectedItem().(models.CleanItem)
				if selectedItem.Size == -1 {
					m.CurrentPath = selectedItem.Path
					m.PathInput.SetValue(selectedItem.Path)
					cmds = append(cmds, m.LoadFiles(), m.CalculateDirSizeAsync())
					return m, tea.Batch(cmds...)
				}
				info, err := os.Stat(selectedItem.Path)
				if err == nil && info.IsDir() {
					m.CurrentPath = selectedItem.Path
					m.PathInput.SetValue(selectedItem.Path)
					cmds = append(cmds, m.LoadFiles(), m.CalculateDirSizeAsync())
					return m, tea.Batch(cmds...)
				}
			} else if m.ShowDirs && m.DirList.SelectedItem() != nil {
				selectedDir := m.DirList.SelectedItem().(models.CleanItem)
				m.CurrentPath = selectedDir.Path
				m.PathInput.SetValue(selectedDir.Path)
				m.ShowDirs = false
				cmds = append(cmds, m.LoadFiles(), m.CalculateDirSizeAsync())
				return m, tea.Batch(cmds...)
			}
		default:
			if strings.HasPrefix(m.FocusedElement, "clean_option_") {
				optionNum := strings.TrimPrefix(m.FocusedElement, "clean_option_")
				idx, err := strconv.Atoi(optionNum)
				if err != nil {
					return m, nil
				}
				if idx < 1 || idx > len(options.DefaultCleanOption) {
					return m, nil
				}
				idx--

				optName := options.DefaultCleanOption[idx]
				m.OptionState[optName] = !m.OptionState[optName]

				if m.Logger != nil {
					m.Logger.Log(logging.DEBUG, fmt.Sprintf("Option '%s' toggled to: %v", optName, m.OptionState[optName]))
				}

				m.FocusedElement = "clean_option_" + optionNum

				if optName == options.ShowHiddenFiles {
					return m, m.LoadFiles()
				}
				return m, nil
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *CleanFilesModel) HandlePressDirButton() (tea.Model, tea.Cmd) {
	if m.ShowDirs {
		m.ShowDirs = false
		return m, m.LoadFiles()
	} else {
		m.ShowDirs = true
		return m, m.LoadDirs()
	}
}

func (m *CleanFilesModel) HandlePressStartButton() (tea.Model, tea.Cmd) {
	path := m.PathInput.Value()
	if path != "" {
		expandedPath := utils.ExpandTilde(path)
		if _, err := os.Stat(expandedPath); err == nil {
			m.CurrentPath = expandedPath
			m.FocusedElement = "pathInput"
			m.PathInput.Focus()
			m.IsLaunched = true
			m.Error = nil // Clear error when valid path is entered
			return m, tea.Batch(m.LoadFiles(), m.CalculateDirSizeAsync())
		} else {
			return m, func() tea.Msg {
				return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid path: %s", path))
			}
		}
	}

	return m, nil
}

func (m *CleanFilesModel) blurAllInputs() {
	m.PathInput.Blur()
	m.ExtInput.Blur()
	m.MinSizeInput.Blur()
	m.MaxSizeInput.Blur()
	m.ExcludeInput.Blur()
	m.OlderInput.Blur()
	m.NewerInput.Blur()
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

func (m *CleanFilesModel) GetSelectedFiles() map[string]bool {
	return m.SelectedFiles
}

func (m *CleanFilesModel) GetSelectedCount() int {
	return m.SelectedCount
}

func (m *CleanFilesModel) GetSelectedSize() int64 {
	return m.SelectedSize
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

func (m *CleanFilesModel) SetMaxSize(size int64) {
	m.MaxSize = size
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
