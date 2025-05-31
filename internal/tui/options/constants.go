package options

// Option names as constants to avoid string literals
const (
	//options for clean view
	ShowHiddenFiles       = "Show hidden files"
	ConfirmDeletion       = "Confirm deletion"
	IncludeSubfolders     = "Include subfolders"
	DeleteEmptySubfolders = "Delete empty subfolders"
	SendFilesToTrash      = "Send files to trash"
	LogOperations         = "Log operations"
	LogToFile             = "Log to file"
	ShowStatistics        = "Show statistics"
	ExitAfterDeletion     = "Exit after deletion"
	//options for cache view
	SystemCache = "System cache"
)

var DefaultCleanOptionState = map[string]bool{
	ShowHiddenFiles:       false,
	ConfirmDeletion:       false,
	IncludeSubfolders:     false,
	DeleteEmptySubfolders: false,
	SendFilesToTrash:      false,
	LogOperations:         false,
	LogToFile:             false,
	ShowStatistics:        true,
	ExitAfterDeletion:     false,
}

var DefaultCleanOption = []string{
	ShowHiddenFiles,
	ConfirmDeletion,
	IncludeSubfolders,
	DeleteEmptySubfolders,
	SendFilesToTrash,
	LogOperations,
	LogToFile,
	ShowStatistics,
	ExitAfterDeletion,
}

var DefaultCacheOptionState = map[string]bool{
	SystemCache: true,
}

var DefaultCacheOption = []string{
	SystemCache,
}
