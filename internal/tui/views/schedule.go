package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/pashkov256/deletor/internal/cleanup"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/tui/errors"
	"github.com/pashkov256/deletor/internal/tui/help"
	"github.com/pashkov256/deletor/internal/tui/styles"
	"github.com/pashkov256/deletor/internal/utils"
	"github.com/pashkov256/deletor/internal/validation"
)

// ScheduledCleanTriggerMsg fires when a scheduled cleanup timer elapses.
type ScheduledCleanTriggerMsg struct {
	ScheduleID int
}

// ScheduledCleanCompletedMsg reports the result of a scheduled cleanup run.
type ScheduledCleanCompletedMsg struct {
	ScheduleID int
	Result     *cleanup.OneOffCleanResult
	Err        error
}

// ScheduleCleanModel manages the one-off scheduled clean page.
type ScheduleCleanModel struct {
	DelayInput       textinput.Model
	FocusedElement   string
	rules            rules.Rules
	filemanager      filemanager.FileManager
	validator        *validation.Validator
	pendingSpec      *cleanup.OneOffCleanSpec
	scheduledFor     time.Time
	activeScheduleID int
	isScheduled      bool
	isRunning        bool
	status           string
	Error            *errors.Error
}

func NewScheduleCleanModel(ruleManager rules.Rules, fm filemanager.FileManager, validator *validation.Validator) *ScheduleCleanModel {
	delayInput := textinput.New()
	delayInput.Placeholder = "Run in (e.g. 10 min, 1 hour, 1 day)"
	delayInput.PromptStyle = styles.TextInputPromptStyle
	delayInput.TextStyle = styles.TextInputTextStyle
	delayInput.Cursor.Style = styles.TextInputCursorStyle

	return &ScheduleCleanModel{
		DelayInput:     delayInput,
		FocusedElement: "delayInput",
		rules:          ruleManager,
		filemanager:    fm,
		validator:      validator,
	}
}

func (m *ScheduleCleanModel) Init() tea.Cmd {
	m.DelayInput.Focus()
	return textinput.Blink
}

func (m *ScheduleCleanModel) View() string {
	var content strings.Builder

	content.WriteString(styles.TitleStyle.Render("Schedule one-off clean"))
	content.WriteString("\n\n")
	content.WriteString("This schedules a single future cleanup run using your saved rules.\n")
	content.WriteString("Deletor must stay open until the scheduled time.\n\n")

	spec, specErr := cleanup.LoadOneOffCleanSpec(m.rules)
	if specErr != nil {
		content.WriteString(styles.InfoStyle.Render(specErr.Error()))
		content.WriteString("\n\n")
	} else {
		extensions := "all files"
		if len(spec.Extensions) > 0 {
			extensions = strings.Join(spec.Extensions, ", ")
		}

		scope := "current directory only"
		if spec.IncludeSubfolders {
			scope = "current directory and subfolders"
		}

		action := "delete permanently"
		if spec.SendFilesToTrash {
			action = "move files to trash"
		}

		content.WriteString(fmt.Sprintf("Saved path: %s\n", spec.Path))
		content.WriteString(fmt.Sprintf("Extensions: %s\n", extensions))
		content.WriteString(fmt.Sprintf("Scope: %s\n", scope))
		content.WriteString(fmt.Sprintf("Action: %s\n", action))
		if spec.DeleteEmptySubfolders {
			content.WriteString("Empty directories will be pruned after the run.\n")
		}
		content.WriteString("\n")
	}

	delayStyle := styles.StandardInputStyle
	if m.FocusedElement == "delayInput" {
		delayStyle = styles.StandardInputFocusedStyle
	}
	content.WriteString(zone.Mark("schedule_delay_input", delayStyle.Render("Run in: "+m.DelayInput.View())))
	content.WriteString("\n\n")

	buttonLabel := "Schedule one-off clean"
	buttonStyle := styles.StandardButtonStyle
	if m.FocusedElement == "scheduleButton" {
		buttonStyle = styles.StandardButtonFocusedStyle
	}
	if m.isRunning {
		buttonLabel = "Running scheduled clean..."
		buttonStyle = styles.LaunchButtonFocusedStyle
	}
	content.WriteString(zone.Mark("schedule_button", buttonStyle.Render(buttonLabel)))

	if m.isScheduled {
		content.WriteString("\n\n")
		content.WriteString(styles.InfoStyle.Render(
			fmt.Sprintf("Scheduled for %s", m.scheduledFor.Format("2006-01-02 15:04:05")),
		))
	}

	if m.Error != nil && m.Error.IsVisible() {
		content.WriteString("\n\n")
		content.WriteString(errors.GetStyle(m.Error.GetType()).Render(m.Error.GetMessage()))
	} else if m.status != "" {
		content.WriteString("\n\n")
		content.WriteString(styles.SuccessStyle.Render(m.status))
	}

	content.WriteString("\n\n")
	content.WriteString(help.NavigateHelpText)

	return zone.Scan(content.String())
}

