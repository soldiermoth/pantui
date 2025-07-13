package views

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"github.com/soldiermoth/pantui/internal/hls"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FFProbeOutput represents ffprobe JSON output
type FFProbeOutput struct {
	Format  FFProbeFormat  `json:"format"`
	Streams []FFProbeStream `json:"streams"`
}

// FFProbeFormat represents format information from ffprobe
type FFProbeFormat struct {
	Filename       string            `json:"filename"`
	NBStreams      int               `json:"nb_streams"`
	NBPrograms     int               `json:"nb_programs"`
	FormatName     string            `json:"format_name"`
	FormatLongName string            `json:"format_long_name"`
	StartTime      string            `json:"start_time"`
	Duration       string            `json:"duration"`
	Size           string            `json:"size"`
	BitRate        string            `json:"bit_rate"`
	ProbeScore     int               `json:"probe_score"`
	Tags           map[string]string `json:"tags"`
}

// FFProbeStream represents stream information from ffprobe
type FFProbeStream struct {
	Index              int               `json:"index"`
	CodecName          string            `json:"codec_name"`
	CodecLongName      string            `json:"codec_long_name"`
	Profile            string            `json:"profile"`
	CodecType          string            `json:"codec_type"`
	CodecTimeBase      string            `json:"codec_time_base"`
	CodecTagString     string            `json:"codec_tag_string"`
	CodecTag           string            `json:"codec_tag"`
	Width              int               `json:"width"`
	Height             int               `json:"height"`
	CodedWidth         int               `json:"coded_width"`
	CodedHeight        int               `json:"coded_height"`
	HasBFrames         int               `json:"has_b_frames"`
	SampleAspectRatio  string            `json:"sample_aspect_ratio"`
	DisplayAspectRatio string            `json:"display_aspect_ratio"`
	PixFmt             string            `json:"pix_fmt"`
	Level              int               `json:"level"`
	ColorRange         string            `json:"color_range"`
	ColorSpace         string            `json:"color_space"`
	ColorTransfer      string            `json:"color_transfer"`
	ColorPrimaries     string            `json:"color_primaries"`
	ChromaLocation     string            `json:"chroma_location"`
	RFrameRate         string            `json:"r_frame_rate"`
	AvgFrameRate       string            `json:"avg_frame_rate"`
	TimeBase           string            `json:"time_base"`
	StartPts           int               `json:"start_pts"`
	StartTime          string            `json:"start_time"`
	DurationTs         int64             `json:"duration_ts"`
	Duration           string            `json:"duration"`
	BitRate            string            `json:"bit_rate"`
	BitsPerRawSample   string            `json:"bits_per_raw_sample"`
	NBFrames           string            `json:"nb_frames"`
	SampleFmt          string            `json:"sample_fmt"`
	SampleRate         string            `json:"sample_rate"`
	Channels           int               `json:"channels"`
	ChannelLayout      string            `json:"channel_layout"`
	BitsPerSample      int               `json:"bits_per_sample"`
	Tags               map[string]string `json:"tags"`
}

// SegmentView displays segment information
type SegmentView struct {
	*BaseView
	textView    *tview.TextView
	segment     *hls.Segment
	resolvedURL string
	probeData   *FFProbeOutput
}

// NewSegmentView creates a new segment view
func NewSegmentView(segment *hls.Segment, resolvedURL string) *SegmentView {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true).
		SetScrollable(true)

	sv := &SegmentView{
		textView:    textView,
		segment:     segment,
		resolvedURL: resolvedURL,
	}

	sv.BaseView = NewBaseView(textView, SegmentViewType, nil)
	sv.setupContent()
	sv.setupKeyBindings()

	return sv
}

// setupContent sets up the content for the segment view
func (sv *SegmentView) setupContent() {
	content := fmt.Sprintf(`[yellow]Segment Information[white]

[cyan]Original URI:[white]
%s

[cyan]Resolved URL:[white]
%s

[cyan]Segment Details:[white]
%s

[cyan]URL Components:[white]
%s

[cyan]Available Actions:[white]
• Press [green]c[white] to copy URL to clipboard
• Press [green]o[white] to open in browser
• Press [green]h[white] to show HTTP headers
• Press [green]i[white] to inspect segment
• Press [yellow]Esc[white] to go back

[darkgray]Note: This view shows segment metadata. 
Actual segment content inspection would require 
downloading and analyzing the media file.[white]`,
		sv.segment.URI,
		sv.resolvedURL,
		sv.formatSegmentDetails(),
		sv.parseURL())

	sv.textView.SetText(content)
	sv.textView.SetTitle(" Segment Details ").SetBorder(true)
}

