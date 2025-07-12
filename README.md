# pantui

A terminal user interface (TUI) for exploring HLS (HTTP Live Streaming) manifests, similar to k9s for Kubernetes but designed specifically for HLS content.

## Features

- **Interactive Navigation**: Navigate through master manifests, media manifests, and segments with intuitive keyboard controls
- **Colorized Display**: Syntax highlighting for different HLS manifest elements
- **Multi-level Exploration**: Drill down from master manifests to media manifests to individual segments
- **Navigation Stack**: Easy backtracking through your exploration path with Esc key
- **Detailed Information**: View comprehensive details about variants, segments, and encryption
- **Multiple Input Sources**: Support for both URLs and local files
- **Real-time Updates**: Refresh manifests to see live changes

## Installation

### From Source

```bash
git clone https://github.com/user/pantui.git
cd pantui
go mod tidy
go build -o pantui
```

### Using Go Install

```bash
go install github.com/user/pantui@latest
```

## Usage

### Basic Usage

```bash
# Load HLS manifest from URL
pantui -u https://example.com/master.m3u8

# Load HLS manifest from local file
pantui -f ./path/to/manifest.m3u8
```

### Navigation

- **Arrow Keys (↑↓)**: Navigate through lists
- **Enter**: Drill down into selected item (variant → manifest → segment)
- **Esc**: Go back to previous view or exit
- **F1**: Show help
- **Ctrl+C**: Exit application

### View-Specific Controls

#### Master Manifest View
- **d**: Show detailed variant information
- **r**: Refresh manifest

#### Media Manifest View
- **d**: Show detailed segment information
- **s**: Show manifest summary (duration, encryption, etc.)
- **r**: Refresh manifest

#### Segment View
- **c**: Copy segment URL to clipboard
- **o**: Open segment in browser
- **h**: Show HTTP headers
- **i**: Inspect segment content

## Examples

### Exploring a Master Manifest

1. Start with a master manifest URL:
   ```bash
   pantui -u https://devstreaming-cdn.apple.com/videos/streaming/examples/img_bipbop_adv_example_fmp4/master.m3u8
   ```

2. Navigate through variants using arrow keys
3. Press Enter on a variant to load its media manifest
4. Explore individual segments in the media manifest
5. Use Esc to navigate back through your path

### Local Development

```bash
# Create a test manifest
echo '#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=720x480
low/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2560000,RESOLUTION=1280x720
mid/index.m3u8' > test.m3u8

# Explore it with pantui
pantui -f test.m3u8
```

## Architecture

### Components

- **HLS Parser**: Robust parsing of M3U8 manifests with support for all major HLS tags
- **TUI Framework**: Built on top of tview/tcell for rich terminal interfaces
- **Navigation System**: Stack-based navigation with state preservation
- **Views**: Specialized views for different manifest types and content

### Project Structure

```
pantui/
├── cmd/                    # CLI commands
├── internal/
│   ├── hls/               # HLS parsing logic
│   └── tui/               # TUI components
│       ├── components/    # Reusable UI components
│       └── views/         # Different view types
├── main.go
└── README.md
```

## Dependencies

- [tview](https://github.com/rivo/tview): Rich TUI framework
- [tcell](https://github.com/gdamore/tcell): Terminal handling
- [cobra](https://github.com/spf13/cobra): CLI framework
- [color](https://github.com/fatih/color): Terminal colors

## Development

### Building

```bash
go build -o pantui
```

### Testing

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Similar Projects

- [hlsq](https://github.com/soldiermoth/hlsq): HLS manifest colorizer (inspiration)
- [k9s](https://github.com/derailed/k9s): Kubernetes TUI (navigation inspiration)

## License

MIT License - see LICENSE file for details

## Roadmap

- [ ] Segment content inspection (media analysis)
- [ ] HTTP header inspection
- [ ] Playlist timeline visualization
- [ ] Export functionality (JSON, CSV)
- [ ] Configuration file support
- [ ] Plugin system for custom analyzers
- [ ] Advanced filtering and search
- [ ] Live manifest monitoring
- [ ] Integration with media analysis tools
