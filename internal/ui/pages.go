package ui

import (
	"github.com/rivo/tview"
)

// Component interface for all views.
type Component interface {
	tview.Primitive
	Name() string
	Start()
	Stop()
	Hints() []KeyHint
}

// Pages manages a stack of views with push/pop navigation.
type Pages struct {
	*tview.Pages
	stack    []Component
	onChange func(Component)
}

// NewPages creates a new pages component.
func NewPages() *Pages {
	return &Pages{
		Pages: tview.NewPages(),
		stack: []Component{},
	}
}

// Push adds a view to the stack and shows it.
func (p *Pages) Push(c Component) {
	// Stop current view
	if len(p.stack) > 0 {
		p.stack[len(p.stack)-1].Stop()
	}

	p.stack = append(p.stack, c)
	p.AddPage(c.Name(), c, true, true)
	c.Start()

	if p.onChange != nil {
		p.onChange(c)
	}
}

// Pop removes the current view and returns to the previous one.
func (p *Pages) Pop() Component {
	if len(p.stack) <= 1 {
		return nil
	}

	// Stop and remove current
	current := p.stack[len(p.stack)-1]
	current.Stop()
	p.RemovePage(current.Name())
	p.stack = p.stack[:len(p.stack)-1]

	// Show previous
	prev := p.stack[len(p.stack)-1]
	p.SwitchToPage(prev.Name())
	prev.Start()

	if p.onChange != nil {
		p.onChange(prev)
	}

	return current
}

// Current returns the current view.
func (p *Pages) Current() Component {
	if len(p.stack) == 0 {
		return nil
	}
	return p.stack[len(p.stack)-1]
}

// SetContent sets content for overlay purposes.
func (p *Pages) SetContent(prim tview.Primitive) {
	// This is used when we need to set content directly
	if len(p.stack) > 0 {
		name := p.stack[len(p.stack)-1].Name()
		p.AddPage(name+"-content", prim, true, true)
	}
}

// CanPop returns true if there's more than one view on the stack.
func (p *Pages) CanPop() bool {
	return len(p.stack) > 1
}

// Depth returns the number of views on the stack.
func (p *Pages) Depth() int {
	return len(p.stack)
}

// SetOnChange sets a callback for when the active view changes.
func (p *Pages) SetOnChange(fn func(Component)) {
	p.onChange = fn
}

// Clear removes all views from the stack.
func (p *Pages) Clear() {
	for _, c := range p.stack {
		c.Stop()
		p.RemovePage(c.Name())
	}
	p.stack = []Component{}
}
