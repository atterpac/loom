package view

import (
	"time"

	"github.com/atterpac/jig/theme"
	"github.com/atterpac/tempo/internal/temporal"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	timelineLabelWidth = 25 // Width for lane labels on the left
	timelineMinWidth   = 40 // Minimum timeline bar area width
)

// TimelineLane represents a horizontal lane in the timeline.
type TimelineLane struct {
	Name      string
	Type      temporal.EventGroupType
	Status    string
	StartTime time.Time
	EndTime   *time.Time
	Node      *temporal.EventTreeNode
}

// TimelineView displays workflow events as a horizontal Gantt-style timeline.
type TimelineView struct {
	*tview.Box
	lanes        []TimelineLane
	startTime    time.Time
	endTime      time.Time
	scrollX      int
	scrollY      int
	zoomLevel    float64
	selectedLane int
	onSelect     func(lane *TimelineLane)
}

// NewTimelineView creates a new timeline/Gantt chart view.
func NewTimelineView() *TimelineView {
	tv := &TimelineView{
		Box:          tview.NewBox(),
		lanes:        []TimelineLane{},
		zoomLevel:    1.0,
		selectedLane: 0,
	}

	tv.SetBackgroundColor(tcell.ColorDefault)
	tv.SetBorder(false)

	return tv
}

// Destroy is a no-op kept for backward compatibility.
func (tv *TimelineView) Destroy() {}

// SetNodes populates the timeline from event tree nodes.
func (tv *TimelineView) SetNodes(nodes []*temporal.EventTreeNode) {
	tv.lanes = nil
	tv.selectedLane = 0

	if len(nodes) == 0 {
		return
	}

	// Find time range
	tv.startTime = time.Now()
	tv.endTime = time.Time{}

	for _, node := range nodes {
		// Skip workflow-level events, only show activities/timers/child workflows
		if node.Type == temporal.GroupWorkflow || node.Type == temporal.GroupWorkflowTask {
			continue
		}

		if node.StartTime.Before(tv.startTime) {
			tv.startTime = node.StartTime
		}
		if node.EndTime != nil && node.EndTime.After(tv.endTime) {
			tv.endTime = *node.EndTime
		} else if node.EndTime == nil && time.Now().After(tv.endTime) {
			tv.endTime = time.Now()
		}

		lane := TimelineLane{
			Name:      node.Name,
			Type:      node.Type,
			Status:    node.Status,
			StartTime: node.StartTime,
			EndTime:   node.EndTime,
			Node:      node,
		}
		tv.lanes = append(tv.lanes, lane)
	}

	// Ensure we have at least some time range
	if tv.endTime.IsZero() || tv.endTime.Before(tv.startTime) {
		tv.endTime = tv.startTime.Add(time.Minute)
	}
}

// Draw renders the timeline view.
// Colors are read dynamically at draw time.
func (tv *TimelineView) Draw(screen tcell.Screen) {
	// Read colors dynamically
	bgColor := theme.Bg()
	tv.SetBackgroundColor(bgColor)

	tv.Box.DrawForSubclass(screen, tv)

	x, y, width, height := tv.GetInnerRect()
	if width < timelineLabelWidth+10 || height < 3 {
		return
	}

	// Draw header with time scale
	tv.drawHeader(screen, x, y, width)

	// Draw lanes starting from y+2 (after header)
	barAreaWidth := width - timelineLabelWidth - 1
	if barAreaWidth < timelineMinWidth {
		barAreaWidth = timelineMinWidth
	}

	timeRange := tv.endTime.Sub(tv.startTime)
	if timeRange <= 0 {
		timeRange = time.Minute
	}

	visibleLanes := height - 3 // Subtract header rows
	startLane := tv.scrollY
	endLane := startLane + visibleLanes
	if endLane > len(tv.lanes) {
		endLane = len(tv.lanes)
	}

	for i := startLane; i < endLane; i++ {
		lane := tv.lanes[i]
		laneY := y + 2 + (i - startLane)

		// Draw lane label
		tv.drawLaneLabel(screen, x, laneY, lane, i == tv.selectedLane)

		// Draw lane bar
		tv.drawLaneBar(screen, x+timelineLabelWidth+1, laneY, barAreaWidth, lane, timeRange, i == tv.selectedLane)
	}

	// Draw legend at bottom if space
	if height > len(tv.lanes)+4 {
		tv.drawLegend(screen, x, y+height-1, width)
	}
}

