package styles

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
			Padding(0, 1).MarginTop(2).Bold(true).Italic(true).Underline(true)

	ListTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#1E90FF")).
			Padding(0, 1).Bold(true)

	DocStyle = lipgloss.NewStyle().
			Padding(1, 1)

	MenuItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).MarginBottom(1)

	// nolint:staticcheck
	SelectedMenuItemStyle = MenuItem.Copy().
				Foreground(lipgloss.Color("#1E90FF"))

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

	// nolint:staticcheck
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

	// nolint:staticcheck
	ListFocusedStyle = ListStyle.Copy().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("#0067cf"))

	// Item styles
	ItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dddddd"))

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1E90FF")).
				Background(lipgloss.Color("#0066ff")).
				Bold(true)

	// Option styles
	OptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5"))

	SelectedOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ad58b3")).
				Bold(true)

	OptionFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#5f5fd7")).
				Background(lipgloss.Color("#333333"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			Padding(0, 1).
			Margin(1, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF00"))

	PathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ebd700")).
			Padding(0, 1)

	ScanResultHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Bold(true).
				Padding(0, 1)

	ScanResultPathStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)

	ScanResultSizeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fff")).
				Bold(true).
				Padding(0, 1)

	ScanResultFilesStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fff")).
				Bold(true).
				Padding(0, 1)

	ScanResultEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fff")).
				Italic(true).
				Padding(0, 1)

	ScanResultSizeCellStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fff")).
				Padding(0, 1).
				Align(lipgloss.Right)

	ScanResultFilesCellStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#fff")).
					Padding(0, 1).
					Align(lipgloss.Right)
)
