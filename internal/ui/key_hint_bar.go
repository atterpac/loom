package ui

import (
	"fmt"
	"strings"

	"github.com/atterpac/loom/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// KeyHintBar displays pill-style key hints at the bottom of modals.
// Renders hints like: [Esc] Cancel   [Enter] Confirm   [j/k] Navigate
type KeyHintBar struct {
	*tview.Box
	hints []KeyHint
}

// NewKeyHintBar creates a new key hint bar.
func NewKeyHintBar() *KeyHintBar {
	khb := &KeyHintBar{
		Box:   tview.NewBox(),
		hints: []KeyHint{},
	}
	khb.SetBackgroundColor(ColorBg())

	OnThemeChange(func(_ *config.ParsedTheme) {
		khb.SetBackgroundColor(ColorBg())
	})

	return khb
}

// SetHints sets the key hints to display.
func (khb *KeyHintBar) SetHints(hints []KeyHint) *KeyHintBar {
	khb.hints = hints
	return khb
}

// Draw renders the pill-style key hints.
func (khb *KeyHintBar) Draw(screen tcell.Screen) {
	khb.Box.DrawForSubclass(screen, khb)

	x, y, width, _ := khb.GetInnerRect()
	if width <= 0 || len(khb.hints) == 0 {
		return
	}

	// Build hint strings and calculate total width
	type renderedHint struct {
		key         string
		description string
		totalWidth  int
	}

	var rendered []renderedHint
	totalWidth := 0
	spacing := 3 // spaces between hints

	for _, h := range khb.hints {
		rh := renderedHint{
			key:         h.Key,
			description: h.Description,
			totalWidth:  len(h.Key) + 3 + len(h.Description), // " Key " + " " + description
		}
		rendered = append(rendered, rh)
		totalWidth += rh.totalWidth
	}
	totalWidth += spacing * (len(rendered) - 1) // Add spacing between hints

	// Center the hints
	startX := x + (width-totalWidth)/2
	if startX < x {
		startX = x
	}

	currentX := startX
	bgStyle := tcell.StyleDefault.Background(ColorBg())
	keyBgStyle := tcell.StyleDefault.
		Foreground(ColorBg()).
		Background(ColorAccent()).
		Bold(true)
	descStyle := tcell.StyleDefault.
		Foreground(ColorFgDim()).
		Background(ColorBg())

	for i, rh := range rendered {
		if currentX >= x+width {
			break
		}

		// Draw key pill: " Key "
		keyText := fmt.Sprintf(" %s ", rh.key)
		for _, r := range keyText {
			if currentX < x+width {
				screen.SetContent(currentX, y, r, nil, keyBgStyle)
				currentX++
			}
		}

		// Draw space after key
		if currentX < x+width {
			screen.SetContent(currentX, y, ' ', nil, bgStyle)
			currentX++
		}

		// Draw description
		for _, r := range rh.description {
			if currentX < x+width {
				screen.SetContent(currentX, y, r, nil, descStyle)
				currentX++
			}
		}

		// Draw spacing between hints (except after last)
		if i < len(rendered)-1 {
			for j := 0; j < spacing && currentX < x+width; j++ {
				screen.SetContent(currentX, y, ' ', nil, bgStyle)
				currentX++
			}
		}
	}
}

// GetHeight returns the height needed for the hint bar.
func (khb *KeyHintBar) GetHeight() int {
	if len(khb.hints) == 0 {
		return 0
	}
	return 1
}

// FormatHintsText returns hints as a formatted text string for TextView.
// Useful for modals that embed hints differently.
func FormatHintsText(hints []KeyHint) string {
	if len(hints) == 0 {
		return ""
	}

	var parts []string
	for _, h := range hints {
		// Pill style using tview color tags
		part := fmt.Sprintf("[%s:%s:b] %s [-:-:-] [%s]%s[-]",
			TagBg(), TagAccent(), h.Key, TagFgDim(), h.Description)
		parts = append(parts, part)
	}

	return strings.Join(parts, "   ")
}
