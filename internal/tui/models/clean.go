package models

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/views/tabs"
)

type CleanItem struct {
	path string
	size int64
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
