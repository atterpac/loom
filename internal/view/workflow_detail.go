package view

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/atterpac/loom/internal/config"
	"github.com/atterpac/loom/internal/temporal"
	"github.com/atterpac/loom/internal/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// WorkflowDetail displays detailed information about a workflow with events.
type WorkflowDetail struct {
	*tview.Flex
	app              *App
	workflowID       string
	runID            string
	workflow         *temporal.Workflow
	events           []temporal.HistoryEvent
	leftFlex         *tview.Flex
	workflowPanel    *ui.Panel
	eventDetailPanel *ui.Panel
	eventsPanel      *ui.Panel
	workflowView     *tview.TextView
	eventDetailView  *tview.TextView
	eventTable       *ui.Table
	loading          bool
	unsubscribeTheme func()
}

// NewWorkflowDetail creates a new workflow detail view.
func NewWorkflowDetail(app *App, workflowID, runID string) *WorkflowDetail {
	wd := &WorkflowDetail{
		Flex:       tview.NewFlex().SetDirection(tview.FlexColumn),
		app:        app,
		workflowID: workflowID,
		runID:      runID,
		eventTable: ui.NewTable(),
	}
	wd.setup()
	return wd
}

func (wd *WorkflowDetail) setup() {
	wd.SetBackgroundColor(ui.ColorBg())

	// Combined workflow info view
	wd.workflowView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	wd.workflowView.SetBackgroundColor(ui.ColorBg())

	// Event detail view
	wd.eventDetailView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	wd.eventDetailView.SetBackgroundColor(ui.ColorBg())

	// Event table
	wd.eventTable.SetHeaders("ID", "TIME", "TYPE")
	wd.eventTable.SetBorder(false)
	wd.eventTable.SetBackgroundColor(ui.ColorBg())

	// Create panels
	wd.workflowPanel = ui.NewPanel("Workflow")
	wd.workflowPanel.SetContent(wd.workflowView)

	wd.eventDetailPanel = ui.NewPanel("Event Detail")
	wd.eventDetailPanel.SetContent(wd.eventDetailView)

	wd.eventsPanel = ui.NewPanel("Events")
	wd.eventsPanel.SetContent(wd.eventTable)

	// Left side: workflow info + event detail stacked
	wd.leftFlex = tview.NewFlex().SetDirection(tview.FlexRow)
	wd.leftFlex.SetBackgroundColor(ui.ColorBg())
	wd.leftFlex.AddItem(wd.workflowPanel, 0, 1, false)
	wd.leftFlex.AddItem(wd.eventDetailPanel, 0, 1, false)

	// Main layout: left stack + right events
	wd.AddItem(wd.leftFlex, 0, 2, false)
	wd.AddItem(wd.eventsPanel, 0, 3, true)

	// Update event detail when selection changes
	wd.eventTable.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row-1 < len(wd.events) {
			wd.updateEventDetail(wd.events[row-1])
		}
	})

	// Show loading state initially
	wd.workflowView.SetText(fmt.Sprintf("\n [%s]Loading...[-]", ui.TagFgDim()))

	// Register for theme changes
	wd.unsubscribeTheme = ui.OnThemeChange(func(_ *config.ParsedTheme) {
		wd.SetBackgroundColor(ui.ColorBg())
		wd.leftFlex.SetBackgroundColor(ui.ColorBg())
		wd.workflowView.SetBackgroundColor(ui.ColorBg())
		wd.eventDetailView.SetBackgroundColor(ui.ColorBg())
		// Re-render with new colors
		if wd.workflow != nil {
			wd.render()
		}
		if len(wd.events) > 0 {
			wd.populateEventTable()
		}
	})
}

func (wd *WorkflowDetail) setLoading(loading bool) {
	wd.loading = loading
}

func (wd *WorkflowDetail) loadData() {
	provider := wd.app.Provider()
	if provider == nil {
		wd.loadMockData()
		return
	}

	wd.setLoading(true)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		workflow, err := provider.GetWorkflow(ctx, wd.app.CurrentNamespace(), wd.workflowID, wd.runID)

		wd.app.UI().QueueUpdateDraw(func() {
			wd.setLoading(false)
			if err != nil {
				wd.showError(err)
				return
			}
			wd.workflow = workflow
			wd.render()
			// Update hints now that we have workflow status
			wd.app.UI().Menu().SetHints(wd.Hints())
		})
	}()

	// Load events in parallel
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		events, err := provider.GetWorkflowHistory(ctx, wd.app.CurrentNamespace(), wd.workflowID, wd.runID)

		wd.app.UI().QueueUpdateDraw(func() {
			if err != nil {
				return
			}
			wd.events = events
			wd.populateEventTable()
		})
	}()
}

