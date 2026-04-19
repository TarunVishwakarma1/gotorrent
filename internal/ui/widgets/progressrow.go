// Package widgets provides custom Fyne widgets for GoTorrent.
package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
)

// ProgressRow renders one torrent entry as a rich card widget.
// It shows: name, size, progress bar + %, speed, peers, ETA,
// downloaded bytes, infohash, and status — plus action buttons.
type ProgressRow struct {
	widget.BaseWidget

	// Status strip — 4px colored left bar
	strip *canvas.Rectangle

	// Header: icon+name | status badge
	iconName   *widget.Label
	badgeRect  *canvas.Rectangle
	badgeLabel *canvas.Text

	// Progress row
	bar     *widget.ProgressBar
	percent *widget.Label
	total   *widget.Label

	// Stats row
	downloaded *widget.Label
	speed      *widget.Label
	peers      *widget.Label
	eta        *widget.Label

	// Footer: hash | buttons
	hash      *widget.Label
	pauseBtn  *widget.Button
	removeBtn *widget.Button
	openBtn   *widget.Button

	// Error row (hidden unless status==Error)
	errLabel *widget.Label

	currentID     string
	currentStatus engine.Status

	OnPause  func(id string)
	OnResume func(id string)
	OnRemove func(id string)
	OnOpen   func(id string)
}

// NewProgressRow creates a ProgressRow.
func NewProgressRow(onPause, onResume func(id string), onRemove, onOpen func(id string)) *ProgressRow {
	r := &ProgressRow{
		OnPause:  onPause,
		OnResume: onResume,
		OnRemove: onRemove,
		OnOpen:   onOpen,
	}

	r.strip = canvas.NewRectangle(color.NRGBA{R: 0x9e, G: 0x9e, B: 0x9e, A: 0xff})

	r.iconName = widget.NewLabel("📄  Loading…")
	r.iconName.TextStyle = fyne.TextStyle{Bold: true}

	r.badgeRect = canvas.NewRectangle(statusColor(engine.StatusQueued))
	r.badgeRect.CornerRadius = 5
	r.badgeLabel = canvas.NewText("Queued", color.White)
	r.badgeLabel.TextSize = 11
	r.badgeLabel.TextStyle = fyne.TextStyle{Bold: true}

	r.bar = widget.NewProgressBar()
	r.percent = widget.NewLabel("0.0%")
	r.total = widget.NewLabel("—")

	r.downloaded = widget.NewLabel("—")
	r.downloaded.TextStyle = fyne.TextStyle{Bold: true}
	r.speed = widget.NewLabel("↓ —")
	r.peers = widget.NewLabel("0 peers")
	r.eta = widget.NewLabel("ETA —")

	r.hash = widget.NewLabel("")
	r.hash.TextStyle = fyne.TextStyle{Monospace: true}

	r.pauseBtn = widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), func() {
		if r.currentID != "" {
			if r.currentStatus == engine.StatusPaused || r.currentStatus == engine.StatusError {
				if r.OnResume != nil {
					r.OnResume(r.currentID)
				}
			} else {
				if r.OnPause != nil {
					r.OnPause(r.currentID)
				}
			}
		}
	})
	r.removeBtn = widget.NewButtonWithIcon("Remove", theme.DeleteIcon(), func() {
		if r.currentID != "" && r.OnRemove != nil {
			r.OnRemove(r.currentID)
		}
	})
	r.openBtn = widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), func() {
		if r.currentID != "" && r.OnOpen != nil {
			r.OnOpen(r.currentID)
		}
	})
	r.removeBtn.Importance = widget.DangerImportance

	r.errLabel = widget.NewLabel("")
	r.errLabel.Wrapping = fyne.TextWrapWord

	r.ExtendBaseWidget(r)
	return r
}