// formatSegmentDetails formats segment-specific details
func (sv *SegmentView) formatSegmentDetails() string {
	if sv.segment == nil {
		return "[red]No segment data available[white]"
	}

	details := fmt.Sprintf("Duration: %.3f seconds\nSequence: %d", 
		sv.segment.Duration, sv.segment.Sequence)

	if sv.segment.ByteRange != "" {
		details += fmt.Sprintf("\nByte Range: %s", sv.segment.ByteRange)
	}

	if sv.segment.Key != nil {
		details += fmt.Sprintf("\nEncryption: %s", sv.segment.Key.Method)
		if sv.segment.Key.URI != "" {
			details += fmt.Sprintf(" (Key URI: %s)", sv.segment.Key.URI)
		}
	}

	if sv.segment.Map != nil {
		details += fmt.Sprintf("\nInit Fragment: %s", sv.segment.Map.URI)
		if sv.segment.Map.ByteRange != "" {
			details += fmt.Sprintf(" (Range: %s)", sv.segment.Map.ByteRange)
		}
	}

	return details
}

// parseURL parses and formats URL components
func (sv *SegmentView) parseURL() string {
	if sv.resolvedURL == "" {
		return "Local file path: " + sv.segment.URI
	}

	parsedURL, err := url.Parse(sv.resolvedURL)
	if err != nil {
		return fmt.Sprintf("Invalid URL: %s", sv.resolvedURL)
	}

	return fmt.Sprintf(`Scheme: %s
Host: %s
Path: %s
Query: %s
Fragment: %s`,
		parsedURL.Scheme,
		parsedURL.Host,
		parsedURL.Path,
		parsedURL.RawQuery,
		parsedURL.Fragment)
}

// setupKeyBindings sets up key bindings for the segment view
func (sv *SegmentView) setupKeyBindings() {
	sv.AddKeyBinding("c", "Copy URL")
	sv.AddKeyBinding("o", "Open in Browser")
	sv.AddKeyBinding("h", "HTTP Headers")
	sv.AddKeyBinding("i", "Inspect")
}

// HandleKey handles key events for the segment view
func (sv *SegmentView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'c':
		sv.copyURL()
		return nil
	case 'o':
		sv.openInBrowser()
		return nil
	case 'h':
		sv.showHTTPHeaders()
		return nil
	case 'i':
		sv.inspectSegment()
		return nil
	}

	// Let the text view handle other keys (scrolling, etc.)
	return event
}

// copyURL copies the URL to clipboard
func (sv *SegmentView) copyURL() {
	url := sv.resolvedURL
	if url == "" {
		url = sv.segment.URI
	}

	// Try to copy to clipboard using system commands
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "windows":
		cmd = exec.Command("clip")
	default:
		sv.showMessage("Clipboard not supported on this platform")
		return
	}

	cmd.Stdin = strings.NewReader(url)
	err := cmd.Run()
	if err != nil {
		sv.showMessage(fmt.Sprintf("Failed to copy to clipboard: %v", err))
	} else {
		sv.showMessage(fmt.Sprintf("URL copied to clipboard: %s", url))
	}
}