func (wd *WorkflowDetail) loadMockData() {
	now := time.Now()
	wd.workflow = &temporal.Workflow{
		ID:        wd.workflowID,
		RunID:     wd.runID,
		Type:      "MockWorkflow",
		Status:    "Running",
		Namespace: wd.app.CurrentNamespace(),
		TaskQueue: "mock-tasks",
		StartTime: now.Add(-5 * time.Minute),
	}
	wd.events = []temporal.HistoryEvent{
		{ID: 1, Type: "WorkflowExecutionStarted", Time: now.Add(-5 * time.Minute), Details: "WorkflowType: MockWorkflow, TaskQueue: mock-tasks"},
		{ID: 2, Type: "WorkflowTaskScheduled", Time: now.Add(-5 * time.Minute), Details: "TaskQueue: mock-tasks"},
		{ID: 3, Type: "WorkflowTaskStarted", Time: now.Add(-5 * time.Minute), Details: "Identity: worker-1@host"},
		{ID: 4, Type: "WorkflowTaskCompleted", Time: now.Add(-5 * time.Minute), Details: "ScheduledEventId: 2"},
		{ID: 5, Type: "ActivityTaskScheduled", Time: now.Add(-4 * time.Minute), Details: "ActivityType: MockActivity, TaskQueue: mock-tasks"},
		{ID: 6, Type: "ActivityTaskStarted", Time: now.Add(-4 * time.Minute), Details: "Identity: worker-1@host, Attempt: 1"},
		{ID: 7, Type: "ActivityTaskCompleted", Time: now.Add(-3 * time.Minute), Details: "ScheduledEventId: 5, Result: {success: true}"},
	}
	wd.render()
	wd.populateEventTable()
}

func (wd *WorkflowDetail) showError(err error) {
	wd.workflowView.SetText(fmt.Sprintf("\n [%s]Error: %s[-]", ui.TagFailed(), err.Error()))
	wd.eventDetailView.SetText("")
}

func (wd *WorkflowDetail) render() {
	if wd.workflow == nil {
		wd.workflowView.SetText(fmt.Sprintf(" [%s]Workflow not found[-]", ui.TagFailed()))
		return
	}

	w := wd.workflow
	now := time.Now()
	statusColor := ui.StatusColorTag(w.Status)
	statusIcon := ui.StatusIcon(w.Status)

	durationStr := "In progress"
	if w.EndTime != nil {
		durationStr = w.EndTime.Sub(w.StartTime).Round(time.Second).String()
	} else if w.Status == "Running" {
		durationStr = time.Since(w.StartTime).Round(time.Second).String()
	}

	// Combined workflow info
	workflowText := fmt.Sprintf(`
[%s::b]ID[-:-:-]           [%s]%s[-]
[%s::b]Type[-:-:-]         [%s]%s[-]
[%s::b]Status[-:-:-]       [%s]%s %s[-]
[%s::b]Started[-:-:-]      [%s]%s[-]
[%s::b]Duration[-:-:-]     [%s]%s[-]
[%s::b]Task Queue[-:-:-]   [%s]%s[-]
[%s::b]Run ID[-:-:-]       [%s]%s[-]`,
		ui.TagFgDim(), ui.TagFg(), w.ID,
		ui.TagFgDim(), ui.TagFg(), w.Type,
		ui.TagFgDim(), statusColor, statusIcon, w.Status,
		ui.TagFgDim(), ui.TagFg(), formatRelativeTime(now, w.StartTime),
		ui.TagFgDim(), ui.TagFg(), durationStr,
		ui.TagFgDim(), ui.TagFg(), w.TaskQueue,
		ui.TagFgDim(), ui.TagFgDim(), truncateStr(w.RunID, 25),
	)
	wd.workflowView.SetText(workflowText)
}

func (wd *WorkflowDetail) updateEventDetail(ev temporal.HistoryEvent) {
	icon := eventIcon(ev.Type)
	colorTag := eventColorTag(ev.Type)

	// Parse and format the details string
	formattedDetails := formatEventDetails(ev.Details)

	detailText := fmt.Sprintf(`
[%s::b]Event ID[-:-:-]     [%s]%d[-]
[%s::b]Type[-:-:-]         [%s]%s %s[-]
[%s::b]Time[-:-:-]         [%s]%s[-]

%s`,
		ui.TagFgDim(), ui.TagFg(), ev.ID,
		ui.TagFgDim(), colorTag, icon, ev.Type,
		ui.TagFgDim(), ui.TagFg(), ev.Time.Format("2006-01-02 15:04:05.000"),
		formattedDetails,
	)
	wd.eventDetailView.SetText(detailText)
}

// formatEventDetails parses event details and formats them with pretty JSON.
func formatEventDetails(details string) string {
	if details == "" {
		return fmt.Sprintf("[%s]No details[-]", ui.TagFgDim())
	}

	// First check if the whole thing is JSON
	trimmed := strings.TrimSpace(details)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		formatted := formatJSONPretty(details)
		return highlightFormattedJSONWorkflow(formatted)
	}

	// Handle key-value format with embedded JSON
	return formatKeyValueDetailsWorkflow(details)
}

