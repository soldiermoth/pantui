package views

import (
	"fmt"
	"os/exec"
	"github.com/soldiermoth/pantui/internal/hls"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MediaView displays a media manifest
type MediaView struct {
	*BaseView
	textView      *tview.TextView
	manifest      *hls.Manifest
	parser        *hls.Parser
	renderer      *ManifestRenderer
	navigableItems map[int]string
	currentLine   int
}

// NewMediaView creates a new media manifest view
func NewMediaView(manifest *hls.Manifest, parser *hls.Parser) *MediaView {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetScrollable(true)

	renderer := NewManifestRenderer(manifest)
	
	mv := &MediaView{
		textView:      textView,
		manifest:      manifest,
		parser:        parser,
		renderer:      renderer,
		navigableItems: renderer.GetNavigableItems(),
		currentLine:   1,
	}

	mv.BaseView = NewBaseView(textView, MediaViewType, manifest)
	mv.setupContent()
	mv.setupKeyBindings()
	mv.setupInputCapture()

	return mv
}

// setupContent sets up the manifest content with syntax highlighting
func (mv *MediaView) setupContent() {
	if mv.manifest == nil {
		mv.textView.SetText("[red]No manifest data available[white]")
		mv.textView.SetTitle(" Media Manifest - Error ").SetBorder(true)
		return
	}

	// Find the first navigable line and set it as current
	for lineNum := 1; lineNum <= len(strings.Split(mv.manifest.Content, "\n")); lineNum++ {
		if _, exists := mv.navigableItems[lineNum]; exists {
			mv.currentLine = lineNum
			break
		}
	}
	
	// Set the highlight line in the renderer
	mv.renderer.SetHighlightLine(mv.currentLine)
	
	// Render the colorized manifest content
	colorizedContent := mv.renderer.RenderColorized()
	mv.textView.SetText(colorizedContent)
	
	// Set title with manifest info
	segmentCount := len(mv.manifest.Segments)
	title := fmt.Sprintf(" Media Manifest - %d segments", segmentCount)
	if mv.manifest.TargetDuration > 0 {
		title += fmt.Sprintf(" (Target: %ds)", mv.manifest.TargetDuration)
	}
	mv.textView.SetTitle(title + " ").SetBorder(true)
}

// setupInputCapture sets up input capture for navigation
func (mv *MediaView) setupInputCapture() {
	mv.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			// Find URI on current line and navigate to it
			if uri, exists := mv.navigableItems[mv.currentLine]; exists {
				// Find the segment by URI
				for _, segment := range mv.manifest.Segments {
					if segment.URI == uri {
						if mv.segmentNavigationCallback != nil {
							mv.segmentNavigationCallback(&segment)
						}
						return nil
					}
				}
				// If no segment found, fall back to regular navigation
				if mv.navigationCallback != nil {
					mv.navigationCallback(uri)
				}
			}
			return nil
		case tcell.KeyUp:
			mv.navigateUp()
			return nil
		case tcell.KeyDown:
			mv.navigateDown()
			return nil
		}
		return event // Let other keys pass through
	})
}

// navigateUp moves to the previous navigable line
func (mv *MediaView) navigateUp() {
	// Find previous navigable line
	for line := mv.currentLine - 1; line >= 1; line-- {
		if _, exists := mv.navigableItems[line]; exists {
			mv.currentLine = line
			mv.highlightCurrentLine()
			return
		}
	}
}

// navigateDown moves to the next navigable line
func (mv *MediaView) navigateDown() {
	// Find next navigable line
	maxLine := len(strings.Split(mv.manifest.Content, "\n"))
	for line := mv.currentLine + 1; line <= maxLine; line++ {
		if _, exists := mv.navigableItems[line]; exists {
			mv.currentLine = line
			mv.highlightCurrentLine()
			return
		}
	}
}

// highlightCurrentLine highlights the current line and refreshes the view
func (mv *MediaView) highlightCurrentLine() {
	// Update the renderer with the new highlight line
	mv.renderer.SetHighlightLine(mv.currentLine)
	
	// Re-render the content with highlighting
	colorizedContent := mv.renderer.RenderColorized()
	mv.textView.SetText(colorizedContent)
	
	// Only scroll if the current line is not visible
	mv.scrollToLineIfNeeded(mv.currentLine)
}

// scrollToLineIfNeeded scrolls to the line only if it's not currently visible
func (mv *MediaView) scrollToLineIfNeeded(lineNum int) {
	if lineNum <= 0 {
		return
	}
	
	// Get current scroll position and viewport size
	_, _, _, height := mv.textView.GetInnerRect()
	currentRow, _ := mv.textView.GetScrollOffset()
	
	// Calculate if the line is visible in the current viewport
	// Account for 0-based indexing (lineNum is 1-based)
	targetRow := lineNum - 1
	bottomVisibleRow := currentRow + height - 1
	
	// Check if line is above the viewport
	if targetRow < currentRow {
		mv.textView.ScrollTo(targetRow, 0)
		return
	}
	
	// Check if line is below the viewport  
	if targetRow > bottomVisibleRow {
		// Scroll so the target line is visible, preferably near the bottom
		newScrollRow := targetRow - height + 2 // Leave some margin
		if newScrollRow < 0 {
			newScrollRow = 0
		}
		mv.textView.ScrollTo(newScrollRow, 0)
		return
	}
	
	// Line is already visible, no scrolling needed
}

