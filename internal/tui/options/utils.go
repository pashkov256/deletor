package options

func GetEmojiByCleanOption(optionName string) string {
	emoji := ""

	switch optionName {
	case ShowHiddenFiles:
		emoji = "👁️"
	case ConfirmDeletion:
		emoji = "⚠️"
	case IncludeSubfolders:
		emoji = "📁"
	case DeleteEmptySubfolders:
		emoji = "🗑️"
	case SendFilesToTrash:
		emoji = "♻️"
	case LogOperations:
		emoji = "📝"
	case LogToFile:
		emoji = "📄"
	case ShowStatistics:
		emoji = "📊"
	case ExitAfterDeletion:
		emoji = "🚪"
	}

	return emoji
}
