package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func TestTipButtonShowsAndHidesHint(t *testing.T) {
	a := test.NewApp()
	a.Settings().SetTheme(brand{theme.DefaultTheme()})
	defer a.Quit()

	hint := textLine("", 12, false, colorNameMuted)
	button := newIconButton(theme.MediaPlayIcon(), "jalankan shortcut", func() {})
	w := a.NewWindow("tip")
	w.SetContent(container.NewBorder(nil, hint, nil, nil, button))
	w.Resize(fyne.NewSize(300, 200))
	w.Show()
	restore := pushHintTarget(hint)
	defer restore()

	if hint.Text != "" {
		t.Fatal("hint muncul tanpa hover")
	}

	test.MoveMouse(w.Canvas(), fyne.NewPos(20, 20))
	if hint.Text != "jalankan shortcut" {
		t.Errorf("hint saat hover = %q", hint.Text)
	}

	// keluar dari tombol: hint dibersihkan
	test.MoveMouse(w.Canvas(), fyne.NewPos(280, 180))
	if hint.Text != "" {
		t.Errorf("hint masih %q setelah kursor keluar", hint.Text)
	}
}

// klik saat tooltip terbuka harus tetap kena tombolnya (dulu PopUp menelan klik).
// klik saat hint tampil harus tetap kena tombolnya (dulu pakai popup yang
// jadi overlay, sehingga klik pertama selalu tertelan).
func TestTipButtonClickGoesThrough(t *testing.T) {
	a := test.NewApp()
	a.Settings().SetTheme(brand{theme.DefaultTheme()})
	defer a.Quit()

	tapped := false
	hint := textLine("", 12, false, colorNameMuted)
	button := newIconButton(theme.MediaPlayIcon(), "jalankan shortcut", func() { tapped = true })
	w := a.NewWindow("tip")
	w.SetContent(container.NewBorder(nil, hint, nil, nil, button))
	w.Resize(fyne.NewSize(300, 200))
	w.Show()
	restore := pushHintTarget(hint)
	defer restore()

	test.MoveMouse(w.Canvas(), fyne.NewPos(20, 20))
	test.TapCanvas(w.Canvas(), fyne.NewPos(20, 20))
	if !tapped {
		t.Error("klik pertama tidak sampai ke tombol")
	}
	if top := w.Canvas().Overlays().Top(); top != nil {
		t.Error("tidak boleh ada overlay untuk keterangan tombol")
	}
}

func TestTipButtonWithoutTipDoesNothing(t *testing.T) {
	a := test.NewApp()
	a.Settings().SetTheme(brand{theme.DefaultTheme()})
	defer a.Quit()

	hint := textLine("", 12, false, colorNameMuted)
	button := newButton("Tanpa tip", nil, "", widget.MediumImportance, func() {})
	w := a.NewWindow("tip")
	w.SetContent(container.NewBorder(nil, hint, nil, nil, button))
	w.Resize(fyne.NewSize(300, 200))
	w.Show()
	restore := pushHintTarget(hint)
	defer restore()

	test.MoveMouse(w.Canvas(), fyne.NewPos(20, 20))
	if hint.Text != "" {
		t.Error("tombol tanpa tip tidak boleh menulis hint")
	}
}
