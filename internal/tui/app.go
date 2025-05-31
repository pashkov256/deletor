package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"

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
}

func NewApp(
	filemanager filemanager.FileManager,
	rules rules.Rules,
) *App {
	return &App{
		menu:        views.NewMainMenu(),
		rulesModel:  views.NewRulesModel(rules),
		page:        menuPage,
		filemanager: filemanager,
		rules:       rules,
	}
}

func (a *App) Init() tea.Cmd {
	a.cleanFilesModel = views.InitialCleanModel(a.rules, a.filemanager)
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
					a.cleanFilesModel = views.InitialCleanModel(a.rules, a.filemanager)
					cmds = append(cmds, a.cleanFilesModel.Init())
				}
				a.page = menuPage
				return a, tea.Batch(cmds...)
			}
		case "enter":
			if a.page == menuPage {
				switch a.menu.List.SelectedItem().(menu.Item).Title() {
				case menu.CleanFIlesTitle:
					a.page = cleanPage
					cmds = append(cmds, a.cleanFilesModel.LoadFiles())
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
	case statsPage:
		content = "Statistics page coming soon..."
	}

	return styles.AppStyle.Render(content)
}
