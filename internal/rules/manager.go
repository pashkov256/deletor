package rules

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pashkov256/deletor/internal/path"
	"github.com/pashkov256/deletor/internal/tui/options"
	"github.com/pashkov256/deletor/internal/utils"
)

func defaultRuleValues() *defaultRules {
	return &defaultRules{
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
}

func (d *defaultRules) clone() *defaultRules {
	if d == nil {
		return defaultRuleValues()
	}

	clone := *d
	clone.Extensions = append([]string(nil), d.Extensions...)
	clone.Exclude = append([]string(nil), d.Exclude...)
	clone.cached = nil
	clone.mu = sync.RWMutex{}
	return &clone
}

func (d *defaultRules) getRulesPath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}
	return filepath.Join(userConfigDir, path.AppDirName, path.RuleFileName), nil
}

func (d *defaultRules) setCache(rules *defaultRules) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cached = rules.clone()
}

func (d *defaultRules) getCached() *defaultRules {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.cached == nil {
		return nil
	}
	return d.cached.clone()
}

func (d *defaultRules) validate() error {
	if d.MinSize != "" {
		if _, err := utils.ToBytes(d.MinSize); err != nil {
			return fmt.Errorf("invalid MinSize: %w", err)
		}
	}
	if d.MaxSize != "" {
		if _, err := utils.ToBytes(d.MaxSize); err != nil {
			return fmt.Errorf("invalid MaxSize: %w", err)
		}
	}
	if d.OlderThan != "" {
		if _, err := utils.ParseTimeDuration(d.OlderThan); err != nil {
			return fmt.Errorf("invalid OlderThan: %w", err)
		}
	}
	if d.NewerThan != "" {
		if _, err := utils.ParseTimeDuration(d.NewerThan); err != nil {
			return fmt.Errorf("invalid NewerThan: %w", err)
		}
	}

	d.Extensions = append([]string(nil), d.Extensions...)
	d.Exclude = append([]string(nil), d.Exclude...)

	return nil
}

func (d *defaultRules) readRulesFromDisk() (*defaultRules, error) {
	filePathRuleConfig, err := d.getRulesPath()
	if err != nil {
		return nil, err
	}

	jsonRules, err := os.ReadFile(filePathRuleConfig)
	if err != nil {
		return nil, err
	}

	rules := defaultRuleValues()
	if err := json.Unmarshal(jsonRules, rules); err != nil {
		return nil, err
	}
	if err := rules.validate(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (d *defaultRules) readRulesForUpdate() (*defaultRules, error) {
	rules, err := d.readRulesFromDisk()
	if err == nil {
		return rules, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return defaultRuleValues(), nil
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) {
		return defaultRuleValues(), nil
	}

	return nil, err
}

func (d *defaultRules) writeRules(rules *defaultRules) error {
	filePathRuleConfig, err := d.getRulesPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(filePathRuleConfig), 0755); err != nil {
		return err
	}

	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePathRuleConfig, rulesJSON, 0644); err != nil {
		return err
	}

	d.setCache(rules)
	return nil
}

// UpdateRules applies the provided options and saves the updated rules to file.
func (d *defaultRules) UpdateRules(options ...RuleOption) error {
	currentRules, err := d.readRulesForUpdate()
	if err != nil {
		return err
	}

	for _, option := range options {
		option(currentRules)
	}

	if err := currentRules.validate(); err != nil {
		return err
	}

	return d.writeRules(currentRules)
}

// GetRules loads and returns the current rules from the configuration file.
func (d *defaultRules) GetRules() (*defaultRules, error) {
	if cached := d.getCached(); cached != nil {
		return cached, nil
	}

	rules, err := d.readRulesFromDisk()
	if err != nil {
		return defaultRuleValues(), err
	}

	d.setCache(rules)
	return rules.clone(), nil
}

// SetupRulesConfig initializes the rules configuration file with default values.
func (d *defaultRules) SetupRulesConfig() error {
	filePathRuleConfig, err := d.getRulesPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(filePathRuleConfig), 0755); err != nil {
		return err
	}

	_, err = os.Stat(filePathRuleConfig)
	if err == nil {
		_, readErr := d.readRulesFromDisk()
		return readErr
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return d.writeRules(defaultRuleValues())
}

// GetRulesPath returns the path to the rules configuration file.
func (d *defaultRules) GetRulesPath() string {
	filePathRuleConfig, err := d.getRulesPath()
	if err != nil {
		return filepath.Join(path.AppDirName, path.RuleFileName)
	}
	return filePathRuleConfig
}
