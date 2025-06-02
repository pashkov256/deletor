package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
)

// FormatSize converts a byte count into a human-readable string with appropriate unit
// Example: 1024 -> "1.00 KB", 1024*1024 -> "1.00 MB"
func FormatSize(bytes int64) string {
	const (
		KB = 1 << 10 // 1024
		MB = 1 << 20 // 1024 * 1024
		GB = 1 << 30 // 1024 * 1024 * 1024
		TB = 1 << 40 // 1024 * 1024 * 1024 * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ExpandTilde expands the tilde (~) in a path to the user's home directory
// Example: "~/Documents" -> "/home/user/Documents"
func ExpandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, path[1:])
}

// ToBytes converts a human-readable size string to bytes
// Example: "1.5MB" -> 1572864, "2GB" -> 2147483648
func ToBytes(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToLower(sizeStr))

	var unitIndex int
	for unitIndex = 0; unitIndex < len(sizeStr); unitIndex++ {
		if sizeStr[unitIndex] < '0' || sizeStr[unitIndex] > '9' {
			if sizeStr[unitIndex] != '.' {
				break
			}
		}
	}

	numStr := sizeStr[:unitIndex]
	unit := sizeStr[unitIndex:]

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %v", err)
	}

	var multiplier float64
	switch unit {
	case "b":
		multiplier = 1
	case "kb":
		multiplier = 1024
	case "mb":
		multiplier = 1024 * 1024
	case "gb":
		multiplier = 1024 * 1024 * 1024
	case "tb":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown unit of measurement: %s", unit)
	}

	bytes := num * multiplier
	return int64(bytes), nil
}

// ToBytesOrDefault converts a size string to bytes, returning 0 if conversion fails
func ToBytesOrDefault(sizeStr string) int64 {
	size, err := ToBytes(sizeStr)
	if err != nil {
		return 0 // Default to 0 if conversion fails
	}
	return size
}

// LogDeletionToFile writes deletion records to a log file
// Each record includes timestamp, file path, and file size
func LogDeletionToFile(files map[string]string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	const (
		DELETION_FILE_NAME = "deletor.log"
	)
	var deletionLogs string
	deletionTimestamp := time.Now().Format("2006-01-02 15:04:05")
	for path, size := range files {
		deletionLogs += fmt.Sprintf("[%s] %s | %s\n", deletionTimestamp, path, size)
	}

	file, err := os.OpenFile(DELETION_FILE_NAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(yellow("Error:"), "Failed to open deleted files")
		return
	}
	_, err = file.WriteString(deletionLogs)
	if err != nil {
		fmt.Println(yellow("Error:"), "Failed to save deleted files")
		return
	}
	defer file.Close()
}

// ParseExtToSlice converts a comma-separated string of extensions into a slice
func ParseExtToSlice(extensions string) []string {
	extSlice := make([]string, 0)
	if extensions != "" {

		for _, ext := range strings.Split(extensions, ",") {
			ext = strings.TrimSpace(ext)
			if ext != "" {
				// Add dot prefix if needed
				if !strings.HasPrefix(ext, ".") {
					ext = "." + ext
				}
				extSlice = append(extSlice, strings.ToLower(ext))
			}
		}
	}
	return extSlice
}

// ParseExcludeToSlice converts a comma-separated string of patterns into a slice
func ParseExcludeToSlice(exclude string) []string {
	excludeSlice := make([]string, 0)

	if exclude != "" {
		for _, exc := range strings.Split(exclude, ",") {
			exc = strings.TrimSpace(exc)
			if exc != "" {
				excludeSlice = append(excludeSlice, exc)
			}
		}
	}

	return excludeSlice
}

// ParseExtToMap converts a slice of extensions into a map for O(1) lookups
func ParseExtToMap(extSlice []string) map[string]struct{} {
	extMap := make(map[string]struct{})

	for _, extItem := range extSlice {
		if extItem == "" {
			continue
		}
		extMap[extItem] = struct{}{}
	}

	return extMap
}

// ParseTimeDuration converts a time duration string to a time.Time
func ParseTimeDuration(timeStr string) (time.Time, error) {
	timeStr = strings.TrimSpace(strings.ToLower(timeStr))

	// Find the first non-digit character
	var unitIndex int
	for unitIndex = 0; unitIndex < len(timeStr); unitIndex++ {
		if timeStr[unitIndex] < '0' || timeStr[unitIndex] > '9' {
			break
		}
	}

	if unitIndex == 0 {
		return time.Time{}, nil
	}

	// Parse the number
	numStr := timeStr[:unitIndex]
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid number: %s", numStr)
	}

	// Get the unit part
	unit := strings.TrimSpace(timeStr[unitIndex:])

	// Calculate the duration
	var duration time.Duration
	switch {
	case strings.HasPrefix(unit, "sec"):
		duration = time.Duration(num) * time.Second
	case strings.HasPrefix(unit, "min"):
		duration = time.Duration(num) * time.Minute
	case strings.HasPrefix(unit, "hour"):
		duration = time.Duration(num) * time.Hour
	case strings.HasPrefix(unit, "day"):
		duration = time.Duration(num) * 24 * time.Hour
	case strings.HasPrefix(unit, "week"):
		duration = time.Duration(num) * 7 * 24 * time.Hour
	case strings.HasPrefix(unit, "month"):
		duration = time.Duration(num) * 30 * 24 * time.Hour
	case strings.HasPrefix(unit, "year"):
		duration = time.Duration(num) * 365 * 24 * time.Hour
	default:
		return time.Time{}, fmt.Errorf("unknown time unit: %s", unit)
	}

	// Return the time that is duration from now
	return time.Now().Add(-duration), nil
}

func GenerateUUID() string {
	return uuid.New().String()
}
