package views

import (
	"fmt"
	"os/exec"
	"github.com/soldiermoth/pantui/internal/hls"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MasterView displays a master manifest
type MasterView struct {
	*BaseView
	textView      *tview.TextView
	manifest      *hls.Manifest
	parser        *hls.Parser
	renderer      *ManifestRenderer
	navigableItems map[int]string
	currentLine   int
}

// NewMasterView creates a new master manifest view
func NewMasterView(manifest *hls.Manifest, parser *hls.Parser) *MasterView {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetScrollable(true)

	renderer := NewManifestRenderer(manifest)
	
	mv := &MasterView{
		textView:      textView,
		manifest:      manifest,
		parser:        parser,
		renderer:      renderer,
		navigableItems: renderer.GetNavigableItems(),
		currentLine:   1,
	}

	mv.BaseView = NewBaseView(textView, MasterViewType, manifest)
	mv.setupContent()
	mv.setupKeyBindings()
	mv.setupInputCapture()

	return mv
}

// setupContent sets up the manifest content with syntax highlighting
func (mv *MasterView) setupContent() {
	if mv.manifest == nil {
		mv.textView.SetText("[red]No manifest data available[white]")
		mv.textView.SetTitle(" Master Manifest - Error ").SetBorder(true)
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
	variantCount := len(mv.manifest.Variants)
	title := fmt.Sprintf(" Master Manifest - %d variants", variantCount)
	mv.textView.SetTitle(title + " ").SetBorder(true)
}

// setupInputCapture sets up input capture for navigation
func (mv *MasterView) setupInputCapture() {
	mv.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			// Find URI on current line and navigate to it
			if uri, exists := mv.navigableItems[mv.currentLine]; exists {
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
func (mv *MasterView) navigateUp() {
	for line := mv.currentLine - 1; line >= 1; line-- {
		if _, exists := mv.navigableItems[line]; exists {
			mv.currentLine = line
			mv.highlightCurrentLine()
			return
		}
	}
}

// navigateDown moves to the next navigable line
func (mv *MasterView) navigateDown() {
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
func (mv *MasterView) highlightCurrentLine() {
	// Update the renderer with the new highlight line
	mv.renderer.SetHighlightLine(mv.currentLine)
	
	// Re-render the content with highlighting
	colorizedContent := mv.renderer.RenderColorized()
	mv.textView.SetText(colorizedContent)
	
	// Only scroll if the current line is not visible
	mv.scrollToLineIfNeeded(mv.currentLine)
}

// scrollToLineIfNeeded scrolls to the line only if it's not currently visible
func (mv *MasterView) scrollToLineIfNeeded(lineNum int) {
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

// setupKeyBindings sets up key bindings for the master view
func (mv *MasterView) setupKeyBindings() {
	mv.AddKeyBinding("Enter", "Open Variant")
	mv.AddKeyBinding("↑↓", "Navigate")
	mv.AddKeyBinding("p", "Play")
	mv.AddKeyBinding("d", "Details")
	mv.AddKeyBinding("r", "Refresh")
}

// formatBandwidth formats bandwidth in human-readable format
func (mv *MasterView) formatBandwidth(bandwidth int) string {
	if bandwidth >= 1000000 {
		return fmt.Sprintf("%.1f Mbps", float64(bandwidth)/1000000)
	} else if bandwidth >= 1000 {
		return fmt.Sprintf("%.1f Kbps", float64(bandwidth)/1000)
	}
	return fmt.Sprintf("%d bps", bandwidth)
}

// HandleKey handles key events for the master view
func (mv *MasterView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'p':
		mv.playManifest()
		return nil
	case 'd':
		mv.showDetails()
		return nil
	case 'r':
		mv.refresh()
		return nil
	}

	// Let the text view handle other keys (Enter is handled in input capture)
	return event
}

// playManifest plays the current manifest using ffplay
func (mv *MasterView) playManifest() {
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

// showDetails shows detailed information about the selected variant
func (mv *MasterView) showDetails() {
	// Find the variant for the current line
	uri, exists := mv.navigableItems[mv.currentLine]
	if !exists {
		return
	}
	
	// Find the variant by URI
	var variant *hls.Variant
	for _, v := range mv.manifest.Variants {
		if v.URI == uri {
			variant = &v
			break
		}
	}
	if variant == nil {
		return
	}
	
	details := fmt.Sprintf(`Variant Stream Details:

URI: %s
Bandwidth: %s
Resolution: %s
Codecs: %s

Additional Attributes:`, 
		variant.URI,
		mv.formatBandwidth(variant.Bandwidth),
		variant.Resolution,
		variant.Codecs)

	for key, value := range variant.Attributes {
		if key != "BANDWIDTH" && key != "RESOLUTION" && key != "CODECS" {
			details += fmt.Sprintf("\n%s: %s", key, value)
		}
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

// refresh refreshes the manifest data
func (mv *MasterView) refresh() {
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
					mv.statusCallback(fmt.Sprintf("Master Manifest - %s", mv.manifest.URL))
				}
			})
		}
	}()
}
