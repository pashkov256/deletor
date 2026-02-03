package views

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/cache"
	"github.com/pashkov256/deletor/internal/filemanager"
	rules "github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/errors"
	"github.com/pashkov256/deletor/internal/tui/help"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/utils"
)

type CacheModel struct {
	OptionState      map[string]bool
	FocusedElement   string
	cacheManager     cache.Manager
	filemanager      filemanager.FileManager
	scanResults      []cache.ScanResult
	isScanning       bool
	rulesOptionState map[string]bool
	status           string
	Error            *errors.Error
}

type CachePath struct {
	Path string
	Size string
}

func InitialCacheModel(fm filemanager.FileManager, rules rules.Rules) *CacheModel {
	latestRules, _ := rules.GetRules()
	return &CacheModel{
		cacheManager:   *cache.NewCacheManager(fm),
		filemanager:    fm,
		OptionState:    options.DefaultCacheOptionState,
		FocusedElement: "option1",
		rulesOptionState: map[string]bool{
			options.ShowHiddenFiles:       latestRules.ShowHiddenFiles,
			options.ConfirmDeletion:       latestRules.ConfirmDeletion,
			options.IncludeSubfolders:     latestRules.IncludeSubfolders,
			options.DeleteEmptySubfolders: latestRules.DeleteEmptySubfolders,
			options.SendFilesToTrash:      latestRules.SendFilesToTrash,
			options.LogOperations:         latestRules.LogOperations,
			options.LogToFile:             latestRules.LogToFile,
			options.ShowStatistics:        latestRules.ShowStatistics,
			options.DisableEmoji:          latestRules.DisableEmoji,
			options.ExitAfterDeletion:     latestRules.ExitAfterDeletion,
		},
		status: "",
	}
}

const pathWidth = 60
const sizeWidth = 15
const filesWidth = 10

func (m *CacheModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *CacheModel) View() string {
	var content strings.Builder
	disableEmoji := m.GetRulesOptionState()[options.DisableEmoji]
	content.WriteString("\n")
	content.WriteString("Select cache types to clear:\n")
	for optionIndex, name := range options.DefaultCacheOption {
		style := styles.OptionStyle
		if m.OptionState[name] {
			style = styles.SelectedOptionStyle
		}
		if m.FocusedElement == fmt.Sprintf("option%d", optionIndex+1) {
			style = styles.OptionFocusedStyle
		}
		content.WriteString(fmt.Sprintf("%-4s", fmt.Sprintf("%d.", optionIndex+1)))

		emoji := ""
		if !disableEmoji {
			switch name {
			case options.SystemCache:
				emoji = "💻"
			}
		}

		optionContent := fmt.Sprintf("[%s] %s %-20s", map[bool]string{true: "✓", false: "○"}[m.OptionState[name]], emoji, name)
		content.WriteString(zone.Mark(fmt.Sprintf("cache_option_%d", optionIndex+1), style.Render(optionContent)))
		content.WriteString("\n")
	}

	if len(m.scanResults) > 0 {
		content.WriteString("\n\n")

		// nolint:staticcheck
		pathStyle := styles.ScanResultPathStyle.Copy().Width(pathWidth).Align(lipgloss.Left)
		// nolint:staticcheck
		sizeStyle := styles.ScanResultSizeStyle.Copy().Width(sizeWidth).Align(lipgloss.Right)
		// nolint:staticcheck
		filesStyle := styles.ScanResultFilesStyle.Copy().Width(filesWidth).Align(lipgloss.Right)

		header := lipgloss.JoinHorizontal(lipgloss.Top,
			pathStyle.Render("Directory"),
			sizeStyle.Render("Size"),
			filesStyle.Render("Files"),
		)
		content.WriteString(styles.ScanResultHeaderStyle.Render(header))
		content.WriteString("\n")

		// Separator line
		separator := styles.ScanResultHeaderStyle.Render(strings.Repeat("─", pathWidth+sizeWidth+filesWidth))
		content.WriteString(separator)
		content.WriteString("\n")

		var totalSize int64
		var totalFiles int64

		// Results
		for _, result := range m.scanResults {
			pathCell := pathStyle.Render(result.Path)
			sizeCell := sizeStyle.Render(utils.FormatSize(result.Size))
			filesCell := filesStyle.Render(fmt.Sprintf("%d", result.FileCount))
			content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, pathCell, sizeCell, filesCell))
			content.WriteString("\n")
			totalSize += result.Size
			totalFiles += result.FileCount
		}

		content.WriteString(separator)
		content.WriteString("\n")

		totalLabel := pathStyle.Render("Total\n\n")
		totalSizeStr := sizeStyle.Render(utils.FormatSize(totalSize))
		totalFilesStr := filesStyle.Render(fmt.Sprintf("%d", totalFiles))
		content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, totalLabel, totalSizeStr, totalFilesStr))
	} else if m.isScanning {
		content.WriteString("\n")
		ScanningMsg := "🔍 Scanning..."
		if disableEmoji {
			newScanningMsg, err := utils.RemoveEmoji(ScanningMsg)
			if err == nil {
				ScanningMsg = newScanningMsg
			}
		}
		content.WriteString(styles.InfoStyle.Render(ScanningMsg))
	} else {
		content.WriteString("\n")
		content.WriteString(styles.ScanResultEmptyStyle.Render("Press 'Scan now' to see cache locations \n"))
	}

	// Show error or status message
	if m.Error != nil && m.Error.IsVisible() {
		errorStyle := errors.GetStyle(m.Error.GetType())
		content.WriteString("\n")
		content.WriteString(errorStyle.Render(m.Error.GetMessage()))
	} else if m.status != "" {
		content.WriteString("\n")
		content.WriteString(styles.SuccessStyle.Render(m.status))
	}

	content.WriteString("\n")

	scanMsg := "🔍 Scan now"
	deleteMsg := "🗑️ Delete selected"
	if disableEmoji { //remove emojis if disabled
		newScanMsg, err := utils.RemoveEmoji(scanMsg)
		if err == nil {
			scanMsg = newScanMsg
		}
		newDeleteMsg, err := utils.RemoveEmoji(deleteMsg)
		if err == nil {
			deleteMsg = newDeleteMsg
		}
	}
	scanBtn := styles.LaunchButtonStyle.Render(scanMsg)
	deleteBtn := styles.DeleteButtonStyle.Render(deleteMsg)

	switch m.FocusedElement {
	case "scanButton":
		scanBtn = styles.LaunchButtonFocusedStyle.Render(scanMsg)
	case "deleteButton":
		deleteBtn = styles.DeleteButtonFocusedStyle.Render(deleteMsg)
	}

	content.WriteString(zone.Mark("cache_scan_button", scanBtn))
	content.WriteString("  ")
	content.WriteString(zone.Mark("cache_delete_button", deleteBtn))
	content.WriteString("\n\n")
	content.WriteString("\n" + help.NavigateHelpText)
	return zone.Scan(content.String())
}

