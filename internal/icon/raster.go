package icon

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"

	"github.com/fyne-io/oksvg"
	"github.com/srwiley/rasterx"
)

// currentColor diisi ke SVG yang memakai fill="currentColor".
const currentColor = "#4f46e5"

// Rasterize mengubah SVG menjadi PNG persegi `size`x`size`.
//
// Ikon dari internet kadang berupa SVG dengan fitur yang tidak didukung renderer;
// SVG seperti itu bisa membuat proses menggambar di UI berhenti (panic) atau tidak
// tampil sama sekali. Karena itu SVG selalu diraster lebih dulu, dan kalau gagal
// pemanggil memakai ikon default.
func Rasterize(svg []byte, size int) (pngBytes []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("svg tidak bisa diraster: %v", r)
		}
	}()

	// banyak ikon internet memakai fill="currentColor" (mis. Material, Remix);
	// warnanya diganti warna brand supaya ikon monokrom tetap terlihat
	// di background terang maupun gelap.
	parsed, err := oksvg.ReadReplacingCurrentColor(bytes.NewReader(svg), currentColor)
	if err != nil {
		return nil, err
	}
	if parsed.ViewBox.W <= 0 || parsed.ViewBox.H <= 0 {
		return nil, errors.New("svg tanpa viewBox")
	}

	parsed.SetTarget(0, 0, float64(size), float64(size))
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	parsed.Draw(rasterx.NewDasher(size, size, scanner), 1.0)

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