// formatKeyValueDetailsWorkflow formats key-value style details with embedded JSON.
func formatKeyValueDetailsWorkflow(details string) string {
	var result strings.Builder

	// Split by commas while preserving JSON objects
	parts := splitPreservingJSONWorkflow(details)

	// First pass: find max key length for alignment
	type kvPair struct {
		key   string
		value string
	}
	var pairs []kvPair
	maxKeyLen := 0

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Find the key-value split point (first colon not inside JSON)
		colonIdx := findKeyColonIndex(part)
		if colonIdx > 0 {
			key := strings.TrimSpace(part[:colonIdx])
			value := strings.TrimSpace(part[colonIdx+1:])
			pairs = append(pairs, kvPair{key, value})
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}
		} else {
			pairs = append(pairs, kvPair{"", part})
		}
	}

	// Second pass: format with aligned keys
	for i, kv := range pairs {
		if i > 0 {
			result.WriteString("\n")
		}

		if kv.key != "" {
			// Pad key for alignment
			paddedKey := kv.key + strings.Repeat(" ", maxKeyLen-len(kv.key))

			// Check if value is JSON
			value := strings.TrimSpace(kv.value)
			if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") {
				formatted := formatJSONPretty(value)
				if formatted != value {
					// JSON was successfully formatted - put it on next line at left margin
					result.WriteString(fmt.Sprintf("[%s::b]%s[-:-:-]\n", ui.TagFgDim(), paddedKey))
					result.WriteString(highlightFormattedJSONWorkflow(formatted))
				} else {
					result.WriteString(fmt.Sprintf("[%s::b]%s[-:-:-]  ", ui.TagFgDim(), paddedKey))
					result.WriteString(highlightJSONLineWorkflow(value))
				}
			} else {
				result.WriteString(fmt.Sprintf("[%s::b]%s[-:-:-]  ", ui.TagFgDim(), paddedKey))
				result.WriteString(fmt.Sprintf("[%s]%s[-]", ui.TagFg(), highlightValuesWorkflow(value)))
			}
		} else {
			result.WriteString(fmt.Sprintf("[%s]%s[-]", ui.TagFg(), kv.value))
		}
	}

	return result.String()
}

// splitPreservingJSONWorkflow splits a string by commas while preserving JSON objects.
func splitPreservingJSONWorkflow(s string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, ch := range s {
		switch ch {
		case '{', '[':
			depth++
			current.WriteRune(ch)
		case '}', ']':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// findKeyColonIndex finds the index of the colon that separates key from value.
// It ignores colons inside JSON objects or strings.
func findKeyColonIndex(s string) int {
	depth := 0
	inString := false
	for i, ch := range s {
		switch ch {
		case '"':
			inString = !inString
		case '{', '[':
			if !inString {
				depth++
			}
		case '}', ']':
			if !inString {
				depth--
			}
		case ':':
			if depth == 0 && !inString {
				return i
			}
		}
	}
	return -1
}

// formatJSONPretty attempts to format a string as pretty JSON.
func formatJSONPretty(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		return s
	}

	pretty, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return s
	}

	return string(pretty)
}

// highlightFormattedJSONWorkflow applies syntax highlighting to formatted JSON.
func highlightFormattedJSONWorkflow(formatted string) string {
	lines := strings.Split(formatted, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, highlightJSONLineWorkflow(line))
	}
	return strings.Join(result, "\n")
}

// highlightJSONLineWorkflow highlights a single line of JSON content.
func highlightJSONLineWorkflow(line string) string {
	// Check for key: value pattern
	if colonIdx := strings.Index(line, ":"); colonIdx > 0 {
		prefix := line[:colonIdx]
		suffix := line[colonIdx+1:]

		trimmed := strings.TrimSpace(prefix)
		if strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"") {
			// JSON key with quotes - use accent color
			return fmt.Sprintf("[%s]%s[-]:[%s]%s[-]", ui.TagAccent(), prefix, ui.TagFg(), highlightValuesWorkflow(suffix))
		}
	}

	return highlightValuesWorkflow(line)
}

// highlightValuesWorkflow highlights JSON values (booleans, null).
func highlightValuesWorkflow(s string) string {
	result := s
	result = strings.ReplaceAll(result, "true", fmt.Sprintf("[%s]true[-]", ui.TagCompleted()))
	result = strings.ReplaceAll(result, "false", fmt.Sprintf("[%s]false[-]", ui.TagFailed()))
	result = strings.ReplaceAll(result, "null", fmt.Sprintf("[%s]null[-]", ui.TagFgDim()))
	return result
}

func (wd *WorkflowDetail) populateEventTable() {
	// Preserve current selection
	currentRow := wd.eventTable.SelectedRow()

	wd.eventTable.ClearRows()
	wd.eventTable.SetHeaders("ID", "TIME", "TYPE")

	for _, ev := range wd.events {
		icon := eventIcon(ev.Type)
		color := eventColor(ev.Type)
		wd.eventTable.AddColoredRow(color,
			fmt.Sprintf("%d", ev.ID),
			ev.Time.Format("15:04:05"),
			icon+" "+truncateStr(ev.Type, 30),
		)
	}

	if wd.eventTable.RowCount() > 0 {
		// Restore previous selection if valid, otherwise select first row
		if currentRow >= 0 && currentRow < len(wd.events) {
			wd.eventTable.SelectRow(currentRow)
			wd.updateEventDetail(wd.events[currentRow])
		} else {
			wd.eventTable.SelectRow(0)
			if len(wd.events) > 0 {
				wd.updateEventDetail(wd.events[0])
			}
		}
	}
}