func (m *ScheduleCleanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.Handle(msg)
	case tea.MouseMsg:
		// nolint:staticcheck
		if msg.Type == tea.MouseLeft && msg.Action == tea.MouseActionPress {
			if zone.Get("schedule_delay_input").InBounds(msg) {
				m.FocusedElement = "delayInput"
				m.DelayInput.Focus()
				return m, nil
			}
			if zone.Get("schedule_button").InBounds(msg) {
				m.FocusedElement = "scheduleButton"
				m.DelayInput.Blur()
				return m.scheduleOnce()
			}
		}
	case ScheduledCleanTriggerMsg:
		if msg.ScheduleID != m.activeScheduleID || m.pendingSpec == nil {
			return m, nil
		}
		m.isScheduled = false
		m.isRunning = true
		m.Error = nil
		m.status = "Running scheduled clean..."
		return m, m.runScheduledClean(msg.ScheduleID)
	case ScheduledCleanCompletedMsg:
		if msg.ScheduleID != m.activeScheduleID {
			return m, nil
		}

		m.isRunning = false
		m.pendingSpec = nil
		m.activeScheduleID = 0
		m.scheduledFor = time.Time{}

		if msg.Err != nil {
			m.status = ""
			m.Error = errors.New(errors.ErrorTypeFileSystem, fmt.Sprintf("Scheduled clean failed: %v", msg.Err))
			return m, nil
		}

		m.Error = nil
		if msg.Result == nil {
			m.status = "Scheduled clean completed."
			return m, nil
		}

		action := "Deleted"
		if msg.Result.UsedTrash {
			action = "Moved to trash"
		}

		status := fmt.Sprintf(
			"%s %d file(s), cleared %s.",
			action,
			msg.Result.FilesCleaned,
			utils.FormatSize(msg.Result.BytesCleared),
		)
		if msg.Result.EmptyDirsDeleted > 0 {
			status += fmt.Sprintf(" Removed %d empty directorie(s).", msg.Result.EmptyDirsDeleted)
		}
		m.status = status
		return m, nil
	}

	if m.FocusedElement == "delayInput" {
		var cmd tea.Cmd
		m.DelayInput, cmd = m.DelayInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *ScheduleCleanModel) Handle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		return m.focusNext()
	case "shift+tab", "up":
		return m.focusPrevious()
	case "enter":
		if m.FocusedElement == "scheduleButton" {
			return m.scheduleOnce()
		}
	case "alt+c":
		m.DelayInput.SetValue("")
		if !m.isScheduled && !m.isRunning {
			m.status = ""
		}
		m.Error = nil
		return m, nil
	}

	if m.FocusedElement == "delayInput" {
		var cmd tea.Cmd
		m.DelayInput, cmd = m.DelayInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *ScheduleCleanModel) focusNext() (tea.Model, tea.Cmd) {
	if m.FocusedElement == "delayInput" {
		m.FocusedElement = "scheduleButton"
		m.DelayInput.Blur()
	} else {
		m.FocusedElement = "delayInput"
		m.DelayInput.Focus()
	}
	return m, nil
}

func (m *ScheduleCleanModel) focusPrevious() (tea.Model, tea.Cmd) {
	return m.focusNext()
}

func (m *ScheduleCleanModel) scheduleOnce() (tea.Model, tea.Cmd) {
	if err := m.validator.ValidateTimeDuration(m.DelayInput.Value()); err != nil {
		m.status = ""
		return m, func() tea.Msg {
			return errors.New(errors.ErrorTypeValidation, "Invalid schedule delay format")
		}
	}

	spec, err := cleanup.LoadOneOffCleanSpec(m.rules)
	if err != nil {
		m.status = ""
		return m, func() tea.Msg {
			return errors.New(errors.ErrorTypeValidation, err.Error())
		}
	}

	parsedTime, err := utils.ParseTimeDuration(m.DelayInput.Value())
	if err != nil {
		m.status = ""
		return m, func() tea.Msg {
			return errors.New(errors.ErrorTypeValidation, fmt.Sprintf("Invalid schedule delay: %v", err))
		}
	}

	delay := time.Since(parsedTime)
	if delay <= 0 {
		m.status = ""
		return m, func() tea.Msg {
			return errors.New(errors.ErrorTypeValidation, "Schedule delay must be greater than zero")
		}
	}

	m.activeScheduleID++
	m.pendingSpec = spec
	m.scheduledFor = time.Now().Add(delay)
	m.isScheduled = true
	m.isRunning = false
	m.Error = nil
	m.status = fmt.Sprintf("One-off clean scheduled for %s", m.scheduledFor.Format("2006-01-02 15:04:05"))

	scheduleID := m.activeScheduleID
	return m, tea.Tick(delay, func(time.Time) tea.Msg {
		return ScheduledCleanTriggerMsg{ScheduleID: scheduleID}
	})
}

func (m *ScheduleCleanModel) runScheduledClean(scheduleID int) tea.Cmd {
	spec := m.pendingSpec
	return func() tea.Msg {
		result, err := cleanup.RunOneOffClean(m.filemanager, spec)
		return ScheduledCleanCompletedMsg{
			ScheduleID: scheduleID,
			Result:     result,
			Err:        err,
		}
	}
}

func (m *ScheduleCleanModel) IsScheduled() bool {
	return m.isScheduled
}

func (m *ScheduleCleanModel) IsRunning() bool {
	return m.isRunning
}

func (m *ScheduleCleanModel) GetStatus() string {
	return m.status
}

func (m *ScheduleCleanModel) GetScheduledFor() time.Time {
	return m.scheduledFor
}
