package icon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fakeIconify(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			got := r.URL.Query().Get("query")
			if got == "" {
				http.Error(w, "no query", http.StatusBadRequest)
				return
			}
			// ikon whatsapp hanya untuk query "whatsapp" polos, supaya fallback ikut teruji
			if !strings.Contains(strings.ToLower(got), "whatsapp") || strings.Contains(got, ".") {
				_, _ = w.Write([]byte(`{"icons":[],"total":0}`))
				return
			}
			_, _ = w.Write([]byte(`{"icons":["ic:baseline-whatsapp","mdi:whatsapp","logos:whatsapp-icon"],"total":3}`))
		case "/logos/whatsapp-icon.svg":
			w.Header().Set("Content-Type", "image/svg+xml")
			_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256"></svg>`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestSearchBuildsDownloadURLs(t *testing.T) {
	srv := fakeIconify(t)
	SetSearchEndpoint(srv.URL)
	t.Cleanup(func() { SetSearchEndpoint("https://api.iconify.design") })

	results, err := Search("whatsapp", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("hasil pencarian = %d, mau 3", len(results))
	}
	if results[0].Name != "ic:baseline-whatsapp" || results[0].ShortName() != "baseline-whatsapp" {
		t.Errorf("hasil pertama salah: %+v", results[0])
	}
	if results[2].URL != srv.URL+"/logos/whatsapp-icon.svg?height=256" {
		t.Errorf("URL unduhan salah: %s", results[2].URL)
	}
}

func TestQueryHintAndSearchBest(t *testing.T) {
	srv := fakeIconify(t)
	SetSearchEndpoint(srv.URL)
	t.Cleanup(func() { SetSearchEndpoint("https://api.iconify.design") })

	// nama seperti host: TLD dibuang, kata terpanjang dipakai
	if got := QueryHint("web.whatsapp.com"); got != "whatsapp" {
		t.Errorf("QueryHint(host) = %q", got)
	}
	if got := QueryHint("WhatsApp Web"); got != "WhatsApp" {
		t.Errorf("QueryHint(nama) = %q", got)
	}
	if got := QueryHint(""); got != "" {
		t.Errorf("QueryHint(kosong) = %q", got)
	}

	// query mentah gagal, tapi varian "whatsapp" berhasil
	results, err := SearchBest("web.whatsapp.com", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Errorf("SearchBest() = %d hasil", len(results))
	}
}

func TestFindIgnoresGenericWords(t *testing.T) {
	srv := fakeIconify(t)
	SetSearchEndpoint(srv.URL)
	t.Cleanup(func() { SetSearchEndpoint("https://api.iconify.design") })

	// kata <= 3 huruf ("my", "go", "app") tidak boleh dipakai untuk mencari
	if _, ok := Find("my-go-app"); ok {
		t.Error("nama tanpa ikon resmi harus gagal, bukan dapat ikon acak")
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	if _, err := Search("  ", 5); err == nil {
		t.Error("query kosong harus error")
	}
}

func TestFindPicksMatchingName(t *testing.T) {
	srv := fakeIconify(t)
	SetSearchEndpoint(srv.URL)
	t.Cleanup(func() { SetSearchEndpoint("https://api.iconify.design") })

	// "WhatsApp Web" harus cocok dengan ikon bernama whatsapp, bukan hasil acak.
	// (hasil pertama "ic:baseline-whatsapp" tidak cocok karena namanya lebih panjang,
	// jadi yang dipakai "mdi:whatsapp")
	got, ok := Find("WhatsApp Web")
	if !ok {
		t.Fatal("Find() tidak menemukan kandidat")
	}
	if got.Name != "mdi:whatsapp" {
		t.Errorf("ikon terpilih = %s", got.Name)
	}

	// Nama yang tidak berhubungan tidak boleh dipaksa memakai ikon apa pun.
	if _, ok := Find("zzz-tidak-ada-xyz"); ok {
		t.Error("nama asing tidak boleh dapat ikon acak")
	}
}

func TestFromURLAlwaysReturnsPNG(t *testing.T) {
	srv := fakeIconify(t)
	data, ext, err := FromURL(srv.URL + "/logos/whatsapp-icon.svg")
	if err != nil {
		t.Fatal(err)
	}
	// SVG dari internet dikonversi ke PNG supaya seragam
	if ext != ".png" || len(data) == 0 {
		t.Errorf("FromURL = ext %q, %d byte", ext, len(data))
	}
	if !bytes.HasPrefix(data, pngMagic) {
		t.Errorf("hasil FromURL bukan PNG: %q", data[:min(8, len(data))])
	}

	if _, _, err := FromURL(srv.URL + "/tidak-ada"); err == nil {
		t.Error("URL 404 harus error")
	}
}

func TestIsImageURLAndDomain(t *testing.T) {
	images := []string{"https://contoh.com/ikon.png", "http://a.b/logo.svg", "https://x.id/i.JPEG?v=2"}
	for _, u := range images {
		if !IsImageURL(u) {
			t.Errorf("IsImageURL(%q) = false", u)
		}
	}
	if IsImageURL("https://contoh.com") || IsImageURL("winbox") {
		t.Error("URL halaman/nama biasa bukan gambar")
	}

	domains := map[string]string{
		"mikrotik.com":            "mikrotik.com",
		"  web.whatsapp.com  ":    "web.whatsapp.com",
		"https://mikrotik.com/":   "mikrotik.com",
		"winbox":                  "",
		"WhatsApp Web":            "",
		"my-app v1.2":             "",
		"lazyrun.png":             "",
		"file:///home/x/icon.png": "",
	}
	for in, want := range domains {
		if got := Domain(in); got != want {
			t.Errorf("Domain(%q) = %q, mau %q", in, got, want)
		}
	}
}

func TestSiteIconsListsSiteAndService(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			_, _ = w.Write([]byte(`<html><head>` +
				`<link rel="manifest" href="/manifest.json">` +
				`<link rel="apple-touch-icon" sizes="180x180" href="/touch.png">` +
				`<link rel="shortcut icon" href="/favicon.ico">` +
				`</head></html>`))
		case "/manifest.json":
			_, _ = w.Write([]byte(`{"icons":[{"src":"/big-192.png","sizes":"192x192"},{"src":"/big-512.png","sizes":"512x512"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	list, err := SiteIcons(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 3 {
		t.Fatalf("ikon situs = %d, mau minimal 3: %+v", len(list), list)
	}
	// ikon manifest terbesar harus di depan, lalu ikon halaman, lalu layanan favicon
	if list[0].URL != srv.URL+"/big-512.png" {
		t.Errorf("kandidat pertama = %+s", list[0])
	}
	found := false
	for _, r := range list {
		if strings.HasPrefix(r.URL, "https://www.google.com/s2/favicons") {
			found = true
		}
	}
	if !found {
		t.Error("layanan favicon tidak disertakan sebagai pilihan")
	}
}

func TestSiteIconsNeedsValidDomain(t *testing.T) {
	if _, err := SiteIcons(""); err == nil {
		t.Error("alamat kosong harus error")
	}
}
