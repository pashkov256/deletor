package rules_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pashkov256/deletor/internal/path"
	"github.com/pashkov256/deletor/internal/rules"
)

// Setup a temporary directory to store test config file
func setupTempConfigDir() func() {
	origAppDirName := path.AppDirName
	origRuleFileName := path.RuleFileName

	path.AppDirName = "deletor_temp"
	path.RuleFileName = "rule_temp.json"

	userConfigDirTemp, _ := os.UserConfigDir()
	filePathRuleConfigTemp := filepath.Join(userConfigDirTemp, path.AppDirName, path.RuleFileName)

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(filePathRuleConfigTemp)
		path.AppDirName = origAppDirName
		path.RuleFileName = origRuleFileName
	}

	return cleanup
}

func TestNewRules(t *testing.T) {
	// Create a new rules instance
	rs := rules.NewRules()

	// Verify that the returned instance is not nil
	if rs == nil {
		t.Error("rules.NewRules() should return a non-nil instance")
	}
}

func TestGetRulesPath(t *testing.T) {
	// Create a new rules instance
	rs := rules.NewRules()

	// Expected path
	userConfigDir, _ := os.UserConfigDir()
	expectedPath := filepath.Join(userConfigDir, path.AppDirName, path.RuleFileName)

	// Verify GetRulesPath returns the correct path
	actualPath := rs.GetRulesPath()
	if actualPath != expectedPath {
		t.Errorf("GetRulesPath() = %v, want %v", actualPath, expectedPath)
	}
}

func TestSetupRulesConfig_NoExistingConfig(t *testing.T) {
	// Setup temporary test directory
	cleanup := setupTempConfigDir()
	defer cleanup()

	// Create rules instance
	rs := rules.NewRules()

	// Get config directory path
	configPath := rs.GetRulesPath()
	configDir := filepath.Dir(configPath)

	// Run setup
	err := rs.SetupRulesConfig()
	if err != nil {
		t.Fatalf("SetupRulesConfig() failed: %v", err)
	}

	// Verify directory was created
	if info, err := os.Stat(configDir); err != nil || !info.IsDir() {
		t.Errorf("Config directory was not created properly")
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Config file was not created properly: %v", err)
	}

	// Check if file has 0644 permissions
	expectedPerm := os.FileMode(0644)
	if runtime.GOOS != "windows" && info.Mode().Perm() != expectedPerm {
		t.Errorf("File permissions = %v, want %v", info.Mode().Perm(), expectedPerm)
	}

	// Read the created config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Parse the config file
	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Failed to parse config file: %v", err)
	}

	// Test boolean fields
	boolFields := map[string]bool{
		"ShowStatistics":        true,
		"ShowHiddenFiles":       false,
		"ConfirmDeletion":       false,
		"IncludeSubfolders":     false,
		"DeleteEmptySubfolders": false,
		"SendFilesToTrash":      false,
		"LogOperations":         false,
		"LogToFile":             false,
		"ExitAfterDeletion":     false,
	}

	for field, expectedValue := range boolFields {
		configValue, exists := config[field]
		if !exists {
			if expectedValue {
				t.Errorf("%s field missing but should be present with value %v", field, expectedValue)
			}
			continue
		}

		boolValue, ok := configValue.(bool)
		if !ok {
			t.Errorf("%s field is not a boolean", field)
			continue
		}

		if boolValue != expectedValue {
			t.Errorf("%s = %v, want %v", field, boolValue, expectedValue)
		}
	}

	// Test empty fields in the default config - these should be omitted due to being empty
	stringFields := []string{"Path", "MinSize", "MaxSize", "OlderThan", "NewerThan", "Extensions", "Exclude"}
	for _, field := range stringFields {
		if value, exists := config[field]; exists {
			t.Errorf("%s should be omitted when empty, but got %v", field, value)
		}
	}

}

