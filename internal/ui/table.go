package ui

import (
	"strings"

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
		selectColor: ColorHighlight,
	}
	t.SetSelectable(true, false)
	t.SetBorders(false)
	t.SetFixed(1, 0) // Fixed header row
	t.SetBackgroundColor(ColorBg)
	t.SetBorderColor(ColorBorder)
	t.SetBordersColor(ColorBorder)
	t.SetSelectedStyle(tcell.StyleDefault.
		Foreground(ColorFg).
		Background(t.selectColor).
		Bold(true))
	t.SetSeparator(' ')
	return t
}

// SetHeaders sets the table column headers.
func (t *Table) SetHeaders(headers ...string) {
	t.headers = headers
	for i, h := range headers {
		// Charm-style: lowercase headers, subtle styling
		cell := tview.NewTableCell(" " + strings.ToLower(h)).
			SetTextColor(ColorFgDim).
			SetBackgroundColor(ColorBg).
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
			SetTextColor(ColorFg).
			SetBackgroundColor(ColorBg).
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
			SetBackgroundColor(ColorBg).
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
			cellColor = ColorFg
		}

		cell := tview.NewTableCell(displayValue).
			SetTextColor(cellColor).
			SetBackgroundColor(ColorBg).
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
