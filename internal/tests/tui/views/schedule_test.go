package views_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/path"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/views"
	"github.com/pashkov256/deletor/internal/validation"
)

func setupScheduleRulesConfig(t *testing.T) func() {
	t.Helper()

	origAppDirName := path.AppDirName
	origRuleFileName := path.RuleFileName

	path.AppDirName = "deletor_schedule_view_test"
	path.RuleFileName = "rule_schedule_view_test.json"

	return func() {
		userConfigDir, _ := os.UserConfigDir()
		_ = os.RemoveAll(filepath.Join(userConfigDir, path.AppDirName))
		path.AppDirName = origAppDirName
		path.RuleFileName = origRuleFileName
	}
}

func setupScheduleModel(t *testing.T) (*views.ScheduleCleanModel, string) {
	t.Helper()

	cleanupConfig := setupScheduleRulesConfig(t)
	t.Cleanup(cleanupConfig)

	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "delete.txt"), []byte("delete"), 0644); err != nil {
		t.Fatalf("Failed to create delete.txt: %v", err)
	}

	ruleManager := rules.NewRules()
	if err := ruleManager.SetupRulesConfig(); err != nil {
		t.Fatalf("Failed to setup rules: %v", err)
	}
	if err := ruleManager.UpdateRules(
		rules.WithPath(tempDir),
		rules.WithExtensions([]string{".txt"}),
		rules.WithOptions(false, false, false, false, false, false, false, false, false, false),
	); err != nil {
		t.Fatalf("Failed to update rules: %v", err)
	}

	model := views.NewScheduleCleanModel(ruleManager, filemanager.NewFileManager(), validation.NewValidator())
	model.Init()
	return model, tempDir
}

func TestScheduleCleanModel_View(t *testing.T) {
	zone.NewGlobal()
	model, _ := setupScheduleModel(t)

	view := model.View()
	if !strings.Contains(view, "Schedule one-off clean") {
		t.Fatal("expected schedule page title in view")
	}
	if !strings.Contains(view, "Run in:") {
		t.Fatal("expected delay input in view")
	}
}

func TestScheduleCleanModel_SchedulesAndRunsOneOffClean(t *testing.T) {
	model, tempDir := setupScheduleModel(t)
	model.DelayInput.SetValue("1 sec")
	model.FocusedElement = "scheduleButton"

	newModel, cmd := model.Handle(tea.KeyMsg{Type: tea.KeyEnter})
	scheduleModel, ok := newModel.(*views.ScheduleCleanModel)
	if !ok {
		t.Fatal("failed to convert scheduled model")
	}
	if cmd == nil {
		t.Fatal("expected schedule command")
	}
	if !scheduleModel.IsScheduled() {
		t.Fatal("expected model to hold a scheduled run")
	}
	if scheduleModel.GetScheduledFor().IsZero() {
		t.Fatal("expected scheduled-for timestamp to be set")
	}

	triggeredModel, runCmd := scheduleModel.Update(views.ScheduledCleanTriggerMsg{ScheduleID: 1})
	scheduleModel = triggeredModel.(*views.ScheduleCleanModel)
	if !scheduleModel.IsRunning() {
		t.Fatal("expected model to enter running state")
	}
	if runCmd == nil {
		t.Fatal("expected cleanup command after trigger")
	}

	completedMsg := runCmd()
	completedModel, _ := scheduleModel.Update(completedMsg)
	scheduleModel = completedModel.(*views.ScheduleCleanModel)

	if scheduleModel.IsRunning() || scheduleModel.IsScheduled() {
		t.Fatal("expected scheduled run state to clear after completion")
	}
	if !strings.Contains(scheduleModel.GetStatus(), "Deleted 1 file(s)") {
		t.Fatalf("unexpected completion status: %q", scheduleModel.GetStatus())
	}
	if _, err := os.Stat(filepath.Join(tempDir, "delete.txt")); !os.IsNotExist(err) {
		t.Fatal("delete.txt should be removed after scheduled clean")
	}
}
