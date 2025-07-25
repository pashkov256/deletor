package runner

import (
	"fmt"
	"os"

	"github.com/pashkov256/deletor/internal/cli/config"
	"github.com/pashkov256/deletor/internal/cli/output"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/utils"
)

const (
	confirmMsgDlt   string = "Delete these files?"
	confirmMsgTrash string = "Move files to trash?"
)

func RunCLI(
	fm filemanager.FileManager,
	rules rules.Rules,
	config *config.Config,
) {
	// Get values from rules if --rules flag is set
	if config.UseRules {
		config = config.GetWithRules(rules)
	}

	extMap := utils.ParseExtToMap(config.Extensions)

	filter := fm.NewFileFilter(
		config.MinSize,
		config.MaxSize,
		extMap,
		config.Exclude,
		config.OlderThan,
		config.NewerThan,
	)

	fileScanner := filemanager.NewFileScanner(fm, filter, config.ShowProgress)
	printer := output.NewPrinter()

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
			var msg string
			if config.MoveFileToTrash {
				msg = confirmMsgTrash
			} else {
				msg = confirmMsgDlt
			}
			actionIsDelete = printer.AskForConfirmation(msg)
		}

		if actionIsDelete {
			if config.MoveFileToTrash {
				for path := range toDeleteMap {
					fm.MoveFileToTrash(path)
				}
				printer.PrintSuccess("Moved to trash: %s", utils.FormatSize(totalClearSize))
			} else {
				for path := range toDeleteMap {
					fm.DeleteFile(path)
				}
				printer.PrintSuccess("Deleted: %s", utils.FormatSize(totalClearSize))
			}

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
