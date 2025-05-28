package views

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pashkov256/deletor/internal/cache"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/utils"
)

type CacheModel struct {
	OptionState    map[string]bool
	FocusedElement string
	cacheManager   cache.Manager
	filemanager    filemanager.FileManager
	scanResults    []cache.ScanResult
	isScanning     bool
	status         string
}

type CachePath struct {
	Path string
	Size string
}

func InitialCacheModel(fm filemanager.FileManager) *CacheModel {
	return &CacheModel{
		cacheManager:   *cache.NewCacheManager(fm),
		filemanager:    fm,
		OptionState:    options.DefaultCacheOptionState,
		FocusedElement: "option1",
		status:         "",
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
		switch name {
		case options.SystemCache:
			emoji = "💻"
		}

		content.WriteString(style.Render(fmt.Sprintf("[%s] %s %-20s", map[bool]string{true: "✓", false: "○"}[m.OptionState[name]], emoji, name)))
		content.WriteString("\n")
	}

	if len(m.scanResults) > 0 {
		content.WriteString("\n\n")

		pathStyle := styles.ScanResultPathStyle.Copy().Width(pathWidth).Align(lipgloss.Left)
		sizeStyle := styles.ScanResultSizeStyle.Copy().Width(sizeWidth).Align(lipgloss.Right)
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
		content.WriteString(styles.InfoStyle.Render("🔍 Scanning..."))
	} else {
		content.WriteString("\n")
		content.WriteString(styles.ScanResultEmptyStyle.Render("Press 'Scan now' to see cache locations \n"))
	}

	// Show status message if exists
	if m.status != "" {
		content.WriteString("\n")
		content.WriteString(styles.InfoStyle.Render(m.status))
	}

	content.WriteString("\n")
	scanBtn := styles.LaunchButtonStyle.Render("🔍 Scan now")
	deleteBtn := styles.DeleteButtonStyle.Render("🗑️ Delete selected")

	if m.FocusedElement == "scanButton" {
		scanBtn = styles.LaunchButtonFocusedStyle.Render("🔍 Scan now")
	} else if m.FocusedElement == "deleteButton" {
		deleteBtn = styles.DeleteButtonFocusedStyle.Render("🗑️ Delete selected")
	}

	content.WriteString(scanBtn)
	content.WriteString("  ")
	content.WriteString(deleteBtn)
	content.WriteString("\n")

	return content.String()
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
		m.status = "" // Clear status when scanning

		results := m.cacheManager.ScanAllLocations()
		m.scanResults = results
		m.isScanning = false

		return m, nil
	} else if m.FocusedElement == "deleteButton" {
		if runtime.GOOS != "windows" {
			m.status = "Currently only Windows is supported for cache clearing\n"
		} else {
			m.cacheManager.ClearCache()
			m.scanResults = []cache.ScanResult{}
			m.status = "Cache clearing completed\n"
		}
		return m, nil
	}
	return m, nil
}
