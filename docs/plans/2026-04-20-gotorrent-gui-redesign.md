# GoTorrent GUI Redesign Implementation Plan

**Goal:** Redesign GoTorrent's Fyne v2 UI to a dark neon blue/indigo theme with icon-only sidebar, glow progress bars, and rich 4-row torrent cards.

**Architecture:** Pure UI layer — engine, config, IPC, and non-download screens untouched. Three files change: `app.go` (colors), `window.go` (sidebar + status bar), `widgets/progressrow.go` (torrent card). Custom `progressFillLayout` replaces `widget.ProgressBar` to enable the gradient glow effect.

**Tech Stack:** Go, Fyne v2 (`fyne.io/fyne/v2`), `fyne.io/fyne/v2/canvas`, `fyne.io/fyne/v2/layout`

---

## File Map

| File | Change |
|---|---|
| `internal/ui/app.go` | Replace `darkColor()` palette with blue/indigo neon |
| `internal/ui/window.go` | Rebuild `buildSidebar()` (icon-only, bg rect, status dot); tighten `buildStatusBar()`; add `statusDot *widget.Label` to `MainWindow` struct; add `fixedSizeLayout` helper; adjust split offset |
| `internal/ui/widgets/progressrow.go` | Replace `*widget.ProgressBar` with custom glow bar; add `progressFillLayout`; rebuild `NewProgressRow`, `Update()`, `CreateRenderer()` |

---

### Task 1: Update dark theme color palette

**Files:**
- Modify: `internal/ui/app.go`

- [ ] **Step 1: Replace `darkColor()` body**

Open `internal/ui/app.go`. Replace the entire `darkColor` method (lines 120–152) with:

```go
func (t *goTorrentTheme) darkColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return hexColor(0x080d1f)
	case theme.ColorNameButton:
		return hexColor(0x0e1530)
	case theme.ColorNamePrimary:
		return hexColor(0x4d9fff)
	case theme.ColorNameFocus:
		return hexColor(0x4d9fff)
	case theme.ColorNameForeground:
		return hexColor(0xe8eeff)
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55}
	case theme.ColorNameInputBackground:
		return hexColor(0x0e1530)
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return hexColor(0x0e1530)
	case theme.ColorNameHeaderBackground:
		return hexColor(0x0a0f20)
	case theme.ColorNameSuccess:
		return hexColor(0x00e676)
	case theme.ColorNameWarning:
		return hexColor(0xffcb6b)
	case theme.ColorNameError:
		return hexColor(0xff5370)
	case theme.ColorNameSelection:
		return hexColor(0x4d9fff)
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 150}
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./internal/ui/...
```
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add internal/ui/app.go
git commit -m "ui: update dark theme to blue/indigo neon palette"
```

---

### Task 2: Rebuild icon-only sidebar

**Files:**
- Modify: `internal/ui/window.go`

- [ ] **Step 1: Add imports to `window.go`**

In the import block, add:
```go
"image/color"