// setupKeyBindings sets up key bindings for the media view
func (mv *MediaView) setupKeyBindings() {
	mv.AddKeyBinding("Enter", "Open Segment")
	mv.AddKeyBinding("↑↓", "Navigate")
	mv.AddKeyBinding("p", "Play")
	mv.AddKeyBinding("d", "Details")
	mv.AddKeyBinding("s", "Summary")
	mv.AddKeyBinding("r", "Refresh")
}

// HandleKey handles key events for the media view
func (mv *MediaView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'p':
		mv.playManifest()
		return nil
	case 'd':
		mv.showDetails()
		return nil
	case 's':
		mv.showSummary()
		return nil
	case 'r':
		mv.refresh()
		return nil
	}

	// Let the text view handle other keys
	return event
}

// playManifest plays the current manifest using ffplay
func (mv *MediaView) playManifest() {
	if mv.manifest == nil {
		if mv.statusCallback != nil {
			mv.statusCallback("No manifest to play")
		}
		return
	}

	// Show status message
	if mv.statusCallback != nil {
		mv.statusCallback(fmt.Sprintf("Launching ffplay for: %s", mv.manifest.URL))
	}

	// Launch ffplay in a goroutine so it doesn't block the UI
	go func() {
		cmd := exec.Command("ffplay", "-hide_banner", mv.manifest.URL)
		
		// Start the process
		err := cmd.Start()
		if err != nil {
			// Update status with error message
			if mv.updateCallback != nil {
				mv.updateCallback(func() {
					if mv.statusCallback != nil {
						mv.statusCallback(fmt.Sprintf("Failed to launch ffplay: %v", err))
					}
				})
			}
			return
		}

		// Optional: Wait for process to finish and update status
		go func() {
			err := cmd.Wait()
			if mv.updateCallback != nil {
				mv.updateCallback(func() {
					if mv.statusCallback != nil {
						if err != nil {
							mv.statusCallback(fmt.Sprintf("ffplay exited with error: %v", err))
						} else {
							mv.statusCallback("ffplay finished")
						}
					}
				})
			}
		}()
	}()
}

// showDetails shows detailed information about the selected segment
func (mv *MediaView) showDetails() {
	// Find the segment for the current line
	uri, exists := mv.navigableItems[mv.currentLine]
	if !exists {
		return
	}
	
	// Find the segment by URI
	var segment *hls.Segment
	for _, s := range mv.manifest.Segments {
		if s.URI == uri {
			segment = &s
			break
		}
	}
	if segment == nil {
		return
	}
	
	details := fmt.Sprintf(`Segment Details:

Sequence: %d
Duration: %.3f seconds
URI: %s
Byte Range: %s`,
		segment.Sequence,
		segment.Duration,
		segment.URI,
		segment.ByteRange)

	if segment.Key != nil {
		details += fmt.Sprintf(`

Encryption:
Method: %s
URI: %s
IV: %s
Key Format: %s`,
			segment.Key.Method,
			segment.Key.URI,
			segment.Key.IV,
			segment.Key.KeyFormat)
	}

	// Create a modal to show details
	modal := tview.NewModal().
		SetText(details).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// Close modal - this would need to be handled by the parent app
		})

	// This would need to be handled by the parent application
	_ = modal
}

// showSummary shows a summary of the manifest
func (mv *MediaView) showSummary() {
	totalDuration := 0.0
	encryptedSegments := 0
	
	for _, segment := range mv.manifest.Segments {
		totalDuration += segment.Duration
		if segment.Key != nil && segment.Key.Method != "NONE" && segment.Key.Method != "" {
			encryptedSegments++
		}
	}

	summary := fmt.Sprintf(`Media Manifest Summary:

Version: %d
Target Duration: %d seconds
Media Sequence: %d
Total Segments: %d
Total Duration: %s
Encrypted Segments: %d

Base URL: %s`,
		mv.manifest.Version,
		mv.manifest.TargetDuration,
		mv.manifest.Sequence,
		len(mv.manifest.Segments),
		mv.formatDuration(totalDuration),
		encryptedSegments,
		mv.manifest.BaseURL)

	// Create a modal to show summary
	modal := tview.NewModal().
		SetText(summary).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// Close modal - this would need to be handled by the parent app
		})

	// This would need to be handled by the parent application
	_ = modal
}

// formatDuration formats duration in human-readable format
func (mv *MediaView) formatDuration(seconds float64) string {
	duration := time.Duration(seconds * float64(time.Second))
	
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// refresh refreshes the manifest data
func (mv *MediaView) refresh() {
	// Show loading indicator via status bar
	if mv.statusCallback != nil {
		mv.statusCallback(fmt.Sprintf("Refreshing manifest: %s", mv.manifest.URL))
	}
	
	// Re-parse the manifest
	go func() {
		newManifest, err := mv.parser.ParseFromURL(mv.manifest.URL)
		// Use QueueUpdateDraw to update UI from goroutine
		if mv.updateCallback != nil {
			mv.updateCallback(func() {
				if err != nil {
					if mv.statusCallback != nil {
						mv.statusCallback(fmt.Sprintf("Failed to refresh manifest: %v", err))
					}
					return
				}
				mv.manifest = newManifest
				mv.BaseView.manifest = newManifest
				mv.renderer = NewManifestRenderer(newManifest)
				mv.navigableItems = mv.renderer.GetNavigableItems()
				mv.setupContent()
				if mv.statusCallback != nil {
					title := fmt.Sprintf(" Media Manifest - %d segments", len(mv.manifest.Segments))
					if mv.manifest.TargetDuration > 0 {
						title += fmt.Sprintf(" (Target: %ds)", mv.manifest.TargetDuration)
					}
					mv.statusCallback(title)
				}
			})
		}
	}()
}
