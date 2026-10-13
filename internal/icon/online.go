package icon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	iconifyMu   sync.RWMutex
	iconifyBase = "https://api.iconify.design"
)

// SetSearchEndpoint mengarahkan pencarian ikon ke endpoint lain (dipakai tes).
func SetSearchEndpoint(base string) {
	iconifyMu.Lock()
	defer iconifyMu.Unlock()
	iconifyBase = base
}

// searchEndpoint dibaca setiap kali mencari, jadi harus lewat lock
// (tes bisa mengganti endpoint sementara goroutine lain mencari).
func searchEndpoint() string {
	iconifyMu.RLock()
	defer iconifyMu.RUnlock()
	return iconifyBase
}

// Result adalah satu ikon dari pencarian online.
type Result struct {
	Name string // contoh: "logos:whatsapp-icon"
	URL  string // URL SVG siap diunduh
}

// ShortName mengembalikan nama ikon tanpa prefix koleksi.
func (r Result) ShortName() string {
	_, name, ok := strings.Cut(r.Name, ":")
	if !ok {
		return r.Name
	}
	return name
}

// Search mencari ikon di Iconify berdasarkan kata kunci.
func Search(query string, limit int) ([]Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("search query is empty")
	}
	if limit <= 0 {
		limit = 32
	}

	base := searchEndpoint()
	endpoint := fmt.Sprintf("%s/search?query=%s&limit=%d", base, url.QueryEscape(query), limit)
	data, err := get(endpoint)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Icons []string `json:"icons"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(payload.Icons))
	for _, name := range payload.Icons {
		prefix, icon, ok := strings.Cut(name, ":")
		if !ok {
			continue
		}
		// height=256 supaya SVG-nya punya ukuran eksplisit (bukan "1em")
		results = append(results, Result{
			Name: name,
			URL:  fmt.Sprintf("%s/%s/%s.svg?height=256", base, url.PathEscape(prefix), url.PathEscape(icon)),
		})
	}
	if len(results) == 0 {
		return nil, errors.New("no icon found for " + query)
	}
	return results, nil
}

// SearchBest mencoba beberapa variasi kata kunci supaya nama seperti
// "web.whatsapp.com" atau "WhatsApp Web" tetap menemukan ikon.
func SearchBest(query string, limit int) ([]Result, error) {
	for _, candidate := range queryVariants(query) {
		if results, err := Search(candidate, limit); err == nil {
			return results, nil
		}
	}
	return nil, errors.New("no icon found for " + strings.TrimSpace(query))
}

// Find mengambil hasil pertama yang namanya cocok dengan kata kunci,
// supaya ikon acak tidak dipasang ke aplikasi yang namanya tidak dikenal.
func Find(query string) (Result, bool) {
	for _, candidate := range queryVariants(query) {
		results, err := Search(candidate, 32)
		if err != nil {
			continue
		}
		want := slug(candidate)
		for _, r := range results {
			name := slug(r.ShortName())
			if name == want || strings.Contains(name, want) || strings.Contains(want, name) {
				return r, true
			}
		}
	}
	return Result{}, false
}

// QueryHint mengembalikan kata kunci paling bermakna dari sebuah nama,
// dipakai untuk mengisi kolom pencarian ikon.
func QueryHint(query string) string {
	variants := queryVariants(query)
	if len(variants) > 1 {
		return variants[1]
	}
	return strings.TrimSpace(query)
}

// queryVariants menyusun kandidat kata kunci: teks asli dulu, lalu kata-kata
// yang lebih sederhana (kata terpanjang didahulukan). Kata < 4 huruf dibuang
// supaya pencarian tidak jatuh ke ikon acak.
func queryVariants(query string) []string {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	out := []string{query}
	base := query
	// host seperti "web.whatsapp.com" → buang TLD-nya
	if !strings.Contains(base, " ") && strings.Contains(base, ".") {
		parts := strings.Split(base, ".")
		if len(parts) > 1 {
			base = strings.Join(parts[:len(parts)-1], " ")
		}
	}

	words := strings.FieldsFunc(base, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	sort.SliceStable(words, func(i, j int) bool { return len(words[i]) > len(words[j]) })
	for _, word := range words {
		if len(word) < 4 || strings.EqualFold(word, query) {
			continue
		}
		out = append(out, word)
	}
	return out
}

// FromURL mengunduh ikon dari URL. Hasilnya selalu PNG: format apa pun
// (svg, webp, gif, ico, jpeg) dikonversi dulu supaya desktop environment
// tidak perlu mendukung banyak format.
func FromURL(rawURL string) ([]byte, string, error) {
	data, err := get(rawURL)
	if err != nil {
		// sekali coba lagi: API ikon publik kadang membatasi permintaan beruntun
		time.Sleep(600 * time.Millisecond)
		data, err = get(rawURL)
	}
	if err != nil {
		return nil, "", err
	}
	converted, err := ToPNG(data)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", rawURL, err)
	}
	return converted, ".png", nil
}

// imageExts adalah ekstensi gambar yang bisa dipakai langsung sebagai ikon.
var imageExts = []string{".png", ".svg", ".jpg", ".jpeg", ".webp", ".gif", ".bmp", ".ico", ".xpm"}

// IsImageURL benar kalau teks yang diketik user berupa URL gambar.
func IsImageURL(text string) bool {
	text = strings.TrimSpace(text)
	if !strings.Contains(text, "://") {
		return false
	}
	lower := strings.ToLower(text)
	for _, ext := range imageExts {
		if strings.HasSuffix(strings.SplitN(lower, "?", 2)[0], ext) {
			return true
		}
	}
	return false
}

// Domain mengembalikan domain kalau user mengetik alamat situs (mis. "mikrotik.com"),
// bukan nama aplikasi biasa.
func Domain(text string) string {
	text = strings.TrimSpace(text)
	if text == "" || strings.ContainsAny(text, " \t") || IsImageURL(text) {
		return ""
	}
	// nama file gambar (mis. "lazyrun.png") bukan domain
	lower := strings.ToLower(strings.SplitN(text, "?", 2)[0])
	for _, ext := range imageExts {
		if strings.HasSuffix(lower, ext) {
			return ""
		}
	}
	host := text
	if strings.Contains(text, "://") {
		if u, err := url.Parse(text); err == nil {
			host = u.Hostname()
		}
	}
	parts := strings.Split(strings.Trim(host, "."), ".")
	if len(parts) < 2 {
		return ""
	}
	tld := parts[len(parts)-1]
	if len(tld) < 2 || len(tld) > 24 {
		return ""
	}
	for _, r := range strings.Join(parts, "") {
		if !(r == '-' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return ""
		}
	}
	return host
}

// slug menyamakan nama untuk perbandingan (huruf kecil, tanpa pemisah).
func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
