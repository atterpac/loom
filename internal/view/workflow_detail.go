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

// WorkflowDetail displays detailed information about a workflow.
type WorkflowDetail struct {
	*tview.Flex
	app        *App
	workflowID string
	runID      string
	workflow   *temporal.Workflow
	header     *tview.TextView
	infoPanel  *tview.TextView
	tabBar     *tview.TextView
	loading    bool
}

// NewWorkflowDetail creates a new workflow detail view.
func NewWorkflowDetail(app *App, workflowID, runID string) *WorkflowDetail {
	wd := &WorkflowDetail{
		Flex:       tview.NewFlex().SetDirection(tview.FlexRow),
		app:        app,
		workflowID: workflowID,
		runID:      runID,
	}
	wd.setup()
	return wd
}

func (wd *WorkflowDetail) setup() {
	wd.SetBackgroundColor(ui.ColorBg)

	// Header section with workflow identity - Charm-style: borderless
	wd.header = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	wd.header.SetBorder(false)
	wd.header.SetBackgroundColor(ui.ColorBg)

	// Tab bar
	wd.tabBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	wd.tabBar.SetBackgroundColor(ui.ColorBg)

	// Info panel - Charm-style: borderless, subtle bg
	wd.infoPanel = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	wd.infoPanel.SetBorder(false)
	wd.infoPanel.SetBackgroundColor(ui.ColorBg)

	wd.AddItem(wd.header, 6, 0, false)
	wd.AddItem(wd.tabBar, 1, 0, false)
	wd.AddItem(wd.infoPanel, 0, 1, true)

	// Show loading state initially
	wd.header.SetText(fmt.Sprintf("\n [%s]loading...[-]", ui.TagFgDim))
}

func (wd *WorkflowDetail) setLoading(loading bool) {
	wd.loading = loading
}

func (wd *WorkflowDetail) loadData() {
	provider := wd.app.Provider()
	if provider == nil {
		// Fallback to mock data if no provider
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
		})
	}()
}

func (wd *WorkflowDetail) loadMockData() {
	// Mock data fallback when no provider is configured
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
	wd.render()
}

func (wd *WorkflowDetail) showError(err error) {
	wd.header.SetText(fmt.Sprintf("\n [%s]error: %s[-]", ui.TagFailed, err.Error()))
	wd.infoPanel.SetText("")
	wd.tabBar.SetText("")
}

func (wd *WorkflowDetail) render() {
	if wd.workflow == nil {
		wd.header.SetText(fmt.Sprintf(" [%s]workflow not found[-]", ui.TagFailed))
		return
	}

	w := wd.workflow
	statusColor := ui.StatusColorTag(w.Status)

	// Charm-style: clean, minimal labels
	headerText := fmt.Sprintf(
		"\n [%s]workflow[%s]   %s\n"+
			" [%s]run[%s]        %s\n"+
			" [%s]type[%s]       %s\n"+
			" [%s]status[%s]     [%s]%s[-]\n"+
			" [%s]namespace[%s]  %s",
		ui.TagFgDim, ui.TagFg, w.ID,
		ui.TagFgDim, ui.TagFg, w.RunID,
		ui.TagFgDim, ui.TagFg, w.Type,
		ui.TagFgDim, ui.TagFg, statusColor, w.Status,
		ui.TagFgDim, ui.TagFg, w.Namespace,
	)
	wd.header.SetText(headerText)

	// Tab bar - Charm-style: simple text tabs
	tabText := fmt.Sprintf(" [%s::b]info[-:-:-]   [%s]events (tab)[-]",
		ui.TagAccent, ui.TagFgDim)
	wd.tabBar.SetText(tabText)

	// Info panel content
	wd.renderInfoPanel()
}

func (wd *WorkflowDetail) renderInfoPanel() {
	if wd.workflow == nil {
		return
	}

	w := wd.workflow
	now := time.Now()

	endTimeStr := "-"
	durationStr := "in progress"
	if w.EndTime != nil {
		endTimeStr = w.EndTime.Format("2006-01-02 15:04:05")
		durationStr = w.EndTime.Sub(w.StartTime).Round(time.Second).String()
	}

	parentStr := "-"
	if w.ParentID != nil {
		parentStr = *w.ParentID
	}

	// Charm-style: clean, minimal labels
	infoText := fmt.Sprintf(
		"\n [%s]started[%s]     %s (%s)\n"+
			" [%s]ended[%s]       %s\n"+
			" [%s]duration[%s]    %s\n"+
			" [%s]task queue[%s]  %s\n"+
			" [%s]parent[%s]      %s\n",
		ui.TagFgDim, ui.TagFg, w.StartTime.Format("2006-01-02 15:04:05"), formatRelativeTime(now, w.StartTime),
		ui.TagFgDim, ui.TagFg, endTimeStr,
		ui.TagFgDim, ui.TagFg, durationStr,
		ui.TagFgDim, ui.TagFg, w.TaskQueue,
		ui.TagFgDim, ui.TagFg, parentStr,
	)

	wd.infoPanel.SetText(infoText)
}

// Name returns the view name.
func (wd *WorkflowDetail) Name() string {
	return "workflow-detail"
}

// Start is called when the view becomes active.
func (wd *WorkflowDetail) Start() {
	wd.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'e':
			wd.app.NavigateToEvents(wd.workflowID, wd.runID)
			return nil
		case 'r':
			wd.loadData()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			// Tab navigates to events view
			wd.app.NavigateToEvents(wd.workflowID, wd.runID)
			return nil
		}
		return event
	})
	// Load data when view becomes active
	wd.loadData()
}

// Stop is called when the view is deactivated.
func (wd *WorkflowDetail) Stop() {
	wd.SetInputCapture(nil)
}

// Hints returns keybinding hints for this view.
func (wd *WorkflowDetail) Hints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "e", Description: "Events"},
		{Key: "r", Description: "Refresh"},
		{Key: "tab", Description: "Switch Tab"},
		{Key: "esc", Description: "Back"},
	}
}
