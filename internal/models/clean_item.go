package models

type CleanItem struct {
	Path  string
	Size  int64
	IsDir bool
}

// For list.Item bubble tea
func (i CleanItem) Title() string {
	return ""
}
func (i CleanItem) Description() string { return i.Path }
func (i CleanItem) FilterValue() string { return i.Path }
