package ui

import (
	"fmt"
	"strings"

	"github.com/atterpac/loom/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ConfirmModal displays a confirmation dialog with command preview.
type ConfirmModal struct {
	*Modal
	textView  *tview.TextView
	message   string
	command   string // CLI equivalent to display
	warning   string // Optional warning for destructive ops
	onConfirm func()
	onCancel  func()
}

// NewConfirmModal creates a confirmation modal.
func NewConfirmModal(title, message, command string) *ConfirmModal {
	// Calculate height based on content
	lines := strings.Count(command, "\n") + 1
	height := 7 + lines // Base height for message + command

	cm := &ConfirmModal{
		Modal: NewModal(ModalConfig{
			Title:    title,
			Width:    60,
			Height:   height,
			Backdrop: true,
		}),
		message: message,
		command: command,
	}
	cm.setup()
	return cm
}

// SetWarning adds a warning message for destructive operations.
func (cm *ConfirmModal) SetWarning(warning string) *ConfirmModal {
	cm.warning = warning
	cm.rebuildContent()
	// Adjust height for warning
	lines := strings.Count(cm.command, "\n") + 1
	height := 9 + lines // Extra height for warning
	cm.SetSize(60, height)
	return cm
}

// SetOnConfirm sets the confirmation callback.
func (cm *ConfirmModal) SetOnConfirm(fn func()) *ConfirmModal {
	cm.onConfirm = fn
	return cm
}

// SetOnCancel sets the cancel callback.
func (cm *ConfirmModal) SetOnCancel(fn func()) *ConfirmModal {
	cm.onCancel = fn
	cm.Modal.SetOnClose(fn) // Wire up base modal close
	return cm
}

func (cm *ConfirmModal) setup() {
	cm.textView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(true)
	cm.textView.SetBackgroundColor(ColorBg())

	cm.rebuildContent()
	cm.SetContent(cm.textView)
	cm.SetHints([]KeyHint{
		{Key: "y/Enter", Description: "Confirm"},
		{Key: "n/Esc", Description: "Cancel"},
	})

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		cm.textView.SetBackgroundColor(ColorBg())
		cm.rebuildContent()
	})
}

func (cm *ConfirmModal) rebuildContent() {
	var sb strings.Builder

	// Message
	sb.WriteString(fmt.Sprintf("\n[%s]%s[-]\n\n", TagFg(), cm.message))

	// Warning (if set)
	if cm.warning != "" {
		sb.WriteString(fmt.Sprintf("[%s]%s %s[-]\n\n", TagFailed(), IconFailed, cm.warning))
	}

	// Command preview
	sb.WriteString(fmt.Sprintf("[%s::b]CLI Command:[-:-:-]\n", TagFgDim()))
	sb.WriteString(fmt.Sprintf("[%s]%s[-]", TagAccent(), cm.command))

	cm.textView.SetText(sb.String())
}

// InputHandler handles keyboard input.
func (cm *ConfirmModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return cm.Flex.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEnter:
			if cm.onConfirm != nil {
				cm.onConfirm()
			}
		case tcell.KeyEscape:
			if cm.onCancel != nil {
				cm.onCancel()
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'y', 'Y':
				if cm.onConfirm != nil {
					cm.onConfirm()
				}
			case 'n', 'N', 'q':
				if cm.onCancel != nil {
					cm.onCancel()
				}
			}
		}
	})
}

// Focus delegates focus to the text view.
func (cm *ConfirmModal) Focus(delegate func(p tview.Primitive)) {
	delegate(cm.textView)
}
