package ui

import (
	"strings"

	"github.com/atterpac/temportui/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Table is a generic table component with selection support.
type Table struct {
	*tview.Table
	headers     []string
	actions     *ActionRegistry
	onSelect    func(row int)
	selectColor tcell.Color
}

// NewTable creates a new table component.
func NewTable() *Table {
	t := &Table{
		Table:       tview.NewTable(),
		actions:     NewActionRegistry(),
		selectColor: ColorHighlight(),
	}
	t.SetSelectable(true, false)
	t.SetBorders(false)
	t.SetFixed(1, 0) // Fixed header row
	t.applyTheme()

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		t.applyTheme()
	})

	return t
}

// applyTheme applies the current theme colors to the table.
func (t *Table) applyTheme() {
	t.selectColor = ColorHighlight()
	t.SetBackgroundColor(ColorBg())
	t.SetBorderColor(ColorBorder())
	t.SetBordersColor(ColorBorder())
	t.SetSelectedStyle(tcell.StyleDefault.
		Foreground(ColorFg()).
		Background(t.selectColor).
		Bold(true))

	// Update all existing cells
	t.RefreshColors()
}

// RefreshColors updates all cell colors to match current theme.
// Call this after theme changes to update the display.
func (t *Table) RefreshColors() {
	rowCount := t.GetRowCount()
	colCount := t.GetColumnCount()
	for row := 0; row < rowCount; row++ {
		for col := 0; col < colCount; col++ {
			cell := t.GetCell(row, col)
			if cell != nil {
				// Update background color for all cells
				cell.SetBackgroundColor(ColorBg())
				// Header row gets dim text, data rows get default fg
				if row == 0 {
					cell.SetTextColor(ColorFgDim())
				}
			}
		}
	}
}

// SetHeaders sets the table column headers.
func (t *Table) SetHeaders(headers ...string) {
	t.headers = headers
	for i, h := range headers {
		// Charm-style: lowercase headers, subtle styling
		cell := tview.NewTableCell(" " + strings.ToLower(h)).
			SetTextColor(ColorFgDim()).
			SetBackgroundColor(ColorBg()).
			SetSelectable(false).
			SetExpansion(1)
		t.SetCell(0, i, cell)
	}
}

// AddRow adds a row to the table.
func (t *Table) AddRow(values ...string) int {
	row := t.GetRowCount()
	for i, v := range values {
		cell := tview.NewTableCell(" " + v).
			SetTextColor(ColorFg()).
			SetBackgroundColor(ColorBg()).
			SetExpansion(1)
		t.SetCell(row, i, cell)
	}
	return row
}

// AddColoredRow adds a row with a specific color.
func (t *Table) AddColoredRow(color tcell.Color, values ...string) int {
	row := t.GetRowCount()
	for i, v := range values {
		cell := tview.NewTableCell(" " + v).
			SetTextColor(color).
			SetBackgroundColor(ColorBg()).
			SetExpansion(1)
		t.SetCell(row, i, cell)
	}
	return row
}

// AddStyledRow adds a row with status icon and color.
func (t *Table) AddStyledRow(status string, values ...string) int {
	row := t.GetRowCount()
	color := StatusColorTcell(status)
	icon := StatusIcon(status)

	for i, v := range values {
		displayValue := " " + v
		cellColor := color

		// Add status icon to the status column (usually column 2 or 3)
		if v == status {
			displayValue = " " + icon + " " + v
		} else {
			cellColor = ColorFg()
		}

		cell := tview.NewTableCell(displayValue).
			SetTextColor(cellColor).
			SetBackgroundColor(ColorBg()).
			SetExpansion(1)
		t.SetCell(row, i, cell)
	}
	return row
}

// ClearRows removes all rows except the header.
func (t *Table) ClearRows() {
	rowCount := t.GetRowCount()
	for i := rowCount - 1; i > 0; i-- {
		t.RemoveRow(i)
	}
}

// SetOnSelect sets the callback for when a row is selected.
func (t *Table) SetOnSelect(fn func(row int)) {
	t.onSelect = fn
	t.SetSelectedFunc(func(row, col int) {
		if row > 0 && fn != nil { // Skip header
			fn(row - 1) // Adjust for header
		}
	})
}

// Actions returns the action registry for this table.
func (t *Table) Actions() *ActionRegistry {
	return t.actions
}

// SelectedRow returns the currently selected row index (0-based, excluding header).
func (t *Table) SelectedRow() int {
	row, _ := t.GetSelection()
	return row - 1 // Adjust for header
}

// SelectRow selects a specific row (0-based, excluding header).
func (t *Table) SelectRow(row int) {
	t.Select(row+1, 0) // Adjust for header
}

// RowCount returns the number of data rows (excluding header).
func (t *Table) RowCount() int {
	count := t.GetRowCount()
	if count > 0 {
		return count - 1 // Exclude header
	}
	return 0
}

// StatusColor returns the appropriate color for a workflow status (deprecated, use StatusColorTcell).
func StatusColor(status string) tcell.Color {
	return StatusColorTcell(status)
}
