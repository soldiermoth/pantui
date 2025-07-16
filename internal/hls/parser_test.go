package hls

import (
	"strings"
	"testing"
)

func TestParseMasterManifest(t *testing.T) {
	masterManifest := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=720x480,CODECS="avc1.77.30,mp4a.40.2"
low/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2560000,RESOLUTION=1280x720,CODECS="avc1.77.30,mp4a.40.2"
mid/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=7680000,RESOLUTION=1920x1080,CODECS="avc1.640028,mp4a.40.2"
high/index.m3u8`

	parser := NewParser()
	manifest, err := parser.parseContent(masterManifest, "test.m3u8")

	if err != nil {
		t.Fatalf("Failed to parse master manifest: %v", err)
	}

	if manifest.Type != MasterManifest {
		t.Errorf("Expected manifest type to be %s, got %s", MasterManifest, manifest.Type)
	}

	if manifest.Version != 3 {
		t.Errorf("Expected version 3, got %d", manifest.Version)
	}

	if len(manifest.Variants) != 3 {
		t.Errorf("Expected 3 variants, got %d", len(manifest.Variants))
	}

	// Test first variant
	variant := manifest.Variants[0]
	if variant.URI != "low/index.m3u8" {
		t.Errorf("Expected URI 'low/index.m3u8', got '%s'", variant.URI)
	}
	if variant.Bandwidth != 1280000 {
		t.Errorf("Expected bandwidth 1280000, got %d", variant.Bandwidth)
	}
	if variant.Resolution != "720x480" {
		t.Errorf("Expected resolution '720x480', got '%s'", variant.Resolution)
	}
	if variant.Codecs != "avc1.77.30,mp4a.40.2" {
		t.Errorf("Expected codecs 'avc1.77.30,mp4a.40.2', got '%s'", variant.Codecs)
	}
}

func TestParseMediaManifest(t *testing.T) {
	mediaManifest := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:9.009,
segment0.ts
#EXTINF:9.009,
segment1.ts
#EXTINF:9.009,
segment2.ts
#EXT-X-ENDLIST`

	parser := NewParser()
	manifest, err := parser.parseContent(mediaManifest, "test.m3u8")

	if err != nil {
		t.Fatalf("Failed to parse media manifest: %v", err)
	}

	if manifest.Type != MediaManifest {
		t.Errorf("Expected manifest type to be %s, got %s", MediaManifest, manifest.Type)
	}

	if manifest.Version != 3 {
		t.Errorf("Expected version 3, got %d", manifest.Version)
	}

	if manifest.TargetDuration != 10 {
		t.Errorf("Expected target duration 10, got %d", manifest.TargetDuration)
	}

	if manifest.Sequence != 0 {
		t.Errorf("Expected media sequence 0, got %d", manifest.Sequence)
	}

	if len(manifest.Segments) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(manifest.Segments))
	}

	// Test first segment
	segment := manifest.Segments[0]
	if segment.URI != "segment0.ts" {
		t.Errorf("Expected URI 'segment0.ts', got '%s'", segment.URI)
	}
	if segment.Duration != 9.009 {
		t.Errorf("Expected duration 9.009, got %f", segment.Duration)
	}
	if segment.Sequence != 0 {
		t.Errorf("Expected sequence 0, got %d", segment.Sequence)
	}
}

func TestParseAttributes(t *testing.T) {
	parser := NewParser()

	// Test simple attributes
	attrs := parser.parseAttributes("BANDWIDTH=1280000,RESOLUTION=720x480")
	if attrs["BANDWIDTH"] != "1280000" {
		t.Errorf("Expected BANDWIDTH '1280000', got '%s'", attrs["BANDWIDTH"])
	}
	if attrs["RESOLUTION"] != "720x480" {
		t.Errorf("Expected RESOLUTION '720x480', got '%s'", attrs["RESOLUTION"])
	}

	// Test quoted attributes
	attrs = parser.parseAttributes(`CODECS="avc1.77.30,mp4a.40.2"`)
	if attrs["CODECS"] != "avc1.77.30,mp4a.40.2" {
		t.Errorf("Expected CODECS 'avc1.77.30,mp4a.40.2', got '%s'", attrs["CODECS"])
	}
}

func TestSplitAttributes(t *testing.T) {
	parser := NewParser()

	// Test simple case
	parts := parser.splitAttributes("A=1,B=2,C=3")
	expected := []string{"A=1", "B=2", "C=3"}
	if len(parts) != len(expected) {
		t.Errorf("Expected %d parts, got %d", len(expected), len(parts))
	}
	for i, part := range parts {
		if part != expected[i] {
			t.Errorf("Expected part '%s', got '%s'", expected[i], part)
		}
	}

	// Test with quoted values containing commas
	parts = parser.splitAttributes(`A=1,B="hello,world",C=3`)
	expected = []string{"A=1", `B="hello,world"`, "C=3"}
	if len(parts) != len(expected) {
		t.Errorf("Expected %d parts, got %d", len(expected), len(parts))
	}
	for i, part := range parts {
		if part != expected[i] {
			t.Errorf("Expected part '%s', got '%s'", expected[i], part)
		}
	}
}

func TestIsMasterManifest(t *testing.T) {
	parser := NewParser()

	// Master manifest
	masterContent := "#EXT-X-STREAM-INF:BANDWIDTH=1000000\ntest.m3u8"
	if !parser.isMasterManifest(masterContent) {
		t.Error("Expected master manifest to be detected")
	}

	// Media manifest
	mediaContent := "#EXTINF:10.0\nsegment.ts"
	if parser.isMasterManifest(mediaContent) {
		t.Error("Expected media manifest to not be detected as master")
	}
}

func TestGetLineType(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		line     string
		expected string
	}{
		{"", "empty"},
		{"#EXTM3U", "tag"},
		{"#EXT-X-VERSION:3", "tag"},
		{"# This is a comment", "comment"},
		{"segment.ts", "uri"},
		{"http://example.com/segment.ts", "uri"},
	}

	for _, test := range tests {
		result := parser.getLineType(test.line)
		if result != test.expected {
			t.Errorf("For line '%s', expected type '%s', got '%s'", test.line, test.expected, result)
		}
	}
}

func TestResolveURL(t *testing.T) {
	parser := NewParser()

	// Test with HTTP base URL
	parser.baseURL = "https://example.com/video"
	
	// Absolute URL should remain unchanged
	result := parser.ResolveURL("https://other.com/segment.ts")
	if result != "https://other.com/segment.ts" {
		t.Errorf("Expected absolute URL to remain unchanged, got '%s'", result)
	}

	// Relative URL should be resolved
	result = parser.ResolveURL("segment.ts")
	if !strings.Contains(result, "segment.ts") {
		t.Errorf("Expected resolved URL to contain 'segment.ts', got '%s'", result)
	}

	// Test with file path base URL
	parser.baseURL = "/local/path"
	result = parser.ResolveURL("segment.ts")
	expected = "/local/path/segment.ts"
	if result != expected {
		t.Errorf("Expected resolved path '%s', got '%s'", expected, result)
	}
}
