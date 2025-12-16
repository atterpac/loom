package ui

import (
	"fmt"
	"strings"

	"github.com/atterpac/loom/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InputField represents a single input field configuration.
type InputField struct {
	Name        string
	Label       string
	Placeholder string
	Required    bool
}

// InputModal displays a modal with one or more input fields.
type InputModal struct {
	*Modal
	message    string
	fields     []InputField
	inputForms []*tview.InputField
	onSubmit   func(values map[string]string)
	onCancel   func()
	focusIndex int

	// Internal components
	form *tview.Form
}

// NewInputModal creates an input modal with the specified fields.
func NewInputModal(title, message string, fields []InputField) *InputModal {
	// Calculate height based on fields
	height := 5 + len(fields)*2
	if message != "" {
		height += 2
	}

	im := &InputModal{
		Modal: NewModal(ModalConfig{
			Title:    title,
			Width:    60,
			Height:   height,
			Backdrop: true,
		}),
		message: message,
		fields:  fields,
	}
	im.setup()
	return im
}

// SetOnSubmit sets the submission callback.
func (im *InputModal) SetOnSubmit(fn func(values map[string]string)) *InputModal {
	im.onSubmit = fn
	return im
}

// SetOnCancel sets the cancel callback.
func (im *InputModal) SetOnCancel(fn func()) *InputModal {
	im.onCancel = fn
	im.Modal.SetOnClose(fn)
	return im
}

func (im *InputModal) setup() {
	im.form = tview.NewForm()
	im.form.SetBackgroundColor(ColorBg())
	im.form.SetFieldBackgroundColor(ColorBgLight())
	im.form.SetFieldTextColor(ColorFg())
	im.form.SetLabelColor(ColorFgDim())
	im.form.SetButtonBackgroundColor(ColorAccent())
	im.form.SetButtonTextColor(ColorBg())
	im.form.SetBorderPadding(0, 0, 1, 1)

	// Add message at top if present
	if im.message != "" {
		messageView := tview.NewTextView().
			SetDynamicColors(true).
			SetText(fmt.Sprintf("[%s]%s[-]", TagFg(), im.message))
		messageView.SetBackgroundColor(ColorBg())
		im.form.AddFormItem(messageView)
	}

	// Add input fields
	im.inputForms = make([]*tview.InputField, len(im.fields))
	for i, field := range im.fields {
		label := field.Label
		if field.Required {
			label += " *"
		}
		inputField := tview.NewInputField().
			SetLabel(label + ": ").
			SetPlaceholder(field.Placeholder).
			SetFieldWidth(40).
			SetFieldBackgroundColor(ColorBgLight()).
			SetFieldTextColor(ColorFg()).
			SetLabelColor(ColorFgDim()).
			SetPlaceholderTextColor(ColorFgDim())
		im.form.AddFormItem(inputField)
		im.inputForms[i] = inputField
	}

	im.SetContent(im.form)
	im.SetHints([]KeyHint{
		{Key: "Tab", Description: "Next field"},
		{Key: "Enter", Description: "Submit"},
		{Key: "Esc", Description: "Cancel"},
	})

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		im.form.SetBackgroundColor(ColorBg())
		im.form.SetFieldBackgroundColor(ColorBgLight())
		im.form.SetFieldTextColor(ColorFg())
		im.form.SetLabelColor(ColorFgDim())
		for _, input := range im.inputForms {
			input.SetFieldBackgroundColor(ColorBgLight())
			input.SetFieldTextColor(ColorFg())
			input.SetLabelColor(ColorFgDim())
			input.SetPlaceholderTextColor(ColorFgDim())
		}
	})
}

// GetValues returns the current values of all input fields.
func (im *InputModal) GetValues() map[string]string {
	values := make(map[string]string)
	for i, field := range im.fields {
		if i < len(im.inputForms) {
			values[field.Name] = im.inputForms[i].GetText()
		}
	}
	return values
}

// Validate checks if all required fields have values.
func (im *InputModal) Validate() error {
	for i, field := range im.fields {
		if field.Required && i < len(im.inputForms) {
			value := strings.TrimSpace(im.inputForms[i].GetText())
			if value == "" {
				return fmt.Errorf("%s is required", field.Label)
			}
		}
	}
	return nil
}

// InputHandler handles keyboard input.
func (im *InputModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return im.Flex.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// Handle special keys first
		switch event.Key() {
		case tcell.KeyEscape:
			if im.onCancel != nil {
				im.onCancel()
			}
			return
		case tcell.KeyEnter:
			// Submit if validation passes
			if err := im.Validate(); err == nil {
				if im.onSubmit != nil {
					im.onSubmit(im.GetValues())
				}
			}
			return
		}

		// Pass everything else to the form
		if handler := im.form.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

// Focus focuses the form.
func (im *InputModal) Focus(delegate func(p tview.Primitive)) {
	delegate(im.form)
}
