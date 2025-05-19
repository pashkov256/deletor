package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pashkov256/deletor/internal/utils"
)

type CleanItem struct {
	Path string
	Size int64
}

func (i CleanItem) Title() string {
	if i.Size == -1 {
		return "📂 .." // Parent directory
	}

	if i.Size == 0 {
		return "📁 " + filepath.Base(i.Path) // Directory
	}

	// Regular file
	filename := filepath.Base(i.Path)
	ext := filepath.Ext(filename)

	// Choose icon based on file extension
	icon := "📄 " // Default file icon
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		icon = "🖼️ " // Image
	case ".mp3", ".wav", ".flac", ".ogg":
		icon = "🎵 " // Audio
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv":
		icon = "🎬 " // Video
	case ".pdf":
		icon = "📕 " // PDF
	case ".doc", ".docx", ".txt", ".rtf":
		icon = "📝 " // Document
	case ".zip", ".rar", ".tar", ".gz", ".7z":
		icon = "🗜️ " // Archive
	case ".exe", ".msi", ".bat":
		icon = "⚙️ " // Executable
	}

	// Format the size with unit
	sizeStr := utils.FormatSize(i.Size)

	// Calculate padding for alignment
	padding := 50 - len(filename)
	if padding < 0 {
		padding = 0
	}

	return fmt.Sprintf("%s%s%s%s", icon, filename, strings.Repeat(" ", padding), sizeStr)
}

func (i CleanItem) Description() string { return i.Path }
func (i CleanItem) FilterValue() string { return i.Path }
