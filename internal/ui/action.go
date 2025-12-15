package ui

import "github.com/gdamore/tcell/v2"

// ActionHandler is a function that handles a key action.
type ActionHandler func() bool

// KeyAction represents a key binding with its handler.
type KeyAction struct {
	Key         tcell.Key
	Rune        rune
	Description string
	Handler     ActionHandler
}

// KeyHint represents a keybinding hint for display in the menu.
type KeyHint struct {
	Key         string
	Description string
}

// ActionRegistry manages key bindings for components.
type ActionRegistry struct {
	actions map[string]*KeyAction
}

// NewActionRegistry creates a new action registry.
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]*KeyAction),
	}
}

// Add registers a key action.
func (r *ActionRegistry) Add(name string, action *KeyAction) {
	r.actions[name] = action
}

// Remove unregisters a key action.
func (r *ActionRegistry) Remove(name string) {
	delete(r.actions, name)
}

// Clear removes all actions.
func (r *ActionRegistry) Clear() {
	r.actions = make(map[string]*KeyAction)
}

// Handle processes a key event and returns true if handled.
func (r *ActionRegistry) Handle(event *tcell.EventKey) bool {
	for _, action := range r.actions {
		if matchesKey(event, action) {
			if action.Handler != nil {
				return action.Handler()
			}
		}
	}
	return false
}

// Hints returns all key hints for menu display.
func (r *ActionRegistry) Hints() []KeyHint {
	var hints []KeyHint
	for _, action := range r.actions {
		if action.Description != "" {
			hints = append(hints, KeyHint{
				Key:         keyToString(action),
				Description: action.Description,
			})
		}
	}
	return hints
}

func matchesKey(event *tcell.EventKey, action *KeyAction) bool {
	if action.Key != tcell.KeyRune {
		return event.Key() == action.Key
	}
	return event.Key() == tcell.KeyRune && event.Rune() == action.Rune
}

func keyToString(action *KeyAction) string {
	switch action.Key {
	case tcell.KeyEnter:
		return "enter"
	case tcell.KeyEscape:
		return "esc"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "backspace"
	case tcell.KeyTab:
		return "tab"
	case tcell.KeyRune:
		return string(action.Rune)
	default:
		return "?"
	}
}

// NewKeyAction creates a key action for a specific key.
func NewKeyAction(key tcell.Key, description string, handler ActionHandler) *KeyAction {
	return &KeyAction{
		Key:         key,
		Description: description,
		Handler:     handler,
	}
}

// NewRuneAction creates a key action for a rune (character).
func NewRuneAction(r rune, description string, handler ActionHandler) *KeyAction {
	return &KeyAction{
		Key:         tcell.KeyRune,
		Rune:        r,
		Description: description,
		Handler:     handler,
	}
}
