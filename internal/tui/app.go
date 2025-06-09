package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/validation"

	"github.com/pashkov256/deletor/internal/tui/menu"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/tui/views"
)

type page int

const (
	menuPage page = iota
	cleanPage
	cachePage
	rulesPage
	statsPage
)

type App struct {
	page            page
	menu            *views.MainMenu
	cleanFilesModel *views.CleanFilesModel
	rulesModel      *views.RulesModel
	cacheModel      *views.CacheModel
	filemanager     filemanager.FileManager
	rules           rules.Rules
	validator       *validation.Validator
}

func NewApp(
	filemanager filemanager.FileManager,
	rules rules.Rules,
	validator *validation.Validator,
) *App {
	return &App{
		menu:        views.NewMainMenu(),
		rulesModel:  views.NewRulesModel(rules, validator),
		page:        menuPage,
		filemanager: filemanager,
		rules:       rules,
		validator:   validator,
	}
}

func (a *App) Init() tea.Cmd {
	a.cleanFilesModel = views.InitialCleanModel(a.rules, a.filemanager, a.validator)
	a.cacheModel = views.InitialCacheModel(a.filemanager)
	return tea.Batch(a.menu.Init(), a.cleanFilesModel.Init(), a.rulesModel.Init())
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
				if a.page == rulesPage {
					a.cleanFilesModel = views.InitialCleanModel(a.rules, a.filemanager, a.validator)
					cmds = append(cmds, a.cleanFilesModel.Init())
				}
				a.page = menuPage
				return a, tea.Batch(cmds...)
			}
		case "enter":
			if a.page == menuPage {
				items := []string{
					menu.CleanFIlesTitle,
					menu.CleanCacheTitle,
					menu.ManageRulesTitle,
					menu.StatisticsTitle,
					menu.ExitTitle,
				}
				switch items[a.menu.SelectedIndex] {
				case menu.CleanFIlesTitle:
					a.cleanFilesModel = views.InitialCleanModel(a.rules, a.filemanager, a.validator)
					cmds = append(cmds, a.cleanFilesModel.Init(), a.cleanFilesModel.LoadFiles())
					a.page = cleanPage
				case menu.CleanCacheTitle:
					a.page = cachePage
				case menu.ManageRulesTitle:
					a.page = rulesPage
				case menu.StatisticsTitle:
					a.page = statsPage
				case menu.ExitTitle:
					return a, tea.Quit
				}
				return a, tea.Batch(cmds...)
			}
		}
	}

	switch a.page {
	case menuPage:
		menuModel, menuCmd := a.menu.Update(msg)
		menu := menuModel.(*views.MainMenu)
		a.menu = menu
		cmd = menuCmd
	case cleanPage:
		cleanModel, cleanCmd := a.cleanFilesModel.Update(msg)
		if m, ok := cleanModel.(*views.CleanFilesModel); ok {
			a.cleanFilesModel = m
		}
		cmd = cleanCmd
	case cachePage:
		cacheModel, cacheCmd := a.cacheModel.Update(msg)
		if m, ok := cacheModel.(*views.CacheModel); ok {
			a.cacheModel = m
		}
		cmd = cacheCmd
	case rulesPage:
		rulesModel, rulesCmd := a.rulesModel.Update(msg)
		if r, ok := rulesModel.(*views.RulesModel); ok {
			a.rulesModel = r
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
		content = a.cleanFilesModel.View()
	case cachePage:
		content = a.cacheModel.View()
	case rulesPage:
		content = a.rulesModel.View()
	}

	return styles.AppStyle.Render(content)
}
