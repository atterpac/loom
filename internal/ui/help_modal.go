package ui

import (
	"fmt"
	"strings"

	"github.com/atterpac/loom/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HelpModal displays a help overlay with keybindings for the current view.
type HelpModal struct {
	*Modal
	textView  *tview.TextView
	nav       *TextViewNavigator
	viewName  string
	viewHints []KeyHint
}

// NewHelpModal creates a new help modal.
func NewHelpModal() *HelpModal {
	hm := &HelpModal{
		Modal: NewModal(ModalConfig{
			Title:     "Help",
			Width:     45,
			Height:    20,
			MinHeight: 10,
			MaxHeight: 25,
			Backdrop:  true,
		}),
		textView: tview.NewTextView(),
	}
	hm.setup()
	return hm
}

// SetViewHints sets the hints for the current view.
func (hm *HelpModal) SetViewHints(viewName string, hints []KeyHint) *HelpModal {
	hm.viewName = viewName
	hm.viewHints = hints
	hm.rebuildContent()
	return hm
}

// SetOnClose sets the callback when the modal is closed.
func (hm *HelpModal) SetOnClose(fn func()) *HelpModal {
	hm.Modal.SetOnClose(fn)
	return hm
}

func (hm *HelpModal) setup() {
	hm.textView.SetDynamicColors(true)
	hm.textView.SetBackgroundColor(ColorBg())
	hm.textView.SetScrollable(true)

	hm.nav = NewTextViewNavigator(hm.textView)

	hm.rebuildContent()

	hm.SetContent(hm.textView)
	hm.SetHints([]KeyHint{
		{Key: "j/k", Description: "Scroll"},
		{Key: "Esc", Description: "Close"},
	})

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		hm.textView.SetBackgroundColor(ColorBg())
		hm.rebuildContent()
	})
}

func (hm *HelpModal) rebuildContent() {
	var sb strings.Builder

	// Title based on view name
	title := "Keybindings"
	if hm.viewName != "" {
		title = formatViewName(hm.viewName) + " Keybindings"
	}

	sb.WriteString(fmt.Sprintf("\n[%s::b]%s[-:-:-]\n\n", TagAccent(), title))

	// Current view hints
	if len(hm.viewHints) > 0 {
		for _, hint := range hm.viewHints {
			sb.WriteString(fmt.Sprintf("  [%s::b]%-12s[-:-:-] [%s]%s[-]\n",
				TagKey(), hint.Key, TagFgDim(), hint.Description))
		}
	}

	// Always show global hints at the bottom
	sb.WriteString(fmt.Sprintf("\n[%s::b]Global[-:-:-]\n", TagPanelTitle()))
	for _, hint := range globalHints() {
		sb.WriteString(fmt.Sprintf("  [%s::b]%-12s[-:-:-] [%s]%s[-]\n",
			TagKey(), hint.Key, TagFgDim(), hint.Description))
	}

	hm.textView.SetText(sb.String())

	// Adjust height based on content
	lineCount := strings.Count(sb.String(), "\n")
	height := min(lineCount+2, 22)
	if height < 10 {
		height = 10
	}
	hm.SetSize(45, height)
}

// InputHandler handles keyboard input.
func (hm *HelpModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return hm.Flex.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEscape:
			hm.Close()
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', '?':
				hm.Close()
			case 'j':
				hm.nav.MoveDown()
			case 'k':
				hm.nav.MoveUp()
			}
		case tcell.KeyDown:
			hm.nav.MoveDown()
		case tcell.KeyUp:
			hm.nav.MoveUp()
		}
	})
}

// Focus delegates focus to the text view.
func (hm *HelpModal) Focus(delegate func(p tview.Primitive)) {
	delegate(hm.textView)
}

// globalHints returns hints that are always available.
func globalHints() []KeyHint {
	return []KeyHint{
		{Key: "?", Description: "Toggle help"},
		{Key: "T", Description: "Theme selector"},
		{Key: "P", Description: "Profile selector"},
		{Key: ":", Description: "Command bar"},
		{Key: "Esc", Description: "Back / Close"},
		{Key: "j/k", Description: "Navigate"},
	}
}

// formatViewName converts internal view names to display names.
func formatViewName(name string) string {
	switch name {
	case "namespaces":
		return "Namespace List"
	case "workflows":
		return "Workflow List"
	case "workflow-detail":
		return "Workflow Detail"
	case "events":
		return "Event History"
	case "task-queues":
		return "Task Queues"
	case "schedules":
		return "Schedules"
	default:
		return name
	}
}
