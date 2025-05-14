package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/tabs"
	"github.com/pashkov256/deletor/internal/utils"
)

type CleanItem struct {
	Path string
	Size int64
}

// Message for directory size updates
type DirSizeMsg struct {
	size int64
}

// Message for filtered files size updates
type FilteredSizeMsg struct {
	size  int64
	count int
}

func (i CleanItem) Title() string {
	if i.Size == -1 {
		return "📂 .." // Parent directory
	}

	if i.Size == 0 {
		return "📁 " + filepath.Base(i.Path) // Directory
	}

	// Regular file
	filename := filepath.Base(i.Path)
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
	sizeStr := utils.FormatSize(i.Size)

	// Calculate padding for alignment
	padding := 50 - len(filename)
	if padding < 0 {
		padding = 0
	}

	return fmt.Sprintf("%s%s%s%s", icon, filename, strings.Repeat(" ", padding), sizeStr)
}

func (i CleanItem) Description() string { return i.Path }
func (i CleanItem) FilterValue() string { return i.Path }

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
	FileToDelete    *CleanItem
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

// Implement the interfaces.CleanModel interface
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

func (m *CleanFilesModel) GetActiveTab() int {
	return m.ActiveTab
}
