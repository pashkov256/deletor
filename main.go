package main

import (
	"fmt"
	"os"

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

func main() {
	var rules = rules.NewRules()
	rules.SetupRulesConfig()
	config := config.GetFlags()
	var fm filemanager.FileManager = filemanager.NewFileManager()

	if !config.IsCLIMode {
		// Start TUI
		if err := tui.Start(fm, rules); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	} else {
		extMap := utils.ParseExtToMap(config.Extensions)

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

			actionIsDelete := true

			fmt.Println() // This is required for formatting
			if config.ConfirmDelete {
				fmt.Println(utils.FormatSize(totalClearSize), "will be cleared.")
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
