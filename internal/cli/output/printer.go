package output

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Printer struct {
	successColor *color.Color
	errorColor   *color.Color
	warningColor *color.Color
	infoColor    *color.Color
	progress     chan int64
}

func NewPrinter() *Printer {
	return &Printer{
		successColor: color.New(color.FgGreen),
		errorColor:   color.New(color.FgRed),
		warningColor: color.New(color.FgYellow),
		infoColor:    color.New(color.FgBlue),
		progress:     make(chan int64),
	}
}

func (p *Printer) PrintSuccess(format string, args ...interface{}) {
	p.successColor.Printf("✓  %s\n", fmt.Sprintf(format, args...))
}

func (p *Printer) PrintError(format string, args ...interface{}) {
	p.errorColor.Printf("✗  %s\n", fmt.Sprintf(format, args...))
}

func (p *Printer) PrintWarning(format string, args ...interface{}) {
	p.warningColor.Printf("⚠  %s\n", fmt.Sprintf(format, args...))
}

func (p *Printer) PrintInfo(format string, args ...interface{}) {
	p.infoColor.Printf("ℹ  %s\n", fmt.Sprintf(format, args...))
}

func (p *Printer) PrintFilesTable(files map[string]string) {
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

func (p *Printer) PrintEmptyDirs(files []string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	for _, path := range files {
		fmt.Printf("%s  %s\n", yellow("DIR"), white(path))
	}
}

func (p *Printer) AskForConfirmation(s string) bool {
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

// func (p *Printer) PrintStats(stats *Stats) {
// 	fmt.Printf("\nStatistics:\n")
// 	fmt.Printf("  Total files scanned: %d\n", stats.FilesScanned)
// 	fmt.Printf("  Total directories scanned: %d\n", stats.DirsScanned)
// 	fmt.Printf("  Total size: %s\n", utils.FormatSize(stats.TotalSize))
// 	fmt.Printf("  Files to delete: %d\n", stats.FilesToDelete)
// 	fmt.Printf("  Size to clear: %s\n", utils.FormatSize(stats.SizeToClear))
// }

// type Stats struct {
// 	FilesScanned  int
// 	DirsScanned   int
// 	TotalSize     int64
// 	FilesToDelete int
// 	SizeToClear   int64
// }
