package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tab - interface for all tabs
type Tab interface {
	View() string
	Update(msg tea.Msg) tea.Cmd
	Init() tea.Cmd
}

// TabStyles - base tab styles
type TabStyles struct {
	TabStyle       lipgloss.Style
	ActiveTabStyle lipgloss.Style
}

// TabFactory - interface for creating tabs
type TabFactory interface {
	NewMainTab(model interface{}) Tab
	NewFiltersTab(model interface{}) Tab
	NewOptionsTab(model interface{}) Tab
	NewHelpTab(model interface{}) Tab
}