"fyne.io/fyne/v2/canvas"
"fyne.io/fyne/v2/layout"
```

- [ ] **Step 2: Add `statusDot` field to `MainWindow` struct**

Replace the struct definition with:

```go
type MainWindow struct {
	win       fyne.Window
	gta       *GoTorrentApp
	content   *fyne.Container
	downloads *downloadsScreen
	currentID screenID

	navBtns   [4]*widget.Button
	statusDot *widget.Label // ● indicator at sidebar bottom, updated by status ticker
}
```

- [ ] **Step 3: Replace `buildSidebar()` body**

```go
func (mw *MainWindow) buildSidebar() fyne.CanvasObject {
	// Sidebar background rectangle
	bg := canvas.NewRectangle(color.NRGBA{R: 0x0e, G: 0x15, B: 0x30, A: 0xff})

	// Logo: accent-colored badge at top
	logoBg := canvas.NewRectangle(color.NRGBA{R: 0x4d, G: 0x9f, B: 0xff, A: 0xff})
	logoBg.CornerRadius = 8
	logoText := canvas.NewText("⬇", color.White)
	logoText.TextSize = 18
	logoText.TextStyle = fyne.TextStyle{Bold: true}
	logo := container.NewStack(
		container.New(&fixedSizeLayout{w: 36, h: 36}, logoBg),
		container.NewCenter(logoText),
	)

	// Nav buttons — icon only, no label text
	mw.navBtns[screenDownloads] = widget.NewButtonWithIcon("", theme.DownloadIcon(), func() {
		mw.showDownloads()
	})
	mw.navBtns[screenAddTorrent] = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		mw.showScreen(screenAddTorrent, addTorrentScreen(mw.gta, mw.win))
	})
	mw.navBtns[screenSettings] = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		mw.showScreen(screenSettings, settingsScreen(mw.gta, mw.win))
	})
	mw.navBtns[screenAbout] = widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		mw.showScreen(screenAbout, aboutScreen(mw.gta))
	})
	for _, btn := range mw.navBtns {
		btn.Alignment = widget.ButtonAlignCenter
	}

	// Status dot — updated by the status bar ticker
	mw.statusDot = widget.NewLabel("⚫")
	mw.statusDot.Alignment = fyne.TextAlignCenter

	nav := container.NewVBox(
		container.NewCenter(logo),
		widget.NewSeparator(),
		container.NewCenter(mw.navBtns[screenDownloads]),
		container.NewCenter(mw.navBtns[screenAddTorrent]),
		widget.NewSeparator(),
		container.NewCenter(mw.navBtns[screenSettings]),
		container.NewCenter(mw.navBtns[screenAbout]),
		layout.NewSpacer(),
		container.NewCenter(mw.statusDot),
		widget.NewLabel(""),
	)

	return container.NewStack(bg, nav)
}
```

- [ ] **Step 4: Add `fixedSizeLayout` helper at end of `window.go`**

Append to the bottom of `window.go`:

```go
// fixedSizeLayout constrains a single child to a fixed size.
type fixedSizeLayout struct{ w, h float32 }

func (l *fixedSizeLayout) Layout(objects []fyne.CanvasObject, _ fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(fyne.NewSize(l.w, l.h))
	}
}

func (l *fixedSizeLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(l.w, l.h)
}
```

- [ ] **Step 5: Narrow the sidebar split offset**

In `newMainWindow()`, change:
```go
split.SetOffset(0.15)
```
to:
```go
split.SetOffset(0.07)
```

- [ ] **Step 6: Build to verify**

```bash
go build ./internal/ui/...
```
Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/window.go
git commit -m "ui: rebuild sidebar as icon-only 56px panel"
```

---

### Task 3: Tighten status bar and wire sidebar dot

**Files:**
- Modify: `internal/ui/window.go`

- [ ] **Step 1: Replace `buildStatusBar()` body**

```go
func (mw *MainWindow) buildStatusBar() fyne.CanvasObject {
	versionLabel := widget.NewLabel("GoTorrent v" + appVersion)

	speedLabel := widget.NewLabel("↓ 0 B/s")
	countLabel := widget.NewLabel("Idle")

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var totalSpeed float64
			var active, total int
			for _, s := range mw.gta.Manager.GetAll() {
				total++
				if s.Status == engine.StatusDownloading {
					active++
					totalSpeed += s.Speed
				}
			}
			speedLabel.SetText("↓ " + formatBytes(totalSpeed) + "/s")
			if active > 0 {
				countLabel.SetText(fmt.Sprintf("%d/%d active", active, total))
				if mw.statusDot != nil {
					mw.statusDot.SetText("🟢")
				}
			} else if total > 0 {
				countLabel.SetText(fmt.Sprintf("%d torrents", total))
				if mw.statusDot != nil {
					mw.statusDot.SetText("⚫")
				}
			} else {
				countLabel.SetText("Idle")
				if mw.statusDot != nil {
					mw.statusDot.SetText("⚫")
				}
			}
		}
	}()

	right := container.NewHBox(speedLabel, widget.NewSeparator(), countLabel)
	return container.NewBorder(nil, nil, versionLabel, right)
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./internal/ui/...
```
Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/window.go
git commit -m "ui: tighten status bar, wire sidebar status dot"
```

---

### Task 4: Add glow progress bar layout

**Files:**
- Modify: `internal/ui/widgets/progressrow.go`

- [ ] **Step 1: Append `progressFillLayout` and `newProgressBar` to the file**

Add at the bottom of `progressrow.go` (before the final closing brace of the file — there is none since it's a package, just append):

```go
// progressFillLayout sizes three objects: track (full width), glow (fill + padding), fill (progress width).
type progressFillLayout struct {
	value *float64
}

