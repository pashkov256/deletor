package menu

import "github.com/charmbracelet/bubbles/list"

type Item struct {
	title string
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return "" }
func (i Item) FilterValue() string { return i.title }

var MenuItems = []list.Item{
	Item{title: CleanFIlesTitle},
	Item{title: CleanCacheTitle},
	Item{title: ManageRulesTitle},
	Item{title: StatisticsTitle},
	Item{title: ExitTitle},
}
