package ui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const appVersion = "1.0.0"
const githubURL = "https://github.com/tarunvishwakarma1/gotorrent"

// aboutScreen builds the About screen content.
func aboutScreen(gta *GoTorrentApp) fyne.CanvasObject {
	icon := widget.NewLabel("📥")
	icon.Alignment = fyne.TextAlignCenter

	appName := widget.NewLabel("GoTorrent")
	appName.Alignment = fyne.TextAlignCenter
	appName.TextStyle = fyne.TextStyle{Bold: true}

	version := widget.NewLabel("Version " + appVersion)
	version.Alignment = fyne.TextAlignCenter

	desc1 := widget.NewLabel("A fast, lightweight BitTorrent client")
	desc1.Alignment = fyne.TextAlignCenter

	desc2 := widget.NewLabel("built with Go and Fyne.")
	desc2.Alignment = fyne.TextAlignCenter

	desc3 := widget.NewLabel("Built on top of a custom BitTorrent engine")
	desc3.Alignment = fyne.TextAlignCenter

	desc4 := widget.NewLabel("written from scratch in pure Go.")
	desc4.Alignment = fyne.TextAlignCenter

	githubBtn := widget.NewButton("View on GitHub", func() {
		u, _ := url.Parse(githubURL)
		_ = gta.App.OpenURL(u)
	})
	githubBtn.Importance = widget.HighImportance

	copyright := widget.NewLabel("© 2025 Tarun Vishwakarma")
	copyright.Alignment = fyne.TextAlignCenter

	license := widget.NewLabel("MIT License")
	license.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		widget.NewSeparator(),
		icon,
		appName,
		version,
		widget.NewSeparator(),
		desc1,
		desc2,
		widget.NewLabel(""),
		desc3,
		desc4,
		widget.NewLabel(""),
		container.NewCenter(githubBtn),
		widget.NewLabel(""),
		copyright,
		license,
	)

	return container.NewScroll(content)
}
