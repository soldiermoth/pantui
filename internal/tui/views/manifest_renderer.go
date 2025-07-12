package views

import (
	"fmt"
	"pantui/internal/hls"
	"pantui/internal/tui/colors"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ManifestRenderer handles rendering HLS manifests with syntax highlighting
type ManifestRenderer struct {
	manifest      *hls.Manifest
	highlightLine int  // Line number to highlight (0 = no highlight)
}

// NewManifestRenderer creates a new manifest renderer
func NewManifestRenderer(manifest *hls.Manifest) *ManifestRenderer {
	return &ManifestRenderer{
		manifest:      manifest,
		highlightLine: 0,
	}
}

// SetHighlightLine sets the line number to highlight
func (mr *ManifestRenderer) SetHighlightLine(lineNum int) {
	mr.highlightLine = lineNum
}

// RenderColorized returns the manifest content with tview color tags
func (mr *ManifestRenderer) RenderColorized() string {
	if mr.manifest == nil || mr.manifest.Content == "" {
		return "[red]No manifest content available[white]"
	}

	lines := strings.Split(mr.manifest.Content, "\n")
	var colorizedLines []string
	
	for i, line := range lines {
		colorizedLine := mr.colorizeLine(line, i+1)
		colorizedLines = append(colorizedLines, colorizedLine)
	}
	
	return strings.Join(colorizedLines, "\n")
}

// GetNavigableItems returns a map of line numbers to URIs for navigation
func (mr *ManifestRenderer) GetNavigableItems() map[int]string {
	navigableItems := make(map[int]string)
	
	if mr.manifest == nil {
		return navigableItems
	}
	
	lines := strings.Split(mr.manifest.Content, "\n")
	
	for i, line := range lines {
		lineNum := i + 1
		line = strings.TrimSpace(line)
		
		if line == "" {
			continue
		}
		
		// Check if this line is a URI (doesn't start with #)
		if !strings.HasPrefix(line, "#") {
			navigableItems[lineNum] = line
			continue
		}
		
		// Check for URIs within HLS tags
		uri := mr.extractURIFromTag(line)
		if uri != "" {
			navigableItems[lineNum] = uri
		}
	}
	
	return navigableItems
}

// extractURIFromTag extracts URI from HLS tags that contain URI attributes
func (mr *ManifestRenderer) extractURIFromTag(line string) string {
	// Handle EXT-X-I-FRAME-STREAM-INF tags with URI attribute
	if strings.HasPrefix(line, "#EXT-X-I-FRAME-STREAM-INF:") {
		return mr.extractAttributeValue(line, "URI")
	}
	
	// Handle EXT-X-MEDIA tags with URI attribute (audio, subtitles, etc.)
	if strings.HasPrefix(line, "#EXT-X-MEDIA:") {
		return mr.extractAttributeValue(line, "URI")
	}
	
	return ""
}

// extractAttributeValue extracts the value of a specific attribute from an HLS tag
func (mr *ManifestRenderer) extractAttributeValue(line, attributeName string) string {
	// Find the attribute in the line
	attrPrefix := attributeName + "="
	attrIndex := strings.Index(line, attrPrefix)
	if attrIndex == -1 {
		return ""
	}
	
	// Extract the value after the attribute name
	valueStart := attrIndex + len(attrPrefix)
	if valueStart >= len(line) {
		return ""
	}
	
	// Handle quoted values
	if line[valueStart] == '"' {
		// Find the closing quote
		valueEnd := strings.Index(line[valueStart+1:], "\"")
		if valueEnd == -1 {
			return ""
		}
		return line[valueStart+1 : valueStart+1+valueEnd]
	}
	
	// Handle unquoted values (find comma or end of line)
	valueEnd := strings.Index(line[valueStart:], ",")
	if valueEnd == -1 {
		return line[valueStart:]
	}
	
	return line[valueStart : valueStart+valueEnd]
}

// colorizeLine applies syntax highlighting to a single line
func (mr *ManifestRenderer) colorizeLine(line string, lineNum int) string {
	line = strings.TrimSpace(line)
	
	if line == "" {
		return ""
	}
	
	var colorizedLine string
	
	// Comment lines
	if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#EXT") {
		colorizedLine = mr.colorText(line, colors.CommentColor)
	} else if strings.HasPrefix(line, "#EXT") {
		// EXT tags
		colorizedLine = mr.colorizeExtTag(line)
	} else if !strings.HasPrefix(line, "#") {
		// URI lines (not starting with #)
		colorizedLine = mr.colorText(line, colors.URIColor)
	} else {
		colorizedLine = line
	}
	
	// Add highlighting if this is the selected line
	if lineNum == mr.highlightLine {
		// Check if this line contains a URI within a tag
		if strings.HasPrefix(line, "#EXT") {
			uri := mr.extractURIFromTag(line)
			if uri != "" {
				// Highlight the entire line but emphasize the URI
				return fmt.Sprintf("[black:white]> %s[-:-]", mr.highlightURIInTag(line, uri))
			}
		}
		// Add background highlight and selection indicator for regular lines
		return fmt.Sprintf("[black:white]> %s[-:-]", colorizedLine)
	}
	
	// Add space for alignment with highlighted lines
	return fmt.Sprintf("  %s", colorizedLine)
}

