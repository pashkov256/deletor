package tabs

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/tui/views"
	"github.com/pashkov256/deletor/internal/utils"
)

type MainTab struct {
	model *views.CleanFilesModel
}

func NewMainTab(model *views.CleanFilesModel) *MainTab {
	return &MainTab{
		model: model,
	}
}

func (t *MainTab) View() string {
	var content strings.Builder
	pathStyle := styles.StandardInputStyle
	if t.model.FocusedElement == "path" {
		pathStyle = styles.StandardInputFocusedStyle
	}
	content.WriteString(pathStyle.Render("Current Path: " + t.model.PathInput.View()))

	if t.model.GetCurrentPath() == "" {
		startButtonStyle := styles.LaunchButtonStyle
		if t.model.GetFocusedElement() == "startButton" {
			startButtonStyle = styles.LaunchButtonFocusedStyle
		}
		content.WriteString("\n")
		content.WriteString(startButtonStyle.Render("📂 Launch"))
	} else {
		// Show full interface when path is set
		extStyle := styles.StandardInputStyle
		if t.model.GetFocusedElement() == "ext" {
			extStyle = styles.StandardInputFocusedStyle
		}
		content.WriteString("\n")
		content.WriteString(extStyle.Render("Extensions: " + t.model.ExtInput.View()))
		content.WriteString("\n")
		var activeList list.Model
		if t.model.GetShowDirs() {
			activeList = t.model.DirList
		} else {
			activeList = t.model.List
		}
		fileCount := len(activeList.Items())
		filteredSizeText := utils.FormatSize(t.model.GetFilteredSize())
		content.WriteString("\n")
		if !t.model.GetShowDirs() {
			content.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Selected files (%d) • Size of selected files: %s",
				t.model.GetFilteredCount(), filteredSizeText)))
		} else {
			content.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Directories in %s (%d)",
				filepath.Base(t.model.GetCurrentPath()), fileCount)))
		}
		content.WriteString("\n")
		listStyle := styles.ListStyle
		if t.model.GetFocusedElement() == "list" {
			listStyle = styles.ListFocusedStyle
		}

		var listContent strings.Builder
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
				item := items[i].(views.CleanItem)

				icon := "📄 "
				if item.Size == -1 {
					icon = "⬆️ "
				} else if item.Size == 0 {
					icon = "📁 "
				} else {
					ext := strings.ToLower(filepath.Ext(item.Path))
					switch ext {
					case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".apng":
						icon = "🖼️ "
					case ".mp3", ".wav", ".flac", ".ogg":
						icon = "🎵 "
					case ".mp4", ".avi", ".mkv", ".mov":
						icon = "🎬 "
					case ".zip", ".rar", ".7z", ".tar", ".gz":
						icon = "🗜️ "
					case ".exe", ".msi":
						icon = "⚙️ "
					case ".pdf":
						icon = "📕 "
					case ".doc", ".docx", ".txt":
						icon = "📝 "
					}
				}
				filename := filepath.Base(item.Path)
				sizeStr := ""
				if item.Size > 0 {
					sizeStr = utils.FormatSize(item.Size)
				} else if item.Size == 0 {
					sizeStr = "DIR"
				} else {
					sizeStr = "UP DIR"
				}

				prefix := "  "
				style := lipgloss.NewStyle()

				if i == selectedIndex {
					prefix = "> "
					style = style.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#0066FF")).Bold(true)
				} else if item.Size == -1 || item.Size == 0 {
					style = style.Foreground(lipgloss.Color("#4DC4FF"))
				}

				displayName := filename
				if len(displayName) > 40 {
					displayName = displayName[:37] + "..."
				}

				padding := 44 - len(displayName)
				if padding < 1 {
					padding = 1
				}

				fileLine := fmt.Sprintf("%s%s%s%s%s",
					prefix,
					icon,
					displayName,
					strings.Repeat(" ", padding),
					sizeStr)

				listContent.WriteString(style.Render(fileLine))
				listContent.WriteString("\n")

			}
			if totalItems > visibleItems {
				scrollInfo := fmt.Sprintf("\nShowing %d-%d of %d items (%.0f%%)",
					startIdx+1, endIdx, totalItems,
					float64(selectedIndex+1)/float64(totalItems)*100)
				listContent.WriteString(lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#999999")).Render(scrollInfo))
			}
		}

		content.WriteString(listStyle.Render(listContent.String()))

		// Buttons section
		content.WriteString("\n\n")
		if t.model.GetFocusedElement() == "dirButton" {
			content.WriteString(styles.StandardButtonFocusedStyle.Render("➡️ Show directories"))
		} else {
			content.WriteString(styles.StandardButtonStyle.Render("➡️ Show directories"))
		}
		content.WriteString("\n\n")

		if t.model.GetFocusedElement() == "button" {
			content.WriteString(styles.DeleteButtonFocusedStyle.Render("🗑️ Start cleaning"))
		} else {
			content.WriteString(styles.DeleteButtonStyle.Render("🗑️ Start cleaning"))
		}
		content.WriteString("\n")
	}

	return content.String()
}

func (t *MainTab) Init() tea.Cmd {
	return nil
}

func (t *MainTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}
