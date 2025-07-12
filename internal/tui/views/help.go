package views

import (
	"pantui/internal/hls"
	"pantui/internal/tui/components"

	"github.com/gdamore/tcell/v2"
)

// HelpView displays help information
type HelpView struct {
	*BaseView
	content string
}

// NewHelpView creates a new help view
func NewHelpView() *HelpView {
	hv := &HelpView{}
	hv.BaseView = NewBaseView(nil, HelpViewType, nil)
	hv.setupContent()
	
	return hv
}

// setupContent sets up the help content
func (hv *HelpView) setupContent() {
	hv.content = `pantui - HLS Manifest Explorer

OVERVIEW:
pantui is a terminal user interface for exploring HLS (HTTP Live Streaming) 
manifests. Navigate through master manifests, media manifests, and segments 
with an intuitive interface similar to k9s.

GLOBAL KEYS:
  F1                Show this help
  Esc               Go back / Exit application
  Ctrl+C            Exit application

MASTER MANIFEST VIEW:
  ↑↓                Navigate variant streams
  Enter             Open selected variant manifest
  d                 Show variant details
  r                 Refresh manifest

MEDIA MANIFEST VIEW:
  ↑↓                Navigate segments
  Enter             Open selected segment details
  d                 Show segment details
  s                 Show manifest summary
  r                 Refresh manifest

SEGMENT VIEW:
  c                 Copy segment URL to clipboard
  o                 Open segment in browser
  h                 Show HTTP headers
  i                 Inspect segment content

NAVIGATION:
- Use arrow keys to navigate through lists
- Press Enter to drill down into sub-manifests or segments
- Press Esc to go back to the previous view
- Navigation stack preserves your path for easy backtracking

FEATURES:
- Colorized manifest display
- Support for both master and media manifests
- URL and file path input
- Detailed segment information
- Encryption status display
- Human-readable duration and bandwidth formatting

USAGE EXAMPLES:
  pantui -u https://example.com/master.m3u8
  pantui -f ./local_manifest.m3u8

For more information, visit: https://github.com/user/pantui`
}

// GetContent returns the help content
func (hv *HelpView) GetContent() string {
	return hv.content
}

// HandleKey handles key events for the help view
func (hv *HelpView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	// Help view doesn't handle any special keys
	return event
}

// GetKeyBindings returns empty key bindings for help view
func (hv *HelpView) GetKeyBindings() []components.KeyBinding {
	return []components.KeyBinding{}
}

// GetManifest returns nil for help view
func (hv *HelpView) GetManifest() *hls.Manifest {
	return nil
}