func (l *progressFillLayout) Layout(objects []fyne.CanvasObject, sz fyne.Size) {
	if len(objects) < 3 {
		return
	}
	// [0] track — full width background
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(sz)

	fillW := sz.Width * float32(*l.value)
	if fillW < 0 {
		fillW = 0
	}

	// [1] glow — slightly taller than the bar, same left edge
	objects[1].Move(fyne.NewPos(0, -2))
	objects[1].Resize(fyne.NewSize(fillW, sz.Height+4))

	// [2] fill — exact progress width
	objects[2].Move(fyne.NewPos(0, 0))
	objects[2].Resize(fyne.NewSize(fillW, sz.Height))
}

func (l *progressFillLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(40, 9)
}

// newProgressBar returns (container, track, glow, fill) for a custom glow progress bar.
// value must be a pointer to a float64 owned by the calling widget; the layout reads it on every resize.
func newProgressBar(value *float64) (fyne.CanvasObject, *canvas.Rectangle, *canvas.Rectangle, *canvas.LinearGradient) {
	track := canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x10})
	track.CornerRadius = 5

	glow := canvas.NewRectangle(color.NRGBA{R: 0x4d, G: 0x9f, B: 0xff, A: 0x40})

	fill := canvas.NewHorizontalGradient(
		color.NRGBA{R: 0x4d, G: 0x9f, B: 0xff, A: 0xff},
		color.NRGBA{R: 0xa7, G: 0x8b, B: 0xfa, A: 0xff},
	)

	bar := container.New(&progressFillLayout{value: value}, track, glow, fill)
	return bar, track, glow, fill
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./internal/ui/widgets/...
```
Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/widgets/progressrow.go
git commit -m "ui: add progressFillLayout and newProgressBar glow helper"
```

---

### Task 5: Rebuild ProgressRow widget

**Files:**
- Modify: `internal/ui/widgets/progressrow.go`

- [ ] **Step 1: Replace `ProgressRow` struct**

Replace the entire `ProgressRow` struct definition with:

```go
// ProgressRow renders one torrent entry as a rich 4-row card widget.
type ProgressRow struct {
	widget.BaseWidget

	// Left status strip
	strip *canvas.Rectangle

	// Row 1: name | badge
	iconName   *widget.Label
	badgeRect  *canvas.Rectangle
	badgeLabel *canvas.Text

	// Row 2: custom glow progress bar
	progressValue float64
	progressTrack *canvas.Rectangle
	progressGlow  *canvas.Rectangle
	progressFill  *canvas.LinearGradient
	progressBar   fyne.CanvasObject // container holding track/glow/fill
	percent       *widget.Label
	total         *widget.Label

	// Row 3: stats
	downloaded *widget.Label
	speed      *widget.Label
	peers      *widget.Label
	eta        *widget.Label

	// Row 4: hash | buttons
	hash      *widget.Label
	pauseBtn  *widget.Button
	removeBtn *widget.Button
	openBtn   *widget.Button

	// Error row
	errLabel *widget.Label

	currentID     string
	currentStatus engine.Status

	OnPause  func(id string)
	OnResume func(id string)
	OnRemove func(id string)
	OnOpen   func(id string)
}
```

- [ ] **Step 2: Replace `NewProgressRow` constructor**

Replace the `NewProgressRow` function:

```go
func NewProgressRow(onPause, onResume func(id string), onRemove, onOpen func(id string)) *ProgressRow {
	r := &ProgressRow{
		OnPause:  onPause,
		OnResume: onResume,
		OnRemove: onRemove,
		OnOpen:   onOpen,
	}

	r.strip = canvas.NewRectangle(statusColor(engine.StatusQueued))

	r.iconName = widget.NewLabel("📄  Loading…")
	r.iconName.TextStyle = fyne.TextStyle{Bold: true}

	r.badgeRect = canvas.NewRectangle(statusColor(engine.StatusQueued))
	r.badgeRect.CornerRadius = 10
	r.badgeLabel = canvas.NewText("Queued", color.White)
	r.badgeLabel.TextSize = 9
	r.badgeLabel.TextStyle = fyne.TextStyle{Bold: true}

	r.progressBar, r.progressTrack, r.progressGlow, r.progressFill = newProgressBar(&r.progressValue)

	r.percent = widget.NewLabel("0.0%")
	r.percent.TextStyle = fyne.TextStyle{Bold: true}
	r.total = widget.NewLabel("—")

	r.downloaded = widget.NewLabel("—")
	r.speed = widget.NewLabel("↓ —")
	r.peers = widget.NewLabel("0 peers")
	r.eta = widget.NewLabel("ETA —")

	r.hash = widget.NewLabel("")
	r.hash.TextStyle = fyne.TextStyle{Monospace: true}

	r.pauseBtn = widget.NewButton("Pause", func() {
		if r.currentID == "" {
			return
		}
		if r.currentStatus == engine.StatusPaused || r.currentStatus == engine.StatusError {
			if r.OnResume != nil {
				r.OnResume(r.currentID)
			}
		} else {
			if r.OnPause != nil {
				r.OnPause(r.currentID)
			}
		}
	})
	r.pauseBtn.Importance = widget.HighImportance

	r.openBtn = widget.NewButton("Open", func() {
		if r.currentID != "" && r.OnOpen != nil {
			r.OnOpen(r.currentID)
		}
	})
	r.openBtn.Importance = widget.MediumImportance

	r.removeBtn = widget.NewButton("Remove", func() {
		if r.currentID != "" && r.OnRemove != nil {
			r.OnRemove(r.currentID)
		}
	})
	r.removeBtn.Importance = widget.DangerImportance

	r.errLabel = widget.NewLabel("")
	r.errLabel.Wrapping = fyne.TextWrapWord

	r.ExtendBaseWidget(r)
	return r
}
```

