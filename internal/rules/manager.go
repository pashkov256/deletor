package rules

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pashkov256/deletor/internal/path"
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
		baseRules := `{
		       "Path": "",
            "Extensions": [],
            "Exclude": [],
            "MinSize": null,
            "MaxSize": null,
            "OlderThan": null,
            "NewerThan": null,
            "ShowHiddenFiles": false,
            "ConfirmDeletion": false,
            "IncludeSubfolders": false,
            "DeleteEmptySubfolders": false,
            "SendFilesToTrash": false,
            "LogOperations": false,
            "LogToFile": false,
            "ShowStatistics": false,
            "ExitAfterDeletion": false
		}`

		err := os.WriteFile(filePathRuleConfig, []byte(baseRules), 0644)

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
