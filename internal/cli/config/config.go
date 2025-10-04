package config

import (
	"time"

	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/utils"
)

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
	MoveFileToTrash    bool      // If true, files will be moved to trash instead of being permanently deleted
	UseRules           bool      // Whether to use rules from configuration file
	JsonLogsEnabled    bool      // Whether to generates JSON-formatted logs
	JsonLogsPath       string    // Path to append JSON-formatted logs
}

// LoadConfig initializes and returns a new Config instance with values from command-line flags
func LoadConfig() *Config {
	return GetFlags()
}

// GetConfig returns the current configuration instance
func (c *Config) GetConfig() *Config {
	return c
}

// GetWithRules returns a value from config if it's set, otherwise returns value from rules
func (c *Config) GetWithRules(rules rules.Rules) *Config {
	if c == nil {
		c = &Config{}
	}

	// Get all rules at once
	defaultRules, err := rules.GetRules()
	if err != nil {
		return c
	}

	// Get values from rules if not set in config
	if len(c.Extensions) == 0 {
		c.Extensions = defaultRules.Extensions
	}
	if c.Directory == "" && defaultRules.Path != "" {
		c.Directory = utils.ExpandTilde(defaultRules.Path)
	} else if c.Directory != "" && defaultRules.Path != "" {
		c.Directory = utils.ExpandTilde(c.Directory)
	}
	if c.MinSize == 0 && defaultRules.MinSize != "" {
		c.MinSize = utils.ToBytesOrDefault(defaultRules.MinSize)
	}
	if c.MaxSize == 0 && defaultRules.MaxSize != "" {
		c.MaxSize = utils.ToBytesOrDefault(defaultRules.MaxSize)
	}
	if len(c.Exclude) == 0 {
		c.Exclude = defaultRules.Exclude
	}
	if c.OlderThan.IsZero() && defaultRules.OlderThan != "" {
		c.OlderThan, _ = utils.ParseTimeDuration(defaultRules.OlderThan)
	}
	if c.NewerThan.IsZero() && defaultRules.NewerThan != "" {
		c.NewerThan, _ = utils.ParseTimeDuration(defaultRules.NewerThan)
	}
	if !c.IncludeSubdirs {
		c.IncludeSubdirs = defaultRules.IncludeSubfolders
	}
	if !c.DeleteEmptyFolders {

		c.DeleteEmptyFolders = defaultRules.DeleteEmptySubfolders
	}
	if !c.MoveFileToTrash {
		c.MoveFileToTrash = defaultRules.SendFilesToTrash
	}

	return c
}
