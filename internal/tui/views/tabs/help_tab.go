package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pashkov256/deletor/internal/tui/models"
	"github.com/pashkov256/deletor/internal/tui/styles"
)

type HelpTab struct {
	model *models.CleanFilesModel
}

func NewHelpTab(model *models.CleanFilesModel) *HelpTab {
	return &HelpTab{
		model: model,
	}
}

func (t *HelpTab) View() string {
	var content strings.Builder
	// Navigation
	content.WriteString(styles.OptionStyle.Render("Navigation:"))
	content.WriteString("\n")
	content.WriteString("  F1-F4    - Switch between tabs\n")
	content.WriteString("  Esc      - Return to main menu\n")
	content.WriteString("  Tab      - Next field\n")
	content.WriteString("  Shift+Tab - Previous field\n")
	content.WriteString("  Ctrl+C   - Exit application\n\n")

	// File Operations
	content.WriteString(styles.OptionStyle.Render("File Operations:"))
	content.WriteString("\n")
	content.WriteString("  Ctrl+R   - Refresh file list\n")
	content.WriteString("  Crtl+O   - Open in explorer\n")
	content.WriteString("  Ctrl+D   - Delete files\n\n")

	// Filter Operations
	content.WriteString(styles.OptionStyle.Render("Filter Operations:"))
	content.WriteString("\n")
	content.WriteString("  Alt+C    - Clear filters\n\n")

	// Options
	content.WriteString(styles.OptionStyle.Render("Options:"))
	content.WriteString("\n")
	content.WriteString("  Alt+1    - Toggle hidden files\n")
	content.WriteString("  Alt+2    - Toggle confirm deletion\n")
	content.WriteString("  Alt+3    - Toggle include subfolders\n")
	content.WriteString("  Alt+4    - Toggle delete empty subfolders\n")

	return content.String()
}

func (t *HelpTab) Init() tea.Cmd {
	return nil
}

func (t *HelpTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}
