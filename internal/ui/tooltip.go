package ui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Keterangan tombol muncul di satu baris hint (bawah window atau dalam dialog),
// bukan sebagai popup: popup Fyne selalu jadi overlay, dan selama overlay ada
// semua klik ke konten di bawahnya tidak diteruskan (tombol jadi perlu 2x klik).
var (
	hintMu     sync.Mutex
	hintTarget *canvas.Text
)

// pushHintTarget mengarahkan keterangan hover ke label tertentu (dipakai dialog).
// Fungsi yang dikembalikan mengembalikan target sebelumnya.
func pushHintTarget(label *canvas.Text) func() {
	hintMu.Lock()
	previous := hintTarget
	hintTarget = label
	hintMu.Unlock()

	return func() {
		hintMu.Lock()
		hintTarget = previous
		hintMu.Unlock()
	}
}

func showHint(text string) {
	hintMu.Lock()
	target := hintTarget
	hintMu.Unlock()

	if target == nil || text == "" {
		return
	}
	target.Text = text
	target.Refresh()
}

// clearHint menghapus hint hanya kalau isinya memang milik tombol ini.
func clearHint(text string) {
	hintMu.Lock()
	target := hintTarget
	hintMu.Unlock()

	if target == nil || target.Text != text {
		return
	}
	target.Text = ""
	target.Refresh()
}

// tipButton adalah tombol yang menampilkan fungsinya saat disorot mouse.
type tipButton struct {
	widget.Button
	tip string
}

// newButton membuat tombol (dengan ikon kalau diberikan) plus keterangan hover.
func newButton(text string, icon fyne.Resource, tip string, importance widget.Importance, tapped func()) *tipButton {
	button := &tipButton{tip: tip}
	button.Text = text
	button.Icon = icon
	button.Importance = importance
	button.OnTapped = tapped
	button.ExtendBaseWidget(button)
	return button
}

// newIconButton adalah tombol ikon saja (tanpa teks), gaya datar.
func newIconButton(icon fyne.Resource, tip string, tapped func()) *tipButton {
	return newButton("", icon, tip, widget.LowImportance, tapped)
}

func (b *tipButton) MouseIn(event *desktop.MouseEvent) {
	b.Button.MouseIn(event)
	showHint(b.tip)
}

func (b *tipButton) MouseOut() {
	b.Button.MouseOut()
	clearHint(b.tip)
}

// showLoading menampilkan dialog proses dengan progress bar yang berjalan,
// dan mengembalikan fungsi untuk menutupnya.
func (ui *app) showLoading(title, message string) func() {
	bar := widget.NewProgressBarInfinite()

	var d *dialog.CustomDialog
	d = dialog.NewCustomWithoutButtons(title, container.NewBorder(
		container.NewVBox(logoBox(newLogo(), 22), muted(message)),
		nil, nil, nil,
		bar,
	), ui.win)
	showDialog(d, ui.win, 420, 170)
	return d.Hide
}
