package options

import (
	"fmt"
	"strconv"
	"strings"
)

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

// GetNextOption returns the next or previous option in a circular manner
func GetNextOption(currentOption string, maxOptions int, forward bool) string {
	currentNum := 1
	if strings.HasPrefix(currentOption, "option") {
		numStr := strings.TrimPrefix(currentOption, "option")
		if num, err := strconv.Atoi(numStr); err == nil {
			currentNum = num
		}
	}

	var nextNum int
	if forward {
		nextNum = currentNum + 1
		if nextNum > maxOptions {
			nextNum = 1
		}
	} else {
		nextNum = currentNum - 1
		if nextNum < 1 {
			nextNum = maxOptions
		}
	}

	return fmt.Sprintf("option%d", nextNum)
}
