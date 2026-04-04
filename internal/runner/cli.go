package runner

import (
	"fmt"
	"os"
	"time"

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

	filter := config.BuildFileFilter()

	fileScanner := filemanager.NewFileScanner(fm, filter, config.ShowProgress)
	printer := output.NewPrinter()

	if config.ShowProgress {
		fileScanner.ProgressBarScanner(config.Directory)
	}

	var toDeleteMap map[string]string
	var totalClearSize int64
	var reportEntries []utils.ReportEntry

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
			actionName := "delete"
			if config.MoveFileToTrash {
				actionName = "trash"
			}
			if config.DryRun {
				actionName = "planned " + actionName
			}

			for path := range toDeleteMap {
				if config.ReportPath != "" {
					info, err := os.Stat(path)
					if err == nil {
						reportEntries = append(reportEntries, utils.ReportEntry{
							Path:      path,
							Action:    actionName,
							Size:      info.Size(),
							Timestamp: info.ModTime(),
						})
					}
				}

				if !config.DryRun {
					if config.MoveFileToTrash {
						fm.MoveFileToTrash(path)
					} else {
						fm.DeleteFile(path)
					}
				}
			}

			if !config.DryRun {
				if config.MoveFileToTrash {
					printer.PrintSuccess("Moved to trash: %s", utils.FormatSize(totalClearSize))
				} else {
					printer.PrintSuccess("Deleted: %s", utils.FormatSize(totalClearSize))
				}

				if config.JsonLogsEnabled {
					utils.LogDeletionToFileAsJson(toDeleteMap, config.JsonLogsPath)
				} else {
					utils.LogDeletionToFile(toDeleteMap)
				}
			} else {
				printer.PrintInfo("Dry run: no files were actually deleted or moved.")
			}
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
				actionName := "delete empty folder"
				if config.DryRun {
					actionName = "planned delete empty folder"
				}

				for i := len(toDeleteEmptyFolders) - 1; i >= 0; i-- {
					path := toDeleteEmptyFolders[i]
					if config.ReportPath != "" {
						info, err := os.Stat(path)
						if err == nil {
							reportEntries = append(reportEntries, utils.ReportEntry{
								Path:      path,
								Action:    actionName,
								Size:      0,
								Timestamp: info.ModTime(),
							})
						}
					}

					if !config.DryRun {
						os.Remove(path)
					}
				}

				if !config.DryRun {
					fmt.Println()
					printer.PrintSuccess("Number of deleted empty folders: %d", len(toDeleteEmptyFolders))
				} else {
					printer.PrintInfo("Dry run: empty folders were not deleted.")
				}
			}
		} else {
			printer.PrintWarning("Empty folders not found")
		}
	}
	if config.ReportPath != "" {
		if err := utils.GenerateReport(reportEntries, config.ReportPath, config.ReportFormat); err != nil {
			printer.PrintError("Failed to generate report: %v", err)
		} else {
			printer.PrintSuccess("Report exported to: %s", config.ReportPath)
		}
	}
}
