package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"lazyrun/internal/store"
)

func sampleItems() []store.Shortcut {
	return []store.Shortcut{
		{
			Slug:        "winbox",
			Name:        "Winbox",
			Type:        store.TypeExecutable,
			Kind:        "windows",
			ExecArgs:    []string{"/usr/bin/wine", "/home/u/.local/bin/winbox"},
			Categories:  "Network;",
			BinaryPath:  "/home/u/.local/bin/winbox",
			IconPath:    "/home/u/.local/share/icons/winbox.png",
			IconSource:  "search:winbox",
			DesktopPath: "/home/u/.local/share/applications/winbox.desktop",
			Targets:     store.Targets{Menu: true, Desktop: true},
			InstalledAt: time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC),
		},
		{
			Slug:            "whatsapp-web",
			Name:            "WhatsApp Web",
			Type:            store.TypeWebApp,
			Kind:            "pwa",
			URL:             "https://web.whatsapp.com",
			Browser:         "flatpak run org.chromium.Chromium",
			ExecArgs:        []string{"flatpak", "run", "org.chromium.Chromium", "--app=https://web.whatsapp.com"},
			Categories:      "Network;WebBrowser;",
			DesktopPath:     "/home/u/.local/share/applications/whatsapp-web.desktop",
			DesktopCopyPath: "/home/u/Desktop/whatsapp-web.desktop",
			Targets:         store.Targets{Menu: true},
			InstalledAt:     time.Date(2026, 10, 13, 10, 5, 0, 0, time.UTC),
		},
	}
}

func TestExportImportRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.lazy")
	items := sampleItems()

	if err := Export(path, "1.2.0", items); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("berkas ekspor tidak ada: %v", err)
	}

	got, err := Import(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(items) {
		t.Fatalf("impor %d item, mau %d", len(got), len(items))
	}

	// urutan hasil impor berdasarkan nama: WhatsApp Web lalu Winbox
	web, exe := got[0].Item, got[1].Item
	if web.Slug != "whatsapp-web" || web.URL != "https://web.whatsapp.com" || !web.Targets.Menu || web.Targets.Desktop {
		t.Errorf("data webapp tidak utuh: %+v", web)
	}
	if exe.Slug != "winbox" || exe.Kind != "windows" || len(exe.ExecArgs) != 2 || exe.ExecArgs[0] != "/usr/bin/wine" {
		t.Errorf("data executable tidak utuh: %+v", exe)
	}
	if exe.Categories != "Network;" || !exe.Targets.Desktop {
		t.Errorf("kategori/target hilang: %+v", exe)
	}
	if !exe.InstalledAt.Equal(items[0].InstalledAt) {
		t.Errorf("installed_at berubah: %v", exe.InstalledAt)
	}
}

func TestExportStoresFileContents(t *testing.T) {
	home := t.TempDir()
	binary := filepath.Join(home, "app")
	if err := os.WriteFile(binary, []byte("isi-binary-penting"), 0o755); err != nil {
		t.Fatal(err)
	}
	iconFile := filepath.Join(home, "app.png")
	iconBytes := append([]byte("\x89PNG\r\n\x1a\n"), []byte("ikon")...)
	if err := os.WriteFile(iconFile, iconBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	items := []store.Shortcut{{
		Slug: "app", Name: "App", Type: store.TypeExecutable, Kind: "binary",
		BinaryPath: binary, IconPath: iconFile, DesktopPath: filepath.Join(home, "app.desktop"),
		Terminal: true, Targets: store.Targets{Menu: true},
	}}

	path := filepath.Join(home, "db.lazy")
	if err := Export(path, "1.2.0", items); err != nil {
		t.Fatal(err)
	}

	got, err := Import(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got[0].Binary) != "isi-binary-penting" {
		t.Errorf("isi binary tidak ikut tercadang: %q", got[0].Binary)
	}
	if string(got[0].Icon) != string(iconBytes) {
		t.Errorf("isi ikon tidak ikut tercadang: %q", got[0].Icon)
	}
	if !got[0].Item.Terminal {
		t.Error("flag terminal hilang")
	}
}

func TestExportSkipsMissingFiles(t *testing.T) {
	home := t.TempDir()
	items := []store.Shortcut{{
		Slug: "hilang", Name: "Hilang", Type: store.TypeExecutable,
		BinaryPath: filepath.Join(home, "tidak-ada"), DesktopPath: filepath.Join(home, "x.desktop"),
		Targets: store.Targets{Menu: true},
	}}

	path := filepath.Join(home, "db.lazy")
	if err := Export(path, "1.2.0", items); err != nil {
		t.Fatal(err)
	}
	got, err := Import(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Binary) != 0 {
		t.Errorf("berkas hilang tidak boleh bikin ekspor gagal: %+v", got)
	}
}

func TestExportOverwritesOldFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.lazy")
	if err := os.WriteFile(path, []byte("bukan sqlite"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Export(path, "1.2.0", sampleItems()[:1]); err != nil {
		t.Fatal(err)
	}
	items, err := Import(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("isi setelah timpa: %d item", len(items))
	}
}

func TestImportRejectsForeignFile(t *testing.T) {
	dir := t.TempDir()

	plain := filepath.Join(dir, "catatan.lazy")
	if err := os.WriteFile(plain, []byte("halo dunia"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Import(plain); err == nil {
		t.Error("berkas bukan SQLite harus ditolak")
	}

	if _, err := Import(filepath.Join(dir, "tidak-ada.lazy")); err == nil {
		t.Error("berkas tidak ada harus error")
	}
}
