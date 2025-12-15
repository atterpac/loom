package ui

import (
	"fmt"
	"strings"

	"github.com/atterpac/temportui/internal/config"
	"github.com/rivo/tview"
)

// Menu displays keybinding hints at the bottom of the screen.
type Menu struct {
	*tview.TextView
	hints []KeyHint
}

// NewMenu creates a new menu component.
func NewMenu() *Menu {
	m := &Menu{
		TextView: tview.NewTextView(),
		hints:    []KeyHint{},
	}
	m.SetDynamicColors(true)
	m.applyTheme()
	m.render()

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		m.applyTheme()
		m.render()
	})

	return m
}

// applyTheme applies the current theme colors to the menu.
func (m *Menu) applyTheme() {
	m.SetBackgroundColor(ColorMenu())
	m.SetTextColor(ColorFg())
}

// SetHints sets the keybinding hints to display.
func (m *Menu) SetHints(hints []KeyHint) {
	m.hints = hints
	m.render()
}

// AddHint adds a single hint.
func (m *Menu) AddHint(hint KeyHint) {
	m.hints = append(m.hints, hint)
	m.render()
}

// Clear removes all hints.
func (m *Menu) Clear() {
	m.hints = []KeyHint{}
	m.render()
}

func (m *Menu) render() {
	if len(m.hints) == 0 {
		m.SetText("")
		return
	}

	var parts []string
	for _, h := range m.hints {
		// Charm-style: key followed by label, simple spacing
		part := fmt.Sprintf("[%s::b]%s[-:-:-] [%s]%s[-]", TagKey(), h.Key, TagFgDim(), h.Description)
		parts = append(parts, part)
	}

	// Simple space separation
	m.SetText(" " + strings.Join(parts, "   "))
}
