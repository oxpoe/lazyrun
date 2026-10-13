package shortcut

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lazyrun/internal/store"
)

func TestRestoreExecutableFromBackup(t *testing.T) {
	home := testEnv(t)

	backup := store.Shortcut{
		Slug:        "winbox",
		Name:        "Winbox",
		Type:        store.TypeExecutable,
		Kind:        "binary",
		BinaryPath:  "/komputer-lama/.local/bin/winbox", // path dari mesin lain
		IconPath:    "/komputer-lama/.local/share/icons/winbox.png",
		DesktopPath: "/komputer-lama/.local/share/applications/winbox.desktop",
		Categories:  "Network;",
		Targets:     store.Targets{Menu: true, Desktop: true},
	}
	restored, err := Restore(backup, append([]byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}, []byte("isi")...), nil)
	if err != nil {
		t.Fatal(err)
	}

	wantBin := filepath.Join(home, ".local", "bin", "winbox")
	if restored.BinaryPath != wantBin {
		t.Errorf("binary = %s, mau %s", restored.BinaryPath, wantBin)
	}
	if st, err := os.Stat(wantBin); err != nil || st.Mode().Perm()&0o111 == 0 {
		t.Fatalf("binary hasil restore tidak executable: %v %v", st, err)
	}
	// path dari mesin lama tidak boleh dipakai
	if strings.Contains(restored.DesktopPath, "komputer-lama") {
		t.Errorf("desktop path masih menunjuk mesin lama: %s", restored.DesktopPath)
	}
	if restored.DesktopCopyPath != filepath.Join(home, "Desktop", "winbox.desktop") {
		t.Errorf("salinan desktop = %s", restored.DesktopCopyPath)
	}
	content, err := os.ReadFile(restored.DesktopPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Exec="+wantBin+"\n") || !strings.Contains(string(content), "Categories=Network;\n") {
		t.Errorf("isi .desktop hasil restore salah:\n%s", content)
	}
	if _, err := os.Stat(restored.IconPath); err != nil {
		t.Errorf("ikon hasil restore tidak ada: %v", err)
	}
}

func TestRestoreScriptGetsWrapper(t *testing.T) {
	home := testEnv(t)

	backup := store.Shortcut{
		Name:       "Deploy",
		Type:       store.TypeExecutable,
		Kind:       "script",
		BinaryPath: "/komputer-lama/proyek/deploy.sh",
		Terminal:   true,
		Targets:    store.Targets{Menu: true},
	}
	restored, err := Restore(backup, []byte("#!/bin/sh\necho halo\n"), nil)
	if err != nil {
		t.Fatal(err)
	}

	if restored.Kind != KindScript {
		t.Errorf("kind = %q", restored.Kind)
	}
	if _, err := os.Stat(filepath.Join(home, ".local", "bin", "deploy")); err != nil {
		t.Errorf("wrapper tidak dibuat: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".local", "bin", "deploy"+scriptSuffix)); err != nil {
		t.Errorf("script asli tidak dibuat: %v", err)
	}

	content, err := os.ReadFile(restored.DesktopPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Terminal=true") {
		t.Errorf("flag terminal hilang:\n%s", content)
	}
}

func TestRestoreWithoutBinaryBlobMarksMissing(t *testing.T) {
	testEnv(t)

	backup := store.Shortcut{
		Name:       "Hilang",
		Type:       store.TypeExecutable,
		BinaryPath: "/komputer-lama/.local/bin/hilang",
		Targets:    store.Targets{Menu: true},
	}
	restored, err := Restore(backup, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if restored.BinaryPath != "" {
		t.Errorf("binary mestinya kosong (tidak bisa dipulihkan), dapat %s", restored.BinaryPath)
	}
	if restored.ExecArgs != nil {
		t.Errorf("exec args mestinya kosong: %v", restored.ExecArgs)
	}
	// .desktop tetap dibuat supaya entri bisa dibetulkan lewat edit
	if _, err := os.Stat(restored.DesktopPath); err != nil {
		t.Errorf(".desktop tidak dibuat: %v", err)
	}
}

func TestRestoreWebAppRebuildsExec(t *testing.T) {
	testEnv(t)

	backup := store.Shortcut{
		Name:    "WhatsApp Web",
		Type:    store.TypeWebApp,
		Kind:    KindPWA,
		URL:     "https://web.whatsapp.com",
		Browser: "flatpak run org.chromium.Chromium",
		Targets: store.Targets{Menu: true},
	}
	restored, err := Restore(backup, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.ExecArgs) != 4 || restored.ExecArgs[3] != "--app=https://web.whatsapp.com" {
		t.Errorf("exec args PWA salah: %v", restored.ExecArgs)
	}
}

func TestRestoreIconFromSource(t *testing.T) {
	testEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(append(pngMagic, []byte("ikon dari internet")...))
	}))
	defer srv.Close()

	backup := store.Shortcut{
		Name:       "Icon",
		Type:       store.TypeExecutable,
		BinaryPath: "",
		IconPath:   "/komputer-lama/ikon.png",
		IconSource: srv.URL + "/ikon.png",
		Targets:    store.Targets{Menu: true},
	}
	restored, err := Restore(backup, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(restored.IconPath)
	if err != nil {
		t.Fatalf("ikon tidak dipasang: %v", err)
	}
	if !strings.Contains(string(data), "ikon dari internet") {
		t.Errorf("ikon tidak diunduh ulang dari sumber: %q", data)
	}
}
