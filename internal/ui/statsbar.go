package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StatsBar displays application status and workflow statistics in a bordered panel.
type StatsBar struct {
	*tview.Box
	profile    string
	namespace  string
	connected  bool
	running    int
	completed  int
	failed     int
	taskQueues int
}

// NewStatsBar creates a new stats bar component.
func NewStatsBar() *StatsBar {
	s := &StatsBar{
		Box:       tview.NewBox(),
		namespace: "default",
		connected: true,
	}
	s.SetBackgroundColor(tcell.ColorDefault)
	return s
}

// SetProfile updates the displayed profile name.
func (s *StatsBar) SetProfile(profile string) {
	s.profile = profile
}

// SetNamespace updates the displayed namespace.
func (s *StatsBar) SetNamespace(ns string) {
	s.namespace = ns
}

// SetConnected updates the connection status.
func (s *StatsBar) SetConnected(connected bool) {
	s.connected = connected
}

// SetWorkflowStats updates the workflow statistics.
func (s *StatsBar) SetWorkflowStats(running, completed, failed int) {
	s.running = running
	s.completed = completed
	s.failed = failed
}

// SetTaskQueueCount updates the task queue count.
func (s *StatsBar) SetTaskQueueCount(count int) {
	s.taskQueues = count
}

// Draw renders the stats bar with rounded borders.
// Colors are read dynamically from the current theme.
func (s *StatsBar) Draw(screen tcell.Screen) {
	// Read colors dynamically at draw time
	bgColor := ColorBg()
	borderColor := ColorPanelBorder()
	titleColor := ColorPanelTitle()
	fgColor := ColorFg()
	fgDimColor := ColorFgDim()

	s.Box.SetBackgroundColor(bgColor)
	s.Box.DrawForSubclass(screen, s)

	x, y, width, height := s.GetInnerRect()
	if width <= 0 || height < 3 {
		return
	}

	borderStyle := tcell.StyleDefault.Foreground(borderColor).Background(bgColor)
	titleStyle := tcell.StyleDefault.Foreground(titleColor).Background(bgColor).Bold(true)
	textStyle := tcell.StyleDefault.Foreground(fgColor).Background(bgColor)
	dimStyle := tcell.StyleDefault.Foreground(fgDimColor).Background(bgColor)

	// Draw rounded border
	screen.SetContent(x, y, '╭', nil, borderStyle)
	screen.SetContent(x+width-1, y, '╮', nil, borderStyle)
	screen.SetContent(x, y+height-1, '╰', nil, borderStyle)
	screen.SetContent(x+width-1, y+height-1, '╯', nil, borderStyle)

	for i := x + 1; i < x+width-1; i++ {
		screen.SetContent(i, y, '─', nil, borderStyle)
		screen.SetContent(i, y+height-1, '─', nil, borderStyle)
	}

	for i := y + 1; i < y+height-1; i++ {
		screen.SetContent(x, i, '│', nil, borderStyle)
		screen.SetContent(x+width-1, i, '│', nil, borderStyle)
	}

	// Draw title in top border
	title := " " + LogoSmall + " "
	titleRunes := []rune(title)
	titleX := x + 2
	for i, r := range titleRunes {
		if titleX+i >= x+width-1 {
			break
		}
		screen.SetContent(titleX+i, y, r, nil, titleStyle)
	}

	// Build content line
	contentY := y + 1
	contentX := x + 2

	// Connection status
	connIcon := IconConnected
	connText := "connected"
	connStyle := tcell.StyleDefault.Foreground(ColorCompleted()).Background(bgColor)
	if !s.connected {
		connIcon = IconDisconnected
		connText = "disconnected"
		connStyle = tcell.StyleDefault.Foreground(ColorFailed()).Background(bgColor)
	}

	accentStyle := tcell.StyleDefault.Foreground(ColorAccent()).Background(bgColor)

	// Draw profile name (if set)
	if s.profile != "" {
		profileText := s.profile
		for i, r := range []rune(profileText) {
			if contentX+i >= x+width-2 {
				break
			}
			screen.SetContent(contentX+i, contentY, r, nil, accentStyle)
		}
		contentX += len(profileText)

		// Separator after profile
		sep := " • "
		for i, r := range []rune(sep) {
			if contentX+i >= x+width-2 {
				break
			}
			screen.SetContent(contentX+i, contentY, r, nil, dimStyle)
		}
		contentX += len(sep)
	}

	// Draw namespace
	nsText := s.namespace
	for i, r := range []rune(nsText) {
		if contentX+i >= x+width-2 {
			break
		}
		screen.SetContent(contentX+i, contentY, r, nil, textStyle)
	}
	contentX += len(nsText)

	// Separator
	nsSep := " • "
	for i, r := range []rune(nsSep) {
		if contentX+i >= x+width-2 {
			break
		}
		screen.SetContent(contentX+i, contentY, r, nil, dimStyle)
	}
	contentX += len(nsSep)

	// Connection status with icon
	connFull := connIcon + " " + connText
	connRunes := []rune(connFull)
	for i, r := range connRunes {
		if contentX+i >= x+width-2 {
			break
		}
		screen.SetContent(contentX+i, contentY, r, nil, connStyle)
	}
	contentX += len(connRunes)

	// Stats section (right-aligned area)
	statsText := s.buildStatsText()
	statsX := x + width - len(statsText) - 3
	if statsX > contentX+3 {
		s.drawStats(screen, statsX, contentY, bgColor)
	}
}

func (s *StatsBar) buildStatsText() string {
	return fmt.Sprintf("Running: %d  Completed: %d  Failed: %d  Queues: %d",
		s.running, s.completed, s.failed, s.taskQueues)
}

func (s *StatsBar) drawStats(screen tcell.Screen, x, y int, bgColor tcell.Color) {
	labelStyle := tcell.StyleDefault.Foreground(ColorFgDim()).Background(bgColor)
	runningStyle := tcell.StyleDefault.Foreground(ColorRunning()).Background(bgColor)
	completedStyle := tcell.StyleDefault.Foreground(ColorCompleted()).Background(bgColor)
	failedStyle := tcell.StyleDefault.Foreground(ColorFailed()).Background(bgColor)
	accentStyle := tcell.StyleDefault.Foreground(ColorAccentDim()).Background(bgColor)

	// Running
	x = s.drawText(screen, x, y, "Running: ", labelStyle)
	x = s.drawText(screen, x, y, fmt.Sprintf("%d", s.running), runningStyle)
	x = s.drawText(screen, x, y, "  ", labelStyle)

	// Completed
	x = s.drawText(screen, x, y, "Completed: ", labelStyle)
	x = s.drawText(screen, x, y, fmt.Sprintf("%d", s.completed), completedStyle)
	x = s.drawText(screen, x, y, "  ", labelStyle)

	// Failed
	x = s.drawText(screen, x, y, "Failed: ", labelStyle)
	x = s.drawText(screen, x, y, fmt.Sprintf("%d", s.failed), failedStyle)
	x = s.drawText(screen, x, y, "  ", labelStyle)

	// Queues
	x = s.drawText(screen, x, y, "Queues: ", labelStyle)
	s.drawText(screen, x, y, fmt.Sprintf("%d", s.taskQueues), accentStyle)
}

func (s *StatsBar) drawText(screen tcell.Screen, x, y int, text string, style tcell.Style) int {
	for _, r := range []rune(text) {
		screen.SetContent(x, y, r, nil, style)
		x++
	}
	return x
}