// Update populates all fields from state. Thread-safe (Fyne handles canvas updates).
func (r *ProgressRow) Update(state *engine.TorrentState) {
	r.currentID = state.ID
	r.currentStatus = state.Status

	// — Status strip colour —
	sc := statusColor(state.Status)
	r.strip.FillColor = sc
	r.strip.Refresh()

	// — Badge —
	r.badgeRect.FillColor = sc
	r.badgeRect.Refresh()
	r.badgeLabel.Text = string(state.Status)
	r.badgeLabel.Color = badgeTextColor(state.Status)
	r.badgeLabel.Refresh()

	// — Name + icon —
	icon := "📄"
	if state.IsMultiFile {
		icon = "🗂"
	} else {
		icon = engine.FileIconForName(state.Name)
	}
	filesInfo := ""
	if state.IsMultiFile && len(state.Files) > 0 {
		filesInfo = fmt.Sprintf("  (%d files)", len(state.Files))
	}
	r.iconName.SetText(fmt.Sprintf("%s  %s%s", icon, state.Name, filesInfo))

	// — Progress —
	r.bar.SetValue(state.Progress)
	r.percent.SetText(fmt.Sprintf("%.1f%%", state.Progress*100))
	r.total.SetText(formatBytes(float64(state.TotalSize)))

	// — Downloaded —
	r.downloaded.SetText(fmt.Sprintf("%s / %s",
		formatBytes(float64(state.Downloaded)),
		formatBytes(float64(state.TotalSize))))

	// — Speed, peers, ETA —
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

	// — Infohash —
	if state.InfoHash != "" {
		h := state.InfoHash
		if len(h) > 16 {
			h = h[:16] + "…"
		}
		r.hash.SetText("#" + h)
	}

	// — Error —
	if state.Status == engine.StatusError && state.Error != "" {
		r.errLabel.SetText("⚠ " + state.Error)
		r.errLabel.Show()
	} else {
		r.errLabel.Hide()
	}

	// — Pause/Resume button —
	switch state.Status {
	case engine.StatusDownloading, engine.StatusConnecting:
		r.pauseBtn.SetText("Pause")
		r.pauseBtn.SetIcon(theme.MediaPauseIcon())
		r.pauseBtn.Show()
	case engine.StatusPaused:
		r.pauseBtn.SetText("Resume")
		r.pauseBtn.SetIcon(theme.MediaPlayIcon())
		r.pauseBtn.Show()
	case engine.StatusError:
		r.pauseBtn.SetText("Retry")
		r.pauseBtn.SetIcon(theme.ViewRefreshIcon())
		r.pauseBtn.Show()
	case engine.StatusComplete:
		r.pauseBtn.Hide()
	case engine.StatusQueued:
		r.pauseBtn.Hide()
	}

	r.Refresh()
}

// CreateRenderer implements fyne.Widget. Called once per list template item.
func (r *ProgressRow) CreateRenderer() fyne.WidgetRenderer {
	// Badge: coloured rect + text layered
	badge := container.New(&badgeLayout{}, r.badgeRect, container.NewCenter(r.badgeLabel))

	// Header row: [icon+name] [spacer] [badge]
	header := container.NewBorder(nil, nil, nil, badge, r.iconName)

	// Progress row: [bar] [percent] [total]
	progressRow := container.NewBorder(nil, nil, nil,
		container.NewHBox(r.percent, r.total),
		r.bar,
	)

	// Stats row
	statsRow := container.NewHBox(
		r.downloaded,
		separatorLabel(),
		r.speed,
		separatorLabel(),
		r.peers,
		separatorLabel(),
		r.eta,
	)

	// Footer: [hash] [spacer] [buttons]
	buttons := container.NewHBox(r.pauseBtn, r.removeBtn, r.openBtn)
	footer := container.NewBorder(nil, nil, r.hash, buttons)

	// Card body (everything except the strip)
	body := container.NewVBox(
		header,
		progressRow,
		statsRow,
		r.errLabel,
		footer,
	)

	// Outer: [4px strip] [body with padding]
	padded := container.NewPadded(body)
	outer := container.New(&stripLayout{stripW: 5}, r.strip, padded)

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

// stripLayout places the first child as a fixed-width left strip,
// then fills the rest of the space with the second child.
type stripLayout struct{ stripW float32 }

func (l *stripLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	objects[0].Resize(fyne.NewSize(l.stripW, size.Height))
	objects[0].Move(fyne.NewPos(0, 0))
	objects[1].Resize(fyne.NewSize(size.Width-l.stripW, size.Height))
	objects[1].Move(fyne.NewPos(l.stripW, 0))
}

func (l *stripLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(l.stripW, 0)
	}
	ms := objects[1].MinSize()
	return fyne.NewSize(ms.Width+l.stripW, ms.Height)
}

// statusColor maps a Status to its indicator colour.
func statusColor(s engine.Status) color.Color {
	switch s {
	case engine.StatusDownloading, engine.StatusConnecting:
		return color.NRGBA{R: 0x21, G: 0x96, B: 0xF3, A: 0xff} // blue
	case engine.StatusComplete:
		return color.NRGBA{R: 0x4c, G: 0xaf, B: 0x50, A: 0xff} // green
	case engine.StatusError:
		return color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff} // red
	case engine.StatusVerifying:
		return color.NRGBA{R: 0xff, G: 0x98, B: 0x00, A: 0xff} // orange
	default: // Queued, Paused
		return color.NRGBA{R: 0x9e, G: 0x9e, B: 0x9e, A: 0xff} // grey
	}
}

// badgeTextColor returns white or black for the badge text.
func badgeTextColor(_ engine.Status) color.Color {
	return color.White
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
