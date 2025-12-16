package ui

import (
	"fmt"

	"github.com/atterpac/loom/internal/config"
	"github.com/rivo/tview"
)

// EmptyState displays a centered message when there is no data to show.
type EmptyState struct {
	*tview.Flex
	textView *tview.TextView
	icon     string
	title    string
	message  string
}

// NewEmptyState creates a new empty state component.
func NewEmptyState(icon, title, message string) *EmptyState {
	es := &EmptyState{
		Flex:     tview.NewFlex().SetDirection(tview.FlexRow),
		textView: tview.NewTextView(),
		icon:     icon,
		title:    title,
		message:  message,
	}
	es.setup()
	return es
}

func (es *EmptyState) setup() {
	es.textView.SetDynamicColors(true)
	es.textView.SetTextAlign(tview.AlignCenter)
	es.textView.SetBackgroundColor(ColorBg())

	es.render()

	// Build centered layout
	es.AddItem(nil, 0, 1, false) // Top spacer
	es.AddItem(es.textView, 5, 0, false)
	es.AddItem(nil, 0, 1, false) // Bottom spacer
	es.SetBackgroundColor(ColorBg())

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		es.SetBackgroundColor(ColorBg())
		es.textView.SetBackgroundColor(ColorBg())
		es.render()
	})
}

func (es *EmptyState) render() {
	text := fmt.Sprintf(`[%s]%s[-]

[%s::b]%s[-:-:-]

[%s]%s[-]`,
		TagFgDim(), es.icon,
		TagFg(), es.title,
		TagFgDim(), es.message)

	es.textView.SetText(text)
}

// SetIcon updates the icon.
func (es *EmptyState) SetIcon(icon string) *EmptyState {
	es.icon = icon
	es.render()
	return es
}

// SetTitle updates the title.
func (es *EmptyState) SetTitle(title string) *EmptyState {
	es.title = title
	es.render()
	return es
}

// SetMessage updates the message.
func (es *EmptyState) SetMessage(message string) *EmptyState {
	es.message = message
	es.render()
	return es
}

// Common empty state presets

// EmptyStateNoNamespaces creates an empty state for no namespaces.
func EmptyStateNoNamespaces() *EmptyState {
	return NewEmptyState(
		IconNamespace,
		"No Namespaces",
		"No namespaces found. Check your connection.",
	)
}

// EmptyStateNoWorkflows creates an empty state for no workflows.
func EmptyStateNoWorkflows() *EmptyState {
	return NewEmptyState(
		IconWorkflow,
		"No Workflows",
		"No workflows found in this namespace.",
	)
}

// EmptyStateNoResults creates an empty state for no filter results.
func EmptyStateNoResults() *EmptyState {
	return NewEmptyState(
		IconDash,
		"No Results",
		"No items match your filter. Try a different search.",
	)
}

// EmptyStateNoTaskQueues creates an empty state for no task queues.
func EmptyStateNoTaskQueues() *EmptyState {
	return NewEmptyState(
		IconTaskQueue,
		"No Task Queues",
		"No task queues found in this namespace.",
	)
}

// EmptyStateNoSchedules creates an empty state for no schedules.
func EmptyStateNoSchedules() *EmptyState {
	return NewEmptyState(
		IconActivity,
		"No Schedules",
		"No schedules found in this namespace.",
	)
}

// EmptyStateNoEvents creates an empty state for no events.
func EmptyStateNoEvents() *EmptyState {
	return NewEmptyState(
		IconEvent,
		"No Events",
		"No events found for this workflow.",
	)
}

// EmptyStateLoading creates a loading state.
func EmptyStateLoading() *EmptyState {
	return NewEmptyState(
		IconActivity,
		"Loading...",
		"Fetching data from Temporal server.",
	)
}

// EmptyStateDisconnected creates a disconnected state.
func EmptyStateDisconnected() *EmptyState {
	return NewEmptyState(
		IconDisconnected,
		"Disconnected",
		"Unable to connect to Temporal server.",
	)
}
