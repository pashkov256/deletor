package cleanup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pashkov256/deletor/internal/cleanup"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/path"
	"github.com/pashkov256/deletor/internal/rules"
)

func setupCleanupRulesConfig(t *testing.T) func() {
	t.Helper()

	origAppDirName := path.AppDirName
	origRuleFileName := path.RuleFileName

	path.AppDirName = "deletor_schedule_test"
	path.RuleFileName = "rule_schedule_test.json"

	return func() {
		userConfigDir, _ := os.UserConfigDir()
		_ = os.RemoveAll(filepath.Join(userConfigDir, path.AppDirName))
		path.AppDirName = origAppDirName
		path.RuleFileName = origRuleFileName
	}
}

func TestRunOneOffClean_UsesSavedRulesRecursively(t *testing.T) {
	cleanupConfig := setupCleanupRulesConfig(t)
	defer cleanupConfig()

	rootDir := t.TempDir()
	nestedDir := filepath.Join(rootDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(rootDir, "keep.log"), []byte("keep"), 0644); err != nil {
		t.Fatalf("Failed to create keep.log: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "delete.txt"), []byte("delete"), 0644); err != nil {
		t.Fatalf("Failed to create delete.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatalf("Failed to create nested.txt: %v", err)
	}

	ruleManager := rules.NewRules()
	if err := ruleManager.SetupRulesConfig(); err != nil {
		t.Fatalf("Failed to setup rules: %v", err)
	}

	if err := ruleManager.UpdateRules(
		rules.WithPath(rootDir),
		rules.WithExtensions([]string{".txt"}),
		rules.WithOptions(false, false, true, true, false, false, false, false, false, false),
	); err != nil {
		t.Fatalf("Failed to update rules: %v", err)
	}

	spec, err := cleanup.LoadOneOffCleanSpec(ruleManager)
	if err != nil {
		t.Fatalf("LoadOneOffCleanSpec failed: %v", err)
	}

	result, err := cleanup.RunOneOffClean(filemanager.NewFileManager(), spec)
	if err != nil {
		t.Fatalf("RunOneOffClean failed: %v", err)
	}

	if result.FilesCleaned != 2 {
		t.Fatalf("FilesCleaned = %d, want 2", result.FilesCleaned)
	}
	if result.UsedTrash {
		t.Fatal("Expected permanent delete mode")
	}
	if result.EmptyDirsDeleted == 0 {
		t.Fatal("Expected at least one empty directory to be deleted")
	}

	if _, err := os.Stat(filepath.Join(rootDir, "delete.txt")); !os.IsNotExist(err) {
		t.Fatal("delete.txt should be removed")
	}
	if _, err := os.Stat(filepath.Join(nestedDir, "nested.txt")); !os.IsNotExist(err) {
		t.Fatal("nested.txt should be removed")
	}
	if _, err := os.Stat(filepath.Join(rootDir, "keep.log")); err != nil {
		t.Fatalf("keep.log should remain: %v", err)
	}
	if _, err := os.Stat(nestedDir); !os.IsNotExist(err) {
		t.Fatal("nested directory should be pruned after the clean")
	}
}

func TestRunOneOffClean_RespectsCurrentLevelOnly(t *testing.T) {
	cleanupConfig := setupCleanupRulesConfig(t)
	defer cleanupConfig()

	rootDir := t.TempDir()
	nestedDir := filepath.Join(rootDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(rootDir, "delete.txt"), []byte("delete"), 0644); err != nil {
		t.Fatalf("Failed to create delete.txt: %v", err)
	}
	nestedPath := filepath.Join(nestedDir, "nested.txt")
	if err := os.WriteFile(nestedPath, []byte("nested"), 0644); err != nil {
		t.Fatalf("Failed to create nested.txt: %v", err)
	}

	ruleManager := rules.NewRules()
	if err := ruleManager.SetupRulesConfig(); err != nil {
		t.Fatalf("Failed to setup rules: %v", err)
	}

	if err := ruleManager.UpdateRules(
		rules.WithPath(rootDir),
		rules.WithExtensions([]string{".txt"}),
		rules.WithOptions(false, false, false, false, false, false, false, false, false, false),
	); err != nil {
		t.Fatalf("Failed to update rules: %v", err)
	}

	spec, err := cleanup.LoadOneOffCleanSpec(ruleManager)
	if err != nil {
		t.Fatalf("LoadOneOffCleanSpec failed: %v", err)
	}

	result, err := cleanup.RunOneOffClean(filemanager.NewFileManager(), spec)
	if err != nil {
		t.Fatalf("RunOneOffClean failed: %v", err)
	}

	if result.FilesCleaned != 1 {
		t.Fatalf("FilesCleaned = %d, want 1", result.FilesCleaned)
	}
	if _, err := os.Stat(nestedPath); err != nil {
		t.Fatalf("nested file should remain when subfolders are disabled: %v", err)
	}
}
