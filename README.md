# GoTorrent

A fast, lightweight BitTorrent desktop client built with Go and the [Fyne v2](https://fyne.io) UI framework.

<!-- Screenshot placeholder -->
> _Screenshot coming soon_

---

## Features

- **Native desktop UI** — dark/light/system theme, resizable window, system tray
- **Drag and drop** — drop `.torrent` files anywhere on the window
- **Torrent preview** — inspect files, size, tracker, and infohash before downloading
- **File selection** — choose which files to download in multi-file torrents
- **Progress tracking** — live speed, peer count, and ETA per torrent
- **Pause / Resume / Remove** — full lifecycle management
- **Single instance** — subsequent launches forward `.torrent` files to the running app
- **Persistent state** — torrent list survives restarts
- **OS notifications** — desktop alert on completion or error
- **File association** — open `.torrent` files directly from your file manager
- **Cross-platform** — macOS, Windows, Linux

---

## Prerequisites

| Requirement | Version |
|-------------|---------|
| Go | 1.22+ |
| Fyne CLI (for packaging) | `go install fyne.io/fyne/v2/cmd/fyne@latest` |
| C compiler | gcc / clang (required by Fyne/OpenGL) |
| **macOS** | Xcode command-line tools |
| **Linux** | `libgl1-mesa-dev xorg-dev` (apt) or equivalent |
| **Windows** | [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MSYS2 |

---

## Build

```bash
# 1. Clone
git clone https://github.com/tarunvishwakarma1/gotorrent
cd gotorrent

# 2. Download dependencies
make deps

# 3. Generate PNG icon from SVG (requires librsvg or Inkscape)
make icon

# 4. Run in development mode
make run

# 5. Build stripped binary
make build
```

---

## Package

### macOS (.app bundle)

```bash
make package-mac
# Produces: GoTorrent.app and gotorrent-1.0.0-mac.tar.gz
```

### Linux

```bash
make package-linux
# Produces: gotorrent binary and gotorrent-1.0.0-linux.tar.gz

# Install system-wide + register MIME type
make install-linux
```

### Windows

```bash
# Cross-compile from Linux/macOS with GOOS=windows
make package-windows
# Produces: GoTorrent.exe and gotorrent-1.0.0-windows.zip
```

---

## Usage

```
# Run directly
./gotorrent

# Open a specific .torrent file
./gotorrent /path/to/file.torrent
```

**Keyboard shortcuts:**

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl+O` | Open file picker |
| `Cmd/Ctrl+,` | Settings |
| `Cmd/Ctrl+Q` | Quit |
| `Space` | Pause / Resume selected |
| `Delete` | Remove selected |

---

## Set as default .torrent handler

### macOS

```bash
# After installing the .app bundle:
duti -s com.gotorrent.app org.bittorrent.torrent all
```

### Linux

```bash
xdg-mime default gotorrent.desktop application/x-bittorrent
```

### Windows

Right-click any `.torrent` file → Open With → Choose another app → gotorrent.

---

## Architecture

```
gotorrent/
├── parser/          # Bencode parser/encoder
├── torrent/         # .torrent file parser (single + multi-file)
├── tracker/         # HTTP tracker client
├── peers/           # Compact peer list decoder
├── handshake/       # BitTorrent handshake protocol
├── messages/        # BitTorrent wire protocol messages
├── bitfield/        # Piece availability bitfield
├── client/          # Per-peer TCP client
├── p2p/             # Download coordinator (piece work queue)
├── cmd/gotorrent/    # Application entry point
└── internal/
    ├── engine/      # Torrent manager, state, download pipeline
    ├── ui/          # Fyne UI: windows, screens, widgets, tray
    ├── config/      # Settings persistence
    └── ipc/         # Single-instance IPC via TCP socket
```

---

## Data locations

| Platform | Config | State | Log |
|----------|--------|-------|-----|
| macOS | `~/Library/Application Support/gotorrent/config.json` | `…/torrents.json` | `…/gotorrent.log` |
| Linux | `~/.local/share/gotorrent/config.json` | `…/torrents.json` | `…/gotorrent.log` |
| Windows | `%APPDATA%\gotorrent\config.json` | `…\torrents.json` | `…\gotorrent.log` |

---

## Credits

Built by **Tarun Vishwakarma**.

BitTorrent engine written from scratch in pure Go.

UI powered by [Fyne](https://fyne.io) — cross-platform GUI framework for Go.

---

## License

MIT — see [LICENSE](LICENSE).
