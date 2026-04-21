package widgets

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Brand colors
var (
	ColorNeonCyan            = color.NRGBA{R: 0x00, G: 0xd4, B: 0xff, A: 0xff}
	ColorNeonGreen           = color.NRGBA{R: 0x00, G: 0xff, B: 0x88, A: 0xff}
	ColorNavyDark            = color.NRGBA{R: 0x0e, G: 0x16, B: 0x24, A: 0xff}
	ColorGreyText            = color.NRGBA{R: 0x8a, G: 0x9a, B: 0xb0, A: 0xff}
	ColorGlassBg             = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x06}
	ColorGlassBorder         = color.NRGBA{R: 0x00, G: 0xd4, B: 0xff, A: 0x15}
	ColorHoverSoftWhiteGreen = color.NRGBA{R: 0xcc, G: 0xff, B: 0xee, A: 0x4d} // Softer Bright White/Green
)

// hoverableGlassCard is a custom widget implementing desktop.Hoverable
type hoverableGlassCard struct {
	widget.BaseWidget
	content     fyne.CanvasObject
	border      *canvas.Rectangle
	ambientAnim *fyne.Animation
	hoverAnim   *fyne.Animation
	baseColor   color.Color
	hoverColor  color.Color
	hovered     bool
}

func (h *hoverableGlassCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

func (h *hoverableGlassCard) MouseIn(*desktop.MouseEvent) {
	h.hovered = true
	if h.hoverAnim != nil {
		h.hoverAnim.Stop()
	}
	h.hoverAnim = canvas.NewColorRGBAAnimation(h.border.StrokeColor, h.hoverColor, time.Millisecond*200, func(c color.Color) {
		h.border.StrokeColor = c
		canvas.Refresh(h.border)
	})
	h.hoverAnim.Start()
}

func (h *hoverableGlassCard) MouseOut() {
	h.hovered = false
	if h.hoverAnim != nil {
		h.hoverAnim.Stop()
	}
	h.hoverAnim = canvas.NewColorRGBAAnimation(h.border.StrokeColor, h.baseColor, time.Millisecond*300, func(c color.Color) {
		h.border.StrokeColor = c
		canvas.Refresh(h.border)
	})
	h.hoverAnim.Start()
}
func (h *hoverableGlassCard) MouseMoved(*desktop.MouseEvent) {}

// NewGlassPanel wraps content in a stylized translucent Neo-Brutalist card which perfectly brightens on mouse hover.
func NewGlassPanel(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorGlassBg)
	bg.CornerRadius = 12

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = 12
	border.StrokeColor = ColorGlassBorder
	border.StrokeWidth = 1

	card := &hoverableGlassCard{
		border:     border,
		baseColor:  ColorGlassBorder,
		hoverColor: ColorHoverSoftWhiteGreen,
	}

	// Pulsing ambient glow for glass panels while idle
	card.ambientAnim = canvas.NewColorRGBAAnimation(
		ColorGlassBorder,
		color.NRGBA{R: 0x00, G: 0xd4, B: 0xff, A: 0x22}, // ambient bright tick
		time.Second*3,
		func(c color.Color) {
			if !card.hovered {
				card.border.StrokeColor = c
				canvas.Refresh(card.border)
			}
		},
	)
	card.ambientAnim.AutoReverse = true
	card.ambientAnim.RepeatCount = fyne.AnimationRepeatForever
	card.ambientAnim.Start()

	padded := container.NewPadded(content)
	card.content = container.NewStack(bg, border, padded)
	card.ExtendBaseWidget(card)

	return card
}

// NewNeonPill returns an animated floating pill for statuses that smoothly breathes.
func NewNeonPill(text string, baseColor color.Color) fyne.CanvasObject {
	bgR, bgG, bgB, _ := baseColor.RGBA()
	r8, g8, b8 := uint8(bgR>>8), uint8(bgG>>8), uint8(bgB>>8)

	bgTint := color.NRGBA{R: r8, G: g8, B: b8, A: 0x11}

	bg := canvas.NewRectangle(bgTint)
	bg.CornerRadius = 8

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = 8
	border.StrokeColor = color.NRGBA{R: r8, G: g8, B: b8, A: 0x55}
	border.StrokeWidth = 1

	lbl := canvas.NewText(text, baseColor)
	lbl.TextSize = 10
	lbl.TextStyle = fyne.TextStyle{Bold: true}

	anim := canvas.NewColorRGBAAnimation(
		bgTint,
		color.NRGBA{R: r8, G: g8, B: b8, A: 0x33},
		time.Second*2, func(c color.Color) {
			bg.FillColor = c
			canvas.Refresh(bg)
		},
	)
	anim.AutoReverse = true
	anim.RepeatCount = fyne.AnimationRepeatForever
	anim.Start()

	return container.New(&pillLayout{}, bg, border, container.NewCenter(lbl))
}

type pillLayout struct{}

func (l *pillLayout) Layout(objects []fyne.CanvasObject, sz fyne.Size) {
	for _, o := range objects {
		o.Resize(sz)
		o.Move(fyne.NewPos(0, 0))
	}
}
func (l *pillLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 3 {
		return fyne.NewSize(60, 20)
	}
	minLbl := objects[2].MinSize()
	return fyne.NewSize(minLbl.Width+16, minLbl.Height+8)
}

// hoverableButton implements interactive glow expansion for primary buttons
type hoverableButton struct {
	widget.BaseWidget
	content    fyne.CanvasObject
	glow       *canvas.Rectangle
	hoverAnim  *fyne.Animation
	baseColor  color.Color
	hoverColor color.Color
	hovered    bool
}

