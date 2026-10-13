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
	web, exe := got[0], got[1]
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
