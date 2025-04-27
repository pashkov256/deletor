package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var mainAppStyle = lipgloss.NewStyle().
	Padding(1, 2, 1, 2)

type page int

const (
	menuPage page = iota
	cleanPage
	rulesPage
	statsPage
)

type App struct {
	menu       *MainMenu
	cleanFiles *CleanFilesModel
	page       page
	err        error
}

func NewApp(startDir string, extensions []string, minSize int64) *App {
	return &App{
		menu:       NewMainMenu(),
		cleanFiles: NewCleanFiles(startDir, extensions, minSize),
		page:       menuPage,
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.menu.Init(), a.cleanFiles.Init())
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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
				case "⚙️ Manage Rules":
					a.page = rulesPage
				case "📊 Statistics":
					a.page = statsPage
				}
				return a, nil
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
		clean := cleanModel.(*CleanFilesModel)
		a.cleanFiles = clean
		cmd = cleanCmd
	}

	return a, cmd
}

func (a *App) View() string {
	switch a.page {
	case menuPage:
		return a.menu.View()
	case cleanPage:
		return a.cleanFiles.View()
	case rulesPage:
		return "Rules page coming soon..."
	case statsPage:
		return "Statistics page coming soon..."
	default:
		return ""
	}
}
