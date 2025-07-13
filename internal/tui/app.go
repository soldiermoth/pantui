package tui

import (
	"fmt"
	"github.com/soldiermoth/pantui/internal/hls"
	"github.com/soldiermoth/pantui/internal/tui/components"
	"github.com/soldiermoth/pantui/internal/tui/views"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the main TUI application
type App struct {
	app            *tview.Application
	pages          *tview.Pages
	parser         *hls.Parser
	navStack       []*views.ViewState
	currentView    views.View
	statusBar      *components.StatusBar
	keyBar         *components.KeyBar
	layout         *tview.Flex
	loadingModal   *tview.Modal
	loadingTicker  *time.Ticker
	spinnerIndex   int
}

// NewApp creates a new TUI application
func NewApp() *App {
	app := &App{
		app:      tview.NewApplication(),
		pages:    tview.NewPages(),
		parser:   hls.NewParser(),
		navStack: make([]*views.ViewState, 0),
	}

	app.setupLayout()
	app.setupKeybindings()
	
	return app
}

// setupLayout sets up the main application layout
func (a *App) setupLayout() {
	a.statusBar = components.NewStatusBar()
	a.keyBar = components.NewKeyBar()
	
	// Main layout: pages + status bar + key bar
	a.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.pages, 0, 1, true).
		AddItem(a.statusBar.GetPrimitive(), 1, 0, false).
		AddItem(a.keyBar.GetPrimitive(), 1, 0, false)
	
	a.app.SetRoot(a.layout, true)
}

// setupKeybindings sets up global key bindings
func (a *App) setupKeybindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		case tcell.KeyEscape:
			if len(a.navStack) > 0 {
				a.navigateBack()
				return nil
			}
			a.app.Stop()
			return nil
		case tcell.KeyF1:
			a.showHelp()
			return nil
		}
		
		// Pass to current view
		if a.currentView != nil {
			return a.currentView.HandleKey(event)
		}
		
		return event
	})
}

// RunWithURL runs the application with a manifest URL
func (a *App) RunWithURL(url string) error {
	// Parse manifest synchronously for initial load
	manifest, err := a.parser.ParseFromURL(url)
	if err != nil {
		return fmt.Errorf("failed to parse manifest from URL: %w", err)
	}

	a.showManifest(manifest)
	return a.app.Run()
}

// RunWithFile runs the application with a manifest file
func (a *App) RunWithFile(filePath string) error {
	// Parse manifest synchronously for initial load
	manifest, err := a.parser.ParseFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse manifest from file: %w", err)
	}

	a.showManifest(manifest)
	return a.app.Run()
}

// showManifest displays the manifest in the appropriate view
func (a *App) showManifest(manifest *hls.Manifest) {
	switch manifest.Type {
	case hls.MasterManifest:
		a.showMasterManifest(manifest)
	case hls.MediaManifest:
		a.showMediaManifest(manifest)
	}
}

// showMasterManifest displays a master manifest
func (a *App) showMasterManifest(manifest *hls.Manifest) {
	view := views.NewMasterView(manifest, a.parser)
	view.SetNavigationCallback(func(uri string) {
		a.navigateToSubManifest(uri)
	})
	view.SetStatusCallback(func(status string) {
		a.statusBar.SetStatus(status)
	})
	view.SetUpdateCallback(func(updateFunc func()) {
		a.app.QueueUpdateDraw(updateFunc)
	})
	
	a.setCurrentView(view, &views.ViewState{
		Type:     views.MasterViewType,
		Manifest: manifest,
		Title:    fmt.Sprintf("Master Manifest - %s", manifest.URL),
	})
}

// showMediaManifest displays a media manifest
func (a *App) showMediaManifest(manifest *hls.Manifest) {
	view := views.NewMediaView(manifest, a.parser)
	view.SetNavigationCallback(func(uri string) {
		// Create a minimal segment object for backwards compatibility
		segment := &hls.Segment{URI: uri}
		a.navigateToSegment(segment)
	})
	view.SetSegmentNavigationCallback(func(segment *hls.Segment) {
		a.navigateToSegment(segment)
	})
	view.SetStatusCallback(func(status string) {
		a.statusBar.SetStatus(status)
	})
	view.SetUpdateCallback(func(updateFunc func()) {
		a.app.QueueUpdateDraw(updateFunc)
	})
	
	a.setCurrentView(view, &views.ViewState{
		Type:     views.MediaViewType,
		Manifest: manifest,
		Title:    fmt.Sprintf("Media Manifest - %s", manifest.URL),
	})
}

// navigateToSubManifest navigates to a sub-manifest
func (a *App) navigateToSubManifest(uri string) {
	resolvedURL := a.parser.ResolveURL(uri)
	
	// Show loading modal
	a.showLoadingModal(fmt.Sprintf("Fetching manifest: %s", resolvedURL))
	
	// Parse manifest in a goroutine to allow UI updates
	go func() {
		var manifest *hls.Manifest
		var err error
		
		// Create a fresh parser instance to avoid state issues
		freshParser := hls.NewParser()
		
		// Try parsing as URL first, then as file if it's a local path
		if strings.HasPrefix(resolvedURL, "http://") || strings.HasPrefix(resolvedURL, "https://") {
			manifest, err = freshParser.ParseFromURL(resolvedURL)
		} else {
			// It's a local file path
			manifest, err = freshParser.ParseFromFile(resolvedURL)
		}
		
		a.app.QueueUpdateDraw(func() {
			a.hideLoadingModal()
			if err != nil {
				a.statusBar.SetError(fmt.Sprintf("Failed to load manifest: %v", err))
				return
			}
			a.showManifest(manifest)
		})
	}()
}

