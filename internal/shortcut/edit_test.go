package shortcut

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditExecutableRenamesAndMovesFiles(t *testing.T) {
	home := testEnv(t)
	src := fakeBinary(t, home, "old-tool")

	sc, err := InstallExecutable(ExecOptions{BinaryPath: src, Name: "Old Tool", Category: "Utility", Menu: true, Desktop: true})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := EditExecutable(sc, ExecOptions{
		BinaryPath: sc.BinaryPath, // binary tidak diganti
		Name:       "New Tool",
		Category:   "Network",
		Menu:       true,
		Desktop:    false, // salinan Desktop dimatikan
	})
	if err != nil {
		t.Fatal(err)
	}

	if updated.Slug != "new-tool" || updated.BinaryPath != filepath.Join(home, ".local", "bin", "new-tool") {
		t.Errorf("slug/binary baru salah: %+v", updated)
	}
	if _, err := os.Stat(updated.DesktopPath); err != nil {
		t.Errorf(".desktop baru tidak ada: %v", err)
	}
	if !strings.HasSuffix(updated.DesktopPath, "new-tool.desktop") {
		t.Errorf("nama .desktop = %s", updated.DesktopPath)
	}
	if updated.DesktopCopyPath != "" {
		t.Errorf("salinan desktop harus kosong, dapat %s", updated.DesktopCopyPath)
	}
	// berkas lama sudah dibersihkan
	if _, err := os.Stat(sc.DesktopPath); !os.IsNotExist(err) {
		t.Errorf(".desktop lama masih ada")
	}
	if _, err := os.Stat(sc.DesktopCopyPath); !os.IsNotExist(err) {
		t.Errorf("salinan desktop lama masih ada")
	}

	content, err := os.ReadFile(updated.DesktopPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Name=New Tool\n") || !strings.Contains(string(content), "Categories=Network;\n") {
		t.Errorf("isi .desktop salah:\n%s", content)
	}
	if !strings.Contains(string(content), "Exec="+updated.BinaryPath+"\n") {
		t.Errorf("Exec tidak menunjuk binary baru:\n%s", content)
	}
}