// drawHeader draws the time scale header.
func (tv *TimelineView) drawHeader(screen tcell.Screen, x, y, width int) {
	barAreaWidth := width - timelineLabelWidth - 1
	timeRange := tv.endTime.Sub(tv.startTime)

	// Draw label column header
	labelStyle := tcell.StyleDefault.Foreground(theme.PanelTitle()).Background(theme.Bg())
	tview.Print(screen, "Event", x, y, timelineLabelWidth, tview.AlignLeft, theme.PanelTitle())

	// Draw time markers
	markerCount := 5
	if barAreaWidth < 60 {
		markerCount = 3
	}

	for i := 0; i <= markerCount; i++ {
		pos := x + timelineLabelWidth + 1 + (barAreaWidth * i / markerCount)
		if pos >= x+width {
			break
		}

		// Calculate time at this position
		offset := time.Duration(float64(timeRange) * float64(i) / float64(markerCount))
		t := tv.startTime.Add(offset)

		// Format time marker
		var marker string
		if timeRange < time.Minute {
			marker = t.Format("04:05.0")
		} else if timeRange < time.Hour {
			marker = t.Format("04:05")
		} else {
			marker = t.Format("15:04")
		}

		// Draw marker
		tview.Print(screen, marker, pos, y, 8, tview.AlignLeft, theme.FgDim())

		// Draw tick mark
		screen.SetContent(pos, y+1, '│', nil, labelStyle)
	}

	// Draw horizontal line under header
	lineStyle := tcell.StyleDefault.Foreground(theme.Border()).Background(theme.Bg())
	for i := x + timelineLabelWidth + 1; i < x+width; i++ {
		screen.SetContent(i, y+1, '─', nil, lineStyle)
	}
}

// drawLaneLabel draws the label for a lane.
func (tv *TimelineView) drawLaneLabel(screen tcell.Screen, x, y int, lane TimelineLane, selected bool) {
	// Truncate name if needed
	name := lane.Name
	maxLen := timelineLabelWidth - 2
	if len(name) > maxLen {
		name = name[:maxLen-1] + "…"
	}

	// Choose style based on selection
	var style tcell.Style
	if selected {
		style = tcell.StyleDefault.Foreground(theme.Bg()).Background(theme.Highlight())
	} else {
		style = tcell.StyleDefault.Foreground(tv.statusColor(lane.Status)).Background(theme.Bg())
	}

	// Clear label area
	for i := 0; i < timelineLabelWidth; i++ {
		screen.SetContent(x+i, y, ' ', nil, style)
	}

	// Draw name
	for i, r := range name {
		if x+i >= x+timelineLabelWidth {
			break
		}
		screen.SetContent(x+i, y, r, nil, style)
	}

	// Draw separator
	sepStyle := tcell.StyleDefault.Foreground(theme.Border()).Background(theme.Bg())
	screen.SetContent(x+timelineLabelWidth, y, '│', nil, sepStyle)
}

// drawLaneBar draws the timeline bar for a lane.
func (tv *TimelineView) drawLaneBar(screen tcell.Screen, x, y, width int, lane TimelineLane, timeRange time.Duration, selected bool) {
	// Calculate bar position and width
	startOffset := lane.StartTime.Sub(tv.startTime)
	barStart := int(float64(width) * float64(startOffset) / float64(timeRange))

	var barEnd int
	if lane.EndTime != nil {
		endOffset := lane.EndTime.Sub(tv.startTime)
		barEnd = int(float64(width) * float64(endOffset) / float64(timeRange))
	} else {
		// Running - extend to current time or end of view
		barEnd = width
	}

	// Ensure minimum bar width
	if barEnd <= barStart {
		barEnd = barStart + 1
	}

	// Apply zoom and scroll
	barStart = int(float64(barStart)*tv.zoomLevel) - tv.scrollX
	barEnd = int(float64(barEnd)*tv.zoomLevel) - tv.scrollX

	// Clamp to visible area
	if barStart < 0 {
		barStart = 0
	}
	if barEnd > width {
		barEnd = width
	}

	// Choose bar character and color based on status
	barChar, barColor := tv.barStyle(lane.Status)
	barStyle := tcell.StyleDefault.Foreground(barColor).Background(theme.Bg())

	if selected {
		barStyle = barStyle.Bold(true)
	}

	// Draw empty space before bar
	emptyStyle := tcell.StyleDefault.Foreground(theme.BgLight()).Background(theme.Bg())
	for i := 0; i < barStart && i < width; i++ {
		screen.SetContent(x+i, y, '·', nil, emptyStyle)
	}

	// Draw the bar
	for i := barStart; i < barEnd && i < width; i++ {
		screen.SetContent(x+i, y, barChar, nil, barStyle)
	}

	// Draw empty space after bar
	for i := barEnd; i < width; i++ {
		screen.SetContent(x+i, y, '·', nil, emptyStyle)
	}
}

