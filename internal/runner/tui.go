package runner

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui"
	"github.com/pashkov256/deletor/internal/validation"
)

func RunTUI(
	filemanager filemanager.FileManager,
	rules rules.Rules, validator *validation.Validator,
) error {
	app := tui.NewApp(filemanager, rules, validator)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
