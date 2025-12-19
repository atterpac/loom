package config

import (
	"github.com/atterpac/jig/theme"
	"github.com/gdamore/tcell/v2"
)

// JigThemeAdapter adapts tempo's ParsedTheme to jig's Theme interface.
type JigThemeAdapter struct {
	parsed *ParsedTheme
}

// NewJigThemeAdapter creates a new adapter for a parsed theme.
func NewJigThemeAdapter(parsed *ParsedTheme) *JigThemeAdapter {
	return &JigThemeAdapter{parsed: parsed}
}

// Base colors
func (a *JigThemeAdapter) Bg() tcell.Color      { return a.parsed.Colors.Bg }
func (a *JigThemeAdapter) BgLight() tcell.Color { return a.parsed.Colors.BgLight }
func (a *JigThemeAdapter) BgDark() tcell.Color  { return a.parsed.Colors.BgDark }
func (a *JigThemeAdapter) Fg() tcell.Color      { return a.parsed.Colors.Fg }
func (a *JigThemeAdapter) FgDim() tcell.Color   { return a.parsed.Colors.FgDim }
func (a *JigThemeAdapter) FgMuted() tcell.Color { return a.parsed.Colors.FgDim } // Map to FgDim

// Accent colors
func (a *JigThemeAdapter) Accent() tcell.Color    { return a.parsed.Colors.Accent }
func (a *JigThemeAdapter) AccentDim() tcell.Color { return a.parsed.Colors.AccentDim }
func (a *JigThemeAdapter) Highlight() tcell.Color { return a.parsed.Colors.Highlight }

// Semantic colors
func (a *JigThemeAdapter) Success() tcell.Color { return a.parsed.Colors.Completed }
func (a *JigThemeAdapter) Warning() tcell.Color { return a.parsed.Colors.Canceled }
func (a *JigThemeAdapter) Error() tcell.Color   { return a.parsed.Colors.Failed }
func (a *JigThemeAdapter) Info() tcell.Color    { return a.parsed.Colors.Running }

// Border colors
func (a *JigThemeAdapter) Border() tcell.Color      { return a.parsed.Colors.Border }
func (a *JigThemeAdapter) BorderFocus() tcell.Color { return a.parsed.Colors.Accent }

// UI element colors
func (a *JigThemeAdapter) Header() tcell.Color      { return a.parsed.Colors.Header }
func (a *JigThemeAdapter) Menu() tcell.Color        { return a.parsed.Colors.Menu }
func (a *JigThemeAdapter) TableHeader() tcell.Color { return a.parsed.Colors.TableHeader }
func (a *JigThemeAdapter) Key() tcell.Color         { return a.parsed.Colors.Key }
func (a *JigThemeAdapter) Crumb() tcell.Color       { return a.parsed.Colors.Crumb }
func (a *JigThemeAdapter) PanelBorder() tcell.Color { return a.parsed.Colors.PanelBorder }
func (a *JigThemeAdapter) PanelTitle() tcell.Color  { return a.parsed.Colors.PanelTitle }

// Verify interface compliance at compile time
var _ theme.Theme = (*JigThemeAdapter)(nil)
