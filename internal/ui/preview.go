package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
	"github.com/tarunvishwakarma1/gotorrent/internal/ui/widgets"
	"github.com/tarunvishwakarma1/gotorrent/torrent"
)

// showPreviewDialog opens the torrent preview modal.
// After the user confirms, the torrent is added to the manager.
func showPreviewDialog(gta *GoTorrentApp, torrentPath string, win fyne.Window) {
	data, err := os.ReadFile(torrentPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("cannot read torrent file: %w", err), win)
		return
	}

	tf, err := torrent.NewTorrentFile(string(data))
	if err != nil {
		dialog.ShowError(fmt.Errorf("invalid torrent file: %w", err), win)
		return
	}

	cfg := gta.Config.Get()
	savePath := cfg.SavePath

	// --- Header info ---
	nameLabel := widget.NewLabel(fmt.Sprintf("🗂  %s", tf.Name))
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}

	totalSize := int64(tf.Length)
	numPieces := len(tf.PieceHashes)
	pieceSizeKB := tf.PieceLength / 1024

	infoGrid := container.NewGridWithColumns(4,
		widget.NewLabel("Size:"), widget.NewLabel(formatBytes(float64(totalSize))),
		widget.NewLabel("Pieces:"), widget.NewLabel(fmt.Sprintf("%d", numPieces)),
		widget.NewLabel("Piece size:"), widget.NewLabel(fmt.Sprintf("%d KB", pieceSizeKB)),
		widget.NewLabel("Tracker:"), widget.NewLabel(truncate(tf.Announce, 30)),
		widget.NewLabel("InfoHash:"), widget.NewLabel(truncate(fmt.Sprintf("%x", tf.InfoHash), 32)),
	)

	// --- File tree ---
	files := make([]engine.FileState, 0)
	if tf.IsMultiFile {
		for _, f := range tf.Files {
			name := ""
			if len(f.Path) > 0 {
				name = f.Path[len(f.Path)-1]
			}
			files = append(files, engine.FileState{
				Path:     f.Path,
				Length:   f.Length,
				Selected: true,
				Icon:     engine.FileIconForName(name),
			})
		}
	} else {
		files = append(files, engine.FileState{
			Path:     []string{tf.Name},
			Length:   int64(tf.Length),
			Selected: true,
			Icon:     engine.FileIconForName(tf.Name),
		})
	}

	// selectedSizeLabel shows updated selection summary
	selectedSizeLabel := widget.NewLabel("")
	updateSelectedSize := func() {
		var sel int64
		for _, f := range files {
			if f.Selected {
				sel += f.Length
			}
		}
		selectedSizeLabel.SetText(fmt.Sprintf("Selected: %s / %s",
			formatBytes(float64(sel)), formatBytes(float64(totalSize))))
	}
	updateSelectedSize()

	fileTree := widgets.NewFileTree(files, updateSelectedSize)

	selectAllBtn := widget.NewButton("☑ Select All", func() {
		fileTree.SelectAll(true)
		updateSelectedSize()
	})
	selectNoneBtn := widget.NewButton("☐ Select None", func() {
		fileTree.SelectAll(false)
		updateSelectedSize()
	})
	selectRow := container.NewHBox(selectAllBtn, selectNoneBtn)

	filesSection := container.NewBorder(
		container.NewBorder(nil, nil, sectionLabel("Files"), selectRow),
		selectedSizeLabel,
		nil, nil,
		fileTree,
	)

	// --- Save path ---
	savePathEntry := widget.NewEntry()
	savePathEntry.SetText(savePath)
	savePathBrowse := widget.NewButton("📁", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			savePathEntry.SetText(uri.Path())
		}, win)
	})
	savePathRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Save to:"),
		savePathBrowse,
		savePathEntry,
	)

	// --- Buttons ---
	var previewDialog *dialog.CustomDialog

	cancelBtn := widget.NewButton("Cancel", func() {
		previewDialog.Hide()
	})
	startBtn := widget.NewButton("▶ Start Download", func() {
		dest := savePathEntry.Text
		if dest == "" {
			dest = cfg.SavePath
		}
		dest = filepath.Clean(dest)

		previewDialog.Hide()

		_, err := gta.Manager.Add(torrentPath, dest)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		// Switch to downloads screen
		if gta.window != nil {
			gta.window.showDownloads()
		}
	})
	startBtn.Importance = widget.HighImportance

	footer := container.NewBorder(nil, nil, nil,
		container.NewHBox(cancelBtn, startBtn),
		savePathRow,
	)

	content := container.NewVBox(
		nameLabel,
		widget.NewSeparator(),
		infoGrid,
		widget.NewSeparator(),
		filesSection,
		widget.NewSeparator(),
		footer,
	)

	previewDialog = dialog.NewCustom("Torrent Preview", "Close", content, win)
	previewDialog.Resize(fyne.NewSize(700, 600))
	previewDialog.Show()
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
