package ui

import (
	"fmt"
	"strings"

	"github.com/atterpac/temportui/internal/config"
	"github.com/rivo/tview"
)

// Crumbs displays breadcrumb navigation path.
type Crumbs struct {
	*tview.TextView
	path []string
}

// NewCrumbs creates a new breadcrumb component.
func NewCrumbs() *Crumbs {
	c := &Crumbs{
		TextView: tview.NewTextView(),
		path:     []string{},
	}
	c.SetDynamicColors(true)
	c.applyTheme()
	c.render()

	// Register for theme changes
	OnThemeChange(func(_ *config.ParsedTheme) {
		c.applyTheme()
		c.render()
	})

	return c
}

// applyTheme applies the current theme colors to the crumbs.
func (c *Crumbs) applyTheme() {
	c.SetBackgroundColor(ColorBg())
	c.SetTextColor(ColorCrumb())
}

// SetPath sets the breadcrumb path.
func (c *Crumbs) SetPath(path []string) {
	c.path = path
	c.render()
}

// Push adds a segment to the path.
func (c *Crumbs) Push(segment string) {
	c.path = append(c.path, segment)
	c.render()
}

// Pop removes the last segment from the path.
func (c *Crumbs) Pop() string {
	if len(c.path) == 0 {
		return ""
	}
	last := c.path[len(c.path)-1]
	c.path = c.path[:len(c.path)-1]
	c.render()
	return last
}

// Clear removes all segments.
func (c *Crumbs) Clear() {
	c.path = []string{}
	c.render()
}

func (c *Crumbs) render() {
	if len(c.path) == 0 {
		c.SetText("")
		return
	}

	// Charm-style: simple slash separators
	var parts []string
	for i, segment := range c.path {
		if i == len(c.path)-1 {
			// Last segment is highlighted
			parts = append(parts, fmt.Sprintf("[%s]%s[-]", TagAccent(), segment))
		} else {
			// Previous segments are dimmed
			parts = append(parts, fmt.Sprintf("[%s]%s[-]", TagFgDim(), segment))
		}
	}

	c.SetText(" " + strings.Join(parts, " / "))
}
