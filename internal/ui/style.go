package ui

import (
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lazyrun/assets"
)

// catatan: widget.Label di Fyne 2.8 selalu memesan tinggi ~2x ukuran font,
// jadi teks satu baris di UI utama memakai canvas.Text supaya rapat.

// textLine membuat teks satu baris berukuran & berwarna tertentu.
func textLine(text string, size float32, bold bool, col fyne.ThemeColorName) *canvas.Text {
	t := canvas.NewText(text, theme.Color(col))
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: bold}
	return t
}

// heading adalah nama aplikasi di header.
func heading(text string) *canvas.Text {
	return textLine(text, 22, true, theme.ColorNameForeground)
}

// subheading adalah nama shortcut di kartu.
func subheading(text string) *canvas.Text {
	return textLine(text, 15, true, theme.ColorNameForeground)
}

// caption adalah teks sekunder kecil (tagline, path, versi).
func caption(text string) *canvas.Text {
	return textLine(text, 12, false, colorNameMuted)
}

// sectionLabel adalah judul bagian huruf kapital.
func sectionLabel(text string) *canvas.Text {
	return textLine(strings.ToUpper(text), 11, true, colorNameMuted)
}

// muted adalah teks sekunder yang bisa wrap/truncate (dipakai di dialog).
func muted(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Importance = widget.LowImportance
	return label
}

// card membungkus konten dengan latar membulat.
func card(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(colorNameCard))
	bg.CornerRadius = 12
	return container.NewStack(bg, container.NewPadded(content))
}

// chip adalah label kecil dengan latar berwarna.
func chip(text string, bg fyne.ThemeColorName) fyne.CanvasObject {
	object, _ := chipLabel(text, bg)
	return object
}

// chipLabel mengembalikan chip sekaligus teksnya supaya bisa diperbarui.
func chipLabel(text string, bg fyne.ThemeColorName) (fyne.CanvasObject, *canvas.Text) {
	rect := canvas.NewRectangle(theme.Color(bg))
	rect.CornerRadius = 7
	label := textLine(text, 11, true, theme.ColorNameForeground)
	label.Alignment = fyne.TextAlignCenter
	return container.NewStack(rect, container.NewPadded(label)), label
}

// accentBar adalah garis gradien brand di bawah header.
func accentBar() *canvas.LinearGradient {
	bar := canvas.NewLinearGradient(brandIndigo, brandCyan, 90)
	bar.SetMinSize(fyne.NewSize(0, 3))
	return bar
}

// iconButton adalah tombol ikon tanpa label, gaya datar.
func iconButton(icon fyne.Resource, tapped func()) *widget.Button {
	button := widget.NewButtonWithIcon("", icon, tapped)
	button.Importance = widget.LowImportance
	return button
}

// thumb menampilkan ikon shortcut, atau fallback kalau file ikonnya hilang.
func thumb(path string, fallback fyne.Resource) fyne.CanvasObject {
	var image *canvas.Image
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			image = canvas.NewImageFromURI(storage.NewFileURI(path))
		}
	}
	if image == nil {
		image = canvas.NewImageFromResource(theme.NewThemedResource(fallback))
	}
	image.FillMode = canvas.ImageFillContain
	image.CornerRadius = 8
	image.SetMinSize(fyne.NewSize(36, 36))
	return container.NewGridWrap(fyne.NewSize(36, 36), image)
}

// newLogo mengembalikan ikon LazyRun (dipakai untuk keperluan animasi).
func newLogo() *canvas.Image {
	image := canvas.NewImageFromResource(fyne.NewStaticResource("lazyrun.png", assets.Icon))
	image.FillMode = canvas.ImageFillContain
	return image
}

// logoBox menampilkan ikon LazyRun pada ukuran tetap.
func logoBox(image *canvas.Image, size float32) fyne.CanvasObject {
	return container.NewGridWrap(fyne.NewSize(size, size), image)
}

// spacer adalah pengisi ruang kosong.
func spacer() fyne.CanvasObject { return layout.NewSpacer() }
