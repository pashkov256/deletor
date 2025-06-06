package runner

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui"
)

func RunTUI(
	filemanager filemanager.FileManager,
	rules rules.Rules,
) error {
	app := tui.NewApp(filemanager, rules)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
