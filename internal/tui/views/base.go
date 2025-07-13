package views

import (
	"pantui/internal/hls"
	"pantui/internal/tui/components"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ViewType represents the type of view
type ViewType int

const (
	MasterViewType ViewType = iota
	MediaViewType
	SegmentViewType
	HelpViewType
)

// String returns the string representation of the view type
func (vt ViewType) String() string {
	switch vt {
	case MasterViewType:
		return "master"
	case MediaViewType:
		return "media"
	case SegmentViewType:
		return "segment"
	case HelpViewType:
		return "help"
	default:
		return "unknown"
	}
}

// ViewState represents the state of a view for navigation
type ViewState struct {
	Type     ViewType
	Manifest *hls.Manifest
	Title    string
}

// NavigationCallback is called when user wants to navigate to a resource
type NavigationCallback func(uri string)

// SegmentNavigationCallback is called when user wants to navigate to a segment
type SegmentNavigationCallback func(segment *hls.Segment)

// StatusCallback is called to update the status bar
type StatusCallback func(status string)

// UpdateCallback is called to queue UI updates from goroutines
type UpdateCallback func(updateFunc func())

// View represents a view in the TUI
type View interface {
	GetPrimitive() tview.Primitive
	GetType() ViewType
	GetManifest() *hls.Manifest
	GetKeyBindings() []components.KeyBinding
	HandleKey(event *tcell.EventKey) *tcell.EventKey
	SetNavigationCallback(callback NavigationCallback)
	SetSegmentNavigationCallback(callback SegmentNavigationCallback)
	SetStatusCallback(callback StatusCallback)
	SetUpdateCallback(callback UpdateCallback)
}

// BaseView provides common functionality for all views
type BaseView struct {
	primitive                 tview.Primitive
	viewType                  ViewType
	manifest                  *hls.Manifest
	keyBindings               []components.KeyBinding
	navigationCallback        NavigationCallback
	segmentNavigationCallback SegmentNavigationCallback
	statusCallback            StatusCallback
	updateCallback            UpdateCallback
}

// NewBaseView creates a new base view
func NewBaseView(primitive tview.Primitive, viewType ViewType, manifest *hls.Manifest) *BaseView {
	return &BaseView{
		primitive:   primitive,
		viewType:    viewType,
		manifest:    manifest,
		keyBindings: make([]components.KeyBinding, 0),
	}
}

// GetPrimitive returns the underlying tview primitive
func (bv *BaseView) GetPrimitive() tview.Primitive {
	return bv.primitive
}

// GetType returns the view type
func (bv *BaseView) GetType() ViewType {
	return bv.viewType
}

// GetManifest returns the associated manifest
func (bv *BaseView) GetManifest() *hls.Manifest {
	return bv.manifest
}

// GetKeyBindings returns the key bindings for this view
func (bv *BaseView) GetKeyBindings() []components.KeyBinding {
	return bv.keyBindings
}

// SetKeyBindings sets the key bindings for this view
func (bv *BaseView) SetKeyBindings(bindings []components.KeyBinding) {
	bv.keyBindings = bindings
}

// AddKeyBinding adds a key binding
func (bv *BaseView) AddKeyBinding(key, description string) {
	bv.keyBindings = append(bv.keyBindings, components.KeyBinding{
		Key:         key,
		Description: description,
	})
}

// SetNavigationCallback sets the navigation callback
func (bv *BaseView) SetNavigationCallback(callback NavigationCallback) {
	bv.navigationCallback = callback
}

// SetSegmentNavigationCallback sets the segment navigation callback
func (bv *BaseView) SetSegmentNavigationCallback(callback SegmentNavigationCallback) {
	bv.segmentNavigationCallback = callback
}

// SetStatusCallback sets the status callback
func (bv *BaseView) SetStatusCallback(callback StatusCallback) {
	bv.statusCallback = callback
}

// SetUpdateCallback sets the update callback
func (bv *BaseView) SetUpdateCallback(callback UpdateCallback) {
	bv.updateCallback = callback
}

// HandleKey handles key events (default implementation)
func (bv *BaseView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	// Default implementation - pass through
	return event
}
