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

// WorkflowList displays a list of workflows.
type WorkflowList struct {
	*tview.Flex
	app           *App
	namespace     string
	table         *ui.Table
	workflows     []temporal.Workflow
	filterText    string
	loading       bool
	autoRefresh   bool
	refreshTicker *time.Ticker
	stopRefresh   chan struct{}
}

// NewWorkflowList creates a new workflow list view.
func NewWorkflowList(app *App, namespace string) *WorkflowList {
	wl := &WorkflowList{
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
		app:         app,
		namespace:   namespace,
		table:       ui.NewTable(),
		workflows:   []temporal.Workflow{},
		stopRefresh: make(chan struct{}),
	}
	wl.setup()
	return wl
}

func (wl *WorkflowList) setup() {
	wl.table.SetHeaders("WORKFLOW ID", "TYPE", "STATUS", "START TIME", "END TIME")
	wl.table.SetBorder(false) // Charm-style: borderless
	wl.table.SetBackgroundColor(ui.ColorBg)
	wl.SetBackgroundColor(ui.ColorBg)

	// Selection handler
	wl.table.SetOnSelect(func(row int) {
		if row >= 0 && row < len(wl.workflows) {
			wf := wl.workflows[row]
			wl.app.NavigateToWorkflowDetail(wf.ID, wf.RunID)
		}
	})

	wl.AddItem(wl.table, 0, 1, true)
}

func (wl *WorkflowList) setLoading(loading bool) {
	wl.loading = loading
}

func (wl *WorkflowList) loadData() {
	provider := wl.app.Provider()
	if provider == nil {
		// Fallback to mock data if no provider
		wl.loadMockData()
		return
	}

	wl.setLoading(true)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		opts := temporal.ListOptions{
			PageSize: 100,
			Query:    wl.filterText,
		}
		workflows, _, err := provider.ListWorkflows(ctx, wl.namespace, opts)

		wl.app.UI().QueueUpdateDraw(func() {
			wl.setLoading(false)
			if err != nil {
				wl.showError(err)
				return
			}
			wl.workflows = workflows
			wl.populateTable()
		})
	}()
}

func (wl *WorkflowList) loadMockData() {
	// Mock data fallback when no provider is configured
	now := time.Now()
	wl.workflows = []temporal.Workflow{
		{
			ID: "order-processing-abc123", RunID: "run-001-xyz", Type: "OrderWorkflow",
			Status: "Running", Namespace: wl.namespace, TaskQueue: "order-tasks",
			StartTime: now.Add(-5 * time.Minute),
		},
		{
			ID: "payment-xyz789", RunID: "run-002-abc", Type: "PaymentWorkflow",
			Status: "Completed", Namespace: wl.namespace, TaskQueue: "payment-tasks",
			StartTime: now.Add(-1 * time.Hour), EndTime: ptr(now.Add(-55 * time.Minute)),
		},
		{
			ID: "shipment-def456", RunID: "run-003-def", Type: "ShipmentWorkflow",
			Status: "Failed", Namespace: wl.namespace, TaskQueue: "shipment-tasks",
			StartTime: now.Add(-30 * time.Minute), EndTime: ptr(now.Add(-25 * time.Minute)),
		},
		{
			ID: "inventory-check-111", RunID: "run-004-ghi", Type: "InventoryWorkflow",
			Status: "Running", Namespace: wl.namespace, TaskQueue: "inventory-tasks",
			StartTime: now.Add(-10 * time.Minute),
		},
		{
			ID: "user-signup-222", RunID: "run-005-jkl", Type: "UserOnboardingWorkflow",
			Status: "Completed", Namespace: wl.namespace, TaskQueue: "user-tasks",
			StartTime: now.Add(-2 * time.Hour), EndTime: ptr(now.Add(-1*time.Hour - 45*time.Minute)),
		},
	}
	wl.populateTable()
}

func ptr[T any](v T) *T {
	return &v
}