// Name returns the view name.
func (wd *WorkflowDetail) Name() string {
	return "workflow-detail"
}

// Start is called when the view becomes active.
func (wd *WorkflowDetail) Start() {
	wd.eventTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r':
			wd.loadData()
			return nil
		case 'e':
			// Navigate to event history/graph view
			wd.app.NavigateToEvents(wd.workflowID, wd.runID)
			return nil
		case 'y':
			wd.yankEventData()
			return nil
		case 'd':
			wd.showEventDetailModal()
			return nil
		case 'c':
			wd.showCancelConfirm()
			return nil
		case 'X':
			wd.showTerminateConfirm()
			return nil
		case 's':
			wd.showSignalInput()
			return nil
		case 'D':
			wd.showDeleteConfirm()
			return nil
		case 'R':
			wd.showResetSelector()
			return nil
		case 'Q':
			wd.showQueryInput()
			return nil
		}
		return event
	})
	wd.loadData()
}

// Stop is called when the view is deactivated.
func (wd *WorkflowDetail) Stop() {
	wd.eventTable.SetInputCapture(nil)
	if wd.unsubscribeTheme != nil {
		wd.unsubscribeTheme()
	}
	// Clean up component theme listeners to prevent memory leaks and visual glitches
	wd.eventTable.Destroy()
	wd.workflowPanel.Destroy()
	wd.eventDetailPanel.Destroy()
	wd.eventsPanel.Destroy()
}

// Hints returns keybinding hints for this view.
func (wd *WorkflowDetail) Hints() []ui.KeyHint {
	hints := []ui.KeyHint{
		{Key: "e", Description: "Event Graph"},
		{Key: "d", Description: "Detail"},
		{Key: "y", Description: "Yank"},
		{Key: "r", Description: "Refresh"},
		{Key: "j/k", Description: "Navigate"},
	}

	// Only show mutation hints if workflow is running
	if wd.workflow != nil && wd.workflow.Status == "Running" {
		hints = append(hints,
			ui.KeyHint{Key: "c", Description: "Cancel"},
			ui.KeyHint{Key: "X", Description: "Terminate"},
			ui.KeyHint{Key: "s", Description: "Signal"},
			ui.KeyHint{Key: "Q", Description: "Query"},
		)
	}

	// Reset is available for completed/failed workflows
	if wd.workflow != nil && (wd.workflow.Status == "Completed" || wd.workflow.Status == "Failed" || wd.workflow.Status == "Terminated" || wd.workflow.Status == "Canceled") {
		hints = append(hints, ui.KeyHint{Key: "R", Description: "Reset"})
	}

	hints = append(hints,
		ui.KeyHint{Key: "D", Description: "Delete"},
		ui.KeyHint{Key: "T", Description: "Theme"},
		ui.KeyHint{Key: "esc", Description: "Back"},
	)

	return hints
}

// Focus sets focus to the event table.
func (wd *WorkflowDetail) Focus(delegate func(p tview.Primitive)) {
	delegate(wd.eventTable)
}

