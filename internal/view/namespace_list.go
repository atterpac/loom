package view

import (
	"context"
	"fmt"
	"time"

	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
	"github.com/atterpac/tempo/internal/temporal"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NamespaceList displays a list of Temporal namespaces with a preview panel.
type NamespaceList struct {
	*tview.Flex
	table         *components.Table
	leftPanel     *components.Panel
	rightPanel    *components.Panel
	preview       *tview.TextView
	emptyState    *components.EmptyState
	app           *App
	namespaces    []temporal.Namespace
	loading       bool
	autoRefresh   bool
	showPreview   bool
	refreshTicker *time.Ticker
	stopRefresh   chan struct{}
}

// NewNamespaceList creates a new namespace list view.
func NewNamespaceList(app *App) *NamespaceList {
	nl := &NamespaceList{
		Flex:        tview.NewFlex().SetDirection(tview.FlexColumn),
		table:       components.NewTable(),
		preview:     tview.NewTextView(),
		app:         app,
		namespaces:  []temporal.Namespace{},
		showPreview: true,
		stopRefresh: make(chan struct{}),
	}
	nl.setup()
	return nl
}

func (nl *NamespaceList) setup() {
	nl.table.SetHeaders("NAME", "STATE", "RETENTION")
	nl.table.SetBorder(false)
	nl.table.SetBackgroundColor(theme.Bg())
	nl.SetBackgroundColor(theme.Bg())

	// Configure preview
	nl.preview.SetDynamicColors(true)
	nl.preview.SetBackgroundColor(theme.Bg())
	nl.preview.SetTextColor(theme.Fg())
	nl.preview.SetWordWrap(true)

	// Create empty state
	nl.emptyState = components.NewEmptyState().
		SetIcon(theme.IconDatabase).
		SetTitle("No Namespaces").
		SetMessage("No namespaces found")

	// Create panels with icons (blubber pattern)
	nl.leftPanel = components.NewPanel().SetTitle(fmt.Sprintf("%s Namespaces", theme.IconNamespace))
	nl.leftPanel.SetContent(nl.table)

	nl.rightPanel = components.NewPanel().SetTitle(fmt.Sprintf("%s Details", theme.IconInfo))
	nl.rightPanel.SetContent(nl.preview)

	// Selection change handler to update preview and hints
	nl.table.SetSelectionChangedFunc(func(row, col int) {
		dataRow := row - 1
		if dataRow >= 0 && dataRow < len(nl.namespaces) {
			nl.updatePreview(nl.namespaces[dataRow])
			nl.app.JigApp().Menu().SetHints(nl.Hints())
		}
	})

	// Selection handler - Enter navigates to workflows
	nl.table.SetOnSelect(func(row int) {
		if row >= 0 && row < len(nl.namespaces) {
			nl.app.NavigateToWorkflows(nl.namespaces[row].Name)
		}
	})

	nl.buildLayout()
}

func (nl *NamespaceList) buildLayout() {
	nl.Clear()
	if nl.showPreview {
		nl.AddItem(nl.leftPanel, 0, 3, true)
		nl.AddItem(nl.rightPanel, 0, 2, false)
	} else {
		nl.AddItem(nl.leftPanel, 0, 1, true)
	}
}

func (nl *NamespaceList) togglePreview() {
	nl.showPreview = !nl.showPreview
	nl.buildLayout()
}

// RefreshTheme updates all component colors after a theme change.
func (nl *NamespaceList) RefreshTheme() {
	bg := theme.Bg()

	// Update main container
	nl.SetBackgroundColor(bg)

	// Update table
	nl.table.SetBackgroundColor(bg)

	// Update preview
	nl.preview.SetBackgroundColor(bg)
	nl.preview.SetTextColor(theme.Fg())

	// Re-render table with new theme colors
	nl.populateTable()
}

func (nl *NamespaceList) updatePreview(ns temporal.Namespace) {
	stateIcon := theme.IconConnected
	stateColor := theme.StatusColorTag("Running")
	if ns.State == "Deprecated" {
		stateIcon = theme.IconDisconnected
		stateColor = theme.StatusColorTag("Failed")
	}

	text := fmt.Sprintf(`[%s::b]Name[-:-:-]
  [%s]%s[-]

[%s::b]State[-:-:-]
  [%s]%s %s[-]

[%s::b]Retention[-:-:-]
  [%s]%s[-]

[%s::b]Description[-:-:-]
  [%s]%s[-]

[%s::b]Owner[-:-:-]
  [%s]%s[-]`,
		theme.TagFgDim(),
		theme.TagFg(), ns.Name,
		theme.TagFgDim(),
		stateColor, stateIcon, ns.State,
		theme.TagFgDim(),
		theme.TagFg(), ns.RetentionPeriod,
		theme.TagFgDim(),
		theme.TagFg(), valueOrEmpty(ns.Description, "No description"),
		theme.TagFgDim(),
		theme.TagFg(), valueOrEmpty(ns.OwnerEmail, "No owner"),
	)
	nl.preview.SetText(text)
}

func valueOrEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func (nl *NamespaceList) setLoading(loading bool) {
	nl.loading = loading
}

func (nl *NamespaceList) loadData() {
	provider := nl.app.Provider()
	if provider == nil {
		nl.loadMockData()
		return
	}

	nl.setLoading(true)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		namespaces, err := provider.ListNamespaces(ctx)

		nl.app.JigApp().QueueUpdateDraw(func() {
			nl.setLoading(false)
			if err != nil {
				nl.showError(err)
				return
			}
			nl.namespaces = namespaces
			nl.populateTable()
		})
	}()
}

