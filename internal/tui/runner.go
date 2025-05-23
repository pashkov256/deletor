package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
)

func Start(
	filemanager filemanager.FileManager,
	rules rules.Rules,
) error {
	app := NewApp(filemanager, rules)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
