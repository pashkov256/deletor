package logging

import "time"

type ScanStatistics struct {
	ScanID       string    `json:"scan_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	ScanDuration float64   `json:"scan_duration"`

	TotalFiles int64 `json:"total_files"`
	TotalDirs  int64 `json:"total_dirs"`
	TotalSize  int64 `json:"total_size"`

	DeletedFiles int64 `json:"deleted_files"`
	DeletedSize  int64 `json:"deleted_size"`
	IgnoredFiles int64 `json:"ignored_files"`
	IgnoredSize  int64 `json:"ignored_size"`
	TrashedFiles int64 `json:"trashed_files"`
	TrashedSize  int64 `json:"trashed_size"`

	ScanPath     string   `json:"scan_path"`
	RulesApplied []string `json:"rules_applied"`
	Extensions   []string `json:"extensions"`
}
