package hls

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

// ManifestType represents the type of HLS manifest
type ManifestType string

const (
	MasterManifest ManifestType = "master"
	MediaManifest  ManifestType = "media"
)

// Manifest represents an HLS manifest
type Manifest struct {
	Type        ManifestType  `json:"type"`
	URL         string        `json:"url"`
	Content     string        `json:"content"`
	Lines       []Line        `json:"lines"`
	Variants    []Variant     `json:"variants,omitempty"`
	Segments    []Segment     `json:"segments,omitempty"`
	Tags        []Tag         `json:"tags"`
	BaseURL     string        `json:"base_url"`
	TargetDuration int        `json:"target_duration,omitempty"`
	Version     int           `json:"version"`
	Sequence    int           `json:"sequence,omitempty"`
}

// Line represents a line in the manifest
type Line struct {
	Number  int    `json:"number"`
	Content string `json:"content"`
	Type    string `json:"type"` // "tag", "uri", "comment", "empty"
}

// Variant represents a variant stream in a master manifest
type Variant struct {
	URI        string            `json:"uri"`
	Bandwidth  int               `json:"bandwidth"`
	Resolution string            `json:"resolution,omitempty"`
	Codecs     string            `json:"codecs,omitempty"`
	Attributes map[string]string `json:"attributes"`
}

// Segment represents a media segment in a media manifest
type Segment struct {
	URI      string  `json:"uri"`
	Duration float64 `json:"duration"`
	Sequence int     `json:"sequence"`
	ByteRange string `json:"byte_range,omitempty"`
	Key      *Key    `json:"key,omitempty"`
	Map      *Map    `json:"map,omitempty"`
}

// Map represents initialization segment information from EXT-X-MAP
type Map struct {
	URI       string `json:"uri"`
	ByteRange string `json:"byte_range,omitempty"`
}

// Key represents encryption key information
type Key struct {
	Method string `json:"method"`
	URI    string `json:"uri,omitempty"`
	IV     string `json:"iv,omitempty"`
	KeyFormat string `json:"key_format,omitempty"`
}

