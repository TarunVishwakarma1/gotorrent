// Package ui implements the GoTorrent desktop UI using Fyne v2.
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"github.com/tarunvishwakarma1/gotorrent/internal/config"
	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
)

// GoTorrentApp is the root application object.
type GoTorrentApp struct {
	App     fyne.App
	Manager *engine.TorrentManager
	Config  *config.Manager
	window  *MainWindow
}

// New creates and initialises the GoTorrent application.
func New(manager *engine.TorrentManager, cfg *config.Manager) *GoTorrentApp {
	a := app.NewWithID("com.gotorrent.app")
	a.SetIcon(appIcon())

	gta := &GoTorrentApp{
		App:     a,
		Manager: manager,
		Config:  cfg,
	}

	gta.applyTheme(cfg.Get().Theme)
	return gta
}

// ShowAndRun opens the main window and blocks until the app exits.
func (gta *GoTorrentApp) ShowAndRun(initialTorrent string) {
	gta.window = newMainWindow(gta)

	if gta.Config.Get().StartMinimized {
		gta.window.win.Hide()
	} else {
		gta.window.win.Show()
	}

	if initialTorrent != "" {
		gta.window.OpenTorrentFile(initialTorrent)
	}

	gta.App.Run()
}

// OpenTorrentFile opens the preview dialog for a torrent path.
// Safe to call from any goroutine.
func (gta *GoTorrentApp) OpenTorrentFile(path string) {
	if gta.window != nil {
		gta.window.OpenTorrentFile(path)
	}
}

// applyTheme sets the Fyne theme based on the user's preference.
func (gta *GoTorrentApp) applyTheme(choice config.ThemeChoice) {
	switch choice {
	case config.ThemeDark:
		gta.App.Settings().SetTheme(&goTorrentTheme{variant: theme.VariantDark, forced: true})
	case config.ThemeLight:
		gta.App.Settings().SetTheme(&goTorrentTheme{variant: theme.VariantLight, forced: true})
	default: // system
		gta.App.Settings().SetTheme(&goTorrentTheme{forced: false})
	}
}

// ApplyTheme is called from the settings screen when the user changes theme.
func (gta *GoTorrentApp) ApplyTheme(choice config.ThemeChoice) {
	gta.applyTheme(choice)
}

// goTorrentTheme is a custom Fyne theme with GoTorrent's brand colours.
type goTorrentTheme struct {
	variant fyne.ThemeVariant
	forced  bool
}

func (t *goTorrentTheme) resolve(v fyne.ThemeVariant) fyne.ThemeVariant {
	if t.forced {
		return t.variant
	}
	return v
}

func (t *goTorrentTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	v := t.resolve(variant)
	if v == theme.VariantDark {
		return t.darkColor(name)
	}
	return t.lightColor(name)
}

func (t *goTorrentTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *goTorrentTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *goTorrentTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 8
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func (t *goTorrentTheme) darkColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground, theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 0x0e, G: 0x16, B: 0x24, A: 0xff} // deep dark blue-grey
	case theme.ColorNameButton:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x0f}
	case theme.ColorNamePrimary, theme.ColorNameFocus:
		return color.NRGBA{R: 0x00, G: 0xd4, B: 0xff, A: 0xff} // vibrant cyan
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xc9, G: 0xd6, B: 0xe3, A: 0xff} // soft light blue-grey text
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x0f}
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x11, G: 0x18, B: 0x27, A: 0xff}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x00, G: 0xff, B: 0x88, A: 0xff} // vibrant green
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xff, G: 0x9f, B: 0x0a, A: 0xff}
	case theme.ColorNameError:
		return color.NRGBA{R: 0xff, G: 0x45, B: 0x3a, A: 0xff}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x0a, G: 0x84, B: 0xff, A: 0x40}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 150}
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (t *goTorrentTheme) lightColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return hexColor(0xf5f5f5)
	case theme.ColorNameButton:
		return hexColor(0x1976d2)
	case theme.ColorNamePrimary:
		return hexColor(0xe94560)
	case theme.ColorNameFocus:
		return hexColor(0xe94560)
	case theme.ColorNameForeground:
		return hexColor(0x212121)
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return hexColor(0x757575)
	case theme.ColorNameInputBackground:
		return hexColor(0xffffff)
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return hexColor(0xffffff)
	case theme.ColorNameHeaderBackground:
		return hexColor(0x1976d2)
	case theme.ColorNameSuccess:
		return hexColor(0x4caf50)
	case theme.ColorNameWarning:
		return hexColor(0xff9800)
	case theme.ColorNameError:
		return hexColor(0xf44336)
	case theme.ColorNameSelection:
		return hexColor(0x1976d2)
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 40}
	}
	return theme.DefaultTheme().Color(name, theme.VariantLight)
}

// hexColor converts a 24-bit RGB value to color.Color.
func hexColor(rgb uint32) color.Color {
	r := uint8((rgb >> 16) & 0xff)
	g := uint8((rgb >> 8) & 0xff)
	b := uint8(rgb & 0xff)
	return color.NRGBA{R: r, G: g, B: b, A: 0xff}
}

// appIcon returns the embedded app icon resource.
func appIcon() fyne.Resource {
	// The icon is loaded from the assets directory via fyne bundle or embed.
	// Falls back to nil (Fyne uses a default icon) if not available.
	return nil
}