// openInBrowser opens the URL in browser
func (sv *SegmentView) openInBrowser() {
	if sv.resolvedURL == "" {
		sv.showMessage("Cannot open local file in browser")
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", sv.resolvedURL)
	case "linux":
		cmd = exec.Command("xdg-open", sv.resolvedURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", sv.resolvedURL)
	default:
		sv.showMessage("Unsupported platform for opening browser")
		return
	}

	err := cmd.Start()
	if err != nil {
		sv.showMessage(fmt.Sprintf("Failed to open browser: %v", err))
	} else {
		sv.showMessage(fmt.Sprintf("Opening in browser: %s", sv.resolvedURL))
	}
}

// showHTTPHeaders shows HTTP headers for the segment
func (sv *SegmentView) showHTTPHeaders() {
	if sv.resolvedURL == "" {
		sv.showMessage("Cannot fetch headers for local file")
		return
	}

	sv.showMessage("Fetching HTTP headers...")
	
	go func() {
		resp, err := http.Head(sv.resolvedURL)
		if err != nil {
			sv.showMessage(fmt.Sprintf("Failed to fetch headers: %v", err))
			return
		}
		defer resp.Body.Close()

		var headers []string
		headers = append(headers, fmt.Sprintf("Status: %s", resp.Status))
		
		for key, values := range resp.Header {
			for _, value := range values {
				headers = append(headers, fmt.Sprintf("%s: %s", key, value))
			}
		}
		
		headerText := strings.Join(headers, "\n")
		sv.updateContentWithHeaders(headerText)
	}()
}

// updateContentWithHeaders updates the content with HTTP headers
func (sv *SegmentView) updateContentWithHeaders(headers string) {
	content := fmt.Sprintf(`[yellow]Segment Information[white]

[cyan]Original URI:[white]
%s

[cyan]Resolved URL:[white]
%s

[cyan]HTTP Headers:[white]
%s

[cyan]URL Components:[white]
%s

[cyan]Available Actions:[white]
• Press [green]c[white] to copy URL to clipboard
• Press [green]o[white] to open in browser
• Press [green]h[white] to show HTTP headers
• Press [green]i[white] to inspect segment
• Press [yellow]Esc[white] to go back`,
		sv.segment.URI,
		sv.resolvedURL,
		headers,
		sv.parseURL())

	sv.textView.SetText(content)
}

// inspectSegment inspects the segment using ffprobe
func (sv *SegmentView) inspectSegment() {
	url := sv.resolvedURL
	if url == "" {
		url = sv.segment.URI
	}

	sv.showMessage("Running ffprobe analysis...")
	
	go func() {
		// Run ffprobe to get detailed information
		// If there's an init fragment, we need to concatenate it with the segment
		probeData, err := sv.runFFProbeWithInit(url)
		if err != nil {
			if sv.updateCallback != nil {
				sv.updateCallback(func() {
					sv.showFFProbeError(err)
				})
			}
			return
		}

		sv.probeData = probeData
		
		// Update the UI with probe results
		if sv.updateCallback != nil {
			sv.updateCallback(func() {
				sv.updateContentWithProbeData()
			})
		}
	}()
}

// runFFProbeWithInit executes ffprobe with init fragment support
func (sv *SegmentView) runFFProbeWithInit(segmentURL string) (*FFProbeOutput, error) {
	// Check if we have an init fragment
	if sv.segment.Map != nil && sv.segment.Map.URI != "" {
		// We have an init fragment, create a temporary concat file
		initURL := sv.resolveInitFragmentURL(sv.segment.Map.URI)
		
		// Create temporary concat file
		concatFile, err := sv.createConcatFile(initURL, segmentURL)
		if err != nil {
			// If we can't create concat file, fall back to segment-only analysis
			concatError := fmt.Errorf("Failed to create concat file: %v", err)
			result, fallbackErr := sv.runFFProbe(segmentURL)
			if fallbackErr != nil {
				return nil, fmt.Errorf("%v\n\nFallback segment-only analysis also failed: %v", concatError, fallbackErr)
			}
			return result, nil
		}
		defer os.Remove(concatFile) // Clean up temp file
		
		cmd := exec.Command("ffprobe",
			"-v", "quiet",
			"-print_format", "json",
			"-show_format",
			"-show_streams",
			"-f", "concat",
			"-safe", "0",
			"-protocol_whitelist", "concat,file,http,https,tcp,tls",
			concatFile)

		output, err := cmd.CombinedOutput()
		if err != nil {
			// Create detailed error message with full command context
			concatError := fmt.Errorf("Init fragment concatenation failed\nCommand: ffprobe -f concat -safe 0 -protocol_whitelist concat,file,http,https,tcp,tls \"%s\"\nConcat file content:\nfile '%s'\nfile '%s'\nError: %v\nOutput: %s", 
				concatFile, initURL, segmentURL, err, string(output))
			
			// Try segment-only analysis and include concat error info
			result, fallbackErr := sv.runFFProbe(segmentURL)
			if fallbackErr != nil {
				return nil, fmt.Errorf("%v\n\nFallback segment-only analysis also failed: %v", concatError, fallbackErr)
			}
			
			// Successful fallback - we could optionally log the concat error
			return result, nil
		}

		var probeOutput FFProbeOutput
		err = json.Unmarshal(output, &probeOutput)
		if err != nil {
			// If parsing fails, fall back to analyzing just the segment
			parseError := fmt.Errorf("Init fragment concat succeeded but JSON parsing failed\nJSON Error: %v\nFFProbe Raw Output:\n%s", err, string(output))
			
			result, fallbackErr := sv.runFFProbe(segmentURL)
			if fallbackErr != nil {
				return nil, fmt.Errorf("%v\n\nFallback segment-only analysis also failed: %v", parseError, fallbackErr)
			}
			
			return result, nil
		}

		return &probeOutput, nil
	}

	// No init fragment, analyze segment directly
	return sv.runFFProbe(segmentURL)
}

// runFFProbe executes ffprobe and returns parsed results
func (sv *SegmentView) runFFProbe(url string) (*FFProbeOutput, error) {
	// ffprobe command with JSON output
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		url)

	// Use CombinedOutput to capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("FFProbe execution failed\nCommand: ffprobe -v quiet -print_format json -show_format -show_streams \"%s\"\nError: %v\nOutput:\n%s", url, err, string(output))
	}

	var probeOutput FFProbeOutput
	err = json.Unmarshal(output, &probeOutput)
	if err != nil {
		return nil, fmt.Errorf("FFProbe succeeded but JSON parsing failed\nJSON Error: %v\nFFProbe Raw Output:\n%s", err, string(output))
	}

	return &probeOutput, nil
}

