package rules

// RuleOption is a function type that modifies rule settings
type RuleOption func(*defaultRules)

// WithPath sets the target directory path
func WithPath(path string) RuleOption {
	return func(r *defaultRules) {
		r.Path = path
	}
}

// WithMinSize sets the minimum file size filter
func WithMinSize(size string) RuleOption {
	return func(r *defaultRules) {
		r.MinSize = size
	}
}

// WithMaxSize sets the maximum file size filter
func WithMaxSize(size string) RuleOption {
	return func(r *defaultRules) {
		r.MaxSize = size
	}
}

// WithExtensions sets the file extensions to process
func WithExtensions(extensions []string) RuleOption {
	return func(r *defaultRules) {
		r.Extensions = extensions
	}
}

// WithExclude sets the patterns to exclude from processing
func WithExclude(exclude []string) RuleOption {
	return func(r *defaultRules) {
		r.Exclude = exclude
	}
}

// WithOlderThan sets the time filter for older files
func WithOlderThan(time string) RuleOption {
	return func(r *defaultRules) {
		r.OlderThan = time
	}
}

// WithNewerThan sets the time filter for newer files
func WithNewerThan(time string) RuleOption {
	return func(r *defaultRules) {
		r.NewerThan = time
	}
}

// WithOptions sets multiple boolean options at once
func WithOptions(showHidden, confirmDeletion, includeSubfolders, deleteEmptySubfolders, sendToTrash, logOps, logToFile, showStats, disableEmoji, exitAfterDeletion bool) RuleOption {
	return func(r *defaultRules) {
		r.ShowHiddenFiles = showHidden
		r.ConfirmDeletion = confirmDeletion
		r.IncludeSubfolders = includeSubfolders
		r.DeleteEmptySubfolders = deleteEmptySubfolders
		r.SendFilesToTrash = sendToTrash
		r.LogOperations = logOps
		r.LogToFile = logToFile
		r.ShowStatistics = showStats
		r.DisableEmoji = disableEmoji
		r.ExitAfterDeletion = exitAfterDeletion
	}
}