// Draw applies theme colors dynamically and draws the view.
func (wd *WorkflowDetail) Draw(screen tcell.Screen) {
	bg := ui.ColorBg()
	wd.SetBackgroundColor(bg)
	wd.leftFlex.SetBackgroundColor(bg)
	wd.workflowView.SetBackgroundColor(bg)
	wd.eventDetailView.SetBackgroundColor(bg)
	wd.Flex.Draw(screen)
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Mutation methods

func (wd *WorkflowDetail) showCancelConfirm() {
	command := fmt.Sprintf(`temporal workflow cancel \
  --workflow-id %s \
  --run-id %s \
  --namespace %s \
  --reason "Cancelled via TUI"`,
		wd.workflowID, wd.runID, wd.app.CurrentNamespace())

	modal := ui.NewConfirmModal(
		"Cancel Workflow",
		fmt.Sprintf("Cancel workflow %s?", wd.workflowID),
		command,
	).SetOnConfirm(func() {
		wd.executeCancelWorkflow()
	}).SetOnCancel(func() {
		wd.closeModal("confirm-cancel")
	})

	wd.app.UI().Pages().AddPage("confirm-cancel", modal, true, true)
	wd.app.UI().SetFocus(modal)
}

func (wd *WorkflowDetail) executeCancelWorkflow() {
	provider := wd.app.Provider()
	if provider == nil {
		wd.closeModal("confirm-cancel")
		wd.showError(fmt.Errorf("no provider connected"))
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := provider.CancelWorkflow(ctx,
			wd.app.CurrentNamespace(),
			wd.workflowID,
			wd.runID,
			"Cancelled via TUI")

		wd.app.UI().QueueUpdateDraw(func() {
			wd.closeModal("confirm-cancel")
			if err != nil {
				wd.showError(err)
			} else {
				wd.loadData() // Refresh to show new status
			}
		})
	}()
}

func (wd *WorkflowDetail) showTerminateConfirm() {
	command := fmt.Sprintf(`temporal workflow terminate \
  --workflow-id %s \
  --run-id %s \
  --namespace %s \
  --reason "Terminated via TUI"`,
		wd.workflowID, wd.runID, wd.app.CurrentNamespace())

	modal := ui.NewConfirmModal(
		"Terminate Workflow",
		fmt.Sprintf("Terminate workflow %s?", wd.workflowID),
		command,
	).SetWarning("This will forcefully terminate the workflow. No cleanup code will run.").
		SetOnConfirm(func() {
			wd.executeTerminateWorkflow()
		}).SetOnCancel(func() {
		wd.closeModal("confirm-terminate")
	})

	wd.app.UI().Pages().AddPage("confirm-terminate", modal, true, true)
	wd.app.UI().SetFocus(modal)
}

func (wd *WorkflowDetail) executeTerminateWorkflow() {
	provider := wd.app.Provider()
	if provider == nil {
		wd.closeModal("confirm-terminate")
		wd.showError(fmt.Errorf("no provider connected"))
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := provider.TerminateWorkflow(ctx,
			wd.app.CurrentNamespace(),
			wd.workflowID,
			wd.runID,
			"Terminated via TUI")

		wd.app.UI().QueueUpdateDraw(func() {
			wd.closeModal("confirm-terminate")
			if err != nil {
				wd.showError(err)
			} else {
				wd.loadData() // Refresh to show new status
			}
		})
	}()
}

func (wd *WorkflowDetail) showDeleteConfirm() {
	command := fmt.Sprintf(`temporal workflow delete \
  --workflow-id %s \
  --run-id %s \
  --namespace %s`,
		wd.workflowID, wd.runID, wd.app.CurrentNamespace())

	modal := ui.NewConfirmModal(
		"Delete Workflow",
		fmt.Sprintf("Delete workflow %s?", wd.workflowID),
		command,
	).SetWarning("This will permanently delete the workflow and its history. This cannot be undone.").
		SetOnConfirm(func() {
			wd.executeDeleteWorkflow()
		}).SetOnCancel(func() {
		wd.closeModal("confirm-delete")
	})

	wd.app.UI().Pages().AddPage("confirm-delete", modal, true, true)
	wd.app.UI().SetFocus(modal)
}

func (wd *WorkflowDetail) executeDeleteWorkflow() {
	provider := wd.app.Provider()
	if provider == nil {
		wd.closeModal("confirm-delete")
		wd.showError(fmt.Errorf("no provider connected"))
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := provider.DeleteWorkflow(ctx,
			wd.app.CurrentNamespace(),
			wd.workflowID,
			wd.runID)

		wd.app.UI().QueueUpdateDraw(func() {
			wd.closeModal("confirm-delete")
			if err != nil {
				wd.showError(err)
			} else {
				// Navigate back to workflow list since this workflow no longer exists
				wd.app.UI().Pages().Pop()
			}
		})
	}()
}

func (wd *WorkflowDetail) showSignalInput() {
	fields := []ui.InputField{
		{
			Name:        "signalName",
			Label:       "Signal Name",
			Placeholder: "e.g., approve, cancel, update",
			Required:    true,
		},
		{
			Name:        "input",
			Label:       "Input (JSON)",
			Placeholder: `e.g., {"approved": true}`,
			Required:    false,
		},
	}

	modal := ui.NewInputModal(
		"Signal Workflow",
		fmt.Sprintf("Send signal to workflow %s", wd.workflowID),
		fields,
	).SetOnSubmit(func(values map[string]string) {
		wd.executeSignalWorkflow(values["signalName"], values["input"])
	}).SetOnCancel(func() {
		wd.closeModal("signal-input")
	})

	wd.app.UI().Pages().AddPage("signal-input", modal, true, true)
	wd.app.UI().SetFocus(modal)
}

func (wd *WorkflowDetail) executeSignalWorkflow(signalName, input string) {
	provider := wd.app.Provider()
	if provider == nil {
		wd.closeModal("signal-input")
		wd.showError(fmt.Errorf("no provider connected"))
		return
	}

	// Convert input string to bytes (empty if no input provided)
	var inputBytes []byte
	if input != "" {
		inputBytes = []byte(input)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := provider.SignalWorkflow(ctx,
			wd.app.CurrentNamespace(),
			wd.workflowID,
			wd.runID,
			signalName,
			inputBytes)

		wd.app.UI().QueueUpdateDraw(func() {
			wd.closeModal("signal-input")
			if err != nil {
				wd.showError(err)
			} else {
				wd.loadData() // Refresh to show signal event in history
			}
		})
	}()
}

func (wd *WorkflowDetail) showResetSelector() {
	provider := wd.app.Provider()
	if provider == nil {
		wd.showError(fmt.Errorf("no provider connected"))
		return
	}

	// Show loading state
	wd.workflowPanel.SetTitle("Workflow (Loading reset points...)")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resetPoints, err := provider.GetResetPoints(ctx,
			wd.app.CurrentNamespace(),
			wd.workflowID,
			wd.runID)

		wd.app.UI().QueueUpdateDraw(func() {
			wd.workflowPanel.SetTitle("Workflow")

			if err != nil {
				wd.showError(fmt.Errorf("failed to get reset points: %w", err))
				return
			}

			if len(resetPoints) == 0 {
				wd.showError(fmt.Errorf("no valid reset points found for this workflow"))
				return
			}

			// Check for failure point - if found, show quick reset modal
			picker := ui.NewResetPicker(resetPoints)
			if failurePoint, found := picker.GetFirstFailurePoint(); found {
				wd.showQuickResetModal(failurePoint, resetPoints)
			} else {
				// No failure point, show full picker directly
				wd.showResetPicker(resetPoints)
			}
		})
	}()
}

