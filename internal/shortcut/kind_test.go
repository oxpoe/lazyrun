package shortcut

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectKind(t *testing.T) {
	dir := t.TempDir()

	write := func(name string, head []byte) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, append(head, make([]byte, 64)...), 0o755); err != nil {
			t.Fatal(err)
		}
		return path
	}

	elf := []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}

	cases := []struct {
		name string
		path string
		want string
	}{
		{"windows exe", write("winbox.exe", []byte("MZ\x90\x00")), KindWindows},
		{"script", write("tool.sh", []byte("#!/bin/sh\n")), KindScript},
		{"appimage", write("app.AppImage", append(append([]byte{}, elf...), 'A', 'I', 0x02, 0x00)), KindAppImage},
		{"binary biasa", write("binary", elf), KindBinary},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DetectKind(c.path); got != c.want {
				t.Errorf("DetectKind() = %q, mau %q", got, c.want)
			}
		})
	}

	// binary Go: test ini sendiri hasil build Go, jadi harus terdeteksi "go"
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if got := DetectKind(exe); got != KindGo {
		t.Errorf("DetectKind(test binary) = %q, mau %q", got, KindGo)
	}
}

func TestDetectKindElectron(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "myapp")
	if err := os.WriteFile(app, append([]byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}, make([]byte, 32)...), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "resources"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "resources", "app.asar"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := DetectKind(app); got != KindElectron {
		t.Errorf("DetectKind(electron) = %q", got)
	}
}

func TestDetectKindUnknownFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catatan.txt")
	if err := os.WriteFile(path, []byte("halo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DetectKind(path); got != "" {
		t.Errorf("teks biasa tidak punya kind, dapat %q", got)
	}
	if got := DetectKind(filepath.Join(t.TempDir(), "tidak-ada")); got != "" {
		t.Errorf("berkas tidak ada harus kosong, dapat %q", got)
	}
}
