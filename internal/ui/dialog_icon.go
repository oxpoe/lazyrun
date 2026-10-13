package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lazyrun/internal/icon"
)

// showIconSearch menampilkan pencarian ikon online (Iconify) dan mengirim hasil pilihan.
func (ui *app) showIconSearch(query string, onPick func(icon.Result, []byte)) {
	entry := widget.NewEntry()
	entry.SetText(icon.QueryHint(query))
	entry.SetPlaceHolder("whatsapp")

	status := muted(T("search_hint"))
	slots := make(chan struct{}, 4)
	grid := container.NewGridWithColumns(6)
	results := container.NewVScroll(container.NewPadded(grid))

	var d *dialog.CustomDialog
	search := func() {
		if strings.TrimSpace(entry.Text) == "" {
			status.SetText(T("search_hint"))
			return
		}
		status.SetText(T("searching"))
		grid.RemoveAll()
		go func() {
			list, err := icon.SearchBest(entry.Text, 24)
			fyne.Do(func() {
				if err != nil {
					status.SetText(T("search_failed"))
					return
				}
				status.SetText(fmt.Sprintf(T("search_found"), len(list)))
				for _, result := range list {
					grid.Add(ui.iconChoice(result, onPick, d, slots))
				}
			})
		}()
	}

	entry.OnSubmitted = func(string) { search() }
	content := container.NewBorder(
		entry,
		container.NewBorder(nil, nil, status,
			container.NewHBox(
				widget.NewButton(T("cancel"), func() { d.Hide() }),
				widget.NewButtonWithIcon(T("search"), theme.SearchIcon(), search),
			)),
		nil, nil,
		results,
	)

	d = dialog.NewCustom(T("icon_search_title"), "", content, ui.win)
	d.Resize(fyne.NewSize(560, 460))
	d.Show()
	search()
}

// iconChoice adalah satu ikon hasil pencarian; gambarnya diunduh di latar belakang.
func (ui *app) iconChoice(result icon.Result, onPick func(icon.Result, []byte), d *dialog.CustomDialog, slot chan struct{}) fyne.CanvasObject {
	var data []byte

	button := widget.NewButton("", func() {
		onPick(result, data)
		if d != nil {
			d.Hide()
		}
	})
	button.Importance = widget.LowImportance

	go func() {
		slot <- struct{}{} // batasi unduhan bersamaan supaya tidak kena rate limit
		defer func() { <-slot }()

		downloaded, _, err := icon.FromURL(result.URL)
		if err != nil {
			return
		}
		data = downloaded
		fyne.Do(func() {
			button.Icon = fyne.NewStaticResource(result.Name, downloaded)
			button.Refresh()
		})
	}()

	return container.NewGridWrap(fyne.NewSize(52, 52), button)
}
