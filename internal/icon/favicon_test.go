package icon

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func icoWithPNG(png []byte) []byte {
	ico := make([]byte, 22+len(png))
	ico[2], ico[3] = 1, 0 // type = icon
	ico[4], ico[5] = 1, 0 // jumlah gambar
	put32 := func(off, v int) {
		ico[off] = byte(v)
		ico[off+1] = byte(v >> 8)
		ico[off+2] = byte(v >> 16)
		ico[off+3] = byte(v >> 24)
	}
	put32(14, len(png)) // bytesInRes
	put32(18, 22)       // imageOffset
	copy(ico[22:], png)
	return ico
}

var png = append([]byte("\x89PNG\r\n\x1a\n"), []byte("payload")...)

func TestFetchUsesLinkedIcon(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			io.WriteString(w, `<html><head>`+
				`<link rel="apple-touch-icon" href="/small.ico">`+
				`<link rel="icon" type="image/png" sizes="192x192" href="/big.png">`+
				`</head></html>`)
		case "/big.png":
			w.Write(png)
		default:
			w.Write(icoWithPNG(png))
		}
	}))
	defer srv.Close()

	data, ext, err := Fetch(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if ext != ".png" || string(data) != string(png) {
		t.Errorf("ikon terpilih salah: ext=%s data=%q", ext, data)
	}
}

func TestFetchFallsBackToFaviconICO(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			io.WriteString(w, `<html><head><title>no icon</title></head></html>`)
			return
		}
		w.Write(icoWithPNG(png))
	}))
	defer srv.Close()

	data, ext, err := Fetch(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if ext != ".png" || string(data) != string(png) {
		t.Errorf("PNG di dalam .ico harus dipakai: ext=%s data=%q", ext, data)
	}
}

func TestFetchFailsWithoutIcon(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	if _, _, err := Fetch(srv.URL); err == nil {
		t.Error("tanpa favicon harus error supaya UI pakai ikon default")
	}
}

func TestLargestPNGFromICO(t *testing.T) {
	if got := largestPNGFromICO(icoWithPNG(png)); string(got) != string(png) {
		t.Errorf("PNG di dalam .ico tidak terbaca: %q", got)
	}
	if got := largestPNGFromICO([]byte("bukan ico")); got != nil {
		t.Errorf("data bukan ICO harus menghasilkan nil, dapat %q", got)
	}
	// ICO berisi BMP lama (bukan PNG) tidak didukung.
	bmp := make([]byte, 40)
	bmp[0], bmp[1], bmp[2], bmp[3] = 0, 0, 1, 0
	bmp[4], bmp[5] = 1, 0
	if got := largestPNGFromICO(bmp); got != nil {
		t.Errorf("ICO BMP harus mengembalikan nil, dapat %q", got)
	}
}

func TestSniffExt(t *testing.T) {
	if ext := sniffExt(png); ext != ".png" {
		t.Errorf("sniffExt(png) = %s", ext)
	}
	if ext := sniffExt([]byte(`<?xml version="1.0"?><svg xmlns="..."><path/></svg>`)); ext != ".svg" {
		t.Errorf("sniffExt(svg) = %s", ext)
	}
	if ext := sniffExt([]byte("\xff\xd8\xff\xe0jpeg")); ext != ".jpg" {
		t.Errorf("sniffExt(jpg) = %s", ext)
	}
	if ext := sniffExt([]byte("plain text")); ext != "" {
		t.Errorf("sniffExt(text) = %s", ext)
	}
}
