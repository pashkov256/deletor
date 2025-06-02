package output

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Printer handles formatted output with color coding for different message types
type Printer struct {
	successColor *color.Color // Green color for success messages
	errorColor   *color.Color // Red color for error messages
	warningColor *color.Color // Yellow color for warning messages
	infoColor    *color.Color // Cyan color for info messages
	progress     chan int64   // Channel for progress updates
}

// NewPrinter creates a new Printer instance with default color settings
func NewPrinter() *Printer {
	return &Printer{
		successColor: color.New(color.FgGreen),
		errorColor:   color.New(color.FgRed),
		warningColor: color.New(color.FgYellow),
		infoColor:    color.New(color.FgCyan),
		progress:     make(chan int64),
	}
}

// PrintSuccess prints a success message with a green checkmark
func (p *Printer) PrintSuccess(format string, args ...interface{}) {
	p.successColor.Printf("✓  %s\n", fmt.Sprintf(format, args...))
}

// PrintError prints an error message with a red X mark
func (p *Printer) PrintError(format string, args ...interface{}) {
	p.errorColor.Printf("✗  %s\n", fmt.Sprintf(format, args...))
}

// PrintWarning prints a warning message with a yellow warning symbol
func (p *Printer) PrintWarning(format string, args ...interface{}) {
	p.warningColor.Printf("⚠  %s\n", fmt.Sprintf(format, args...))
}

// PrintInfo prints an info message with a blue info symbol
func (p *Printer) PrintInfo(format string, args ...interface{}) {
	p.infoColor.Printf("ℹ  %s\n", fmt.Sprintf(format, args...))
}

// PrintFilesTable prints a formatted table of files with their sizes
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

// PrintEmptyDirs prints a list of empty directories
func (p *Printer) PrintEmptyDirs(files []string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	for _, path := range files {
		fmt.Printf("%s  %s\n", yellow("DIR"), white(path))
	}
}

// AskForConfirmation prompts the user for confirmation with a yes/no question
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