// Tag represents an HLS tag
type Tag struct {
	Name       string            `json:"name"`
	Value      string            `json:"value,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	LineNumber int               `json:"line_number"`
}

// Parser handles HLS manifest parsing
type Parser struct {
	baseURL string
}

// NewParser creates a new HLS parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFromURL parses an HLS manifest from a URL
func (p *Parser) ParseFromURL(manifestURL string) (*Manifest, error) {
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	p.baseURL = p.getBaseURL(manifestURL)
	return p.parseContent(string(content), manifestURL)
}

// ParseFromFile parses an HLS manifest from a local file
func (p *Parser) ParseFromFile(filePath string) (*Manifest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	p.baseURL = path.Dir(filePath)
	return p.parseContent(string(content), filePath)
}

// parseContent parses the manifest content
func (p *Parser) parseContent(content, sourceURL string) (*Manifest, error) {
	manifest := &Manifest{
		URL:     sourceURL,
		Content: content,
		BaseURL: p.baseURL,
		Tags:    make([]Tag, 0),
	}

	lines := strings.Split(content, "\n")
	manifest.Lines = make([]Line, len(lines))

	// Parse lines
	for i, line := range lines {
		line = strings.TrimSpace(line)
		manifest.Lines[i] = Line{
			Number:  i + 1,
			Content: line,
			Type:    p.getLineType(line),
		}
	}

	// Determine manifest type and parse accordingly
	if p.isMasterManifest(content) {
		manifest.Type = MasterManifest
		if err := p.parseMasterManifest(manifest); err != nil {
			return nil, err
		}
	} else {
		manifest.Type = MediaManifest
		if err := p.parseMediaManifest(manifest); err != nil {
			return nil, err
		}
	}

	return manifest, nil
}

// getLineType determines the type of a line
func (p *Parser) getLineType(line string) string {
	if line == "" {
		return "empty"
	}
	if strings.HasPrefix(line, "#EXT") {
		return "tag"
	}
	if strings.HasPrefix(line, "#") {
		return "comment"
	}
	return "uri"
}

// isMasterManifest checks if the manifest is a master manifest
func (p *Parser) isMasterManifest(content string) bool {
	return strings.Contains(content, "#EXT-X-STREAM-INF") || strings.Contains(content, "#EXT-X-I-FRAME-STREAM-INF")
}

// parseMasterManifest parses a master manifest
func (p *Parser) parseMasterManifest(manifest *Manifest) error {
	scanner := bufio.NewScanner(strings.NewReader(manifest.Content))
	lineNumber := 0
	variants := make([]Variant, 0)
	
	var currentVariant *Variant
	
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#EXT") {
			continue
		}
		
		if strings.HasPrefix(line, "#EXT-X-VERSION:") {
			if version, err := strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-VERSION:")); err == nil {
				manifest.Version = version
			}
		} else if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			attributes := p.parseAttributes(strings.TrimPrefix(line, "#EXT-X-STREAM-INF:"))
			currentVariant = &Variant{
				Attributes: attributes,
			}
			
			if bandwidth, ok := attributes["BANDWIDTH"]; ok {
				if bw, err := strconv.Atoi(bandwidth); err == nil {
					currentVariant.Bandwidth = bw
				}
			}
			
			if resolution, ok := attributes["RESOLUTION"]; ok {
				currentVariant.Resolution = resolution
			}
			
			if codecs, ok := attributes["CODECS"]; ok {
				currentVariant.Codecs = codecs
			}
		} else if currentVariant != nil && !strings.HasPrefix(line, "#") {
			currentVariant.URI = line
			variants = append(variants, *currentVariant)
			currentVariant = nil
		}
		
		// Parse all tags
		if strings.HasPrefix(line, "#EXT") {
			tag := p.parseTag(line, lineNumber)
			manifest.Tags = append(manifest.Tags, tag)
		}
	}
	
	manifest.Variants = variants
	return nil
}

// parseMediaManifest parses a media manifest
func (p *Parser) parseMediaManifest(manifest *Manifest) error {
	scanner := bufio.NewScanner(strings.NewReader(manifest.Content))
	lineNumber := 0
	segments := make([]Segment, 0)
	sequence := 0
	
	var currentSegment *Segment
	var currentKey *Key
	var currentMap *Map
	
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#EXT") {
			continue
		}
		
		if strings.HasPrefix(line, "#EXT-X-VERSION:") {
			if version, err := strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-VERSION:")); err == nil {
				manifest.Version = version
			}
		} else if strings.HasPrefix(line, "#EXT-X-TARGETDURATION:") {
			if duration, err := strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-TARGETDURATION:")); err == nil {
				manifest.TargetDuration = duration
			}
		} else if strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:") {
			if seq, err := strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-MEDIA-SEQUENCE:")); err == nil {
				manifest.Sequence = seq
				sequence = seq
			}
		} else if strings.HasPrefix(line, "#EXTINF:") {
			durationStr := strings.TrimPrefix(line, "#EXTINF:")
			if commaIdx := strings.Index(durationStr, ","); commaIdx != -1 {
				durationStr = durationStr[:commaIdx]
			}
			
			if duration, err := strconv.ParseFloat(durationStr, 64); err == nil {
				currentSegment = &Segment{
					Duration: duration,
					Sequence: sequence,
					Key:      currentKey,
					Map:      currentMap,
				}
				sequence++
			}
		} else if strings.HasPrefix(line, "#EXT-X-BYTERANGE:") {
			// Handle byte range for current segment
			if currentSegment != nil {
				currentSegment.ByteRange = strings.TrimPrefix(line, "#EXT-X-BYTERANGE:")
			}
		} else if strings.HasPrefix(line, "#EXT-X-KEY:") {
			attributes := p.parseAttributes(strings.TrimPrefix(line, "#EXT-X-KEY:"))
			currentKey = &Key{
				Method: attributes["METHOD"],
				URI:    attributes["URI"],
				IV:     attributes["IV"],
				KeyFormat: attributes["KEYFORMAT"],
			}
		} else if strings.HasPrefix(line, "#EXT-X-MAP:") {
			attributes := p.parseAttributes(strings.TrimPrefix(line, "#EXT-X-MAP:"))
			currentMap = &Map{
				URI:       attributes["URI"],
				ByteRange: attributes["BYTERANGE"],
			}
		} else if currentSegment != nil && !strings.HasPrefix(line, "#") {
			currentSegment.URI = line
			segments = append(segments, *currentSegment)
			currentSegment = nil
		}
		
		// Parse all tags
		if strings.HasPrefix(line, "#EXT") {
			tag := p.parseTag(line, lineNumber)
			manifest.Tags = append(manifest.Tags, tag)
		}
	}
	
	manifest.Segments = segments
	return nil
}

// parseTag parses an HLS tag
func (p *Parser) parseTag(line string, lineNumber int) Tag {
	tag := Tag{
		LineNumber: lineNumber,
	}
	
	if colonIdx := strings.Index(line, ":"); colonIdx != -1 {
		tag.Name = line[:colonIdx]
		tag.Value = line[colonIdx+1:]
		tag.Attributes = p.parseAttributes(tag.Value)
	} else {
		tag.Name = line
	}
	
	return tag
}

// parseAttributes parses HLS tag attributes
func (p *Parser) parseAttributes(attrStr string) map[string]string {
	attributes := make(map[string]string)
	
	// Handle quoted values
	parts := p.splitAttributes(attrStr)
	
	for _, part := range parts {
		if equalIdx := strings.Index(part, "="); equalIdx != -1 {
			key := strings.TrimSpace(part[:equalIdx])
			value := strings.TrimSpace(part[equalIdx+1:])
			
			// Remove quotes if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			
			attributes[key] = value
		}
	}
	
	return attributes
}

// splitAttributes splits attribute string handling quoted values
func (p *Parser) splitAttributes(attrStr string) []string {
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
					parts = append(parts, current.String())
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
		parts = append(parts, current.String())
	}
	
	return parts
}

// getBaseURL extracts the base URL from a full URL
func (p *Parser) getBaseURL(fullURL string) string {
	if parsedURL, err := url.Parse(fullURL); err == nil {
		parsedURL.Path = path.Dir(parsedURL.Path)
		return parsedURL.String()
	}
	return ""
}

// ResolveURL resolves a relative URL against the base URL
func (p *Parser) ResolveURL(relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL
	}
	
	if strings.HasPrefix(p.baseURL, "http://") || strings.HasPrefix(p.baseURL, "https://") {
		if baseURL, err := url.Parse(p.baseURL); err == nil {
			if resolvedURL, err := baseURL.Parse(relativeURL); err == nil {
				return resolvedURL.String()
			}
		}
	}
	
	return path.Join(p.baseURL, relativeURL)
}
