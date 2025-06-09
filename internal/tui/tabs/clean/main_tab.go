package clean

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/models"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/utils"
)

type MainTab struct {
	model interfaces.CleanModel
}

func (t *MainTab) Init() tea.Cmd              { return nil }
func (t *MainTab) Update(msg tea.Msg) tea.Cmd { return nil }

func (t *MainTab) View() string {
	var content strings.Builder

	// Path input
	pathStyle := styles.StandardInputStyle
	if t.model.GetFocusedElement() == "pathInput" {
		pathStyle = styles.StandardInputFocusedStyle
	}
	content.WriteString(zone.Mark("main_path_input", pathStyle.Render("Current Path: "+t.model.GetPathInput().View())))
	content.WriteString("\n")

	// If no path is set, show only the start button
	if t.model.GetCurrentPath() == "" {
		startButtonStyle := styles.LaunchButtonStyle
		if t.model.GetFocusedElement() == "startButton" {
			startButtonStyle = styles.LaunchButtonFocusedStyle
		}
		content.WriteString("\n")
		content.WriteString(zone.Mark("main_start_button", startButtonStyle.Render("📂 Launch")))
		content.WriteString("\n")
	} else {
		// Show full interface when path is set
		extStyle := styles.StandardInputStyle
		if t.model.GetFocusedElement() == "extInput" {
			extStyle = styles.StandardInputFocusedStyle
		}
		content.WriteString(zone.Mark("main_ext_input", extStyle.Render("Extensions: "+t.model.GetExtInput().View())))
		content.WriteString("\n\n")

		if !t.model.GetShowDirs() {
			content.WriteString(styles.ListTitleStyle.Render(fmt.Sprintf("Selected files (%d) • Size of selected files: %s",
				t.model.GetFilteredCount(), utils.FormatSize(t.model.GetFilteredSize()))))
		} else {
			content.WriteString(styles.ListTitleStyle.Render(fmt.Sprintf("Directories in %s (%d)",
				filepath.Base(t.model.GetCurrentPath()), len(t.model.GetDirList().Items()))))
		}
		content.WriteString("\n")
		// List content
		var listContent strings.Builder

		// List items
		var activeList list.Model
		if t.model.GetShowDirs() {
			activeList = t.model.GetDirList()
		} else {
			activeList = t.model.GetList()
		}

		if len(activeList.Items()) == 0 {
			if !t.model.GetShowDirs() {
				listContent.WriteString("No files match your filters. Try changing extensions or size filters.")
			} else {
				listContent.WriteString("No directories found in this location.")
			}
		} else {
			items := activeList.Items()
			selectedIndex := activeList.Index()
			totalItems := len(items)

			visibleItems := 10
			if visibleItems > totalItems {
				visibleItems = totalItems
			}

			startIdx := 0
			if selectedIndex > visibleItems-3 && totalItems > visibleItems {
				startIdx = selectedIndex - (visibleItems / 2)
				if startIdx+visibleItems > totalItems {
					startIdx = totalItems - visibleItems
				}
			}
			if startIdx < 0 {
				startIdx = 0
			}

			endIdx := startIdx + visibleItems
			if endIdx > totalItems {
				endIdx = totalItems
			}

			for i := startIdx; i < endIdx; i++ {
				item := items[i].(models.CleanItem)
				icon := utils.GetFileIcon(item.Size, item.Path, item.IsDir)
				filename := filepath.Base(item.Path)
				sizeStr := ""
				if item.Size >= 0 && !item.IsDir {
					sizeStr = utils.FormatSize(item.Size)
				} else if item.Size == -1 {
					sizeStr = "UP TO DIR"
				} else if item.IsDir {
					sizeStr = "DIR"
				}
				prefix := "  "
				style := lipgloss.NewStyle()

				if i == selectedIndex {
					prefix = "> "
					style = style.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#0066ff")).Bold(true)
				} else if item.IsDir && item.Size != -1 {
					style = style.Foreground(lipgloss.Color("#ccc"))
				} else if item.Size == -1 {
					style = style.Foreground(lipgloss.Color("#578cdb"))
				}

				const iconWidth = 3
				const filenameWidth = 45
				const sizeWidth = 10

				iconDisplay := fmt.Sprintf("%-*s", iconWidth, icon)
				displayName := filename
				if len(displayName) > filenameWidth {
					displayName = displayName[:filenameWidth-3] + "..."
				}
				sizeDisplay := fmt.Sprintf("%-*s", sizeWidth, sizeStr)

				line := fmt.Sprintf("%s%s%-*s%s",
					prefix,
					iconDisplay,
					filenameWidth, displayName,
					sizeDisplay)

				listContent.WriteString(style.Render(line))
				listContent.WriteString("\n")
			}

			if totalItems > visibleItems {
				scrollInfo := fmt.Sprintf("\nShowing %d-%d of %d items (%.0f%%)",
					startIdx+1, endIdx, totalItems,
					float64(selectedIndex+1)/float64(totalItems)*100)
				listContent.WriteString(lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#999999")).Render(scrollInfo))
			}
		}

		listStyle := styles.ListStyle
		if t.model.GetFocusedElement() == "list" {
			listStyle = styles.ListFocusedStyle
		}
		content.WriteString(listStyle.Render(listContent.String()))

		// Buttons section
		content.WriteString("\n\n")
		if t.model.GetFocusedElement() == "dirButton" {
			content.WriteString(zone.Mark("main_dir_button", styles.StandardButtonFocusedStyle.Render("➡️  Show directories")))
		} else {
			content.WriteString(zone.Mark("main_dir_button", styles.StandardButtonStyle.Render("➡️  Show directories")))
		}
		content.WriteString("  ")

		if t.model.GetFocusedElement() == "deleteButton" {
			content.WriteString(zone.Mark("main_delete_button", styles.DeleteButtonFocusedStyle.Render("🗑️  Start cleaning")))
		} else {
			content.WriteString(zone.Mark("main_delete_button", styles.DeleteButtonStyle.Render("🗑️  Start cleaning")))
		}
		content.WriteString("\n")
	}

	return content.String()
}
