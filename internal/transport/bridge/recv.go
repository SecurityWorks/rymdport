package bridge

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Jacalz/wormhole-gui/v2/internal/transport"
	"github.com/Jacalz/wormhole-gui/v2/internal/util"
)

// RecvItem is the item that is being received
type RecvItem struct {
	URI      fyne.URI
	Progress *util.ProgressBar
	Name     string
}

// RecvList is a list of progress bars that track send progress.
type RecvList struct {
	widget.List

	client *transport.Client

	Items []*RecvItem

	window fyne.Window
}

// Length returns the length of the data.
func (p *RecvList) Length() int {
	return len(p.Items)
}

// CreateItem creates a new item in the list.
func (p *RecvList) CreateItem() fyne.CanvasObject {
	return container.New(&listLayout{},
		widget.NewFileIcon(nil),
		&widget.Label{Text: "Waiting for filename...", Wrapping: fyne.TextTruncate},
		util.NewProgressBar(),
	)
}

// UpdateItem updates the data in the list.
func (p *RecvList) UpdateItem(i int, item fyne.CanvasObject) {
	item.(*fyne.Container).Objects[0].(*widget.FileIcon).SetURI(p.Items[i].URI)
	item.(*fyne.Container).Objects[1].(*widget.Label).SetText(p.Items[i].Name)
	p.Items[i].Progress = item.(*fyne.Container).Objects[2].(*util.ProgressBar)
}

// RemoveItem removes the item at the specified index.
func (p *RecvList) RemoveItem(i int) {
	copy(p.Items[i:], p.Items[i+1:])
	p.Items[p.Length()-1] = nil // Make sure that GC run on removed element
	p.Items = p.Items[:p.Length()-1]
	p.Refresh()
}

// OnSelected handles removing items and stopping send (in the future)
func (p *RecvList) OnSelected(i int) {
	dialog.ShowConfirm("Remove from list", "Do you wish to remove the item from the list?", func(remove bool) {
		if remove {
			p.RemoveItem(i)
			p.Refresh()
		}
	}, p.window)

	p.Unselect(i)
}

// NewReceive adds data about a new send to the list and then returns the channel to update the code.
func (p *RecvList) NewReceive(code string) {
	p.Items = append(p.Items, &RecvItem{Name: "Waiting for filename..."})
	p.Refresh()

	path := make(chan string)
	index := p.Length() - 1

	go func() {
		name := <-path
		p.Items[index].URI = storage.NewFileURI(name)
		if name != "text" {
			p.Items[index].Name = p.Items[index].URI.Name()
		} else {
			p.Items[index].Name = "Text Snippet"
		}

		close(path)
		p.Refresh()
	}()

	go func(code string) {
		if err := p.client.NewReceive(code, path, p.Items[index].Progress); err != nil {
			p.client.ShowNotification("Receive failed", "An error occurred when receiving the data.")
			p.Items[index].Progress.Failed()
			dialog.ShowError(err, p.window)
		} else {
			p.client.ShowNotification("Receive completed", "The data was received successfully.")
		}

		p.Refresh()
	}(code)
}

// NewRecvList greates a list of progress bars.
func NewRecvList(window fyne.Window, client *transport.Client) *RecvList {
	p := &RecvList{client: client, window: window}
	p.List.Length = p.Length
	p.List.CreateItem = p.CreateItem
	p.List.UpdateItem = p.UpdateItem
	p.List.OnSelected = p.OnSelected
	p.ExtendBaseWidget(p)

	return p
}
