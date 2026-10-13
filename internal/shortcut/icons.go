package shortcut

import (
	"errors"
	"os"
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
func installIconFor(o iconOptions) (path, source string, err error) {
	if o.IconPath != "" {
		path, err = icon.Save(o.IconPath, "", "", o.Slug)
		return path, iconSourceFile + o.IconPath, err
	}
	if len(o.IconData) > 0 && o.IconURL != "" {
		if path, err := icon.SaveData(o.IconData, o.Slug); err == nil {
			return path, o.IconURL, nil
		}
	}
	if o.IconURL != "" {
		if data, ext, err := icon.FromURL(o.IconURL); err == nil {
			path, err := icon.SaveBytes(data, ext, o.Slug)
			return path, o.IconURL, err
		}
	}
	if nearby := icon.Nearby(o.NearbyDir, o.NearbyBase); nearby != "" {
		path, err = icon.Save(nearby, "", "", o.Slug)
		return path, iconSourceFile + nearby, err
	}
	if o.PageURL != "" {
		if data, ext, err := icon.Fetch(o.PageURL); err == nil {
			path, err := icon.SaveBytes(data, ext, o.Slug)
			return path, iconSourceFavicon + o.PageURL, err
		}
	}
	if o.Query != "" {
		if result, ok := icon.Find(o.Query); ok {
			if data, ext, err := icon.FromURL(result.URL); err == nil {
				path, err := icon.SaveBytes(data, ext, o.Slug)
				return path, iconSourceSearch + o.Query, err
			}
		}
	}
	path, err = icon.Default(o.Slug)
	return path, "", err
}

// RefreshIcon mengunduh ulang ikon shortcut dari sumber online-nya (sync manual).
func RefreshIcon(sc store.Shortcut) (string, error) {
	var data []byte
	var ext string
	var err error

	switch {
	case sc.IconSource == "":
		return "", errors.New("this icon has no online source")
	case strings.HasPrefix(sc.IconSource, iconSourceFavicon):
		data, ext, err = icon.Fetch(strings.TrimPrefix(sc.IconSource, iconSourceFavicon))
	case strings.HasPrefix(sc.IconSource, iconSourceSearch):
		result, ok := icon.Find(strings.TrimPrefix(sc.IconSource, iconSourceSearch))
		if !ok {
			return "", errors.New("no matching icon found online anymore")
		}
		data, ext, err = icon.FromURL(result.URL)
	case strings.HasPrefix(sc.IconSource, "http"):
		data, ext, err = icon.FromURL(sc.IconSource)
	default:
		return "", errors.New("this icon comes from a local file, not the internet")
	}
	if err != nil {
		return "", err
	}

	path, err := icon.SaveBytes(data, ext, sc.Slug)
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
