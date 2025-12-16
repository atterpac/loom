package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/atterpac/loom/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NamespaceFormData holds the form field values.
type NamespaceFormData struct {
	Name          string
	RetentionDays int
	Description   string
	OwnerEmail    string
}

// NamespaceForm displays a modal for creating or editing namespaces.
type NamespaceForm struct {
	*Modal
	isEdit       bool
	original     string // Original namespace name when editing
	nameDisabled bool   // Track if name field is disabled

	// Form fields
	nameInput        *tview.InputField
	retentionInput   *tview.InputField
	descriptionInput *tview.InputField
	ownerEmailInput  *tview.InputField

	// Internal components
	form *tview.Form

	// Callbacks
	onSubmit func(data NamespaceFormData)
	onCancel func()
}

// NewNamespaceForm creates a new namespace form modal.
func NewNamespaceForm() *NamespaceForm {
	nf := &NamespaceForm{
		Modal: NewModal(ModalConfig{
			Title:    "Create Namespace",
			Width:    60,
			Height:   12,
			Backdrop: true,
		}),
	}
	nf.setup()
	return nf
}

// SetOnSubmit sets the submission callback.
func (nf *NamespaceForm) SetOnSubmit(fn func(data NamespaceFormData)) *NamespaceForm {
	nf.onSubmit = fn
	return nf
}

// SetOnCancel sets the cancel callback.
func (nf *NamespaceForm) SetOnCancel(fn func()) *NamespaceForm {
	nf.onCancel = fn
	nf.Modal.SetOnClose(fn)
	return nf
}

// SetNamespace populates the form for editing an existing namespace.
func (nf *NamespaceForm) SetNamespace(name string, retentionDays int, description, ownerEmail string) {
	nf.isEdit = true
	nf.original = name
	nf.nameDisabled = true

	nf.nameInput.SetText(name)
	nf.retentionInput.SetText(strconv.Itoa(retentionDays))
	nf.descriptionInput.SetText(description)
	nf.ownerEmailInput.SetText(ownerEmail)

	// Update title
	nf.SetTitle("Edit Namespace")

	// Disable name field when editing (visual only - we ignore input)
	nf.nameInput.SetFieldBackgroundColor(ColorBgDark())
}

// ClearFields resets all form fields.
func (nf *NamespaceForm) ClearFields() {
	nf.isEdit = false
	nf.original = ""
	nf.nameDisabled = false

	nf.nameInput.SetText("")
	nf.retentionInput.SetText("30")
	nf.descriptionInput.SetText("")
	nf.ownerEmailInput.SetText("")

	// Update title
	nf.SetTitle("Create Namespace")

	// Enable name field
	nf.nameInput.SetFieldBackgroundColor(ColorBgLight())
}

