package ui

import (
	"github.com/atterpac/loom/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ProfileForm displays a form for creating or editing a connection profile.
type ProfileForm struct {
	*Modal
	form          *tview.Form
	nameInput     *tview.InputField
	addressInput  *tview.InputField
	nsInput       *tview.InputField
	certInput     *tview.InputField
	keyInput      *tview.InputField
	caInput       *tview.InputField
	serverInput   *tview.InputField
	skipVerify    *tview.Checkbox
	isEdit        bool
	originalName  string
	onSave        func(name string, cfg config.ConnectionConfig)
	onCancel      func()
}

// NewProfileForm creates a new profile form.
func NewProfileForm() *ProfileForm {
	pf := &ProfileForm{
		Modal: NewModal(ModalConfig{
			Title:    "New Profile",
			Width:    60,
			Height:   16,
			Backdrop: true,
		}),
		form: tview.NewForm(),
	}
	pf.setup()
	return pf
}

// SetProfile populates the form with an existing profile for editing.
func (pf *ProfileForm) SetProfile(name string, cfg config.ConnectionConfig) *ProfileForm {
	pf.isEdit = true
	pf.originalName = name
	pf.nameInput.SetText(name)
	pf.addressInput.SetText(cfg.Address)
	pf.nsInput.SetText(cfg.Namespace)
	pf.certInput.SetText(cfg.TLS.Cert)
	pf.keyInput.SetText(cfg.TLS.Key)
	pf.caInput.SetText(cfg.TLS.CA)
	pf.serverInput.SetText(cfg.TLS.ServerName)
	pf.skipVerify.SetChecked(cfg.TLS.SkipVerify)
	pf.SetTitle("Edit Profile")
	return pf
}

// SetOnSave sets the callback when the form is saved.
func (pf *ProfileForm) SetOnSave(fn func(name string, cfg config.ConnectionConfig)) *ProfileForm {
	pf.onSave = fn
	return pf
}

// SetOnCancel sets the callback when the form is cancelled.
func (pf *ProfileForm) SetOnCancel(fn func()) *ProfileForm {
	pf.onCancel = fn
	pf.Modal.SetOnClose(fn)
	return pf
}

func (pf *ProfileForm) setup() {
	pf.form.SetBackgroundColor(ColorBg())
	pf.form.SetFieldBackgroundColor(ColorBgLight())
	pf.form.SetFieldTextColor(ColorFg())
	pf.form.SetLabelColor(ColorFgDim())
	pf.form.SetButtonBackgroundColor(ColorAccent())
	pf.form.SetButtonTextColor(ColorBg())

	// Create input fields
	pf.nameInput = tview.NewInputField().
		SetLabel("Name: ").
		SetFieldWidth(30).
		SetPlaceholder("my-profile")
	pf.nameInput.SetFieldBackgroundColor(ColorBgLight())
	pf.nameInput.SetFieldTextColor(ColorFg())
	pf.nameInput.SetLabelColor(ColorFgDim())
	pf.nameInput.SetPlaceholderTextColor(ColorFgDim())

	pf.addressInput = tview.NewInputField().
		SetLabel("Address: ").
		SetFieldWidth(30).
		SetPlaceholder("localhost:7233")
	pf.addressInput.SetFieldBackgroundColor(ColorBgLight())
	pf.addressInput.SetFieldTextColor(ColorFg())
	pf.addressInput.SetLabelColor(ColorFgDim())
	pf.addressInput.SetPlaceholderTextColor(ColorFgDim())

	pf.nsInput = tview.NewInputField().
		SetLabel("Namespace: ").
		SetFieldWidth(30).
		SetPlaceholder("default")
	pf.nsInput.SetFieldBackgroundColor(ColorBgLight())
	pf.nsInput.SetFieldTextColor(ColorFg())
	pf.nsInput.SetLabelColor(ColorFgDim())
	pf.nsInput.SetPlaceholderTextColor(ColorFgDim())

	pf.certInput = tview.NewInputField().
		SetLabel("TLS Cert: ").
		SetFieldWidth(30).
		SetPlaceholder("/path/to/cert.pem (optional)")
	pf.certInput.SetFieldBackgroundColor(ColorBgLight())
	pf.certInput.SetFieldTextColor(ColorFg())
	pf.certInput.SetLabelColor(ColorFgDim())
	pf.certInput.SetPlaceholderTextColor(ColorFgDim())

	pf.keyInput = tview.NewInputField().
		SetLabel("TLS Key: ").
		SetFieldWidth(30).
		SetPlaceholder("/path/to/key.pem (optional)")
	pf.keyInput.SetFieldBackgroundColor(ColorBgLight())
	pf.keyInput.SetFieldTextColor(ColorFg())
	pf.keyInput.SetLabelColor(ColorFgDim())
	pf.keyInput.SetPlaceholderTextColor(ColorFgDim())

	pf.caInput = tview.NewInputField().
		SetLabel("TLS CA: ").
		SetFieldWidth(30).
		SetPlaceholder("/path/to/ca.pem (optional)")
	pf.caInput.SetFieldBackgroundColor(ColorBgLight())
	pf.caInput.SetFieldTextColor(ColorFg())
	pf.caInput.SetLabelColor(ColorFgDim())
	pf.caInput.SetPlaceholderTextColor(ColorFgDim())

	pf.serverInput = tview.NewInputField().
		SetLabel("TLS Server: ").
		SetFieldWidth(30).
		SetPlaceholder("server.example.com (optional)")
	pf.serverInput.SetFieldBackgroundColor(ColorBgLight())
	pf.serverInput.SetFieldTextColor(ColorFg())
	pf.serverInput.SetLabelColor(ColorFgDim())
	pf.serverInput.SetPlaceholderTextColor(ColorFgDim())

	pf.skipVerify = tview.NewCheckbox().
		SetLabel("Skip TLS Verify: ")
	pf.skipVerify.SetBackgroundColor(ColorBg())
	pf.skipVerify.SetLabelColor(ColorFgDim())
	pf.skipVerify.SetFieldBackgroundColor(ColorBg())

	// Add items to form
	pf.form.AddFormItem(pf.nameInput)
	pf.form.AddFormItem(pf.addressInput)
	pf.form.AddFormItem(pf.nsInput)
	pf.form.AddFormItem(pf.certInput)
	pf.form.AddFormItem(pf.keyInput)
	pf.form.AddFormItem(pf.caInput)
	pf.form.AddFormItem(pf.serverInput)
	pf.form.AddFormItem(pf.skipVerify)

	pf.SetContent(pf.form)
	pf.SetHints([]KeyHint{
		{Key: "Tab", Description: "Next"},
		{Key: "Ctrl+S", Description: "Save"},
		{Key: "Esc", Description: "Cancel"},
	})

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		pf.form.SetBackgroundColor(ColorBg())
		pf.form.SetFieldBackgroundColor(ColorBgLight())
		pf.form.SetFieldTextColor(ColorFg())
		pf.form.SetLabelColor(ColorFgDim())
		pf.form.SetButtonBackgroundColor(ColorAccent())
		pf.form.SetButtonTextColor(ColorBg())

		for _, input := range []*tview.InputField{pf.nameInput, pf.addressInput, pf.nsInput, pf.certInput, pf.keyInput, pf.caInput, pf.serverInput} {
			input.SetFieldBackgroundColor(ColorBgLight())
			input.SetFieldTextColor(ColorFg())
			input.SetLabelColor(ColorFgDim())
			input.SetPlaceholderTextColor(ColorFgDim())
		}
		pf.skipVerify.SetBackgroundColor(ColorBg())
		pf.skipVerify.SetLabelColor(ColorFgDim())
	})
}

