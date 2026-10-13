package icon

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"

	"golang.org/x/image/bmp"
)

func solidImg(size int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			img.Set(x, y, color.NRGBA{R: 0x12, G: 0x9f, B: 0x55, A: 0xFF})
		}
	}
	return img
}

func encode(t *testing.T, fn func(*bytes.Buffer) error) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := fn(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestToPNGConvertsFormats(t *testing.T) {
	src := solidImg(64)

	cases := map[string][]byte{
		"png": encode(t, func(b *bytes.Buffer) error { return png.Encode(b, src) }),
		"gif": encode(t, func(b *bytes.Buffer) error { return gif.Encode(b, src, nil) }),
		"bmp": encode(t, func(b *bytes.Buffer) error { return bmp.Encode(b, src) }),
		"jpg": encode(t, func(b *bytes.Buffer) error { return jpeg.Encode(b, src, nil) }),
		"svg": []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8 8"><rect width="8" height="8"/></svg>`),
		"ico": icoWithPNG(encode(t, func(b *bytes.Buffer) error { return png.Encode(b, solidImg(16)) })),
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := ToPNG(data)
			if err != nil {
				t.Fatalf("ToPNG(%s) gagal: %v", name, err)
			}
			if !bytes.HasPrefix(out, pngMagic) {
				t.Fatalf("hasil %s bukan PNG", name)
			}
			if _, err := png.Decode(bytes.NewReader(out)); err != nil {
				t.Fatalf("hasil %s tidak bisa di-decode: %v", name, err)
			}
		})
	}
}

func TestToPNGRejectsUnknown(t *testing.T) {
	if _, err := ToPNG([]byte("halo dunia")); err == nil {
		t.Error("data bukan gambar harus error")
	}
}

func TestToPNGShrinksHugeImages(t *testing.T) {
	out, err := ToPNG(encode(t, func(b *bytes.Buffer) error { return png.Encode(b, solidImg(1024)) }))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) >= 1024*1024 {
		t.Errorf("gambar besar tidak diperkecil: %d byte", len(out))
	}
}

func TestSaveAlwaysWritesPNG(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	// ikon webp tak disupport ditest lewat gif: hasil harus selalu .png
	path, err := Save(encode(t, func(b *bytes.Buffer) error { return gif.Encode(b, solidImg(32), nil) }), "contoh")
	if err != nil {
		t.Fatal(err)
	}
	if got := path[len(path)-4:]; got != ".png" {
		t.Errorf("path ikon = %s", path)
	}
}
