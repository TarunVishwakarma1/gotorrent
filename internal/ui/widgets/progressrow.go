package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
)

// ProgressRow renders one torrent entry in a minimalist grid layout,
// wrapped in an elegant glass-morphic panel.
type ProgressRow struct {
	widget.BaseWidget

	// Row 1
	iconName            *canvas.Text
	statusPillContainer *fyne.Container

	// Row 2
	progressValue float64
	progressTrack *canvas.Rectangle
	progressGlow1 *canvas.Rectangle
	progressGlow2 *canvas.Rectangle
	progressFill  *canvas.LinearGradient
	progressBar   fyne.CanvasObject

	// Row 3 (Labels)
	sizeLabel *canvas.Text
	percent   *canvas.Text
	speed     *canvas.Text
	peers     *canvas.Text

	// Row 3 (Buttons)
	pauseBtn  *widget.Button
	removeBtn *widget.Button
	openBtn   *widget.Button

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

	r.iconName = StyledText("Loading…", 13, color.White, true, false)
	r.statusPillContainer = container.NewStack() // Will dynamically load NewNeonPill in Update()

	r.progressBar, r.progressTrack, r.progressGlow1, r.progressGlow2, r.progressFill = NewAnimatedProgressBar(&r.progressValue, false)

	r.sizeLabel = StyledText("—", 11, ColorGreyText, false, true)
	r.percent = StyledText("0.0%", 11, ColorGreyText, true, true)
	r.speed = StyledText("—", 11, ColorNeonCyan, true, true)
	r.peers = StyledText("", 11, ColorGreyText, false, true)

	r.pauseBtn = widget.NewButton("⏸", func() {
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
	r.pauseBtn.Importance = widget.LowImportance

	r.openBtn = widget.NewButton("📂", func() {
		if r.currentID != "" && r.OnOpen != nil {
			r.OnOpen(r.currentID)
		}
	})
	r.openBtn.Importance = widget.LowImportance

	r.removeBtn = widget.NewButton("✖", func() {
		if r.currentID != "" && r.OnRemove != nil {
			r.OnRemove(r.currentID)
		}
	})
	r.removeBtn.Importance = widget.LowImportance

	r.errLabel = widget.NewLabel("")
	r.errLabel.Wrapping = fyne.TextWrapWord
	r.errLabel.Hide()

	r.ExtendBaseWidget(r)
	return r
}

func (r *ProgressRow) Update(state *engine.TorrentState) {
	r.currentID = state.ID
	r.currentStatus = state.Status

	r.iconName.Text = state.Name
	r.iconName.Refresh()

	// Update pill dynamically
	var pill fyne.CanvasObject
	if state.Status == engine.StatusComplete {
		pill = NewNeonPill("Seeding", ColorNeonGreen)
	} else if state.Status == engine.StatusDownloading {
		pill = NewNeonPill("Downloading", ColorNeonCyan)
	} else {
		pill = NewNeonPill(string(state.Status), ColorGreyText)
	}
	r.statusPillContainer.Objects = []fyne.CanvasObject{pill}
	r.statusPillContainer.Refresh()

	// Progress updates
	r.progressValue = state.Progress
	if state.Status == engine.StatusComplete {
		r.progressFill.StartColor = ColorNeonGreen
		r.progressFill.EndColor = color.NRGBA{R: 0x00, G: 0xcc, B: 0x66, A: 0xff}
		r.progressGlow1.FillColor = color.NRGBA{R: ColorNeonGreen.R, G: ColorNeonGreen.G, B: ColorNeonGreen.B, A: 0x33}
		r.progressGlow2.FillColor = color.NRGBA{R: ColorNeonGreen.R, G: ColorNeonGreen.G, B: ColorNeonGreen.B, A: 0x11}
	} else {
		r.progressFill.StartColor = ColorNeonCyan
		r.progressFill.EndColor = color.NRGBA{R: 0x00, G: 0x55, B: 0xff, A: 0xff}
		r.progressGlow1.FillColor = color.NRGBA{R: ColorNeonCyan.R, G: ColorNeonCyan.G, B: ColorNeonCyan.B, A: 0x33}
		r.progressGlow2.FillColor = color.NRGBA{R: ColorNeonCyan.R, G: ColorNeonCyan.G, B: ColorNeonCyan.B, A: 0x11}
	}
	r.progressFill.Refresh()
	r.progressGlow1.Refresh()
	r.progressGlow2.Refresh()

	r.sizeLabel.Text = FormatBytes(float64(state.TotalSize))
	r.sizeLabel.Refresh()

	r.percent.Text = fmt.Sprintf("%.1f%%", state.Progress*100)
	r.percent.Refresh()

	if state.Status == engine.StatusDownloading {
		r.speed.Text = fmt.Sprintf("%s/s", FormatBytes(state.Speed))
		r.speed.Color = ColorNeonCyan
		r.peers.Text = fmt.Sprintf("%d peers", state.Peers)
	} else if state.Status == engine.StatusComplete {
		r.speed.Text = "Complete"
		r.speed.Color = ColorNeonGreen
		r.peers.Text = ""
	} else {
		r.speed.Text = "—"
		r.speed.Color = ColorGreyText
		r.peers.Text = ""
	}
	r.speed.Refresh()
	r.peers.Refresh()

	if state.Status == engine.StatusError && state.Error != "" {
		r.errLabel.SetText("⚠ " + state.Error)
		r.errLabel.Show()
	} else {
		r.errLabel.Hide()
	}

	if state.Status == engine.StatusDownloading || state.Status == engine.StatusConnecting {
		r.pauseBtn.SetText("⏸")
	} else {
		r.pauseBtn.SetText("▶")
	}

	r.Refresh()
}

func (r *ProgressRow) CreateRenderer() fyne.WidgetRenderer {
	row1 := container.NewBorder(nil, nil, r.iconName, r.statusPillContainer)
	row2 := container.NewPadded(r.progressBar)

	stats := container.NewHBox(
		r.sizeLabel,
		layout.NewSpacer(), r.percent,
		layout.NewSpacer(), r.speed,
		layout.NewSpacer(), r.peers,
	)

	btns := container.NewHBox(r.pauseBtn, r.openBtn, r.removeBtn)
	row3 := container.NewBorder(nil, nil, stats, btns)

	bodyOuter := container.NewVBox(row1, row2, row3)
	if r.errLabel.Visible() {
		bodyOuter.Add(r.errLabel)
	}

	// Wrap the entire list item inside our new sleek glass panel
	glassCard := NewGlassPanel(bodyOuter)

	// Separate rows with a tiny bit of bottom margin outside the card
	paddedForList := container.NewBorder(nil, container.New(&marginBottom{h: 8}), nil, nil, glassCard)
	return widget.NewSimpleRenderer(paddedForList)
}

type marginBottom struct{ h float32 }

func (l *marginBottom) Layout(objects []fyne.CanvasObject, sz fyne.Size) {}
func (l *marginBottom) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(10, l.h)
}
