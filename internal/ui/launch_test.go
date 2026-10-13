package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"lazyrun/internal/store"
)

func TestLaunchCommandPrefersDesktopFile(t *testing.T) {
	dir := t.TempDir()
	desk := filepath.Join(dir, "tool.desktop")
	if err := os.WriteFile(desk, []byte("[Desktop Entry]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sc := store.Shortcut{DesktopPath: desk, BinaryPath: "/bin/true", ExecArgs: []string{"/bin/true"}}
	args, err := launchCommand(sc)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := exec.LookPath("gio"); err != nil {
		// tanpa gio: langsung jalankan perintahnya
		if len(args) != 1 || args[0] != "/bin/true" {
			t.Errorf("tanpa gio harus jalankan binary, dapat %v", args)
		}
		return
	}
	if len(args) != 3 || args[1] != "launch" || args[2] != desk {
		t.Errorf("harus lewat gio launch supaya Terminal=true dihormati, dapat %v", args)
	}
}

func TestLaunchCommandFallsBack(t *testing.T) {
	// .desktop tidak ada -> pakai exec args
	sc := store.Shortcut{DesktopPath: "/tidak/ada/x.desktop", ExecArgs: []string{"/usr/bin/wine", "/x/winbox"}}
	args, err := launchCommand(sc)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != "/usr/bin/wine" {
		t.Errorf("fallback exec args salah: %v", args)
	}

	// web app tanpa exec args
	web := store.Shortcut{Type: store.TypeWebApp, URL: "https://example.com"}
	args, err = launchCommand(web)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != "xdg-open" {
		t.Errorf("fallback web app salah: %v", args)
	}

	// tanpa apa-apa
	if _, err := launchCommand(store.Shortcut{}); err == nil {
		t.Error("shortcut kosong harus error, bukan jalan tanpa perintah")
	}
}
