package update

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"v1.0.0", "1.0.0", 0},
		{"1.1.0", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.10.0", "1.9.0", 1},
		{"2.0", "1.9.9", 1},
		{"1.0", "1.0.0", 0},
		{"1.0.0-rc1", "1.0.0", -1},
		{"1.0.0", "1.0.0-rc1", 1},
		{"dev", "1.0.0", -1},
	}
	for _, c := range cases {
		if got := CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("CompareVersions(%q, %q) = %d, mau %d", c.a, c.b, got, c.want)
		}
	}
}

func TestNewerIgnoresSameAndOlder(t *testing.T) {
	if Newer("1.0.0", Release{Version: "1.0.0"}) {
		t.Error("versi sama bukan update")
	}
	if Newer("1.0.0", Release{Version: "0.9.0"}) {
		t.Error("versi lama bukan update")
	}
	if !Newer("1.0.0", Release{Version: "1.0.1"}) {
		t.Error("versi lebih baru harus terdeteksi")
	}
}

func TestLatestParsesRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("GitHub butuh header User-Agent")
		}
		_, _ = w.Write([]byte(`{
			"tag_name": "v1.2.0",
			"body": "  - ikon online\n  - animasi  ",
			"html_url": "https://github.com/oxpoe/lazyrun/releases/tag/v1.2.0",
			"assets": [
				{"name": "lazyrun-1.2.0-x86_64.AppImage.zsync", "browser_download_url": "https://contoh/zsync"},
				{"name": "lazyrun-1.2.0-x86_64.AppImage", "browser_download_url": "https://contoh/appimage"}
			]
		}`))
	}))
	defer srv.Close()

	SetEndpoint(srv.URL)
	t.Cleanup(func() { SetEndpoint("https://api.github.com/repos/oxpoe/lazyrun/releases") })

	rel, err := Latest()
	if err != nil {
		t.Fatal(err)
	}
	if rel.Version != "1.2.0" || rel.Tag != "v1.2.0" {
		t.Errorf("versi = %q / %q", rel.Version, rel.Tag)
	}
	if rel.AppImage != "https://contoh/appimage" || rel.Zsync != "https://contoh/zsync" {
		t.Errorf("aset salah: %q %q", rel.AppImage, rel.Zsync)
	}
	if rel.Notes != "- ikon online\n  - animasi" {
		t.Errorf("catatan rilis = %q", rel.Notes)
	}
	if !Newer("1.1.9", rel) {
		t.Error("1.2.0 harus dianggap lebih baru dari 1.1.9")
	}
}

func TestDownloadReplacesAppImageAtomically(t *testing.T) {
	appImage := fakeAppImage(2048)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(appImage)
	}))
	defer srv.Close()

	target := filepath.Join(t.TempDir(), "lazyrun.AppImage")
	if err := os.WriteFile(target, fakeAppImage(1024), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Download(Release{AppImage: srv.URL}, target); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(appImage) {
		t.Error("isi AppImage tidak terganti")
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Error("AppImage hasil update harus executable")
	}
	// tidak boleh menyisakan berkas sementara
	entries, err := os.ReadDir(filepath.Dir(target))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("ada berkas sisa: %v", entries)
	}
}

func TestDownloadKeepsOldAppImageOnBadData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html>404</html>"))
	}))
	defer srv.Close()

	target := filepath.Join(t.TempDir(), "lazyrun.AppImage")
	old := fakeAppImage(512)
	if err := os.WriteFile(target, old, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Download(Release{AppImage: srv.URL}, target); err == nil {
		t.Fatal("berkas bukan AppImage harus ditolak")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(old) {
		t.Error("AppImage lama harus tetap utuh saat update gagal")
	}
}

func TestDownloadWithoutAsset(t *testing.T) {
	target := filepath.Join(t.TempDir(), "lazyrun.AppImage")
	if err := Download(Release{}, target); err == nil {
		t.Error("rilis tanpa AppImage harus error")
	}
}

func TestIsAppImage(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good")
	if err := os.WriteFile(good, fakeAppImage(16), 0o755); err != nil {
		t.Fatal(err)
	}
	if !IsAppImage(good) {
		t.Error("AppImage sah tidak dikenali")
	}

	plain := filepath.Join(dir, "plain")
	if err := os.WriteFile(plain, []byte("\x7fELF\x02\x01\x01\x00no-magic-here"), 0o755); err != nil {
		t.Fatal(err)
	}
	if IsAppImage(plain) {
		t.Error("ELF biasa bukan AppImage")
	}
}

// fakeAppImage membuat berkas dengan magic ELF + AI (cukup untuk verifikasi).
func fakeAppImage(size int) []byte {
	data := make([]byte, size)
	copy(data, []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00})
	copy(data[8:], []byte{'A', 'I', 0x02})
	return data
}
