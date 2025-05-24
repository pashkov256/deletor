package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type CacheModel struct {
}

func NewCacheModel() *CacheModel {
	return &CacheModel{}
}

func (m *CacheModel) Init() tea.Cmd {

	// Otherwise just return the blink command for the path input
	return textinput.Blink
}

func (m *CacheModel) View() string {

	// --- Content rendering ---
	var content strings.Builder
	content.WriteString("sdfsdf")
	return content.String()
}

func (m *CacheModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	return m, nil
}
