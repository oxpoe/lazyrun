package shortcut

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lazyrun/internal/icon"
	"lazyrun/internal/store"
)

// iconSetEndpoint mengarahkan pencarian ikon ke server tes.
func iconSetEndpoint(t *testing.T, base string) {
	t.Helper()
	icon.SetSearchEndpoint(base)
	t.Cleanup(func() { icon.SetSearchEndpoint("http://127.0.0.1:1") })
}

func TestRefreshIconFromURL(t *testing.T) {
	testEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(append(pngMagic, []byte("ikon-baru")...))
	}))
	defer srv.Close()

	deskDst := filepath.Join(t.TempDir(), "x.desktop")
	old := Entry{Name: "X", ExecArgs: []string{"/x"}, Icon: "/lama/icon.png"}
	if err := WriteDesktop(deskDst, old); err != nil {
		t.Fatal(err)
	}

	sc := store.Shortcut{
		Slug:        "x",
		Name:        "X",
		IconPath:    "/lama/icon.png",
		IconSource:  srv.URL + "/icon.png",
		DesktopPath: deskDst,
	}
	path, err := RefreshIcon(sc)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "x.png" {
		t.Errorf("ikon baru di %s", path)
	}
	if data, err := os.ReadFile(path); err != nil || len(data) == 0 {
		t.Errorf("ikon tidak tertulis: %v", err)
	}

	content, err := os.ReadFile(deskDst)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Icon="+path+"\n") {
		t.Errorf("baris Icon= tidak diperbarui:\n%s", content)
	}
}

func TestRefreshIconKeepsDesktopCopyMode(t *testing.T) {
	home := testEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(append(pngMagic, []byte("baru")...))
	}))
	defer srv.Close()

	deskDst := filepath.Join(home, ".local/share/applications/y.desktop")
	copyDst := filepath.Join(home, "Desktop", "y.desktop")
	for _, path := range []string{deskDst, copyDst} {
		if err := WriteDesktop(path, Entry{Name: "Y", ExecArgs: []string{"/y"}, Icon: "/lama.png"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(copyDst, 0o755); err != nil {
		t.Fatal(err)
	}

	sc := store.Shortcut{
		Slug:            "y",
		IconSource:      srv.URL,
		DesktopPath:     deskDst,
		DesktopCopyPath: copyDst,
	}
	if _, err := RefreshIcon(sc); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(copyDst)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm()&0o111 == 0 {
		t.Errorf("salinan desktop kehilangan mode executable: %v", st.Mode())
	}
}

func TestRefreshIconWithoutOnlineSource(t *testing.T) {
	testEnv(t)
	sc := store.Shortcut{Slug: "z", IconSource: "file:/tmp/icon.png", DesktopPath: "/tmp/z.desktop"}
	if _, err := RefreshIcon(sc); err == nil {
		t.Error("ikon dari file lokal harus error, bukan diunduh")
	}
}

func TestInstallExecutableSearchesIconOnline(t *testing.T) {
	home := testEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/search"):
			_, _ = w.Write([]byte(`{"icons":["mdi:winbox"]}`))
		case r.URL.Path == "/mdi/winbox.svg":
			_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256"></svg>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	iconSetEndpoint(t, srv.URL)

	src := fakeBinary(t, home, "winbox")
	sc, err := InstallExecutable(ExecOptions{BinaryPath: src, Name: "Winbox", Menu: true})
	if err != nil {
		t.Fatal(err)
	}
	if sc.IconSource != "search:Winbox" {
		t.Errorf("icon_source = %q", sc.IconSource)
	}
	if filepath.Ext(sc.IconPath) != ".png" {
		t.Errorf("SVG online harus diraster jadi .png, dapat %s", sc.IconPath)
	}
	assertPNG(t, sc.IconPath)
}

func TestInstallWebAppUsesIconURL(t *testing.T) {
	testEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pilih.svg":
			w.Header().Set("Content-Type", "image/svg+xml")
			_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256"></svg>`))
		case "/":
			_, _ = w.Write([]byte(`<html><head><link rel="icon" href="/situs.png"></head></html>`))
		default:
			_, _ = w.Write(append(pngMagic, []byte("favicon-situs")...))
		}
	}))
	defer srv.Close()

	sc, err := InstallWebApp(WebOptions{
		URL:     "https://example.com",
		Name:    "Example",
		Browser: BrowserArg{Name: "Chromium", Exec: []string{"/usr/bin/chromium"}},
		IconURL: srv.URL + "/pilih.svg",
		Menu:    true,
		Desktop: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sc.IconSource != srv.URL+"/pilih.svg" {
		t.Errorf("ikon pilihan user harus dipakai, icon_source = %q", sc.IconSource)
	}
	if filepath.Ext(sc.IconPath) != ".png" {
		t.Errorf("ikon = %s", sc.IconPath)
	}
	assertPNG(t, sc.IconPath)
}

func TestInstallWebAppUsesIconDataFromUI(t *testing.T) {
	testEnv(t)

	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256"></svg>`)
	sc, err := InstallWebApp(WebOptions{
		URL:      "https://example.com",
		Name:     "Example",
		Browser:  BrowserArg{Name: "Chromium", Exec: []string{"/usr/bin/chromium"}},
		IconURL:  "https://api.iconify.design/mdi/example.svg?height=256",
		IconData: svg, // sudah diunduh UI: tidak perlu request ulang
		Menu:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sc.IconSource != "https://api.iconify.design/mdi/example.svg?height=256" {
		t.Errorf("icon_source = %q", sc.IconSource)
	}
	assertPNG(t, sc.IconPath)
}

// assertPNG memastikan file ikon berisi PNG (bukan SVG mentah).
func assertPNG(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ikon tidak terbaca: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) {
		t.Errorf("%s bukan PNG: %q", path, data[:min(12, len(data))])
	}
}