// createConcatFile creates a temporary concat file for ffprobe
func (sv *SegmentView) createConcatFile(initURL, segmentURL string) (string, error) {
	// Create temporary file
	tmpFile, err := ioutil.TempFile("", "pantui_concat_*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	// Write concat file format
	// Format: https://ffmpeg.org/ffmpeg-formats.html#concat-1
	concatContent := fmt.Sprintf("file '%s'\nfile '%s'\n", initURL, segmentURL)
	
	_, err = tmpFile.WriteString(concatContent)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write concat file: %v", err)
	}

	return tmpFile.Name(), nil
}

// resolveInitFragmentURL resolves the init fragment URL relative to the segment
func (sv *SegmentView) resolveInitFragmentURL(initURI string) string {
	// If init URI is absolute, return as-is
	if strings.HasPrefix(initURI, "http://") || strings.HasPrefix(initURI, "https://") {
		return initURI
	}

	// If we have a resolved segment URL, resolve init relative to it
	if sv.resolvedURL != "" {
		if baseURL, err := url.Parse(sv.resolvedURL); err == nil {
			if initURL, err := baseURL.Parse(initURI); err == nil {
				return initURL.String()
			}
		}
	}

	// Fallback: return the init URI as-is (might be a local file)
	return initURI
}

// updateContentWithProbeData updates the content with ffprobe analysis
func (sv *SegmentView) updateContentWithProbeData() {
	if sv.probeData == nil {
		sv.showMessage("No probe data available")
		return
	}

	content := fmt.Sprintf(`[yellow]Segment Analysis[white]

[cyan]Original URI:[white]
%s

[cyan]Resolved URL:[white]
%s

%s

[cyan]Available Actions:[white]
• Press [green]c[white] to copy URL to clipboard
• Press [green]o[white] to open in browser
• Press [green]h[white] to show HTTP headers
• Press [green]i[white] to inspect segment
• Press [yellow]Esc[white] to go back`,
		sv.segment.URI,
		sv.resolvedURL,
		sv.formatProbeData())

	sv.textView.SetText(content)
	sv.textView.SetTitle(" Segment Analysis ").SetBorder(true)
}

