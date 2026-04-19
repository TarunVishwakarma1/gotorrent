package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/ui/widgets"
)

// addTorrentScreen builds the Add Torrent screen with a drop zone and browse button.
func addTorrentScreen(gta *GoTorrentApp, win fyne.Window) fyne.CanvasObject {
	dropZone := widgets.NewDropZone(func() {
		openTorrentFilePicker(gta, win)
	})

	hint := widget.NewLabel("Or drag a .torrent file anywhere onto this window")
	hint.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		widget.NewLabel(""),
		dropZone,
		widget.NewLabel(""),
		hint,
	)

	return container.NewCenter(content)
}

// openTorrentFilePicker shows a file open dialog filtered to .torrent files.
func openTorrentFilePicker(gta *GoTorrentApp, win fyne.Window) {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		path := reader.URI().Path()
		if !strings.HasSuffix(strings.ToLower(path), ".torrent") {
			dialog.ShowInformation("Invalid File",
				"Please select a .torrent file.", win)
			return
		}
		showPreviewDialog(gta, path, win)
	}, win)
	fd.Show()
}

// handleDroppedURIs processes files dropped onto the window.
// Only .torrent files are accepted; others show an error dialog.
func handleDroppedURIs(gta *GoTorrentApp, win fyne.Window, uris []fyne.URI) {
	for _, uri := range uris {
		path := uri.Path()
		if !strings.HasSuffix(strings.ToLower(path), ".torrent") {
			dialog.ShowInformation("Invalid File",
				"Only .torrent files are accepted.\nDropped: "+path, win)
			continue
		}
		showPreviewDialog(gta, path, win)
	}
}
