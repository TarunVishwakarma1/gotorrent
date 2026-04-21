package ui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
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
	nameLabel := widgets.StyledText(fmt.Sprintf("🗂  %s", tf.Name), 16, color.White, true, false)
	
	totalSize := int64(tf.Length)
	numPieces := len(tf.PieceHashes)
	pieceSizeKB := tf.PieceLength / 1024

	infoGrid := container.NewGridWithColumns(4,
		widgets.StyledText("Size:", 12, widgets.ColorGreyText, false, false), 
		widgets.StyledText(widgets.FormatBytes(float64(totalSize)), 13, widgets.ColorNeonCyan, true, true),
		
		widgets.StyledText("Pieces:", 12, widgets.ColorGreyText, false, false), 
		widgets.StyledText(fmt.Sprintf("%d", numPieces), 13, color.White, true, true),
		
		widgets.StyledText("Piece size:", 12, widgets.ColorGreyText, false, false), 
		widgets.StyledText(fmt.Sprintf("%d KB", pieceSizeKB), 13, color.White, true, true),
		
		widgets.StyledText("Tracker:", 12, widgets.ColorGreyText, false, false), 
		widgets.StyledText(truncate(tf.Announce, 30), 12, color.White, false, false),
		
		widgets.StyledText("InfoHash:", 12, widgets.ColorGreyText, false, false), 
		widgets.StyledText(truncate(fmt.Sprintf("%x", tf.InfoHash), 32), 11, widgets.ColorGreyText, false, true),
	)
	glassInfo := widgets.NewGlassPanel(infoGrid)

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

	selectedSizeLabel := widgets.StyledText("", 12, widgets.ColorNeonCyan, true, true)
	updateSelectedSize := func() {
		var sel int64
		for _, f := range files {
			if f.Selected {
				sel += f.Length
			}
		}
		selectedSizeLabel.Text = fmt.Sprintf("Selected: %s / %s",
			widgets.FormatBytes(float64(sel)), widgets.FormatBytes(float64(totalSize)))
		selectedSizeLabel.Refresh()
	}
	updateSelectedSize()

	fileTree := widgets.NewFileTree(files, updateSelectedSize)

	selectAllBtn := widget.NewButton("Select All", func() {
		fileTree.SelectAll(true)
		updateSelectedSize()
	})
	selectNoneBtn := widget.NewButton("Select None", func() {
		fileTree.SelectAll(false)
		updateSelectedSize()
	})
	selectRow := container.NewHBox(selectAllBtn, selectNoneBtn)

	filesSection := container.NewBorder(
		container.NewBorder(nil, nil, widgets.StyledText("Files Included", 13, color.White, true, false), selectRow),
		container.NewPadded(selectedSizeLabel),
		nil, nil,
		fileTree,
	)
	glassTree := widgets.NewGlassPanel(filesSection)

	// --- Save path ---
	savePathEntry := widget.NewEntry()
	savePathEntry.SetText(savePath)
	savePathBrowse := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			savePathEntry.SetText(uri.Path())
		}, win)
	})
	savePathRow := container.NewBorder(
		nil, nil,
		widgets.StyledText("Save to:", 12, widgets.ColorGreyText, false, false),
		savePathBrowse,
		savePathEntry,
	)

	// --- Buttons ---
	var previewDialog *dialog.CustomDialog

	cancelBtn := widget.NewButton("Cancel", func() {
		previewDialog.Hide()
	})
	
	startBtn := widgets.NewAnimatedPrimaryButton("Start Download", theme.MediaPlayIcon(), func() {
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
		if gta.window != nil {
			gta.window.showDownloads()
		}
	})

	footer := container.NewBorder(nil, nil, nil,
		container.NewHBox(cancelBtn, startBtn),
		savePathRow,
	)

	content := container.NewVBox(
		container.NewPadded(nameLabel),
		glassInfo,
		glassTree,
		container.NewPadded(footer),
	)

	previewDialog = dialog.NewCustom("", "Close", content, win)
	// Make it more compact & dense "Apple Style"
	previewDialog.Resize(fyne.NewSize(650, 480))
	previewDialog.Show()
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