func (h *hoverableButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

func (h *hoverableButton) MouseIn(*desktop.MouseEvent) {
	h.hovered = true
	if h.hoverAnim != nil {
		h.hoverAnim.Stop()
	}
	h.hoverAnim = canvas.NewColorRGBAAnimation(h.glow.FillColor, h.hoverColor, time.Millisecond*200, func(c color.Color) {
		h.glow.FillColor = c
		canvas.Refresh(h.glow)
	})
	h.hoverAnim.Start()
}

func (h *hoverableButton) MouseOut() {
	h.hovered = false
	if h.hoverAnim != nil {
		h.hoverAnim.Stop()
	}
	h.hoverAnim = canvas.NewColorRGBAAnimation(h.glow.FillColor, h.baseColor, time.Millisecond*300, func(c color.Color) {
		h.glow.FillColor = c
		canvas.Refresh(h.glow)
	})
	h.hoverAnim.Start()
}
func (h *hoverableButton) MouseMoved(*desktop.MouseEvent) {}

// NewAnimatedPrimaryButton returns a strongly glowing action button which intensifies upon hover.
func NewAnimatedPrimaryButton(text string, icon fyne.Resource, action func()) fyne.CanvasObject {
	baseGlow := color.NRGBA{R: 0x00, G: 0xd4, B: 0xff, A: 0x11}

	glow := canvas.NewRectangle(baseGlow)
	glow.CornerRadius = 8

	btn := widget.NewButtonWithIcon(text, icon, action)
	btn.Importance = widget.HighImportance

	hb := &hoverableButton{
		glow:       glow,
		baseColor:  baseGlow,
		hoverColor: color.NRGBA{R: 0x00, G: 0xd4, B: 0xff, A: 0x66}, // vibrant surge
	}
	hb.content = container.NewStack(glow, btn)
	hb.ExtendBaseWidget(hb)

	// Add an ambient pulse loop as well
	anim := canvas.NewColorRGBAAnimation(
		baseGlow,
		color.NRGBA{R: 0x00, G: 0xd4, B: 0xff, A: 0x22},
		time.Second*2, func(c color.Color) {
			if !hb.hovered { // limit override if hovered
				hb.glow.FillColor = c
				canvas.Refresh(hb.glow)
			}
		},
	)
	anim.AutoReverse = true
	anim.RepeatCount = fyne.AnimationRepeatForever
	anim.Start()

	return hb
}

// StyledText is a typography helper for monospace, bold sizes.
func StyledText(text string, size float32, clr color.Color, bold bool, mono bool) *canvas.Text {
	txt := canvas.NewText(text, clr)
	txt.TextSize = size
	txt.TextStyle = fyne.TextStyle{Bold: bold, Monospace: mono}
	return txt
}

// NewAnimatedProgressBar builds a breathing neon progress bar with multiple bloom drop-shadow layers.
func NewAnimatedProgressBar(value *float64, isComplete bool) (fyne.CanvasObject, *canvas.Rectangle, *canvas.Rectangle, *canvas.Rectangle, *canvas.LinearGradient) {
	track := canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x06})
	track.CornerRadius = 3

	glowColor := ColorNeonCyan
	fillColors := []color.Color{ColorNeonCyan, color.NRGBA{R: 0x00, G: 0x55, B: 0xff, A: 0xff}}

	if isComplete {
		glowColor = ColorNeonGreen
		fillColors = []color.Color{ColorNeonGreen, color.NRGBA{R: 0x00, G: 0xcc, B: 0x66, A: 0xff}}
	}

	// Layer 1: Core bar Glow
	glowColor.A = 0x33
	glow1 := canvas.NewRectangle(glowColor)
	glow1.CornerRadius = 3

	// Layer 2: Deep outer Bloom (larger expansion)
	bloomColor := glowColor
	bloomColor.A = 0x11
	glow2 := canvas.NewRectangle(bloomColor)
	glow2.CornerRadius = 6

	fill := canvas.NewHorizontalGradient(fillColors[0], fillColors[1])

	anim := canvas.NewColorRGBAAnimation(glowColor, color.NRGBA{R: glowColor.R, G: glowColor.G, B: glowColor.B, A: 0x55}, time.Second*2, func(c color.Color) {
		glow1.FillColor = c
		canvas.Refresh(glow1)
	})
	anim.AutoReverse = true
	anim.RepeatCount = fyne.AnimationRepeatForever
	anim.Start()

	barContainer := container.New(&animatedProgressFillLayout{value: value}, track, glow2, glow1, fill)
	return barContainer, track, glow1, glow2, fill
}

func FormatBytes(b float64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", b/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", b/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", b/(1<<10))
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

type animatedProgressFillLayout struct{ value *float64 }

func (l *animatedProgressFillLayout) Layout(objects []fyne.CanvasObject, sz fyne.Size) {
	if len(objects) < 4 {
		return
	}

	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(sz)

	fillW := sz.Width * float32(*l.value)
	if fillW < 0 {
		fillW = 0
	}

	// Deep Bloom
	objects[1].Move(fyne.NewPos(0, -4))
	objects[1].Resize(fyne.NewSize(fillW, sz.Height+8))

	// Tight Glow
	objects[2].Move(fyne.NewPos(0, -2))
	objects[2].Resize(fyne.NewSize(fillW, sz.Height+4))

	// Core Fill
	objects[3].Move(fyne.NewPos(0, 0))
	objects[3].Resize(fyne.NewSize(fillW, sz.Height))
}
func (l *animatedProgressFillLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(60, 6)
}