- [ ] **Step 3: Replace `Update()` method**

```go
func (r *ProgressRow) Update(state *engine.TorrentState) {
	r.currentID = state.ID
	r.currentStatus = state.Status

	sc := statusColor(state.Status)

	// Strip
	r.strip.FillColor = sc
	r.strip.Refresh()

	// Badge
	r.badgeRect.FillColor = sc
	r.badgeRect.Refresh()
	r.badgeLabel.Text = string(state.Status)
	r.badgeLabel.Color = badgeTextColor(state.Status)
	r.badgeLabel.Refresh()

	// Name
	icon := engine.FileIconForName(state.Name)
	if state.IsMultiFile {
		icon = "🗂"
	}
	filesInfo := ""
	if state.IsMultiFile && len(state.Files) > 0 {
		filesInfo = fmt.Sprintf("  (%d files)", len(state.Files))
	}
	r.iconName.SetText(fmt.Sprintf("%s  %s%s", icon, state.Name, filesInfo))

	// Progress bar value + colors
	r.progressValue = state.Progress
	switch state.Status {
	case engine.StatusComplete:
		green := color.NRGBA{R: 0x00, G: 0xe6, B: 0x76, A: 0xff}
		r.progressFill.StartColor = green
		r.progressFill.EndColor = green
		r.progressGlow.FillColor = color.NRGBA{R: 0x00, G: 0xe6, B: 0x76, A: 0x55}
	case engine.StatusPaused, engine.StatusQueued:
		grey := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x33}
		r.progressFill.StartColor = grey
		r.progressFill.EndColor = grey
		r.progressGlow.FillColor = color.NRGBA{A: 0}
	case engine.StatusError:
		red := color.NRGBA{R: 0xff, G: 0x53, B: 0x70, A: 0xff}
		r.progressFill.StartColor = red
		r.progressFill.EndColor = red
		r.progressGlow.FillColor = color.NRGBA{R: 0xff, G: 0x53, B: 0x70, A: 0x40}
	default: // Downloading, Connecting, Verifying
		r.progressFill.StartColor = color.NRGBA{R: 0x4d, G: 0x9f, B: 0xff, A: 0xff}
		r.progressFill.EndColor = color.NRGBA{R: 0xa7, G: 0x8b, B: 0xfa, A: 0xff}
		r.progressGlow.FillColor = color.NRGBA{R: 0x4d, G: 0x9f, B: 0xff, A: 0x55}
	}
	r.progressFill.Refresh()
	r.progressGlow.Refresh()

	// Percent + total
	r.percent.SetText(fmt.Sprintf("%.1f%%", state.Progress*100))
	r.total.SetText(formatBytes(float64(state.TotalSize)))

	// Downloaded
	r.downloaded.SetText(fmt.Sprintf("%s / %s",
		formatBytes(float64(state.Downloaded)),
		formatBytes(float64(state.TotalSize))))

	// Speed, peers, ETA
	if state.Status == engine.StatusDownloading {
		r.speed.SetText(fmt.Sprintf("↓ %s/s", formatBytes(state.Speed)))
		r.peers.SetText(fmt.Sprintf("%d peers", state.Peers))
		if state.ETA != "" && state.ETA != "Unknown" {
			r.eta.SetText("ETA " + state.ETA)
		} else {
			r.eta.SetText("ETA …")
		}
	} else if state.Status == engine.StatusComplete {
		r.speed.SetText("Complete")
		r.peers.SetText("")
		r.eta.SetText("")
	} else {
		r.speed.SetText("↓ —")
		r.peers.SetText("—")
		r.eta.SetText("—")
	}

	// Hash
	if state.InfoHash != "" {
		h := state.InfoHash
		if len(h) > 16 {
			h = h[:16] + "…"
		}
		r.hash.SetText("#" + h)
	}

	// Error label
	if state.Status == engine.StatusError && state.Error != "" {
		r.errLabel.SetText("⚠ " + state.Error)
		r.errLabel.Show()
	} else {
		r.errLabel.Hide()
	}

	// Pause / Resume / Retry button text
	switch state.Status {
	case engine.StatusDownloading, engine.StatusConnecting:
		r.pauseBtn.SetText("Pause")
		r.pauseBtn.Show()
	case engine.StatusPaused:
		r.pauseBtn.SetText("Resume")
		r.pauseBtn.Show()
	case engine.StatusError:
		r.pauseBtn.SetText("Retry")
		r.pauseBtn.Show()
	default:
		r.pauseBtn.Hide()
	}

	r.Refresh()
}
```