// highlightURIInTag highlights the URI portion within an HLS tag
func (mr *ManifestRenderer) highlightURIInTag(line, uri string) string {
	// Find the URI in the line and highlight it
	uriIndex := strings.Index(line, fmt.Sprintf(`"%s"`, uri))
	if uriIndex == -1 {
		uriIndex = strings.Index(line, uri)
	}
	
	if uriIndex == -1 {
		// Fallback to regular colorization
		return mr.colorizeExtTag(line)
	}
	
	// Split the line into parts: before URI, URI, after URI
	beforeURI := line[:uriIndex]
	afterURIStart := uriIndex + len(uri)
	
	// Handle quoted URIs
	if uriIndex > 0 && line[uriIndex-1] == '"' {
		beforeURI = line[:uriIndex-1]
		if afterURIStart < len(line) && line[afterURIStart] == '"' {
			afterURIStart++
		}
	}
	
	afterURI := ""
	if afterURIStart < len(line) {
		afterURI = line[afterURIStart:]
	}
	
	// Colorize each part and emphasize the URI
	colorizedBefore := mr.colorizeExtTag(beforeURI + `"`)
	colorizedURI := fmt.Sprintf("[red:white]%s[-:-]", uri) // Red text on white background for emphasis
	colorizedAfter := mr.colorizeExtTag(`"` + afterURI)
	
	return colorizedBefore + colorizedURI + colorizedAfter
}

// colorizeExtTag colorizes EXT tags with their attributes
func (mr *ManifestRenderer) colorizeExtTag(line string) string {
	// Split tag name from value
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		// Tag without value
		return mr.colorText(line, colors.TagColor)
	}
	
	tagName := line[:colonIdx]
	tagValue := line[colonIdx+1:]
	
	coloredTag := mr.colorText(tagName+":", colors.TagColor)
	coloredValue := mr.colorizeTagValue(tagName, tagValue)
	
	return coloredTag + coloredValue
}

// colorizeTagValue colorizes tag values based on tag type
func (mr *ManifestRenderer) colorizeTagValue(tagName, value string) string {
	switch tagName {
	case "#EXTINF":
		return mr.colorizeExtInf(value)
	case "#EXT-X-STREAM-INF":
		return mr.colorizeStreamInf(value)
	case "#EXT-X-BYTERANGE":
		return mr.colorText(value, colors.DurationColor)
	case "#EXT-X-KEY":
		return mr.colorizeKey(value)
	case "#EXT-X-TARGETDURATION", "#EXT-X-MEDIA-SEQUENCE", "#EXT-X-VERSION":
		return mr.colorText(value, colors.SequenceColor)
	default:
		return mr.colorText(value, colors.TagValueColor)
	}
}