func (nl *NamespaceList) loadMockData() {
	nl.namespaces = []temporal.Namespace{
		{Name: "default", State: "Active", RetentionPeriod: "7 days"},
		{Name: "production", State: "Active", RetentionPeriod: "30 days"},
		{Name: "staging", State: "Active", RetentionPeriod: "3 days"},
		{Name: "development", State: "Active", RetentionPeriod: "1 day"},
		{Name: "archived", State: "Deprecated", RetentionPeriod: "90 days"},
	}
	nl.populateTable()
}

func (nl *NamespaceList) populateTable() {
	currentRow := nl.table.SelectedRow()

	nl.table.ClearRows()
	nl.table.SetHeaders("NAME", "STATE", "RETENTION")

	if len(nl.namespaces) == 0 {
		nl.leftPanel.SetContent(nl.emptyState)
		nl.preview.SetText("")
		return
	}

	nl.leftPanel.SetContent(nl.table)

	for _, ns := range nl.namespaces {
		nl.table.AddStyledRowSimple(ns.State,
			theme.IconDatabase+" "+ns.Name,
			ns.State,
			ns.RetentionPeriod,
		)
	}

	if nl.table.RowCount() > 0 {
		if currentRow >= 0 && currentRow < len(nl.namespaces) {
			nl.table.SelectRow(currentRow)
			nl.updatePreview(nl.namespaces[currentRow])
		} else {
			nl.table.SelectRow(0)
			if len(nl.namespaces) > 0 {
				nl.updatePreview(nl.namespaces[0])
			}
		}
	}
}

func (nl *NamespaceList) showError(err error) {
	nl.table.ClearRows()
	nl.table.SetHeaders("NAME", "STATE", "RETENTION")
	nl.table.AddRowWithColor(theme.Error(),
		theme.IconError+" Error loading namespaces",
		err.Error(),
		"",
	)
}

func (nl *NamespaceList) toggleAutoRefresh() {
	nl.autoRefresh = !nl.autoRefresh
	if nl.autoRefresh {
		nl.startAutoRefresh()
	} else {
		nl.stopAutoRefresh()
	}
}

func (nl *NamespaceList) startAutoRefresh() {
	nl.refreshTicker = time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-nl.refreshTicker.C:
				nl.app.JigApp().QueueUpdateDraw(func() {
					nl.loadData()
				})
			case <-nl.stopRefresh:
				return
			}
		}
	}()
}

func (nl *NamespaceList) stopAutoRefresh() {
	if nl.refreshTicker != nil {
		nl.refreshTicker.Stop()
		nl.refreshTicker = nil
	}
	select {
	case nl.stopRefresh <- struct{}{}:
	default:
	}
}

// Name returns the view name.
func (nl *NamespaceList) Name() string {
	return "namespaces"
}

// Start is called when the view becomes active.
func (nl *NamespaceList) Start() {
	nl.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			nl.app.Stop()
			return nil
		case 'a':
			nl.toggleAutoRefresh()
			return nil
		case 'r':
			nl.loadData()
			return nil
		case 'p':
			nl.togglePreview()
			return nil
		case 'i':
			ns := nl.getSelectedNamespace()
			if ns != nil {
				nl.app.NavigateToNamespaceDetail(ns.Name)
			}
			return nil
		case 'n':
			// TODO: Create namespace form
			return nil
		case 'e':
			// TODO: Edit namespace form
			return nil
		case 'D':
			// TODO: Deprecate confirm
			return nil
		case 'X':
			// TODO: Delete confirm
			return nil
		}
		return event
	})
	nl.loadData()
}

// Stop is called when the view is deactivated.
func (nl *NamespaceList) Stop() {
	nl.table.SetInputCapture(nil)
	nl.stopAutoRefresh()
}

// Hints returns keybinding hints for this view.
func (nl *NamespaceList) Hints() []KeyHint {
	hints := []KeyHint{
		{Key: "enter", Description: "Workflows"},
		{Key: "i", Description: "Info"},
		{Key: "n", Description: "Create"},
		{Key: "e", Description: "Edit"},
	}

	ns := nl.getSelectedNamespace()
	if ns != nil && ns.State == "Deprecated" {
		hints = append(hints, KeyHint{Key: "X", Description: "Delete"})
	} else {
		hints = append(hints, KeyHint{Key: "D", Description: "Deprecate"})
	}

	hints = append(hints,
		KeyHint{Key: "p", Description: "Preview"},
		KeyHint{Key: "r", Description: "Refresh"},
		KeyHint{Key: "a", Description: "Auto-refresh"},
		KeyHint{Key: "T", Description: "Theme"},
		KeyHint{Key: "?", Description: "Help"},
		KeyHint{Key: "q", Description: "Quit"},
	)
	return hints
}

// Focus sets focus to the table.
func (nl *NamespaceList) Focus(delegate func(p tview.Primitive)) {
	if len(nl.namespaces) == 0 {
		delegate(nl.Flex)
		return
	}
	delegate(nl.table)
}

// Draw applies theme colors dynamically and draws the view.
func (nl *NamespaceList) Draw(screen tcell.Screen) {
	bg := theme.Bg()
	nl.SetBackgroundColor(bg)
	nl.preview.SetBackgroundColor(bg)
	nl.preview.SetTextColor(theme.Fg())
	nl.Flex.Draw(screen)
}

// getSelectedNamespace returns the currently selected namespace.
func (nl *NamespaceList) getSelectedNamespace() *temporal.Namespace {
	row := nl.table.SelectedRow()
	if row >= 0 && row < len(nl.namespaces) {
		return &nl.namespaces[row]
	}
	return nil
}