func (nf *NamespaceForm) setup() {
	nf.form = tview.NewForm()
	nf.form.SetBackgroundColor(ColorBg())
	nf.form.SetFieldBackgroundColor(ColorBgLight())
	nf.form.SetFieldTextColor(ColorFg())
	nf.form.SetLabelColor(ColorFgDim())
	nf.form.SetButtonBackgroundColor(ColorAccent())
	nf.form.SetButtonTextColor(ColorBg())
	nf.form.SetBorderPadding(0, 0, 1, 1)

	// Name field (required)
	nf.nameInput = tview.NewInputField().
		SetLabel("Name *: ").
		SetPlaceholder("my-namespace").
		SetFieldWidth(40).
		SetFieldBackgroundColor(ColorBgLight()).
		SetFieldTextColor(ColorFg()).
		SetLabelColor(ColorFgDim()).
		SetPlaceholderTextColor(ColorFgDim())
	nf.form.AddFormItem(nf.nameInput)

	// Retention field (required)
	nf.retentionInput = tview.NewInputField().
		SetLabel("Retention Days *: ").
		SetPlaceholder("30").
		SetText("30").
		SetFieldWidth(10).
		SetFieldBackgroundColor(ColorBgLight()).
		SetFieldTextColor(ColorFg()).
		SetLabelColor(ColorFgDim()).
		SetPlaceholderTextColor(ColorFgDim()).
		SetAcceptanceFunc(func(text string, lastChar rune) bool {
			// Only accept digits
			return lastChar >= '0' && lastChar <= '9'
		})
	nf.form.AddFormItem(nf.retentionInput)

	// Description field (optional)
	nf.descriptionInput = tview.NewInputField().
		SetLabel("Description: ").
		SetPlaceholder("Optional description").
		SetFieldWidth(40).
		SetFieldBackgroundColor(ColorBgLight()).
		SetFieldTextColor(ColorFg()).
		SetLabelColor(ColorFgDim()).
		SetPlaceholderTextColor(ColorFgDim())
	nf.form.AddFormItem(nf.descriptionInput)

	// Owner Email field (optional)
	nf.ownerEmailInput = tview.NewInputField().
		SetLabel("Owner Email: ").
		SetPlaceholder("owner@example.com").
		SetFieldWidth(40).
		SetFieldBackgroundColor(ColorBgLight()).
		SetFieldTextColor(ColorFg()).
		SetLabelColor(ColorFgDim()).
		SetPlaceholderTextColor(ColorFgDim())
	nf.form.AddFormItem(nf.ownerEmailInput)

	nf.SetContent(nf.form)
	nf.SetHints([]KeyHint{
		{Key: "Tab", Description: "Next"},
		{Key: "Ctrl+S", Description: "Save"},
		{Key: "Esc", Description: "Cancel"},
	})

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		nf.form.SetBackgroundColor(ColorBg())
		nf.form.SetFieldBackgroundColor(ColorBgLight())
		nf.form.SetFieldTextColor(ColorFg())
		nf.form.SetLabelColor(ColorFgDim())

		// Update field colors
		for _, field := range []*tview.InputField{nf.retentionInput, nf.descriptionInput, nf.ownerEmailInput} {
			field.SetFieldBackgroundColor(ColorBgLight())
			field.SetFieldTextColor(ColorFg())
			field.SetLabelColor(ColorFgDim())
			field.SetPlaceholderTextColor(ColorFgDim())
		}

		// Handle name field separately (may be disabled)
		if nf.nameDisabled {
			nf.nameInput.SetFieldBackgroundColor(ColorBgDark())
		} else {
			nf.nameInput.SetFieldBackgroundColor(ColorBgLight())
		}
		nf.nameInput.SetFieldTextColor(ColorFg())
		nf.nameInput.SetLabelColor(ColorFgDim())
		nf.nameInput.SetPlaceholderTextColor(ColorFgDim())
	})
}

// GetData returns the current form data.
func (nf *NamespaceForm) GetData() NamespaceFormData {
	retention, _ := strconv.Atoi(nf.retentionInput.GetText())
	if retention < 1 {
		retention = 1
	}

	name := nf.nameInput.GetText()
	if nf.isEdit {
		name = nf.original // Use original name when editing
	}

	return NamespaceFormData{
		Name:          strings.TrimSpace(name),
		RetentionDays: retention,
		Description:   strings.TrimSpace(nf.descriptionInput.GetText()),
		OwnerEmail:    strings.TrimSpace(nf.ownerEmailInput.GetText()),
	}
}

// Validate checks if required fields have valid values.
func (nf *NamespaceForm) Validate() error {
	name := strings.TrimSpace(nf.nameInput.GetText())
	if !nf.isEdit && name == "" {
		return fmt.Errorf("namespace name is required")
	}

	retentionStr := strings.TrimSpace(nf.retentionInput.GetText())
	if retentionStr == "" {
		return fmt.Errorf("retention days is required")
	}

	retention, err := strconv.Atoi(retentionStr)
	if err != nil {
		return fmt.Errorf("retention days must be a number")
	}
	if retention < 1 {
		return fmt.Errorf("retention days must be at least 1")
	}

	return nil
}

// InputHandler handles keyboard input.
func (nf *NamespaceForm) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return nf.Flex.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// Handle special keys first
		switch event.Key() {
		case tcell.KeyEscape:
			if nf.onCancel != nil {
				nf.onCancel()
			}
			return
		case tcell.KeyCtrlS:
			if err := nf.Validate(); err == nil {
				if nf.onSubmit != nil {
					nf.onSubmit(nf.GetData())
				}
			}
			return
		}

		// Pass everything else to the form
		if handler := nf.form.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

// Focus returns the focusable component.
func (nf *NamespaceForm) Focus(delegate func(p tview.Primitive)) {
	// Focus the form - it will manage field focus
	delegate(nf.form)
}
