package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pashkov256/deletor/internal/utils"
)

type CleanItem struct {
	Path  string
	Size  int64
	IsDir bool
}

func (i CleanItem) Title() string {
	if i.Size == -1 {
		return "📂 .." // Parent directory
	}

	if i.IsDir {
		return "📁 " + filepath.Base(i.Path) // Directory
	}

	// Regular file
	filename := filepath.Base(i.Path)
	ext := filepath.Ext(filename)

	// Choose icon based on file extension
	icon := "📄 " // Default file icon
	switch strings.ToLower(ext) {
	// Programming languages
	case ".go":
		icon = "🐹 " // Go mascot
	case ".js", ".jsx":
		icon = "📜 " // JavaScript
	case ".ts", ".tsx":
		icon = "📘 " // TypeScript
	case ".py":
		icon = "🐍 " // Python
	case ".java":
		icon = "☕ " // Java
	case ".cpp", ".c", ".h":
		icon = "⚙️ " // C/C++
	case ".rs":
		icon = "🦀 " // Rust
	case ".php":
		icon = "🐘 " // PHP
	case ".rb":
		icon = "💎 " // Ruby
	case ".swift":
		icon = "🐦 " // Swift
	case ".kt", ".kts":
		icon = "⚡ " // Kotlin
	case ".scala":
		icon = "⚡ " // Scala
	case ".hs":
		icon = "λ " // Haskell
	case ".lua":
		icon = "🌙 " // Lua
	case ".sh", ".bash":
		icon = "🐚 " // Shell
	case ".ps1":
		icon = "💻 " // PowerShell
	case ".bat", ".cmd":
		icon = "🪟 " // Windows batch
	case ".env":
		icon = "⚙️ " // Environment file
	case ".json":
		icon = "📋 " // JSON
	case ".xml":
		icon = "📑 " // XML
	case ".yaml", ".yml":
		icon = "📝 " // YAML
	case ".toml":
		icon = "⚙️ " // TOML
	case ".ini", ".cfg", ".conf":
		icon = "⚙️ " // Config files
	case ".md", ".markdown":
		icon = "📖 " // Markdown
	case ".txt":
		icon = "📝 " // Text
	case ".log":
		icon = "📋 " // Log files
	case ".csv":
		icon = "📊 " // CSV
	case ".xlsx", ".xls":
		icon = "📊 " // Excel
	case ".doc", ".docx":
		icon = "📄 " // Word
	case ".pdf":
		icon = "📕 " // PDF
	case ".ppt", ".pptx":
		icon = "📑 " // PowerPoint
	case ".html", ".htm":
		icon = "🌐 " // HTML
	case ".css":
		icon = "🎨 " // CSS
	case ".scss", ".sass":
		icon = "🎨 " // SASS/SCSS
	case ".sql":
		icon = "🗄️ " // SQL
	case ".db", ".sqlite", ".sqlite3":
		icon = "🗄️ " // Database
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg":
		icon = "🖼️ " // Images
	case ".mp3", ".wav", ".flac", ".ogg", ".m4a":
		icon = "🎵 " // Audio
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".webm":
		icon = "🎬 " // Video
	case ".zip", ".rar", ".tar", ".gz", ".7z", ".bz2":
		icon = "🗜️ " // Archives
	case ".exe", ".msi", ".app":
		icon = "⚙️ " // Executables
	case ".dll", ".so", ".dylib":
		icon = "🔧 " // Libraries
	case ".iso", ".img":
		icon = "💿 " // Disk images
	case ".ttf", ".otf", ".woff", ".woff2":
		icon = "📝 " // Fonts
	case ".gitignore":
		icon = "🚫 " // Git ignore
	case ".git":
		icon = "📦 " // Git
	case ".dockerfile", ".dockerignore":
		icon = "🐳 " // Docker
	case ".lock":
		icon = "🔒 " // Lock files
	case ".key", ".pem", ".crt", ".cer":
		icon = "🔑 " // Certificates/Keys
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
