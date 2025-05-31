package interfaces

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type RulesModel interface {
	// Getters
	GetPathInput() textinput.Model
	GetExtInput() textinput.Model
	GetMinSizeInput() textinput.Model
	GetMaxSizeInput() textinput.Model
	GetExcludeInput() textinput.Model
	GetOlderInput() textinput.Model
	GetNewerInput() textinput.Model
	GetFocusedElement() string
	GetOptionState() map[string]bool

	// Setters
	SetFocusedElement(element string)
	SetOptionState(option string, state bool)

	// Other
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
}
