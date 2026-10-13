package shortcut

import (
	"bytes"
	"debug/buildinfo"
	"os"
	"path/filepath"
)

// Jenis aplikasi yang bisa dikenali LazyRun.
const (
	KindGo       = "go"
	KindAppImage = "appimage"
	KindElectron = "electron"
	KindScript   = "script"
	KindWindows  = "windows"
	KindBinary   = "binary"
	KindPWA      = "pwa"
)

// DetectKind menebak jenis berkas yang dipilih user:
// binary Go, AppImage, aplikasi Electron, script, .exe Windows, atau binary biasa.
func DetectKind(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	head := make([]byte, 12)
	n, _ := f.Read(head)
	f.Close()
	head = head[:n]

	switch {
	case bytes.HasPrefix(head, []byte("MZ")):
		return KindWindows
	case bytes.HasPrefix(head, []byte("#!")):
		return KindScript
	case isAppImageHead(head):
		return KindAppImage
	case bytes.HasPrefix(head, []byte{0x7f, 'E', 'L', 'F'}):
		if _, err := buildinfo.ReadFile(path); err == nil {
			return KindGo
		}
		if hasElectronResources(path) {
			return KindElectron
		}
		return KindBinary
	}
	return ""
}

// hasElectronResources mencari berkas penanda Electron di sebelah binary.
func hasElectronResources(path string) bool {
	resources := filepath.Join(filepath.Dir(path), "resources")
	for _, name := range []string{"app.asar", "default_app.asar", "app"} {
		if _, err := os.Stat(filepath.Join(resources, name)); err == nil {
			return true
		}
	}
	return false
}

// isAppImageHead memeriksa magic AppImage (ELF + "AI" pada offset 8).
// Sengaja terpisah dari internal/update supaya paket ini tidak tergantung ke sana.
func isAppImageHead(head []byte) bool {
	return len(head) >= 11 &&
		head[0] == 0x7f && head[1] == 'E' && head[2] == 'L' && head[3] == 'F' &&
		head[8] == 'A' && head[9] == 'I' && (head[10] == 0x01 || head[10] == 0x02)
}
