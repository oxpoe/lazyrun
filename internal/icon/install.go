// Package icon menyimpan ikon shortcut ke direktori ikon XDG.
package icon

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"

	"lazyrun/assets"
	"lazyrun/internal/xdg"
)

var exts = []string{".png", ".svg", ".jpg", ".jpeg", ".webp", ".gif", ".bmp", ".ico", ".xpm"}

var pngMagic = []byte("\x89PNG\r\n\x1a\n")

// maxSide adalah sisi terpanjang ikon yang disimpan.
const maxSide = 512

// Default memasang ikon bawaan LazyRun.
func Default(slug string) (string, error) {
	return Save(assets.Icon, slug)
}

// Nearby mencari file ikon bernama base.* di folder dir (dipakai untuk auto-deteksi).
func Nearby(dir, base string) string {
	if dir == "" || base == "" {
		return ""
	}
	return findNearby(dir, base)
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

// Save menulis ikon ke direktori ikon XDG sebagai PNG.
//
// Semua format (SVG, webp, gif, bmp, ico, jpeg) dikonversi dulu ke PNG karena
// desktop environment tidak konsisten membacanya, dan sebagian SVG dari internet
// punya fitur yang bikin proses menggambar berhenti di UI.
func Save(data []byte, slug string) (string, error) {
	converted, err := ToPNG(data)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(xdg.IconsDir(), 0o755); err != nil {
		return "", err
	}
	dest := filepath.Join(xdg.IconsDir(), slug+".png")
	return dest, os.WriteFile(dest, converted, 0o644)
}

// Ext menebak ekstensi file dari isi data.
func Ext(data []byte) string { return sniffExt(data) }

// IconBytes adalah ikon default LazyRun.
func IconBytes() []byte { return assets.Icon }

// ToPNG mengubah data gambar apa pun yang didukung menjadi PNG.
// Data yang sudah PNG dikembalikan apa adanya.
func ToPNG(data []byte) ([]byte, error) {
	if bytes.HasPrefix(data, pngMagic) {
		return data, nil
	}

	if sniffExt(data) == ".svg" {
		return Rasterize(data, 256)
	}

	// jpeg, gif, bmp, webp dikenali dari isi file
	if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
		return encodePNG(img)
	}

	// ico: bisa berisi PNG langsung atau gambar BMP di dalamnya
	if img, err := decodeICO(data); err == nil {
		return encodePNG(img)
	}
	return nil, errors.New("format ikon tidak didukung")
}

// encodePNG menulis gambar jadi PNG, diperkecil kalau lebih besar dari maxSide.
func encodePNG(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	if bounds.Dx() > maxSide || bounds.Dy() > maxSide {
		scale := float64(maxSide) / float64(max(bounds.Dx(), bounds.Dy()))
		size := image.Rect(0, 0, int(float64(bounds.Dx())*scale), int(float64(bounds.Dy())*scale))
		scaled := image.NewNRGBA(size)
		xdraw.CatmullRom.Scale(scaled, size, img, bounds, xdraw.Over, nil)
		img = scaled
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
