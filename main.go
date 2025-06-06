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

func main() {
	var rules = rules.NewRules()
	rules.SetupRulesConfig()
	config := config.GetFlags()
	fm := filemanager.NewFileManager()

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
			MaxSize:    config.MaxSize,
			Extensions: extMap,
			Exclude:    config.Exclude,
			OlderThan:  config.OlderThan,
			NewerThan:  config.NewerThan,
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
			if !config.SkipConfirm {
				fmt.Println(utils.FormatSize(totalClearSize), "will be cleared.")
				actionIsDelete = printer.AskForConfirmation("Delete these files?")
			}

			if actionIsDelete {
				for path := range toDeleteMap {
					os.Remove(path)
				}
				printer.PrintSuccess("Deleted: %s", utils.FormatSize(totalClearSize))

				utils.LogDeletionToFile(toDeleteMap)
			}

		} else {
			printer.PrintWarning("File not found")
		}
		if config.DeleteEmptyFolders {
			printer.PrintInfo("Scan empty subfolders")
			toDeleteEmptyFolders := fileScanner.ScanEmptySubFolders(config.Directory)
			if len(toDeleteEmptyFolders) != 0 {
				printer.PrintEmptyDirs(toDeleteEmptyFolders)

				actionIsEmptyDeleteFolders := true

				if !config.SkipConfirm {
					actionIsEmptyDeleteFolders = printer.AskForConfirmation("Delete these empty folders?")
				}

				if actionIsEmptyDeleteFolders {
					for i := len(toDeleteEmptyFolders) - 1; i >= 0; i-- {
						os.Remove(toDeleteEmptyFolders[i])
					}
					fmt.Println()
					printer.PrintSuccess("Number of deleted empty folders: %d", len(toDeleteEmptyFolders))
				}
			} else {
				printer.PrintWarning("Empty folders not found")
			}
		}
	}
}
