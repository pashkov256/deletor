package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	cli "github.com/urfave/cli/v2"
)

type Task struct {
	info os.FileInfo
}

func main() {
	app := &cli.App{
		Name:  "deletor",
		Usage: "A utility for deleting files by extension and size",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "extensions",
				Aliases:  []string{"e"},
				Usage:    "Comma-separated list of file extensions (e.g. mp4,zip,rtf)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "directory",
				Aliases:  []string{"d"},
				Usage:    "File search directory",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "size",
				Aliases: []string{"s"},
				Usage:   "Maximum file size (for example, 10mb, 1gb)",
			},
		},
		Action: func(c *cli.Context) error {
			ext := strings.Split(c.String("extensions"), ",")
			dir := c.String("directory")
			size := c.String("size")

			sizeBytes, _ := toBytes(size)

			toDeleteMap := make(map[string]string, 16)

			files := make([]struct {
				Name string
				Size int64
			}, 0, 0)
			var wg sync.WaitGroup
			numCPU := runtime.NumCPU()
			taskCh := make(chan Task, numCPU)

			var totalClearSize int64

			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				wg.Add(1)
				if info == nil {
					return nil
				}
				go func() {
					taskCh <- Task{info: info}
					defer wg.Done()
					for i := 0; i < len(ext); i++ {
						if info.Size() > sizeBytes && fmt.Sprint(".", ext[i]) == filepath.Ext(info.Name()) {
							files = append(files, struct {
								Name string
								Size int64
							}{path, info.Size()})
							toDeleteMap[path] = formatSize(info.Size())
							totalClearSize += info.Size()
							break
						}

					}

					<-taskCh
				}()

				return nil
			})

			wg.Wait()

			if totalClearSize != 0 {
				printFilesTable(toDeleteMap)

				fmt.Println()
				fmt.Println(formatSize(totalClearSize), " will be cleared.\n")

				actionIsDelete := askForConfirmation("Delete these files?")

				if !actionIsDelete {
					return nil
				}

				fmt.Println(color.New(color.FgGreen).SprintFunc()("✓"), "Deleted:", formatSize(totalClearSize))
				for _, file := range files {
					os.Remove(file.Name)
				}
			} else {
				red := color.New(color.FgRed).SprintFunc()
				fmt.Println(red("Error:"), "File not found")
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
	}
}
func printFilesTable(files map[string]string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	maxSizeLen := 0
	for _, size := range files {
		if len(size) > maxSizeLen {
			maxSizeLen = len(size)
		}
	}

	for path, size := range files {
		fmt.Printf("%s  %s\n", yellow(fmt.Sprintf("%-*s", maxSizeLen, size)), white(path))
	}
}

func askForConfirmation(s string) bool {
	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s %s ", bold(s), green("[y/n]:"))

	for {

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		fmt.Print("\n")

		switch response {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}

func toBytes(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToLower(sizeStr))

	var unitIndex int
	for unitIndex = 0; unitIndex < len(sizeStr); unitIndex++ {
		if sizeStr[unitIndex] < '0' || sizeStr[unitIndex] > '9' {
			if sizeStr[unitIndex] != '.' {
				break
			}
		}
	}

	numStr := sizeStr[:unitIndex]
	unit := sizeStr[unitIndex:]

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("неверный формат числа: %v", err)
	}

	var multiplier float64
	switch unit {
	case "b":
		multiplier = 1
	case "kb":
		multiplier = 1024
	case "mb":
		multiplier = 1024 * 1024
	case "gb":
		multiplier = 1024 * 1024 * 1024
	case "tb":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown unit of measurement: %s", unit)
	}

	bytes := num * multiplier
	return int64(bytes), nil
}

func formatSize(bytes int64) string {
	const (
		KB = 1 << 10 // 1024
		MB = 1 << 20 // 1024 * 1024
		GB = 1 << 30 // 1024 * 1024 * 1024
		TB = 1 << 40 // 1024 * 1024 * 1024 * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
