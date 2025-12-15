package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// App wraps tview.Application with additional functionality.
type App struct {
	*tview.Application
	main    *tview.Flex
	header  *Header
	crumbs  *Crumbs
	menu    *Menu
	pages   *Pages
	content tview.Primitive
}

// NewApp creates a new application wrapper.
func NewApp() *App {
	app := &App{
		Application: tview.NewApplication(),
		header:      NewHeader(),
		crumbs:      NewCrumbs(),
		menu:        NewMenu(),
		pages:       NewPages(),
	}
	app.buildLayout()
	return app
}

func (a *App) buildLayout() {
	// Set global background
	tview.Styles.PrimitiveBackgroundColor = ColorBg
	tview.Styles.ContrastBackgroundColor = ColorBgLight
	tview.Styles.MoreContrastBackgroundColor = ColorBgDark
	tview.Styles.BorderColor = ColorBorder
	tview.Styles.TitleColor = ColorAccent
	tview.Styles.PrimaryTextColor = ColorFg
	tview.Styles.SecondaryTextColor = ColorFgDim

	a.main = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header, 1, 0, false).
		AddItem(a.crumbs, 1, 0, false).
		AddItem(a.pages, 0, 1, true).
		AddItem(a.menu, 1, 0, false)

	a.main.SetBackgroundColor(ColorBg)
	a.SetRoot(a.main, true)
}

// Header returns the header component.
func (a *App) Header() *Header {
	return a.header
}

// Crumbs returns the breadcrumb component.
func (a *App) Crumbs() *Crumbs {
	return a.crumbs
}

// Menu returns the menu component.
func (a *App) Menu() *Menu {
	return a.menu
}

// Pages returns the pages component.
func (a *App) Pages() *Pages {
	return a.pages
}

// SetContent sets the main content area (used by views).
func (a *App) SetContent(p tview.Primitive) {
	a.content = p
	a.pages.SetContent(p)
}

// Header displays app title, namespace, and connection status.
type Header struct {
	*tview.TextView
	namespace string
	connected bool
}

// NewHeader creates a new header component.
func NewHeader() *Header {
	h := &Header{
		TextView:  tview.NewTextView(),
		namespace: "default",
		connected: true,
	}
	h.SetDynamicColors(true)
	h.SetBackgroundColor(ColorHeader)
	h.SetTextColor(ColorFg)
	h.render()
	return h
}

// SetNamespace updates the displayed namespace.
func (h *Header) SetNamespace(ns string) {
	h.namespace = ns
	h.render()
}

// SetConnected updates the connection status.
func (h *Header) SetConnected(connected bool) {
	h.connected = connected
	h.render()
}

func (h *Header) render() {
	connColor := TagCompleted
	connText := "connected"
	if !h.connected {
		connColor = TagFailed
		connText = "disconnected"
	}

	// Charm-style: clean, minimal, spaced
	text := fmt.Sprintf(
		" [%s::b]%s[-]  [%s]%s[-]  [%s]%s[-]",
		TagAccent, LogoSmall,
		TagFgDim, h.namespace,
		connColor, connText,
	)
	h.SetText(text)
}

// SetText allows overriding the header text temporarily.
func (h *Header) SetText(text string) {
	h.TextView.SetText(text)
}
