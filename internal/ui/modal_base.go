package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ModalConfig holds configuration for modal creation.
type ModalConfig struct {
	Title     string
	Width     int  // Fixed width, 0 = use MinWidth
	Height    int  // Fixed height, 0 = use MinHeight
	MinWidth  int  // Minimum width
	MaxWidth  int  // Maximum width
	MinHeight int  // Minimum height
	MaxHeight int  // Maximum height
	Backdrop  bool // Use ColorBgDark backdrop
}

// DefaultModalConfig returns sensible defaults for modal configuration.
func DefaultModalConfig() ModalConfig {
	return ModalConfig{
		Width:     50,
		MinWidth:  30,
		MaxWidth:  80,
		MinHeight: 6,
		MaxHeight: 30,
		Backdrop:  true,
	}
}

// Modal is the base component for all modal dialogs.
// It provides consistent centering, backdrop, rounded borders via Panel,
// and a pill-style key hint bar.
type Modal struct {
	*tview.Flex
	panel       *Panel
	hintBar     *KeyHintBar
	content     tview.Primitive
	innerFlex   *tview.Flex // Holds panel + hint bar
	config      ModalConfig
	onClose     func()
	width       int
	height      int
}

// NewModal creates a new modal with the given configuration.
func NewModal(cfg ModalConfig) *Modal {
	// Apply defaults for zero values
	if cfg.MinWidth == 0 {
		cfg.MinWidth = 30
	}
	if cfg.MinHeight == 0 {
		cfg.MinHeight = 6
	}

	m := &Modal{
		Flex:    tview.NewFlex().SetDirection(tview.FlexRow),
		panel:   NewPanel(cfg.Title),
		hintBar: NewKeyHintBar(),
		config:  cfg,
	}

	m.setup()
	return m
}

func (m *Modal) setup() {
	// Calculate dimensions
	m.width = m.config.Width
	if m.width == 0 {
		m.width = m.config.MinWidth
	}
	if m.config.MaxWidth > 0 && m.width > m.config.MaxWidth {
		m.width = m.config.MaxWidth
	}

	m.height = m.config.Height
	if m.height == 0 {
		m.height = m.config.MinHeight
	}
	if m.config.MaxHeight > 0 && m.height > m.config.MaxHeight {
		m.height = m.config.MaxHeight
	}

	// Build inner flex: panel on top, hint bar at bottom
	m.innerFlex = tview.NewFlex().SetDirection(tview.FlexRow)
	m.innerFlex.AddItem(m.panel, 0, 1, true)
	m.innerFlex.AddItem(m.hintBar, 1, 0, false)
	m.innerFlex.SetBackgroundColor(ColorBg())

	m.rebuild()
}

func (m *Modal) rebuild() {
	m.Clear()

	// Total height includes hint bar
	hintHeight := m.hintBar.GetHeight()
	totalHeight := m.height + hintHeight

	// Build centered layout
	m.AddItem(nil, 0, 1, false) // Top spacer
	m.AddItem(tview.NewFlex().
		AddItem(nil, 0, 1, false).             // Left spacer
		AddItem(m.innerFlex, m.width, 0, true). // Centered modal
		AddItem(nil, 0, 1, false),              // Right spacer
		totalHeight, 0, true)
	m.AddItem(nil, 0, 1, false) // Bottom spacer

	if m.config.Backdrop {
		m.SetBackgroundColor(ColorBgDark())
	} else {
		m.SetBackgroundColor(ColorBg())
	}
}

func (m *Modal) applyTheme() {
	m.panel.SetBorderColor(ColorPanelBorder())
	m.panel.SetTitleColor(ColorPanelTitle())
	m.panel.SetBackgroundColor(ColorBg())
	m.innerFlex.SetBackgroundColor(ColorBg())
	m.hintBar.SetBackgroundColor(ColorBg())

	if m.config.Backdrop {
		m.SetBackgroundColor(ColorBgDark())
	} else {
		m.SetBackgroundColor(ColorBg())
	}
}

// Draw applies theme colors dynamically and draws the modal.
func (m *Modal) Draw(screen tcell.Screen) {
	m.applyTheme()
	m.Flex.Draw(screen)
}

// SetContent sets the content primitive inside the modal panel.
func (m *Modal) SetContent(content tview.Primitive) *Modal {
	m.content = content
	m.panel.SetContent(content)
	return m
}

// SetHints sets the key hints to display at the bottom.
func (m *Modal) SetHints(hints []KeyHint) *Modal {
	m.hintBar.SetHints(hints)
	// Rebuild to adjust for hint bar visibility
	m.rebuild()
	return m
}

// SetTitle sets the modal title.
func (m *Modal) SetTitle(title string) *Modal {
	m.config.Title = title
	m.panel.SetTitle(title)
	return m
}

// SetOnClose sets the callback when the modal is closed via Esc or q.
func (m *Modal) SetOnClose(fn func()) *Modal {
	m.onClose = fn
	return m
}

// Close triggers the onClose callback.
func (m *Modal) Close() {
	if m.onClose != nil {
		m.onClose()
	}
}

// SetSize updates the modal dimensions and rebuilds the layout.
func (m *Modal) SetSize(width, height int) *Modal {
	m.width = width
	m.height = height
	m.rebuild()
	return m
}

// GetPanel returns the inner panel for advanced customization.
func (m *Modal) GetPanel() *Panel {
	return m.panel
}

// Focus delegates focus to the content.
func (m *Modal) Focus(delegate func(p tview.Primitive)) {
	if m.content != nil {
		delegate(m.content)
	} else {
		delegate(m.panel)
	}
}

// HasFocus returns whether the modal has focus.
func (m *Modal) HasFocus() bool {
	if m.content != nil {
		return m.content.HasFocus()
	}
	return m.panel.HasFocus()
}

// BaseInputHandler provides common modal key handling.
// Returns true if the event was handled.
func (m *Modal) BaseInputHandler(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyEscape:
		m.Close()
		return true
	case tcell.KeyRune:
		if event.Rune() == 'q' {
			m.Close()
			return true
		}
	}
	return false
}

// WrapInputHandler creates an input handler that first tries the base handler,
// then falls back to the provided custom handler.
func (m *Modal) WrapInputHandler(custom func(event *tcell.EventKey, setFocus func(p tview.Primitive))) func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// Don't use base handler for 'q' in forms where it might be typed
		if event.Key() == tcell.KeyEscape {
			m.Close()
			return
		}
		if custom != nil {
			custom(event, setFocus)
		}
	}
}

// ModalInputHandler returns a basic input handler for modals without custom handling.
func (m *Modal) ModalInputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		m.BaseInputHandler(event)
	}
}
