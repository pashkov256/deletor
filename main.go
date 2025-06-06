package main

import (
	"fmt"
	"os"

	"github.com/pashkov256/deletor/internal/cli/config"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/runner"
)

func main() {
	var rules = rules.NewRules()
	rules.SetupRulesConfig()
	config := config.GetFlags()
	fm := filemanager.NewFileManager()

	if config.IsCLIMode {
		runner.RunCLI(fm, rules, config)
	} else {
		if err := runner.RunTUI(fm, rules); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}
}
