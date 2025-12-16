package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SponsorURL is the GitHub sponsors URL
const SponsorURL = "github.com/sponsors/atterpac"

// Menu displays keybinding hints at the bottom of the screen with sponsor link.
type Menu struct {
	*tview.Flex
	hintsView   *tview.TextView
	sponsorView *tview.TextView
	hints       []KeyHint
}

// NewMenu creates a new menu component.
func NewMenu() *Menu {
	hintsView := tview.NewTextView()
	hintsView.SetDynamicColors(true)
	hintsView.SetBackgroundColor(tcell.ColorDefault)

	sponsorView := tview.NewTextView()
	sponsorView.SetDynamicColors(true)
	sponsorView.SetTextAlign(tview.AlignRight)
	sponsorView.SetBackgroundColor(tcell.ColorDefault)

	m := &Menu{
		Flex:        tview.NewFlex().SetDirection(tview.FlexColumn),
		hintsView:   hintsView,
		sponsorView: sponsorView,
		hints:       []KeyHint{},
	}

	m.SetBackgroundColor(tcell.ColorDefault)
	m.AddItem(hintsView, 0, 1, false)
	m.AddItem(sponsorView, 45, 0, false)

	m.render()

	return m
}

// Draw applies theme colors dynamically before drawing.
func (m *Menu) Draw(screen tcell.Screen) {
	menuColor := ColorMenu()
	m.SetBackgroundColor(menuColor)
	m.hintsView.SetBackgroundColor(menuColor)
	m.hintsView.SetTextColor(ColorFg())
	m.sponsorView.SetBackgroundColor(menuColor)
	m.sponsorView.SetTextColor(ColorFgDim())

	m.Flex.Draw(screen)
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
	// Render hints
	if len(m.hints) == 0 {
		m.hintsView.SetText("")
	} else {
		var parts []string
		for _, h := range m.hints {
			part := fmt.Sprintf("[%s::b]%s[-:-:-] [%s]%s[-]", TagKey(), h.Key, TagFgDim(), h.Description)
			parts = append(parts, part)
		}
		m.hintsView.SetText(" " + strings.Join(parts, "   "))
	}

	// Render sponsor (subtle, no underline)
	m.sponsorView.SetText(fmt.Sprintf(
		"[%s]%s %s[-] ",
		TagFgDim(), IconHeart, SponsorURL,
	))
}
