package rules

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pashkov256/deletor/internal/path"
	"github.com/pashkov256/deletor/internal/tui/options"
)

// UpdateRules applies the provided options and saves the updated rules to file
func (d *defaultRules) UpdateRules(options ...RuleOption) error {
	// Update the struct fields
	for _, option := range options {
		option(d)
	}

	// Marshal and save to file
	rulesJSON, err := json.Marshal(d)
	if err != nil {
		return err
	}

	err = os.WriteFile(d.GetRulesPath(), rulesJSON, 0644)
	if err != nil {
		return err
	}
	return nil
}

// GetRules loads and returns the current rules from the configuration file
func (d *defaultRules) GetRules() (*defaultRules, error) {
	jsonRules, err := os.ReadFile(d.GetRulesPath())
	if err != nil {
		return nil, err
	}

	// Create a new instance to avoid modifying the receiver
	rules := &defaultRules{}
	err = json.Unmarshal(jsonRules, rules)
	if err != nil {
		return nil, err
	}

	return rules, nil
}

// SetupRulesConfig initializes the rules configuration file with default values
func (d *defaultRules) SetupRulesConfig() error {
	filePathRuleConfig := d.GetRulesPath()

	err := os.MkdirAll(filepath.Dir(filePathRuleConfig), 0755)
	if err != nil {
		return err
	}

	_, err = os.Stat(filePathRuleConfig)

	if err != nil {
		// Create a new defaultRules instance with values from DefaultCleanOptionState
		rules := &defaultRules{
			Path:                  "",
			Extensions:            []string{},
			Exclude:               []string{},
			MinSize:               "",
			MaxSize:               "",
			OlderThan:             "",
			NewerThan:             "",
			ShowHiddenFiles:       options.DefaultCleanOptionState[options.ShowHiddenFiles],
			ConfirmDeletion:       options.DefaultCleanOptionState[options.ConfirmDeletion],
			IncludeSubfolders:     options.DefaultCleanOptionState[options.IncludeSubfolders],
			DeleteEmptySubfolders: options.DefaultCleanOptionState[options.DeleteEmptySubfolders],
			SendFilesToTrash:      options.DefaultCleanOptionState[options.SendFilesToTrash],
			LogOperations:         options.DefaultCleanOptionState[options.LogOperations],
			LogToFile:             options.DefaultCleanOptionState[options.LogToFile],
			ShowStatistics:        options.DefaultCleanOptionState[options.ShowStatistics],
			DisableEmoji:          options.DefaultCleanOptionState[options.DisableEmoji],
			ExitAfterDeletion:     options.DefaultCleanOptionState[options.ExitAfterDeletion],
		}

		rulesJSON, err := json.Marshal(rules)
		if err != nil {
			return err
		}

		err = os.WriteFile(filePathRuleConfig, rulesJSON, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetRulesPath returns the path to the rules configuration file
func (d *defaultRules) GetRulesPath() string {
	userConfigDir, _ := os.UserConfigDir()
	filePathRuleConfig := filepath.Join(userConfigDir, path.AppDirName, path.RuleFileName)

	return filePathRuleConfig
}