- [ ] **Step 4: Replace `CreateRenderer()` method**

```go
func (r *ProgressRow) CreateRenderer() fyne.WidgetRenderer {
	// Badge: colored pill bg + centered text
	badge := container.New(&badgeLayout{}, r.badgeRect, container.NewCenter(r.badgeLabel))

	// Row 1: [icon + name]  [badge]
	row1 := container.NewBorder(nil, nil, nil, badge, r.iconName)

	// Row 2: [progress bar (stretches)]  [percent]  [total]
	row2 := container.NewBorder(nil, nil, nil,
		container.NewHBox(r.percent, r.total),
		r.progressBar,
	)

	// Row 3: stats
	row3 := container.NewHBox(
		r.speed,
		separatorLabel(),
		r.peers,
		separatorLabel(),
		r.eta,
		separatorLabel(),
		r.downloaded,
	)

	// Row 4: [hash]  [buttons]
	buttons := container.NewHBox(r.pauseBtn, r.openBtn, r.removeBtn)
	row4 := container.NewBorder(nil, nil, r.hash, buttons)

	// Card body
	body := container.NewVBox(row1, row2, row3, r.errLabel, row4)
	padded := container.NewPadded(body)

	// Outer: 5px left strip + padded card body
	outer := container.New(&stripLayout{stripW: 5}, r.strip, padded)

	return widget.NewSimpleRenderer(outer)
}
```

- [ ] **Step 5: Remove now-unused `theme` import if compiler warns**

Check the import block. The `theme` import in `progressrow.go` was used for `theme.MediaPauseIcon()` etc. Since buttons are now text-only (no icons), the `theme` import is unused. Remove it:

```go
import (
    "fmt"
    "image/color"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"

    "github.com/tarunvishwakarma1/gotorrent/internal/engine"
)
```

- [ ] **Step 6: Build to verify**

```bash
go build ./internal/ui/...
```
Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/widgets/progressrow.go
git commit -m "ui: rebuild ProgressRow with 4-row card and glow progress bar"
```

---

### Task 6: Full build and smoke test

- [ ] **Step 1: Full build**

```bash
go build ./...
```
Expected: no output.

- [ ] **Step 2: Run the app**

```bash
go run ./cmd/gotorrent/
```

Verify visually:
- Window opens with narrow dark sidebar (~56px wide)
- Sidebar shows icons only (no text labels on buttons)
- Logo badge (blue rounded rect + arrow icon) at sidebar top
- Main area background is deep navy `#080d1f`
- Status bar: version on left, speed + count on right

- [ ] **Step 3: Add a test torrent**

Drag a `.torrent` file onto the window or press `Cmd+O`.

Verify the torrent card:
- 4 rows visible: name+badge / progress bar / stats / hash+buttons
- Progress bar shows blue→purple gradient with glow behind fill
- Paused torrents show grey bar, no glow
- Complete torrents show solid green bar with green glow
- Pause / Open / Remove buttons in row 4
- Error state shows red strip + red bar + error label

- [ ] **Step 4: Commit any fixups**

```bash
git add -p
git commit -m "ui: fix visual nits from smoke test"
```
