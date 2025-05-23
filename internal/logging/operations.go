package logging

import (
	"time"
)

type OperationType string

const (
	OperationDeleted OperationType = "deleted"
	OperationIgnored OperationType = "ignored"
	OperationTrashed OperationType = "trashed"
)

type FileOperation struct {
	ScanID        string        `json:"scan_id"`
	Timestamp     time.Time     `json:"timestamp"`
	FilePath      string        `json:"file_path"`
	FileSize      int64         `json:"file_size"`
	OperationType OperationType `json:"operation_type"`
	Reason        string        `json:"reason"`
	RuleApplied   string        `json:"rule_applied"`
}

func NewFileOperation(scanID, filePath string, size int64, opType OperationType, reason, rule string) *FileOperation {
	return &FileOperation{
		ScanID:        scanID,
		Timestamp:     time.Now(),
		FilePath:      filePath,
		FileSize:      size,
		OperationType: opType,
		Reason:        reason,
		RuleApplied:   rule,
	}
}
