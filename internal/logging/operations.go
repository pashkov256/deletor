package logging

import (
	"time"
)

// OperationType represents the type of file operation performed
type OperationType string

const (
	OperationDeleted OperationType = "deleted" // File was permanently deleted
	OperationIgnored OperationType = "ignored" // File was skipped
	OperationTrashed OperationType = "trashed" // File was moved to trash
)

// FileOperation records details about a single file operation
type FileOperation struct {
	Timestamp     time.Time     `json:"timestamp"`      // When the operation occurred
	FilePath      string        `json:"file_path"`      // Path to the affected file
	FileSize      int64         `json:"file_size"`      // Size of the file in bytes
	OperationType OperationType `json:"operation_type"` // Type of operation performed
	Reason        string        `json:"reason"`         // Why the operation was performed
	RuleApplied   string        `json:"rule_applied"`   // Which rule triggered the operation
}

// NewFileOperation creates a new file operation record
func NewFileOperation(filePath string, size int64, opType OperationType, reason, rule string) *FileOperation {
	return &FileOperation{
		Timestamp:     time.Now(),
		FilePath:      filePath,
		FileSize:      size,
		OperationType: opType,
		Reason:        reason,
		RuleApplied:   rule,
	}
}