func (wd *WorkflowDetail) showQuickResetModal(failurePoint temporal.ResetPoint, allPoints []temporal.ResetPoint) {
	modal := ui.NewQuickResetModal(wd.workflowID, failurePoint)

	modal.SetOnConfirm(func() {
		wd.closeModal("quick-reset")
		wd.showResetConfirm(failurePoint.EventID)
	})

	modal.SetOnAdvanced(func() {
		wd.closeModal("quick-reset")
		wd.showResetPicker(allPoints)
	})

	modal.SetOnCancel(func() {
		wd.closeModal("quick-reset")
	})

	wd.app.UI().Pages().AddPage("quick-reset", modal, true, true)
	wd.app.UI().SetFocus(modal)
}

func (wd *WorkflowDetail) showResetPicker(resetPoints []temporal.ResetPoint) {
	picker := ui.NewResetPicker(resetPoints)

	picker.SetOnSelect(func(eventID int64, description string) {
		wd.closeModal("reset-picker")
		wd.showResetConfirm(eventID)
	})

	picker.SetOnCancel(func() {
		wd.closeModal("reset-picker")
	})

	// Create a centered modal layout for the picker
	height := picker.GetHeight()
	width := 80

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(picker, width, 0, true).
			AddItem(nil, 0, 1, false), height, 0, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(ui.ColorBg())

	wd.app.UI().Pages().AddPage("reset-picker", flex, true, true)
	wd.app.UI().SetFocus(picker)
}

func (wd *WorkflowDetail) showResetConfirm(eventID int64) {
	wd.closeModal("reset-selector")

	command := fmt.Sprintf(`temporal workflow reset \
  --workflow-id %s \
  --run-id %s \
  --namespace %s \
  --event-id %d \
  --reason "Reset via TUI"`,
		wd.workflowID, wd.runID, wd.app.CurrentNamespace(), eventID)

	modal := ui.NewConfirmModal(
		"Reset Workflow",
		fmt.Sprintf("Reset workflow %s to event %d?", wd.workflowID, eventID),
		command,
	).SetWarning("This will create a new run from the specified event. The current run will remain unchanged.").
		SetOnConfirm(func() {
			wd.executeResetWorkflow(eventID)
		}).SetOnCancel(func() {
		wd.closeModal("confirm-reset")
	})

	wd.app.UI().Pages().AddPage("confirm-reset", modal, true, true)
	wd.app.UI().SetFocus(modal)
}

func (wd *WorkflowDetail) executeResetWorkflow(eventID int64) {
	provider := wd.app.Provider()
	if provider == nil {
		wd.closeModal("confirm-reset")
		wd.showError(fmt.Errorf("no provider connected"))
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		newRunID, err := provider.ResetWorkflow(ctx,
			wd.app.CurrentNamespace(),
			wd.workflowID,
			wd.runID,
			eventID,
			"Reset via TUI")

		wd.app.UI().QueueUpdateDraw(func() {
			wd.closeModal("confirm-reset")
			if err != nil {
				wd.showError(err)
			} else {
				// Navigate to the new run
				wd.runID = newRunID
				wd.loadData()
			}
		})
	}()
}

func (wd *WorkflowDetail) closeModal(name string) {
	wd.app.UI().Pages().RemovePage(name)
	// Restore focus to current view
	if current := wd.app.UI().Pages().Current(); current != nil {
		wd.app.UI().SetFocus(current)
	}
}

// Query methods

