package rules

// RuleOption provides a function for setting an option
type RuleOption func(*defaultRules)

func WithPath(path string) RuleOption {
	return func(r *defaultRules) {
		r.Path = path
	}
}

func WithMinSize(size string) RuleOption {
	return func(r *defaultRules) {
		r.MinSize = size
	}
}

func WithMaxSize(size string) RuleOption {
	return func(r *defaultRules) {
		r.MaxSize = size
	}
}

func WithExtensions(extensions []string) RuleOption {
	return func(r *defaultRules) {
		r.Extensions = extensions
	}
}

func WithExclude(exclude []string) RuleOption {
	return func(r *defaultRules) {
		r.Exclude = exclude
	}
}

func WithOlderThan(time string) RuleOption {
	return func(r *defaultRules) {
		r.OlderThan = time
	}
}

func WithNewerThan(time string) RuleOption {
	return func(r *defaultRules) {
		r.NewerThan = time
	}
}

func WithOptions(showHidden, confirmDeletion, includeSubfolders, deleteEmptySubfolders, sendToTrash, logOps, logToFile, showStats, exitAfterDeletion bool) RuleOption {
	return func(r *defaultRules) {
		r.ShowHiddenFiles = showHidden
		r.ConfirmDeletion = confirmDeletion
		r.IncludeSubfolders = includeSubfolders
		r.DeleteEmptySubfolders = deleteEmptySubfolders
		r.SendFilesToTrash = sendToTrash
		r.LogOperations = logOps
		r.LogToFile = logToFile
		r.ShowStatistics = showStats
		r.ExitAfterDeletion = exitAfterDeletion
	}
}
