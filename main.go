package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	rules "github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/tui"
	"github.com/schollz/progressbar/v3"
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
	excludeFlag := flag.String("exclude", "", "Exclude specific files/paths (e.g. data,backup)")
	size := flag.String("s", "", "Minimum file size to delete (e.g. 10kb, 10mb, 10b)")
	dir := flag.String("d", ".", "Directory to scan")
	isCLIMode := flag.Bool("cli", false, "CLI mode")
	progress := *flag.Bool("progress", false, "Display a progress bar during file scanning")
	flag.Parse()

	// Parse exclude patterns after flag.Parse()
	var exclude []string
	if *excludeFlag != "" {
		exclude = strings.Split(*excludeFlag, ",")
	}

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
	if !*isCLIMode {
		// Start TUI
		if err := startTUI(absPath, extSlice, minSize); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	} else {
		var mutex sync.Mutex
		var wg sync.WaitGroup
		var totalClearSize int64
		var totalScanSize int64
		var progressChan chan int64

		toDeleteMap := make(map[string]string, 16)
		numCPU := runtime.NumCPU()
		taskCh := make(chan Task, numCPU)
		extMap := make(map[string]bool)

		files := make([]struct {
			Name string
			Size int64
		}, 0, 0)

		// Use command line extensions if provided, otherwise use env
		extensionsToUse := extSlice
		if len(extensionsToUse) == 0 && !extensionFromFlag {
			extensionsToUse = ext
		}

		// Populate extension map
		for _, extItem := range extensionsToUse {
			if extItem == "" {
				continue
			}
			extMap[fmt.Sprint(".", extItem)] = true
		}

		// If no extensions specified, print usage
		if len(extMap) == 0 {
			fmt.Println("Error: No file extensions specified. Use -e flag or EXTENSIONS environment variable")
			fmt.Println("Example: -e \"jpg,png,mp4\" or EXTENSIONS=jpg,png,mp4")
			os.Exit(1)
		}

		if progress {
			filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {

				if info == nil {
					return nil
				}

				if len(exclude) != 0 {
					for _, excludePattern := range exclude {
						if strings.Contains(filepath.ToSlash(path), excludePattern+"/") {
							return nil
						} else if strings.HasPrefix(info.Name(), excludePattern) {
							return nil
						}
					}
				}

				if info.Size() > minSize && extMap[filepath.Ext(info.Name())] {
					totalScanSize += info.Size()
				}

				return nil
			})

			bar := progressbar.NewOptions64(
				totalScanSize,
				progressbar.OptionSetDescription("Scanning files..."),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetWidth(10),
				progressbar.OptionThrottle(65*time.Millisecond),
				progressbar.OptionShowCount(),
				progressbar.OptionOnCompletion(func() {
					fmt.Fprint(os.Stderr, "\n")
				}),
				progressbar.OptionSpinnerType(14),
				progressbar.OptionFullWidth(),
				progressbar.OptionSetRenderBlankState(true))

			progressChan = make(chan int64)
			go func() {
				for incr := range progressChan {
					bar.Add64(incr)
				}
			}()
		}

		filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				fmt.Printf("Warning: Nil FileInfo for path: %s (err: %v)\n", path, err)

				return nil
			}

			if err != nil {
				fmt.Printf("Warning: Error accessing path %s: %v\n", path, err)
				return nil
			}

			wg.Add(1)
			go func(path string, info os.FileInfo) {
				// Acquire token from channel first
				taskCh <- Task{info: info}
				defer func() { <-taskCh }() // Release token when done
				defer wg.Done()

				if len(exclude) != 0 {
					for _, excludePattern := range exclude {
						if strings.Contains(filepath.ToSlash(path), excludePattern+"/") ||
							strings.HasPrefix(info.Name(), excludePattern) {
							fmt.Printf("Skipping excluded path: %s\n", path)
							return
						}
					}
				}

				if info.Size() > minSize && extMap[filepath.Ext(info.Name())] {
					mutex.Lock()
					files = append(files, struct {
						Name string
						Size int64
					}{path, info.Size()})
					toDeleteMap[path] = formatSize(info.Size())
					totalClearSize += info.Size()
					mutex.Unlock()
					if progress {
						progressChan <- info.Size()
					}
				}
			}(path, info)

			return nil
		})

		wg.Wait()

		if totalClearSize != 0 {
			printFilesTable(toDeleteMap)

			fmt.Println()
			fmt.Println(formatSize(totalClearSize), "will be cleared.")

			actionIsDelete := askForConfirmation("Delete these files?")

			if actionIsDelete {
				fmt.Println(color.New(color.FgGreen).SprintFunc()("✓"), "Deleted:", formatSize(totalClearSize))
				for _, file := range files {
					os.Remove(file.Name)
				}
				logDeletionToFile(toDeleteMap)
			}

		} else {
			red := color.New(color.FgRed).SprintFunc()
			fmt.Println(red("Error:"), "File not found")
		}
	}

}

func startTUI(dir string, extensions []string, minSize int64) error {
	app := tui.NewApp(dir, extensions, minSize)
	p := tea.NewProgram(app, tea.WithAltScreen())
	rules.SetupRulesConfig()
	_, err := p.Run()
	return err
}
