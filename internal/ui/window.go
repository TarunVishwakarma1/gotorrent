package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
	"github.com/tarunvishwakarma1/gotorrent/internal/ui/widgets"
)

// screenID identifies which screen is currently visible.
type screenID int

const (
	screenDownloads screenID = iota
	screenAddTorrent
	screenSettings
	screenAbout
)

// MainWindow is the application's primary window.
type MainWindow struct {
	win       fyne.Window
	gta       *GoTorrentApp
	content   *fyne.Container
	downloads *downloadsScreen
	currentID screenID
}

// newMainWindow creates and configures the main application window.
func newMainWindow(gta *GoTorrentApp) *MainWindow {
	mw := &MainWindow{gta: gta}

	mw.win = gta.App.NewWindow("GoTorrent")
	mw.win.Resize(fyne.NewSize(1200, 700))
	mw.win.CenterOnScreen()
	mw.win.SetMaster()

	// Set up OS file drop on the window.
	mw.win.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		handleDroppedURIs(gta, mw.win, uris)
	})

	// Build screens.
	mw.content = container.NewStack()
	downloadsObj, downloadsContent := newDownloadsScreen(gta, mw.win)
	mw.downloads = downloadsObj

	// Pre-build the content area with the downloads screen.
	mw.content.Objects = []fyne.CanvasObject{downloadsContent}

	topBar := mw.buildTopBar()
	statusBar := mw.buildStatusBar()

	root := container.NewBorder(topBar, statusBar, nil, nil, mw.content)
	mw.win.SetContent(root)

	// Keyboard shortcuts.
	mw.registerShortcuts()

	// System tray.
	gta.setupTray(mw.win)

	// Handle window close — minimize to tray if configured.
	mw.win.SetCloseIntercept(func() {
		if gta.Config.Get().MinimizeToTray {
			mw.win.Hide()
		} else {
			gta.Manager.Shutdown()
			gta.App.Quit()
		}
	})

	mw.selectNav(screenDownloads)
	return mw
}

// buildTopBar creates the top toolbar area to match the mockup.
func (mw *MainWindow) buildTopBar() fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 0x0d, G: 0x15, B: 0x20, A: 0xff})

	addBtn := widgets.NewAnimatedPrimaryButton("Add Torrent", theme.ContentAddIcon(), func() {
		openTorrentFilePicker(mw.gta, mw.win)
	})

	pauseAllBtn := widget.NewButtonWithIcon("Pause All", theme.MediaPauseIcon(), func() {
		for _, s := range mw.gta.Manager.GetAll() {
			if s.Status == engine.StatusDownloading || s.Status == engine.StatusConnecting {
				_ = mw.gta.Manager.Pause(s.ID)
			}
		}
	})

	// Speed labels on the right
	downSpeed := widgets.StyledText("↓ 0 MB/s", 12, widgets.ColorNeonCyan, true, true)
	upSpeed := widgets.StyledText("↑ 0 MB/s", 12, widgets.ColorNeonGreen, true, true)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var totalDown float64
			for _, s := range mw.gta.Manager.GetAll() {
				if s.Status == engine.StatusDownloading {
					totalDown += s.Speed
				}
			}
			downSpeed.Text = fmt.Sprintf("↓ %s/s", widgets.FormatBytes(totalDown))
			downSpeed.Refresh()
		}
	}()

	left := container.NewHBox(addBtn, pauseAllBtn)
	right := container.NewHBox(downSpeed, upSpeed)

	bar := container.NewBorder(nil, nil, left, right)
	padded := container.NewPadded(bar)

	// A subtle bottom border
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = widgets.ColorGlassBorder
	border.StrokeWidth = 1

	return container.NewStack(bg, border, padded)
}

// buildStatusBar creates the bottom minimally-styled status bar.
func (mw *MainWindow) buildStatusBar() fyne.CanvasObject {
	barBg := canvas.NewRectangle(color.NRGBA{R: 0x0a, G: 0x0f, B: 0x1a, A: 0xff})

	countLabel := widget.NewLabel("0 downloading · 0 total")
	downloadedLabel := widget.NewLabel("Downloaded: 0 GB")

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var active, total int
			var totalDownloaded int64
			for _, s := range mw.gta.Manager.GetAll() {
				total++
				if s.Status == engine.StatusDownloading || s.Status == engine.StatusConnecting {
					active++
				}
				totalDownloaded += s.Downloaded
			}
			countLabel.SetText(fmt.Sprintf("%d downloading · %d total", active, total))
			downloadedLabel.SetText(fmt.Sprintf("Downloaded: %s", formatBytes(float64(totalDownloaded))))
		}
	}()

	topBorder := canvas.NewRectangle(color.Transparent)
	topBorder.StrokeColor = color.NRGBA{R: 0x00, G: 0xd4, B: 0xff, A: 0x15}
	topBorder.StrokeWidth = 1

	bar := container.NewBorder(nil, nil, countLabel, downloadedLabel)
	padded := container.NewPadded(bar)

	return container.NewStack(barBg, topBorder, padded)
}

// registerShortcuts adds keyboard shortcuts to the window canvas.
func (mw *MainWindow) registerShortcuts() {
	// Cmd/Ctrl+O — open file picker
	mw.win.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierShortcutDefault},
		func(fyne.Shortcut) { openTorrentFilePicker(mw.gta, mw.win) },
	)
	// Cmd/Ctrl+, — settings
	mw.win.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: fyne.KeyModifierShortcutDefault},
		func(fyne.Shortcut) { mw.showScreen(screenSettings, settingsScreen(mw.gta, mw.win)) },
	)
	// Cmd/Ctrl+Q — quit
	mw.win.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierShortcutDefault},
		func(fyne.Shortcut) {
			mw.gta.Manager.Shutdown()
			mw.gta.App.Quit()
		},
	)
}

// showDownloads switches the content area to the downloads screen.
func (mw *MainWindow) showDownloads() {
	mw.showScreen(screenDownloads, mw.downloads.stack)
}

// showScreen switches to an arbitrary screen.
func (mw *MainWindow) showScreen(id screenID, content fyne.CanvasObject) {
	mw.selectNav(id)
	mw.content.Objects = []fyne.CanvasObject{content}
	mw.content.Refresh()
}

// selectNav switches the navigation state. (Kept for compatibility with other screens)
func (mw *MainWindow) selectNav(id screenID) {
	mw.currentID = id
	if id == screenDownloads {
		allStates := mw.gta.Manager.GetAll()
		mw.downloads.updateStates(allStates)
	}
}

// OpenTorrentFile shows the preview dialog for a torrent file path.
// Safe to call from any goroutine (e.g. IPC handler).
func (mw *MainWindow) OpenTorrentFile(path string) {
	if !strings.HasSuffix(strings.ToLower(path), ".torrent") {
		return
	}
	mw.win.Show()
	mw.win.RequestFocus()
	showPreviewDialog(mw.gta, path, mw.win)
}

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
