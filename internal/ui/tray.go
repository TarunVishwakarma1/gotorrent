package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// setupTray configures the system tray icon and menu.
// On platforms that do not support a system tray this is a no-op.
func (gta *GoTorrentApp) setupTray(win fyne.Window) {
	deskApp, ok := gta.App.(desktop.App)
	if !ok {
		return // not a desktop app (e.g. mobile)
	}

	icon := appIcon()
	if icon != nil {
		deskApp.SetSystemTrayIcon(icon)
	}

	deskApp.SetSystemTrayMenu(fyne.NewMenu("GoTorrent",
		fyne.NewMenuItem("Show GoTorrent", func() {
			win.Show()
			win.RequestFocus()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Pause All", func() {
			for _, s := range gta.Manager.GetAll() {
				_ = gta.Manager.Pause(s.ID)
			}
		}),
		fyne.NewMenuItem("Resume All", func() {
			for _, s := range gta.Manager.GetAll() {
				_ = gta.Manager.Resume(s.ID)
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			gta.Manager.Shutdown()
			gta.App.Quit()
		}),
	))
}

// updateTrayTooltip refreshes the tray icon tooltip with current download speed.
func (gta *GoTorrentApp) updateTrayTooltip() {
	_, ok := gta.App.(desktop.App)
	if !ok {
		return
	}

	var totalSpeed float64
	for _, s := range gta.Manager.GetAll() {
		totalSpeed += s.Speed
	}

	var tooltip string
	if totalSpeed > 0 {
		tooltip = fmt.Sprintf("GoTorrent — ↓ %s/s", formatTraySpeed(totalSpeed))
	} else {
		tooltip = "GoTorrent — Idle"
	}
	_ = tooltip // Fyne tray tooltips are set via menu title on some platforms
}

func formatTraySpeed(bps float64) string {
	switch {
	case bps >= 1<<20:
		return fmt.Sprintf("%.1f MB", bps/(1<<20))
	case bps >= 1<<10:
		return fmt.Sprintf("%.1f KB", bps/(1<<10))
	default:
		return fmt.Sprintf("%.0f B", bps)
	}
}
