package ui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"lazyrun/internal/icon"
)

// TestNoEmptyDismissButton memastikan modal tidak memakai dialog.NewCustom,
// yang selalu menambah tombol dismiss walau labelnya kosong (tombol hantu
// di bagian bawah modal). Kita punya baris tombol sendiri, jadi harus memakai
// NewCustomWithoutButtons.
func TestNoEmptyDismissButton(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "dialog.NewCustom(") {
			t.Errorf("%s memakai dialog.NewCustom (menambah tombol kosong); gunakan NewCustomWithoutButtons", file)
		}
	}
}

// TestDialogsOpenWithoutPanic membuka semua modal di test driver.
func TestDialogsOpenWithoutPanic(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	// pencarian ikon diarahkan ke port mati: tes tidak boleh menyentuh internet
	// dan tidak boleh mengubah UI dari goroutine latar.
	icon.SetSearchEndpoint("http://127.0.0.1:1")
	defer icon.SetSearchEndpoint("https://api.iconify.design")

	ui, err := build(a, "test")
	if err != nil {
		t.Fatal(err)
	}
	ui.win.Show()

	ui.showExecutableDialog()
	if top := ui.win.Canvas().Overlays().Top(); top == nil {
		t.Fatal("modal executable tidak muncul")
	}
	ui.showWebAppDialog()
	if len(ui.browsers) > 0 {
		if top := ui.win.Canvas().Overlays().Top(); top == nil {
			t.Fatal("modal web app tidak muncul")
		}
	}
	ui.showIconSearch("whatsapp", func(icon.Result, []byte) {})
	if top := ui.win.Canvas().Overlays().Top(); top == nil {
		t.Fatal("modal pencarian ikon tidak muncul")
	}

	// tunggu goroutine pencarian selesai: test driver menjalankan fyne.Do inline,
	// jadi sisa goroutine bisa bertabrakan dengan tes berikutnya kalau dibiarkan.
	time.Sleep(300 * time.Millisecond)
}

// TestStringsCoverUsedKeys memastikan setiap kunci T() yang dipakai ada di tabel
// terjemahan (kalau tidak, UI menampilkan nama kunci mentah).
func TestStringsCoverUsedKeys(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	used := map[string]bool{}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range regexp.MustCompile(`T\("([a-z_]+)"\)`).FindAllStringSubmatch(string(data), -1) {
			used[m[1]] = true
		}
	}
	if len(used) == 0 {
		t.Fatal("tidak ada kunci T() yang ditemukan")
	}
	for key := range used {
		if _, ok := tr[key]; !ok {
			t.Errorf("kunci %q dipakai tapi tidak ada di tabel terjemahan", key)
		}
	}

	// dan sebaliknya: kunci terjemahan yang tidak terpakai sebaiknya dihapus
	for key := range tr {
		if !used[key] {
			t.Errorf("kunci %q ada di tabel terjemahan tapi tidak dipakai", key)
		}
	}
}
