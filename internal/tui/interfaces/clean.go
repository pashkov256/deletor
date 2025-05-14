package interfaces

// CleanModel defines the interface that models must implement to work with clean tabs
type CleanModel interface {
	GetCurrentPath() string
	GetExtensions() []string
	GetMinSize() int64
	GetExclude() []string
	GetOptions() []string
	GetOptionState() map[string]bool
	GetFocusedElement() string
	GetShowDirs() bool
	GetDirSize() int64
	GetCalculatingSize() bool
	GetFilteredSize() int64
	GetFilteredCount() int
	GetActiveTab() int
}
