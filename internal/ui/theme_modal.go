package ui

import (
	"github.com/atterpac/loom/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ThemeSelectorModal displays a modal for selecting themes with live preview.
type ThemeSelectorModal struct {
	*Modal
	table         *tview.Table
	nav           *TableListNavigator
	themes        []string
	originalTheme string
	onSelect      func(name string)
	onCancel      func()
}

// NewThemeSelectorModal creates a new theme selector modal.
func NewThemeSelectorModal() *ThemeSelectorModal {
	tsm := &ThemeSelectorModal{
		Modal: NewModal(ModalConfig{
			Title:     "Select Theme",
			Width:     40,
			Height:    8,
			MinHeight: 8,
			MaxHeight: 20,
			Backdrop:  true,
		}),
		table:  tview.NewTable(),
		themes: config.ThemeNames(),
	}

	if t := ActiveTheme(); t != nil {
		tsm.originalTheme = t.Key
	}

	tsm.setup()
	return tsm
}

// SetOnSelect sets the callback when a theme is selected.
func (tsm *ThemeSelectorModal) SetOnSelect(fn func(name string)) *ThemeSelectorModal {
	tsm.onSelect = fn
	return tsm
}

// SetOnCancel sets the callback when selection is cancelled.
func (tsm *ThemeSelectorModal) SetOnCancel(fn func()) *ThemeSelectorModal {
	tsm.onCancel = fn
	return tsm
}

func (tsm *ThemeSelectorModal) setup() {
	// Table will pick up colors dynamically via tview.Styles
	tsm.table.SetBackgroundColor(tcell.ColorDefault)
	tsm.table.SetSelectable(true, false)
	tsm.table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(ColorFg()).
		Background(ColorHighlight()).
		Bold(true))

	// No header rows in theme list
	tsm.nav = NewTableListNavigator(tsm.table, 0)

	// Adjust modal height based on theme count
	height := len(tsm.themes) + 2
	if height < 8 {
		height = 8
	}
	if height > 18 {
		height = 18
	}
	tsm.SetSize(40, height)

	tsm.SetContent(tsm.table)
	tsm.SetHints([]KeyHint{
		{Key: "j/k", Description: "Navigate"},
		{Key: "Enter", Description: "Select"},
		{Key: "Esc", Description: "Cancel"},
	})

	tsm.rebuildTable()

	// Select current theme row BEFORE setting up the selection changed callback
	// This prevents triggering SetTheme during initialization
	for i, name := range tsm.themes {
		if name == tsm.originalTheme {
			tsm.table.Select(i, 0)
			break
		}
	}

	// Preview theme on selection change (set up AFTER initial selection)
	tsm.table.SetSelectionChangedFunc(func(row, col int) {
		if row >= 0 && row < len(tsm.themes) {
			themeName := tsm.themes[row]
			_ = SetTheme(themeName)
		}
	})

	// Handle final selection
	tsm.table.SetSelectedFunc(func(row, col int) {
		if tsm.onSelect != nil {
			selected := tsm.GetSelectedTheme()
			if selected != "" {
				tsm.onSelect(selected)
			}
		}
	})

	// Set input capture directly on table for navigation
	tsm.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			tsm.restoreAndCancel()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				tsm.nav.MoveDown()
				return nil
			case 'k':
				tsm.nav.MoveUp()
				return nil
			case 'q':
				tsm.restoreAndCancel()
				return nil
			}
		}
		return event
	})
}

func (tsm *ThemeSelectorModal) restoreAndCancel() {
	// Restore original theme on cancel
	if tsm.originalTheme != "" {
		_ = SetTheme(tsm.originalTheme)
	}
	if tsm.onCancel != nil {
		tsm.onCancel()
	}
}

func (tsm *ThemeSelectorModal) rebuildTable() {
	if len(tsm.themes) == 0 {
		return
	}

	// Preserve current selection
	currentRow, _ := tsm.table.GetSelection()

	tsm.table.Clear()

	// Add themes to table
	for i, name := range tsm.themes {
		marker := "  "
		if name == tsm.originalTheme {
			marker = IconCompleted + " "
		}
		cell := tview.NewTableCell(marker + name).
			SetTextColor(tcell.ColorDefault).
			SetBackgroundColor(tcell.ColorDefault)
		tsm.table.SetCell(i, 0, cell)
	}

	// Restore selection (or default to original theme row if no prior selection)
	if currentRow >= 0 && currentRow < len(tsm.themes) {
		tsm.table.Select(currentRow, 0)
	} else {
		for i, name := range tsm.themes {
			if name == tsm.originalTheme {
				tsm.table.Select(i, 0)
				break
			}
		}
	}
}

// GetSelectedTheme returns the currently highlighted theme name.
func (tsm *ThemeSelectorModal) GetSelectedTheme() string {
	idx := tsm.nav.GetSelectedIndex()
	if idx >= 0 && idx < len(tsm.themes) {
		return tsm.themes[idx]
	}
	return ""
}

// GetOriginalTheme returns the theme that was active when the modal opened.
func (tsm *ThemeSelectorModal) GetOriginalTheme() string {
	return tsm.originalTheme
}

// Draw updates colors dynamically and draws the modal.
func (tsm *ThemeSelectorModal) Draw(screen tcell.Screen) {
	// Update table colors dynamically for theme preview
	bg := ColorBg()
	fg := ColorFg()
	tsm.table.SetBackgroundColor(bg)
	tsm.table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(fg).
		Background(ColorHighlight()).
		Bold(true))

	// Update all cell colors to match current theme
	for row := 0; row < tsm.table.GetRowCount(); row++ {
		if cell := tsm.table.GetCell(row, 0); cell != nil {
			cell.SetTextColor(fg)
			cell.SetBackgroundColor(bg)
		}
	}

	// Update modal colors
	tsm.Modal.SetBackgroundColor(ColorBgDark())
	tsm.Modal.GetPanel().SetBackgroundColor(bg)

	tsm.Modal.Draw(screen)
}

// Focus delegates focus to the table.
func (tsm *ThemeSelectorModal) Focus(delegate func(p tview.Primitive)) {
	delegate(tsm.table)
}
