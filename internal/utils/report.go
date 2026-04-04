package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// ReportEntry represents an entry in the exported report
type ReportEntry struct {
	Path      string    `json:"path"`
	Action    string    `json:"action"`
	Size      int64     `json:"size"`
	Timestamp time.Time `json:"timestamp"`
}

// GenerateReport exports the collected file actions to a JSON or CSV file
func GenerateReport(entries []ReportEntry, path string, format string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer file.Close()

	if format == "csv" {
		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write header
		if err := writer.Write([]string{"Path", "Action", "Size (bytes)", "Timestamp"}); err != nil {
			return fmt.Errorf("failed to write CSV header: %v", err)
		}

		// Write entries
		for _, entry := range entries {
			if err := writer.Write([]string{
				entry.Path,
				entry.Action,
				strconv.FormatInt(entry.Size, 10),
				entry.Timestamp.Format(time.RFC3339),
			}); err != nil {
				return fmt.Errorf("failed to write CSV entry: %v", err)
			}
		}
	} else {
		// Default to JSON
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(entries); err != nil {
			return fmt.Errorf("failed to encode JSON report: %v", err)
		}
	}

	return nil
}