func (wl *WorkflowList) populateTable() {
	wl.table.ClearRows()
	wl.table.SetHeaders("WORKFLOW ID", "TYPE", "STATUS", "START TIME", "END TIME")

	now := time.Now()
	for _, w := range wl.workflows {
		endTime := ui.IconDot + " -"
		if w.EndTime != nil {
			endTime = formatRelativeTime(now, *w.EndTime)
		}

		// Use AddStyledRow for status icon and coloring
		wl.table.AddStyledRow(w.Status,
			truncate(w.ID, 25),
			w.Type,
			w.Status,
			formatRelativeTime(now, w.StartTime),
			endTime,
		)
	}

	if wl.table.RowCount() > 0 {
		wl.table.SelectRow(0)
	}
}

func (wl *WorkflowList) showError(err error) {
	wl.table.ClearRows()
	wl.table.SetHeaders("WORKFLOW ID", "TYPE", "STATUS", "START TIME", "END TIME")
	wl.table.AddColoredRow(ui.ColorFailed,
		ui.IconFailed+" Error loading workflows",
		err.Error(),
		"",
		"",
		"",
	)
}

func (wl *WorkflowList) toggleAutoRefresh() {
	wl.autoRefresh = !wl.autoRefresh
	if wl.autoRefresh {
		wl.startAutoRefresh()
	} else {
		wl.stopAutoRefresh()
	}
}

func (wl *WorkflowList) startAutoRefresh() {
	wl.refreshTicker = time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-wl.refreshTicker.C:
				wl.app.UI().QueueUpdateDraw(func() {
					wl.loadData()
				})
			case <-wl.stopRefresh:
				return
			}
		}
	}()
}

func (wl *WorkflowList) stopAutoRefresh() {
	if wl.refreshTicker != nil {
		wl.refreshTicker.Stop()
		wl.refreshTicker = nil
	}
	// Signal stop to the goroutine
	select {
	case wl.stopRefresh <- struct{}{}:
	default:
	}
}

// Name returns the view name.
func (wl *WorkflowList) Name() string {
	return "workflows"
}

// Start is called when the view becomes active.
func (wl *WorkflowList) Start() {
	wl.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '/':
			wl.showFilter()
			return nil
		case 't':
			wl.app.NavigateToTaskQueues()
			return nil
		case 'a':
			wl.toggleAutoRefresh()
			return nil
		case 'r':
			wl.loadData()
			return nil
		}
		return event
	})
	// Load data when view becomes active
	wl.loadData()
}

// Stop is called when the view is deactivated.
func (wl *WorkflowList) Stop() {
	wl.table.SetInputCapture(nil)
	wl.stopAutoRefresh()
}

// Hints returns keybinding hints for this view.
func (wl *WorkflowList) Hints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "enter", Description: "Detail"},
		{Key: "/", Description: "Filter"},
		{Key: "r", Description: "Refresh"},
		{Key: "a", Description: "Auto-refresh"},
		{Key: "t", Description: "Task Queues"},
		{Key: "j/k", Description: "Navigate"},
		{Key: "esc", Description: "Back"},
	}
}

func (wl *WorkflowList) showFilter() {
	// Create filter input with styling
	input := tview.NewInputField().
		SetLabel(" " + ui.IconArrowRight + " Filter: ").
		SetFieldWidth(30).
		SetFieldBackgroundColor(ui.ColorBgLight).
		SetFieldTextColor(ui.ColorFg).
		SetLabelColor(ui.ColorAccent)

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			wl.filterText = input.GetText()
			wl.loadData() // Reload with filter
		}
		// Remove input and restore focus
		wl.RemoveItem(input)
		wl.app.UI().SetFocus(wl.table)
	})

	wl.AddItem(input, 1, 0, false)
	wl.app.UI().SetFocus(input)
}

func formatRelativeTime(now time.Time, t time.Time) string {
	d := now.Sub(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd ago", days)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
