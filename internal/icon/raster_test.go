package icon

import (
	"bytes"
	"encoding/binary"
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
	// brand dusk pink #C25C9E: merah > biru > hijau
	if !(r > b && b > g) {
		t.Errorf("warna pengganti bukan warna brand: r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

// ico32bpp membuat ICO berisi gambar BMP 32-bit (format favicon lama).
func ico32bpp(width int, pixel func(x, y int) (byte, byte, byte, byte)) []byte {
	stride := (width*4 + 3) & ^3
	dib := make([]byte, 40+stride*width)
	binary.LittleEndian.PutUint32(dib[0:], 40)
	binary.LittleEndian.PutUint32(dib[4:], uint32(width))
	binary.LittleEndian.PutUint32(dib[8:], uint32(width*2)) // XOR + AND
	binary.LittleEndian.PutUint16(dib[12:], 1)
	binary.LittleEndian.PutUint16(dib[14:], 32)

	for y := 0; y < width; y++ {
		row := dib[40+(width-1-y)*stride:]
		for x := 0; x < width; x++ {
			b, g, r, a := pixel(x, y)
			row[x*4], row[x*4+1], row[x*4+2], row[x*4+3] = b, g, r, a
		}
	}

	ico := make([]byte, 22+len(dib))
	ico[2], ico[3] = 1, 0
	ico[4], ico[5] = 1, 0
	put32 := func(off, v int) {
		binary.LittleEndian.PutUint32(ico[off:], uint32(v))
	}
	ico[6] = byte(width) // lebar (0 = 256)
	ico[7] = byte(width)
	put32(14, len(dib))
	put32(18, 22)
	copy(ico[22:], dib)
	return ico
}

func TestToPNGConvertsBMPIcon(t *testing.T) {
	ico := ico32bpp(16, func(x, y int) (byte, byte, byte, byte) {
		return 0x20, 0x80, 0xE0, 0xFF // BGR + alpha
	})

	out, err := ToPNG(ico)
	if err != nil {
		t.Fatalf("ico BMP harus bisa dikonversi: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	r, g, b, a := img.At(8, 8).RGBA()
	if a == 0 {
		t.Fatal("piksel ICO BMP transparan, seharusnya opaque")
	}
	// warna BGR 0x20,0x80,0xE0 = RGB (0xE0,0x80,0x20)
	if r>>8 != 0xE0 || g>>8 != 0x80 || b>>8 != 0x20 {
		t.Errorf("warna salah: r=%#x g=%#x b=%#x", r>>8, g>>8, b>>8)
	}
}

func TestDecodeICORejectsGarbage(t *testing.T) {
	if _, err := decodeICO([]byte("bukan ico")); err == nil {
		t.Error("data bukan ICO harus error")
	}
	if _, err := decodeICO(icoWithPNG([]byte("bukan png"))); err == nil {
		t.Error("entry PNG rusak harus error")
	}
}
