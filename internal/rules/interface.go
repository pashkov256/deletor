package rules

// Rules defines the interface for managing file operation rules
type Rules interface {
	UpdateRules(options ...RuleOption) error // Updates rules with provided options
	GetRules() (*defaultRules, error)        // Returns current file rules configuration
	SetupRulesConfig() error                 // Initializes rules configuration
	GetRulesPath() string                    // Returns path to rules configuration file
}

// defaultRules holds the configuration for file operations
type defaultRules struct {
	Path                  string   `json:",omitempty"` // Target directory path
	Extensions            []string `json:",omitempty"` // File extensions to process
	Exclude               []string `json:",omitempty"` // Patterns to exclude
	MinSize               string   `json:",omitempty"` // Minimum file size
	MaxSize               string   `json:",omitempty"` // Maximum file size
	OlderThan             string   `json:",omitempty"` // Only process files older than
	NewerThan             string   `json:",omitempty"` // Only process files newer than
	ShowHiddenFiles       bool     `json:",omitempty"` // Whether to show hidden files
	ConfirmDeletion       bool     `json:",omitempty"` // Whether to confirm deletions
	IncludeSubfolders     bool     `json:",omitempty"` // Whether to process subfolders
	DeleteEmptySubfolders bool     `json:",omitempty"` // Whether to remove empty folders
	SendFilesToTrash      bool     `json:",omitempty"` // Whether to use trash instead of delete
	LogOperations         bool     `json:",omitempty"` // Whether to log operations
	LogToFile             bool     `json:",omitempty"` // Whether to write logs to file
	ShowStatistics        bool     `json:",omitempty"` // Whether to display statistics
	DisableEmoji          bool     `json:",omitempty"` // Whether to disable emoji
	ExitAfterDeletion     bool     `json:",omitempty"` // Whether to exit after deletion
}

// NewRules creates a new instance of the default rules
func NewRules() Rules {
	return &defaultRules{}
}
