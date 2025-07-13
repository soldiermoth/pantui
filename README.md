# PanTUI - HLS Manifest Navigator

A powerful Terminal User Interface (TUI) for exploring and analyzing HLS (HTTP Live Streaming) manifests. Navigate through master manifests, media playlists, and individual segments with detailed technical analysis powered by FFProbe integration.

## 🚀 Features

### 🔍 **Interactive Navigation**
- **Master Manifests** → Navigate variant streams, audio groups, subtitle tracks, I-frame streams
- **Media Manifests** → Browse individual segments with complete metadata  
- **Segment Analysis** → Deep technical inspection with FFProbe integration
- **Smart Scrolling** → Only scrolls when navigation targets are off-screen
- **Navigation Stack** → Easy backtracking with Esc key

### 🎨 **Rich Visual Experience**
- **Syntax Highlighting** → Colorized HLS manifest display (hlsq-style)
- **Manifest Structure** → Preserves original HLS format and hierarchy
- **Visual Selection** → Clear highlighting of navigable elements with `>` indicator
- **Status Information** → Real-time feedback and progress updates
- **Key Bindings Display** → Always-visible available actions

### 🔧 **Advanced Analysis**
- **FFProbe Integration** → Complete media segment technical analysis
- **Init Fragment Support** → Automatic handling of `#EXT-X-MAP` segments for fMP4
- **Network Protocol Support** → HTTP/HTTPS manifest and segment loading
- **Comprehensive Metadata** → Codec info, resolution, bitrate, duration, encryption status
- **Error Diagnostics** → Detailed FFProbe error reporting with full command output

### ⚡ **Media Operations**
- **FFPlay Integration** → Direct manifest playback with 'p' key
- **Segment Inspection** → Detailed codec, resolution, and bitrate information
- **HTTP Headers** → View complete HTTP response headers
- **URL Management** → Copy URLs to clipboard, open in browser

## 📦 Installation

### Prerequisites
- **Go 1.19+** - For building the application
- **FFmpeg** - For media analysis and playback functionality
  ```bash
  # macOS
  brew install ffmpeg
  
  # Ubuntu/Debian
  sudo apt install ffmpeg
  
  # Windows
  # Download from https://ffmpeg.org/download.html
  ```

### Build from Source
```bash
git clone https://github.com/soldiermoth/pantui.git
cd pantui
go mod tidy
go build -o pantui .
```

### Using Go Install
```bash
go install github.com/soldiermoth/pantui@latest
```

## 🚀 Usage

### Basic Commands
```bash
# Analyze HLS manifest from URL
./pantui https://example.com/master.m3u8

# Analyze local manifest file  
./pantui /path/to/manifest.m3u8

# Show help
./pantui -h
```

### 🎮 Navigation Keys

| Key | Action | Context |
|-----|--------|---------|
| `↑↓` | Navigate between URIs | All views |
| `Enter` | Open selected item | All views |
| `Esc` | Go back / Exit | All views |
| `F1` | Show help | All views |
| `Ctrl+C` | Exit application | All views |

### 🎯 View-Specific Controls

#### Master Manifest View
| Key | Action |
|-----|--------|
| `p` | Play manifest with ffplay |
| `d` | Show variant details |
| `r` | Refresh manifest |

#### Media Manifest View  
| Key | Action |
|-----|--------|
| `p` | Play manifest with ffplay |
| `d` | Show segment details |
| `s` | Show manifest summary |
| `r` | Refresh manifest |

#### Segment View
| Key | Action |
|-----|--------|
| `i` | **Inspect with FFProbe** |
| `c` | Copy URL to clipboard |
| `o` | Open in browser |
| `h` | Show HTTP headers |

## 🎬 Examples

### Analyzing Apple's Sample HLS Stream
```bash
./pantui "https://devstreaming-cdn.apple.com/videos/streaming/examples/img_bipbop_adv_example_fmp4/master.m3u8"
```

### Complete Workflow
1. **Master Manifest** → Navigate through variants, audio groups, subtitles
2. **Media Manifest** → Browse segments, view encryption status
3. **Segment Analysis** → Press `i` for detailed FFProbe analysis
4. **Playback Testing** → Press `p` to test with ffplay

