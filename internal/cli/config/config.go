package config

import "time"

// Config holds all command-line configuration options for the application
type Config struct {
	Directory          string    // Target directory to process
	Extensions         []string  // File extensions to include
	MinSize            int64     // Minimum file size in bytes
	MaxSize            int64     // Maximum file size in bytes
	Exclude            []string  // Patterns to exclude
	IncludeSubdirs     bool      // Whether to process subdirectories
	ShowProgress       bool      // Whether to display progress
	IsCLIMode          bool      // Whether running in CLI mode
	HaveProgress       bool      // Whether progress tracking is available
	SkipConfirm        bool      // Whether to skip confirmation prompts
	DeleteEmptyFolders bool      // Whether to remove empty directories
	OlderThan          time.Time // Only process files older than this time
	NewerThan          time.Time // Only process files newer than this time
	MoveFileToTrash    bool      // // If true, files will be moved to trash instead of being permanently deleted
}

// LoadConfig initializes and returns a new Config instance with values from command-line flags
func LoadConfig() *Config {
	return GetFlags()
}

// GetConfig returns the current configuration instance
func (c *Config) GetConfig() *Config {
	return c
}
