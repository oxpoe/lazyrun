package ui

import (
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lazyrun/internal/icon"
)

// iconCache menyimpan ikon yang sudah diunduh supaya pencarian ulang tidak
// mengunduh berkali-kali: API ikon publik membatasi jumlah permintaan beruntun.
type iconCache struct {
	mu    sync.Mutex
	items map[string][]byte
}

func newIconCache() *iconCache { return &iconCache{items: map[string][]byte{}} }

func (c *iconCache) load(url string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, ok := c.items[url]
	return data, ok
}

func (c *iconCache) store(url string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[url] = data
}

// fetchIcon mengunduh ikon (atau mengambil dari cache).
func (ui *app) fetchIcon(url string) ([]byte, error) {
	if data, ok := ui.icons.load(url); ok {
		return data, nil
	}
	data, _, err := icon.FromURL(url)
	if err != nil {
		return nil, err
	}
	ui.icons.store(url, data)
	return data, nil
}

// showIconSearch menampilkan pencarian ikon online (Iconify) dan mengirim hasil pilihan.
func (ui *app) showIconSearch(query string, onPick func(icon.Result, []byte)) {
	entry := widget.NewEntry()
	entry.SetText(icon.QueryHint(query))
	entry.SetPlaceHolder("whatsapp")

	status := muted(T("search_hint"))
	entry.SetPlaceHolder("whatsapp / mikrotik.com")
	grid := container.NewGridWithColumns(6)
	results := container.NewVScroll(container.NewPadded(grid))
	slots := make(chan struct{}, 3)

	var d *dialog.CustomDialog
	search := func() {
		if strings.TrimSpace(entry.Text) == "" {
			status.SetText(T("search_hint"))
			return
		}
		status.SetText(T("searching"))
		grid.RemoveAll()
		go func() {
			list, err := lookupIcons(entry.Text)
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
	// status di barisnya sendiri: supaya tidak bertabrakan dengan tombol
	// di layar sempit atau ber-penskalaan besar.
	bottom := container.NewVBox(status, container.NewHBox(
		spacer(),
		widget.NewButtonWithIcon(T("back"), theme.NavigateBackIcon(), func() { d.Hide() }),
		widget.NewButtonWithIcon(T("search"), theme.SearchIcon(), search),
	))
	content := container.NewBorder(
		container.NewVBox(closeRow(func() { d.Hide() }), entry),
		bottom, nil, nil,
		results,
	)

	d = dialog.NewCustomWithoutButtons(T("icon_search_title"), content, ui.win)
	showDialog(d, ui.win, 560, 470) // sekaligus menampilkan dialog
	search()
}

// lookupIcons memilih sumber sesuai yang diketik user:
// URL gambar → dipakai langsung, domain situs → ikon resmi situs itu,
// selain itu → pencarian nama di Iconify.
func lookupIcons(query string) ([]icon.Result, error) {
	query = strings.TrimSpace(query)
	if icon.IsImageURL(query) {
		return []icon.Result{{Name: "URL ikon", URL: query}}, nil
	}
	if domain := icon.Domain(query); domain != "" {
		if list, err := icon.SiteIcons(domain); err == nil && len(list) > 0 {
			return list, nil
		}
	}
	return icon.SearchBest(query, 24)
}

// iconChoice adalah satu ikon hasil pencarian. Gambarnya diunduh di latar belakang
// dan digambar sebagai PNG: SVG dari internet kadang tidak bisa digambar sehingga
// thumbnail tidak muncul atau proses menggambar berhenti.
func (ui *app) iconChoice(result icon.Result, onPick func(icon.Result, []byte), d *dialog.CustomDialog, slot chan struct{}) fyne.CanvasObject {
	var mu sync.Mutex
	var raw []byte

	button := widget.NewButton("", func() {
		mu.Lock()
		picked := raw
		mu.Unlock()
		onPick(result, picked)
		if d != nil {
			d.Hide()
		}
	})
	button.Importance = widget.LowImportance
	button.Icon = theme.NewThemedResource(theme.FileImageIcon()) // placeholder selama diunduh

	go func() {
		slot <- struct{}{}
		defer func() { <-slot }()

		downloaded, err := ui.fetchIcon(result.URL)
		if err != nil {
			return
		}
		thumb, err := icon.ToPNG(downloaded)
		if err != nil {
			return // placeholder tetap tampil
		}

		mu.Lock()
		raw = downloaded
		mu.Unlock()

		fyne.Do(func() {
			button.Icon = fyne.NewStaticResource(result.Name+".png", thumb)
			button.Refresh()
		})
	}()

	return container.NewGridWrap(fyne.NewSize(52, 52), button)
}
