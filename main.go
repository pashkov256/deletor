package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pashkov256/deletor/internal/cli/config"
	"github.com/pashkov256/deletor/internal/cli/output"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui"
	"github.com/pashkov256/deletor/internal/utils"
)

type Task struct {
	info os.FileInfo
}

var extensionFromFlag bool
var sizeFromFlag bool
var ext []string
var size string

func init() {
	extensionFromFlag = false
	sizeFromFlag = false

	err := godotenv.Load()
	if err != nil {
		extensionFromFlag = true
		sizeFromFlag = true
	}

	ext = strings.Split(os.Getenv("EXTENSIONS"), ",")
	if len(ext) == 1 && ext[0] == "" {
		extensionFromFlag = true
	}

	size = os.Getenv("MAX_SIZE")
	if size == "" {
		sizeFromFlag = true
	}
}

func main() {
	config := config.GetFlags()

	var fm filemanager.FileManager = filemanager.NewFileManager()
	rules := rules.NewRules()

	if !config.IsCLIMode {
		// Start TUI
		if err := tui.Start(fm, rules); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	} else {
		extMap := make(map[string]bool)

		if len(config.Extensions) == 0 && !extensionFromFlag {
			config.Extensions = ext
		}

		// Populate extension map
		for _, extItem := range config.Extensions {
			if extItem == "" {
				continue
			}
			extMap[fmt.Sprint(".", extItem)] = true
		}

		fileScanner := filemanager.NewFileScanner(fm, &filemanager.FileFilter{
			MinSize:    config.MinSize,
			Extensions: extMap,
			Exclude:    config.Exclude,
		}, config.ShowProgress)

		printer := output.NewPrinter()

		// If no extensions specified, print usage
		if len(extMap) == 0 {
			fmt.Println("Error: No file extensions specified. Use -e flag or EXTENSIONS environment variable")
			fmt.Println("Example: -e \"jpg,png,mp4\" or EXTENSIONS=jpg,png,mp4")
			os.Exit(1)
		}

		if config.ShowProgress {
			fileScanner.ProgressBarScanner(config.Directory)
		}

		var toDeleteMap map[string]string
		var totalClearSize int64

		if config.IncludeSubdirs {
			toDeleteMap, totalClearSize = fileScanner.ScanFilesRecursively(config.Directory)
		} else {
			toDeleteMap, totalClearSize = fileScanner.ScanFilesCurrentLevel(config.Directory)
		}
		if len(toDeleteMap) != 0 {
			printer.PrintFilesTable(toDeleteMap)

			fmt.Println(utils.FormatSize(totalClearSize), "will be cleared.")

			actionIsDelete := true

			if !config.ConfirmDelete {
				actionIsDelete = printer.AskForConfirmation("Delete these files?")
			}

			if actionIsDelete {
				printer.PrintSuccess("Deleted: %s", utils.FormatSize(totalClearSize))

				for path := range toDeleteMap {
					os.Remove(path)
				}

				utils.LogDeletionToFile(toDeleteMap)
			}
		} else {
			printer.PrintWarning("File not found")
		}
	}

}