// drawLegend draws the status legend at the bottom.
func (tv *TimelineView) drawLegend(screen tcell.Screen, x, y, width int) {
	legend := []struct {
		char   rune
		status string
		color  tcell.Color
	}{
		{'█', "Completed", theme.Success()},
		{'▓', "Running", theme.Warning()},
		{'░', "Failed", theme.Error()},
		{'▒', "Pending", theme.FgDim()},
	}

	pos := x
	for _, item := range legend {
		if pos+15 > x+width {
			break
		}

		style := tcell.StyleDefault.Foreground(item.color).Background(theme.Bg())
		screen.SetContent(pos, y, item.char, nil, style)
		pos++

		labelStyle := tcell.StyleDefault.Foreground(theme.FgDim()).Background(theme.Bg())
		for _, r := range item.status {
			screen.SetContent(pos, y, r, nil, labelStyle)
			pos++
		}
		pos += 2 // spacing
	}
}

// barStyle returns the bar character and color for a status.
func (tv *TimelineView) barStyle(status string) (rune, tcell.Color) {
	switch status {
	case "Running":
		return '▓', theme.Warning()
	case "Completed", "Fired":
		return '█', theme.Success()
	case "Failed", "TimedOut":
		return '░', theme.Error()
	case "Canceled", "Terminated":
		return '▒', theme.Warning()
	case "Scheduled", "Initiated", "Pending":
		return '▒', theme.FgDim()
	default:
		return '▒', theme.Fg()
	}
}

// statusColor returns the color for a status.
func (tv *TimelineView) statusColor(status string) tcell.Color {
	return theme.StatusColor(status)
}

// InputHandler handles keyboard input.
func (tv *TimelineView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return tv.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			tv.moveSelection(-1)
		case tcell.KeyDown:
			tv.moveSelection(1)
		case tcell.KeyLeft:
			tv.scroll(-5)
		case tcell.KeyRight:
			tv.scroll(5)
		case tcell.KeyEnter:
			if tv.onSelect != nil && tv.selectedLane >= 0 && tv.selectedLane < len(tv.lanes) {
				tv.onSelect(&tv.lanes[tv.selectedLane])
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'k':
				tv.moveSelection(-1)
			case 'j':
				tv.moveSelection(1)
			case 'h':
				tv.scroll(-5)
			case 'l':
				tv.scroll(5)
			case '+', '=':
				tv.zoom(1.2)
			case '-':
				tv.zoom(0.8)
			case '0':
				tv.resetView()
			}
		}
	})
}

// moveSelection moves the lane selection up or down.
func (tv *TimelineView) moveSelection(delta int) {
	if len(tv.lanes) == 0 {
		return
	}

	tv.selectedLane += delta
	if tv.selectedLane < 0 {
		tv.selectedLane = 0
	}
	if tv.selectedLane >= len(tv.lanes) {
		tv.selectedLane = len(tv.lanes) - 1
	}

	// Adjust scroll to keep selection visible
	_, _, _, height := tv.GetInnerRect()
	visibleLanes := height - 3

	if tv.selectedLane < tv.scrollY {
		tv.scrollY = tv.selectedLane
	}
	if tv.selectedLane >= tv.scrollY+visibleLanes {
		tv.scrollY = tv.selectedLane - visibleLanes + 1
	}
}

// scroll horizontally scrolls the timeline.
func (tv *TimelineView) scroll(delta int) {
	tv.scrollX += delta
	if tv.scrollX < 0 {
		tv.scrollX = 0
	}
}

// zoom adjusts the zoom level.
func (tv *TimelineView) zoom(factor float64) {
	tv.zoomLevel *= factor
	if tv.zoomLevel < 0.5 {
		tv.zoomLevel = 0.5
	}
	if tv.zoomLevel > 5.0 {
		tv.zoomLevel = 5.0
	}
}

// resetView resets zoom and scroll.
func (tv *TimelineView) resetView() {
	tv.zoomLevel = 1.0
	tv.scrollX = 0
	tv.scrollY = 0
}

// SetOnSelect sets the callback for lane selection.
func (tv *TimelineView) SetOnSelect(fn func(lane *TimelineLane)) {
	tv.onSelect = fn
}

// SelectedLane returns the currently selected lane.
func (tv *TimelineView) SelectedLane() *TimelineLane {
	if tv.selectedLane >= 0 && tv.selectedLane < len(tv.lanes) {
		return &tv.lanes[tv.selectedLane]
	}
	return nil
}

// LaneCount returns the number of lanes.
func (tv *TimelineView) LaneCount() int {
	return len(tv.lanes)
}

// Focus implements tview.Primitive.
func (tv *TimelineView) Focus(delegate func(p tview.Primitive)) {
	tv.Box.Focus(delegate)
}

// HasFocus implements tview.Primitive.
func (tv *TimelineView) HasFocus() bool {
	return tv.Box.HasFocus()
}
