# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

pantui is a Terminal User Interface (TUI) application for navigating HLS (HTTP Live Streaming) manifests. The application allows users to:

- Start with a master manifest (.m3u8)
- Navigate into details about media manifests 
- Inspect individual segments
- Maintain the visual structure of HLS manifests with colorized output (similar to hlsq)

## Commands

### Building and Running
- `go build -o pantui .` - Build the application
- `go run . -u <URL>` - Run with HLS manifest URL
- `go run . -f <file>` - Run with local HLS manifest file
- `go test ./...` - Run all tests
- `go test ./internal/hls -v` - Run HLS parser tests specifically

### Development Commands
- `go mod tidy` - Clean up dependencies
- `go mod download` - Download dependencies

## Architecture

### Core Components
- `internal/hls/` - HLS manifest parsing and data structures
- `internal/tui/` - Terminal user interface components
- `internal/tui/views/` - Different view types (master, media, segment)
- `internal/tui/components/` - Reusable UI components (status bar, key bar)
- `internal/tui/colors/` - Color scheme definitions for hlsq-like appearance

### Key Files
- `main.go` - Application entry point
- `cmd/root.go` - Cobra command-line interface setup
- `internal/tui/app.go` - Main TUI application controller
- `internal/hls/parser.go` - HLS manifest parser implementation

## Key Concepts

- **Master Manifest**: Top-level playlist containing references to media playlists, audio groups, subtitle tracks, and I-frame streams
- **Media Manifest**: Playlist containing actual media segments with byte-range support
- **Segments**: Individual media files referenced in media playlests
- **Audio Groups**: Alternative audio tracks (different languages, codecs, channel counts)
- **Subtitle Tracks**: Text overlay streams
- **I-Frame Streams**: Keyframe-only streams for seeking
- **Navigation**: Users can drill down from master → media → segments using Enter key
- **Color Scheme**: Uses tcell colors to mimic hlsq appearance
- **Smart Scrolling**: Only scrolls when the next URI is outside the viewport

## Navigation
- Enter: Open selected item (variant/segment/audio/subtitle)
- Esc: Go back to previous view or quit
- F1: Show help
- Ctrl+C: Exit application
- Up/Down: Navigate between URIs (smart scrolling - only scrolls when needed)
- p: Play current manifest with ffplay
- d: Show details for selected item
- r: Refresh current manifest
- c: Copy URL to clipboard (in segment view)
- o: Open URL in browser (in segment view)
- h: Show HTTP headers (in segment view)
- i: Inspect segment with ffprobe (in segment view)

## Development Notes

- The UI preserves the structure and format of HLS manifests
- Colorization follows patterns similar to hlsq (github.com/soldiermoth/hlsq)
- Support for both local files and HTTP URLs for manifest loading
- Interactive navigation between different levels of the HLS hierarchy
- Uses tview and tcell for terminal UI components
- Play functionality requires `ffplay` (part of FFmpeg) to be installed and available in PATH
- Segment inspection requires `ffprobe` (part of FFmpeg) to be installed and available in PATH

## Requirements

- Go 1.19 or later
- FFmpeg (ffplay and ffprobe) for media playback and analysis functionality

## FFProbe Integration

The segment inspection feature uses `ffprobe` to analyze media segments and display:
- Container format information
- Video stream details (codec, resolution, frame rate, bit rate)
- Audio stream details (codec, sample rate, channels, bit rate)
- File size and duration
- Color space and pixel format information
- Encoding profiles and levels