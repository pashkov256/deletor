package tui

import "github.com/charmbracelet/lipgloss"

var buttonColor = lipgloss.Color("#1E90FF")
var buttonFocusedColor = lipgloss.Color("#1570CC")
var buttonWidth = 30

// Common styles for the entire application
var (
	// Base styles
	AppStyle = lipgloss.NewStyle().Padding(0, 2)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#1E90FF")).
			Padding(0, 1)

	// Input styles
	StandardInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#666666")).
				Padding(0, 0).
				Width(100)

	StandardInputFocusedStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#1E90FF")).
					Padding(0, 0).
					Width(100)

	// Text input styles
	TextInputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1E90FF"))

	TextInputTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	TextInputCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6666"))

	// Button styles
	StandardButtonStyle = lipgloss.NewStyle().
				Background(buttonColor).
				Foreground(lipgloss.Color("#fff")).
				Width(buttonWidth).
				Bold(true).
				AlignHorizontal(lipgloss.Center)

	StandardButtonFocusedStyle = lipgloss.NewStyle().
					Background(buttonFocusedColor).
					Foreground(lipgloss.Color("#fff")).
					Width(buttonWidth).
					Bold(true).
					AlignHorizontal(lipgloss.Center)

	// Special button styles (for delete and launch)
	DeleteButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fff")).
				Background(lipgloss.Color("#FF6666 ")).
				BorderForeground(lipgloss.Color("#FF6666")).
				Width(buttonWidth).
				AlignHorizontal(lipgloss.Center)

	DeleteButtonFocusedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#fff")).
					Background(lipgloss.Color("#CC5252")).
					Padding(0, 1).
					Width(buttonWidth).
					AlignHorizontal(lipgloss.Center)

	LaunchButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fff")).
				Background(lipgloss.Color("#42bd48")).
				Width(buttonWidth).
				Bold(true).
				AlignHorizontal(lipgloss.Center)

	LaunchButtonFocusedStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#2E7D32")).
					Foreground(lipgloss.Color("#fff")).
					Width(buttonWidth).
					Bold(true).
					AlignHorizontal(lipgloss.Center)

	// Tab styles
	TabStyle = lipgloss.NewStyle().
			Border(lipgloss.Border{
			Bottom: "─",
		}).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(0, 1).
		MarginRight(1)

	ActiveTabStyle = TabStyle.Copy().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1E90FF")).
			Foreground(lipgloss.Color("#1E90FF")).
			Bold(true)

	// List styles
	ListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(0, 0).
			Width(100).
			Height(9)

	ListFocusedStyle = ListStyle.Copy().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#1E90FF"))

	// Option styles
	OptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5"))

	SelectedOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ad58b3")).
				Bold(true)

	OptionFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#5f5fd7")).
				Background(lipgloss.Color("#333333"))

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	// Path style
	PathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)
)