func (pf *ProfileForm) save() {
	if pf.onSave == nil {
		return
	}

	name := pf.nameInput.GetText()
	if name == "" {
		return
	}

	cfg := config.ConnectionConfig{
		Address:   pf.addressInput.GetText(),
		Namespace: pf.nsInput.GetText(),
		TLS: config.TLSConfig{
			Cert:       pf.certInput.GetText(),
			Key:        pf.keyInput.GetText(),
			CA:         pf.caInput.GetText(),
			ServerName: pf.serverInput.GetText(),
			SkipVerify: pf.skipVerify.IsChecked(),
		},
	}

	// Set defaults if empty
	if cfg.Address == "" {
		cfg.Address = "localhost:7233"
	}
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	pf.onSave(name, cfg)
}

// ClearFields resets the form for a new profile.
func (pf *ProfileForm) ClearFields() *ProfileForm {
	pf.isEdit = false
	pf.originalName = ""
	pf.nameInput.SetText("")
	pf.addressInput.SetText("")
	pf.nsInput.SetText("")
	pf.certInput.SetText("")
	pf.keyInput.SetText("")
	pf.caInput.SetText("")
	pf.serverInput.SetText("")
	pf.skipVerify.SetChecked(false)
	pf.SetTitle("New Profile")
	return pf
}

// InputHandler handles keyboard input.
func (pf *ProfileForm) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pf.Flex.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// Handle special keys first
		switch event.Key() {
		case tcell.KeyEscape:
			if pf.onCancel != nil {
				pf.onCancel()
			}
			return
		case tcell.KeyCtrlS:
			pf.save()
			return
		}

		// Pass everything else to the form
		if handler := pf.form.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

// Focus sets focus to the form.
func (pf *ProfileForm) Focus(delegate func(p tview.Primitive)) {
	delegate(pf.form)
}
