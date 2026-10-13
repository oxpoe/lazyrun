// Package update memeriksa rilis terbaru LazyRun di GitHub dan memasangnya.
package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	apiBase = "https://api.github.com/repos/oxpoe/lazyrun/releases"
	repoURL = "https://github.com/oxpoe/lazyrun"
)

// LAZYRUN_UPDATE_API bisa dipakai untuk mengarahkan cek update ke server lain
// (mis. untuk tes atau build dari fork).
func init() {
	if v := os.Getenv("LAZYRUN_UPDATE_API"); v != "" {
		apiBase = v
	}
}

// SetEndpoint mengarahkan pemeriksaan rilis ke alamat lain (dipakai tes).
func SetEndpoint(base string) {
	apiBase = base
}

// RepoURL adalah halaman repo LazyRun.
func RepoURL() string { return repoURL }

// Release adalah data rilis terbaru yang dibutuhkan LazyRun.
type Release struct {
	Tag      string // tag asli, mis. "v1.1.0"
	Version  string // versi tanpa awalan v, mis. "1.1.0"
	Notes    string // catatan rilis (markdown GitHub)
	PageURL  string // halaman rilis di GitHub
	AppImage string // URL unduhan AppImage (kosong kalau tidak ada)
	Zsync    string // URL berkas .zsync (opsional)
}

var client = &http.Client{Timeout: 20 * time.Second}

// Latest mengambil rilis terbaru dari GitHub.
func Latest() (Release, error) {
	var rel Release
	req, err := http.NewRequest(http.MethodGet, apiBase+"/latest", nil)
	if err != nil {
		return rel, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "LazyRun")

	resp, err := client.Do(req)
	if err != nil {
		return rel, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return rel, fmt.Errorf("GitHub menjawab %s", resp.Status)
	}

	var payload struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return rel, err
	}

	rel = Release{
		Tag:     payload.TagName,
		Version: strings.TrimPrefix(payload.TagName, "v"),
		Notes:   strings.TrimSpace(payload.Body),
		PageURL: payload.HTMLURL,
	}
	if rel.PageURL == "" {
		rel.PageURL = repoURL + "/releases/latest"
	}
	for _, asset := range payload.Assets {
		switch {
		case strings.HasSuffix(asset.Name, ".AppImage"):
			rel.AppImage = asset.URL
		case strings.HasSuffix(asset.Name, ".AppImage.zsync"):
			rel.Zsync = asset.URL
		}
	}
	return rel, nil
}

// Newer benar kalau rilis lebih baru dari versi yang sedang dipakai.
func Newer(current string, rel Release) bool {
	return CompareVersions(rel.Version, current) > 0
}

// CompareVersions membandingkan dua versi sederhana (mis. "1.2.0" vs "1.10"):
// hasil -1 kalau a lebih lama, 0 sama, 1 lebih baru. Awalan "v" diabaikan, dan
// versi dengan akhiran (mis. 1.0.0-rc1) dianggap lebih lama dari versi final.
func CompareVersions(a, b string) int {
	pa, sa := parseVersion(a)
	pb, sb := parseVersion(b)

	for i := 0; i < len(pa) || i < len(pb); i++ {
		var na, nb int
		if i < len(pa) {
			na = pa[i]
		}
		if i < len(pb) {
			nb = pb[i]
		}
		if na != nb {
			if na < nb {
				return -1
			}
			return 1
		}
	}
	if sa == sb {
		return 0
	}
	if sa == "" {
		return 1 // versi final lebih baru dari versi bersuffix
	}
	if sb == "" {
		return -1
	}
	return strings.Compare(sa, sb)
}

func parseVersion(v string) ([]int, string) {
	v = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "v"))
	main, suffix, _ := strings.Cut(v, "-")
	if plus := strings.IndexByte(main, '+'); plus >= 0 {
		main = main[:plus]
	}

	var numbers []int
	for _, part := range strings.Split(main, ".") {
		digits := part
		if i := strings.IndexFunc(part, func(r rune) bool { return r < '0' || r > '9' }); i >= 0 {
			digits = part[:i]
		}
		n, err := strconv.Atoi(digits)
		if err != nil {
			n = 0
		}
		numbers = append(numbers, n)
	}
	return numbers, suffix
}

// IsAppImage memeriksa apakah sebuah berkas benar-benar AppImage (ELF dengan
// magic AI di offset 8). Dipakai sebelum menimpa AppImage yang sedang dipakai.
func IsAppImage(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	head := make([]byte, 12)
	if _, err := f.Read(head); err != nil {
		return false
	}
	if !(head[0] == 0x7f && head[1] == 'E' && head[2] == 'L' && head[3] == 'F') {
		return false
	}
	return head[8] == 'A' && head[9] == 'I' && (head[10] == 0x01 || head[10] == 0x02)
}

// Download mengganti AppImage tujuan dengan rilis terbaru.
// Unduhan ditulis ke berkas sementara di folder yang sama, diverifikasi, baru
// ditukar secara atomik: kalau gagal, aplikasi lama tetap utuh.
func Download(rel Release, target string) error {
	if rel.AppImage == "" {
		return errors.New("rilis ini tidak menyertakan AppImage")
	}
	if target == "" {
		return errors.New("tidak tahu lokasi AppImage yang berjalan")
	}

	req, err := http.NewRequest(http.MethodGet, rel.AppImage, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "LazyRun")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unduhan gagal: %s", resp.Status)
	}

	tmp, err := os.CreateTemp(filepath.Dir(target), ".lazyrun-update-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := copyWithLimit(tmp, resp.Body, maxDownload); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if !IsAppImage(tmp.Name()) {
		return errors.New("berkas yang diunduh bukan AppImage yang sah")
	}
	return os.Rename(tmp.Name(), target)
}
