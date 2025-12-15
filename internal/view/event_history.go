package view

import (
	"context"
	"fmt"
	"time"

	"github.com/atterpac/temportui/internal/temporal"
	"github.com/atterpac/temportui/internal/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// EventHistory displays workflow event history with a side panel for details.
type EventHistory struct {
	*tview.Flex
	app         *App
	workflowID  string
	runID       string
	table       *ui.Table
	sidePanel   *tview.TextView
	events      []temporal.HistoryEvent
	sidePanelOn bool
	loading     bool
}

// NewEventHistory creates a new event history view.
func NewEventHistory(app *App, workflowID, runID string) *EventHistory {
	eh := &EventHistory{
		Flex:       tview.NewFlex().SetDirection(tview.FlexColumn),
		app:        app,
		workflowID: workflowID,
		runID:      runID,
		table:      ui.NewTable(),
		events:     []temporal.HistoryEvent{},
	}
	eh.setup()
	return eh
}

func (eh *EventHistory) setup() {
	eh.SetBackgroundColor(ui.ColorBg)

	eh.table.SetHeaders("ID", "TIME", "TYPE", "DETAILS")
	eh.table.SetBorder(false) // Charm-style: borderless
	eh.table.SetBackgroundColor(ui.ColorBg)

	// Create side panel (hidden initially) - keep minimal border for separation
	eh.sidePanel = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	eh.sidePanel.SetBorder(false)
	eh.sidePanel.SetBackgroundColor(ui.ColorBgLight)

	// Selection change handler
	eh.table.SetSelectionChangedFunc(func(row, col int) {
		if eh.sidePanelOn && row > 0 {
			eh.updateSidePanel(row - 1)
		}
	})

	// Selection handler (Enter key)
	eh.table.SetSelectedFunc(func(row, col int) {
		if row > 0 {
			eh.toggleSidePanel()
			if eh.sidePanelOn {
				eh.updateSidePanel(row - 1)
			}
		}
	})

	// Main layout
	eh.AddItem(eh.table, 0, 2, true)
}

func (eh *EventHistory) setLoading(loading bool) {
	eh.loading = loading
}

func (eh *EventHistory) loadData() {
	provider := eh.app.Provider()
	if provider == nil {
		// Fallback to mock data if no provider
		eh.loadMockData()
		return
	}

	eh.setLoading(true)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Longer timeout for history
		defer cancel()

		events, err := provider.GetWorkflowHistory(ctx, eh.app.CurrentNamespace(), eh.workflowID, eh.runID)

		eh.app.UI().QueueUpdateDraw(func() {
			eh.setLoading(false)
			if err != nil {
				eh.showError(err)
				return
			}
			eh.events = events
			eh.populateTable()
		})
	}()
}

func (eh *EventHistory) loadMockData() {
	// Mock data fallback when no provider is configured
	now := time.Now()
	eh.events = []temporal.HistoryEvent{
		{ID: 1, Type: "WorkflowExecutionStarted", Time: now.Add(-5 * time.Minute), Details: "WorkflowType: MockWorkflow, TaskQueue: mock-tasks"},
		{ID: 2, Type: "WorkflowTaskScheduled", Time: now.Add(-5 * time.Minute), Details: "TaskQueue: mock-tasks"},
		{ID: 3, Type: "WorkflowTaskStarted", Time: now.Add(-5 * time.Minute), Details: "Identity: worker-1@host"},
		{ID: 4, Type: "WorkflowTaskCompleted", Time: now.Add(-5 * time.Minute), Details: "ScheduledEventId: 2"},
		{ID: 5, Type: "ActivityTaskScheduled", Time: now.Add(-4 * time.Minute), Details: "ActivityType: MockActivity, TaskQueue: mock-tasks"},
		{ID: 6, Type: "ActivityTaskStarted", Time: now.Add(-4 * time.Minute), Details: "Identity: worker-1@host, Attempt: 1"},
		{ID: 7, Type: "ActivityTaskCompleted", Time: now.Add(-3 * time.Minute), Details: "ScheduledEventId: 5, Result: {success: true}"},
	}
	eh.populateTable()
}

