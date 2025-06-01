package rules

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pashkov256/deletor/internal/path"
	"github.com/pashkov256/deletor/internal/tui/options"
)

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

func (d *defaultRules) SetupRulesConfig() error {
	filePathRuleConfig := d.GetRulesPath()
	os.MkdirAll(filepath.Dir(filePathRuleConfig), 0755)

	_, err := os.Stat(filePathRuleConfig)

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
			ExitAfterDeletion:     options.DefaultCleanOptionState[options.ExitAfterDeletion],
		}

		// Marshal the rules to JSON
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

func (d *defaultRules) GetRulesPath() string {
	userConfigDir, _ := os.UserConfigDir()
	filePathRuleConfig := filepath.Join(userConfigDir, path.AppDirName, path.RuleFileName)

	return filePathRuleConfig
}

func (d *defaultRules) Equals(other Rules) bool {
	if other == nil {
		return false
	}

	otherRules, err := other.GetRules()
	if err != nil {
		return false
	}

	if d.Path != otherRules.Path || d.MinSize != otherRules.MinSize {
		return false
	}

	if len(d.Extensions) != len(otherRules.Extensions) {
		return false
	}

	if len(d.Exclude) != len(otherRules.Exclude) {
		return false
	}

	for i := range d.Extensions {
		if d.Extensions[i] != otherRules.Extensions[i] {
			return false
		}
	}

	for i := range d.Exclude {
		if d.Exclude[i] != otherRules.Exclude[i] {
			return false
		}
	}

	return true
}
