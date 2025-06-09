package runner

import (
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui"
	"github.com/pashkov256/deletor/internal/validation"
)

func RunTUI(
	filemanager filemanager.FileManager,
	rules rules.Rules, validator *validation.Validator,
) error {
	zone.NewGlobal()
	app := tui.NewApp(filemanager, rules, validator)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
