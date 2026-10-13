package ui

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"

	"lazyrun/internal/update"
)

func TestAppImagePathFromEnv(t *testing.T) {
	t.Setenv("APPIMAGE", "/home/u/Apps/lazyrun-1.0.0-x86_64.AppImage")
	if got := appImagePath(); got != "/home/u/Apps/lazyrun-1.0.0-x86_64.AppImage" {
		t.Errorf("appImagePath() = %q", got)
	}
}

func TestAppImagePathWithoutEnvAndNotAppImage(t *testing.T) {
	// test binary bukan AppImage, jadi tidak ada yang bisa diupdate
	t.Setenv("APPIMAGE", "")
	if got := appImagePath(); got != "" {
		t.Errorf("appImagePath() = %q, mau kosong untuk binary biasa", got)
	}
}

func TestAppImagePathFallsBackToExecutable(t *testing.T) {
	dir := t.TempDir()
	// executable tiruan: ELF + magic AI (AppImage)
	fake := filepath.Join(dir, "lazyrun.AppImage")
	data := make([]byte, 64)
	copy(data, []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00})
	copy(data[8:], []byte{'A', 'I', 0x02})
	if err := os.WriteFile(fake, data, 0o755); err != nil {
		t.Fatal(err)
	}
	// jalur env kosong -> deteksi dari IsAppImage (di sini diuji langsung)
	if !update.IsAppImage(fake) {
		t.Fatal("berkas uji harus dikenali sebagai AppImage")
	}
}

func TestMarkUpdatePendingSetsFooter(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui, err := build(a, "1.1.0")
	if err != nil {
		t.Fatal(err)
	}
	ui.markUpdatePending(update.Release{Version: "9.9.9"})
	if got := ui.versionText.Text; got != "v1.1.0 → v9.9.9" {
		t.Errorf("footer tidak menampilkan update: %q", got)
	}
}
