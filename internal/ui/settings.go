package ui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/config"
)

// settingsScreen builds the Settings screen content.
func settingsScreen(gta *GoTorrentApp, win fyne.Window) fyne.CanvasObject {
	cfg := gta.Config.Get()

	// --- Downloads section ---
	savePathEntry := widget.NewEntry()
	savePathEntry.SetText(cfg.SavePath)
	savePathEntry.PlaceHolder = "~/Downloads"

	browseBtn := widget.NewButton("📁", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			savePathEntry.SetText(uri.Path())
		}, win)
	})

	savePathRow := container.NewBorder(nil, nil, nil, browseBtn, savePathEntry)

	maxConcEntry := widget.NewEntry()
	maxConcEntry.SetText(strconv.Itoa(cfg.MaxConcurrent))

	maxConnEntry := widget.NewEntry()
	maxConnEntry.SetText(strconv.Itoa(cfg.MaxConnections))

	portEntry := widget.NewEntry()
	portEntry.SetText(strconv.Itoa(cfg.ListenPort))

	// --- Theme section ---
	themeOpts := []string{"Light", "Dark", "System"}
	themeDefault := 2
	switch cfg.Theme {
	case config.ThemeLight:
		themeDefault = 0
	case config.ThemeDark:
		themeDefault = 1
	}
	themeRadio := widget.NewRadioGroup(themeOpts, nil)
	themeRadio.SetSelected(themeOpts[themeDefault])
	themeRadio.Horizontal = true

	// --- Behaviour section ---
	startMinCheck := widget.NewCheck("Start minimized to tray", nil)
	startMinCheck.SetChecked(cfg.StartMinimized)

	notifyCheck := widget.NewCheck("Show notification on download complete", nil)
	notifyCheck.SetChecked(cfg.NotifyOnComplete)

	minToTrayCheck := widget.NewCheck("Minimize to tray on close", nil)
	minToTrayCheck.SetChecked(cfg.MinimizeToTray)

	autoStartCheck := widget.NewCheck("Start downloads automatically", nil)
	autoStartCheck.SetChecked(cfg.AutoStart)

	// --- Save button ---
	saveBtn := widget.NewButton("Save Settings", func() {
		maxConc, _ := strconv.Atoi(maxConcEntry.Text)
		if maxConc < 1 {
			maxConc = 1
		}
		maxConn, _ := strconv.Atoi(maxConnEntry.Text)
		if maxConn < 1 {
			maxConn = 1
		}
		port, _ := strconv.Atoi(portEntry.Text)
		if port < 1 || port > 65535 {
			port = 6881
		}

		var themeChoice config.ThemeChoice
		switch themeRadio.Selected {
		case "Light":
			themeChoice = config.ThemeLight
		case "Dark":
			themeChoice = config.ThemeDark
		default:
			themeChoice = config.ThemeSystem
		}

		newCfg := &config.Config{
			SavePath:         savePathEntry.Text,
			MaxConcurrent:    maxConc,
			MaxConnections:   maxConn,
			ListenPort:       port,
			Theme:            themeChoice,
			StartMinimized:   startMinCheck.Checked,
			NotifyOnComplete: notifyCheck.Checked,
			MinimizeToTray:   minToTrayCheck.Checked,
			AutoStart:        autoStartCheck.Checked,
		}
		if err := gta.Config.Save(newCfg); err != nil {
			dialog.ShowError(err, win)
			return
		}
		gta.Manager.SetMaxConcurrent(maxConc)
		gta.ApplyTheme(themeChoice)
		dialog.ShowInformation("Settings", "Settings saved successfully.", win)
	})
	saveBtn.Importance = widget.HighImportance

	form := container.NewVBox(
		sectionLabel("Downloads"),
		widget.NewSeparator(),
		formRow("Default save location", savePathRow),
		formRow("Max concurrent downloads", maxConcEntry),
		formRow("Max connections per torrent", maxConnEntry),
		widget.NewLabel(""),

		sectionLabel("Network"),
		widget.NewSeparator(),
		formRow("Listening port", portEntry),
		widget.NewLabel(""),

		sectionLabel("Appearance"),
		widget.NewSeparator(),
		formRow("Theme", themeRadio),
		widget.NewLabel(""),

		sectionLabel("Behavior"),
		widget.NewSeparator(),
		startMinCheck,
		notifyCheck,
		minToTrayCheck,
		autoStartCheck,
		widget.NewLabel(""),

		container.NewBorder(nil, nil, nil, saveBtn),
	)

	return container.NewScroll(form)
}

// sectionLabel returns a bold section heading label.
func sectionLabel(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Bold: true}
	return l
}

// formRow returns a two-column form row: label on the left, widget on the right.
func formRow(label string, w fyne.CanvasObject) fyne.CanvasObject {
	l := widget.NewLabel(label)
	return container.NewGridWithColumns(2, l, w)
}
