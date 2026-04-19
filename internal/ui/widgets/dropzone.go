package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// DropZone is a visual widget indicating that files can be dropped here.
// Actual OS file drop is handled at the window level via SetOnDropped;
// this widget provides the visual affordance.
type DropZone struct {
	widget.BaseWidget

	bg      *canvas.Rectangle
	icon    *widget.Label
	title   *widget.Label
	sub     *widget.Label
	browseBtn *widget.Button

	// OnBrowse is called when the user clicks the browse button.
	OnBrowse func()
}

// NewDropZone creates a DropZone widget.
// onBrowse is called when the user clicks "Browse for file".
func NewDropZone(onBrowse func()) *DropZone {
	dz := &DropZone{OnBrowse: onBrowse}

	dz.bg = canvas.NewRectangle(color.Transparent)
	dz.bg.StrokeWidth = 2
	dz.bg.CornerRadius = 8

	dz.icon = widget.NewLabel("📥")
	dz.icon.Alignment = fyne.TextAlignCenter

	dz.title = widget.NewLabel("Drop .torrent file here")
	dz.title.Alignment = fyne.TextAlignCenter
	dz.title.TextStyle = fyne.TextStyle{Bold: true}

	dz.sub = widget.NewLabel("or")
	dz.sub.Alignment = fyne.TextAlignCenter

	dz.browseBtn = widget.NewButtonWithIcon("Browse for file", theme.FolderOpenIcon(), func() {
		if dz.OnBrowse != nil {
			dz.OnBrowse()
		}
	})
	dz.browseBtn.Importance = widget.HighImportance

	dz.ExtendBaseWidget(dz)
	return dz
}

// SetHover visually highlights the drop zone when a file is dragged over.
func (dz *DropZone) SetHover(active bool) {
	if active {
		dz.bg.StrokeColor = theme.PrimaryColor()
	} else {
		dz.bg.StrokeColor = theme.ShadowColor()
	}
	dz.bg.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (dz *DropZone) CreateRenderer() fyne.WidgetRenderer {
	dz.bg.StrokeColor = theme.ShadowColor()

	inner := container.NewVBox(
		dz.icon,
		dz.title,
		dz.sub,
		container.NewCenter(dz.browseBtn),
	)
	content := container.New(&dropZoneLayout{}, dz.bg, container.NewCenter(inner))
	return widget.NewSimpleRenderer(content)
}

// dropZoneLayout stacks children and enforces a minimum widget size.
type dropZoneLayout struct{}

func (*dropZoneLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}

func (*dropZoneLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(400, 220)
}