// colorizeExtInf colorizes EXTINF duration and title
func (mr *ManifestRenderer) colorizeExtInf(value string) string {
	// EXTINF format: duration,title
	commaIdx := strings.Index(value, ",")
	if commaIdx == -1 {
		return mr.colorText(value, colors.DurationColor)
	}
	
	duration := value[:commaIdx]
	title := value[commaIdx:]
	
	return mr.colorText(duration, colors.DurationColor) + mr.colorText(title, colors.TagValueColor)
}

// colorizeStreamInf colorizes stream attributes
func (mr *ManifestRenderer) colorizeStreamInf(value string) string {
	// Parse attributes and colorize them
	parts := mr.splitAttributes(value)
	var colorizedParts []string
	
	for _, part := range parts {
		equalIdx := strings.Index(part, "=")
		if equalIdx == -1 {
			colorizedParts = append(colorizedParts, mr.colorText(part, colors.TagValueColor))
			continue
		}
		
		key := part[:equalIdx]
		val := part[equalIdx+1:]
		
		var coloredValue string
		switch key {
		case "BANDWIDTH", "AVERAGE-BANDWIDTH":
			coloredValue = mr.colorText(val, colors.BandwidthColor)
		case "RESOLUTION":
			coloredValue = mr.colorText(val, colors.ResolutionColor)
		case "CODECS":
			coloredValue = mr.colorText(val, colors.CodecsColor)
		default:
			coloredValue = mr.colorText(val, colors.TagValueColor)
		}
		
		colorizedParts = append(colorizedParts, 
			mr.colorText(key+"=", colors.TagValueColor)+coloredValue)
	}
	
	return strings.Join(colorizedParts, mr.colorText(",", colors.TagValueColor))
}

// colorizeKey colorizes encryption key attributes
func (mr *ManifestRenderer) colorizeKey(value string) string {
	parts := mr.splitAttributes(value)
	var colorizedParts []string
	
	for _, part := range parts {
		equalIdx := strings.Index(part, "=")
		if equalIdx == -1 {
			colorizedParts = append(colorizedParts, mr.colorText(part, colors.TagValueColor))
			continue
		}
		
		key := part[:equalIdx]
		val := part[equalIdx+1:]
		
		var coloredValue string
		if key == "METHOD" {
			if val == "NONE" || val == `"NONE"` {
				coloredValue = mr.colorText(val, colors.UnencryptedColor)
			} else {
				coloredValue = mr.colorText(val, colors.EncryptedColor)
			}
		} else {
			coloredValue = mr.colorText(val, colors.TagValueColor)
		}
		
		colorizedParts = append(colorizedParts, 
			mr.colorText(key+"=", colors.TagValueColor)+coloredValue)
	}
	
	return strings.Join(colorizedParts, mr.colorText(",", colors.TagValueColor))
}

// splitAttributes splits attribute string handling quoted values
func (mr *ManifestRenderer) splitAttributes(attrStr string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	
	for _, char := range attrStr {
		switch char {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(char)
		case ',':
			if !inQuotes {
				if current.Len() > 0 {
					parts = append(parts, strings.TrimSpace(current.String()))
					current.Reset()
				}
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}
	
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	
	return parts
}

// colorText applies tview color tags to text
func (mr *ManifestRenderer) colorText(text string, color tcell.Color) string {
	colorName := mr.colorToName(color)
	return fmt.Sprintf("[%s]%s[white]", colorName, text)
}

// colorToName converts tcell color to tview color name
func (mr *ManifestRenderer) colorToName(color tcell.Color) string {
	switch color {
	case tcell.ColorYellow:
		return "yellow"
	case tcell.ColorGreen:
		return "green"
	case tcell.ColorBlue:
		return "blue"
	case tcell.ColorPurple:
		return "purple"
	case tcell.ColorRed:
		return "red"
	case tcell.ColorLightCyan:
		return "aqua"
	case tcell.ColorAqua:
		return "aqua"
	case tcell.ColorDarkGray:
		return "darkgray"
	case tcell.ColorWhite:
		return "white"
	default:
		return "white"
	}
}