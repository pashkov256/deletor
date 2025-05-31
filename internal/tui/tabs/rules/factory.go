package rules

import (
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/tabs/base"
)

type RulesTabFactory struct{}

func NewRulesTabFactory() *RulesTabFactory {
	return &RulesTabFactory{}
}

func (f *RulesTabFactory) CreateTabs(model interfaces.RulesModel) []base.Tab {
	return []base.Tab{
		&MainTab{model: model},
		&FiltersTab{model: model},
		&OptionsTab{model: model},
	}
}
