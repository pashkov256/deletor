package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type page int

const (
	menuPage page = iota
	cleanPage
	rulesPage
	statsPage
)

type App struct {
	menu       *MainMenu
	cleanFiles *model
	rules      *RulesModel
	page       page
	err        error
	startDir   string
	exclude    []string
	extensions []string
	minSize    int64
}

func NewApp(startDir string, extensions []string, exclude []string, minSize int64) *App {
	return &App{
		menu:       NewMainMenu(),
		rules:      NewRulesModel(),
		page:       menuPage,
		startDir:   startDir,
		extensions: extensions,
		exclude:    exclude,
		minSize:    minSize,
	}
}

func (a *App) Init() tea.Cmd {
	a.cleanFiles = initialModel()
	return tea.Batch(a.menu.Init(), a.cleanFiles.Init(), a.rules.Init())
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "esc":
			if a.page != menuPage {
				a.page = menuPage
				return a, nil
			}
		case "enter":
			if a.page == menuPage {
				switch a.menu.list.SelectedItem().(item).Title() {
				case "🧹 Clean Files":
					a.page = cleanPage
					cmds = append(cmds, a.cleanFiles.loadFiles())
				case "⚙️ Manage Rules":
					a.page = rulesPage
				case "📊 Statistics":
					a.page = statsPage
				case "🚪 Exit":
					return a, tea.Quit
				}
				return a, tea.Batch(cmds...)
			}
		}
	}

	switch a.page {
	case menuPage:
		menuModel, menuCmd := a.menu.Update(msg)
		menu := menuModel.(*MainMenu)
		a.menu = menu
		cmd = menuCmd
	case cleanPage:
		cleanModel, cleanCmd := a.cleanFiles.Update(msg)
		if m, ok := cleanModel.(*model); ok {
			a.cleanFiles = m
		}
		cmd = cleanCmd
	case rulesPage:
		rulesModel, rulesCmd := a.rules.Update(msg)
		if r, ok := rulesModel.(*RulesModel); ok {
			a.rules = r
		}
		cmd = rulesCmd
	}

	return a, tea.Batch(cmd, tea.Batch(cmds...))
}

func (a *App) View() string {
	var content string
	switch a.page {
	case menuPage:
		content = a.menu.View()
	case cleanPage:
		content = a.cleanFiles.View()
	case rulesPage:
		content = a.rules.View()
	case statsPage:
		content = "Statistics page coming soon..."
	}
	return AppStyle.Render(content)
}
