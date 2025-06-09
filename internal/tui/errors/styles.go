package errors

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base error style
	baseErrorStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Margin(1, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF0000"))

	// nolint:staticcheck
	validationErrorStyle = baseErrorStyle.Copy().
				BorderForeground(lipgloss.Color("#FFA500")).
				Foreground(lipgloss.Color("#FFA500"))

	// nolint:staticcheck
	fileSystemErrorStyle = baseErrorStyle.Copy().
				BorderForeground(lipgloss.Color("#FF0000")).
				Foreground(lipgloss.Color("#FF0000"))

	// nolint:staticcheck
	permissionErrorStyle = baseErrorStyle.Copy().
				BorderForeground(lipgloss.Color("#FF00FF")).
				Foreground(lipgloss.Color("#FF00FF"))
	// nolint:staticcheck
	configurationErrorStyle = baseErrorStyle.Copy().
				BorderForeground(lipgloss.Color("#00FFFF")).
				Foreground(lipgloss.Color("#00FFFF"))
)

// GetStyle returns the appropriate style for the given error type
func GetStyle(errType ErrorType) lipgloss.Style {
	switch errType {
	case ErrorTypeValidation:
		return validationErrorStyle
	case ErrorTypeFileSystem:
		return fileSystemErrorStyle
	case ErrorTypePermission:
		return permissionErrorStyle
	case ErrorTypeConfiguration:
		return configurationErrorStyle
	default:
		return baseErrorStyle
	}
}
