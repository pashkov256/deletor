package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "deletor",
		Usage: "Утилита для удаления файлов по расширению и размеру",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "extensions",
				Aliases:  []string{"e"},
				Usage:    "Список расширений файлов через запятую (например, mp4,zip,ttf)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "directory",
				Aliases:  []string{"d"},
				Usage:    "Директория для поиска файлов",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "size",
				Aliases: []string{"s"},
				Usage:   "Максимальный размер файла (например, 10mb, 1gb)",
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
			var totalClearSize int64

			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				for i := 0; i < len(ext); i++ {
					if info.Size() > sizeBytes {
						files = append(files, struct {
							Name string
							Size int64
						}{path, info.Size()})
						toDeleteMap[path] = formatSize(info.Size())
						totalClearSize += info.Size()
					}

				}
				return nil
			})
			printFilesTable(toDeleteMap)

			fmt.Println()
			fmt.Println(formatSize(totalClearSize), " will be cleared.")
			fmt.Println()

			actionIsDelete := askForConfirmation("Delete these files?")

			if !actionIsDelete {
				return nil
			}
			fmt.Println(formatSize(totalClearSize), " it was deleted.")

			for _, file := range files {
				os.Remove(file.Name)
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
	}
}

func printFilesTable(files map[string]string) {
	maxSizeLen := 0
	for _, file := range files {
		sizeStr := len(file)
		if sizeStr > maxSizeLen {
			maxSizeLen = sizeStr
		}
	}

	for name, file := range files {
		fmt.Printf("%-*s  %s\n", maxSizeLen, file, name)
	}
}

func deleteFiles(toDeleteMap map[string]string) {
	for dir, _ := range toDeleteMap {
		fmt.Println(dir)
		err := os.Remove(dir)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

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
		return 0, fmt.Errorf("неизвестная единица измерения: %s", unit)
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
