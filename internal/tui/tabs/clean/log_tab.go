package clean

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pashkov256/deletor/internal/logging"
	"github.com/pashkov256/deletor/internal/tui/interfaces"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/utils"
)

type LogTab struct {
	model      interfaces.CleanModel
	stats      *logging.ScanStatistics
	startTime  time.Time
	totalStats *logging.ScanStatistics
}

func (t *LogTab) Init() tea.Cmd {
	// Initialize with empty statistics and program start time
	t.startTime = time.Now()
	t.stats = &logging.ScanStatistics{
		StartTime:     t.startTime,
		Directory:     t.model.GetCurrentPath(),
		OperationType: "none",
	}
	// Initialize total statistics
	t.totalStats = &logging.ScanStatistics{
		StartTime:     t.startTime,
		Directory:     t.model.GetCurrentPath(),
		OperationType: "total",
		TotalFiles:    0,
		TotalSize:     0,
		DeletedFiles:  0,
		DeletedSize:   0,
		TrashedFiles:  0,
		TrashedSize:   0,
		IgnoredFiles:  0,
		IgnoredSize:   0,
	}
	return nil
}

func (t *LogTab) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (t *LogTab) View() string {
	var content strings.Builder
	disableEmoji := t.model.GetOptionState()[options.DisableEmoji]

	// Check if statistics are enabled
	if !t.model.GetOptionState()[options.ShowStatistics] {
		statsMsg := "Statistics display is disabled. Enable 'Show statistics' in Options tab (F3)."
		if !disableEmoji {
			statsMsg = "⚠️" + statsMsg + "⚠️"
		}
		return styles.InfoStyle.Render("\n" + statsMsg)
	}

	tableStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(1, 2)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Width(25).
		Align(lipgloss.Left)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		PaddingLeft(1)

	// Initialize stats if nil
	if t.stats == nil {
		t.stats = &logging.ScanStatistics{
			StartTime:     t.startTime,
			Directory:     t.model.GetCurrentPath(),
			OperationType: "none",
		}
	}

	// Format duration - use program start time
	duration := time.Since(t.startTime)
	durationStr := fmt.Sprintf("%.2f seconds", duration.Seconds())

	// Format time - use program start time
	timeStr := t.startTime.Format("02.01.2006 15:04:05 ")

	rows := []struct {
		emoji        string
		label        string
		value        string
		newlineAfter bool
	}{
		{"🔄", "Last operation", t.stats.OperationType, false},
		{"📂", "Directory", t.stats.Directory, false},
		{"⏰", "Start Time", timeStr, false},
		{"⏱️", "Program lifetime", durationStr, true},
		{"📝", "Total Files", fmt.Sprintf("%d", t.totalStats.TotalFiles), false},
		{"💾", "Total Size", utils.FormatSize(t.totalStats.TotalSize), true},
		{"🗑️", "Deleted Files", fmt.Sprintf("%d", t.totalStats.DeletedFiles), false},
		{"📈", "Deleted Size", utils.FormatSize(t.totalStats.DeletedSize), true},
		{"♻️", "Trashed Files", fmt.Sprintf("%d", t.totalStats.TrashedFiles), false},
		{"📈", "Trashed Size", utils.FormatSize(t.totalStats.TrashedSize), true},
		{"🚫", "Ignored Files", fmt.Sprintf("%d", t.totalStats.IgnoredFiles), false},
		{"📈", "Ignored Size", utils.FormatSize(t.totalStats.IgnoredSize), false},
	}
	// Create table content
	var tableContent strings.Builder
	for _, row := range rows {
		label := row.label
		if !disableEmoji {
			label = row.emoji + " " + label
		}
		tableContent.WriteString(labelStyle.Render(label))
		tableContent.WriteString(valueStyle.Render(row.value))
		tableContent.WriteString("\n")
		if row.newlineAfter {
			tableContent.WriteString("\n")
		}
	}

	// Render table with border
	content.WriteString(tableStyle.Render(tableContent.String()))
	content.WriteString(styles.PathStyle.Render(fmt.Sprintf("\n\nLog are stored in: %s", logging.GetLogFilePath())))
	return content.String()
}

func (t *LogTab) UpdateStats(stats *logging.ScanStatistics) {
	if stats != nil {
		// Update current operation stats
		t.stats = stats
		t.stats.StartTime = t.startTime

		// Update total statistics
		t.totalStats.TotalFiles += stats.TotalFiles
		t.totalStats.TotalSize += stats.TotalSize
		t.totalStats.DeletedFiles += stats.DeletedFiles
		t.totalStats.DeletedSize += stats.DeletedSize
		t.totalStats.TrashedFiles += stats.TrashedFiles
		t.totalStats.TrashedSize += stats.TrashedSize
		t.totalStats.IgnoredFiles += stats.IgnoredFiles
		t.totalStats.IgnoredSize += stats.IgnoredSize

		// Force a redraw by sending a nil message to the model
		t.model.Update(nil)
	}
}