func (wd *WorkflowDetail) showQueryInput() {
	// Check if workflow is running - queries only work on running workflows
	if wd.workflow == nil || wd.workflow.Status != "Running" {
		wd.showError(fmt.Errorf("queries can only be executed on running workflows"))
		return
	}

	fields := []ui.InputField{
		{
			Name:        "queryType",
			Label:       "Query Type",
			Placeholder: "__stack_trace (or custom query handler name)",
			Required:    true,
		},
		{
			Name:        "args",
			Label:       "Arguments (JSON)",
			Placeholder: `e.g., {"key": "value"} (optional)`,
			Required:    false,
		},
	}

	modal := ui.NewInputModal(
		"Query Workflow",
		fmt.Sprintf("Execute query on workflow %s", wd.workflowID),
		fields,
	).SetOnSubmit(func(values map[string]string) {
		wd.executeQuery(values["queryType"], values["args"])
	}).SetOnCancel(func() {
		wd.closeModal("query-input")
	})

	wd.app.UI().Pages().AddPage("query-input", modal, true, true)
	wd.app.UI().SetFocus(modal)
}

func (wd *WorkflowDetail) executeQuery(queryType, args string) {
	wd.closeModal("query-input")

	provider := wd.app.Provider()
	if provider == nil {
		wd.showError(fmt.Errorf("no provider connected"))
		return
	}

	// Convert args string to bytes (empty if no args provided)
	var argsBytes []byte
	if args != "" {
		argsBytes = []byte(args)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := provider.QueryWorkflow(ctx,
			wd.app.CurrentNamespace(),
			wd.workflowID,
			wd.runID,
			queryType,
			argsBytes)

		wd.app.UI().QueueUpdateDraw(func() {
			if err != nil {
				wd.showQueryError(queryType, err.Error())
				return
			}
			if result.Error != "" {
				wd.showQueryError(queryType, result.Error)
				return
			}
			wd.showQueryResult(queryType, result.Result)
		})
	}()
}

func (wd *WorkflowDetail) showQueryResult(queryType, result string) {
	modal := ui.NewQueryResultModal().
		SetResult(queryType, result).
		SetOnClose(func() {
			wd.closeModal("query-result")
		})

	wd.app.UI().Pages().AddPage("query-result", modal, true, true)
	wd.app.UI().SetFocus(modal)
}

func (wd *WorkflowDetail) showQueryError(queryType, errMsg string) {
	modal := ui.NewQueryResultModal().
		SetError(queryType, errMsg).
		SetOnClose(func() {
			wd.closeModal("query-result")
		})

	wd.app.UI().Pages().AddPage("query-result", modal, true, true)
	wd.app.UI().SetFocus(modal)
}

// Yank and detail methods

// getSelectedEventDetails returns the details for the currently selected event.
func (wd *WorkflowDetail) getSelectedEventDetails() (string, string) {
	row := wd.eventTable.SelectedRow()
	if row < 0 || row >= len(wd.events) {
		return "", ""
	}
	ev := wd.events[row]
	return ev.Type, prettyPrintJSONDetail(ev.Details)
}

// yankEventData copies the selected event's details to clipboard.
func (wd *WorkflowDetail) yankEventData() {
	eventType, data := wd.getSelectedEventDetails()
	if data == "" {
		return
	}

	if err := ui.CopyToClipboard(data); err != nil {
		wd.eventDetailView.SetText(fmt.Sprintf("[%s]%s Failed to copy: %s[-]",
			ui.TagFailed(), ui.IconFailed, err.Error()))
		return
	}

	// Show success feedback
	wd.eventDetailView.SetText(fmt.Sprintf(`
[%s::b]Copied to clipboard[-:-:-]

[%s]%s[-]

[%s]%s[-]`,
		ui.TagPanelTitle(),
		ui.TagAccent(), eventType,
		ui.TagCompleted(), "Event data copied!"))

	// Restore detail after a brief delay
	go func() {
		time.Sleep(1500 * time.Millisecond)
		wd.app.UI().QueueUpdateDraw(func() {
			row := wd.eventTable.SelectedRow()
			if row >= 0 && row < len(wd.events) {
				wd.updateEventDetail(wd.events[row])
			}
		})
	}()
}