### Modern fMP4 with Init Fragments
When you encounter `#EXT-X-MAP` tags, PanTUI automatically:
1. ✅ Detects the init fragment URI
2. ✅ Creates temporary concat files for FFProbe
3. ✅ Provides complete technical analysis
4. ✅ Shows codec, resolution, bitrate details

### Sample Output
```
┌ Master Manifest - 4 variants ──────────────────────────────────┐
│  #EXTM3U                                                        │
│  #EXT-X-VERSION:6                                              │
│> #EXT-X-STREAM-INF:BANDWIDTH=2177116,RESOLUTION=960x540,...    │
│  variant1.m3u8                                                 │
│  #EXT-X-STREAM-INF:BANDWIDTH=6312875,RESOLUTION=1920x1080,... │
│  variant2.m3u8                                                 │
│  #EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",...  │
│  audio_en.m3u8                                                 │
└─────────────────────────────────────────────────────────────────┘
│ Enter=Open Variant  ↑↓=Navigate  p=Play  d=Details  r=Refresh  │
```

### FFProbe Analysis Example
```
[yellow]Segment Analysis[white]

[cyan]Format Information:[white]
Container: QuickTime / MOV
Duration: 6.006 seconds  
File Size: 2.1 MB
Overall Bitrate: 2.8 Mbps

[cyan]Stream 0 (video):[white]
Codec: h264 (High Profile)
Resolution: 1920x1080
Frame Rate: 29.97 fps
Pixel Format: yuv420p
Bitrate: 2.5 Mbps

[cyan]Stream 1 (audio):[white]  
Codec: aac (LC)
Sample Rate: 48000 Hz
Channels: 2 (stereo)
Bitrate: 128 Kbps
```

## 🏗️ Technical Details

### HLS Support
- **Master Manifests** - Variant streams, audio groups, subtitle tracks, I-frame streams
- **Media Manifests** - Segments, byte ranges, encryption keys, target duration  
- **Modern Features** - fMP4 segments, init fragments (`#EXT-X-MAP`), HTTPS delivery
- **EXT Tags** - Complete support for HLS specification tags

### FFProbe Integration
When inspecting segments, PanTUI provides:
- **Container Information** - Format, duration, file size, overall bitrate
- **Video Streams** - Codec (H.264/H.265/AV1), resolution, frame rate, pixel format
- **Audio Streams** - Codec (AAC/MP3), sample rate, channels, channel layout
- **Advanced Metadata** - Color space, encoding profiles, levels

### Init Fragment Handling
For segments with init fragments:
```bash
# PanTUI automatically creates:
file 'https://cdn.example.com/init.m4s'
file 'https://cdn.example.com/segment.m4s'

# And analyzes with:
ffprobe -f concat -safe 0 -protocol_whitelist concat,file,http,https,tcp,tls /tmp/concat.txt
```

### Architecture

```
pantui/
├── cmd/                    # Command-line interface
├── internal/
│   ├── hls/               # HLS manifest parsing & data structures
│   └── tui/               # Terminal UI components
│       ├── views/         # Master, Media, Segment views
│       ├── components/    # Status bar, key bindings  
│       └── colors/        # Syntax highlighting
├── main.go                # Application entry point
└── README.md              # This file
```

### Dependencies
- **[tview](https://github.com/rivo/tview)** - Rich TUI framework
- **[tcell](https://github.com/gdamore/tcell)** - Terminal handling
- **[cobra](https://github.com/spf13/cobra)** - CLI framework

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

## 🗺️ Roadmap

### ✅ Completed
- [x] Segment content inspection (FFProbe integration)
- [x] HTTP header inspection  
- [x] Media analysis tools integration (FFmpeg)
- [x] Init fragment support for fMP4
- [x] Comprehensive error reporting

### 🚧 Planned Features
- [ ] Playlist timeline visualization
- [ ] Export functionality (JSON, CSV)
- [ ] Configuration file support
- [ ] Plugin system for custom analyzers
- [ ] Advanced filtering and search
- [ ] Live manifest monitoring
- [ ] Bandwidth utilization analysis
- [ ] Segment download performance metrics
