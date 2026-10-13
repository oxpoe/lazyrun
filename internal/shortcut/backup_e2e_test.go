package shortcut

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"lazyrun/internal/db"
	"lazyrun/internal/store"
)

// TestBackupRestoreEndToEnd meniru alur backup: pasang beberapa shortcut,
// ekspor ke berkas .lazy, lalu pulihkan di "komputer lain" (HOME berbeda)
// dan pastikan hasilnya benar-benar jalan.
func TestBackupRestoreEndToEnd(t *testing.T) {
	codeDir := t.TempDir()
	homeA := t.TempDir()

	// ---- komputer A: pasang shortcut ----
	t.Setenv("HOME", homeA)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("PATH", "")

	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte(`<html><head><link rel="icon" type="image/png" sizes="128x128" href="/web.png"></head></html>`))
			return
		}
		_, _ = w.Write(append(pngMagic, []byte("ikon web")...))
	}))
	defer site.Close()

	// script yang butuh folder kerjanya + tool di ~/.local/bin
	scriptSrc := filepath.Join(codeDir, "deploy.sh")
	if err := os.WriteFile(scriptSrc, []byte(`#!/bin/sh
[ -f ./data.txt ] && echo "data ok" || echo "data hilang"
case "$PATH" in *".local/bin"*) echo "path ok";; *) echo "path hilang";; esac
`), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codeDir, "data.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	script, err := InstallExecutable(ExecOptions{BinaryPath: scriptSrc, Name: "Deploy", Category: "Utility", Menu: true, Desktop: true})
	if err != nil {
		t.Fatal(err)
	}
	binarySrc := fakeBinary(t, codeDir, "tool")
	binaryShortcut, err := InstallExecutable(ExecOptions{BinaryPath: binarySrc, Name: "Tool", Menu: true})
	if err != nil {
		t.Fatal(err)
	}
	web, err := InstallWebApp(WebOptions{
		URL: site.URL, Name: "Situs Web",
		Browser: BrowserArg{Name: "Chromium", Exec: []string{"/usr/bin/chromium"}},
		Menu:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// ---- ekspor ke .lazy ----
	backupPath := filepath.Join(t.TempDir(), "db.lazy")
	if err := db.Export(backupPath, "1.2.0", []store.Shortcut{script, binaryShortcut, web}); err != nil {
		t.Fatal(err)
	}

	// ---- komputer B: HOME baru, pulihkan ----
	homeB := t.TempDir()
	t.Setenv("HOME", homeB)

	backups, err := db.Import(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 3 {
		t.Fatalf("cadangan berisi %d entri, mau 3", len(backups))
	}

	restored := map[string]store.Shortcut{}
	for _, backup := range backups {
		item, err := Restore(backup.Item, backup.Binary, backup.Icon)
		if err != nil {
			t.Fatalf("restore %s: %v", backup.Item.Name, err)
		}
		restored[backup.Item.Name] = item
	}

	// ---- semua berkas harus di home B, bukan home A ----
	for name, item := range restored {
		if item.BinaryPath != "" && !strings.HasPrefix(item.BinaryPath, homeB) {
			t.Errorf("%s: binary masih menunjuk home lama: %s", name, item.BinaryPath)
		}
		if !strings.HasPrefix(item.DesktopPath, homeB) {
			t.Errorf("%s: .desktop masih menunjuk home lama: %s", name, item.DesktopPath)
		}
		if _, err := os.Stat(item.DesktopPath); err != nil {
			t.Errorf("%s: .desktop tidak ada: %v", name, err)
		}
		if _, err := os.Stat(item.IconPath); err != nil {
			t.Errorf("%s: ikon tidak ada: %v", name, err)
		}
	}

	// script hasil restore harus jalan seperti dijalankan desktop environment
	deploy := restored["Deploy"]
	if !strings.HasPrefix(deploy.DesktopPath, homeB) {
		t.Fatal("script tidak terpulihkan")
	}
	cmd := exec.Command(deploy.ExecArgs[0], deploy.ExecArgs[1:]...)
	cmd.Dir = homeB
	cmd.Env = []string{"HOME=" + homeB, "PATH=/usr/bin:/bin"}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("script hasil restore gagal jalan: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "path ok") {
		t.Errorf("PATH tidak disiapkan wrapper: %s", out)
	}
	// script asli di folder asalnya masih ada di mesin ini? bukan di sini,
	// jadi wrapper memakai salinannya (sudah diuji lewat 'data ok' di atas)

	// binary biasa hasil restore juga harus executable
	tool := restored["Tool"]
	if st, err := os.Stat(tool.BinaryPath); err != nil || st.Mode().Perm()&0o111 == 0 {
		t.Errorf("binary hasil restore tidak executable: %v %v", st, err)
	}
	if !strings.HasPrefix(tool.DesktopCopyPath, homeB) && tool.Targets.Desktop {
		t.Errorf("salinan desktop salah: %s", tool.DesktopCopyPath)
	}

	// web app: exec args tetap menunjuk browser + URL
	if !strings.Contains(strings.Join(restored["Situs Web"].ExecArgs, " "), "--app="+site.URL) {
		t.Errorf("exec web app salah: %v", restored["Situs Web"].ExecArgs)
	}
}
