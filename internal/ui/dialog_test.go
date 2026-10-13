package ui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"lazyrun/internal/db"
	"lazyrun/internal/icon"
	"lazyrun/internal/store"
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
	ui.showEditDialog(store.Shortcut{Slug: "winbox", Name: "Winbox", Type: store.TypeExecutable, Kind: "windows",
		BinaryPath: "/tmp/winbox", IconPath: "", Categories: "Network;", Targets: store.Targets{Menu: true, Desktop: true}})
	if top := ui.win.Canvas().Overlays().Top(); top == nil {
		t.Fatal("modal edit executable tidak muncul")
	}
	ui.showEditDialog(store.Shortcut{Slug: "wa", Name: "WhatsApp Web", Type: store.TypeWebApp, Kind: "pwa",
		URL: "https://web.whatsapp.com", Browser: "/usr/bin/vivaldi", Targets: store.Targets{Menu: true}})
	if top := ui.win.Canvas().Overlays().Top(); top == nil {
		t.Fatal("modal edit webapp tidak muncul")
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

// TestRestoreImported memasang ulang entri cadangan: berkas dibuat di HOME saat ini
// dan entri yang sudah ada tidak ditimpa.
func TestRestoreImported(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("PATH", "")

	script := []byte("#!/bin/sh\necho halo\n")
	items := []db.Backup{
		{Item: store.Shortcut{Slug: "lama", Name: "Deploy", Type: store.TypeExecutable, Kind: "script",
			BinaryPath: "/komputer-lama/.local/bin/deploy", Targets: store.Targets{Menu: true}}, Binary: script},
		{Item: store.Shortcut{Name: "Figma", Type: store.TypeWebApp, Kind: "pwa",
			URL: "https://figma.com", Browser: "/usr/bin/vivaldi", Targets: store.Targets{Menu: true}}},
	}

	a := test.NewApp()
	defer a.Quit()
	ui, err := build(a, "test")
	if err != nil {
		t.Fatal(err)
	}
	ui.store.Items = nil
	ui.store.Add(store.Shortcut{Slug: "figma", Name: "Figma"}) // sudah ada -> dilewati

	added, skipped := ui.restoreImported(items)
	if added != 1 || skipped != 1 {
		t.Fatalf("added=%d skipped=%d, mau 1 dan 1", added, skipped)
	}

	deploy, ok := ui.store.Get("deploy")
	if !ok {
		t.Fatal("shortcut hasil restore tidak masuk daftar")
	}
	sc := ui.store.Items[deploy]
	want := filepath.Join(home, ".local", "bin", "deploy")
	if sc.BinaryPath != want {
		t.Errorf("binary = %s, mau %s", sc.BinaryPath, want)
	}
	if _, err := os.Stat(want + ".real"); err != nil {
		t.Errorf("script asli (.real) tidak ditulis: %v", err)
	}
	if _, err := os.Stat(sc.DesktopPath); err != nil {
		t.Errorf(".desktop tidak dibuat: %v", err)
	}
	if sc.IconPath == "" {
		t.Error("ikon tidak dipasang")
	}
	if sc.Kind != "script" {
		t.Errorf("kind hasil restore = %q", sc.Kind)
	}
}