// navigateToSegment shows segment details
func (a *App) navigateToSegment(segment *hls.Segment) {
	resolvedURL := a.parser.ResolveURL(segment.URI)
	
	view := views.NewSegmentView(segment, resolvedURL)
	view.SetStatusCallback(func(status string) {
		a.statusBar.SetStatus(status)
	})
	view.SetUpdateCallback(func(updateFunc func()) {
		a.app.QueueUpdateDraw(updateFunc)
	})
	
	a.setCurrentView(view, &views.ViewState{
		Type:  views.SegmentViewType,
		Title: fmt.Sprintf("Segment - %s", segment.URI),
	})
}

// setCurrentView sets the current view and updates the navigation stack
func (a *App) setCurrentView(view views.View, state *views.ViewState) {
	// Save current view state if there is one
	if a.currentView != nil {
		a.navStack = append(a.navStack, a.getCurrentViewState())
	}
	
	a.currentView = view
	a.pages.AddAndSwitchToPage(state.Type.String(), view.GetPrimitive(), true)
	
	// Update status and key bars
	a.statusBar.SetStatus(state.Title)
	a.keyBar.SetKeys(view.GetKeyBindings())
}

// getCurrentViewState gets the current view state
func (a *App) getCurrentViewState() *views.ViewState {
	if a.currentView == nil {
		return nil
	}
	
	return &views.ViewState{
		Type:     a.currentView.GetType(),
		Manifest: a.currentView.GetManifest(),
		Title:    a.statusBar.GetStatus(),
	}
}

// navigateBack navigates back to the previous view
func (a *App) navigateBack() {
	if len(a.navStack) == 0 {
		return
	}
	
	// Pop the last state
	lastState := a.navStack[len(a.navStack)-1]
	a.navStack = a.navStack[:len(a.navStack)-1]
	
	// Create view based on state
	var view views.View
	switch lastState.Type {
	case views.MasterViewType:
		view = views.NewMasterView(lastState.Manifest, a.parser)
		view.SetNavigationCallback(func(uri string) {
			a.navigateToSubManifest(uri)
		})
		view.SetStatusCallback(func(status string) {
			a.statusBar.SetStatus(status)
		})
		view.SetUpdateCallback(func(updateFunc func()) {
			a.app.QueueUpdateDraw(updateFunc)
		})
	case views.MediaViewType:
		view = views.NewMediaView(lastState.Manifest, a.parser)
		view.SetNavigationCallback(func(uri string) {
			// Create a minimal segment object for backwards compatibility
			segment := &hls.Segment{URI: uri}
			a.navigateToSegment(segment)
		})
		view.SetSegmentNavigationCallback(func(segment *hls.Segment) {
			a.navigateToSegment(segment)
		})
		view.SetStatusCallback(func(status string) {
			a.statusBar.SetStatus(status)
		})
		view.SetUpdateCallback(func(updateFunc func()) {
			a.app.QueueUpdateDraw(updateFunc)
		})
	case views.SegmentViewType:
		// For segment view, we need to go back to the previous media view
		if len(a.navStack) > 0 {
			a.navigateBack()
			return
		}
	}
	
	a.currentView = view
	a.pages.AddAndSwitchToPage(lastState.Type.String(), view.GetPrimitive(), true)
	
	// Update status and key bars
	a.statusBar.SetStatus(lastState.Title)
	a.keyBar.SetKeys(view.GetKeyBindings())
}

// showHelp shows the help dialog
func (a *App) showHelp() {
	helpView := views.NewHelpView()
	
	modal := tview.NewModal().
		SetText(helpView.GetContent()).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("help")
		})
	
	a.pages.AddPage("help", modal, false, true)
}

// Stop stops the application
func (a *App) Stop() {
	a.app.Stop()
}

// showLoadingModal shows a loading modal with the given message
func (a *App) showLoadingModal(message string) {
	// Stop any existing loading animation
	a.hideLoadingModal()
	
	// Create animated loading modal
	a.loadingModal = tview.NewModal().
		SetBackgroundColor(tcell.ColorDarkBlue)
	
	// Spinner characters for animation
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	a.spinnerIndex = 0
	
	// Set initial text
	spinner := spinnerFrames[0]
	a.loadingModal.SetText(fmt.Sprintf("%s %s\n\nPlease wait...", spinner, message))
	
	// Add to pages
	a.pages.AddPage("loading", a.loadingModal, false, true)
	
	// Start animation ticker after adding to pages
	a.loadingTicker = time.NewTicker(100 * time.Millisecond)
	go func() {
		for range a.loadingTicker.C {
			a.app.QueueUpdateDraw(func() {
				if a.loadingModal != nil {
					spinner := spinnerFrames[a.spinnerIndex%len(spinnerFrames)]
					a.loadingModal.SetText(fmt.Sprintf("%s %s\n\nPlease wait...", spinner, message))
					a.spinnerIndex++
				}
			})
		}
	}()
}

// hideLoadingModal hides the loading modal
func (a *App) hideLoadingModal() {
	// Stop animation ticker
	if a.loadingTicker != nil {
		a.loadingTicker.Stop()
		a.loadingTicker = nil
	}
	
	// Remove loading page
	a.pages.RemovePage("loading")
	a.loadingModal = nil
}

// showErrorModal shows an error modal with the given message
func (a *App) showErrorModal(message string) {
	modal := tview.NewModal().
		SetText(message).
		SetBackgroundColor(tcell.ColorDarkRed).
		AddButtons([]string{"Exit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.app.Stop()
		})
	
	a.pages.AddPage("error", modal, false, true)
}

// GetApp returns the underlying tview application
func (a *App) GetApp() *tview.Application {
	return a.app
}
