package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// KeyBinding represents a key binding
type KeyBinding struct {
	Key         string
	Description string
}

// KeyBar represents the key bindings bar
type KeyBar struct {
	textView *tview.TextView
	bindings []KeyBinding
}

// NewKeyBar creates a new key bar
func NewKeyBar() *KeyBar {
	kb := &KeyBar{
		textView: tview.NewTextView(),
		bindings: []KeyBinding{
			{"F1", "Help"},
			{"Esc", "Back/Quit"},
			{"Ctrl+C", "Exit"},
		},
	}
	
	kb.textView.
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetBorder(false).
		SetBackgroundColor(tcell.ColorDarkGray)
	
	kb.updateDisplay()
	
	return kb
}

// SetKeys sets the key bindings
func (kb *KeyBar) SetKeys(bindings []KeyBinding) {
	// Always include global bindings
	globalBindings := []KeyBinding{
		{"F1", "Help"},
		{"Esc", "Back/Quit"},
		{"Ctrl+C", "Exit"},
	}
	
	kb.bindings = append(bindings, globalBindings...)
	kb.updateDisplay()
}

// AddKey adds a key binding
func (kb *KeyBar) AddKey(key, description string) {
	kb.bindings = append(kb.bindings, KeyBinding{
		Key:         key,
		Description: description,
	})
	kb.updateDisplay()
}

// updateDisplay updates the key bar display
func (kb *KeyBar) updateDisplay() {
	var parts []string
	
	for _, binding := range kb.bindings {
		part := fmt.Sprintf("[white]%s[darkgray]=%s", binding.Key, binding.Description)
		parts = append(parts, part)
	}
	
	content := " " + strings.Join(parts, " [darkgray]|[white] ")
	kb.textView.SetText(content)
}

// GetPrimitive returns the underlying tview primitive
func (kb *KeyBar) GetPrimitive() tview.Primitive {
	return kb.textView
}