// formatProbeData formats ffprobe data for display
func (sv *SegmentView) formatProbeData() string {
	if sv.probeData == nil {
		return "[red]No probe data available[white]"
	}

	var content strings.Builder
	
	// Format information
	content.WriteString("[cyan]Format Information:[white]\n")
	format := sv.probeData.Format
	
	content.WriteString(fmt.Sprintf("Container: %s\n", format.FormatLongName))
	if format.Duration != "" {
		if duration, err := strconv.ParseFloat(format.Duration, 64); err == nil {
			content.WriteString(fmt.Sprintf("Duration: %s\n", sv.formatDuration(duration)))
		}
	}
	if format.Size != "" {
		if size, err := strconv.ParseInt(format.Size, 10, 64); err == nil {
			content.WriteString(fmt.Sprintf("File Size: %s\n", sv.formatBytes(size)))
		}
	}
	if format.BitRate != "" {
		if bitrate, err := strconv.ParseInt(format.BitRate, 10, 64); err == nil {
			content.WriteString(fmt.Sprintf("Overall Bitrate: %s\n", sv.formatBitrate(bitrate)))
		}
	}

	// Stream information
	for i, stream := range sv.probeData.Streams {
		content.WriteString(fmt.Sprintf("\n[cyan]Stream %d (%s):[white]\n", i, stream.CodecType))
		
		content.WriteString(fmt.Sprintf("Codec: %s", stream.CodecName))
		if stream.CodecLongName != "" {
			content.WriteString(fmt.Sprintf(" (%s)", stream.CodecLongName))
		}
		content.WriteString("\n")

		if stream.CodecType == "video" {
			if stream.Width > 0 && stream.Height > 0 {
				content.WriteString(fmt.Sprintf("Resolution: %dx%d\n", stream.Width, stream.Height))
			}
			if stream.AvgFrameRate != "" {
				content.WriteString(fmt.Sprintf("Frame Rate: %s fps\n", sv.formatFrameRate(stream.AvgFrameRate)))
			}
			if stream.PixFmt != "" {
				content.WriteString(fmt.Sprintf("Pixel Format: %s\n", stream.PixFmt))
			}
			if stream.Profile != "" {
				content.WriteString(fmt.Sprintf("Profile: %s\n", stream.Profile))
			}
			if stream.Level > 0 {
				content.WriteString(fmt.Sprintf("Level: %d\n", stream.Level))
			}
		} else if stream.CodecType == "audio" {
			if stream.SampleRate != "" {
				content.WriteString(fmt.Sprintf("Sample Rate: %s Hz\n", stream.SampleRate))
			}
			if stream.Channels > 0 {
				content.WriteString(fmt.Sprintf("Channels: %d\n", stream.Channels))
			}
			if stream.ChannelLayout != "" {
				content.WriteString(fmt.Sprintf("Channel Layout: %s\n", stream.ChannelLayout))
			}
			if stream.SampleFmt != "" {
				content.WriteString(fmt.Sprintf("Sample Format: %s\n", stream.SampleFmt))
			}
		}

		if stream.BitRate != "" {
			if bitrate, err := strconv.ParseInt(stream.BitRate, 10, 64); err == nil {
				content.WriteString(fmt.Sprintf("Bitrate: %s\n", sv.formatBitrate(bitrate)))
			}
		}

		if stream.Duration != "" {
			if duration, err := strconv.ParseFloat(stream.Duration, 64); err == nil {
				content.WriteString(fmt.Sprintf("Duration: %s\n", sv.formatDuration(duration)))
			}
		}
	}

	return content.String()
}

// formatDuration formats duration in seconds to human-readable format
func (sv *SegmentView) formatDuration(seconds float64) string {
	duration := time.Duration(seconds * float64(time.Second))
	
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	millisecs := int(duration.Milliseconds()) % 1000
	
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", hours, minutes, secs, millisecs)
	}
	return fmt.Sprintf("%d:%02d.%03d", minutes, secs, millisecs)
}

// formatBytes formats bytes to human-readable format
func (sv *SegmentView) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatBitrate formats bitrate to human-readable format
func (sv *SegmentView) formatBitrate(bitrate int64) string {
	if bitrate >= 1000000 {
		return fmt.Sprintf("%.1f Mbps", float64(bitrate)/1000000)
	} else if bitrate >= 1000 {
		return fmt.Sprintf("%.1f Kbps", float64(bitrate)/1000)
	}
	return fmt.Sprintf("%d bps", bitrate)
}

// formatFrameRate formats frame rate from fraction string
func (sv *SegmentView) formatFrameRate(frameRate string) string {
	// Frame rate is often in format "num/den"
	parts := strings.Split(frameRate, "/")
	if len(parts) == 2 {
		if num, err1 := strconv.ParseFloat(parts[0], 64); err1 == nil {
			if den, err2 := strconv.ParseFloat(parts[1], 64); err2 == nil && den != 0 {
				return fmt.Sprintf("%.2f", num/den)
			}
		}
	}
	return frameRate
}

// showMessage shows a temporary message
func (sv *SegmentView) showMessage(message string) {
	if sv.statusCallback != nil {
		sv.statusCallback(message)
	}
}

// showFFProbeError displays detailed ffprobe error information
func (sv *SegmentView) showFFProbeError(err error) {
	// Show brief message in status bar
	if sv.statusCallback != nil {
		sv.statusCallback("FFProbe analysis failed - see details below")
	}

	// Display full error details in the main content area
	content := fmt.Sprintf(`[yellow]Segment Analysis Failed[white]

[cyan]Original URI:[white]
%s

[cyan]Resolved URL:[white]
%s

[cyan]Segment Details:[white]
%s

[red]FFProbe Error Details:[white]
%s

[cyan]Available Actions:[white]
• Press [green]c[white] to copy URL to clipboard
• Press [green]o[white] to open in browser
• Press [green]h[white] to show HTTP headers
• Press [green]i[white] to retry inspection
• Press [yellow]Esc[white] to go back

[darkgray]Tip: Check if the segment URL is accessible and if ffprobe is installed.[white]`,
		sv.segment.URI,
		sv.resolvedURL,
		sv.formatSegmentDetails(),
		err.Error())

	sv.textView.SetText(content)
	sv.textView.SetTitle(" Segment Analysis - Error ").SetBorder(true)
}
