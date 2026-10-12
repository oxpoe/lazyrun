// Package icon menyimpan ikon shortcut ke direktori ikon XDG.
package icon

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"lazyrun/assets"
	"lazyrun/internal/xdg"
)

var exts = []string{".png", ".svg", ".jpg", ".jpeg", ".xpm"}
var nearbyNames = []string{"icon", "logo"}

// Save memasang ikon untuk shortcut dan mengembalikan path absolutnya.
// Urutan pilihan: ikon dari user, ikon di folder binary, lalu ikon default LazyRun.
func Save(userIcon, nearbyDir, nearbyBase, slug string) (string, error) {
	if userIcon != "" {
		return fromFile(userIcon, slug)
	}
	for _, name := range append([]string{nearbyBase}, nearbyNames...) {
		if name == "" {
			continue
		}
		if p := findNearby(nearbyDir, name); p != "" {
			return fromFile(p, slug)
		}
	}
	return Default(slug)
}

// Default memasang ikon bawaan LazyRun.
func Default(slug string) (string, error) {
	return SaveBytes(assets.Icon, ".png", slug)
}

func findNearby(dir, base string) string {
	base = strings.ToLower(base)
	for _, ext := range exts {
		p := filepath.Join(dir, base+ext)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func fromFile(src, slug string) (string, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(src))
	// ICO tidak dipakai desktop theme; pakai PNG di dalamnya kalau ada.
	if ext == ".ico" {
		png := largestPNGFromICO(data)
		if png == nil {
			return "", errors.New("this .ico has no PNG inside it, pick another icon")
		}
		return SaveBytes(png, ".png", slug)
	}
	return SaveBytes(data, ext, slug)
}

// SaveBytes menulis data ikon ke direktori ikon XDG.
func SaveBytes(data []byte, ext, slug string) (string, error) {
	if ext == "" {
		ext = ".png"
	}
	if err := os.MkdirAll(xdg.IconsDir(), 0o755); err != nil {
		return "", err
	}
	// shortcut: ikon disimpan apa adanya di hicolor/512x512 tanpa resize;
	// desktop environment men-skalanya sendiri. Tambah resize kalau ikon kecil terlihat pecah.
	dest := filepath.Join(xdg.IconsDir(), slug+ext)
	return dest, os.WriteFile(dest, data, 0o644)
}

// IconBytes adalah ikon default LazyRun.
func IconBytes() []byte { return assets.Icon }
