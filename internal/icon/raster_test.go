package icon

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRasterizeSimpleSVG(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">` +
		`<circle cx="12" cy="12" r="10" fill="#f0f"/></svg>`)

	out, err := Rasterize(svg, 64)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("hasil raster bukan PNG: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 64 || b.Dy() != 64 {
		t.Errorf("ukuran hasil = %v", b)
	}
}

func TestRasterizeSVGWithoutViewBox(t *testing.T) {
	// banyak ikon internet hanya punya width/height
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256"><rect width="256" height="256"/></svg>`)
	if _, err := Rasterize(svg, 32); err != nil {
		t.Errorf("svg tanpa viewBox harus tetap bisa: %v", err)
	}
}

func TestRasterizeRejectsGarbageWithoutPanic(t *testing.T) {
	cases := map[string][]byte{
		"bukan svg":            []byte("<html><body>404</body></html>"),
		"svg kosong":           []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`),
		"path ngawur":          []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><path d="M NaN NaN L x y"/></svg>`),
		"data terpotong":       []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><path d="M0 0 L5 5"`),
		"koordinat tak hingga": []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><path d="M0 0 L1e400 1e400"/></svg>`),
	}

	for name, svg := range cases {
		t.Run(name, func(t *testing.T) {
			// inti tes: tidak panic. Error boleh.
			_, _ = Rasterize(svg, 32)
		})
	}
}

func TestExtSniffsSVG(t *testing.T) {
	if got := Ext([]byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"/>`)); got != ".svg" {
		t.Errorf("Ext(svg) = %q", got)
	}
	if got := Ext(pngBytes); got != ".png" {
		t.Errorf("Ext(png) = %q", got)
	}
}

func TestRasterizeColorsCurrentColorIcons(t *testing.T) {
	// ikon Material/Remix memakai fill="currentColor" tanpa warna eksplisit;
	// tanpa penggantian warna, rasterisasi gagal dan thumbnail tidak muncul.
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">` +
		`<path fill="currentColor" d="M3 3h18v18H3z"/></svg>`)

	out, err := Rasterize(svg, 32)
	if err != nil {
		t.Fatalf("ikon currentColor harus bisa diraster: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	r, g, b, a := img.At(16, 16).RGBA()
	if a == 0 {
		t.Fatal("ikon currentColor tidak terwarnai (transparan)")
	}
	// brand indigo #4F46E5: biru > merah > hijau
	if !(b > r && r > g) {
		t.Errorf("warna pengganti bukan brand indigo: r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}