func TestSetupRulesConfig_ExistingConfig(t *testing.T) {
	// Setup temporary test directory
	cleanup := setupTempConfigDir()
	defer cleanup()

	// Create rules instance
	rs := rules.NewRules()

	// Create existing config with custom values
	configPath := rs.GetRulesPath()
	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	existingConfig := map[string]interface{}{
		"Test1": "Test1_Value",
		"Test2": []string{"Test2_Value"},
	}
	data, _ := json.Marshal(existingConfig)
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write existing config: %v", err)
	}

	// Run setup
	err = rs.SetupRulesConfig()
	if err != nil {
		t.Fatalf("SetupRulesConfig() failed: %v", err)
	}

	// Verify existing config was not overwritten
	configData, _ := os.ReadFile(configPath)
	var loadedConfig map[string]interface{}
	err = json.Unmarshal(configData, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to parse existing config: %v", err)
	}

	test1, ok := loadedConfig["Test1"].(string)
	if !ok {
		t.Error("Test1 field not found or not a string")
	} else if test1 != "Test1_Value" {
		t.Errorf("Test1 = %q, want Test1_Value", test1)
	}

	test2, ok := loadedConfig["Test2"].([]interface{})
	if !ok {
		t.Error("Test2 field not found or not an array")
	} else if len(test2) != 1 || test2[0] != "Test2_Value" {
		t.Errorf("Extensions = %v, want {Test2_Value}", test2...)
	}
}

func TestUpdateRules_SingleOption(t *testing.T) {
	// Setup temporary test directory
	cleanup := setupTempConfigDir()
	defer cleanup()

	// Create rules instance and setup default config
	rs := rules.NewRules()
	err := rs.SetupRulesConfig()
	if err != nil {
		t.Fatalf("Failed to setup default config: %v", err)
	}

	// Update with single option
	testPath := "/test/path"
	err = rs.UpdateRules(rules.WithPath(testPath))
	if err != nil {
		t.Fatalf("UpdateRules failed: %v", err)
	}

	// Verify the update
	data, err := os.ReadFile(rs.GetRulesPath())
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if path, ok := config["Path"].(string); !ok || path != testPath {
		t.Errorf("Path = %v, want %v", config["Path"], testPath)
	}
}

func TestUpdateRules_MultipleOptions(t *testing.T) {
	// Setup temporary test directory
	cleanup := setupTempConfigDir()
	defer cleanup()

	// Create rules instance and setup default config
	rs := rules.NewRules()
	err := rs.SetupRulesConfig()
	if err != nil {
		t.Fatalf("Failed to setup default config: %v", err)
	}

	// Update with multiple options
	testPath := "/test/path"
	testExtensions := []string{".txt", ".md"}
	testExclude := []string{"*.log", "*.tmp"}
	err = rs.UpdateRules(
		rules.WithPath(testPath),
		rules.WithMinSize("1GB"),
		rules.WithMaxSize("10GB"),
		rules.WithOlderThan("10d"),
		rules.WithNewerThan("10d"),
		rules.WithExtensions(testExtensions),
		rules.WithOptions(true, true, true, false, false, true, false, true, false),
		rules.WithExclude(testExclude),
	)
	if err != nil {
		t.Fatalf("UpdateRules failed: %v", err)
	}

	// Verify the updates
	data, err := os.ReadFile(rs.GetRulesPath())
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Check path
	if path, ok := config["Path"].(string); !ok || path != testPath {
		t.Errorf("Path = %v, want %v", config["Path"], testPath)
	}

	// Check extensions
	if ext, ok := config["Extensions"].([]interface{}); !ok {
		t.Error("Extensions field not found or not an array")
	} else {
		if len(ext) != len(testExtensions) {
			t.Errorf("Extensions length = %d, want %d", len(ext), len(testExtensions))
		}
		for i, v := range ext {
			if v != testExtensions[i] {
				t.Errorf("Extensions[%d] = %v, want %v", i, v, testExtensions[i])
			}
		}
	}

	// Check exclude
	if exclude, ok := config["Exclude"].([]interface{}); !ok {
		t.Error("Exclude field not found or not an array")
	} else {
		if len(exclude) != len(testExclude) {
			t.Errorf("Exclude length = %d, want %d", len(exclude), len(testExclude))
		}
		for i, v := range exclude {
			if v != testExclude[i] {
				t.Errorf("Exclude[%d] = %v, want %v", i, v, testExclude[i])
			}
		}
	}

	// Check boolean options
	expectedTrueFields := []string{
		"ShowHiddenFiles",
		"ConfirmDeletion",
		"IncludeSubfolders",
		"LogOperations",
		"ShowStatistics",
	}

	// Verify true fields are present and set to true
	for _, field := range expectedTrueFields {
		if value, ok := config[field].(bool); !ok || !value {
			t.Errorf("%s = %v, want true", field, value)
		}
	}

	// Verify false fields are omitted
	falseFields := []string{
		"DeleteEmptySubfolders",
		"SendFilesToTrash",
		"LogToFile",
		"ExitAfterDeletion",
	}

	for _, field := range falseFields {
		if _, exists := config[field]; exists {
			t.Errorf("%s should be omitted when false, but it exists in config", field)
		}
	}

	// Check string fields
	stringFields := map[string]string{
		"MinSize":   "1GB",
		"MaxSize":   "10GB",
		"OlderThan": "10d",
		"NewerThan": "10d",
	}
	for field, expectedValue := range stringFields {
		if value, ok := config[field].(string); !ok || value != expectedValue {
			t.Errorf("%s = %v, want %v", field, value, expectedValue)
		}
	}

}

