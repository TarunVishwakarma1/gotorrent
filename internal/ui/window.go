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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
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

	navBtns   [4]*widget.Button
	statusDot *widget.Label // ● indicator at sidebar bottom, updated by status ticker
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

	sidebar := mw.buildSidebar()
	statusBar := mw.buildStatusBar()

	split := container.NewHSplit(sidebar, mw.content)
	split.SetOffset(0.07)

	root := container.NewBorder(nil, statusBar, nil, nil, split)
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

// buildSidebar creates the glass navigation panel.
func (mw *MainWindow) buildSidebar() fyne.CanvasObject {
	// Glass background panel
	bg := canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x12})
	bg.CornerRadius = 18
	bg.StrokeColor = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x1a}
	bg.StrokeWidth = 1

	// Logo badge: blue circle
	logoBg := canvas.NewRectangle(color.NRGBA{R: 0x0a, G: 0x84, B: 0xff, A: 0xff})
	logoBg.CornerRadius = 12
	logoText := canvas.NewText("⬇", color.White)
	logoText.TextSize = 18
	logoText.TextStyle = fyne.TextStyle{Bold: true}
	logo := container.NewStack(
		container.New(&fixedSizeLayout{w: 36, h: 36}, logoBg),
		container.NewCenter(logoText),
	)

	// Nav buttons — icon only
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

	// Status dot: glow effect with two stacked rectangles
	dotGlow := canvas.NewRectangle(color.NRGBA{R: 0x30, G: 0xd1, B: 0x58, A: 0x55})
	dotGlow.CornerRadius = 6
	dotSolid := canvas.NewRectangle(color.NRGBA{R: 0x30, G: 0xd1, B: 0x58, A: 0xff})
	dotSolid.CornerRadius = 4
	statusDotCanvas := container.NewStack(
		container.New(&fixedSizeLayout{w: 12, h: 12}, dotGlow),
		container.New(&fixedSizeLayout{w: 8, h: 8}, dotSolid),
	)
	// Keep widget.Label hidden for status ticker compatibility
	mw.statusDot = widget.NewLabel("")
	mw.statusDot.Hide()

	nav := container.NewVBox(
		container.NewCenter(logo),
		widget.NewSeparator(),
		container.NewCenter(mw.navBtns[screenDownloads]),
		container.NewCenter(mw.navBtns[screenAddTorrent]),
		widget.NewSeparator(),
		container.NewCenter(mw.navBtns[screenSettings]),
		container.NewCenter(mw.navBtns[screenAbout]),
		layout.NewSpacer(),
		container.NewCenter(statusDotCanvas),
		widget.NewLabel(""), // bottom padding
	)

	return container.NewStack(bg, nav)
}

// buildStatusBar creates the bottom status bar with live aggregate stats.
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

// selectNav highlights the active nav button.
func (mw *MainWindow) selectNav(id screenID) {
	mw.currentID = id
	for i, btn := range mw.navBtns {
		if screenID(i) == id {
			btn.Importance = widget.HighImportance
		} else {
			btn.Importance = widget.LowImportance
		}
		btn.Refresh()
	}
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
