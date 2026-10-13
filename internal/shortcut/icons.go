package shortcut

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"lazyrun/internal/icon"
	"lazyrun/internal/store"
)

// Prefix sumber ikon yang disimpan di metadata, dipakai untuk sync ulang.
const (
	iconSourceFile    = "file:"
	iconSourceFavicon = "favicon:"
	iconSourceSearch  = "search:"
)

type iconOptions struct {
	IconPath   string // file pilihan user
	IconURL    string // URL ikon pilihan user
	IconData   []byte // ikon pilihan user yang sudah diunduh UI
	NearbyDir  string // folder binary, untuk auto-deteksi ikon di sebelahnya
	NearbyBase string
	PageURL    string // halaman PWA, untuk favicon
	Query      string // nama aplikasi, untuk cari ikon online
	Slug       string
}

// installIconFor menentukan ikon shortcut dan sumbernya.
// Urutan: file user → URL user → ikon di folder binary → favicon situs → cari online → ikon default.
// Kandidat yang gagal (file rusak, unduhan gagal, SVG tidak bisa diraster) dilewati,
// jadi install tidak pernah batal hanya karena ikon.
func installIconFor(o iconOptions) (path, source string, err error) {
	attempts := []struct {
		source string
		load   func() ([]byte, string, error)
	}{
		{iconSourceFile + o.IconPath, func() ([]byte, string, error) {
			if o.IconPath == "" {
				return nil, "", errSkipIcon
			}
			data, err := os.ReadFile(o.IconPath)
			return data, strings.ToLower(filepath.Ext(o.IconPath)), err
		}},
		{o.IconURL, func() ([]byte, string, error) {
			// ikon pilihan dari pencarian online: datanya sudah diunduh UI
			if len(o.IconData) > 0 {
				return o.IconData, icon.Ext(o.IconData), nil
			}
			if o.IconURL == "" {
				return nil, "", errSkipIcon
			}
			return icon.FromURL(o.IconURL)
		}},
		{iconSourceFile + icon.Nearby(o.NearbyDir, o.NearbyBase), func() ([]byte, string, error) {
			nearby := icon.Nearby(o.NearbyDir, o.NearbyBase)
			if nearby == "" {
				return nil, "", errSkipIcon
			}
			data, err := os.ReadFile(nearby)
			return data, strings.ToLower(filepath.Ext(nearby)), err
		}},
		{iconSourceFavicon + o.PageURL, func() ([]byte, string, error) {
			if o.PageURL == "" {
				return nil, "", errSkipIcon
			}
			return icon.Fetch(o.PageURL)
		}},
		{iconSourceSearch + o.Query, func() ([]byte, string, error) {
			if o.Query == "" {
				return nil, "", errSkipIcon
			}
			result, ok := icon.Find(o.Query)
			if !ok {
				return nil, "", errSkipIcon
			}
			return icon.FromURL(result.URL)
		}},
	}

	for _, attempt := range attempts {
		data, _, err := attempt.load()
		if err != nil || len(data) == 0 {
			continue
		}
		path, err := icon.Save(data, o.Slug)
		if err != nil {
			continue
		}
		return path, attempt.source, nil
	}

	path, err = icon.Default(o.Slug)
	return path, "", err
}

// errSkipIcon menandai kandidat ikon yang tidak tersedia.
var errSkipIcon = errors.New("icon source not available")

// RefreshIcon mengunduh ulang ikon shortcut dari sumber online-nya (sync manual).
func RefreshIcon(sc store.Shortcut) (string, error) {
	var data []byte
	var err error

	switch {
	case sc.IconSource == "":
		return "", errors.New("this icon has no online source")
	case strings.HasPrefix(sc.IconSource, iconSourceFavicon):
		data, _, err = icon.Fetch(strings.TrimPrefix(sc.IconSource, iconSourceFavicon))
	case strings.HasPrefix(sc.IconSource, iconSourceSearch):
		result, ok := icon.Find(strings.TrimPrefix(sc.IconSource, iconSourceSearch))
		if !ok {
			return "", errors.New("no matching icon found online anymore")
		}
		data, _, err = icon.FromURL(result.URL)
	case strings.HasPrefix(sc.IconSource, "http"):
		data, _, err = icon.FromURL(sc.IconSource)
	default:
		return "", errors.New("this icon comes from a local file, not the internet")
	}
	if err != nil {
		return "", err
	}

	path, err := icon.Save(data, sc.Slug)
	if err != nil {
		return "", err
	}
	if sc.IconPath != "" && sc.IconPath != path {
		_ = os.Remove(sc.IconPath)
	}
	if err := replaceIconLine(sc.DesktopPath, path); err != nil {
		return path, err
	}
	if sc.DesktopCopyPath != "" {
		if err := replaceIconLine(sc.DesktopCopyPath, path); err != nil {
			return path, err
		}
	}
	go RefreshCache()
	return path, nil
}

// replaceIconLine menimpa baris Icon= di file .desktop tanpa mengubah mode file.
func replaceIconLine(path, iconPath string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "Icon=") {
			lines[i] = "Icon=" + escapeValue(iconPath)
			break
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), info.Mode().Perm())
}