func TestUpdateRules_InvalidJSON(t *testing.T) {
	// Setup temporary test directory
	cleanup := setupTempConfigDir()
	defer cleanup()

	// Create rules instance and setup default config
	rs := rules.NewRules()
	err := rs.SetupRulesConfig()
	if err != nil {
		t.Fatalf("Failed to setup default config: %v", err)
	}

	// Create config with invalid JSON structure
	configPath := rs.GetRulesPath()
	invalidJSON := []byte("{invalid json")
	err = os.WriteFile(configPath, invalidJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Attempt to update rules - this should succeed since UpdateRules doesn't read the existing file
	err = rs.UpdateRules(rules.WithPath("/new/path"))
	if err != nil {
		t.Errorf("UpdateRules failed but should have succeeded: %v", err)
	}

	// Verify the file was updated with valid JSON
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read updated config: %v", err)
	}

	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		t.Errorf("Updated config contains invalid JSON: %v", err)
	}

	// Verify the path was updated
	if path, ok := config["Path"].(string); !ok || path != "/new/path" {
		t.Errorf("Path = %v, want %v", config["Path"], "/new/path")
	}
}

func TestGetRules_MissingFile(t *testing.T) {
	// Setup temporary test directory
	cleanup := setupTempConfigDir()
	defer cleanup()

	// Create rules instance and setup default config
	rs := rules.NewRules()
	err := rs.SetupRulesConfig()
	if err != nil {
		t.Fatalf("Failed to setup default config: %v", err)
	}

	// Get config path and ensure it doesn't exist
	configPath := rs.GetRulesPath()
	os.RemoveAll(filepath.Dir(configPath))

	// Try to get rules from the file that doesn't exist
	_, err = rs.GetRules()
	if err == nil {
		t.Error("Expected error when reading non-existent file, got nil")
	}
}

func TestGetRules_InvalidJSON(t *testing.T) {
	// Setup temporary test directory
	cleanup := setupTempConfigDir()
	defer cleanup()

	// Create rules instance
	rs := rules.NewRules()

	// Create directory and write invalid JSON
	configPath := rs.GetRulesPath()
	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	invalidJSON := []byte("{invalid json")
	err = os.WriteFile(configPath, invalidJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Try to get rules from file with invalid JSON
	_, err = rs.GetRules()
	if err == nil {
		t.Error("Expected error when parsing invalid JSON, got nil")
	}
}

func TestGetRules_Success(t *testing.T) {
	// Setup temporary test directory
	cleanup := setupTempConfigDir()
	defer cleanup()

	// Create rules instance
	rs := rules.NewRules()

	// Create test config
	configPath := rs.GetRulesPath()
	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	testConfig := map[string]interface{}{
		"Path":            "/test/path",
		"Extensions":      []string{".txt", ".md"},
		"ShowHiddenFiles": true,
		"ShowStatistics":  true,
	}

	configJSON, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	err = os.WriteFile(configPath, configJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Get rules and verify content
	rules, err := rs.GetRules()
	if err != nil {
		t.Fatalf("GetRules failed: %v", err)
	}

	if rules == nil {
		t.Fatal("GetRules returned empty")
	}

	if rules.Path != "/test/path" {
		t.Errorf("Path = %q, want %q", rules.Path, "/test/path")
	}

	if len(rules.Extensions) != 2 || rules.Extensions[0] != ".txt" || rules.Extensions[1] != ".md" {
		t.Errorf("Extensions = %v, want [.txt .md]", rules.Extensions)
	}

	if !rules.ShowHiddenFiles {
		t.Error("ShowHiddenFiles = false, want true")
	}

	if !rules.ShowStatistics {
		t.Error("ShowStatistics = false, want true")
	}
}
