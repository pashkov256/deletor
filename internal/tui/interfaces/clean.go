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
	GetList() list.Model
	GetDirList() list.Model
	GetFileToDelete() *models.CleanItem
	GetShowDirs() bool
	GetCalculatingSize() bool
	GetDirSize() int64
	GetFilteredSize() int64
	GetFilteredCount() int
	GetOptionState() map[string]bool
	GetMinSize() int64
	GetExclude() []string
	GetRules() rules.Rules
	GetFilemanager() filemanager.FileManager
	GetFocusedElement() string
	GetPathInput() textinput.Model
	GetExtInput() textinput.Model
	GetMinSizeInput() textinput.Model
	GetMaxSizeInput() textinput.Model
	GetExcludeInput() textinput.Model
	GetNewerInput() textinput.Model
	GetOlderInput() textinput.Model

	// Setters and state updates
	SetFocusedElement(element string)
	SetShowDirs(show bool)
	SetOptionState(option string, state bool)
	SetMinSize(size int64)
	SetExclude(exclude []string)
	SetExtensions(extensions []string)
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	CalculateDirSizeAsync() tea.Cmd
	LoadFiles() tea.Cmd
	LoadDirs() tea.Cmd
}
