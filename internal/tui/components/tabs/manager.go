package tabs

// TabManager - base tab manager
type TabManager[m any] struct {
	tabs      []Tab
	activeTab int
	model     *m
}

func (t *TabManager[m]) GetActiveTab() Tab {
	return t.tabs[t.activeTab]
}

func (t *TabManager[m]) GetActiveTabIndex() int {
	return t.activeTab
}

func (t *TabManager[m]) SetActiveTabIndex(index int) {
	t.activeTab = index
}

func NewTabManager[m any](tabs []Tab, model *m) *TabManager[m] {
	return &TabManager[m]{
		model:     model,
		activeTab: 0,
		tabs:      tabs,
	}
}
