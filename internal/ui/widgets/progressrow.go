// Package widgets provides custom Fyne widgets for GoTorrent.
package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
)

// ProgressRow renders one torrent entry as a rich 4-row glass card widget.
type ProgressRow struct {
	widget.BaseWidget

	// Card glass background
	cardBg     *canvas.Rectangle
	cardBorder *canvas.Rectangle

	// Row 1: name | badge
	iconName    *widget.Label
	badgeBg     *canvas.Rectangle
	badgeBorder *canvas.Rectangle
	badgeLabel  *canvas.Text

	// Row 2: custom glow progress bar
	progressValue float64
	progressTrack *canvas.Rectangle
	progressGlow  *canvas.Rectangle
	progressFill  *canvas.LinearGradient
	progressBar   fyne.CanvasObject
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

func NewProgressRow(onPause, onResume func(id string), onRemove, onOpen func(id string)) *ProgressRow {
	r := &ProgressRow{
		OnPause:  onPause,
		OnResume: onResume,
		OnRemove: onRemove,
		OnOpen:   onOpen,
	}

	// Glass card background
	r.cardBg = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x0f})
	r.cardBg.CornerRadius = 16

	r.cardBorder = canvas.NewRectangle(color.Transparent)
	r.cardBorder.CornerRadius = 16
	r.cardBorder.StrokeColor = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x17}
	r.cardBorder.StrokeWidth = 1

	// Name
	r.iconName = widget.NewLabel("📄  Loading…")
	r.iconName.TextStyle = fyne.TextStyle{Bold: true}

	// Badge: glass pill
	r.badgeBg = canvas.NewRectangle(glassStatusBadgeBg(engine.StatusQueued))
	r.badgeBg.CornerRadius = 10
	r.badgeBorder = canvas.NewRectangle(color.Transparent)
	r.badgeBorder.CornerRadius = 10
	r.badgeBorder.StrokeColor = glassStatusBadgeBorder(engine.StatusQueued)
	r.badgeBorder.StrokeWidth = 1
	r.badgeLabel = canvas.NewText("Queued", glassStatusColor(engine.StatusQueued))
	r.badgeLabel.TextSize = 9
	r.badgeLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Progress bar
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

func (r *ProgressRow) Update(state *engine.TorrentState) {
	r.currentID = state.ID
	r.currentStatus = state.Status

	// Badge colors
	r.badgeBg.FillColor = glassStatusBadgeBg(state.Status)
	r.badgeBg.Refresh()
	r.badgeBorder.StrokeColor = glassStatusBadgeBorder(state.Status)
	r.badgeBorder.Refresh()
	r.badgeLabel.Text = string(state.Status)
	r.badgeLabel.Color = glassStatusColor(state.Status)
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
		green := color.NRGBA{R: 0x30, G: 0xd1, B: 0x58, A: 0xff}
		r.progressFill.StartColor = green
		r.progressFill.EndColor = green
		r.progressGlow.FillColor = color.NRGBA{R: 0x30, G: 0xd1, B: 0x58, A: 0x40}
	case engine.StatusPaused, engine.StatusQueued:
		grey := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x33}
		r.progressFill.StartColor = grey
		r.progressFill.EndColor = grey
		r.progressGlow.FillColor = color.NRGBA{A: 0}
	case engine.StatusError:
		red := color.NRGBA{R: 0xff, G: 0x45, B: 0x3a, A: 0xff}
		r.progressFill.StartColor = red
		r.progressFill.EndColor = red
		r.progressGlow.FillColor = color.NRGBA{R: 0xff, G: 0x45, B: 0x3a, A: 0x40}
	default: // Downloading, Connecting, Verifying
		r.progressFill.StartColor = color.NRGBA{R: 0x0a, G: 0x84, B: 0xff, A: 0xff}
		r.progressFill.EndColor = color.NRGBA{R: 0x5e, G: 0x5c, B: 0xe6, A: 0xff}
		r.progressGlow.FillColor = color.NRGBA{R: 0x0a, G: 0x84, B: 0xff, A: 0x40}
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

func (r *ProgressRow) CreateRenderer() fyne.WidgetRenderer {
	// Badge: pill bg + border overlay + centered text
	badge := container.New(&badgeLayout{},
		r.badgeBg,
		r.badgeBorder,
		container.NewCenter(r.badgeLabel),
	)

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

	// Divider above row 4
	divider := canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x0f})

	// Card body
	body := container.NewVBox(row1, row2, row3, r.errLabel, divider, row4)
	padded := container.NewPadded(body)

	// Outer: glass card bg + border overlay + padded body
	outer := container.NewStack(r.cardBg, r.cardBorder, padded)

	return widget.NewSimpleRenderer(outer)
}

// separatorLabel returns a styled separator dot.
func separatorLabel() *widget.Label {
	l := widget.NewLabel("•")
	return l
}

// badgeLayout gives the badge a fixed minimum size.
type badgeLayout struct{}

func (*badgeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}

func (*badgeLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 24)
}

// glassStatusColor returns the Apple-accent text color for a status badge.
func glassStatusColor(s engine.Status) color.Color {
	switch s {
	case engine.StatusDownloading, engine.StatusConnecting, engine.StatusVerifying:
		return color.NRGBA{R: 0x0a, G: 0x84, B: 0xff, A: 0xff}
	case engine.StatusComplete:
		return color.NRGBA{R: 0x30, G: 0xd1, B: 0x58, A: 0xff}
	case engine.StatusError:
		return color.NRGBA{R: 0xff, G: 0x45, B: 0x3a, A: 0xff}
	default: // Paused, Queued
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x66}
	}
}

// glassStatusBadgeBg returns the tinted pill background for a status badge.
func glassStatusBadgeBg(s engine.Status) color.Color {
	switch s {
	case engine.StatusDownloading, engine.StatusConnecting, engine.StatusVerifying:
		return color.NRGBA{R: 0x0a, G: 0x84, B: 0xff, A: 0x26}
	case engine.StatusComplete:
		return color.NRGBA{R: 0x30, G: 0xd1, B: 0x58, A: 0x1f}
	case engine.StatusError:
		return color.NRGBA{R: 0xff, G: 0x45, B: 0x3a, A: 0x1f}
	default:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x14}
	}
}

// glassStatusBadgeBorder returns the tinted pill border for a status badge.
func glassStatusBadgeBorder(s engine.Status) color.Color {
	switch s {
	case engine.StatusDownloading, engine.StatusConnecting, engine.StatusVerifying:
		return color.NRGBA{R: 0x0a, G: 0x84, B: 0xff, A: 0x40}
	case engine.StatusComplete:
		return color.NRGBA{R: 0x30, G: 0xd1, B: 0x58, A: 0x33}
	case engine.StatusError:
		return color.NRGBA{R: 0xff, G: 0x45, B: 0x3a, A: 0x33}
	default:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x1a}
	}
}

// formatBytes returns human-readable size string.
func formatBytes(b float64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.2f GB", b/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", b/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.0f KB", b/(1<<10))
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

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
