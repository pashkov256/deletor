package interfaces

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/models"
	"github.com/pashkov256/deletor/internal/rules"
)

// CleanModel defines the interface that models must implement to work with clean tabs
type CleanModel interface {
	// Getters
	GetCurrentPath() string
	GetExtensions() []string
	GetMinSize() int64
	GetExclude() []string
	GetOptions() []string
	GetOptionState() map[string]bool
	GetFocusedElement() string
	GetShowDirs() bool
	GetDirSize() int64
	GetCalculatingSize() bool
	GetFilteredSize() int64
	GetFilteredCount() int
	GetList() list.Model
	GetDirList() list.Model
	GetRules() rules.Rules
	GetFilemanager() filemanager.FileManager
	GetFileToDelete() *models.CleanItem
	GetPathInput() textinput.Model
	GetExtInput() textinput.Model
	GetMinSizeInput() textinput.Model
	GetMaxSizeInput() textinput.Model
	GetExcludeInput() textinput.Model
	GetOlderInput() textinput.Model
	GetNewerInput() textinput.Model
	GetSelectedFiles() map[string]bool
	GetSelectedCount() int
	GetSelectedSize() int64

	// Setters and state updates
	SetFocusedElement(element string)
	SetShowDirs(show bool)
	SetOptionState(option string, state bool)
	SetMinSize(size int64)
	SetMaxSize(size int64)
	SetExclude(exclude []string)
	SetExtensions(extensions []string)
	SetCurrentPath(path string)
	SetPathInput(input textinput.Model)
	SetExtInput(input textinput.Model)
	SetExcludeInput(input textinput.Model)
	SetSizeInput(input textinput.Model)
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	CalculateDirSizeAsync() tea.Cmd
	LoadFiles() tea.Cmd
	LoadDirs() tea.Cmd
}