func (m *CacheModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			return m.handleTab()
		case "shift+tab":
			return m.handleShiftTab()
		case "enter", " ":
			return m.handleSpace()
		}
	case tea.MouseMsg:
		// nolint:staticcheck
		if msg.Type == tea.MouseLeft && msg.Action == tea.MouseActionPress {
			// Handle option clicks
			for i := range options.DefaultCacheOption {
				if zone.Get(fmt.Sprintf("cache_option_%d", i+1)).InBounds(msg) {
					m.FocusedElement = fmt.Sprintf("option%d", i+1)
					return m.handleSpace()
				}
			}

			// Handle scan button click
			if zone.Get("cache_scan_button").InBounds(msg) {
				m.FocusedElement = "scanButton"
				return m.handleSpace()
			}

			// Handle delete button click
			if zone.Get("cache_delete_button").InBounds(msg) {
				m.FocusedElement = "deleteButton"
				return m.handleSpace()
			}
		}
	}
	return m, nil
}

func (m *CacheModel) handleTab() (tea.Model, tea.Cmd) {
	switch m.FocusedElement {
	case "option1":
		m.FocusedElement = "scanButton"
	case "scanButton":
		m.FocusedElement = "deleteButton"
	case "deleteButton":
		m.FocusedElement = "option1"
	default:
		m.FocusedElement = "option1"
	}
	return m, nil
}

func (m *CacheModel) handleShiftTab() (tea.Model, tea.Cmd) {
	switch m.FocusedElement {
	case "option1":
		m.FocusedElement = "deleteButton"
	case "scanButton":
		m.FocusedElement = "option1"
	case "deleteButton":
		m.FocusedElement = "scanButton"
	default:
		m.FocusedElement = "option1"
	}
	return m, nil
}

func (m *CacheModel) handleSpace() (tea.Model, tea.Cmd) {
	if strings.HasPrefix(m.FocusedElement, "option") {
		optionNum := strings.TrimPrefix(m.FocusedElement, "option")
		idx, err := strconv.Atoi(optionNum)
		if err != nil {
			return m, nil
		}
		if idx < 1 || idx > len(options.DefaultCacheOption) {
			return m, nil
		}
		idx--

		optName := options.DefaultCacheOption[idx]
		m.OptionState[optName] = !m.OptionState[optName]

		m.FocusedElement = "option" + optionNum

		return m, nil
	} else if m.FocusedElement == "scanButton" {
		m.isScanning = true
		m.scanResults = nil
		m.status = ""
		m.Error = nil

		results := m.cacheManager.ScanAllLocations()
		m.scanResults = results
		m.isScanning = false

		return m, nil
	} else if m.FocusedElement == "deleteButton" {
		m.Error = nil
		m.status = ""

		if runtime.GOOS == "darwin" {
			m.Error = errors.New(errors.ErrorTypeFileSystem, "Currently only Windows and Linux is supported for cache clearing")
		} else {
			if err := m.cacheManager.ClearCache(); err != nil {
				m.Error = errors.New(errors.ErrorTypeFileSystem, "Not all files were successfully deleted")
			} else {
				m.scanResults = []cache.ScanResult{}
				m.status = "Cache clearing completed"
			}
		}
		return m, nil
	}
	return m, nil
}

func (m *CacheModel) GetRulesOptionState() map[string]bool {
	return m.rulesOptionState
}