func (eh *EventHistory) populateTable() {
	eh.table.ClearRows()
	eh.table.SetHeaders("ID", "TIME", "TYPE", "DETAILS")

	for _, ev := range eh.events {
		icon := eventIcon(ev.Type)
		color := eventColor(ev.Type)
		eh.table.AddColoredRow(color,
			fmt.Sprintf("%d", ev.ID),
			ev.Time.Format("15:04:05"),
			icon+" "+ev.Type,
			truncate(ev.Details, 40),
		)
	}

	if eh.table.RowCount() > 0 {
		eh.table.SelectRow(0)
	}
}

func (eh *EventHistory) showError(err error) {
	eh.table.ClearRows()
	eh.table.SetHeaders("ID", "TIME", "TYPE", "DETAILS")
	eh.table.AddColoredRow(ui.ColorFailed,
		"",
		"",
		ui.IconFailed+" Error loading events",
		err.Error(),
	)
}

func (eh *EventHistory) toggleSidePanel() {
	if eh.sidePanelOn {
		eh.RemoveItem(eh.sidePanel)
		eh.sidePanelOn = false
	} else {
		eh.AddItem(eh.sidePanel, 0, 1, false)
		eh.sidePanelOn = true
	}
}

func (eh *EventHistory) updateSidePanel(index int) {
	if index < 0 || index >= len(eh.events) {
		return
	}

	ev := eh.events[index]
	icon := eventIcon(ev.Type)
	colorTag := eventColorTag(ev.Type)

	text := fmt.Sprintf(
		"\n [%s]%s Event ID:[%s]    %d\n\n"+
			" [%s]%s Type:[%s]\n   [%s]%s %s[-]\n\n"+
			" [%s]%s Time:[%s]\n   %s\n\n"+
			" [%s]%s Details:[%s]\n   %s\n",
		ui.TagFgDim, ui.IconBullet, ui.TagFg, ev.ID,
		ui.TagFgDim, ui.IconBullet, ui.TagFg, colorTag, icon, ev.Type,
		ui.TagFgDim, ui.IconBullet, ui.TagFg, ev.Time.Format("2006-01-02 15:04:05.000"),
		ui.TagFgDim, ui.IconBullet, ui.TagFg, ev.Details,
	)
	eh.sidePanel.SetText(text)
}

// Name returns the view name.
func (eh *EventHistory) Name() string {
	return "events"
}

// Start is called when the view becomes active.
func (eh *EventHistory) Start() {
	eh.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'p':
			eh.toggleSidePanel()
			if eh.sidePanelOn {
				row := eh.table.SelectedRow()
				if row >= 0 {
					eh.updateSidePanel(row)
				}
			}
			return nil
		case 'r':
			eh.loadData()
			return nil
		}
		return event
	})
	// Load data when view becomes active
	eh.loadData()
}

// Stop is called when the view is deactivated.
func (eh *EventHistory) Stop() {
	eh.table.SetInputCapture(nil)
}

// Hints returns keybinding hints for this view.
func (eh *EventHistory) Hints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "enter", Description: "Toggle Detail"},
		{Key: "p", Description: "Panel"},
		{Key: "r", Description: "Refresh"},
		{Key: "j/k", Description: "Navigate"},
		{Key: "esc", Description: "Back"},
	}
}

// eventIcon returns an icon for the event type.
func eventIcon(eventType string) string {
	switch {
	case contains(eventType, "Started"):
		return ui.IconRunning
	case contains(eventType, "Completed"):
		return ui.IconCompleted
	case contains(eventType, "Failed"):
		return ui.IconFailed
	case contains(eventType, "Scheduled"):
		return ui.IconPending
	case contains(eventType, "Timer"):
		return ui.IconTimedOut
	case contains(eventType, "Signal"):
		return ui.IconActivity
	case contains(eventType, "Child"):
		return ui.IconWorkflow
	default:
		return ui.IconEvent
	}
}

// eventColor returns a color for the event type.
func eventColor(eventType string) tcell.Color {
	switch {
	case contains(eventType, "Started"):
		return ui.ColorRunning
	case contains(eventType, "Completed"):
		return ui.ColorCompleted
	case contains(eventType, "Failed"):
		return ui.ColorFailed
	case contains(eventType, "Scheduled"):
		return ui.ColorFgDim
	default:
		return ui.ColorFg
	}
}

// eventColorTag returns a color tag for the event type.
func eventColorTag(eventType string) string {
	switch {
	case contains(eventType, "Started"):
		return ui.TagRunning
	case contains(eventType, "Completed"):
		return ui.TagCompleted
	case contains(eventType, "Failed"):
		return ui.TagFailed
	default:
		return ui.TagFg
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
