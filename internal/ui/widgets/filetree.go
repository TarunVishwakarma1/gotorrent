package widgets

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
)

// FileTreeItem represents one row in the file selection list.
type FileTreeItem struct {
	State    *engine.FileState
	OnToggle func(selected bool)
}

// FileTree displays a list of files from a torrent with checkboxes.
// For single-file torrents it shows one row; for multi-file it shows all files.
type FileTree struct {
	widget.BaseWidget

	items  []*FileTreeItem
	rows   []*fileTreeRow
	scroll *container.Scroll

	// OnSelectionChanged is called whenever any checkbox changes.
	OnSelectionChanged func()
}

// NewFileTree constructs a FileTree for the given file states.
func NewFileTree(files []engine.FileState, onChanged func()) *FileTree {
	ft := &FileTree{OnSelectionChanged: onChanged}
	ft.setFiles(files)
	ft.ExtendBaseWidget(ft)
	return ft
}

// SetFiles replaces the displayed file list.
func (ft *FileTree) SetFiles(files []engine.FileState) {
	ft.setFiles(files)
	ft.Refresh()
}

// SelectedFiles returns copies of selected FileState entries.
func (ft *FileTree) SelectedFiles() []engine.FileState {
	out := make([]engine.FileState, 0, len(ft.items))
	for _, item := range ft.items {
		if item.State.Selected {
			out = append(out, *item.State)
		}
	}
	return out
}

// SelectedSize returns total bytes of selected files.
func (ft *FileTree) SelectedSize() int64 {
	var total int64
	for _, item := range ft.items {
		if item.State.Selected {
			total += item.State.Length
		}
	}
	return total
}

// TotalSize returns total bytes of all files.
func (ft *FileTree) TotalSize() int64 {
	var total int64
	for _, item := range ft.items {
		total += item.State.Length
	}
	return total
}

// SelectAll selects or deselects all files.
func (ft *FileTree) SelectAll(selected bool) {
	for _, item := range ft.items {
		item.State.Selected = selected
	}
	for _, row := range ft.rows {
		row.check.SetChecked(selected)
	}
	if ft.OnSelectionChanged != nil {
		ft.OnSelectionChanged()
	}
}

func (ft *FileTree) setFiles(files []engine.FileState) {
	ft.items = make([]*FileTreeItem, len(files))
	ft.rows = make([]*fileTreeRow, len(files))

	for i := range files {
		i := i
		fs := &files[i]
		item := &FileTreeItem{State: fs}
		item.OnToggle = func(selected bool) {
			fs.Selected = selected
			if ft.OnSelectionChanged != nil {
				ft.OnSelectionChanged()
			}
		}
		ft.items[i] = item
		ft.rows[i] = newFileTreeRow(item)
	}
}

// CreateRenderer implements fyne.Widget.
func (ft *FileTree) CreateRenderer() fyne.WidgetRenderer {
	rows := make([]fyne.CanvasObject, len(ft.rows))
	for i, r := range ft.rows {
		rows[i] = r
	}
	list := container.NewVBox(rows...)
	ft.scroll = container.NewScroll(list)
	ft.scroll.SetMinSize(fyne.NewSize(400, 200))
	return widget.NewSimpleRenderer(ft.scroll)
}

// fileTreeRow is one row in the file tree: checkbox + icon + name + size.
type fileTreeRow struct {
	widget.BaseWidget
	item    *FileTreeItem
	check   *widget.Check
	content fyne.CanvasObject
}

func newFileTreeRow(item *FileTreeItem) *fileTreeRow {
	r := &fileTreeRow{item: item}

	name := ""
	if len(item.State.Path) > 0 {
		name = filepath.Join(item.State.Path...)
	}

	icon := engine.FileIconForName(filepath.Base(name))
	label := fmt.Sprintf("%s %s", icon, name)
	size := formatBytes(float64(item.State.Length))

	r.check = widget.NewCheck(label, func(checked bool) {
		if item.OnToggle != nil {
			item.OnToggle(checked)
		}
	})
	r.check.SetChecked(item.State.Selected)

	sizeLabel := widget.NewLabel(size)
	sizeLabel.Alignment = fyne.TextAlignTrailing

	r.content = container.NewBorder(nil, nil, nil, sizeLabel, r.check)
	r.ExtendBaseWidget(r)
	return r
}

func (r *fileTreeRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}
