package components

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StatusBar represents the application status bar
type StatusBar struct {
	textView *tview.TextView
	status   string
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	sb := &StatusBar{
		textView: tview.NewTextView(),
		status:   "Ready",
	}
	
	sb.textView.
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetBorder(false).
		SetBackgroundColor(tcell.ColorDarkBlue)
	
	sb.updateDisplay()
	
	return sb
}

// SetStatus sets the status message
func (sb *StatusBar) SetStatus(status string) {
	sb.status = status
	sb.updateDisplay()
}

// SetError sets an error message
func (sb *StatusBar) SetError(message string) {
	sb.status = fmt.Sprintf("[red]ERROR: %s[white]", message)
	sb.updateDisplay()
}

// SetWarning sets a warning message
func (sb *StatusBar) SetWarning(message string) {
	sb.status = fmt.Sprintf("[yellow]WARNING: %s[white]", message)
	sb.updateDisplay()
}

// SetSuccess sets a success message
func (sb *StatusBar) SetSuccess(message string) {
	sb.status = fmt.Sprintf("[green]SUCCESS: %s[white]", message)
	sb.updateDisplay()
}

// GetStatus returns the current status
func (sb *StatusBar) GetStatus() string {
	return sb.status
}

// updateDisplay updates the status bar display
func (sb *StatusBar) updateDisplay() {
	timestamp := time.Now().Format("15:04:05")
	content := fmt.Sprintf(" [white]%s[darkgray] | [white]%s", sb.status, timestamp)
	sb.textView.SetText(content)
}

// GetPrimitive returns the underlying tview primitive
func (sb *StatusBar) GetPrimitive() tview.Primitive {
	return sb.textView
}
