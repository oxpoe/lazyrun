// Package icon: favicon.go mengambil ikon dari halaman web untuk shortcut PWA.
package icon

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

var client = &http.Client{Timeout: 8 * time.Second}

// shortcut: parsing HTML pakai regex, bukan parser penuh; cukup untuk tag <link rel=icon>.
var (
	linkTag  = regexp.MustCompile(`(?is)<link\b[^>]*>`)
	relAttr  = regexp.MustCompile(`(?is)\brel\s*=\s*["']?([^"'>]+)`)
	hrefAttr = regexp.MustCompile(`(?is)\bhref\s*=\s*["']?([^"'\s>]+)`)
	typeAttr = regexp.MustCompile(`(?is)\btype\s*=\s*["']?([^"'\s>]+)`)
	sizeAttr = regexp.MustCompile(`(?is)\bsizes\s*=\s*["']?(\d+)`)
)

// Fetch mengambil favicon dari halaman. Return data + ekstensi file (".png", ".jpg", ".svg").
func Fetch(pageURL string) ([]byte, string, error) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, "", err
	}

	var lastErr error
	for _, candidate := range iconCandidates(base) {
		data, err := get(candidate)
		if err != nil {
			lastErr = err
			continue
		}
		if ext := sniffExt(data); ext != "" {
			return data, ext, nil
		}
		if png := largestPNGFromICO(data); png != nil {
			return png, ".png", nil
		}
		lastErr = errors.New("unsupported icon format at " + candidate)
	}
	if lastErr == nil {
		lastErr = errors.New("no favicon found")
	}
	return nil, "", lastErr
}

// iconCandidate adalah kandidat URL ikon dengan skor prioritas.
type iconCandidate struct {
	url   string
	score int
}

// iconCandidates mengumpulkan URL ikon dari HTML, ikon besar didahulukan, favicon.ico terakhir.
func iconCandidates(base *url.URL) []string {
	var found []iconCandidate

	if page, err := get(base.String()); err == nil {
		for _, tag := range linkTag.FindAllString(string(page), -1) {
			rel := strings.ToLower(attr(relAttr, tag))
			if !strings.Contains(rel, "icon") || strings.Contains(rel, "mask-icon") {
				continue
			}
			href := attr(hrefAttr, tag)
			if href == "" {
				continue
			}
			ref, err := url.Parse(href)
			if err != nil {
				continue
			}
			score := 1
			if strings.Contains(strings.ToLower(attr(typeAttr, tag)), "png") {
				score += 10
			}
			if size := attr(sizeAttr, tag); size != "" {
				fmt.Sscanf(size, "%d", &score)
				score += 10
			}
			found = append(found, iconCandidate{base.ResolveReference(ref).String(), score})
		}
	}

	sort.SliceStable(found, func(i, j int) bool { return found[i].score > found[j].score })

	out := make([]string, 0, len(found)+1)
	seen := map[string]bool{}
	for _, c := range found {
		if !seen[c.url] {
			seen[c.url] = true
			out = append(out, c.url)
		}
	}
	fallback := base.ResolveReference(&url.URL{Path: "/favicon.ico"}).String()
	if !seen[fallback] {
		out = append(out, fallback)
	}
	return out
}

func attr(re *regexp.Regexp, tag string) string {
	if m := re.FindStringSubmatch(tag); m != nil {
		return m[1]
	}
	return ""
}

func get(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) LazyRun")
	req.Header.Set("Accept", "text/html,image/*;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", rawURL, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 2<<20))
}

func sniffExt(data []byte) string {
	head := data
	if len(head) > 512 {
		head = head[:512]
	}
	switch {
	case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")):
		return ".png"
	case bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}):
		return ".jpg"
	case bytes.Contains(bytes.ToLower(head), []byte("<svg")):
		return ".svg"
	}
	return ""
}

// largestPNGFromICO mengambil PNG terbesar di dalam file .ico.
// ICO dari BMP lama tidak didukung (langka untuk favicon).
func largestPNGFromICO(data []byte) []byte {
	if len(data) < 22 || data[0] != 0 || data[1] != 0 || data[2] != 1 || data[3] != 0 {
		return nil
	}
	count := int(data[4]) | int(data[5])<<8
	var best []byte
	for i := 0; i < count; i++ {
		off := 6 + i*16
		if off+16 > len(data) {
			break
		}
		size := int(data[off+8]) | int(data[off+9])<<8 | int(data[off+10])<<16 | int(data[off+11])<<24
		start := int(data[off+12]) | int(data[off+13])<<8 | int(data[off+14])<<16 | int(data[off+15])<<24
		if size <= 0 || start < 0 || start+size > len(data) {
			continue
		}
		blob := data[start : start+size]
		if bytes.HasPrefix(blob, []byte("\x89PNG\r\n\x1a\n")) && len(blob) > len(best) {
			best = blob
		}
	}
	return best
}