// showEventDetailModal shows a full-screen modal with the event details.
func (wd *WorkflowDetail) showEventDetailModal() {
	eventType, data := wd.getSelectedEventDetails()
	if data == "" {
		return
	}

	// Create scrollable text view for the detail
	textView := tview.NewTextView()
	textView.SetDynamicColors(true)
	textView.SetScrollable(true)
	textView.SetWordWrap(true)
	textView.SetBackgroundColor(ui.ColorBg())
	textView.SetTextColor(ui.ColorFg())

	// Format with syntax highlighting
	formattedData := formatDetailViewWithHighlighting(data)
	textView.SetText(formattedData)

	// Create modal using the base Modal component for consistent styling
	modal := ui.NewModal(ui.ModalConfig{
		Title:     "Detail",
		Width:     80,
		Height:    20,
		MinHeight: 10,
		MaxHeight: 30,
		Backdrop:  true,
	})
	modal.SetContent(textView)
	modal.SetHints([]ui.KeyHint{
		{Key: "j/k", Description: "Scroll"},
		{Key: "g/G", Description: "Top/Bottom"},
		{Key: "y", Description: "Yank"},
		{Key: "q", Description: "Close"},
	})

	// Update panel title with event type
	modal.GetPanel().SetTitle(fmt.Sprintf("Detail: %s", truncateEventTypeStr(eventType)))

	modal.SetOnClose(func() {
		wd.closeModal("event-detail")
	})

	// Input handler for the modal
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			wd.closeModal("event-detail")
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				wd.closeModal("event-detail")
				return nil
			case 'y':
				if err := ui.CopyToClipboard(data); err == nil {
					// Show brief feedback
					title := fmt.Sprintf("Detail: %s", truncateEventTypeStr(eventType))
					modal.GetPanel().SetTitle("Copied!")
					modal.GetPanel().SetTitleColor(ui.ColorCompleted())
					go func() {
						time.Sleep(1 * time.Second)
						wd.app.UI().QueueUpdateDraw(func() {
							modal.GetPanel().SetTitle(title)
							modal.GetPanel().SetTitleColor(tcell.ColorDefault)
						})
					}()
				}
				return nil
			case 'j':
				row, _ := textView.GetScrollOffset()
				textView.ScrollTo(row+1, 0)
				return nil
			case 'k':
				row, _ := textView.GetScrollOffset()
				if row > 0 {
					textView.ScrollTo(row-1, 0)
				}
				return nil
			case 'G':
				textView.ScrollToEnd()
				return nil
			case 'g':
				textView.ScrollTo(0, 0)
				return nil
			}
		}
		return event
	})

	wd.app.UI().Pages().AddPage("event-detail", modal, true, true)
	wd.app.UI().SetFocus(textView)
}

// truncateEventTypeStr shortens long event type names for the title.
func truncateEventTypeStr(eventType string) string {
	if len(eventType) > 30 {
		return eventType[:27] + "..."
	}
	return eventType
}

// prettyPrintJSONDetail attempts to format JSON in the details string.
func prettyPrintJSONDetail(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Try to parse the whole thing as JSON first
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		var parsed interface{}
		if err := json.Unmarshal([]byte(s), &parsed); err == nil {
			pretty, err := json.MarshalIndent(parsed, "", "  ")
			if err == nil {
				return string(pretty)
			}
		}
	}

	// Otherwise, try to find and format JSON embedded in the string
	// Look for patterns like "Result: {...}" or "Input: {...}"
	var result strings.Builder
	parts := strings.Split(s, ", ")
	for i, part := range parts {
		if i > 0 {
			result.WriteString("\n")
		}

		// Check if this part has embedded JSON
		if colonIdx := strings.Index(part, ": "); colonIdx > 0 {
			key := part[:colonIdx]
			value := part[colonIdx+2:]

			// Try to parse and pretty-print the value as JSON
			if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") {
				var parsed interface{}
				if err := json.Unmarshal([]byte(value), &parsed); err == nil {
					pretty, err := json.MarshalIndent(parsed, "", "  ")
					if err == nil {
						result.WriteString(fmt.Sprintf("%s:\n%s", key, string(pretty)))
						continue
					}
				}
			}
			result.WriteString(fmt.Sprintf("%s: %s", key, value))
		} else {
			result.WriteString(part)
		}
	}

	return result.String()
}

// formatDetailViewWithHighlighting adds color tags for syntax highlighting.
func formatDetailViewWithHighlighting(data string) string {
	lines := strings.Split(data, "\n")
	var result []string

	for _, line := range lines {
		highlighted := highlightDetailLine(line)
		result = append(result, highlighted)
	}

	return strings.Join(result, "\n")
}

// highlightDetailLine adds tview color tags to a single line.
func highlightDetailLine(line string) string {
	// If line contains a colon that looks like a JSON key
	if idx := strings.Index(line, ":"); idx > 0 {
		prefix := line[:idx]
		suffix := line[idx:]

		trimmed := strings.TrimSpace(prefix)
		if strings.HasPrefix(trimmed, "\"") || strings.HasPrefix(trimmed, "'") {
			return fmt.Sprintf("[%s]%s[-]%s", ui.TagAccent(), prefix, highlightDetailValue(suffix))
		} else if !strings.Contains(trimmed, " ") && len(trimmed) > 0 {
			return fmt.Sprintf("[%s::b]%s[-:-:-]%s", ui.TagAccent(), prefix, highlightDetailValue(suffix))
		}
	}

	return highlightDetailValue(line)
}

// highlightDetailValue highlights JSON values.
func highlightDetailValue(s string) string {
	result := s
	result = strings.ReplaceAll(result, "true", fmt.Sprintf("[%s]true[-]", ui.TagCompleted()))
	result = strings.ReplaceAll(result, "false", fmt.Sprintf("[%s]false[-]", ui.TagFailed()))
	result = strings.ReplaceAll(result, "null", fmt.Sprintf("[%s]null[-]", ui.TagFgDim()))
	return result
}
