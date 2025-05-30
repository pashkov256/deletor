package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/pashkov256/deletor/internal/utils"
)

func GetFlags() *Config {
	config := &Config{}

	extensions := flag.String("e", "", "File extensions to delete (comma-separated)")
	excludeFlag := flag.String("exclude", "", "Exclude specific files/paths (e.g. data,backup)")
	minSize := flag.String("min-size", "", "Minimum file size to delete (e.g. 10kb, 10mb, 10b)")
	maxSize := flag.String("max-size", "", "Maximum file size to delete (e.g. 10kb, 10mb, 10b)")
	dir := flag.String("d", ".", "Directory to scan")
	includeSubdirsScan := flag.Bool("subdirs", false, "Include subdirectories in scan")
	isCLIMode := flag.Bool("cli", false, "CLI mode (default is TUI)")
	progress := flag.Bool("progress", false, "Display a progress bar during file scanning")
	confirmDelete := flag.Bool("confirm-delete", false, "Confirm that files are to be deleted?")

	flag.Parse()

	*dir = utils.ExpandTilde(*dir)

	// Parse exclude patterns
	if *excludeFlag != "" {
		config.Exclude = utils.ParseExcludeToSlice(*excludeFlag)
	}

	// Convert extensions to slice
	if *extensions != "" {
		config.Extensions = utils.ParseExtToSlice(*extensions)
	}

	// Convert size to bytes
	if *minSize != "" {
		sizeBytes, err := utils.ToBytes(*minSize)
		if err != nil {
			fmt.Printf("Error parsing size: %v\n", err)
			os.Exit(1)
		}
		config.MinSize = sizeBytes
	}

	// Convert size to bytes
	if *maxSize != "" {
		sizeBytes, err := utils.ToBytes(*maxSize)
		if err != nil {
			fmt.Printf("Error parsing size: %v\n", err)
			os.Exit(1)
		}
		config.MaxSize = sizeBytes
	}

	config.IsCLIMode = *isCLIMode
	config.HaveProgress = *progress
	config.IncludeSubdirs = *includeSubdirsScan
	config.Directory = *dir
	config.ConfirmDelete = *confirmDelete
	return config
}