func TestEditExecutableReplacesBinaryAndKind(t *testing.T) {
	home := testEnv(t)
	first := fakeBinary(t, home, "first")
	second := filepath.Join(home, "second.sh")
	if err := os.WriteFile(second, []byte("#!/bin/sh\necho hai\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	sc, err := InstallExecutable(ExecOptions{BinaryPath: first, Name: "Tool", Menu: true})
	if err != nil {
		t.Fatal(err)
	}
	if sc.Kind != KindBinary {
		t.Fatalf("kind awal = %q", sc.Kind)
	}

	updated, err := EditExecutable(sc, ExecOptions{BinaryPath: second, Name: "Tool", Menu: true})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Kind != KindScript {
		t.Errorf("kind setelah ganti binary = %q, mau %q", updated.Kind, KindScript)
	}
	if updated.IconPath == "" {
		t.Error("ikon harus tetap ada")
	}
}

func TestEditExecutableKeepsIconUnlessChanged(t *testing.T) {
	home := testEnv(t)
	src := fakeBinary(t, home, "tool")
	iconFile := filepath.Join(home, "ikon.png")
	if err := os.WriteFile(iconFile, append(pngMagic, []byte("ikon")...), 0o644); err != nil {
		t.Fatal(err)
	}

	sc, err := InstallExecutable(ExecOptions{BinaryPath: src, Name: "Tool", Menu: true, IconPath: iconFile})
	if err != nil {
		t.Fatal(err)
	}
	before := sc.IconPath

	updated, err := EditExecutable(sc, ExecOptions{BinaryPath: sc.BinaryPath, Name: "Tool", Menu: true})
	if err != nil {
		t.Fatal(err)
	}
	if updated.IconPath != before {
		t.Errorf("ikon berubah tanpa diminta: %s -> %s", before, updated.IconPath)
	}

	// ganti ikon dengan berkas lain: nama berkasnya sama (slug sama),
	// jadi yang diperiksa isinya
	beforeBytes, err := os.ReadFile(before)
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(home, "lain.png")
	if err := os.WriteFile(other, append(pngMagic, []byte("ikon lain yang lebih panjang")...), 0o644); err != nil {
		t.Fatal(err)
	}
	updated, err = EditExecutable(sc, ExecOptions{BinaryPath: sc.BinaryPath, Name: "Tool", Menu: true, IconPath: other})
	if err != nil {
		t.Fatal(err)
	}
	afterBytes, err := os.ReadFile(updated.IconPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(beforeBytes) == string(afterBytes) {
		t.Error("isi ikon tidak ikut berubah saat diganti")
	}
}

func TestEditWebAppChangesURLAndTargets(t *testing.T) {
	testEnv(t)
	sc, err := InstallWebApp(WebOptions{
		URL:     "https://lama.example",
		Name:    "Contoh",
		Browser: BrowserArg{Name: "Chromium", Exec: []string{"/usr/bin/chromium"}},
		Menu:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := EditWebApp(sc, WebOptions{
		URL:     "lama.example/baru",
		Name:    "Contoh Baru",
		Browser: BrowserArg{Name: "Firefox", Exec: []string{"/usr/bin/firefox"}, Firefox: true},
		Menu:    false,
		Desktop: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if updated.URL != "https://lama.example/baru" {
		t.Errorf("URL = %s", updated.URL)
	}
	if updated.DesktopCopyPath == "" {
		t.Error("salinan desktop harus dibuat")
	}
	content, err := os.ReadFile(updated.DesktopPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Exec=/usr/bin/firefox --new-window https://lama.example/baru\n") {
		t.Errorf("Exec salah:\n%s", content)
	}
	if !strings.Contains(string(content), "NoDisplay=true") {
		t.Errorf("menu dimatikan harus NoDisplay:\n%s", content)
	}
}

func TestEditExecutableRejectsDuplicateName(t *testing.T) {
	home := testEnv(t)
	a := fakeBinary(t, home, "app-a")
	b := fakeBinary(t, home, "app-b")

	if _, err := InstallExecutable(ExecOptions{BinaryPath: a, Name: "App A", Menu: true}); err != nil {
		t.Fatal(err)
	}
	scB, err := InstallExecutable(ExecOptions{BinaryPath: b, Name: "App B", Menu: true})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := EditExecutable(scB, ExecOptions{BinaryPath: scB.BinaryPath, Name: "App A", Menu: true}); err == nil {
		t.Error("nama yang sudah dipakai harus ditolak")
	}
}

func TestEditWebAppRefreshesIconFromSite(t *testing.T) {
	testEnv(t)

	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			_, _ = w.Write([]byte(`<html><head><link rel="icon" type="image/png" sizes="64x64" href="/baru.png"></head></html>`))
		case "/paksa.png":
			_, _ = w.Write(append(pngMagic, []byte("ikon paksa yang panjang")...))
		default:
			_, _ = w.Write(append(pngMagic, []byte("ikon lama")...))
		}
	}))
	defer page.Close()

	sc, err := InstallWebApp(WebOptions{
		URL:     page.URL,
		Name:    "Situs",
		Browser: BrowserArg{Name: "Chromium", Exec: []string{"/usr/bin/chromium"}},
		Menu:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// ikon sebelum diedit (harus dibaca sebelum berkasnya ditimpa)
	before, err := os.ReadFile(sc.IconPath)
	if err != nil {
		t.Fatal(err)
	}

	// ubah URL ke halaman lain -> ikon lama harus diganti sumbernya
	updated, err := EditWebApp(sc, WebOptions{
		URL:     page.URL + "/halaman",
		Name:    "Situs",
		Browser: BrowserArg{Name: "Chromium", Exec: []string{"/usr/bin/chromium"}},
		IconURL: page.URL + "/paksa.png",
		Menu:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.IconSource != page.URL+"/paksa.png" {
		t.Errorf("icon_source = %q", updated.IconSource)
	}
	after, err := os.ReadFile(updated.IconPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) == string(after) {
		t.Error("isi ikon harus diperbarui dari sumber baru")
	}
}
