package ui

import (
	"fmt"
	"net/url"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
	"github.com/tarunvishwakarma1/gotorrent/internal/ui/widgets"
)

// downloadsScreen manages the downloads list view.
type downloadsScreen struct {
	mu     sync.Mutex
	states []*engine.TorrentState
	list   *widget.List
	empty  fyne.CanvasObject
	stack  *fyne.Container
	gta    *GoTorrentApp
	win    fyne.Window
}

// newDownloadsScreen constructs the downloads list screen.
func newDownloadsScreen(gta *GoTorrentApp, win fyne.Window) (*downloadsScreen, fyne.CanvasObject) {
	ds := &downloadsScreen{gta: gta, win: win}

	// Empty state shown when no torrents are loaded.
	ds.empty = buildEmptyState(func() {
		openTorrentFilePicker(gta, win)
	})

	ds.list = widget.NewList(
		func() int {
			ds.mu.Lock()
			defer ds.mu.Unlock()
			return len(ds.states)
		},
		func() fyne.CanvasObject {
			return widgets.NewProgressRow(
				func(id string) { _ = gta.Manager.Pause(id) },
				func(id string) { _ = gta.Manager.Resume(id) },
				func(id string) { ds.confirmRemove(id) },
				func(id string) { ds.openDownload(id) },
			)
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			ds.mu.Lock()
			if i >= len(ds.states) {
				ds.mu.Unlock()
				return
			}
			s := ds.states[i]
			ds.mu.Unlock()
			obj.(*widgets.ProgressRow).Update(s)
		},
	)

	ds.stack = container.NewStack(ds.empty)

	// Load initial state from manager.
	initial := gta.Manager.GetAll()
	ds.updateStates(initial)

	// Subscribe to manager updates.
	gta.Manager.OnUpdate = func(state *engine.TorrentState) {
		ds.onStateUpdate(state)
	}
	gta.Manager.OnComplete = func(state *engine.TorrentState) {
		if gta.Config.Get().NotifyOnComplete {
			gta.App.SendNotification(&fyne.Notification{
				Title:   "GoTorrent",
				Content: state.Name + " — Download complete",
			})
		}
	}
	gta.Manager.OnError = func(state *engine.TorrentState) {
		gta.App.SendNotification(&fyne.Notification{
			Title:   "GoTorrent — Error",
			Content: state.Name + " — Download failed: " + state.Error,
		})
	}

	return ds, ds.stack
}

// onStateUpdate merges a state update and refreshes the list.
func (ds *downloadsScreen) onStateUpdate(state *engine.TorrentState) {
	ds.mu.Lock()
	found := false
	for i, s := range ds.states {
		if s.ID == state.ID {
			ds.states[i] = state
			found = true
			break
		}
	}
	if !found {
		ds.states = append(ds.states, state)
	}
	ds.mu.Unlock()
	ds.refreshView()
}

// updateStates replaces the state list wholesale.
func (ds *downloadsScreen) updateStates(states []*engine.TorrentState) {
	ds.mu.Lock()
	ds.states = states
	ds.mu.Unlock()
	ds.refreshView()
}

// refreshView updates the stack to show either the list or the empty state.
func (ds *downloadsScreen) refreshView() {
	ds.mu.Lock()
	n := len(ds.states)
	ds.mu.Unlock()

	if n == 0 {
		ds.stack.Objects = []fyne.CanvasObject{ds.empty}
	} else {
		ds.stack.Objects = []fyne.CanvasObject{ds.list}
	}
	ds.stack.Refresh()
	ds.list.Refresh()
}

// confirmRemove shows a confirmation dialog before removing a torrent.
func (ds *downloadsScreen) confirmRemove(id string) {
	state, err := ds.gta.Manager.Get(id)
	if err != nil {
		return
	}

	deleteCheck := widget.NewCheck("Delete downloaded files", nil)

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Remove %q from GoTorrent?", state.Name)),
		deleteCheck,
	)

	dialog.ShowCustomConfirm("Remove Torrent", "Remove", "Cancel", content, func(confirm bool) {
		if !confirm {
			return
		}
		if err := ds.gta.Manager.Remove(id, deleteCheck.Checked); err != nil {
			dialog.ShowError(err, ds.win)
			return
		}
		// Remove from local state slice.
		ds.mu.Lock()
		for i, s := range ds.states {
			if s.ID == id {
				ds.states = append(ds.states[:i], ds.states[i+1:]...)
				break
			}
		}
		ds.mu.Unlock()
		ds.refreshView()
	}, ds.win)
}

// openDownload opens the download's output directory in the OS file manager.
func (ds *downloadsScreen) openDownload(id string) {
	state, err := ds.gta.Manager.Get(id)
	if err != nil {
		return
	}
	u, err := url.Parse("file://" + state.SavePath)
	if err != nil {
		return
	}
	_ = ds.gta.App.OpenURL(u)
}

// buildEmptyState returns the "no torrents yet" placeholder.
func buildEmptyState(onAdd func()) fyne.CanvasObject {
	icon := widget.NewLabel("📥")
	icon.Alignment = fyne.TextAlignCenter

	title := widget.NewLabel("No torrents yet")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	sub := widget.NewLabel("Drag a .torrent file here to get started")
	sub.Alignment = fyne.TextAlignCenter

	addBtn := widget.NewButton("+ Add Torrent", onAdd)
	addBtn.Importance = widget.HighImportance

	return container.NewCenter(container.NewVBox(
		icon, title, sub,
		widget.NewLabel(""),
		container.NewCenter(addBtn),
	))
}

// formatBytes returns human-readable file size (ui package copy).
func formatBytes(b float64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", b/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", b/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", b/(1<<10))
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}
