package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"flag"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
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
	// Parse command line arguments
	extensions := flag.String("e", "", "File extensions to delete (comma-separated)")
	size := flag.String("s", "", "Minimum file size to delete (e.g. 10kb, 10mb, 10b)")
	dir := flag.String("d", ".", "Directory to scan")
	flag.Parse()

	// Convert extensions to slice
	var extSlice []string
	if *extensions != "" {
		extSlice = strings.Split(*extensions, ",")
		for i := range extSlice {
			extSlice[i] = strings.TrimSpace(extSlice[i])
		}
	}

	// Convert size to bytes
	var minSize int64
	if *size != "" {
		sizeBytes, err := toBytes(*size)
		if err != nil {
			fmt.Printf("Error parsing size: %v\n", err)
			os.Exit(1)
		}
		minSize = sizeBytes
	}

	// Get absolute path
	absPath, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	// Start TUI
	if err := startTUI(absPath, extSlice, minSize); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
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

func logDeletionToFile(files map[string]string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	const (
		DELETION_FILE_NAME = "deletor.log"
	)
	var deletionLogs string
	deletionTimestamp := time.Now().Format("2006-01-02 15:04:05")
	for path, size := range files {
		deletionLogs += fmt.Sprintf("[%s] %s | %s\n", deletionTimestamp, path, size)
	}
	fmt.Println(deletionLogs)
	file, err := os.OpenFile(DELETION_FILE_NAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(yellow("Error:"), "Failed to open deleted files")
		return
	}
	_, err = file.WriteString(deletionLogs)
	if err != nil {
		fmt.Println(yellow("Error:"), "Failed to save deleted files")
		return
	}
	defer file.Close()
}
