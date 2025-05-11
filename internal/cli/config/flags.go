package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pashkov256/deletor/internal/utils"
)

func GetFlags() *Config {
	config := &Config{}

	extensions := flag.String("e", "", "File extensions to delete (comma-separated)")
	excludeFlag := flag.String("exclude", "", "Exclude specific files/paths (e.g. data,backup)")
	minSize := flag.String("min-size", "", "Minimum file size to delete (e.g. 10kb, 10mb, 10b)")
	dir := flag.String("d", ".", "Directory to scan")
	includeSubdirsScan := flag.Bool("subdirs", false, "Include subdirectories in scan")
	isCLIMode := flag.Bool("cli", false, "CLI mode (default is TUI)")
	progress := flag.Bool("progress", false, "Display a progress bar during file scanning")

	flag.Parse()

	*dir = utils.ExpandTilde(*dir)

	// Parse exclude patterns
	if *excludeFlag != "" {
		config.Exclude = strings.Split(*excludeFlag, ",")
	}

	// Convert extensions to slice
	if *extensions != "" {
		config.Extensions = strings.Split(*extensions, ",")
		for i := range config.Extensions {
			config.Extensions[i] = strings.TrimSpace(config.Extensions[i])
		}
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

	config.IsCLIMode = *isCLIMode
	config.HaveProgress = *progress
	config.IncludeSubdirs = *includeSubdirsScan
	config.Directory = *dir
	return config
}
