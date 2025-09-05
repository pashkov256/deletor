package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pashkov256/deletor/internal/cli/config"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/runner"
	"github.com/pashkov256/deletor/internal/validation"
)

func main() {
	// Initialize rules
	rules := rules.NewRules()
	rules.SetupRulesConfig()

	// Load CLI flags/config
	cfg := config.GetFlags()

	// Initialize validator & file manager
	validator := validation.NewValidator()
	fm := filemanager.NewFileManager()

	// Ensure resources are cleaned up if needed

	// defer fm.Close()

	if cfg.IsCLIMode {
		if err := runner.RunCLI(fm, rules, cfg); err != nil {
			log.Printf("CLI Error: %v\n", err)
			os.Exit(2)
		}
	} else {
		if err := runner.RunTUI(fm, rules, validator); err != nil {
			log.Printf("TUI Error: %v\n", err)
			os.Exit(3)
		}
	}
}
