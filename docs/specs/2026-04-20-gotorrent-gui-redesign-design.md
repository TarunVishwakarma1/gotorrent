# GoTorrent GUI Redesign — Design Spec

**Date:** 2026-04-20  
**Status:** Approved  
**Approach:** Option B — Full Visual Overhaul

---

## Overview

Full visual redesign of the GoTorrent Fyne v2 desktop UI. No engine or logic changes. All work confined to `internal/ui/` and `internal/ui/widgets/`. Delivers a Dark Neon Compact aesthetic with blue/indigo accent palette, icon-only sidebar, glow progress bars, and rich 4-row torrent cards.

---

## Color System

All colors defined in `internal/ui/app.go` (`goTorrentTheme.darkColor`).

| Role | Hex | Usage |
|---|---|---|
| Background | `#080d1f` | `ColorNameBackground` |
| Card / sidebar bg | `#0e1530` | `ColorNameButton`, `ColorNameHeaderBackground`, `ColorNameInputBackground`, `ColorNameMenuBackground` |
| Accent (primary) | `#4d9fff` | `ColorNamePrimary`, `ColorNameFocus`, `ColorNameSelection` |
| Foreground | `#e8eeff` | `ColorNameForeground` |
| Muted / placeholder | `#ffffff44` | `ColorNameDisabled`, `ColorNamePlaceHolder` |
| Success / complete | `#00e676` | `ColorNameSuccess` |
| Warning | `#ffcb6b` | `ColorNameWarning` |
| Error | `#ff5370` | `ColorNameError` |
| Shadow | `rgba(0,0,0,0.6)` | `ColorNameShadow` |

Light theme colors remain unchanged (light theme is secondary, dark is primary target).

Progress bar gradient: `#4d9fff` → `#a78bfa` (implemented via `canvas.LinearGradient` in `ProgressRow`, not via theme).

---

## Sidebar (`internal/ui/window.go`)

### Layout
- Fixed width: split offset `0.07` (~56px at 1200px window width)
- Background: `canvas.Rectangle` filled with `#0e1530`
- Right border: 1px `#ffffff0a`

### Structure (top to bottom)
1. **Logo** — 32×32 `canvas.Rectangle` with rounded corners (radius 8), gradient fill `#4d9fff→#a78bfa`, centered download icon `canvas.Text`, glow shadow. Fixed at top with 14px padding.
2. **Nav items** — 4 icon-only `widget.Button` (40×40), `widget.MediumImportance`. Each wrapped in a fixed-size container.
3. **Active indicator** — 3px `canvas.Rectangle` strip on left edge of active nav item, color `#4d9fff`, glow shadow. Implemented by swapping a custom `navItem` wrapper that shows/hides the strip.
4. **Spacer** — fills remaining vertical space.
5. **Status dot** — 8px `canvas.Circle` at bottom center. Green `#00e676` with glow when any download is active; grey `#ffffff33` when idle. Updated by the same ticker that drives the status bar.

### Nav items
| Icon | Screen |
|---|---|
| `theme.DownloadIcon()` | Downloads |
| `theme.ContentAddIcon()` | Add Torrent |
| `theme.SettingsIcon()` | Settings |
| `theme.InfoIcon()` | About |

Buttons use `widget.LowImportance`; active button uses `widget.HighImportance` (existing `selectNav` logic unchanged).

---

## ProgressRow Widget (`internal/ui/widgets/progressrow.go`)

### Card structure (full 4-row)

```
┌── 4px color strip ──────────────────────────────────────────────┐
│  Row 1: [icon + name (bold)]          [status badge pill]       │
│  Row 2: [━━━━━━━━━━━━━━━░░░░] 65%  3.2 GB                      │
│  Row 3: ↓ 4.2 MB/s  ·  3 peers  ·  ETA 3m  ·  2.08 / 3.2 GB   │
│  Row 4: #a1b2c3d4…              [Pause] [Open] [✕ Remove]       │
└─────────────────────────────────────────────────────────────────┘
```

### Progress bar
- Height: 9px, corner radius 5px
- Track: `canvas.Rectangle`, fill `#ffffff10`
- Fill: `canvas.LinearGradient` (Fyne v2 native), start `#4d9fff`, end `#a78bfa`, horizontal. For complete state: solid `canvas.Rectangle` fill `#00e676`.
- Glow: `canvas.Rectangle` behind fill bar, same color at 40% alpha, 2px larger on each side — approximates glow since Fyne has no box-shadow. Colors: `#4d9fff66` downloading, `#00e67666` complete, none for paused/error.

### Status strip colors
| Status | Color |
|---|---|
| Downloading / Connecting | `#4d9fff` |
| Complete | `#00e676` |
| Error | `#ff5370` |
| Verifying | `#ffcb6b` |
| Paused / Queued | `#ffffff33` |

### Status badge
- Pill shape: `canvas.Rectangle` corner radius 10, filled with status color
- Text: `canvas.Text` size 9, bold, black (`#000000`) for bright colors, white for dark/muted
- Min size: 80×20px

### Action buttons (footer row)
| Button | Style | Visibility |
|---|---|---|
| Pause / Resume / Retry | `background #4d9fff22`, text `#4d9fff` | Hidden when Complete |
| Open folder | `background #ffffff0a`, text `#ffffff66` | Always visible |
| Remove | `background #ff537018`, text `#ff5370` | Always visible |

Buttons are plain `widget.Button` with custom importance — no icon text, label only, to keep footer row compact.

### Fields removed / added vs current
- Removed: `badgeRect canvas.Rectangle`, `badgeLabel canvas.Text` (replaced by new `statusBadge` composite)
- Added: `glowRect *canvas.Rectangle` (progress glow layer), `trackRect *canvas.Rectangle` (progress track)
- Retained: all data fields (`bar`, `percent`, `total`, `downloaded`, `speed`, `peers`, `eta`, `hash`, `errLabel`)

---

## Status Bar (`internal/ui/window.go`)

Layout: `container.NewBorder(nil, nil, versionLabel, right)` — unchanged structure, updated styling.

- Version label: `#ffffff22`, size 10
- Speed label: `#4d9fff` when active, `#ffffff33` when idle
- Count label: `#e8eeff`
- Status dot: `canvas.Circle` 7px, `#00e676` glow when active
- Background: `#0a0f20` (slightly darker than card bg) — set via `ColorNameHeaderBackground` or wrapper rectangle

---

## Files Changed

| File | Change |
|---|---|
| `internal/ui/app.go` | Update `darkColor()` palette; add `hexColor` helper (already exists) |
| `internal/ui/window.go` | Rebuild `buildSidebar()` with icon-only layout + status dot; tighten `buildStatusBar()` |
| `internal/ui/widgets/progressrow.go` | Rebuild `CreateRenderer()` with 4-row layout, glow bar, new badge, compact buttons |

No other files change. Engine, config, IPC, settings, about, add-torrent screens: untouched.

---

## Constraints

- Fyne v2 only — no external UI libraries
- No CSS/HTML — all glow effects approximated with layered `canvas.Rectangle` objects
- Light theme: not redesigned in this spec (existing light colors kept as-is)
- No new dependencies
- `.superpowers/` should be added to `.gitignore`
